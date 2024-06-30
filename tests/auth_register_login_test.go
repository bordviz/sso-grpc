package tests

import (
	"fmt"
	"grpc/internal/lib/jwt"
	"grpc/tests/suite"
	"testing"
	"time"

	ssov1 "github.com/bordviz/sso-protos/gen/go/sso"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	appID            = 1
	appSecret        = "iYDFTuUHLIKyhads879fasyodfIRTUYDF"
	appRefreshSecret = "fgsuyTERTCV7812RDCfasdTYFK78ghjlBuygsd78gliueBN"

	passDefaultLength = 10
	userCount         = 10
)

var (
	ErrEmptyPass        = status.Error(codes.InvalidArgument, "empty password")
	ErrEmptyEmail       = status.Error(codes.InvalidArgument, "empty email")
	ErrEmptyName        = status.Error(codes.InvalidArgument, "empty name")
	ErrAppID            = status.Error(codes.InvalidArgument, "empty app id")
	ErrInvPassOrEmail   = status.Error(codes.InvalidArgument, "invalid password or email")
	ErrUserAlreadyExist = status.Error(codes.AlreadyExists, "user already registered")
)

func TestAuthRegisterLogin(t *testing.T) {
	ctx, st := suite.New(t)

	fakeUsers := generateFakeUsers(userCount)
	for idx, user := range fakeUsers {
		fakeUser := user
		t.Run(fmt.Sprintf("test attempt: %d", idx), func(t *testing.T) {
			t.Parallel()

			registerResp, err := st.AuthClient.Register(ctx, fakeUser)
			require.NoError(t, err)
			assert.NotEmpty(t, registerResp.GetUserId())

			loginReq := &ssov1.LoginRequest{
				Email:    fakeUser.Email,
				Password: fakeUser.Password,
				AppId:    appID,
			}

			loginResp, err := st.AuthClient.Login(ctx, loginReq)

			loginTime := time.Now()

			require.NoError(t, err)
			require.NotEmpty(t, loginResp.GetAccessToken())
			require.NotEmpty(t, loginResp.GetRefreshToken())

			token := loginResp.GetAccessToken()
			refreshToken := loginResp.GetRefreshToken()
			data, err := jwt.DecodeToken(token, appSecret)
			require.NoError(t, err)
			refreshData, err := jwt.DecodeToken(refreshToken, appRefreshSecret)
			require.NoError(t, err)

			assert.Equal(t, registerResp.GetUserId(), data.UserID)
			assert.Equal(t, registerResp.GetUserId(), refreshData.UserID)
			assert.Equal(t, appID, data.AppID)
			assert.Equal(t, appID, refreshData.AppID)

			const deltaSeconds = 2.0
			assert.InDelta(t, loginTime.Add(st.Cfg.TokenExpires).Unix(), data.Expires, deltaSeconds)
			assert.InDelta(t, loginTime.Add(st.Cfg.RefreshTokenExpires).Unix(), refreshData.Expires, deltaSeconds)
		})
	}
}

func TestFailRegister(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name    string
		request *ssov1.RegisterRequest
		err     error
	}{
		{
			name: "success",
			request: &ssov1.RegisterRequest{
				Email:    "test@test.com",
				Password: "somePassword123",
				Name:     "test",
			},
		},
		{
			name: "empty email",
			request: &ssov1.RegisterRequest{
				Email:    "",
				Password: "123",
				Name:     "test",
			},
			err: ErrEmptyEmail,
		},
		{
			name: "empty password",
			request: &ssov1.RegisterRequest{
				Email:    "test",
				Password: "",
				Name:     "test",
			},
			err: ErrEmptyPass,
		},
		{
			name: "empty name",
			request: &ssov1.RegisterRequest{
				Email:    "test",
				Password: "123",
				Name:     "",
			},
			err: ErrEmptyName,
		},
		{
			name: "already exist user",
			request: &ssov1.RegisterRequest{
				Email:    "test@test.com",
				Password: "somePassword123",
				Name:     "test",
			},
			err: ErrUserAlreadyExist,
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			registerResp, err := st.AuthClient.Register(ctx, test.request)
			if err == nil {
				require.NoError(t, err)
				require.NotEmpty(t, registerResp.GetUserId())
			} else {
				require.Equal(t, test.err.Error(), err.Error())
				require.Empty(t, registerResp.GetUserId())
			}
		})
	}
}

func TestFailLogin(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name    string
		request *ssov1.LoginRequest
		err     error
	}{
		{
			name: "empty email",
			request: &ssov1.LoginRequest{
				Email:    "",
				Password: "123",
				AppId:    appID,
			},
			err: ErrEmptyEmail,
		},
		{
			name: "empty password",
			request: &ssov1.LoginRequest{
				Email:    "test",
				Password: "",
				AppId:    appID,
			},
			err: ErrEmptyPass,
		},
		{
			name: "empty app id",
			request: &ssov1.LoginRequest{
				Email:    "test",
				Password: "123",
				AppId:    0,
			},
			err: ErrAppID,
		},
		{
			name: "invalid password or email",
			request: &ssov1.LoginRequest{
				Email:    "fjasjlasjfl;asl;f",
				Password: "fashsfasfasfy7af8923",
				AppId:    appID,
			},
			err: ErrInvPassOrEmail,
		},
	}

	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loginResp, err := st.AuthClient.Login(ctx, tt.request)
			require.Equal(t, tt.err.Error(), err.Error())
			require.Empty(t, loginResp.GetAccessToken())
			require.Empty(t, loginResp.GetRefreshToken())
		})
	}
}

func generateFakeUsers(usersCount int) []*ssov1.RegisterRequest {
	users := make([]*ssov1.RegisterRequest, usersCount)

	for i := 0; i < usersCount; i++ {
		users[i] = &ssov1.RegisterRequest{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
			Name:     gofakeit.FirstName(),
		}
	}

	return users
}

func randomFakePassword() string {
	return gofakeit.Password(
		true,
		true,
		true,
		true,
		false,
		passDefaultLength,
	)
}
