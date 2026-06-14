package cron

import (
	"time"

	"coupon-service/internal/pkg/logger"
	"coupon-service/internal/repository"
)

type ExpireCron struct {
	repo        *repository.Repository
	ticker      *time.Ticker
	stopChan    chan struct{}
	intervalMin int
}

func NewExpireCron(repo *repository.Repository, intervalMin int) *ExpireCron {
	if intervalMin <= 0 {
		intervalMin = 1
	}
	return &ExpireCron{
		repo:        repo,
		intervalMin: intervalMin,
		stopChan:    make(chan struct{}),
	}
}

func (c *ExpireCron) Start() {
	c.ticker = time.NewTicker(time.Duration(c.intervalMin) * time.Minute)
	logger.Info("expire cron started, interval=%d minutes", c.intervalMin)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("expire cron panic recovered: %v", r)
			}
		}()
		for {
			select {
			case <-c.ticker.C:
				c.run()
			case <-c.stopChan:
				logger.Info("expire cron stopped")
				return
			}
		}
	}()
}

func (c *ExpireCron) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
	close(c.stopChan)
}

func (c *ExpireCron) run() {
	now := logger.Now()
	logger.Info("expire cron running at %v", now)
	cnt, err := c.repo.BatchMarkExpired(now)
	if err != nil {
		logger.Error("batch mark expired failed: %v", err)
		return
	}
	logger.Info("expire cron finished, marked %d records as expired", cnt)
}
