package auth

import (
	"auth-service/internal/domains/models"
	"auth-service/internal/lib/crypto"
	"auth-service/internal/lib/jwt"
	"auth-service/internal/lib/logger/sl"
	"auth-service/internal/storage"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"log/slog"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}
type UserSaver interface {
	SaveUser(ctx context.Context, id int64, firstName string, lastName string, userName string, photoUrl string, isAdmin bool) error
}

type UserProvider interface {
	User(ctx context.Context, userHash string) (models.User, error)
	IsAdmin(ctx context.Context, userHash string) (isAdmin bool, err error)
}
type AppProvider interface {
	App(ctx context.Context, serviceId int32) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid Credentials")
	ErrInvalidApp         = errors.New("invalid App")
)
var secretKey = GenerateSecretKey("7901019694:AAEjOz9nQNZkmtByby8QljOehunWLez2xCk")

const MaxTimeDiff = 300

func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log, userSaver, userProvider, appProvider, tokenTTL,
	}
}
func GenerateSecretKey(botToken string) []byte {
	mac := hmac.New(sha256.New, []byte("WebAppData"))
	mac.Write([]byte(botToken))
	return mac.Sum(nil)
}

func (a Auth) Login(ctx context.Context, userHash string, serviceId int64) (string, error) {
	log := a.log.With(slog.String("op", "app.LoginUser"))

	log.Info("login user")
	if err := ValidateInitData(userHash); err != nil {

		return "", status.Errorf(codes.Internal, "internal error")
	}
	tgHash, err := crypto.HashTgID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err != nil {
		http.Error(w, "Ошибка начала транзакции", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	return "", nil
}
func (a Auth) RegisterUser(Hash string, userData string, userNameLocale string, serviceId int64) (token string, err error) {

	log := a.log.With(slog.String("op", "app.RegisterUser"), slog.String("email", email))

	log.Info("register user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Error("Failed to hash password", sl.Err(err))

		return 0, fmt.Errorf("app.RegisterUser, %w", err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("Failed to save user", sl.Err(err))
		return 0, fmt.Errorf("app.RegisterUser, %w", err)
	}
	log.Info("User register")

	return id, nil
}
func (a Auth) IsAdmin(ctx context.Context, userHash string) (bool, error) {
	log := a.log.With(slog.String("op", "app.IsAdmin"), slog.Int("userId", int(userId)))

	log.Info("authorise user")

	isAdmin, err := a.userProvider.IsAdmin(ctx, userHash)

	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("app.IsAdmin, %w", ErrInvalidApp)
		}
		log.Error("Failed to authorise user", sl.Err(err))
		return false, fmt.Errorf("app.IsAdmin, %w", err)
	}

	return isAdmin, nil
}
func ValidateInitData(initData string) error {
	params, err := url.ParseQuery(initData)
	if err != nil {
		return fmt.Errorf("ошибка парсинга initData: %v", err)
	}

	hash := params.Get("hash")
	if hash == "" {
		return errors.New(" hash не найден")
	}

	fmt.Println("Hash найден:", hash)

	params.Del("hash")

	// Сборка data_check_string
	var pairs []string
	for key, values := range params {
		for _, value := range values {
			pairs = append(pairs, key+"="+value)
		}
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// Генерация подписи HMAC
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(dataCheckString))
	generatedHash := hex.EncodeToString(h.Sum(nil))

	fmt.Println("🔍 Сгенерированный hash:", generatedHash)

	if generatedHash != hash {
		return errors.New("❌ Ошибка: hash не совпадает! данные могли быть подделаны")
	}

	authDateStr := params.Get("auth_date")
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return errors.New("❌ Неверный auth_date")
	}

	currentTime := time.Now().Unix()
	if currentTime-authDate > MaxTimeDiff {
		return errors.New("❌ Данные слишком старые")
	}

	return nil
}
