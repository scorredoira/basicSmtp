package basicSmtp

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"net"
	"net/smtp"
)

// Send connects to server without checking TLS and sends the email message.
func Send(subject, body, server, user, password, from string, to []string, html bool) error {
	c, err := smtp.Dial(server)
	host, _, err := net.SplitHostPort(server)
	if err != nil {
		return err
	}

	tlc := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	if err = c.StartTLS(tlc); err != nil {
		return err
	}

	a := &unEncryptedAuth{user, password}
	if err = c.Auth(a); err != nil {
		return err
	}

	if err := c.Mail(from); err != nil {
		return err
	}

	for _, rcp := range to {
		if err := c.Rcpt(rcp); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	data := messageData(subject, body, from, html)
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	err = c.Quit()
	if err != nil {
		return err
	}

	return nil
}

func messageData(subject, body, from string, html bool) []byte {
	var buf bytes.Buffer
	buf.WriteString("From: ")
	buf.WriteString(from)

	// allow utf-8 characters in the subject as specified in RFC 1342
	buf.WriteString("\nSubject: =?utf-8?B?")
	buf.WriteString(base64.StdEncoding.EncodeToString([]byte(subject)))
	buf.WriteString("?=")

	buf.WriteString("\nMIME-version: 1.0;")
	buf.WriteString("\nContent-Type: text/")
	buf.WriteString(contentType(html))
	buf.WriteString("; charset=\"UTF-8\";\n\n")
	buf.WriteString(body)
	return buf.Bytes()
}

func contentType(html bool) string {
	if html {
		return "text/html"
	}
	return "text/plain"
}

// unEncryptedAuth uses the given username and password to authenticate
// without checking a TLS connection or host like smtp.PlainAuth does.
type unEncryptedAuth struct {
	username, password string
}

func (a *unEncryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	resp := []byte("\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *unEncryptedAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		// We've already sent everything.
		return nil, errors.New("unexpected server challenge")
	}

	return nil, nil
}
