package mailserver

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/NeoGitCrt1/go-snowflake"
	"github.com/NeoGitCrt1/gomail4dev/dblink"
	"github.com/NeoGitCrt1/gomail4dev/mailparse"
	"github.com/mhale/smtpd"
)

type ServerRelayOptions struct {
	SmtpServer      string   `json:"smtpServer"`
	SmtpPort        int      `json:"smtpPort"`
	AutomaticEmails []string `json:"automaticEmails"`
	SenderAddress   string   `json:"senderAddress"`
	Login           string   `json:"login"`
	Password        string   `json:"password"`
}

var ServerOptions *ServerRelayOptions
var cleanSql string
func init() {
	flag.IntVar(&port, "smtp_port", 25, "smtp server port")
	flag.IntVar(&maxRecord, "smtp_max", 1000, "smtp max record")
}

var port int
var maxRecord int

func Serve(wg *sync.WaitGroup) {
	defer wg.Done()
	ServerOptions = &ServerRelayOptions{
		SmtpServer: "localhost",
		SmtpPort:   port,
	}
	cleanSql = "delete from Message where id < (select id from Message order by id desc limit "+ strconv.Itoa(maxRecord) + ", 1 )"
	srv := &smtpd.Server{Addr: ":" + strconv.Itoa(port), Handler: mailHandler, Appname: "gomail", Hostname: ""}
	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, smtpd.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

}

var count uint32

func mailHandler(origin net.Addr, from string, to []string, data []byte) (err error) {
	stmt, err := dblink.Db.Prepare("INSERT INTO Message ( id, [from], [to], subject,receivedDate, data, isUnread, mimeParseError, attachmentCount ) values (?,?,?,?,?,?,?,?,?)")
	m, err := mailparse.ReadMailFromRaw(&data)
	aCnt := 0
	mimeParseError := ""
	if err != nil {
		mimeParseError = err.Error()
	} 
	aCnt = len(*m.Parts) - 1
	_ , err = stmt.Exec(snowflake.ID(), from, strings.Join(to, ","), m.Subject,
		time.Now(),
		data, 1, mimeParseError,
		aCnt,
	)
	stmt.Close()
	
	c := atomic.AddUint32(&count, 1)
	if c == 11 {
		dblink.Db.Exec(cleanSql)
		atomic.StoreUint32(&count, 0)
	}
	return
}
