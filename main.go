package main

import (
	"flag"
	"fmt"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var result struct {
	Type    string
	StepNos string
	Content string
}

func main() {
	smtpHost := flag.String("smtpHost", "smtp.mail.me.com", "SMTP host")
	smtpPort := flag.Int("smtpPort", 587, "SMTP port")
	username := flag.String("username", "", "e-mail account username")
	password := flag.String("password", "", "e-mail account password")
	to := flag.String("to", "", "e-mail destination")
	port := flag.Int("port", 62115, "WebSocket server port")
	flag.Parse()

	u := url.URL{Scheme: "ws", Host: "localhost:" + strconv.Itoa(*port)}
	fmt.Println("Connecting to " + u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			err := c.ReadJSON(&result)
			if err != nil {
				fmt.Println(err)
				return
			}

			msg := ""
			switch result.Type {
			case "step":
				msg = "🍅 Go to the next step #" + result.StepNos
			case "finish":
				msg = "🍅 Finished! #" + result.StepNos
			}
			if len(msg) > 0 {
				fmt.Println("Send notification: " + msg)
				sendMail(*smtpHost, *smtpPort, *username, *password, *to, msg)
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		}
	}
}

func sendMail(smtpHost string, smtpPort int, username string, password string, to string, msg string) {
	toList := []string{
		to,
	}
	body := []byte("From: " + username + "\r\nTo: " + strings.Join(toList[:], ",") + "\r\n\r\n" + msg)

	auth := smtp.PlainAuth("", username, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+strconv.Itoa(smtpPort), auth, username, toList, body)
	if err != nil {
		fmt.Println(err)
		return
	}
}
