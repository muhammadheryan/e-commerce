package user

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/muhammadheryan/e-commerce/model"
)

type SQL struct {
	conn *sqlx.DB
}

type UserRepository interface {
	Create(ctx context.Context, req *model.UserEntity) (*model.UserEntity, error)
	Get(ctx context.Context, filter *model.UserFilter) (*model.UserEntity, error)
}

func NewUserRepository(conn *sqlx.DB) UserRepository {
	return &SQL{conn: conn}
}

const (
	insertUserQuery = `INSERT INTO user (name, email, phone, password_hash, created_at) VALUES (?, ?, ?, ?, NOW())`
	getUserBase     = `SELECT id, name, email, phone, password_hash, created_at, updated_at FROM user WHERE true`
)

func (s *SQL) Create(ctx context.Context, data *model.UserEntity) (*model.UserEntity, error) {
	result, err := s.conn.ExecContext(ctx, insertUserQuery, data.Name, data.Email, data.Phone, data.PasswordHash)
	if err != nil {
		return nil, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	data.ID = uint64(lastID)
	return data, nil
}

func (s *SQL) Get(ctx context.Context, filter *model.UserFilter) (*model.UserEntity, error) {
	query := getUserBase
	args := make([]any, 0, 3)

	if filter.ID != 0 {
		query += " AND id = ?"
		args = append(args, filter.ID)
	}
	if filter.Email != "" {
		query += " AND email = ?"
		args = append(args, filter.Email)
	}
	if filter.Phone != "" {
		query += " AND phone = ?"
		args = append(args, filter.Phone)
	}

	var entity model.UserEntity
	if err := s.conn.QueryRowxContext(ctx, query, args...).StructScan(&entity); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}
