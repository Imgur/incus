package main

import (
    //"errors"
    "log"
    
    "menteslibres.net/gosexy/redis"
)

const ClientsKey = "SocketClients"
const CountKey   = "SocketClientsCount"

type RedisStore struct {
    clientCount int
}

//func (this *RedisStore) GetConn() (redis.Conn) {
//    //if this.Pool.ActiveCount() == 0 {
//    //    c, err := redis.Dial("tcp", ":6379")
//    //    if err != nil {
//    //        return nil, errors.New("Couldn't connect to redis :(")
//    //    }
//    //    
//    //    return c, nil
//    //}
//    
//    return this.Conn
//}

func (this *RedisStore) Save(UID string, s *Socket) (bool, error) {
    client = redis.New()
    err = client.Connect("localhost", "6379")

    if err != nil {
        log.Fatalf("Connect failed: %s\n", err.Error())
        return
    }
 
    _, err1 := client.SAdd(ClientsKey, UID)
    if err1 != nil {
        log.Printf("%s\n", err1.Error())
    }
    
    client.Close()
    return true, nil
}
//
//func (this *RedisStore) Remove(UID string) (bool, error) {
//    exists, _ := this.GetConn().Do("Hexists", redis.Args{}.Add(ClientsKey).Add(UID))
//    
//    //delete(m.clients, UID)
//    
//    if(exists) { // only subtract if the client was in the store in the first place.
//        this.clientCount--
//    }
//    
//    return true, nil
//}
//
//func (this *RedisStore) Client(UID string) (*Socket, error) {
//    exists, _ := this.GetConn().Do("Hexists", redis.Args{}.Add(ClientsKey).Add(UID))
//    
//    //if(!exists) {
//    //    return nil, errors.New("ClientID doesn't exist")
//    //}
//    
//    return nil, nil
//}
//
//func (m *RedisStore) Clients() (map[string] *Socket, error) {
//    return nil, nil
//}
//
//func (m *RedisStore) Count() (int, error) {
//    return 1, nil
//}
