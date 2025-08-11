package rule_manager

import (
	"correlator/logger"
	"time"
)

type Worker struct {
	R           *Rule
	LogChan     chan []map[string]interface{}
	AlertChan   chan []byte
	ContPointer []*Container
}

func (w *Worker) InitWorker() {
	if err := w.defineStepsAndAlert(); err != nil {
		logger.ErrorLogger.Printf("Failed to parse conditions rule [%s]\n", w.R.Rule_Name)
		return
	}

	logger.InfoLogger.Printf("Worker [%s] started", w.R.Rule_Name)

	var logs []map[string]interface{}
	var ok bool
	if len(w.R.StickyFields) > 0 {
		for {
			logs, ok = <-w.LogChan // getting slice of parsed logs
			if !ok {
				logger.InfoLogger.Printf("Log channel closed, stopping worker [%s]\n", w.R.Rule_Name)
				return
			}
			w.multiContProcess(logs)
		}
	} else {
		err := w.newContainer() // creating default container for simple rules
		if err != nil {
			logger.ErrorLogger.Printf("Stopping Worker [%s]: %v", w.R.Rule_Name, err)
		}
		for {
			logs, ok = <-w.LogChan
			if !ok {
				logger.InfoLogger.Printf("Log channel closed, stopping worker [%s]\n", w.R.Rule_Name)
				return
			}
			w.defaultProcess(logs)
		}
	}
}

// func to check conditions in single container
func (w *Worker) defaultProcess(logs []map[string]interface{}) error {
	c := w.ContPointer[0]

	for _, l := range logs {
		if c.checkLogic(l, c.CurrentStep) && c.isTimeoutOver() {
			c.Alert.inheritFields(l)
			c.Alert.createAlert()
			c.TimeoutTrigger = time.Now()
		}
	}
	return nil
}

// func to check conditions in all containers
func (w *Worker) multiContProcess(logs []map[string]interface{}) error {
	var newCP []*Container
	for _, c := range w.ContPointer {
		if c.isExpired() { // Checking if some of containers has expired
			continue
		}
		newCP = append(newCP, c)
	}
	w.ContPointer = newCP

	for _, l := range logs {
		sorted := false
		for _, c := range w.ContPointer {
			sutisfies := true
			for _, s := range w.R.StickyFields {
				val := getValueFromLog(l, s)
				strVal, ok := val.(string)
				if !ok || c.Selectors[s] == "" || c.Selectors[s] != strVal {
					sutisfies = false
					break
				}
			}
			if sutisfies {
				sorted = true
			} else {
				continue
			}

			if c.checkLogic(l, c.CurrentStep) { // checking if condition satisfied
				if c.CurrentStep.NextStep == nil {
					c.Alert.inheritFields(l) // saving fields for alert
					if c.isTimeoutOver() {
						c.Alert.createAlert()
						c.TimeoutTrigger = time.Now()
					}
				} else {
					c.nextStep()
				}
				c.Alert.inheritFields(l) // saving fields for alert
			}
		}
		if !sorted { // If no container for such log creating new one
			if err := w.newContainerWithTTL(); err != nil {
				logger.ErrorLogger.Printf("Error creating new container for rule [%s]: %s\n", w.R.Rule_Name, err.Error())
				continue
			}
			cont := w.ContPointer[len(w.ContPointer)-1]
			for _, s := range w.R.StickyFields {
				val := getValueFromLog(l, s)
				strVal, ok := val.(string)
				if !ok {
					strVal = ""
				}
				cont.Selectors[s] = strVal
			}
			if cont.checkLogic(l, cont.CurrentStep) { // checking if condition satisfied
				if cont.CurrentStep.NextStep == nil {
					cont.Alert.inheritFields(l)
					if cont.isTimeoutOver() {
						cont.Alert.createAlert()
						cont.TimeoutTrigger = time.Now()
					}
				} else {
					cont.nextStep()
				}
				cont.Alert.inheritFields(l) // saving fields for alert
			}
		}
	}
	return nil
}

func (w *Worker) defineStepsAndAlert() (err error) {
	// saving first pointer on list of steps
	_, err = w.R.ConvertMapToSteps(w.R.Condition)
	if err != nil {
		return err
	}
	return nil
}

func (w *Worker) newContainerWithTTL() error {
	ttl, err := w.R.ParseTTL()
	if err != nil {
		return err
	}

	timeout, err := w.R.ParseTimeout()
	if err != nil {
		return err
	}

	al_ex, err := w.R.ConvertToAlert()
	if err != nil {
		return err
	}

	c := &Container{Trigger: time.Now(), CurrentStep: w.R.StepHead, Ttl: ttl, Timeout: timeout, TimeoutTrigger: time.Now().Add(-2 * timeout), Selectors: make(map[string]string), Alert: al_ex}
	w.ContPointer = append(w.ContPointer, c)
	return nil
}

func (w *Worker) newContainer() error {
	timeout, err := w.R.ParseTimeout()
	if err != nil {
		return err
	}

	al_ex, err := w.R.ConvertToAlert()
	if err != nil {
		return err
	}

	c := &Container{CurrentStep: w.R.StepHead, Timeout: timeout, TimeoutTrigger: time.Now().Add(-2 * timeout), Alert: al_ex}
	w.ContPointer = append(w.ContPointer, c)
	return nil
}
