package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	pb "Distributed_load_balancer/proto"

	"google.golang.org/grpc"
)

var servers []string // Esta lista se llenará dinámicamente con los servidores
var nextServer = 0
var mu sync.Mutex

// getNextServer obtiene el siguiente servidor de manera circular
func getNextServer() string {
	mu.Lock()
	server := servers[nextServer]
	nextServer = (nextServer + 1) % len(servers)
	mu.Unlock()
	return server
}

// handleRequest maneja las solicitudes de cliente y realiza reintentos si la conexión falla
func handleRequest(client pb.LoadBalancerServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var res *pb.Response
	var err error
	for i := 0; i < 3; i++ { // Intentar 3 veces
		res, err = client.ProcessRequest(ctx, &pb.Request{Payload: "Heavy Task"})
		if err == nil {
			break
		}
		log.Printf("Error en la solicitud, intento %d: %v", i+1, err)
		time.Sleep(time.Second) // Esperar un segundo antes de reintentar
	}

	if err != nil {
		log.Fatalf("Error en la solicitud: %v", err)
	}
	fmt.Println("Respuesta:", res.Result)
}

func main() {
	// Leer el número de servidores desde los argumentos
	if len(os.Args) < 2 {
		log.Fatalf("Debes proporcionar el número de servidores a utilizar como argumento")
	}
	numServers, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Error al convertir el número de servidores: %v", err)
	}

	// Llenar la lista de servidores con los puertos correspondientes
	for i := 0; i < numServers; i++ {
		serverAddr := fmt.Sprintf(":500%02d", 51+i)
		servers = append(servers, serverAddr)
	}

	// Hacer solicitudes a los servidores
	for i := 0; i < 25; i++ {
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
