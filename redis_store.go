package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/garyburd/redigo/redis"
)

const ClientsKey = "SocketClients"
const PageKey = "PageClients"

type RedisStore struct {
	clientsKey string
	pageKey    string

	server string
	port   int
	pool   redisPool
}

//connection pool implimentation
type redisPool struct {
	connections chan redis.Conn
	maxIdle     int
	connFn      func() (redis.Conn, error) // function to create new connection.
}

func newRedisStore(redis_host string, redis_port int) *RedisStore {

	return &RedisStore{
		ClientsKey,
		PageKey,

		redis_host,
		redis_port,

		redisPool{
			connections: make(chan redis.Conn, 6),
			maxIdle:     6,

			connFn: func() (redis.Conn, error) {
				client, err := redis.Dial("tcp", fmt.Sprintf("%s:%v", redis_host, redis_port))
				if err != nil {
					log.Printf("Redis connect failed: %s\n", err.Error())
					return nil, err
				}

				return client, nil
			},
		},
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

	psc := redis.PubSubConn{conn}
	psc.Subscribe(channel)
	go func() {
		defer conn.Close()
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				c <- v.Data
			case redis.Subscription:
			case error:
				conn, err = this.GetConn()
				if err != nil {
					log.Println(err)
				}

				psc = redis.PubSubConn{conn}
				psc.Subscribe(channel)
			}
		}
	}()

	return conn, nil
}

func (this *RedisStore) Publish(channel string, message string) {
	publisher, err := this.GetConn()
	if err != nil {
		return
	}
	defer this.CloseConn(publisher)

	publisher.Do("PUBLISH", channel, message)
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
