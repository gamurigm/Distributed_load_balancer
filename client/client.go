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

var (
	servers    []string   // Lista de servidores
	nextServer = 0        // Índice del siguiente servidor
	mu         sync.Mutex // Mutex para acceder de manera segura a la lista de servidores
)

// getNextServer obtiene el siguiente servidor de manera circular
func getNextServer() string {
	mu.Lock()
	server := servers[nextServer]
	nextServer = (nextServer + 1) % len(servers)
	mu.Unlock()
	return server
}

// handleRequest maneja las solicitudes de cliente
func handleRequest(client pb.LoadBalancerServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
		log.Printf("Error final en la solicitud después de 3 intentos: %v", err)
		return
	}
	fmt.Println("Respuesta del servidor:", res.Result)
}

// initServers inicializa la lista de servidores desde los argumentos
func initServers(numServers int) {
	for i := 0; i < numServers; i++ {
		serverAddr := fmt.Sprintf(":500%02d", 51+i) // Asume que los servidores están en puertos consecutivos
		servers = append(servers, serverAddr)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Debes proporcionar el número de servidores como argumento")
	}
	numServers, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Error al convertir el número de servidores: %v", err)
	}

	// Inicializar la lista de servidores
	initServers(numServers)

	// Crear y ejecutar solicitudes concurrentemente
	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			serverAddr := getNextServer()
			conn, err := grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock()) // Conexión bloqueante hasta que esté lista
			if err != nil {
				log.Printf("No se pudo conectar al servidor %s: %v", serverAddr, err)
				return
			}
			defer conn.Close()

			client := pb.NewLoadBalancerServiceClient(conn)
			handleRequest(client)
		}()
	}

	wg.Wait() // Espera que todas las solicitudes terminen
}
