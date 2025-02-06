package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"sync"
	"time"

	pb "Distributed_load_balancer/proto"

	"google.golang.org/grpc"
)

var (
	requestCount int        // Contador de solicitudes
	mu           sync.Mutex // Mutex para proteger el contador
)

type server struct {
	pb.UnimplementedLoadBalancerServiceServer
}

func heavyComputation(n int) float64 {
	result := 0.0
	for i := 1; i < n; i++ {
		// Operaciones matemáticas intensivas
		sqrt := math.Sqrt(float64(i))
		log := math.Log(float64(i + 1))
		sin := math.Sin(float64(i))
		cos := math.Cos(float64(i))
		tan := math.Tan(float64(i))
		exp := math.Exp(float64(i) / 1000000) // Exponencial para mayor intensidad
		result += sqrt*log + sin*cos + tan*exp
	}
	return result
}

// ProcessRequest maneja las solicitudes de los clientes.
func (s *server) ProcessRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	mu.Lock()
	requestCount++
	currentRequest := requestCount
	mu.Unlock()

	start := time.Now()
	computationResult := heavyComputation(905587) // Cálculo pesado
	elapsed := time.Since(start)

	log.Printf("Servidor en puerto %s procesó solicitud #%d en %s", port, currentRequest, elapsed)

	return &pb.Response{Result: fmt.Sprintf("Server [%s]: Computo = %.2f, Tiempo = %s", port, computationResult, elapsed)}, nil
}

var port string

func main() {
	// Leer el puerto desde los argumentos de la línea de comandos
	if len(os.Args) < 2 {
		log.Fatal("Debes proporcionar un puerto (ejemplo: ./server :50051)")
	}
	port = os.Args[1]

	// Iniciar el servidor gRPC
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterLoadBalancerServiceServer(s, &server{})

	log.Printf("Servidor corriendo en el puerto %s", port)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("Error en el servidor: %v", err)
	}
}
