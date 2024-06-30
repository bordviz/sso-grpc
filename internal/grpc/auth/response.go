package auth

import (
	service "grpc/internal/services/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ResponseError(err error) error {
	switch err {
	case service.ErrInvPassOrEmail:
		return status.Error(codes.InvalidArgument, service.ErrInvPassOrEmail.Error())
	case service.ErrUserAlreadyExist:
		return status.Error(codes.AlreadyExists, service.ErrUserAlreadyExist.Error())
	case service.ErrInvalidData:
		return status.Error(codes.InvalidArgument, service.ErrInvalidData.Error())
	case service.ErrUnauthorized:
		return status.Error(codes.Unauthenticated, service.ErrUnauthorized.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
