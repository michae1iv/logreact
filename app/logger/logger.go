package logger

import (
	"correlator/config"
	"fmt"
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

// * Init loggers to write success and error logs in files
func InitLoggers(cfg *config.LoggingConfig) error {
	// creating dir for logs if it doesn't exists
	log_path := cfg.LogPath
	err := os.MkdirAll(log_path, os.ModePerm)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	// Open log files
	infoFile, err := os.OpenFile(log_path+"/info.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	errorFile, err := os.OpenFile(log_path+"/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Init loggers
	InfoLogger = log.New(infoFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}
