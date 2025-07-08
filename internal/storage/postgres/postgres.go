package postgres

import (
	"auth-service/internal/storage"
	"context"
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

func Open(storagPath string) (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pgxCfg, err := pgxpool.ParseConfig(storagPath)
	if err != nil {
		log.Fatal("❌ Ошибка парсинга строки подключения:", err)
	}
	pgxCfg.MaxConns = 1
	pgxCfg.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err := pool.Ping(ctx); err != nil {
		fmt.Println("Ошибка подключения к базе данных ")
		return nil, err
	}

	log.Println("✅ Подключение к PostgreSQL успешно")
	fmt.Println("ready")
	return &Storage{db: pool}, nil

	// Жёстко задаём ограничения

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
