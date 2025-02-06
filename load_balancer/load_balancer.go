package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

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

// Server es la implementación de nuestro servicio de balanceo de carga
type server struct {
	pb.UnimplementedLoadBalancerServiceServer
}

// ProcessRequest maneja las solicitudes de los clientes, redirigiéndolas a los servidores correctos
func (s *server) ProcessRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	serverAddr := getNextServer()

	// Aquí puedes enviar la solicitud al servidor correspondiente, pero no estamos haciendo eso aquí.
	// Este código solo redirige la solicitud al servidor.
	return &pb.Response{Result: fmt.Sprintf("Solicitud enviada al servidor: %s", serverAddr)}, nil
}

func main() {

	// Iniciar un servidor gRPC
	lis, err := net.Listen("tcp", ":50000") // El balanceador escuchará en el puerto 50000
	if err != nil {
		log.Fatalf("No se pudo iniciar el servidor: %v", err)
	}

	// Crear un servidor gRPC
	grpcServer := grpc.NewServer()
	pb.RegisterLoadBalancerServiceServer(grpcServer, &server{})

	// Iniciar el servidor gRPC
	log.Println("Balanceador de carga en ejecución en el puerto :50040")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error al iniciar el servidor gRPC: %v", err)
	}
}
