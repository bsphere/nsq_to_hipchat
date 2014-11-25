package main

import (
	"flag"
	"github.com/andybons/hipchat"
	nsq "github.com/bitly/go-nsq"
	"log"
	"strconv"
	"sync"
	"time"
)

func main() {
	var lookupd, topic, room, from, token, color string

	flag.StringVar(&lookupd, "lookupd", "http://127.0.0.1:4161", "lookupd HTTP address")
	flag.StringVar(&topic, "topic", "", "NSQD topic")
	flag.StringVar(&room, "room", "", "HipChat room")
	flag.StringVar(&from, "from", "nsq_to_hipchat", "HipChat announcement user name")
	flag.StringVar(&color, "color", "purple", "Message color: yellow, red, green, purple, gray, random")
	flag.StringVar(&token, "token", "", "HipChat auth token")
	flag.Parse()

	if lookupd == "" || topic == "" || room == "" || from == "" || token == "" {
		flag.PrintDefaults()
	}

	if lookupd == "" || topic == "" || room == "" || from == "" || color == "" ||
		token == "" {
		log.Fatal("invalid options")
	}

	h := hipchat.Client{AuthToken: token}

	channel := "nsq_to_hipchat" + strconv.FormatInt(time.Now().Unix(), 10) + "#ephemeral"
	c, err := nsq.NewConsumer(topic, channel, nsq.NewConfig())
	if err != nil {
		log.Fatal(err)
	}

	c.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		req := hipchat.MessageRequest{
			RoomId:        room,
			From:          from,
			Message:       string(m.Body),
			Color:         color,
			MessageFormat: hipchat.FormatText,
			Notify:        false,
		}

		if err := h.PostMessage(req); err != nil {
			log.Print(err)
		}

		m.Finish()

		return nil
	}))

	if err := c.ConnectToNSQLookupd(lookupd); err != nil {
		log.Fatal(err)
	}

	req := hipchat.MessageRequest{
		RoomId:        room,
		From:          from,
		Message:       "nsq_to_hipchat announcing events from topic '" + topic + "' to this room.",
		Color:         color,
		MessageFormat: hipchat.FormatText,
		Notify:        false,
	}

	if err := h.PostMessage(req); err != nil {
		log.Print(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
