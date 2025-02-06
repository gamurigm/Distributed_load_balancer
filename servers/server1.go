package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"sync"
	"time"

	pb "Distributed_load_balancer/proto"

	"google.golang.org/grpc"
)

var requestCount int
var mu sync.Mutex

type server struct {
	pb.UnimplementedLoadBalancerServiceServer
}

func heavyComputation(n int) float64 {
	result := 0.0
	for i := 1; i < n; i++ {
		result += math.Sqrt(float64(i)) * math.Log(float64(i+1)) * math.Pow(float64(i), 2)
	}
	return result
}

func (s *server) ProcessRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	mu.Lock()
	requestCount++
	mu.Unlock()

	start := time.Now()
	computationResult := heavyComputation(5000000) // Cálculo pesado
	elapsed := time.Since(start)

	log.Printf("Servidor 1 procesó solicitud #%d en %s", requestCount, elapsed)

	return &pb.Response{Result: fmt.Sprintf("Server1 [%d]: Computo = %.2f, Tiempo = %s", requestCount, computationResult, elapsed)}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50049")
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterLoadBalancerServiceServer(s, &server{})

	log.Println("Servidor 1 corriendo en el puerto 50049")
	if err := s.Serve(listener); err != nil {
		log.Fatalf("Error en el servidor: %v", err)
	}
}
