package auth

import (
	"auth-service/internal/app"
	"auth-service/internal/lib/jwt"
	"context"
	ssov1 "github.com/Durov-Fans/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, userHash string, serviceId int64) (token string, err error)
	RegisterUser(hash string, userData string, userNameLocale string, serviceId int64) (token string, err error)
	IsAdmin(ctx context.Context, userHash int) (isAdmin bool, err error)
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
	token, err := s.auth.Login(ctx, req.GetUserHash(), req.GetServiceId())
	if err != nil {
		return nil, status.Error(codes.Internal, "Ошибка")
	}
	return &ssov1.LoginResponse{Token: token}, nil

}
func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetUserHash() == "" {
		return status.Error(codes.InvalidArgument, "Юзер обязателен")
	}
	if req.GetServiceId() == 0 {
		return status.Error(codes.InvalidArgument, "Неизвестный сервис")
	}
	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetUserNameLocale() == "" {
		return status.Error(codes.InvalidArgument, "Внутренний никнейм обязателен")
	}
	if req.GetHash() == "" {
		return status.Error(codes.InvalidArgument, "Хеш обязателен")
	}

	if req.GetServiceId() == 0 {
		return status.Error(codes.InvalidArgument, "Неизвестный сервис")
	}

	if req.GetUserData() == "" {
		return status.Error(codes.InvalidArgument, "Данные пользователя обязательны")
	}

	return nil
}
func (s *serverApi) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	user, err := s.auth.RegisterUser(req.GetHash(), req.GetUserData(), req.GetUserNameLocale(), req.GetServiceId())
	if err != nil {
		return nil, status.Error(codes.Internal, "Ошибка")
	}
	app := s.
	return &ssov1.RegisterResponse{Token: jwt.NewToken(user, app, duration)}, nil
}

func (s *serverApi) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	panic("implement me")
}
