package rw

import (
	"context"
	"correlator/config"
	"correlator/logger"
	kafka "correlator/rw/kafka"
	"reflect"
)

// Struct for defining agents and their struct, if you want to include new reader you need to write a struct and methods thats defined in ReaderAgent
type Readers struct {
	Kafka *kafka.KafkaR
}

// Initialising objects
var readers = &Readers{
	Kafka: kafka.NewReader(),
}

// Methods for Readers agents
type ReaderAgent interface {
	LoadConfig(interface{}) error // because we don't know how many configs are and their format
	Connect() error               // Inint connection or open file
	Resume() error                // function to recreate resume work after errors, recommended to sleep for some time
	Disconnect()                  // close connecntion, file, etc
	Read() ([][]byte, error)      // Reads chunk of data from source
}

func StartReader(ctx context.Context, r ReaderAgent, cfg interface{}, gh_chn chan []byte) {
	err := r.LoadConfig(cfg)
	if err != nil {
		logger.ErrorLogger.Printf("Error loading reader agent config: %s\n", err.Error())
		return
	}
	if err = r.Connect(); err != nil {
		logger.ErrorLogger.Printf("Error starting reader agent: %s\n", err.Error())
		return
	}
	defer r.Disconnect()
	for {
		data, err := r.Read()
		if err != nil {
			logger.ErrorLogger.Printf("Reader error: %s\n", err.Error())
			if err = r.Resume(); err != nil {
				logger.ErrorLogger.Printf("Error resuming reader agent, quiting: %s\n", err.Error())
			}
			continue
		}
		for _, d := range data {
			gh_chn <- d
		}
	}
}

func InitReaders(ctx context.Context, cfg *config.ReaderConfig, gh_chn chan []byte) error {

	vCfg := reflect.ValueOf(*cfg)
	tCfg := reflect.TypeOf(*cfg)

	vReaders := reflect.ValueOf(readers).Elem()

	for i := 0; i < tCfg.NumField(); i++ {
		cfgField := vCfg.Field(i)
		cfgFieldName := tCfg.Field(i).Name

		readerField := vReaders.FieldByName(cfgFieldName)
		if !readerField.IsValid() || readerField.IsNil() {
			continue // if no such field in Readers or = nil
		}

		reader := readerField.Interface()
		readerAgent, ok := reader.(ReaderAgent)
		if !ok {
			continue // If not reader agent
		}

		// Starting in goroutine
		go StartReader(ctx, readerAgent, cfgField.Interface(), gh_chn)
		logger.InfoLogger.Printf("Reader agent %s started", cfgFieldName)
	}

	return nil
}
