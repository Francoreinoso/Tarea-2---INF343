package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	//db "tarea2/database"
	pb "tarea2/protos" // Importa el paquete generado a partir de la definición del servicio gRPC
	repos "tarea2/repositorio"

	//"google.golang.org/protobuf/proto"
	"google.golang.org/grpc"
)

type ventasServer struct {
	pb.UnimplementedVentasServiceServer
}

var order *repos.Order

var products []*pb.Product

func (s *ventasServer) SendData(ctx context.Context, req *pb.Data) (*pb.OrderResponse, error) {
	// Aquí puedes procesar tanto los productos como los datos del cliente recibidos a través de gRPC
	// Por ejemplo, puedes guardarlos en una base de datos o realizar operaciones comerciales.
	// Imprime los productos recibidos
	fmt.Println("Orden recibida.")
	for _, product := range req.Products {
		productData:= &pb.Product{
			Title: product.Title,
			Author: product.Author,
			Genre: product.Genre,
			Pages: product.Pages,
			Publication: product.Publication,
			Quantity: product.Quantity,
			Price: product.Price,
		}
		products= append(products, productData)
	 }
	// Imprime los datos del cliente recibidos

	order := &repos.Order{
		Products: products,
		Customer: req.Customer,
	}
	orderId,err := repos.Create(order)
	if err != nil {
		fmt.Println("Error al insertar el documento en MongoDB:", err)
	}
	// La ID generada por MongoDB para la orden
	response := &pb.OrderResponse{
        OrderId: orderId,
        Products: req.Products,
        Customer: req.Customer,
        Deliveries: []*pb.Delivery{
            {
                ShippingAddress: req.Customer,
                ShippingMethod: "USPS",
                TrackingNumber: "12345678901234567890",
            },
        },
    }

	// Ahora puedes devolver la orderID como respuesta a la solicitud gRPC
	fmt.Printf("Document inserted with ID: %s\n", orderId)

	// Puedes realizar otras operaciones aquí según tus necesidades.
	sendToRabbitMQ(response)

	// Retorna una respuesta vacía.
	return response, nil
}

func sendToRabbitMQ(response *pb.OrderResponse) {
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

    message, err := json.Marshal(response)
    if err != nil {
        log.Printf("Failed to marshal response: %v", err)
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err = ch.PublishWithContext(ctx,
        "logs", // exchange
        "",     // routing key
        false,  // mandatory
        false,  // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        message,
        })
    failOnError(err, "Failed to publish a message")

    //log.Printf(" [x] Sent %s", message)
}

func failOnError(err error, msg string) {
	if err != nil {
			log.Panicf("%s: %s", msg, err)
	}
}


func main() {
	
	listen, err := net.Listen("tcp", "localhost:50051") // Reemplaza localhost y el puerto si es necesario.
	if err != nil {
		log.Fatalf("Error al escuchar: %v", err)
	}
	server := grpc.NewServer()
	pb.RegisterVentasServiceServer(server, &ventasServer{})

	log.Printf("Servidor gRPC de Ventas escuchando en :50051")
	if err := server.Serve(listen); err != nil {
		log.Fatalf("Error al servir: %v", err)
	}
}

func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
			s = "hello"
	} else {
			s = strings.Join(args[1:], " ")
	}
	return s
}