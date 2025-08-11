package rule_manager

import (
	"correlator/logger"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Struct for rule
type Rule struct {
	UKeyValue    string                 `json:"ukey,omitempty"`
	Rule_Name    string                 `json:"rule"`
	Params       map[string]interface{} `json:"params"`
	Condition    map[string]interface{} `json:"condition"`
	Alert        map[string]interface{} `json:"alert"`
	StickyFields []string               `json:"-"`
	StepHead     *Step                  `json:"-"`
}

// Struct that defines steps in rules, there is minimum one step in each rule
type Step struct {
	Logic    []string      // Reverse polish notation of logic
	Times    int           // defines amount of times for freq
	Per      time.Duration // defines time period for freq
	Count    int           // Count for amount of alarms
	NextStep *Step         // Pointer to a next step of the rule
}

// Takes Rule and unmarshall it to Rule struct
func (r *Rule) ParseRule(input interface{}) error {
	var data []byte

	switch v := input.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("unsupported input type %T", input)
	}

	if err := json.Unmarshal(data, r); err != nil {
		logger.ErrorLogger.Printf("Error parsing rule: %s\n", err)
		return err
	}

	return nil
}

// recursive function which inits steps of the rule, returns list of steps, each step contains pointer on next steps
func (r *Rule) ConvertMapToSteps(cond map[string]interface{}) (*Step, error) {
	// logic
	var err error
	step := &Step{}

	logic, ok := cond["logic"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid logic field")
	}
	// Converting to reverse polish notation
	step.Logic, err = ParseCondToRPN(logic)
	if err != nil {
		return nil, err
	}
	// looking for $STICKY$ fields
	for i, l := range step.Logic {
		if l == "$STICKY$" {
			if str := strings.Join(r.StickyFields, ","); !strings.Contains(str, step.Logic[i-1]) {
				r.StickyFields = append(r.StickyFields, step.Logic[i-1])
			}
		}
	}

	// freq (not required)
	if freqRaw, ok := cond["freq"].(string); ok && freqRaw != "" {
		times, per, err := ParseFreq(freqRaw)
		if err != nil {
			return nil, err
		}
		step.Times = times
		step.Per = per
	} else {
		// default:
		step.Times = 1
		step.Per = 0
	}

	// count (not required)
	if count, ok := cond["times"].(float64); ok && count > 0 {
		step.Count = int(count)
	} else {
		step.Count = 0
	}

	// then
	if thenRaw, ok := cond["then"].(map[string]interface{}); ok {
		thenStep, err := r.ConvertMapToSteps(thenRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid 'then': %v", err)
		}
		if thenStep != nil {
			step.NextStep = thenStep
		}
	}
	r.StepHead = step

	return step, nil
}

func (r Rule) ConvertToAlert() (*Alert, error) {

	raw := r.Alert
	alert := &Alert{
		Name:      r.Rule_Name,
		Fields:    make(map[string]string),
		AddFields: make(map[string]string),
	}

	// splitting fields by ,
	if fieldsStr, ok := raw["fields"].(string); ok {
		fields := strings.Split(fieldsStr, ",")
		for _, field := range fields {
			clean := strings.TrimSpace(field)
			alert.Fields[clean] = "" // inserting values later with inheritFields
		}
	}

	// selecting fields at "addfields"
	if addFields, ok := raw["addfields"].(map[string]interface{}); ok {
		for k, v := range addFields {
			if strVal, ok := v.(string); ok {
				alert.AddFields[k] = strVal
			}
		}
	}

	if sev_level, ok := r.Params["sev_level"].(float64); ok && sev_level >= 0 {
		alert.SeverityLevel = int(sev_level)
	} else {
		alert.SeverityLevel = 1
	}

	if no_alert, ok := r.Params["no_alert"].(bool); ok {
		alert.NoAlert = no_alert
	} else {
		alert.NoAlert = false
	}

	return alert, nil
}

// Function to parse rule from json, either returns error or Rule's object with parsed params
func ParseRule(input string) (Rule, error) {
	rule := Rule{}
	err := json.Unmarshal([]byte(input), &rule)
	if err != nil {
		logger.ErrorLogger.Printf("Error parsing rule: %s\n", err)
		return Rule{}, err
	}

	return rule, nil
}

// ParseFreq enters freq (f.e, "10/min") and returns count in ms.
func ParseFreq(freq string) (count int, duration time.Duration, err error) {
	if freq == "0" {
		// Однократное срабатывание
		return 1, 0, nil
	}

	parts := strings.Split(freq, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid freq format: %s", freq)
	}

	count, err = strconv.Atoi(parts[0])
	if err != nil || count <= 0 {
		return 0, 0, fmt.Errorf("invalid count in freq: %s", parts[0])
	}

	switch parts[1] {
	case "sec":
		duration = time.Second
	case "min":
		duration = time.Minute
	case "hour":
		duration = time.Hour
	case "day":
		duration = 24 * time.Hour
	case "week":
		duration = 7 * 24 * time.Hour
	case "month":
		duration = 30 * 24 * time.Hour // month is calculated as 30 days
	default:
		return 0, 0, fmt.Errorf("invalid time unit in freq: %s", parts[1])
	}

	return count, duration, nil
}

// parses ttl and returns duration
func (r Rule) ParseTTL() (time.Duration, error) {
	var (
		ttl      string
		duration time.Duration
		err      error
	)

	if value, ok := r.Params["ttl"].(string); !ok {
		return 0, fmt.Errorf("invalid ttl format")
	} else {
		ttl = value
	}
	if duration, err = time.ParseDuration(ttl); err != nil {
		return 0, fmt.Errorf("unsuported ttl format")
	}

	return duration, nil
}

// parses timeout and returns duration
func (r Rule) ParseTimeout() (time.Duration, error) {
	var (
		timeout  string
		duration time.Duration
		err      error
	)

	if value, ok := r.Params["timeout"].(string); !ok {
		return 0, nil
	} else {
		timeout = value
	}
	if timeout == "" { // if not defined
		return 0, nil
	}
	if duration, err = time.ParseDuration(timeout); err != nil {
		return 0, fmt.Errorf("unsuported timeout format")
	}

	return duration, nil
}
