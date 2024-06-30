package grpcapp

import (
	"fmt"
	authGRPC "grpc/internal/grpc/auth"
	"grpc/internal/lib/logger/sl"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, port int, authService authGRPC.Auth) *App {
	gRPCServer := grpc.NewServer()

	authGRPC.Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	const op = "app.grpcapp.MustRun"

	if err := a.Run(); err != nil {
		a.log.Error("failed to run gRPC server", sl.OpErr(op, err))
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app.grpcapp.Run"

	a.log.Info("stating gRPC server")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		a.log.Error("failed to listen gRPC server", sl.OpErr(op, err))
		return err
	}

	a.log.Info("gRPC server is running", slog.String("addres", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		a.log.Error("failed to start gRPC server", sl.OpErr(op, err))
		return err
	}

	return nil
}

func (a *App) Stop() {
	const op = "app.grpcapp.Stop"

	a.log.Info("stopping gRPC server", slog.String("op", op), slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
