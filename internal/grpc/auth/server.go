package auth

import (
	"context"
	"grpc/internal/domain/models"

	ssov1 "github.com/bordviz/sso-protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	emptyValue = 0
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (tokens models.TokensPair, err error)
	Register(ctx context.Context, email string, password string, name string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64, appID int) (bool, error)
	RefreshToken(ctx context.Context, token string, appID int) (tokens models.TokensPair, err error)
	CurrentUser(ctx context.Context, token string, appID int) (models.UserRead, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	tokens, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		return nil, ResponseError(err)
	}

	return &ssov1.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetName())
	if err != nil {
		return nil, ResponseError(err)
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId(), int(req.GetAppId()))
	if err != nil {
		return nil, ResponseError(err)
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func (s *serverAPI) RefreshToken(ctx context.Context, req *ssov1.RefreshTokenRequest) (*ssov1.RefreshTokenResponse, error) {
	if err := validateRefreshToken(req); err != nil {
		return nil, err
	}

	tokens, err := s.auth.RefreshToken(ctx, req.GetToken(), int(req.GetAppId()))
	if err != nil {
		return nil, ResponseError(err)
	}

	return &ssov1.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *serverAPI) CurrentUser(ctx context.Context, req *ssov1.CurrentUserRequest) (*ssov1.CurrentUserResponse, error) {
	if err := validateCurrentUser(req); err != nil {
		return nil, err
	}

	user, err := s.auth.CurrentUser(ctx, req.GetToken(), int(req.GetAppId()))
	if err != nil {
		return nil, ResponseError(err)
	}

	return &ssov1.CurrentUserResponse{
		UserId: user.ID,
		Email:  user.Email,
		Name:   user.Name,
	}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "empty email")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "empty password")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "empty app id")
	}

	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "empty email")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "empty password")
	}

	if req.GetName() == "" {
		return status.Error(codes.InvalidArgument, "empty name")
	}

	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "empty user id")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "empty app id")
	}

	return nil
}

func validateRefreshToken(req *ssov1.RefreshTokenRequest) error {
	if req.GetToken() == "" {
		return status.Error(codes.InvalidArgument, "empty token")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "empty app id")
	}

	return nil
}

func validateCurrentUser(req *ssov1.CurrentUserRequest) error {
	if req.GetToken() == "" {
		return status.Error(codes.InvalidArgument, "empty token")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "empty app id")
	}

	return nil
}
