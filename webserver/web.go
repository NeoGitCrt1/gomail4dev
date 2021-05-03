package webserver

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/NeoGitCrt1/gomail4dev/dblink"
	"github.com/NeoGitCrt1/gomail4dev/mailparse"
	"github.com/NeoGitCrt1/gomail4dev/mailserver"
	"github.com/bdwilliams/go-jsonify/jsonify"
	"github.com/gin-gonic/gin"

	"github.com/DeanThompson/ginpprof"
)

type ServerOptions struct {
	BasePath string `json:"smtpServer"`
	Port     int    `json:"smtpPort"`
}

var opt *ServerOptions

var basePath string
var port int

func init() {
	flag.StringVar(&basePath, "base_path", "/", "base url part")
	flag.IntVar(&port, "web_port", 5000, "web site port")
}

func Serve(wg *sync.WaitGroup) {
	defer wg.Done()
	opt = &ServerOptions{
		basePath,
		port,
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	base := router.Group((*opt).BasePath + "/")
	{
		base.StaticFile("/", "./wwwroot/index.html")
		base.Static("/js", "./wwwroot/js")
		base.Static("/css", "./wwwroot/css")
		base.Static("/fonts", "./wwwroot/fonts")
		base.StaticFile("/logo.png", "./wwwroot/logo.png")
		api := base.Group("/api")
		{
			api.GET("/Messages", func(c *gin.Context) {
				r, err := dblink.Db.Query("select cast(id as text) id, [from], [to], receivedDate, subject, attachmentCount, isUnread from Message order by receivedDate desc")
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
				data := jsonify.Jsonify(r)
				c.String(http.StatusOK, "[%s]", strings.Join(data, ""))
			})
			api.DELETE("/Messages/:id", func(c *gin.Context) {
				id := c.Param("id")
				if id == "*" {
					dblink.Db.Exec("delete from Message", id)
				} else {
					dblink.Db.Exec("delete from Message where id=?", id)
				}
			})
			api.GET("/Messages/:id", func(c *gin.Context) {
				id := c.Param("id")
				r := dblink.Db.QueryRow("select [from], receivedDate, data from Message where id=?", id)
				dblink.Db.Exec("update Message set isUnread = 0 where id=? and isUnread = 1", id)
				var recv string
				var from string
				b := make([]byte, 0)
				err := r.Scan(&from, &recv, &b)
				if err != nil {
					c.String(http.StatusOK, "{}")
					return
				}
				msg, err := mailparse.ReadMailFromRaw(&b)
				if err != nil {
					c.String(http.StatusOK, "{}")
					return
				}
				msg.From = from
				msg.Recv = recv
				head := make([]kv, 0)
				for k := range msg.Head {
					head = append(head, kv{k, msg.Head[k][0]})
				}
	
				c.JSON(http.StatusOK, gin.H{
					"headers":          head,
					"subject":          msg.Subject,
					"to":               msg.To,
					"from":             msg.From,
					"id":               c.Param("id"),
					"receivedDate":     msg.Recv,
					"secureConnection": "false",
				})
			})
			api.GET("/Messages/:id/html", func(c *gin.Context) {
				r := dblink.Db.QueryRow("select data from Message where id=?", c.Param("id"))
				b := make([]byte, 0)
				err := r.Scan(&b)
				if err != nil {
					c.String(http.StatusOK, "{}")
					return
				}
				msg, err := mailparse.ReadMailFromRaw(&b)
				// for i := range *(msg.Parts) {
				// 	(*(msg.Parts) )[i].ContentType
				// }
				con, isPlain := msg.TextBody()
				if isPlain {
					c.String(http.StatusOK, "<pre>%s</pre>", con)
				} else {
					c.String(http.StatusOK, con)
				}
	
			})
			api.GET("/Messages/:id/raw", func(c *gin.Context) {
				r := dblink.Db.QueryRow("select data from Message where id=?", c.Param("id"))
				b := make([]byte, 0)
				err := r.Scan(&b)
				if err != nil {
					c.String(http.StatusOK, "{}")
					return
				}
				c.DataFromReader(http.StatusOK,
					int64(len(b)),
					"text/plain",
					bytes.NewReader(b),
					map[string]string{},
				)
	
			})
			api.GET("/Messages/:id/download", func(c *gin.Context) {
				r := dblink.Db.QueryRow("select data from Message where id=?", c.Param("id"))
				b := make([]byte, 0)
				err := r.Scan(&b)
				if err != nil {
					c.String(http.StatusOK, "%s", err.Error())
					return
				}
				//			body, _ := ioutil.ReadAll(msg.Body)
				c.DataFromReader(http.StatusOK,
					int64(len(b)),
					"application/octet-stream",
					bytes.NewReader(b),
					map[string]string{"Content-Disposition": "attachment;filename=" + c.Param("id") + ".eml"},
				)
	
			})
			api.GET("/Server", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"relayOptions": mailserver.ServerOptions,
					"isRunning": true,
				})
			})
	
		}
	}

	ginpprof.Wrap(router)

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa((*opt).Port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
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

type kv struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
