package eventbus

type Env string

const (
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
)

type TopicSet struct {
	UserEvents string
}

func BuildTopics(env Env) TopicSet {
	prefix := string(env)
	return TopicSet{
		UserEvents: prefix + ".user.service.user-events",
	}
}
