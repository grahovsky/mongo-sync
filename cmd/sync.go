package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"mongo-sync/internal/config"
	"mongo-sync/internal/db"
	"mongo-sync/internal/logger"
)

var checkTime time.Time

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer func() {
		db.CleanClients()
		cancel()
	}()

	config.Init()
	logger.Set(logger.Options{Level: config.Settings.Log.Level, AddSource: false})

	slog.Info(version())

	// Set the start time of the check synchronization
	var err error
	switch config.Settings.Common.From {
	case "":
		checkTime = time.Now().Add(-time.Hour * 2)
	default:
		checkTime, err = time.Parse(time.RFC3339, config.Settings.Common.From)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}
	slog.Info(checkTime.String())
	slog.Info("start sync")

	// ticker scan
	go func() {
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()

		firstTick := true

		for {
			select {
			case <-ticker.C:
				if firstTick {
					ticker.Reset(config.Settings.Common.Timeout)
					firstTick = false
				}
				slog.Debug("start sync iteration")
				syncIterate()
				slog.Debug("end sync iteration")
			case <-ctx.Done():
				slog.Info("stopped sync..")
				return
			}
		}
	}()

	<-ctx.Done()
}

func syncIterate() {
	// time to filter data for processing
	requestTime := checkTime.Add(-2 * config.Settings.Common.Timeout)
	slog.Debug(fmt.Sprintf("request time: %s", requestTime))

	// Update the check time to the beginning of the request
	checkTime = time.Now()

	// create channel
	docsChan := make(chan db.Doc, 5000)

	var wg sync.WaitGroup
	go db.RequestNewDocs(docsChan, requestTime)
	wg.Add(config.Settings.Common.NumWorkers)

	for w := 0; w < config.Settings.Common.NumWorkers; w++ {
		go db.Worker(docsChan, &wg)
	}

	wg.Wait()
}
