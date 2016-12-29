package piper

import "sync"

type Job interface{}
type Result interface{}

const DoneNil = 0
const DoneReq = 1
const DoneAck = 2

type Msg struct {
	done int
	job  Job
}

type Handler func(w *Worker, job Job) Result
type Callback func(w *Worker, job Job, result Result)

type Worker struct {
	wgWorker sync.WaitGroup
	wgJob    sync.WaitGroup
	handler  Handler
	cb       Callback
	queue    chan *Msg
	prev     *Worker
	next     *Worker
}

func NewWorker(fanoutSize int, handler Handler) *Worker {
	return &Worker{
		sync.WaitGroup{},
		sync.WaitGroup{},
		handler,
		nil,
		make(chan *Msg, fanoutSize),
		nil,
		nil}
}

func (w *Worker) Chain(next *Worker) *Worker {
	w.next = next
	next.prev = w
	return next
}

func (w *Worker) GoWait() {
	w.Go()
	w.Wait()
}

func (w *Worker) Go() {
	w.wgWorker.Add(1)
	go func() {
		for msg := range w.queue {
			if msg.done == DoneNil {
				w.work(msg.job)
			} else {
				switch msg.done {
				case DoneReq:
					w.chainDoneReq()
				case DoneAck:
					w.chainClosing()
				default:
					break
				}

			}
		}
		w.wgWorker.Done()
	}()

	if w.next != nil {
		w.next.Go()
	}
}

func (w *Worker) Wait() {
	w.wgWorker.Wait()
}

func (w *Worker) Next() *Worker {
	return w.next
}

func (w *Worker) Queue(job Job) {
	w.qMsg(0, job)
}

func (w *Worker) Done() {
	w.qMsg(DoneReq, nil)
}

func (w *Worker) Get(cb Callback) *Worker {
	w.cb = cb
	return w
}

func (w *Worker) work(job Job) {
	w.wgJob.Add(1)
	go func(job Job) {
		r := w.handler(w, job)
		if w.cb != nil {
			w.cb(w, job, r)
		}
		if w.next != nil {
			w.next.Queue(r)
		}
		w.wgJob.Done()
	}(job)
}

func (w *Worker) chainDoneReq() {
	w.wgJob.Wait()
	if w.next != nil {
		w.next.Done()
	} else {
		w.qMsg(DoneAck, nil)
	}
}

func (w *Worker) chainClosing() {
	close(w.queue)
	if w.prev != nil {
		w.prev.qMsg(DoneAck, nil)
	}
}

func (w *Worker) qMsg(done int, job Job) {
	w.queue <- &Msg{done, job}
}
