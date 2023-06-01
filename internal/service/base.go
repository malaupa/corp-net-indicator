package service

import (
	"time"

	"com.telekom-mms.corp-net-indicator/internal/logger"
)

const pollInterval = 5

type client[T interface{}] interface {
	Ping() error
	Subscribe() (chan T, error)
}

func waitAndSubscribe[T interface{}](client client[T], statusChan chan T, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
		}
		if client.Ping() == nil {
			break
		}
		logger.Verbosef("Wait %d seconds for service to come up...", pollInterval)
		time.Sleep(time.Second * pollInterval)
	}
	c, err := client.Subscribe()
	if err != nil {
		panic(err)
	}
	for status := range c {
		select {
		case statusChan <- status:
		case <-done:
			return
		}
	}
}