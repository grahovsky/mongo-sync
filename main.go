package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/grahovsky/mongo-sync/common"
	"github.com/grahovsky/mongo-sync/database"
)

func main() {
	// Чтение значения переменной окружения "HOME"
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config"
	}

	// Read config file
	config, err := common.LoadConfig(configPath)
	if err != nil {
		fmt.Println("config not founded")
		os.Exit(1)
	}

	common.InitLogger()
	logger := common.GetLogger()

	// Set the start time of the check synchronization
	var checkTime time.Time
	switch config.FirstSyncDate {
	case "":
		checkTime = time.Now().Add(-time.Hour * 2)
	default:
		checkTime, err = time.Parse("2006-01-02", config.FirstSyncDate)
		if err != nil {
			logger.Fatal(err)
		}
	}
	logger.Println(checkTime.String())
	timeout := time.Second * time.Duration(config.Timeout)

	for {
		// time to filter data for processing
		requestTime := checkTime.Add(-2 * timeout)

		// Update the check time to the beginning of the request
		checkTime = time.Now()

		// create channel
		jobs := make(chan database.Doc, 500)

		var wg sync.WaitGroup
		wg.Add(1)
		go database.Worker(jobs, config, &wg)

		database.PrepareSend(*config, jobs, requestTime)

		if len(jobs) > 0 {
			wg.Add(config.NumWorkers)

			// create workers
			for w := 1; w <= config.NumWorkers; w++ {
				go database.Worker(jobs, config, &wg)
			}
		}

		close(jobs)
		wg.Wait()

		logger.Println("End sync iteration")

		duration := time.Since(checkTime)
		if duration < timeout {
			time.Sleep(timeout - duration)
		}
	}
}
