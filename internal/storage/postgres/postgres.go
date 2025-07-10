package postgres

import (
	"auth-service/internal/domains/models"
	"auth-service/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type Storage struct {
	db *pgxpool.Pool
}

func (s *Storage) User(ctx context.Context, userHash string) (models.User, error) {
	_, err := s.db.Begin(ctx)
	if err != nil {
		return models.User{}, fmt.Errorf("Transaction Error", err)
	}
	var user models.User
	//TODO Добавить Валидацию и Сериализацию юзера
	err = s.db.QueryRow(ctx, `SELECT * FROM users WHERE hash = $1`, userHash).Scan(&user.ID, &user.Hash, &user.FirstName, &user.LastName, &user.Username, &user.UserNameLocale, &user.PhotoURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf(" Пользователь не найден %s ", storage.ErrAppNotFound)
		}

		return models.User{}, fmt.Errorf("Ошибка", err)
	}
}

func (s *Storage) IsAdmin(ctx context.Context, userHash string) (bool, error) {
	_, err := s.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("Transaction Error", err)
	}
	var isAdmin bool
	err = s.db.QueryRow(ctx, `SELECT is_admin FROM users where hash = $1`, userHash).Scan(&isAdmin)
	if err != nil {
		return false, fmt.Errorf("IsAdmin Error: %v", err)
	}
	fmt.Println(isAdmin)
	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, serviceId int32) (models.App, error) {
	_, err := s.db.Begin(ctx)
	if err != nil {
		return models.App{}, err
	}
	var app models.App
	err = s.db.QueryRow(ctx, `SELECT * FROM apps WHERE id = $1`, serviceId).Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf(" Клиент сервис не найден %s ", storage.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf(" Ошибка", err)
	}

	return app, nil
}

func InitDB(storagPath string) (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pgxCfg, err := pgxpool.ParseConfig(storagPath)
	if err != nil {
		log.Fatal(" Ошибка парсинга строки подключения:", err)
	}
	pgxCfg.MaxConns = 1
	pgxCfg.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err := pool.Ping(ctx); err != nil {
		fmt.Println("Ошибка подключения к базе данных ")
		return nil, err
	}

	log.Println("Подключение к PostgresSQL успешно")
	fmt.Println("ready")
	return &Storage{db: pool}, nil

}

func (s *Storage) Close() {
	s.db.Close()
}

func (s Storage) SaveUser(ctx context.Context, id int64, firstName string, lastName string, userName string, photoUrl string, isAdmin bool) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return status.Error(codes.Internal, "Ошибка начала транзакции")
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO users (id, first_name, last_name,user_name, last_login, photo_url,is_admin)
         VALUES ($1, $2, $3,$4, NOW(), $5,$6)`, id, firstName, lastName, userName, photoUrl, isAdmin)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("SaveUser", storage.ErrUserExist)
		}
		return status.Error(codes.Internal, "Ошибка транзакции")
	}
	if err := tx.Commit(ctx); err != nil {
		fmt.Println("SaveUser", err)
		return status.Error(codes.Internal, "Ошибка базы данных")

	}
	return nil
}
