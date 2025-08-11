package postgre

import (
	"correlator/config"
	"correlator/db"
	"encoding/json"
	"fmt"
	"time"
)

type PostgreW struct {
	Conf   *config.PostgreWriterConfig
	IncChn chan []byte
}

func NewWriter() *PostgreW {
	return &PostgreW{
		IncChn: make(chan []byte),
	}
}

func (w PostgreW) GetChannel() chan []byte {
	return w.IncChn
}

func (w *PostgreW) LoadConfig(cfg interface{}) error {
	var ok bool

	w.Conf, ok = cfg.(*config.PostgreWriterConfig)
	if !ok {
		return fmt.Errorf("invalid config type, expected PostgreWriterConfig")
	}

	if !w.Conf.Enable {
		return fmt.Errorf("postgre writer is disabled in conf file")
	}

	return nil
}

func (w PostgreW) Connect() error {
	if err := db.DB.AutoMigrate(db.Alert{}); err != nil {
		return fmt.Errorf("migration error: %v", err)
	}
	return nil
}

func (w PostgreW) Resume() error {
	return nil
}

func (w PostgreW) Disconnect() {
}

func (w PostgreW) Write(message []byte) error {
	var data map[string]interface{}
	var timestamp time.Time
	if err := json.Unmarshal(message, &data); err != nil {
		return fmt.Errorf("failed to parse alert before saving in Postgre: %w", err)
	}

	v, ok := data["timestamp"].(string)
	if !ok || v == "" {
		return fmt.Errorf("failed to read timestamp")
	} else {
		layout := time.RFC3339Nano
		if parsedTime, err := time.Parse(layout, v); err != nil {
			return fmt.Errorf("failed to parse timestamp")
		} else {
			timestamp = parsedTime
		}
	}

	alert := db.Alert{Text: message, Timestamp: timestamp}

	if err := db.DB.Create(&alert).Error; err != nil {
		return fmt.Errorf("error occured while saving alert rule: %w", err)
	}

	return nil
}
