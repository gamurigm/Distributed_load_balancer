package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
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

func heavyComputationPart(start, end, port int, wg *sync.WaitGroup, resultChan chan<- float64) {
	defer wg.Done()
	result := 0.0
	for i := start; i < end; i++ {
		// Operaciones matemáticas intensivas con el número de puerto como parámetro
		sqrt := math.Sqrt(float64(i))
		log := math.Log(float64(i + 1))
		sin := math.Sin(float64(i))
		cos := math.Cos(float64(i))
		tan := math.Tan(float64(i))
		exp := math.Exp(float64(i) / float64(port)) // Utiliza el puerto en el cálculo exponencial
		result += sqrt*log + sin*cos + tan*exp
	}
	resultChan <- result
}

func heavyComputation(port int) float64 {
	numCPU := runtime.NumCPU() // Obtener la cantidad de núcleos de CPU disponibles
	runtime.GOMAXPROCS(numCPU) // Configurar Go para usar todos los núcleos

	// Dividir la carga de trabajo en varias partes
	numGoroutines := numCPU
	chunkSize := port / numGoroutines
	var wg sync.WaitGroup
	resultChan := make(chan float64, numGoroutines)

	// Llamar a las goroutines
	for i := 0; i < numGoroutines; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numGoroutines-1 {
			end = port // El último fragmento toma el resto
		}
		wg.Add(1)
		go heavyComputationPart(start, end, port, &wg, resultChan)
	}

	wg.Wait()
	close(resultChan)

	// Combina los resultados de todas las goroutines
	var totalResult float64
	for result := range resultChan {
		totalResult += result
	}
	return totalResult
}

// ProcessRequest maneja las solicitudes de los clientes.
func (s *server) ProcessRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	mu.Lock()
	requestCount++
	currentRequest := requestCount
	mu.Unlock()

	// Eliminar el prefijo ":" del puerto
	portNumStr := strings.TrimPrefix(port, ":")
	portNum, err := strconv.Atoi(portNumStr) // Convertir el puerto sin ":"
	if err != nil {
		log.Fatalf("Error al convertir el puerto: %v", err)
	}

	start := time.Now()
	computationResult := heavyComputation(portNum) // Cálculo pesado dependiente del puerto
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
	log.Printf("GOMAXPROCS (núcleos utilizados): %d", runtime.GOMAXPROCS(0))
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
