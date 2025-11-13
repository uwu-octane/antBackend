package svc

import (
	"time"

	"github.com/uwu-octane/antBackend/common/eventbus/publisher"
	kpub "github.com/uwu-octane/antBackend/common/eventbus/publisher/kafka"

	dbutil "github.com/uwu-octane/antBackend/common/db/util"
	eventbus "github.com/uwu-octane/antBackend/common/eventbus/event"
	"github.com/uwu-octane/antBackend/user/internal/config"
	event "github.com/uwu-octane/antBackend/user/internal/event"
	"github.com/uwu-octane/antBackend/user/internal/model"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config           config.Config
	Master           sqlx.SqlConn
	Replica          sqlx.SqlConn
	Users            model.UserModel
	UserEventsPusher *publisher.EventBusPublisher
}

func NewServiceContext(c config.Config) *ServiceContext {
	master := sqlx.NewSqlConn(c.UserDatabase.Driver, c.UserDatabase.MasterDSN)
	replica := sqlx.NewSqlConn(c.UserDatabase.Driver, c.UserDatabase.ReplicaDSN)

	selector := dbutil.NewSelector(replica, master, c.UserReadStrategy.FromReplica, c.UserReadStrategy.FallbackToMasterOnReadError, nil)
	users := model.NewUsersModel(replica, master, selector)

	userEventsPusher := kafkaUserEventsPusher(c)

	return &ServiceContext{
		Config:           c,
		Master:           master,
		Replica:          replica,
		Users:            users,
		UserEventsPusher: userEventsPusher,
	}
}

func kafkaUserEventsPusher(c config.Config) *publisher.EventBusPublisher {
	topics := eventbus.BuildTopics(eventbus.Env(c.Kafka.Env), event.TopicSuffixUserEvents)
	opts := kpub.ProducerOptions{
		Brokers:         c.Kafka.Brokers,
		Acks:            c.KafkaUserProducer.Acks,
		Idempotent:      c.KafkaUserProducer.Idempotent,
		RetryMax:        c.KafkaUserProducer.RetryMax,
		Compression:     c.KafkaUserProducer.Compression,
		FlushBytes:      c.KafkaUserProducer.FlushBytes,
		FlushMessages:   c.KafkaUserProducer.FlushMessages,
		FlushFrequency:  time.Duration(c.KafkaUserProducer.FlushFrequencyMs) * time.Millisecond,
		MaxMessageBytes: c.KafkaUserProducer.MaxMessageBytes,
		EnableSASL:      c.KafkaUserProducer.SASL.Enable,
		SASLMechanism:   c.KafkaUserProducer.SASL.Mechanism,
		SASLUsername:    c.KafkaUserProducer.SASL.Username,
		SASLPassword:    c.KafkaUserProducer.SASL.Password,
		EnableTLS:       c.KafkaUserProducer.TLS.Enable,
	}
	pub, err := kpub.NewSaramaPublisher(&opts)
	if err != nil {
		logx.Errorw("create kafka user events publisher failed", logx.Field("error", err))
		return nil
	}
	return publisher.NewEventBusPublisher(pub, topics)
}
