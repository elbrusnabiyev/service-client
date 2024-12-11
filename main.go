// client/main.go
// установление соединения с сервером

package main

import (
	"context"
	"io"
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
	// Удалённое подключение к серверу.  Setting up a connection to the server.
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderManagementClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Add Order:
	order1 := pb.Order{Id: "101", Items: []string{"iPhone XS", "Mac Book Pro"}, Destination: "San Jose, CA", Price: 2300.00}
	res, _ := client.AddOrder(ctx, &order1)
	if res != nil {
		log.Print("AddOrder Response -> ", res.Value)
	}

	// Get Order:
	retrievedOrder, err := client.GetOrder(ctx, &pb.OrderID{Value: "106"})
	log.Print("GetOrder Response -> : ", retrievedOrder)

	// Search Order : Server streaming scenario
	searchStream, _ := client.SearchOrders(ctx, &pb.OrderID{Value: "Google"})
	for {
		searchOrder, err := searchStream.Recv()
		if err == io.EOF {
			log.Print("EOF")
			break
		}
		if err == nil {
			log.Print("Search Result : ", searchOrder)
		}
	}

	// UpdateOrders : Клиентский потоковый сценарий - Client streaming scenario
	updOrder1 := pb.Order{Id: "102", Items: []string{"google Pixel 3A", "Google Pixel Book"}, Destination: "Mountain View, CA", Price: 1100.00}
	updOrder2 := pb.Order{Id: "103", Items: []string{"Apple Watch S4", "iPad Pro"}, Destination: "San Jose, CA", Price: 2800.00}
	updOrder3 := pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub", "iPad Mini"}, Destination: "Mountain View, CA", Price: 2200.00}

	updateStream, err := client.UpdateOrders(ctx) // Вызов удалённого метода.

	if err != nil {
		log.Fatalf("%v.UpdateOrders(_) = _, %v", client, err)
	}

	// обновляем заказ 1 - Updating order 1
	if err := updateStream.Send(&updOrder1); err != nil { // Отправка обновленных заказов в клиентский поток
		log.Fatalf("%v.Send(%v) = %v", updateStream, updOrder1, err)
	}

	// обновляем заказ 2 - Updating order 2
	if err := updateStream.Send(&updOrder2); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updateStream, updOrder2, err)
	}

	// обновляем заказ 3 - Updating order 3
	if err := updateStream.Send(&updOrder3); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updateStream, updOrder3, err)
	}

	updateRes, err := updateStream.CloseAndRecv() // Закрытие поока и получение ответа.
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", updateStream, err, nil)
	}
	log.Printf("Update Orders Res : %s", updateRes)

	// ============================
	// Process Order. Двунаправленный потоковый сценарий - Bi-Directional Streaming scenatio
	streamProcOrder, err := client.ProcessOrders(ctx)
	if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = _, %v", client, err)
	}

	if err := streamProcOrder.Send(&pb.OrderID{Value: "102"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "102", err)
	}
	if err := streamProcOrder.Send(&pb.OrderID{Value: "103"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "103", err)
	}
	if err := streamProcOrder.Send(&pb.OrderID{Value: "104"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "104", err)
	}

	chanel := make(chan struct{})
	go asncClientBidirectionalRPC(streamProcOrder, chanel)
	time.Sleep(time.Millisecond * 1000)

	if err := streamProcOrder.Send(&pb.OrderID{Value: "101"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "101", err)
	}
	if err := streamProcOrder.CloseSend(); err != nil {
		log.Fatal(err)
	}
	chanel <- struct{}{}

}

func asncClientBidirectionalRPC(streamProcOrder pb.OrderManagement_ProcessOrdersClient, c chan struct{}) {
	for {
		combinedShipment, errProcOrder := streamProcOrder.Recv()
		if errProcOrder == io.EOF {
			break
		}
		log.Printf("Combined shipment : ", combinedShipment.OrdersList)
	}
	<-c
}
