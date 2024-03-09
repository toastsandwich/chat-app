package main

import (
	"os"

	"github.com/toastsandwich/chat-app/v1.1.0/server/pkg/server"
)

func main() {
	addr := os.Args[1] + ":" + os.Args[2]
	app := server.NewServer(addr)
	app.Start()
}
