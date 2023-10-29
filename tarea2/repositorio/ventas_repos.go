package user_repository

import (
	//"log"
	"context"
	"log"
	"tarea2/database"
	pb "tarea2/protos"

	//"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Order struct {
	Products []*pb.Product `bson:"products"`
    Customer *pb.Customer   `bson:"customer"`
    Deliveries *pb.Delivery `json:"deliveries"`
}

var err error

func Create(order *Order) (string, error) {
    collection := database.GetCollection("orders")

    result, err := collection.InsertOne(context.Background(), order)
    if err != nil {
        return "", err
    }

    // La ID generada por MongoDB para la orden
    orderID := result.InsertedID.(primitive.ObjectID).Hex()

    return orderID, err
}


func Read (orderID string) (*Order,error){
        collection := database.GetCollection("orders")
        objectID, err := primitive.ObjectIDFromHex(orderID)
        if err != nil {
        log.Fatal(err)
        }


        // Definir el filtro para buscar la orden por su orderID
        filter := bson.M{"_id": objectID}
    
        // Declara una variable para almacenar la orden
        var order Order
        // Realiza la consulta a la base de datos
        err = collection.FindOne(context.TODO(), filter).Decode(&order)
        if err != nil {
            // Manejar el error si la orden no se encontró
            return nil, err
        }
    
        // Imprimir la orden encontrada
        //fmt.Printf("Orden encontrada: %+v\n", order)
        return &order,nil
    
}

func Update (orderID string, deliveries *pb.Delivery) {
    collection := database.GetCollection("orders")
    objectID, err := primitive.ObjectIDFromHex(orderID)
	// Define un filtro para encontrar el documento que deseas actualizar
	filter := bson.M{"_id": objectID} 

	// Define la actualización
	update := bson.M{
        "$set": bson.M{
            "deliveries": deliveries, // Reemplaza con el nuevo valor
        },
    }
	// Realiza la operación de actualización
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
        return
	}
}

func ReducirInv (producto *pb.Product) {
    collection := database.GetCollection("products")

    // Define un filtro para encontrar el documento que deseas actualizar
    filter := bson.M{"title": producto.Title}

    // Realiza una consulta para buscar el producto por título
    var product pb.Product
    err := collection.FindOne(context.TODO(), filter).Decode(&product)

    if err != nil {
        if err == mongo.ErrNoDocuments {
            // El producto no existe en la base de datos, crea uno nuevo con stock inicial de 100.
            newProduct := &pb.Product{
            	Title:       producto.Title,
            	Author:      producto.Author,
            	Genre:       producto.Genre,
            	Pages:       producto.Pages,
            	Publication: producto.Publication,
            	Quantity:    100,
            	Price:       producto.Price,
            }

            _, err := collection.InsertOne(context.TODO(), newProduct)
            if err != nil {
                log.Fatal(err)
                return
            }
        } else {
            log.Fatal(err)
            return
        }
        update := bson.M{
            "$inc": bson.M{"quantity": -producto.Quantity},
        }
        _, err := collection.UpdateOne(context.TODO(), filter, update)
        if err != nil {
            log.Fatal(err)
            return
        }
    }
    
}


// func LoadOrder (orderID string) (*Order, error) {
    


//     if err != nil {
//         return nil, err
//     }

//     return &order, nil
// }