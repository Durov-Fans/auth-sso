package auth

import (
	"context"
	ssov1 "github.com/Durov-Fans/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, serviceId int32) (token string, err error)
	RegisterUser(email string, password string) (userId int64, err error)
	IsAdmin(ctx context.Context, userId int) (isAdmin bool, err error)
}

type serverApi struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverApi{auth: auth})
}

func (s *serverApi) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetServiceId())
	if err != nil {
		return nil, status.Error(codes.Internal, "Ошибка")
	}
	return &ssov1.LoginResponse{Token: token}, nil

}
func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "Email обязателен")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "Неверный пароль")
	}
	if req.GetServiceId() == 0 {
		return status.Error(codes.InvalidArgument, "Неизвестный сервис")
	}
	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "Email обязателен")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "Неверный пароль")
	}

	return nil
}
func (s *serverApi) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userId, err := s.auth.RegisterUser(req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "Ошибка")
	}
	return &ssov1.RegisterResponse{UserId: userId}, nil
}

func (s *serverApi) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	panic("implement me")
}
