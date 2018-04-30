package runner

import (
	"os"
	"time"
	"os/signal"
	"errors"
)

// Task 任务
type Task func(int)

// Runner 执行任务
type Runner struct {
	interrupt chan os.Signal
	complete  chan error
	timeout   <-chan time.Time
	tasks     []Task
}

// ErrTimeout 超时错误
var ErrTimeout = errors.New("received timeout")
// ErrInterrupt 中断错误
var ErrInterrupt = errors.New("received interrupt")

// New 新建Runner
func New(d time.Duration) *Runner {
	return &Runner{
		interrupt: make(chan os.Signal, 1),
		complete:  make(chan error),
		timeout:   time.After(d),
	}
}

// Add 添加任务
func (r *Runner) Add(tasks ...Task) {
	r.tasks = append(r.tasks, tasks...)
}

// Start 开始执行任务
func (r *Runner) Start() error {
	signal.Notify(r.interrupt, os.Interrupt)

	go func() {
		r.complete <- r.run()
	}()
	select {
	case err := <-r.complete:
		return err
	case <-r.timeout:
		return ErrTimeout
	}
}

func (r *Runner) run() error {
	for k, task := range r.tasks {
		if r.gotInterrupt() {
			return ErrInterrupt
		}
		task(k)
	}
	return nil
}

func (r *Runner) gotInterrupt() bool {
	select {
	case <-r.interrupt:
		signal.Stop(r.interrupt)
		return true
	default:
		return false
	}
}
