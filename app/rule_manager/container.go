package rule_manager

import "time"

// struct that contains how much container will live
type Container struct {
	Ttl             time.Duration
	Timeout         time.Duration       // How long new alert won't be created
	TimeoutTrigger  time.Time           // time since last alert
	Trigger         time.Time           // When container created
	Selectors       map[string]string   // STICKY fields and their values
	CurrentStep     *Step               // contains list of steps (each step contains pointer on next and orherwise step)
	TriggerTimes    []time.Time         // Timestamps of alarms
	DifferentValues map[string][]string // fields and values that already were in logs
	Alert           *Alert
}

// Function to perform operations
func (c *Container) evaluateOperation(a, b interface{}, operator string, ld map[string]interface{}) bool {
	if v1, ok1 := a.(bool); ok1 {
		if v2, ok2 := b.(bool); ok2 {
			switch operator {
			case "AND":
				return v1 && v2
			case "and":
				return v1 && v2
			case "OR":
				return v1 || v2
			case "or":
				return v1 || v2
			default:
				return false
			}

		}
	}

	if v1, ok1 := a.(string); ok1 {
		if v2, ok2 := b.(string); ok2 {
			//! Container already checks if sticky field sattisfies
			if v1 == "$STICKY$" || v2 == "$STICKY$" {
				return true
			}
			if v2 == "$ANY$" {
				return isFieldExists(ld, v1)
			}
			if v2 == "$DIFF$" && isFieldExists(ld, v1) {
				val := getValueFromLog(ld, v1).(string)
				if containsString(c.DifferentValues[v1], val) {
					return false
				} else {
					c.DifferentValues[v1] = append(c.DifferentValues[v1], val)
					return true
				}
			}
			switch operator {
			case ":":
				return CheckValueFromLog(ld, v1, v2)
			case "!:":
				return !CheckValueFromLog(ld, v1, v2)
			case "CONTAINS":
				return InList(ld, v1, v2)
			case "contains":
				return InList(ld, v1, v2)
			case "->":
				return InList(ld, v1, v2)
			case "!->":
				return !InList(ld, v1, v2)
			case "==":
				return CheckPerciseValueFromLog(ld, v1, v2)
			case "=":
				return CheckPerciseValueFromLog(ld, v1, v2)
			case "!=":
				return !CheckPerciseValueFromLog(ld, v1, v2)
			default:
				return false
			}
		}
	}
	return false
}

// change pointer to a next step
func (c *Container) nextStep() {
	c.CurrentStep = c.CurrentStep.NextStep
	clear(c.TriggerTimes)
}

// adding new timestamp of trigger
func (c *Container) recordTrigger() {
	c.TriggerTimes = append(c.TriggerTimes, time.Now())
}

// checking if freq or times already exceeded
func (c *Container) isLimitExceeded(s *Step) bool {
	now := time.Now()
	validTimes := make([]time.Time, 0, len(c.TriggerTimes))

	// Saving alarms during `Per`
	if s.Per != 0 {
		for _, t := range c.TriggerTimes {
			if now.Sub(t) <= s.Per {
				validTimes = append(validTimes, t)
			}
		}
		c.TriggerTimes = validTimes
	}

	// Checking frequency if freq is above 0 (defined)
	if s.Times > 0 && len(validTimes) >= s.Times {
		return true
	}

	// Checking count if count is above 0 (defined)
	if s.Count > 0 && len(c.TriggerTimes) >= s.Count {
		return true
	}

	return false
}

// checking if container ttl is expired if ttl is 0 (not defined) than container without expiration
func (c *Container) isExpired() bool {
	if c.Ttl == 0 {
		return false
	}
	return time.Since(c.Trigger) > c.Ttl
}

func (c *Container) isTimeoutOver() bool {
	if c.Timeout == 0 {
		return true
	}
	return time.Since(c.TimeoutTrigger) > c.Timeout
}

// Function which performs operations step by step and returns true if condition satisfied
func (c *Container) checkLogic(logData map[string]interface{}, s *Step) bool {
	var stack []interface{}
	var result bool

	for _, token := range s.Logic {
		if isOperand(token) { // if operand adding to stack
			stack = append(stack, token)
		} else if isOperator(token) { // if operator performing operation with 2 top values from stack
			if len(stack) < 2 {
				return false
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2] // Deleting top 2 values from stack
			result = c.evaluateOperation(a, b, token, logData)
			stack = append(stack, result)
		}
	}
	if len(stack) != 1 {
		return false // Error: no valid result
	}

	if result && (s.Count != 0 || s.Times != 1) {
		c.recordTrigger()
		if c.isLimitExceeded(s) {
			return true
		} else {
			return false
		}
	}
	return result
}
