package svc

import (
	"github.com/uwu-octane/antBackend/user/internal/config"
	"github.com/uwu-octane/antBackend/user/internal/model"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                      config.Config
	Master                      sqlx.SqlConn
	Replica                     sqlx.SqlConn
	ReadFromReplica             bool
	FallbackToMasterOnReadError bool
	Users                       model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	master := sqlx.NewSqlConn(c.UserDatabase.Driver, c.UserDatabase.MasterDSN)
	replica := sqlx.NewSqlConn(c.UserDatabase.Driver, c.UserDatabase.ReplicaDSN)
	users := model.NewUsersModel(replica, master)
	return &ServiceContext{
		Config:                      c,
		Master:                      master,
		Replica:                     replica,
		ReadFromReplica:             c.UserReadStrategy.FromReplica,
		FallbackToMasterOnReadError: c.UserReadStrategy.FallbackToMasterOnReadError,
		Users:                       users,
	}
}
