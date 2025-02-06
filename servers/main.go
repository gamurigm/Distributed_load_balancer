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

var servers []string
var nextServer = 0
var mu sync.Mutex

// Función para obtener el siguiente servidor de manera cíclica
func getNextServer() string {
	mu.Lock()
	server := servers[nextServer]
	nextServer = (nextServer + 1) % len(servers)
	mu.Unlock()
	return server
}

// Función para manejar la solicitud a un servidor
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
	// Leer el número de servidores desde la entrada de la consola
	fmt.Print("Ingresa el número de servidores a utilizar: ")
	var numServers int
	_, err := fmt.Scanf("%d", &numServers)
	if err != nil || numServers <= 0 {
		log.Fatalf("Número de servidores no válido: %v", err)
	}

	// Generar dinámicamente las direcciones de los servidores
	for i := 0; i < numServers; i++ {
		port := 50051 + i // Servidores en puertos consecutivos: 50051, 50052, ...
		serverAddr := fmt.Sprintf(":%d", port)
		servers = append(servers, serverAddr)
	}

	// Procesar 10 solicitudes
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
