package webserver

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/NeoGitCrt1/gomail4dev/dblink"
	"github.com/NeoGitCrt1/gomail4dev/mailparse"
	"github.com/NeoGitCrt1/gomail4dev/mailserver"
	"github.com/bdwilliams/go-jsonify/jsonify"
	"github.com/gin-gonic/gin"
)

func Serve() {
	router := gin.Default()
	router.StaticFile("/", "./wwwroot/index.html")
	router.Static("/js", "./wwwroot/js")
	router.Static("/css", "./wwwroot/css")
	router.Static("/fonts", "./wwwroot/fonts")
	router.StaticFile("/logo.png", "./wwwroot/logo.png")

	api := router.Group("/api")
	{
		api.GET("/Messages", func(c *gin.Context) {
			r, err := dblink.Db().Query("select id, [from], [to], receivedDate, subject, attachmentCount from Message order by receivedDate desc")
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
				dblink.Db().Exec("delete from Message", id)
			} else {
				dblink.Db().Exec("delete from Message where id=?", id)
			}
		})
		api.GET("/Messages/:id", func(c *gin.Context) {
			r := dblink.Db().QueryRow("select [from], receivedDate, data from Message where id=?", c.Param("id"))
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
			//body, _ := ioutil.ReadAll(msg.Body)

			head := make([]kv, 0)
			for k := range msg.Head {
				head = append(head, kv{k, msg.Head[k][0]})
			}

			c.JSON(http.StatusOK, gin.H{"headers": head,
				"subject":          msg.Subject,
				"to":               msg.To,
				"from":             from,
				"id":               c.Param("id"),
				"receivedDate":     recv,
				"secureConnection": "false",
			})
		})
		api.GET("/Messages/:id/html", func(c *gin.Context) {
			r := dblink.Db().QueryRow("select data from Message where id=?", c.Param("id"))
			b := make([]byte, 0)
			err := r.Scan(&b)
			if err != nil {
				c.String(http.StatusOK, "{}")
				return
			}
			msg, err := mailparse.ReadMailFromRaw(&b)

			c.String(http.StatusOK, msg.TextBody())

		})
		api.GET("/Messages/:id/raw", func(c *gin.Context) {
			r := dblink.Db().QueryRow("select data from Message where id=?", c.Param("id"))
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
			r := dblink.Db().QueryRow("select data from Message where id=?", c.Param("id"))
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

	router.Run(":5000")

}

type kv struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
