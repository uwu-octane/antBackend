package svc

import (
	"github.com/uwu-octane/antBackend/user/internal/config"
	"github.com/uwu-octane/antBackend/user/internal/model"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config  config.Config
	Master  sqlx.SqlConn
	Replica sqlx.SqlConn
	Users   model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	master := sqlx.NewSqlConn(c.Database.Driver, c.Database.MasterDSN)
	replica := sqlx.NewSqlConn(c.Database.Driver, c.Database.ReplicaDSN)
	users := model.NewUsersModel(replica, master)
	return &ServiceContext{
		Config:  c,
		Master:  master,
		Replica: replica,
		Users:   users,
	}
}
