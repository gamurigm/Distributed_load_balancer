package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "Distributed_load_balancer/proto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar al balanceador: %v", err)
	}
	defer conn.Close()

	client := pb.NewLoadBalancerServiceClient(conn)

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res, err := client.ProcessRequest(ctx, &pb.Request{Payload: "Tarea pesada"})
		if err != nil {
			log.Fatalf("Error en la solicitud: %v", err)
		}

		fmt.Println("Respuesta:", res.Result)
	}
}
