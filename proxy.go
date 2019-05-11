package main

import (
	"fmt"
	"strings"
	"net/http"
	"encoding/json"
	"github.com/fiorix/go-eventsocket/eventsocket"
)

const (
	addr = "134.209.167.66"
	port = 8020
	eslURL = "127.0.0.1:8021"
	eslPWD = "ClueCon"
)

type Broker struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Request string `json:"request"`
	Response string `json:"response"`
	Authorized bool `json:"authorized"`
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

func sendESL(b *Broker) {
	c, err := eventsocket.Dial(eslURL, eslPWD)
	check(err)
	c.Send("event json ALL")
	c.Send(fmt.Sprintf("bgapi user_data %s@%s param password", b.Username, addr))
	for {
		e, err := c.ReadEvent()
		check(err)
		if strings.Contains(e.Get("Job-Command"), "user_data") {
			if e.Body == b.Password {
				break
			} else {
				c.Close()
				return
			}
		}
	}
	c.Send(fmt.Sprintf("bgapi user_data %s@%s var esl_use", b.Username, addr))
	for {
		e, err := c.ReadEvent()
		check(err)
		if strings.Contains(e.Get("Job-Command"), "user_data") {
			if strings.Contains(e.Body, strings.SplitN(b.Request," ",2)[0]) {
				b.Authorized = true
				break
			} else {
				c.Close()
				return
			}
		}
	}
	c.Send(fmt.Sprintf("bgapi %s", b.Request))
	for {
		e, err := c.ReadEvent()
		check(err)
		if strings.Contains(e.Get("Job-Command"), strings.SplitN(b.Request," ",2)[0]) {
			b.Response = e.Body
			c.Close()
			break
		}
	}
}

func httpESL (w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	b := &Broker{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		Request: r.FormValue("request"),
	}
	sendESL(b)
	j, err := json.Marshal(b)
	check(err)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(j))
}


func main() {
	http.HandleFunc("/esl/", httpESL)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", addr, port), nil); err != nil {
		fmt.Println("ListenAndServe Error:", err)
	}
}
