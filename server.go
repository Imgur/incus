package main

type Server struct {
    Config  *Configuration
    Store   *Storage
}

func main() {
    conf  := initConfig()
    store := initStore()
    //initLogger()
    server := Server{&conf, &store}
    
    server.initAppListner()
    go server.initSocketListener()
    
}

func (this *Server) initSocketListener() {
    
}