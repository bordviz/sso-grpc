package main

import (
	"grpc/internal/app"
	"grpc/internal/config"
	"grpc/internal/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	log := logger.New(cfg.Env)
	log.Debug("debug messages enabled")
	log.Info("info messages enabled")
	log.Warn("warn messages enabled")
	log.Error("error messages enabled")

	application := app.New(log, cfg)

	stop := make(chan os.Signal, 1)
	go application.GRPCSrv.MustRun()
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	stopSignal := <-stop
	log.Info("stoppping application", slog.String("signal", stopSignal.String()))

	application.GRPCSrv.Stop()

	log.Info("application stopped")
}
