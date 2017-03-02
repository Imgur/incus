package incus

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
)

const ClientsKey = "SocketClients"
const PageKey = "PageClients"
const PresenceKeyPrefix = "ClientPresence"

var timedOut = errors.New("Timed out waiting for Redis")

type RedisCallback (func(redis.Conn) (interface{}, error))

type RedisCommandResult struct {
	Error error
	Value interface{}
}

type RedisCommand struct {
	Callback RedisCallback
	Result   chan RedisCommandResult
}

type RedisStore struct {
	clientsKey        string
	pageKey           string
	presenceKeyPrefix string
	presenceDuration  int64

	server                    string
	port                      int
	pool                      *redisPool
	pollingFreq               time.Duration
	incomingRedisActivityCmds chan RedisCommand
	redisPendingQueue         *RedisQueue
}

//connection pool implimentation
type redisPool struct {
	connections chan redis.Conn
	maxIdle     int
	connFn      func() (redis.Conn, error) // function to create new connection.
}

func newRedisStore(redisHost string, redisPort, numberOfActivityConsumers, connPoolSize int, stats RuntimeStats) *RedisStore {

	pool := &redisPool{
		connections: make(chan redis.Conn, connPoolSize),
		maxIdle:     connPoolSize,

		connFn: func() (redis.Conn, error) {
			client, err := redis.Dial("tcp", fmt.Sprintf("%s:%v", redisHost, redisPort))
			if err != nil {
				log.Printf("Redis connect failed: %s\n", err.Error())
				return nil, err
			}

			return client, nil
		},
	}

	redisPendingQueue := NewRedisQueue(numberOfActivityConsumers, stats, pool)

	return &RedisStore{
		redisPendingQueue: redisPendingQueue,
		clientsKey:        ClientsKey,
		pageKey:           PageKey,
		presenceKeyPrefix: PresenceKeyPrefix,
		presenceDuration:  60,
		server:            redisHost,
		port:              redisPort,
		pool:              pool,
		pollingFreq:       time.Millisecond * 100,
	}

}

func (this *redisPool) Get() (redis.Conn, bool) {

	var conn redis.Conn
	select {
	case conn = <-this.connections:
	default:
		conn, err := this.connFn()
		if err != nil {
			return nil, false
		}

		return conn, true
	}

	if err := this.testConn(conn); err != nil {
		return this.Get() // if connection is bad, get the next one in line until base case is hit, then create new client
	}

	return conn, true
}

func (this *redisPool) Close(conn redis.Conn) {
	select {
	case this.connections <- conn:
		return
	default:
		conn.Close()
	}
}

func (this *redisPool) testConn(conn redis.Conn) error {
	if _, err := conn.Do("PING"); err != nil {
		conn.Close()
		return err
	}

	return nil
}

func (this *RedisStore) GetConn() (redis.Conn, error) {

	client, ok := this.pool.Get()
	if !ok {
		return nil, errors.New("Error while getting redis connection")
	}

	return client, nil
}

func (this *RedisStore) CloseConn(conn redis.Conn) {
	this.pool.Close(conn)
}

func (this *RedisStore) Subscribe(c chan []byte, channel string) (redis.Conn, error) {
	conn, err := this.GetConn()
	if err != nil {
		return nil, err
	}

	psc := redis.PubSubConn{Conn: conn}
	psc.Subscribe(channel)
	go func() {
		defer conn.Close()
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				c <- v.Data
			case redis.Subscription:
			case error:
				log.Printf("Error receiving: %s. Reconnecting...", v.Error())
				conn, err = this.GetConn()
				if err != nil {
					log.Println(err)
				}

				psc = redis.PubSubConn{Conn: conn}
				psc.Subscribe(channel)
			}
		}
	}()

	return conn, nil
}

func (this *RedisStore) Poll(c chan []byte, queue string) error {
	go func() {
		consumer, _ := this.GetConn()
		defer this.CloseConn(consumer)
		var err error

		for {
			consumer, err = this.GetConn()
			if err != nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			message, err := redis.Bytes(consumer.Do("LPOP", queue))
			this.CloseConn(consumer)
            log.Printf("Message from redis: %s\n", message)

			if err == nil && len(message) > 0 {
				c <- message
			} else {
				time.Sleep(this.pollingFreq)
			}
		}
	}()

	return nil
}

func (this *RedisStore) MarkActive(user, socket_id string, timestamp int64) error {
	return this.redisPendingQueue.RunAsyncTimeout(5*time.Second, func(conn redis.Conn) (result interface{}, err error) {
		userSortedSetKey := this.presenceKeyPrefix + ":" + user

		conn.Send("MULTI")
		conn.Send("ZADD", userSortedSetKey, timestamp, socket_id)
		conn.Send("EXPIRE", userSortedSetKey, timestamp+this.presenceDuration)
		return conn.Do("EXEC")
	}).Error
}

