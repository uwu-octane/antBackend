package model

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AuthUsers struct {
	Id           string         `db:"id"`
	Email        string         `db:"email"`
	PasswordHash string         `db:"password_hash"`
	PasswordAlgo sql.NullString `db:"password_algo"`
	CreatedAt    string         `db:"created_at"`
	UpdatedAt    string         `db:"updated_at"`
}

type AuthUsersModel interface {
	FindByEmail(ctx context.Context, email string) (*AuthUsers, error)
	FindByUsername(ctx context.Context, username string) (*AuthUsers, error)
}

type defaultAuthUsersModel struct {
	replica sqlx.SqlConn
	master  sqlx.SqlConn
}

func NewAuthUsersModel(replica sqlx.SqlConn, master sqlx.SqlConn) *defaultAuthUsersModel {
	return &defaultAuthUsersModel{
		replica: replica,
		master:  master,
	}
}

const authUsersFields = "id, email, password_hash, password_algo, created_at, updated_at"

func (m *defaultAuthUsersModel) FindByEmail(ctx context.Context, email string) (*AuthUsers, error) {
	var user AuthUsers
	query := "SELECT " + authUsersFields + " FROM auth_users WHERE email = $1 LIMIT 1"
	if err := m.replica.QueryRowCtx(ctx, &user, query, email); err != nil {
		if err == sql.ErrNoRows {
			logx.WithContext(ctx).Errorf("auth user not found: %s", email)
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

func (m *defaultAuthUsersModel) FindByUsername(ctx context.Context, username string) (*AuthUsers, error) {
	var user AuthUsers
	query := "SELECT " + authUsersFields + " FROM auth_users WHERE username = $1 LIMIT 1"
	if err := m.replica.QueryRowCtx(ctx, &user, query, username); err != nil {
		if err == sql.ErrNoRows {
			logx.WithContext(ctx).Errorf("auth user not found: %s", username)
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}
