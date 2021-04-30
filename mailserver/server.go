package mailserver

import (
	"net"
	"strconv"
	"strings"
	"sync/atomic"
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
	
}

var MPort int

func Serve() {
	ServerOptions = &ServerRelayOptions{
		SmtpServer: "localhost",
		SmtpPort:   MPort,
	}
	smtpd.ListenAndServe(":"+strconv.Itoa(MPort), mailHandler, "gomail", "")
}

var count uint32

func mailHandler(origin net.Addr, from string, to []string, data []byte) error {
	stmt, err := dblink.Db.Prepare("INSERT INTO Message ( id, [from], [to], subject,receivedDate, data, isUnread, mimeParseError, attachmentCount ) values (?,?,?,?,?,?,?,?,?)")
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
	// not for high traffic scenario
	c := atomic.AddUint32(&count, 1)
	if (c > 10) {
		atomic.StoreUint32(&count, 0)
		// I konw this delete sucks. I have old data with UUID as id, so I have to do this in this way
		dblink.Db.Exec("delete from Message where id not in (select id from Message order by receivedDate desc limit 1000 )")
	}
	
	return err
}
