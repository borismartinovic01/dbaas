package main

import (
	"net"
	"net/http"
	"net/rpc"
)

type App struct{}

func main() {
	app := &App{}
	PubSubServer := NewPubSub()
	rpc.Register(PubSubServer)
	rpc.HandleHTTP()
	app.listenRPC()
}

func (app *App) listenRPC() error {
	listen, err := net.Listen("tcp", ":3000")
	if err != nil {
		return nil
	}
	defer listen.Close()

	http.Serve(listen, nil)

	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}

		go rpc.ServeConn(conn)
	}
}
