package auth

import (
	"auth-service/internal/domains/models"
	"context"
	"encoding/json"
	"fmt"
	ssov1 "github.com/Durov-Fans/protos/gen/go/sso"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"google.golang.org/grpc"
	"net/http"
)

type Auth interface {
	ValidateUser(ctx context.Context, userData string, serviceId int64, w http.ResponseWriter) (token string, err error)
	RegisterUser(ctx context.Context, userData string, userNameLocale string, serviceId int64) (token string, err error)
	IsAdmin(ctx context.Context, userHash string) (isAdmin bool, err error)
}

type serverApi struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverApi{auth: auth})
}

func (s *serverApi) ValidateUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req models.InitDataRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка десериализации", http.StatusBadRequest)
		return
	}
	if err := validateLogin(req, w); err != nil {
		return
	}

	token, err := s.auth.ValidateUser(ctx, req.InitData, req.ServiceId, w)
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
	if err != nil {
		return
	}

}
func validateLogin(req models.InitDataRequest, w http.ResponseWriter) error {
	if req.InitData == "" {
		http.Error(w, "Юзер обязателен", http.StatusBadRequest)
		return fmt.Errorf("Юзер обязателен")
	}

	if req.ServiceId == 0 {
		http.Error(w, "Неизвестный сервис", http.StatusBadRequest)
		return fmt.Errorf("Неизвестный сервис")
	}
	return nil
}

func validateRegister(req models.RegisterRequest, w http.ResponseWriter) error {
	if req.UserNameLocale == "" {
		http.Error(w, "Внутренний никнейм обязателен", http.StatusBadRequest)
		return fmt.Errorf("Внутренний никнейм обязателен")
	}
	if req.UserHash == "" {
		http.Error(w, "Хеш обязателен", http.StatusBadRequest)
		return fmt.Errorf("Хеш обязателен")
	}

	if req.ServiceID == 0 {
		http.Error(w, "Неизвестный сервис", http.StatusBadRequest)
		return fmt.Errorf("Неизвестный сервис")
	}

	return nil
}
func (s *serverApi) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка десериализации", http.StatusBadRequest)
		return
	}
	if err := validateRegister(req, w); err != nil {
		return
	}

	token, err := s.auth.RegisterUser(ctx, req.UserHash, req.UserNameLocale, req.ServiceID)
	if err != nil {
		http.Error(w, "Ошибка", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
	if err != nil {
		return
	}
}

func (s *serverApi) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	panic("implement me")
}
