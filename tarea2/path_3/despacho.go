package main

import (
	"encoding/json"
	"fmt"

	//"fmt"
	"log"
	pb "tarea2/protos"
	repos "tarea2/repositorio"

	amqp "github.com/rabbitmq/amqp091-go"
	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

func failOnError(err error, msg string) {
    if err != nil {
        log.Panicf("%s: %s", msg, err)
    }
}

type Order struct {
    Products   []Product  `json:"products"`
    Customer   Customer   `json:"customer"`
    Deliveries Delivery `json:"deliveries"`
}

type OrderResponse struct {
    OrderID    string        `json:"orderID"`
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

type Delivery struct {
    ShippingAddress   Location `json:"shippingAddress"`
    ShippingMethod    string   `json:"shippingMethod"`
    TrackingNumber    string   `json:"trackingNumber"`
}

var order Order
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
        "despacho-queue", // name
        false,           // durable
        false,           // delete when unused
        true,            // exclusive
        false,           // no-wait
        nil,             // arguments
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
            fmt.Println("Despacho recibido")

            // Decodificar el mensaje JSON
            var orderResponse OrderResponse
            if err := json.Unmarshal(d.Body, &orderResponse); err != nil {
                log.Printf("Error al decodificar el mensaje JSON: %v", err)
                continue
            }

            // Procesar el mensaje de despacho aquí
            // Puedes acceder a los datos del pedido en orderResponse

            //log.Printf("OrderID: %s", orderResponse.OrderID)

            order, err := repos.Read(orderResponse.OrderID)
            if err != nil {
                log.Printf("Error al cargar la orden desde MongoDB: %v", err)
                continue
            }
            
            delivery := &pb.Delivery{
                ShippingAddress: &pb.Customer{
                	Name:     orderResponse.Customer.Name,
                	LastName: orderResponse.Customer.LastName,
                	Email:    order.Customer.Email,
                	Location: order.Customer.Location,
                	Phone:    order.Customer.Phone,
                },
                ShippingMethod:  "USPS",
                TrackingNumber:  "12345678901234567890",
            }

            // Agregar el detalle de despacho a la orden
            // Debes tener acceso a la orden cargada desde MongoDB en esta parte del código
            // La variable "order" debe ser la orden que quieres actualizar
            repos.Update(orderResponse.OrderID,delivery)
            fmt.Println("Despacho adicionado")
        }
    }()

    log.Printf(" [*] Despacho is waiting for logs. To exit press CTRL+C")
    <-forever
}