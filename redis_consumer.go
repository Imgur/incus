package incus

import "log"

type RedisConsumer struct {
	commands <-chan RedisCommand
	pool     *redisPool
}

func NewRedisConsumer(commands <-chan RedisCommand, pool *redisPool) *RedisConsumer {
	consumer := &RedisConsumer{
		commands: commands,
		pool:     pool,
	}

	go consumer.ConsumeForever()

	return consumer
}

func (r *RedisConsumer) ConsumeForever() {
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
