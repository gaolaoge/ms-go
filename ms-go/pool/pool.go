package pool

/*
需求：
1. 希望可以控制至多创建固定数量的协程
2. 有一个任务队列，等待协程进行调度执行
3. 协程用完时，其它任务处于等待状态，一旦有协程空余，立即获取任务执行
4. 协程长时间空余则对其清理，以避免浪费
5. 有超时任务，主动让出协程
*/

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultExpire = 3

var (
	ErrorInValidCap    = errors.New("pool cap cannot <= 0")
	ErrorInValidExpire = errors.New("pool expire cannot <= 0")
	ErrorHasClose      = errors.New("pool has bean released")
)

type sig struct{}

type Pool struct {
	cap          int32         // 容量
	running      int32         // 正在运行中的 worker 数量
	workers      []*Worker     // 若干空闲 worker
	expire       time.Duration // 过期时间，规定 worker 的空闲时间超过此值就回收掉
	release      chan sig      // 释放资源信号，接收到它表示pool不再使用
	lock         sync.Mutex    // 用以保护 pool 里面的相关资源的安全
	once         sync.Once     // 标记释放动作只能调用一次
	workerCache  sync.Pool     // 缓存
	cond         *sync.Cond
	PanicHandler func(err any)
}

// Submit 获取池中 worker 并执行任务
func (p *Pool) Submit(task func()) error {
	if len(p.release) > 0 {
		// pool 已释放
		return ErrorHasClose
	}
	w := p.GetWorker()
	w.pool.incRunning()
	w.task <- task
	return nil
}

func (p *Pool) GetWorker() *Worker {
	// 1. Get the pool
	idleWorkers := p.workers
	n := len(idleWorkers)
	if n > 0 {
		// 2. 存在空闲 worker 直接使用
		p.lock.Lock()
		w := idleWorkers[n-1]
		idleWorkers[n-1] = nil
		p.workers = idleWorkers[:n-1]
		p.lock.Unlock()
		return w
	}
	// 3. 若没有空闲 worker，判断是否可以创建新 worker
	if p.running < p.cap {
		// 4. 若 cap 大于所有正在运行 workers 长度，表示可以创建
		var w *Worker
		c := p.workerCache.Get()
		if c == nil {
			w = &Worker{
				pool: p,
				task: make(chan func(), 1),
			}
		} else {
			w = c.(*Worker)
		}

		w.run()
		return w
	}
	// 5. 若 cap 已经等于所有正在运行 workers 长度，则只能堵塞等待
	return p.waitIdleWorker()
}

func (p *Pool) waitIdleWorker() *Worker {
	p.lock.Lock()
	p.cond.Wait()
	idleWorkers := p.workers
	n := len(idleWorkers)
	if n == 0 {
		p.lock.Unlock()
		return p.waitIdleWorker()
	}
	w := idleWorkers[n-1]
	idleWorkers[n-1] = nil
	p.workers = idleWorkers[:n-1]
	p.lock.Unlock()
	return w
}

func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) PutWorker(w *Worker) {
	w.lastTime = time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, w)
	p.cond.Signal()
	p.lock.Unlock()
}

func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

// Release 销毁
func (p *Pool) Release() {
	p.once.Do(func() {
		p.lock.Lock()
		workers := p.workers
		for i, w := range workers {
			w.task = nil
			w.pool = nil
			workers[i] = nil
		}
		p.lock.Unlock()
		p.release <- sig{}
	})
}

func (p *Pool) IsClosed() bool {
	return len(p.release) > 0
}

func (p *Pool) Restart() bool {
	if !p.IsClosed() {
		return true
	}
	_ = <-p.release
	go p.expireWorker()
	return true
}

// 程监听已有 worker 是否需要被释放
func (p *Pool) expireWorker() {
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		if p.IsClosed() {
			break
		}
		p.lock.Lock()
		// 循环空闲 workers ，若空闲时间大于预设 expire 则释放掉
		idleWorkers := p.workers
		n := len(idleWorkers)
		if n > 0 {
			for i, w := range idleWorkers {
				if time.Now().Sub(w.lastTime) <= p.expire {
					break
				}
				n = i
				w.task = nil
			}
			if n >= len(idleWorkers)-1 {
				p.workers = idleWorkers[:0]
			} else {
				p.workers = idleWorkers[n+1:]
			}
		}
		p.lock.Unlock()
	}
}

func NewPool(cap int) (*Pool, error) {
	return NewTimePool(cap, DefaultExpire)
}

func NewTimePool(cap int, expire int) (*Pool, error) {
	if cap <= 0 {
		return nil, ErrorInValidCap
	}
	if expire <= 0 {
		return nil, ErrorInValidExpire
	}
	p := &Pool{
		cap:     int32(cap),
		expire:  time.Duration(expire) * time.Second,
		release: make(chan sig, 1),
	}

	p.workerCache.New = func() any {
		return &Worker{
			pool: p,
			task: make(chan func(), 1),
		}
	}

	p.cond = sync.NewCond(&p.lock)

	go p.expireWorker()

	return p, nil
}