func (this *RedisStore) MarkInactive(user, socket_id string) error {
	return this.redisPendingQueue.RunAsyncTimeout(5*time.Second, func(conn redis.Conn) (result interface{}, err error) {
		userSortedSetKey := this.presenceKeyPrefix + ":" + user

		return conn.Do("ZREM", userSortedSetKey, socket_id)
	}).Error
}

func (this *RedisStore) QueryIsUserActive(user string, nowTimestamp int64) (bool, error) {
	result := this.redisPendingQueue.RunAsyncTimeout(5*time.Second, func(conn redis.Conn) (result interface{}, err error) {
		userSortedSetKey := this.presenceKeyPrefix + ":" + user

		reply, err := conn.Do("ZRANGEBYSCORE", userSortedSetKey, nowTimestamp-this.presenceDuration, nowTimestamp)

		els := reply.([]interface{})

		return (len(els) > 0), nil
	})

	if result.Error != nil {
		return false, result.Error
	} else {
		return result.Value.(bool), nil
	}
}

func (this *RedisStore) GetIsLongpollKillswitchActive() (bool, error) {
	killswitchKey := viper.Get("longpoll_killswitch")

	result := this.redisPendingQueue.RunAsyncTimeout(5*time.Second, func(conn redis.Conn) (result interface{}, err error) {
		return conn.Do("TTL", killswitchKey)
	})

	if result.Error == nil {
		if result.Value.(int64) >= -1 {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, timedOut
}

func (this *RedisStore) ActivateLongpollKillswitch(seconds int64) error {
	killswitchKey := viper.Get("longpoll_killswitch")

	return this.redisPendingQueue.RunAsyncTimeout(5*time.Second, func(conn redis.Conn) (result interface{}, err error) {
		return conn.Do("SETEX", killswitchKey, seconds, "1")
	}).Error
}

func (this *RedisStore) DeactivateLongpollKillswitch() error {
	killswitchKey := viper.Get("longpoll_killswitch")

	return this.redisPendingQueue.RunAsyncTimeout(5*time.Second, func(conn redis.Conn) (result interface{}, err error) {
		return conn.Do("DEL", killswitchKey)
	}).Error
}

func (this *RedisStore) Publish(channel string, message string) {
	publisher, err := this.GetConn()
	if err != nil {
		return
	}
	defer this.CloseConn(publisher)

	publisher.Do("PUBLISH", channel, message)
}

func (this *RedisStore) Push(queue string, message string) {
	client, err := this.GetConn()
	if err != nil {
		return
	}
	defer this.CloseConn(client)

	client.Do("RPUSH", queue, message)
}

func (this *RedisStore) Save(sock *Socket) error {
	client, err := this.GetConn()
	if err != nil {
		return err
	}
	defer this.CloseConn(client)

	_, err = client.Do("SADD", this.clientsKey, sock.UID)
	if err != nil {
		return err
	}

	return nil
}

func (this *RedisStore) Remove(sock *Socket) error {
	client, err := this.GetConn()
	if err != nil {
		return err
	}
	defer this.CloseConn(client)

	_, err = client.Do("SREM", this.clientsKey, sock.UID)
	if err != nil {
		return err
	}

	return nil
}

func (this *RedisStore) Clients() ([]string, error) {
	client, err := this.GetConn()
	if err != nil {
		return nil, err
	}
	defer this.CloseConn(client)

	socks, err1 := redis.Strings(client.Do("SMEMBERS", this.clientsKey))
	if err1 != nil {
		return nil, err1
	}

	return socks, nil
}

func (this *RedisStore) Count() (int64, error) {
	client, err := this.GetConn()
	if err != nil {
		return 0, err
	}
	defer this.CloseConn(client)

	socks, err1 := redis.Int64(client.Do("SCARD", this.clientsKey))
	if err1 != nil {
		return 0, err1
	}

	return socks, nil
}

func (this *RedisStore) SetPage(sock *Socket) error {
	client, err := this.GetConn()
	if err != nil {
		return err
	}
	defer this.CloseConn(client)

	_, err = client.Do("HINCRBY", this.pageKey, sock.Page, 1)
	if err != nil {
		return err
	}

	return nil
}

func (this *RedisStore) UnsetPage(sock *Socket) error {
	client, err := this.GetConn()
	if err != nil {
		return err
	}
	defer this.CloseConn(client)

	var i int64
	i, err = redis.Int64(client.Do("HINCRBY", this.pageKey, sock.Page, -1))
	if err != nil {
		return err
	}

	if i <= 0 {
		client.Do("HDEL", this.pageKey, sock.Page)
	}

	return nil
}
