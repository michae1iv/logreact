package rule_manager

import (
	"context"
	"correlator/logger"
	"time"
)

type Frame struct {
	WorkersPointer map[string]*Worker          // Pool of workers
	Ukeys          string                      // unique keys, if slice is given separator is ,
	LogChan        chan map[string]interface{} // Channel for logs
	RuleChan       chan *Rule                  // Channel for rules, if no Rule with such Rule_Name then creating new one, else deleting rule from map
}

// Main function of frame
func (f *Frame) StartFrame(ctx context.Context) {
	logger.InfoLogger.Printf("Frame [%s] started", f.Ukeys)
	for {
		select {
		case <-ctx.Done():
			logger.InfoLogger.Println("Stop signal recieved, stopping handler")
			return
		case r, ok := <-f.RuleChan: // getting parsed log
			if !ok {
				logger.InfoLogger.Println("Rule channel closed, stopping handler")
				return
			}
			if err := f.manageRule(r); err != nil {
				logger.ErrorLogger.Printf("Error adding rule %s to frame: %s\n", r.Rule_Name, f.Ukeys)
			}
		case log, ok := <-f.LogChan: // getting parsed log
			if !ok {
				logger.InfoLogger.Println("Log channel closed, stopping handler")
				return
			}
			var logs []map[string]interface{}
			logs = append(logs, log)
			logs = append(logs, drainChannel(f.LogChan)...)
			f.toWorkers(logs)
		default:
			f.doBackgroundWork()
		}
	}
}

func (f *Frame) manageRule(r *Rule) error {
	// if such rule already exists, then delete it
	if f.WorkersPointer[r.Rule_Name] != nil {
		close(f.WorkersPointer[r.Rule_Name].AlertChan)
		close(f.WorkersPointer[r.Rule_Name].LogChan)
		delete(f.WorkersPointer, r.Rule_Name)
		return nil
	}
	// if worker doesn't exist, create new one
	worker := &Worker{
		R:         r,
		AlertChan: make(chan []byte),
		LogChan:   make(chan []map[string]interface{}),
	}
	go worker.InitWorker()
	f.WorkersPointer[r.Rule_Name] = worker

	return nil
}

// Sending logs to workers
func (f Frame) toWorkers(log []map[string]interface{}) {
	for _, worker := range f.WorkersPointer {
		worker.LogChan <- log
	}
}

// Waiting 10 milliseconds if no work to do to not block CPU if no work to do
func (f Frame) doBackgroundWork() {
	time.Sleep(10 * time.Millisecond)
}
