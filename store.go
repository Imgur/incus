package main

import (
    "log"    
    "menteslibres.net/gosexy/redis"
)

type Storage struct {
    memory      MemoryStore
    redis       RedisStore
    StorageType string
}

func initStore(Config *Configuration) Storage{
    var Store = Storage{
        MemoryStore{make(map[string]*Socket), 0},
        RedisStore{
            ClientsKey,
            
            "localhost",
            6379,
            
            redisPool{
                connections: []*redis.Client{},
                maxIdle:     6,
                
                connFn:      func () (*redis.Client, error) {
                    client := redis.New()
                    err := client.Connect("localhost", 6379)
                    
                    if err != nil {
                        log.Fatalf("Connect failed: %s\n", err.Error())
                        return nil, err
                    }
                    
                    return client, nil
                },
            },
            
        },
        
        "redis",
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
