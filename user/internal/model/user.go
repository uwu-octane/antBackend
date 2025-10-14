package model

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type User struct {
	Id          string         `db:"id"`
	Username    string         `db:"username"`
	Email       sql.NullString `db:"email"`
	DisplayName sql.NullString `db:"display_name"`
	AvatarUrl   sql.NullString `db:"avatar_url"`
	CreatedAt   string         `db:"created_at"`
	UpdatedAt   string         `db:"updated_at"`
}

type UserModel interface {
	// read only: replica
	FindOne(ctx context.Context, id string) (*User, error)
	// write: master
	//todo implement insert, update, delete
}

type defaultUserModel struct {
	replica sqlx.SqlConn
	master  sqlx.SqlConn
}

func NewUsersModel(replica sqlx.SqlConn, master sqlx.SqlConn) *defaultUserModel {
	return &defaultUserModel{
		replica: replica,
		master:  master,
	}
}

const userFields = "id, username, email, display_name, avatar_url, created_at, updated_at"

func (m *defaultUserModel) FindOne(ctx context.Context, id string) (*User, error) {
	var user User
	query := "SELECT " + userFields + " FROM users WHERE id = $1 LIMIT 1"
	if err := m.replica.QueryRowCtx(ctx, &user, query, id); err != nil {
		if err == sql.ErrNoRows {
			logx.WithContext(ctx).Errorf("user not found: %s", id)
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}
