package common

import (
	"log"
	"os"
)

var logger *log.Logger

func InitLogger() {
	logger = log.New(os.Stdout, "", log.LstdFlags)

	// logger = &log.Logger{}
	// logger.SetFlags(log.LstdFlags)
	// logger.SetOutput(os.Stdout)

	// Create a log file
	if config.LogFile != "" {
		logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			logger.Fatal(err)
		}
		defer logFile.Close()
		// Set log output to file
		logger.SetOutput(logFile)
	}
}

func GetLogger() *log.Logger {
	return logger
}
