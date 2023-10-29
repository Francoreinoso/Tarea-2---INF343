package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	ventas "tarea2/protos" // Asegúrate de que esta importación esté configurada correctamente

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Product struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Genre       string  `json:"genre"`
	Pages       int     `json:"pages"`
	Publication string  `json:"publication"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type Location struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

type Customer struct {
	Name       string   `json:"name"`
	LastName   string   `json:"lastname"`
	Email      string   `json:"email"`
	Location   Location `json:"location"`
	Phone      string   `json:"phone"`
}

type Data struct {
	Products []Product `json:"products"`
	Customer  Customer `json:"customer"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: ", os.Args[0], " <archivo.json>")
		return
	}

	// Carga el archivo JSON y decodifica los datos en productData.
	filePath := os.Args[1]
	jsonData, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error al abrir el archivo JSON: %v", err)
		return
	}


	// Decodifica el archivo JSON en una estructura de datos.
	var productData Data
	if err := json.Unmarshal(jsonData, &productData); err != nil {
		fmt.Printf("Error al analizar el JSON: %v\n", err)
		return
	}

	// Establece la conexión gRPC con el servidor.
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials())) // Reemplaza localhost y el puerto si es necesario.
	if err != nil {
		fmt.Printf("Error al conectar: %v", err)
		return
	}
	defer conn.Close()

	// Crea un cliente gRPC
	client := ventas.NewVentasServiceClient(conn)

	// Convierte los datos al formato esperado por el servicio gRPC
	grpcProducts := make([]*ventas.Product, 0)
	for _, p := range productData.Products {
		grpcProducts = append(grpcProducts, &ventas.Product{
			Title:       p.Title,
			Author:      p.Author,
			Genre:       p.Genre,
			Pages:       int32(p.Pages),
			Publication: p.Publication,
			Quantity:    int32(p.Quantity),
			Price:       float32(p.Price),
		})
	}

	grpcCustomer := &ventas.Customer{
		Name:     productData.Customer.Name,
		LastName: productData.Customer.LastName,
		Email:    productData.Customer.Email,
		Location: &ventas.Location{
			Address1:   productData.Customer.Location.Address1,
			Address2:   productData.Customer.Location.Address2,
			City:       productData.Customer.Location.City,
			State:      productData.Customer.Location.State,
			PostalCode: productData.Customer.Location.PostalCode,
			Country:    productData.Customer.Location.Country,
		},
		Phone: productData.Customer.Phone,
	}
	
	grpcData := &ventas.Data{
		Products: grpcProducts,
		Customer: grpcCustomer,
	}

	// Llama al servicio SendProducts con los datos en el formato gRPC.
	response, err := client.SendData(context.Background(), grpcData)
	if err != nil {
    	fmt.Printf("Error al llamar al servicio SendData: %v\n", err)
    	return
	}

	if response != nil {
		fmt.Printf("Orden ID recibido: %s\n", response.OrderId)
	} else {
		fmt.Println("Respuesta vacía o sin OrderId.")
	}

	//fmt.Println("Productos enviados exitosamente al servidor gRPC.")
}