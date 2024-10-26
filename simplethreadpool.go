package simplethreadpool

import (
	"context"

	"github.com/golang-collections/collections/queue"
)

type threadTask struct {
	process func(ctx context.Context)
}

func (t *threadTask) Run(releaseFunc func(), ctx context.Context) {
	go func() {
		// TODO: add waiting group
		t.process(ctx)
		releaseFunc()
	}()
}

func newTask(process func(ctx context.Context)) *threadTask {
	return &threadTask{
		process: process,
	}
}

type ThreadPool struct {
	activeTasks  []*threadTask
	waitingTasks *queue.Queue
	maxThreads   int
}

func NewThreadPool(maxThreads int) *ThreadPool {
	return &ThreadPool{
		activeTasks:  make([]*threadTask, 0),
		waitingTasks: queue.New(),
		maxThreads:   maxThreads,
	}
}

func (t *ThreadPool) Put(process func(ctx context.Context)) {
	if len(t.activeTasks) < t.maxThreads {
		lastId := len(t.activeTasks)
		task := newTask(process)
		t.activeTasks = append(t.activeTasks, task)
		task.Run(t.releaseFunc(lastId), context.Background())
	} else {
		for i, item := range t.activeTasks {
			if item == nil {
				t.activeTasks[i] = newTask(process)
				t.activeTasks[i].Run(t.releaseFunc(i), context.Background())
				return
			}
		}
		t.waitingTasks.Enqueue(process)
	}
}

func (t *ThreadPool) releaseFunc(index int) func() {
	return func() {
		t.activeTasks[index] = nil
		t.next(index)
	}
}

func (t *ThreadPool) next(index int) {
	if process := t.waitingTasks.Dequeue(); process != nil {
		t.activeTasks[index] = newTask(process.(func(ctx context.Context)))
		t.activeTasks[index].Run(t.releaseFunc(index), context.Background())
	}
}

func (t *ThreadPool) Wait() {
	for func() bool {
		for _, item := range t.activeTasks {
			if item != nil {
				return false
			}
		}
		return true
	}() || t.waitingTasks.Len() > 0 {
	}
}
