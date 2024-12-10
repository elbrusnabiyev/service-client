// client/main.go
// установление соединения с сервером

package main

import (
	"context"
	"log"
	"time"

	pb "github.com/elbrusnabiyev/service-client/ecommerceorder"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewOrderManagementClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Get Order:
	retrievedOrder, err := client.GetOrder(ctx, &pb.OrderID{Value: "106"})
	log.Print("GetOrder Response -> : ", retrievedOrder)

}
