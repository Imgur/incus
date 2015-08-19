package incus

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

const ClientsKey = "SocketClients"
const PageKey = "PageClients"
const PresenceKeyPrefix = "ClientPresence"

var timedOut = errors.New("Timed out waiting for Redis")

type RedisCommand (func(redis.Conn) error)

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
	pendingList               *RedisConsumerPendingList
}

//connection pool implimentation
type redisPool struct {
	connections chan redis.Conn
	maxIdle     int
	connFn      func() (redis.Conn, error) // function to create new connection.
}

func newRedisStore(redisHost string, redisPort, activityConsumers int, stats RuntimeStats) *RedisStore {

	pool := &redisPool{
		connections: make(chan redis.Conn, 10),
		maxIdle:     10,

		connFn: func() (redis.Conn, error) {
			client, err := redis.Dial("tcp", fmt.Sprintf("%s:%v", redisHost, redisPort))
			if err != nil {
				log.Printf("Redis connect failed: %s\n", err.Error())
				return nil, err
			}

			return client, nil
		},
	}

	incomingRedisActivityCmds := make(chan RedisCommand)
	incomingRedisActivityCmdsSend := (<-chan RedisCommand)(incomingRedisActivityCmds)
	redisConsumerPending := NewRedisConsumerPendingList(activityConsumers, incomingRedisActivityCmdsSend, stats, pool)

	return &RedisStore{
		incomingRedisActivityCmds: incomingRedisActivityCmds,
		pendingList:               redisConsumerPending,
		clientsKey:                ClientsKey,
		pageKey:                   PageKey,
		presenceKeyPrefix:         PresenceKeyPrefix,
		presenceDuration:          60,
		server:                    redisHost,
		port:                      redisPort,
		pool:                      pool,
		pollingFreq:               time.Millisecond * 100,
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

			if err == nil && len(message) > 0 {
				c <- message
			} else {
				time.Sleep(this.pollingFreq)
			}
		}
	}()

	return nil
}

func (this *RedisStore) MarkActive(user, socket_id string, timestamp int64) {
	work := func(conn redis.Conn) error {
		log.Println("Marking as active")

		userSortedSetKey := this.presenceKeyPrefix + ":" + user

		conn.Send("MULTI")
		conn.Send("ZADD", userSortedSetKey, timestamp, socket_id)
		conn.Send("EXPIRE", userSortedSetKey, timestamp+this.presenceDuration)
		_, err := conn.Do("EXEC")
		if err != nil {
			log.Printf("Error marking user as active: %s", err.Error())
		}

		return err
	}

	this.incomingRedisActivityCmds <- work
}

func (this *RedisStore) MarkInactive(user, socket_id string) {
	work := func(conn redis.Conn) error {
		log.Println("Marking as inactive")

		userSortedSetKey := this.presenceKeyPrefix + ":" + user

		_, err := conn.Do("ZREM", userSortedSetKey, socket_id)
		return err
	}

	this.incomingRedisActivityCmds <- work
}

func (this *RedisStore) QueryIsUserActive(user string, nowTimestamp int64) (bool, error) {
	resultChan := make(chan bool)

	work := func(conn redis.Conn) error {
		userSortedSetKey := this.presenceKeyPrefix + ":" + user

		log.Println("Querying if a user is active")
		reply, err := conn.Do("ZRANGEBYSCORE", userSortedSetKey, nowTimestamp-this.presenceDuration, nowTimestamp)

		els := reply.([]interface{})

		resultChan <- (len(els) > 0)

		return err
	}

	this.incomingRedisActivityCmds <- work

	select {
	case active := <-resultChan:
		return active, nil
	case <-time.After(5 * time.Second):
		return false, timedOut
	}
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
