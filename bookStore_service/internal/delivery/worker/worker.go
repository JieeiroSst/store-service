package worker

import (
	"fmt"

	"github.com/robfig/cron/v3"
)

// RunJob chạy một job theo lịch trình định sẵn
func RunJob(c *cron.Cron, schedule string, handler func()) {
	c.AddFunc(schedule, handler)
	c.Start()
}

type Worker struct {
}

func NewWorker() *Worker {
	return &Worker{}
}

func (w *Worker) RunWorker() {
	RunJob(cron.New(), "*/5 * * * *", func() {
		fmt.Println("Chạy job 1")
	})
	RunJob(cron.New(), "*/10 * * * *", func() {
		fmt.Println("Chạy job 2")
	})
	
}
