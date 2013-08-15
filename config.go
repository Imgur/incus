package main

import (
    "log"
    "fmt"

    "github.com/briankassouf/cfg"
)

type Configuration struct {
    vars map[string]string
}

func initConfig() Configuration {
    mymap := make(map[string]string)
    err := cfg.Load("sockets.conf", mymap)
    if err != nil {
        log.Fatal(err)
    }

    return Configuration{mymap}    
}

func (this *Configuration) Get(name string) (string, error) {
    val, ok := this.vars[name]
    if !ok {
        return "", fmt.Errorf("Config Error: variable '%s' not found", name)
    }
 
    return val, nil;
}
