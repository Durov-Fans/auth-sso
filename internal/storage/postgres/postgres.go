package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func Open(storagPath string) (*Storage, error) {
	db, err := sql.Open("postgres", s.config.DatabaseUrl)
	fmt.Println(s.config.DatabaseUrl)

	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		fmt.Println("defs")
		return nil, err
	}

	fmt.Println("ready")
	return &Storage{db: db}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s Storage) SaveUser(ctx context.Context, firstName string, lastName string, userName string, photoUrl string, isAdmin bool) error {
	stmt, err := s.db.Prepare("INSERT INTO users values ()")
}
