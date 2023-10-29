package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

type OrderResponse struct {
	OrderID    string        `json:"orderID"`
	GroupID    string        `json:"groupID"`
	Products   []Product     `json:"products"`
	Customer   Customer      `json:"customer"`
}

type Product struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Genre       string  `json:"genre"`
	Pages       int     `json:"pages"`
	Publication string  `json:"publication"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type Customer struct {
	Name     string      `json:"name"`
	LastName string      `json:"lastname"`
	Email    string      `json:"email"`
	Location Location    `json:"location"`
	Phone    string      `json:"phone"`
}

type Location struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

func sendEmail(orderResponse OrderResponse) error {
	// Convierte la estructura OrderResponse a JSON
	emailData, err := json.Marshal(orderResponse)
	if err != nil {
		return err
	}

	// Realiza una solicitud HTTP POST a la API de AWS
	url := "https://sjwc0tz9e4.execute-api.us-east-2.amazonaws.com/Prod"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(emailData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Verifica el código de respuesta
	if resp.StatusCode != http.StatusOK {
		return errors.New("Error en la solicitud HTTP")
	}

	return nil
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		"logs", // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			//log.Printf(" [x] Received: %s", d.Body)

			// Decodificar el mensaje JSON
			var orderResponse OrderResponse
			if err := json.Unmarshal(d.Body, &orderResponse); err != nil {
				log.Printf("Error al decodificar el mensaje JSON: %v", err)
				continue
			}

			// Establecer GroupID
			orderResponse.GroupID = "9G#1f8S2m!"

			// Enviar un correo
			err := sendEmail(orderResponse)
			if err != nil {
				log.Printf("Error al enviar el correo: %v", err)
			} else {
				log.Printf("Correo enviado")
			}
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit, press CTRL+C")
	<-forever
}