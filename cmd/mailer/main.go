package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/Hickar/gin-rush/internal/broker"
	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/mailer"
)

func main() {
	if len(os.Args) < 1 {
		log.Fatalf("mailer worker initialization error: no configuration file was provided")
	}

	conf := config.NewConfig(os.Args[1])

	mailClient, err := mailer.NewMailer(&conf.Gmail)
	if err != nil {
		log.Fatalf("mailer setup error: %s", err)
	}

	conn, err := broker.NewBroker(&conf.RabbitMQ)
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn broker.Broker) {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conn)

	messages, err := conn.Consume("mailer_ex", "topic", "mailer")
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})

	go func() {
		for d := range messages {
			var msg mailer.ConfirmationMessage

			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Fatalf("unable to decode queue message: %s", err)
			}

			err = mailClient.SendConfirmationCode(msg.Username, msg.Email, msg.Code)
			if err != nil {
				log.Fatalf("unable to send confirmation message: %s", err)
			}
		}
	}()

	<-done
	log.Println("Waiting for new messages...")
}