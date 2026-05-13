package asynqq

import "github.com/hibiken/asynq"

// NewClient cria cliente asynq para a API enfileirar jobs.
func NewClient(redisAddr string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
}

// NewServer cria servidor asynq para o worker.
func NewServer(redisAddr string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
		},
	)
}
