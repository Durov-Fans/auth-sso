package postgres

import (
	"auth-service/internal/domains/models"
	"auth-service/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type Storage struct {
	db *pgxpool.Pool
}

func (s *Storage) ValidateUser(ctx context.Context, tgHash string) (models.UserResponse, error) {
	tx, err := s.db.Begin(ctx)

	if err != nil {
		return models.UserResponse{}, fmt.Errorf("Ошибка сервера")
	}

	defer tx.Rollback(ctx)

	var user models.UserResponse

	err = tx.QueryRow(ctx, `UPDATE users 
SET last_login = NOW() 
WHERE tgId = $1 
RETURNING tgid, id, first_name, last_name, user_name, user_name_locale, photo_url`, tgHash).Scan(&user.TgId, &user.ID, &user.FirstName, &user.LastName, &user.Username, &user.UserNameLocale, &user.PhotoURL)
	if errors.Is(err, sql.ErrNoRows) {
		log.Println("Пользователь не найден")
		return models.UserResponse{}, fmt.Errorf(" Пользователь не найден %s ", storage.ErrUserNotFound)
	}
	if user.IsBanned {
		log.Println("Пользователь забанен")
		return models.UserResponse{}, fmt.Errorf("Пользователь забанен")
	}
	if err := tx.Commit(ctx); err != nil {
		return models.UserResponse{}, fmt.Errorf("Ошибка комита")
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, tgHash string) (bool, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("Transaction Error", err)
	}
	defer tx.Rollback(ctx)
	var isAdmin bool
	err = tx.QueryRow(ctx, `SELECT is_admin FROM users where tgid = $1`, tgHash).Scan(&isAdmin)
	if err != nil {
		return false, fmt.Errorf("IsAdmin Error: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		fmt.Println("SaveUser", err)
		return false, fmt.Errorf("Ошибка базы данных")

	}
	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, serviceId int64) (models.App, error) {
	tx, err := s.db.Begin(ctx)

	if err != nil {
		return models.App{}, err
	}
	defer tx.Rollback(ctx)
	var app models.App
	err = tx.QueryRow(ctx, `SELECT * FROM apps WHERE id = $1`, serviceId).Scan(&app.ID, &app.Name, &app.Secret)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf(" Клиент сервис не найден %s ", storage.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf(" Ошибка", err)
	}
	if err := tx.Commit(ctx); err != nil {
		fmt.Println("SaveUser", err)
		return models.App{}, fmt.Errorf("Ошибка базы данных")

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
	return &Storage{db: pool}, nil

}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) SaveUser(ctx context.Context, tgId string, User models.User) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Ошибка начала транзакции")
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO users (tgid, first_name, last_name,user_name,user_name_locale, last_login, photo_url,is_admin)
         VALUES ($1, $2, $3,$4,$5, NOW(), $6,$7)`, tgId, User.FirstName, User.LastName, User.Username, User.UserNameLocale, User.PhotoURL, User.IsAdmin)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return storage.ErrUserExist
		}
		return fmt.Errorf("Ошибка транзакции")
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Ошибка базы данных")

	}
	return nil
}
