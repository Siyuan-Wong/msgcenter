package main

import (
	"msgcenter/server"
)

func main() {
	exitLoader()
	server.GetServer().Start()
}
