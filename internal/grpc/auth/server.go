package auth

import (
	"auth-service/internal/domains/models"
	"auth-service/internal/services/auth"
	"auth-service/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/mux"

	"net/http"
)

type Auth interface {
	ValidateUser(ctx context.Context, userData string, serviceId int64) (token string, err error)
	RegisterUser(ctx context.Context, userData string, userNameLocale string, serviceId int64) (token string, err error)
	IsAdmin(ctx context.Context, tgId int64) (isAdmin bool, err error)
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
	return &http.Server{Addr: api.port, Handler: &router}
}
func (s *ServerApi) configureRouting() mux.Router {
	r := *mux.NewRouter()

	r.HandleFunc("/register", s.RegisterUser).Methods("POST")
	r.HandleFunc("/validate", s.ValidateUser).Methods("GET")
	r.HandleFunc("/isAdmin", s.IsAdmin).Methods("GET")

	return r
}
func (s *ServerApi) ValidateUser(w http.ResponseWriter, r *http.Request) {
	var req models.InitDataRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка десериализации", http.StatusBadRequest)
		return
	}
	if err := validateValidation(req, w); err != nil {
		http.Error(w, "ошибка валидации полей", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	user, token, err := s.services.ValidateUser(ctx, req.InitData, req.ServiceId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  user,
	})
	if err != nil {
		return
	}

}
func validateValidation(req models.InitDataRequest, w http.ResponseWriter) error {
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
		if errors.As(err, &storage.ErrUserExist) {
			http.Error(w, "Ошибка", http.StatusConflict)
			return
		}
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
	var req models.IsAdmin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "ошибка десериализации", http.StatusBadRequest)
		return
	}
	if req.InitData == "" {
		http.Error(w, "хеш обязателен", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	isAdmin, err := s.services.IsAdmin(ctx, req.InitData)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidApp) {
			http.Error(w, "Неизвестный сервис", http.StatusBadRequest)
		}
		http.Error(w, "Ошибка", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]bool{
		"isAdmin": isAdmin,
	})
	if err != nil {
		return
	}
}
