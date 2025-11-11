package svc

import (
	dbutil "github.com/uwu-octane/antBackend/common/db/util"
	"github.com/uwu-octane/antBackend/common/eventbus"
	"github.com/uwu-octane/antBackend/user/internal/config"
	"github.com/uwu-octane/antBackend/user/internal/model"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config           config.Config
	Master           sqlx.SqlConn
	Replica          sqlx.SqlConn
	Users            model.UserModel
	UserEventsPusher *kq.Pusher
}

func NewServiceContext(c config.Config) *ServiceContext {
	master := sqlx.NewSqlConn(c.UserDatabase.Driver, c.UserDatabase.MasterDSN)
	replica := sqlx.NewSqlConn(c.UserDatabase.Driver, c.UserDatabase.ReplicaDSN)

	selector := dbutil.NewSelector(replica, master, c.UserReadStrategy.FromReplica, c.UserReadStrategy.FallbackToMasterOnReadError, nil)
	users := model.NewUsersModel(replica, master, selector)

	topics := eventbus.BuildTopics(eventbus.Env(c.Kafka.Env))

	var userEventsPusher *kq.Pusher
	if len(c.Kafka.Brokers) > 0 {
		userEventsPusher = kq.NewPusher(c.Kafka.Brokers, topics.UserEvents)
	}

	return &ServiceContext{
		Config:           c,
		Master:           master,
		Replica:          replica,
		Users:            users,
		UserEventsPusher: userEventsPusher,
	}
}
