package pool

import (
	"errors"
	"sync"
	"time"
)

const DefaultExpire = 3

var (
	ErrorInValidCap    = errors.New("pool cap cannot <= 0")
	ErrorInValidExpire = errors.New("pool expire cannot <= 0")
)

type sig struct{}

type Pool struct {
	cap     int32         // 容量
	running int32         // 正在运行中的 worker 数量
	workers []*Worker     // 若干空闲 worker
	expire  time.Duration // 过期时间，规定 worker 的空闲时间超过此值就回收掉
	release chan sig      // 释放资源信号，接收到它表示pool不再使用
	lock    sync.Mutex    // 用以保护 pool 里面的相关资源的安全
	once    sync.Once     // 标记释放动作只能调用一次
}

// Submit 获取池中 worker 并执行任务
func (p *Pool) Submit(task func()) error {
	w := p.GetWorker()
	w.task <- task
	return nil
}

func (p *Pool) GetWorker() *Worker {
	// 1. Get the pool
	idleWorkers := p.workers
	n := len(idleWorkers)
	if n == 0 {
		// 3. 若没有空闲 worker，判断是否可以创建新 worker
		if p.running < p.cap {
			// 4. 若 cap 大于所有正在运行 workers 长度，表示可以创建
			w := &Worker{
				pool: p,
				task: make(chan func(), 1),
			}
			w.run()
			return w
		}
		// 5. 若 cap 已经等于所有正在运行 workers 长度，则只能堵塞等待
	}
	// 2. 存在空闲 worker 直接使用
	w := idleWorkers[n-1]
	idleWorkers[n-1] = nil
	p.workers = idleWorkers[:n-1]
	return w
}

func (p *Pool) inRunning() {

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
	return &Pool{
		cap:     int32(cap),
		expire:  time.Duration(expire) * time.Second,
		release: make(chan sig, 1),
	}, nil
}
