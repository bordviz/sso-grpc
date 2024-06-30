package suite

import (
	"grpc/internal/config"
	"net"
	"strconv"
	"testing"

	ssov1 "github.com/bordviz/sso-protos/gen/go/sso"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

const gRPCHost = "localhost"

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadWithPath("../config/local_tests.yaml")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	cc, err := grpc.NewClient(
		gRPCAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("gRPC server connection error: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(cc),
	}
}

func gRPCAddress(cfg *config.Config) string {
	return net.JoinHostPort(gRPCHost, strconv.Itoa(cfg.GRPC.Port))
}
