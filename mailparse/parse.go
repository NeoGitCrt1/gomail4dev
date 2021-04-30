package mailparse

// reference
// https://gist.github.com/tejainece/9940151
// https://github.com/kirabou/parseMIMEemail.go

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"strings"
)

const (
	Enc = "Content-Transfer-Encoding"
	Tp  = "Content-Type"
)

func ReadMailFromRaw(data *[]byte) (m *Mail, err error) {
	return ReadMail(bytes.NewReader(*data))
}

func ReadMail(r io.Reader) (m *Mail, err error) {
	m = nil
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return
	}

	mediaType, params, _ := mime.ParseMediaType(msg.Header.Get(Tp))

	parts := make([]*Part, 0)
	if !strings.HasPrefix(mediaType, "multipart/") {
		buf := new(bytes.Buffer)
		buf.ReadFrom(msg.Body)
		data := buf.Bytes()
		parts = append(parts, &Part{mediaType, strings.ToUpper(msg.Header.Get(Enc)), &data, false})
	} else {
		reader := multipart.NewReader(msg.Body, params["boundary"])
		for {

			newPart, err := reader.NextPart()
			if err == io.EOF {
				err = nil
				break
			}
			if err != nil {
				break
			}

			// Do something with the newPart being processed.
			// newPart can itself be a nested new multipart part,
			// requiring some kind of recursive processing.
			partData, err := ioutil.ReadAll(newPart)
			if err != nil {
				continue
			}

			mediaType, _, err := mime.ParseMediaType(newPart.Header.Get(Tp))
			if err != nil {
				continue
			}

			parts = append(parts, &Part{mediaType, strings.ToUpper(newPart.Header.Get(Enc)), &partData, !(mediaType == "text/plain" || mediaType == "text/html")})
		}
	}

	dec := new(mime.WordDecoder)
	from, _ := dec.DecodeHeader(msg.Header.Get("From"))
	to, _ := dec.DecodeHeader(msg.Header.Get("To"))
	subject, _ := dec.DecodeHeader(msg.Header.Get("Subject"))
	m = &Mail{
		To:      to,
		From:    from,
		Subject: subject,
		Head:    msg.Header,
		Parts:   &parts,
	}

	return

}

func getMediaType(h textproto.MIMEHeader) (mediaType string, params map[string]string, err error) {
	return mime.ParseMediaType(h.Get(Tp))
}
