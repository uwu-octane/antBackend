package model

import (
	"context"
	"database/sql"

	"github.com/uwu-octane/antBackend/common/commonutil"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AuthUsers struct {
	Id           string         `db:"id"`
	Username     sql.NullString `db:"username"`
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
	replica  sqlx.SqlConn
	master   sqlx.SqlConn
	selector *commonutil.Selector
}

func NewAuthUsersModel(replica sqlx.SqlConn, master sqlx.SqlConn, selector *commonutil.Selector) *defaultAuthUsersModel {
	return &defaultAuthUsersModel{
		replica:  replica,
		master:   master,
		selector: selector,
	}
}

const authUsersFields = "id, username, email, password_hash, password_algo, created_at, updated_at"

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

func (m *defaultAuthUsersModel) FindOneByIDWithCallBack(ctx context.Context, id string) (*AuthUsers, error) {
	var user AuthUsers
	query := "SELECT " + authUsersFields + " FROM auth_users WHERE id = $1 LIMIT 1"

	err := m.selector.Do(ctx, func(ctx context.Context, conn sqlx.SqlConn) error {
		//idempotent
		var u AuthUsers
		if err := conn.QueryRowCtx(ctx, &u, query, id); err != nil {
			return err
		}
		user = u
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}
