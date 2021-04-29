package mailserver

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	m "net/mail"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/NeoGitCrt1/go-snowflake"
	"github.com/NeoGitCrt1/gomail4dev/dblink"
	"github.com/mhale/smtpd"
)

type ServerRelayOptions struct {
	BasePath        string
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
		SmtpPort:   25,
	}
}

func Serve() {
	fmt.Println("start mail tcp:")
	smtpd.ListenAndServe(":25", mailHandler, "gomail", "")

}
func mailHandler(origin net.Addr, from string, to []string, data []byte) error {
	stmt, err := dblink.Db().Prepare("INSERT INTO Message ( id, [from], [to], receivedDate, subject, data, mimeParseError, sessionId, attachmentCount, isUnread ) values (?,?,?,?,?,?,?,?,?,?)")
	msg, attachmentCount, mimeParseError := readMsg(bytes.NewReader(data))
	_, err = stmt.Exec(strconv.FormatUint(snowflake.ID(), 10), from, strings.Join(to, ","), time.Now(),
		msg.Header.Get("Subject"),
		data, mimeParseError, "", attachmentCount, 0,
	)
	stmt.Close()
	return err
}

func readMsg(r io.Reader) (mail *m.Message, attachmentCount int, mimeParseError string) {
	attachmentCount = 0
	mimeParseError = ""
	mail, err := m.ReadMessage(r)
	if err != nil {
		mimeParseError = err.Error()
		return
	}
	mediaType, params, err := getMediaType(textproto.MIMEHeader(mail.Header))
	if err != nil {
		mimeParseError = err.Error()
		return
	}
	if !strings.HasPrefix(mediaType, "multipart/") {
		return
	}

	mr := multipart.NewReader(mail.Body, params["boundary"])
	for {
		p, err := mr.NextPart()
		switch err {
		case nil:
			// Carry on
		case io.EOF:
			break
		default:
			mimeParseError = err.Error()
			return
		}

		mediaType, _, err := getMediaType(p.Header)

		if err != nil {
			mimeParseError = err.Error()
			return
		}

		if mediaType == "text/plain" || mediaType == "text/html" {
			// Carry on
		} else {
			attachmentCount++
		}
	}
}

func getMediaType(h textproto.MIMEHeader) (mediaType string, params map[string]string, err error) {
	return mime.ParseMediaType(h.Get("Content-Type"))
}
