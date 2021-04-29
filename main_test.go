package main

import (
	"net/smtp"
	"testing"
	_ "mime/multipart"
	"time"
)

func TestRepeatAce(t *testing.T) {
	
	// Choose auth method and set it up
	//auth := smtp.PlainAuth("", "piotr@mailtrap.io", "extremely_secret_pass", "smtp.mailtrap.io")

	// Here we do it all: connect to our server, set up a message and send it
	to := []string{"bill@gates.com"}
	msg := []byte("To: bill@gates.com\r\n" +
		"Subject: 汉字Why are you not using Mailtrap yet?" + time.Now().String() + "\r\n" +
		"\r\n" +
		"Here’s the 汉字space for our great sales pitch\r\n")
	err := smtp.SendMail(":25", nil, "piotr@mailtrap.io", to, msg)
	if err != nil {
		t.Error(err)
	}
}
