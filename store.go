package incus

import (
	"sync"

	"github.com/spf13/viper"
)

type Storage struct {
	memory      *MemoryStore
	redis       *RedisStore
	StorageType string

	userMu sync.RWMutex
	pageMu sync.RWMutex
}

func NewStore(stats RuntimeStats) *Storage {
	storeType := "memory"
	var redisStore *RedisStore

	if viper.GetBool("redis_enabled") {
		redisHost := viper.GetString("redis_port_6379_tcp_addr")
		redisPort := viper.GetInt("redis_port_6379_tcp_port")
		connPoolSize := viper.GetInt("redis_connection_pool_size")
		numConsumers := viper.GetInt("redis_activity_consumers")

		redisStore = newRedisStore(redisHost, redisPort, numConsumers, connPoolSize, stats)
		storeType = "redis"
	}

	var Store = Storage{
		&MemoryStore{make(map[string]map[string]*Socket), make(map[string]map[string]*Socket), make(map[string]map[string]*Socket), 0},
		redisStore,
		storeType,

		sync.RWMutex{},
		sync.RWMutex{},
	}

	return &Store
}

func (this *Storage) Save(sock *Socket) error {
	this.userMu.Lock()
	this.memory.Save(sock)
	this.userMu.Unlock()

	if this.StorageType == "redis" {
		if err := this.redis.Save(sock); err != nil {
			return err
		}
	}

	return nil
}

func (this *Storage) Remove(sock *Socket) error {
	this.userMu.Lock()
	this.memory.Remove(sock)
	this.userMu.Unlock()

	if this.StorageType == "redis" {
		if err := this.redis.Remove(sock); err != nil {
			return err
		}
	}

	return nil
}

func (this *Storage) Client(UID string) (map[string]*Socket, error) {
	defer this.userMu.RUnlock()
	this.userMu.RLock()

	return this.memory.Client(UID)
}

func (this *Storage) Clients() map[string]map[string]*Socket {
	defer this.userMu.RUnlock()
	this.userMu.RLock()

	return this.memory.Clients()
}

func (this *Storage) ClientList() ([]string, error) {
	if this.StorageType == "redis" {
		return this.redis.Clients()
	}

	return nil, nil
}

func (this *Storage) Count() (int64, error) {
	if this.StorageType == "redis" {
		return this.redis.Count()
	}

	return this.memory.Count()
}

func (this *Storage) SetPage(sock *Socket) error {
	this.pageMu.Lock()
	this.memory.SetPage(sock)
	this.pageMu.Unlock()

	if this.StorageType == "redis" {
		if err := this.redis.SetPage(sock); err != nil {
			return err
		}
	}

	return nil
}

func (this *Storage) UnsetPage(sock *Socket) error {
	this.pageMu.Lock()
	this.memory.UnsetPage(sock)
	this.pageMu.Unlock()

	if this.StorageType == "redis" {
		if err := this.redis.UnsetPage(sock); err != nil {
			return err
		}
	}

	return nil
}

func (this *Storage) getPage(page string) map[string]*Socket {
	defer this.pageMu.RUnlock()
	this.pageMu.RLock()
	return this.memory.getPage(page)
}

func (this *Storage) SetGroups(sock *Socket) error {
	this.pageMu.Lock()
	this.memory.SetGroups(sock)
	this.pageMu.Unlock()

	if this.StorageType == "redis" {
		if err := this.redis.SetGroups(sock); err != nil {
			return err
		}
	}

	return nil
}

func (this *Storage) UnsetGroups(sock *Socket) error {
	this.pageMu.Lock()
	this.memory.UnsetGroups(sock)
	this.pageMu.Unlock()

	if this.StorageType == "redis" {
		if err := this.redis.UnsetPage(sock); err != nil {
			return err
		}
	}

	return nil
}

func (this *Storage) getGroups(groups []string) []map[string]*Socket {
	defer this.pageMu.RUnlock()
	this.pageMu.RLock()
	return this.memory.getGroups(groups)
}
