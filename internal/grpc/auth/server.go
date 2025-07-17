package auth

import (
	"auth-service/internal/domains/models"
	"auth-service/internal/services/auth"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/mux"

	"net/http"
)

type Auth interface {
	ValidateUser(ctx context.Context, userData string, serviceId int64, w http.ResponseWriter) (token string, err error)
	RegisterUser(ctx context.Context, userData string, userNameLocale string, serviceId int64) (token string, err error)
	IsAdmin(ctx context.Context, userHash string) (isAdmin bool, err error)
}

type ServerApi struct {
	services auth.Auth
	port     string
}

func Register(authService auth.Auth, port string) *http.Server {
	api := ServerApi{
		services: authService,
		port:     port,
	}
	router := api.configureRouting()
	fmt.Println("fdasdfsdfadsf %s", api.port)
	return &http.Server{Addr: api.port, Handler: &router}
}
func (s *ServerApi) configureRouting() mux.Router {
	r := *mux.NewRouter()

	r.HandleFunc("/register", s.RegisterUser)
	r.HandleFunc("/validate", s.ValidateUser)
	r.HandleFunc("/isAdmin", s.IsAdmin)

	return r
}
func (s *ServerApi) ValidateUser(w http.ResponseWriter, r *http.Request) {
	var req models.InitDataRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка десериализации", http.StatusBadRequest)
		return
	}
	if err := validateLogin(req, w); err != nil {
		return
	}
	ctx := r.Context()
	token, err := s.services.ValidateUser(ctx, req.InitData, req.ServiceId, w)
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
		return fmt.Errorf("юзер обязателен")
	}

	if req.ServiceId == 0 {
		http.Error(w, "Неизвестный сервис", http.StatusBadRequest)
		return fmt.Errorf("неизвестный сервис")
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
func (s *ServerApi) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка десериализации", http.StatusBadRequest)
		return
	}
	if err := validateRegister(req, w); err != nil {
		return
	}
	ctx := r.Context()
	token, err := s.services.RegisterUser(ctx, req.UserHash, req.UserNameLocale, req.ServiceID)
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

func (s *ServerApi) IsAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	return
}
