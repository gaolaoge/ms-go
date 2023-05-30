package pool

import "time"

type Worker struct {
	pool     *Pool
	task     chan func() // 任务
	lastTime time.Time   // 最后一次执行任务的时间，若时间过久可以判定为空余 Worker 然后回收
}

func (w *Worker) run() {
	go w.running()
}

func (w *Worker) running() {
	for f := range w.task {
		if f == nil {
			return
		}
		f()
		// 任务执行完成，偿还 worker
		w.pool.PutWorker(w)
		w.pool.decRunning()
	}
}
