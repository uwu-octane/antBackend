package svc

import (
	"github.com/uwu-octane/antBackend/auth/internal/config"
	"github.com/uwu-octane/antBackend/auth/internal/model"
	"github.com/uwu-octane/antBackend/auth/internal/util"
	dbutil "github.com/uwu-octane/antBackend/common/db/util"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/sync/singleflight"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Redis
	Key    string

	Master      sqlx.SqlConn
	Replica     sqlx.SqlConn
	RfGroup     *singleflight.Group
	TokenHelper *util.TokenHelper

	AuthUsers model.AuthUsersModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	redis := redis.MustNewRedis(c.AuthRedis.RedisConf)
	master := sqlx.NewSqlConn(c.AuthDatabase.Driver, c.AuthDatabase.MasterDSN)
	replica := sqlx.NewSqlConn(c.AuthDatabase.Driver, c.AuthDatabase.ReplicaDSN)
	selector := dbutil.NewSelector(replica, master, c.AuthReadStrategy.FromReplica, c.AuthReadStrategy.FallbackToMasterOnReadError, nil)
	return &ServiceContext{
		Config:      c,
		Redis:       redis,
		Key:         c.AuthRedis.Key,
		Master:      master,
		Replica:     replica,
		RfGroup:     &singleflight.Group{},
		TokenHelper: util.CreateTokenHelper(c.JwtAuth),
		AuthUsers:   model.NewAuthUsersModel(replica, master, selector),
	}
}
