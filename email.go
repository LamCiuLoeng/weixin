package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/mail"
	"net/smtp"
	"path"
	"strings"
)

func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

func encodeStr(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func encodeTitle(str string) string {
	// "=?UTF-8?X?  X -> B for Base64, Q for Quoted-Printable
	return fmt.Sprintf("=?UTF-8?b?%s?=", encodeStr(str))
}

func createTextMsg(from string, to string, cc string, subject string, content string) string {
	hdr := make(map[string]string)
	hdr["From"] = from
	hdr["To"] = to
	hdr["Cc"] = cc
	hdr["Subject"] = encodeTitle(subject)
	hdr["MIME-Version"] = "1.0"
	hdr["Content-Type"] = "text/html; charset=UTF-8"
	hdr["Content-Transfer-Encoding"] = "base64"

	msg := ""
	for k, v := range hdr {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += fmt.Sprintf("\r\n%s", encodeStr(content))
	return msg
}

func createMultipartMsg(from string, to string, cc string, subject string, body string, files []string) (string, error) {
	msg := ""
	boundary := randomBoundary()
	msg += "MIME-Version: 1.0"
	msg += fmt.Sprintf("\r\nFrom: %s", from)
	msg += fmt.Sprintf("\r\nTo: %s", to)
	msg += fmt.Sprintf("\r\nCc: %s", cc)
	msg += fmt.Sprintf("\r\nSubject: %s", encodeTitle(subject))
	msg += fmt.Sprintf("\r\nContent-Type: multipart/mixed; boundary=%s", boundary)

	//add the email body
	msg += fmt.Sprintf("\r\n\r\n--%s", boundary)
	msg += "\r\nContent-Type: text/html; charset=UTF-8"
	msg += "\r\nContent-Transfer-Encoding: base64"
	msg += fmt.Sprintf("\r\n\r\n%s", encodeStr(body))

	//add the attachment
	for _, f := range files {
		filename := path.Base(f)
		content, err := ioutil.ReadFile(f)
		if err != nil {
			return "", err
		}
		encodedContent := encodeStr(string(content))
		msg += fmt.Sprintf("\r\n--%s", boundary)
		msg += "\r\nContent-Type: application/octet-stream"
		msg += "\r\nMIME-Version: 1.0"
		msg += "\r\nContent-Transfer-Encoding: base64"
		msg += fmt.Sprintf("\r\nContent-Disposition: attachment; filename=\"%s\"", encodeTitle(filename))
		msg += fmt.Sprintf("\r\n\r\n%s", encodedContent)
	}

	msg += fmt.Sprintf("\r\n--%s--", boundary)
	return msg, nil

}

func SendEmail(from string,
	to []string,
	cc []string,
	subject string,
	body string,
	files []string,
	serverip string,
	port string,
	auth smtp.Auth) error {

	emailfrom := mail.Address{from, from}
	var sendto, emailto, emailcc []string
	for _, t := range to {
		tmp := mail.Address{"", t}
		emailto = append(emailto, tmp.String())
		sendto = append(sendto, tmp.Address)
	}

	for _, c := range cc {
		tmp := mail.Address{"", c}
		emailcc = append(emailcc, tmp.String())
		sendto = append(sendto, tmp.Address)
	}

	var msg string
	var err error
	if len(files) < 1 {
		msg = createTextMsg(emailfrom.String(),
			strings.Join(emailto, ","),
			strings.Join(emailcc, ","),
			subject,
			body)
	} else {
		msg, err = createMultipartMsg(emailfrom.String(),
			strings.Join(emailto, ","),
			strings.Join(emailcc, ","),
			subject,
			body,
			files)
		if err != nil {
			return err
		}
	}

	srv := serverip + ":" + port
	emailerr := smtp.SendMail(srv, auth, emailfrom.Address, sendto, []byte(msg))
	return emailerr
}
