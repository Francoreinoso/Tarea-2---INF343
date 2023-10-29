package main

import (
	"encoding/json"
	"fmt"
	"log"
	pb "tarea2/protos"
	repos "tarea2/repositorio"

	amqp "github.com/rabbitmq/amqp091-go"
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
    Pages       int32     `json:"pages"`
    Publication string  `json:"publication"`
    Quantity    int32     `json:"quantity"`
    Price       float32 `json:"price"`
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
        "inventario-queue", // name
        false,             // durable
        false,             // delete when unused
        true,              // exclusive
        false,             // no-wait
        nil,               // arguments
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
            //log.Printf(" [x] Inventario Received: %s", d.Body)

            // Decodificar el mensaje JSON
            var orderResponse OrderResponse
            if err := json.Unmarshal(d.Body, &orderResponse); err != nil {
                log.Printf("Error al decodificar el mensaje JSON: %v", err)
                continue
            }
            
            for _,product := range orderResponse.Products{
                productData:= &pb.Product{
                    Title: product.Title,
                    Author: product.Author,
                    Genre: product.Genre,
                    Pages: product.Pages,
                    Publication: product.Publication,
                    Quantity: product.Quantity,
                    Price: product.Price,
                }
                
                
                repos.ReducirInv(productData)

            }
            // Procesar el mensaje de inventario aquí
            // Puedes acceder a los datos del pedido en orderResponse

            //log.Printf("OrderID: %s", orderResponse.OrderID)
            //log.Printf("GroupID: %s", orderResponse.GroupID)
            fmt.Println("Inventario reducido")
        }
    }()

    log.Printf(" [*] Inventario is waiting for logs. To exit press CTRL+C")
    <-forever
}