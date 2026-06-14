package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"coupon-service/config"
	"coupon-service/internal/cron"
	"coupon-service/internal/handler"
	"coupon-service/internal/middleware"
	"coupon-service/internal/pkg/logger"
	"coupon-service/internal/repository"
	"coupon-service/internal/service"
)

func main() {
	cfg := config.Load()

	if err := logger.Init("./logs"); err != nil {
		log.Fatalf("init logger failed: %v", err)
	}

	logger.Info("starting coupon service, db=%s, port=%s, ticker=%dmin",
		cfg.DBPath, cfg.Port, cfg.TickerMinutes)

	repo, err := repository.New(cfg.DBPath)
	if err != nil {
		logger.Error("init repository failed: %v", err)
		log.Fatalf("init repository failed: %v", err)
	}
	logger.Info("repository initialized")

	templateService := service.NewTemplateService(repo)
	claimService := service.NewClaimService(repo)
	useService := service.NewUseService(repo)
	newUserService := service.NewNewUserService(repo, claimService)

	h := handler.NewHandler(templateService, claimService, useService, newUserService)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	h.RegisterRoutes(r)

	expireCron := cron.NewExpireCron(repo, cfg.TickerMinutes)
	expireCron.Start()

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		logger.Info("http server starting on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutdown signal received")

	expireCron.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error: %v", err)
	}
	logger.Info("coupon service stopped")
}
