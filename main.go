package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/NeoGitCrt1/gomail4dev/dblink"
	"github.com/NeoGitCrt1/gomail4dev/mailserver"
	"github.com/NeoGitCrt1/gomail4dev/webserver"
)

func main() {

	flag.IntVar(&mailserver.MPort, "smtp_port", 25, "smtp server port")
	flag.StringVar(&webserver.BasePath, "base_path", "/", "base url part")
	flag.IntVar(&webserver.WPort, "web_port", 5000, "web site port")
	flag.StringVar(&dblink.DbData, "data_path", "./data.db", "data file path")
	flag.Parse()
	
	dblink.InitDb()
	go mailserver.Serve()
	go webserver.Serve()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c
	os.Exit(0)
}
