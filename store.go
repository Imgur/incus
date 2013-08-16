package main

import (
    "log"
)

type Storage struct {
    memory      MemoryStore
    redis       RedisStore
    StorageType string
}

func initStore(Config *Configuration) Storage{
    store_type := "memory"
    var redisStore RedisStore
    
    redis_enabled := Config.Get("redis_enabled");
    if(redis_enabled == "true") {
        redis_host := Config.Get("redis_host")
        redis_port := uint(Config.GetInt("redis_port"))
        log.Println("%s, %v", redis_host, redis_port)
        
        redisStore = newRedisStore(redis_host, redis_port)
        store_type = "redis"
    }
    
    var Store = Storage{
        MemoryStore{make(map[string] *Socket), 0, make(map[string] *Page)},
        redisStore,
        
        store_type,
    }
    
    return Store
}

func (this *Storage) Save(UID string, s *Socket) (error) {
    this.memory.Save(UID, s)
    
    if this.StorageType == "redis" {
        if err := this.redis.Save(UID); err != nil {
            return err
        }
    }
    
    return nil
}

func (this *Storage) Remove(UID string) (error) {
    this.memory.Remove(UID)
    
    if this.StorageType == "redis" {
        if err := this.redis.Remove(UID); err != nil {
            return err
        }
    }
    
    return nil
}

func (this *Storage) Client(UID string) (*Socket, error) {
    return this.memory.Client(UID)
}

func (this *Storage) Clients() (map[string] *Socket) {
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

func (this *Storage) SetPage(UID string, page string) error {
    this.memory.SetPage(UID, page)
    
    if this.StorageType == "redis" {
        if err := this.redis.SetPage(UID, page); err != nil {
            return err
        }
    }
    
    return nil
}

func (this *Storage) UnsetPage(UID string, page string) error {
    this.memory.UnsetPage(UID, page)
    
    if this.StorageType == "redis" {
        if err := this.redis.UnsetPage(UID, page); err != nil {
            return err
        }
    }
    
    return nil
}

func (this *Storage) getPage(page string) *Page {
    return this.memory.getPage(page)
}
