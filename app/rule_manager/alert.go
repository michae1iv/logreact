package rule_manager

import (
	"correlator/logger"
	"correlator/regulars"
	"correlator/rw"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Alert struct {
	NoAlert       bool              `json:"-"`
	Name          string            `json:"rule,omitempty"`
	SeverityLevel int               `json:"sev_level,omitempty"`
	Fields        map[string]string `json:"inherited_fields,omitempty"`
	AddFields     map[string]string `json:"added_fields,omitempty"`
	Time          time.Time         `json:"timestamp"`
}

func (a *Alert) createAlert() {
	a.Time = time.Now().UTC()
	for k, v := range a.AddFields {
		a.AddFields[k] = a.formatTemplate(v)
	}
	AlertByte, err := json.Marshal(a)
	if err != nil {
		logger.ErrorLogger.Printf("Alert error, failed to convert to JSON: %s\n", err.Error())
		return
	}
	if a.NoAlert {
		logger.InfoLogger.Printf("%s", string(AlertByte))
		return
	}
	rw.WChan <- AlertByte
	println("here")
}

func (a *Alert) inheritFields(log map[string]interface{}) {
	for k, v := range a.Fields {
		if log[k] != nil {
			if val, ok := log[k].(string); ok && val != v {
				a.Fields[k] = val
			}
		}
	}

}

func (a *Alert) formatTemplate(template string) string {

	result := regulars.Message_reg.ReplaceAllStringFunc(template, func(m string) string {
		key := strings.Trim(m, "%")
		if a.Fields[key] != "" {
			return fmt.Sprintf("%v", a.Fields[key])
		}
		return m
	})

	return result
}
