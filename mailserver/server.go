package mailserver

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
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

func init() {
	ServerOptions = &ServerRelayOptions{
		SmtpServer: "localhost",
		SmtpPort:   *mPort,
	}
}

var mPort = flag.Int("smtp_port", 25, "smtp server port")

func Serve() {
	fmt.Println("start mail tcp:")
	smtpd.ListenAndServe(":" + strconv.Itoa(*mPort), mailHandler, "gomail", "")
}
func mailHandler(origin net.Addr, from string, to []string, data []byte) error {
	stmt, err := dblink.Db().Prepare("INSERT INTO Message ( id, [from], [to], subject,receivedDate, data, isUnread, mimeParseError, attachmentCount ) values (?,?,?,?,?,?,?,?,?)")
	m, err := mailparse.ReadMailFromRaw(&data)
	aCnt := 0
	mimeParseError := ""
	if err != nil {
		mimeParseError = err.Error()
	}
	aCnt = len(*m.Parts) - 1
	stmt.Exec(strconv.FormatUint(snowflake.ID(), 10), from, strings.Join(to, ","), m.Subject,
		time.Now(),
		data, 1, mimeParseError,
		aCnt,
	)
	stmt.Close()
	return err
}
