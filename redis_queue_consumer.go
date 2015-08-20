package incus

import "log"

type RedisQueueConsumer struct {
	commands <-chan RedisCommand
	pool     *redisPool
}

func NewRedisQueueConsumer(commands <-chan RedisCommand, pool *redisPool) *RedisQueueConsumer {
	consumer := &RedisQueueConsumer{
		commands: commands,
		pool:     pool,
	}

	go consumer.ConsumeForever()

	return consumer
}

func (r *RedisQueueConsumer) ConsumeForever() {
	for {
		command := <-r.commands

		if DEBUG {
			log.Println("Dequeued one command in consumer")
		}

		conn, success := r.pool.Get()

		if success {
			result, err := command.Callback(conn)

			// The Result channel is a buffered channel of length 1 so this does not block.
			command.Result <- RedisCommandResult{
				Value: result,
				Error: err,
			}

			r.pool.Close(conn)
		} else {
			log.Println("Failed to get redis connection")
		}
	}
}
