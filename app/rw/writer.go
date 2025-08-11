package rw

import (
	"context"
	"correlator/config"
	"correlator/logger"
	kafka "correlator/rw/kafka"
	"correlator/rw/postgre"
	"fmt"
	"reflect"
)

var WChan chan []byte

// Struct for defining agents and their struct, if you want to include new reader you need to write a struct and methods thats defined in ReaderAgent
type Writers struct {
	Kafka   *kafka.KafkaW // There must be a channel for each writer agent
	Postgre *postgre.PostgreW
}

// Initialising new readers here
var writers = &Writers{
	Kafka:   kafka.NewWriter(),
	Postgre: postgre.NewWriter(),
}

// Methods for Writers agents
type WriterAgent interface {
	LoadConfig(interface{}) error // Loading config
	Connect() error               // Inint connection or open file
	Resume() error                // function to recreate resume work after errors, recommended to sleep for some time
	Disconnect()                  // close connecntion, file, etc
	Write([]byte) error           // Writes alert
	GetChannel() chan []byte      // returns channel
}

func StartWriter(ctx context.Context, w WriterAgent, cfg interface{}) {
	err := w.LoadConfig(cfg)
	if err != nil {
		logger.ErrorLogger.Printf("Error loading writer agent config: %s\n", err.Error())
		return
	}
	if err = w.Connect(); err != nil {
		logger.ErrorLogger.Printf("Error starting writer agent: %s\n", err.Error())
		return
	}
	defer w.Disconnect()
	for {
		data, ok := <-w.GetChannel()
		if !ok {
			logger.InfoLogger.Printf("Writer channel closed\n")
			return
		}
		if err := w.Write(data); err != nil {
			logger.ErrorLogger.Printf("Writer error: %s\n", err.Error())
			if err = w.Resume(); err != nil {
				logger.ErrorLogger.Printf("Error resuming writer agent, quiting: %s\n", err.Error())
				return
			}
		}
	}
}

func InitWriters(ctx context.Context, cfg *config.WriterConfig) error { //

	var agents []WriterAgent

	vCfg := reflect.ValueOf(*cfg)
	tCfg := reflect.TypeOf(*cfg)

	vWriters := reflect.ValueOf(writers).Elem()

	for i := 0; i < tCfg.NumField(); i++ {
		cfgField := vCfg.Field(i)
		cfgFieldName := tCfg.Field(i).Name

		writerField := vWriters.FieldByName(cfgFieldName)
		if !writerField.IsValid() || writerField.IsNil() {
			continue // if no such field in Writers or = nil
		}

		wr := writerField.Interface()
		writerAgent, ok := wr.(WriterAgent)
		if !ok {
			continue // If not writer agent
		}

		// Starting in goroutine
		agents = append(agents, writerAgent)
		go StartWriter(ctx, writerAgent, cfgField.Interface())
		logger.InfoLogger.Printf("Writer agent %s started", cfgFieldName)
	}

	for {
		alert, ok := <-WChan
		if !ok {
			logger.InfoLogger.Printf("Writer channel closed\n")
			return fmt.Errorf("writer channel closed")
		}
		for _, ag := range agents {
			ag.GetChannel() <- alert
		}
	}
}
