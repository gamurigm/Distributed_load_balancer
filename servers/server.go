package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync/atomic"

	pb "Distributed_load_balancer/proto"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedLoadBalancerServiceServer       // Implementa la interfaz generada
	load                                      int32 // Carga actual del servidor
}

// ProcessRequest maneja las solicitudes de procesamiento de trabajo
func (s *Server) ProcessRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	atomic.AddInt32(&s.load, 1)        // Incrementar la carga
	defer atomic.AddInt32(&s.load, -1) // Decrementar la carga al finalizar

	// Simulación de procesamiento de trabajo
	result := fmt.Sprintf("Trabajo %d procesado con éxito", req.WorkId)
	return &pb.Response{Result: result}, nil
}

// GetLoad devuelve la carga actual del servidor
func (s *Server) GetLoad(ctx context.Context, req *pb.LoadRequest) (*pb.LoadResponse, error) {
	load := atomic.LoadInt32(&s.load) // Obtener la carga actual
	return &pb.LoadResponse{Load: load}, nil
}

func main() {
	// Leer el puerto desde los argumentos de la línea de comandos
	if len(os.Args) < 2 {
		log.Fatal("Debes proporcionar un puerto (ejemplo: ./server :50051)")
	}
	port := os.Args[1]

	// Iniciar el servidor gRPC
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}

	// Crear un servidor gRPC
	s := grpc.NewServer()
	pb.RegisterLoadBalancerServiceServer(s, &Server{})

	// Log de inicio del servidor
	log.Printf("Servidor corriendo en el puerto %s", port)

	// Iniciar el servidor
	if err := s.Serve(listener); err != nil {
		log.Fatalf("Error en el servidor: %v", err)
	}
}
