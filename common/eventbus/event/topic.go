package event

type Env string

const (
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
)

type TopicSet struct {
	UserEvents string
}

func BuildTopics(env Env, suffix string) TopicSet {
	prefix := string(env)
	return TopicSet{
		UserEvents: prefix + suffix,
	}
}
