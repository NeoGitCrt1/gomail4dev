package main

import (
	"flag"
	"sync"

	"github.com/NeoGitCrt1/gomail4dev/dblink"
	"github.com/NeoGitCrt1/gomail4dev/mailserver"
	"github.com/NeoGitCrt1/gomail4dev/webserver"
)

func init() {
}

func main() {
	flag.Parse()

	dblink.InitDb()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go mailserver.Serve(&wg)
	go webserver.Serve(&wg)

	wg.Wait()
}
