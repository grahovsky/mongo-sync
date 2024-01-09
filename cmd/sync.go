package main

import (
	"context"
	"log"
	"log/slog"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"mongo-sync/internal/config"
	"mongo-sync/internal/db"
	"mongo-sync/internal/logger"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer cancel()

	config.Init()
	logger.Set(logger.Options{Level: config.Settings.Log.Level, AddSource: false})

	slog.Info(version())

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
				slog.Info("start sync iteration")

				slog.Info("end scan iteration")
			case <-ctx.Done():
				slog.Info("stopped scan..")
				db.CleanClients()
				return
			}
		}
	}()

	// Set the start time of the check synchronization
	var checkTime time.Time
	var err error

	switch config.Settings.Common.From {
	case "":
		checkTime = time.Now().Add(-time.Hour * 2)
	default:
		checkTime, err = time.Parse(time.RFC3339, config.Settings.Common.From)
		if err != nil {
			slog.Error(err.Error())
			log.Fatal(err)
		}
	}
	slog.Info(checkTime.String())

	for {
		// time to filter data for processing
		requestTime := checkTime.Add(-2 * config.Settings.Common.Timeout)

		// Update the check time to the beginning of the request
		checkTime = time.Now()

		// create channel
		docsChan := make(chan db.Doc, 500)

		var wg sync.WaitGroup
		wg.Add(1)
		go db.Worker(docsChan, &wg)

		db.RequestNewDocs(docsChan, requestTime)

		if len(docsChan) > 0 {
			wg.Add(config.Settings.Common.NumWorkers)

			// create workers
			for w := 1; w <= config.Settings.Common.NumWorkers; w++ {
				go db.Worker(docsChan, &wg)
			}
		}

		close(docsChan)
		wg.Wait()

		slog.Info("End sync iteration")

		duration := time.Since(checkTime)
		if duration < config.Settings.Common.Timeout {
			time.Sleep(config.Settings.Common.Timeout - duration)
		}
	}
}
