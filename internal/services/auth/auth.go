package auth

import (
	"auth-service/internal/domains/models"
	"auth-service/internal/lib/crypto"
	"auth-service/internal/lib/jwt"
	"auth-service/internal/lib/logger/sl"
	"auth-service/internal/storage"
	"context"
	"errors"
	"fmt"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"log/slog"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
	tgToken      string
}
type UserSaver interface {
	SaveUser(ctx context.Context, tgId string, User models.User) error
}

type UserProvider interface {
	IsAdmin(ctx context.Context, tgId string) (isAdmin bool, err error)
	ValidateUser(ctx context.Context, userHash string) (models.UserResponse, error)
}
type AppProvider interface {
	App(ctx context.Context, serviceId int64) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid Credentials")
	ErrInvalidApp         = errors.New("invalid App")
)

func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration, tgToken string) *Auth {

	return &Auth{
		log, userSaver, userProvider, appProvider, tokenTTL, tgToken,
	}
}

func (a Auth) ValidateUser(ctx context.Context, userHash string, serviceId int64) (models.UserResponse, string, error) {
	log := a.log.With(slog.String("op", "app.ValidateUser"))

	log.Info("валидация пользователя")
	//expIn := 24 * time.Hour
	if err := initdata.Validate(userHash, a.tgToken, 0); err != nil {
		return models.UserResponse{}, "", fmt.Errorf("validate user: %w", err)
	}
	userDecodeHash, err := initdata.Parse(userHash)

	app, err := a.appProvider.App(ctx, serviceId)
	tgHash, err := crypto.HashTgID(userDecodeHash.User.ID)

	if err != nil {
		return models.UserResponse{}, "", fmt.Errorf("Ошибка хеширования: %w", err)
	}
	user, err := a.userProvider.ValidateUser(ctx, tgHash)
	if err != nil {
		return models.UserResponse{}, "", err
	}
	token, err := jwt.NewToken(tgHash, app, a.tokenTTL)

	if err != nil {
		return models.UserResponse{}, "", fmt.Errorf("Ошибка генерации токена: %w", err)

	}
	return user, token, nil
}

//func (s *Storage) User(ctx context.Context, userHash string) (models.User, error) {
//	_, err := s.db.Begin(ctx)
//	if err != nil {
//		return models.User{}, fmt.Errorf("Transaction Error", err)
//	}
//	var user models.User
//	//TODO Добавить Валидацию и Сериализацию юзера
//	err = s.db.QueryRow(ctx, `SELECT * FROM users WHERE hash = $1`, userHash).Scan(&user.ID, &user.Hash, &user.FirstName, &user.LastName, &user.Username, &user.UserNameLocale, &user.PhotoURL)
//	if err != nil {
//		if errors.Is(err, sql.ErrNoRows) {
//			return models.User{}, fmt.Errorf(" Пользователь не найден %s ", storage.ErrAppNotFound)
//		}
//
//		return models.User{}, fmt.Errorf("Ошибка", err)
//	}
//	return user, nil
//}

func (a Auth) RegisterUser(ctx context.Context, userHash string, userNameLocale string, serviceId int64) (string, error) {

	log := a.log.With(slog.String("op", "app.RegisterUser"), slog.Int("serviceId", int(serviceId)))

	log.Info("Регистрация")

	err := initdata.Validate(userHash, a.tgToken, 0)
	if err != nil {
		log.Error("Ошибка валидации", err)
		return "", status.Errorf(codes.Unauthenticated, "Токен не прошел валидацию")
	}
	userDecodeHash, err := initdata.Parse(userHash)

	if err != nil {
		log.Error("Ошибка десереализации")
		return "", status.Errorf(codes.Internal, "internal error")
	}

	tgHash, err := crypto.HashTgID(userDecodeHash.User.ID)
	if err != nil {
		log.Error("ошибка хеширования тг айди", sl.Err(err))

		return "", status.Errorf(codes.Internal, "Ошибка хеширования")
	}
	User := models.User{
		ID:             tgHash,
		FirstName:      userDecodeHash.User.FirstName,
		LastName:       userDecodeHash.User.LastName,
		PhotoURL:       userDecodeHash.User.PhotoURL,
		UserNameLocale: userNameLocale,
		Username:       userDecodeHash.User.Username,
		IsAdmin:        false,
	}
	err = a.userSaver.SaveUser(ctx, tgHash, User)
	if err != nil {
		log.Error("Ошибка сохранениня юзера", sl.Err(err))

		if errors.As(err, &storage.ErrUserExist) {
			return "", err
		}
		return "", status.Errorf(codes.Internal, "internal error")
	}

	app, err := a.appProvider.App(ctx, serviceId)

	log.Info("Пользователь зарегистрирован")

	token, err := jwt.NewToken(User.ID, app, a.tokenTTL)
	if err != nil {
		log.Error("Ошибка генерации токена", sl.Err(err))
		status.Errorf(codes.Internal, "Ошибка генерации токена")
	}
	log.Info("asdasddasdds")
	return token, nil
}
func (a Auth) IsAdmin(ctx context.Context, initData string) (bool, error) {
	log := a.log.With(slog.String("op", "app.IsAdmin"))

	log.Info("authorise user")
	err := initdata.Validate(initData, a.tgToken, 0)
	if err != nil {
		log.Error("Ошибка валидации", err)
		return false, fmt.Errorf("Токен не прошел валидацию: %w", err)
	}
	userDecodeHash, err := initdata.Parse(initData)
	tgHash, err := crypto.HashTgID(userDecodeHash.User.ID)
	isAdmin, err := a.userProvider.IsAdmin(ctx, tgHash)

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
