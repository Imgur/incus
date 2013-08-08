package main

import (

)

type Configuration struct {
    conf string
}

func initConfig() Configuration {
    return Configuration{"TEST"}
}