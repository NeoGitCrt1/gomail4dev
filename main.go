package main

import (
	"flag"
	"sync"

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
	wg := sync.WaitGroup{}
	wg.Add(2)

	go mailserver.Serve(&wg)
	go webserver.Serve(&wg)

	wg.Wait()
}
