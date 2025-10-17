package model

import (
	"context"
	"database/sql"

	"github.com/uwu-octane/antBackend/common/commonutil"
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
	replica  sqlx.SqlConn
	master   sqlx.SqlConn
	selector *commonutil.Selector
}

func NewUsersModel(replica sqlx.SqlConn, master sqlx.SqlConn, selector *commonutil.Selector) *defaultUserModel {
	return &defaultUserModel{
		replica:  replica,
		master:   master,
		selector: selector,
	}
}

const userFields = "id, username, email, display_name, avatar_url, created_at, updated_at"

func (m *defaultUserModel) FindOneWithCallBack(ctx context.Context, id string) (*User, error) {
	var result User
	const query = "SELECT " + userFields + " FROM users WHERE id = $1 LIMIT 1"

	err := m.selector.Do(ctx, func(ctx context.Context, conn sqlx.SqlConn) error {
		//idempotent
		var u User
		if err := conn.QueryRowCtx(ctx, &u, query, id); err != nil {
			return err
		}
		result = u
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *defaultUserModel) FindOne(ctx context.Context, id string) (*User, error) {
	var user User
	const query = "SELECT " + userFields + " FROM users WHERE id = $1 LIMIT 1"
	err := m.replica.QueryRowCtx(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
