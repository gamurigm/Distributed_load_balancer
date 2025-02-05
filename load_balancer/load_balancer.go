package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "Distributed_load_balancer/proto"
	"google.golang.org/grpc"
)

var servers = []string{":50051", ":50052"}
var nextServer = 0
var mu sync.Mutex

func getNextServer() string {
	mu.Lock()
	server := servers[nextServer]
	nextServer = (nextServer + 1) % len(servers)
	mu.Unlock()
	return server
}

func handleRequest(client pb.LoadBalancerServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.ProcessRequest(ctx, &pb.Request{Payload: "Heavy Task"})
	if err != nil {
		log.Fatalf("Error en la solicitud: %v", err)
	}
	fmt.Println("Respuesta:", res.Result)
}

func main() {
	for i := 0; i < 10; i++ {
		serverAddr := getNextServer()
		conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("No se pudo conectar al servidor %s: %v", serverAddr, err)
		}
		defer conn.Close()

		client := pb.NewLoadBalancerServiceClient(conn)
		handleRequest(client)
	}
}
