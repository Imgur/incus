package main

import (
    "log"
    "fmt"
    "strconv"

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

func (this *Configuration) Get(name string) string {
    val, ok := this.vars[name]
    if !ok {
        panic(fmt.Sprintf("Config Error: variable '%s' not found", name))
    }
 
    return val;
}

func (this *Configuration) GetInt(name string) int {
    val, ok := this.vars[name]
    if !ok {
        panic(fmt.Sprintf("Config Error: variable '%s' not found", name))
    }
 
    i, err := strconv.Atoi(val)
    if err != nil {
        panic(fmt.Sprintf("Config Error: '%s' could not be cast as an int", name))
    }
    
    return i
}