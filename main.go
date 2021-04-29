package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"github.com/NeoGitCrt1/gomail4dev/mailserver"
	"github.com/NeoGitCrt1/gomail4dev/webserver"
)

func main() {

	go mailserver.Serve()
	
	go webserver.Serve()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <- c
	fmt.Println("退出:", s)
	os.Exit(0)
}