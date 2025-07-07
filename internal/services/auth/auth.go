package auth

import (
	"auth-service/internal/domains/models"
	"auth-service/internal/lib/jwt"
	"auth-service/internal/lib/logger/sl"
	"auth-service/internal/storage"
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"

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
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userId int32) (bool, error)
}
type AppProvider interface {
	App(ctx context.Context, serviceId int32) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid Credentials")
	ErrInvalidApp         = errors.New("invalid App")
)

func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log, userSaver, userProvider, appProvider, tokenTTL,
	}
}
func (a Auth) Login(ctx context.Context, email string, password []byte, serviceId int32) (string, error) {
	log := a.log.With(slog.String("op", "app.LoginUser"), slog.String("email", email))

	log.Info("login user")

	user, err := a.userProvider.User(ctx, email)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("app.LoginUser, %w", ErrInvalidCredentials)
		}

		a.log.Error("Failed to get user", sl.Err(err))

		return "", fmt.Errorf("app.LoginUser, %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, password); err != nil {
		a.log.Error("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("app.LoginUser, %w", ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, serviceId)

	if err != nil {
		return "", fmt.Errorf("app.LoginUser, %w", err)
	}

	log.Info("user is logged in")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("Failed to get user", sl.Err(err))

	}
	return token, nil
}
func (a Auth) RegisterUser(ctx context.Context, email string, password string) (userId int64, err error) {

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
func (a Auth) IsAdmin(ctx context.Context, userId int32) (bool, error) {
	log := a.log.With(slog.String("op", "app.IsAdmin"), slog.Int("userId", int(userId)))

	log.Info("authorise user")

	isAdmin, err := a.userProvider.IsAdmin(ctx, userId)

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
