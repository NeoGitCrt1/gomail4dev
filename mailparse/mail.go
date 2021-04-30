package mailparse

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"mime/quotedprintable"
	"net/mail"
)

// https://github.com/kirabou/parseMIMEemail.go
type Mail struct {
	To string
	From string
	Subject string
	Head mail.Header
	Parts *[]*Part
}

type Part struct {
	ContentType string
	enc string
	data *[]byte
	isAttach bool
}

func (m *Mail) TextBody() (string , bool){
	for k := range *m.Parts {
		p := (*m.Parts)[k]
		if (!p.isAttach) {
			return p.ContentString() , p.ContentType == "text/plain"
		}
	}
	return "" , true
}


func (p *Part) ContentString() string{
	switch p.enc {
	case "BASE64":
		c, err := base64.StdEncoding.DecodeString(string(*p.data))
		if err != nil {
			return string(*p.data)
		}
		return string(c)

	case "QUOTED-PRINTABLE":
		c, err := ioutil.ReadAll(quotedprintable.NewReader(bytes.NewReader(*p.data)))
		if err != nil {
			return string(*p.data)
		} 
		return string(c)
	default:
		return string(*p.data)
	}
}