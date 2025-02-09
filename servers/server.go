package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"sync/atomic"
	"time"

	pb "Distributed_load_balancer/proto"

	"google.golang.org/grpc"
)

// Estructura del servidor que implementa el servicio de balanceo de carga
type Server struct {
	pb.UnimplementedLoadBalancerServiceServer
	activeLoads  int32  // Carga actual del servidor (solicitudes activas)
	totalHandled int32  // Total de solicitudes manejadas
	port         string // Puerto del servidor
}

// Función para manejar la carga del servidor y devolver la carga actual
func (s *Server) GetLoad(ctx context.Context, req *pb.LoadRequest) (*pb.LoadResponse, error) {
	currentLoad := atomic.LoadInt32(&s.activeLoads)
	log.Printf("[Server %s] Reportando carga actual: %d", s.port, currentLoad)
	return &pb.LoadResponse{Load: currentLoad}, nil
}

// Función para procesar solicitudes y guardar la respuesta en un archivo CSV
func (s *Server) ProcessRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	// Aumentar carga activa
	atomic.AddInt32(&s.activeLoads, 1)
	defer atomic.AddInt32(&s.activeLoads, -1) // Disminuir carga al final

	// Simulando procesamiento de la solicitud
	log.Printf("[Server %s] Procesando solicitud %d", s.port, req.WorkId)

	// Simular resultado de la solicitud
	result := fmt.Sprintf("Resultado de trabajo %d", req.WorkId)

	// Obtener la carga activa actual
	currentLoad := atomic.LoadInt32(&s.activeLoads)

	// Registrar en CSV
	go s.saveToCSV(req.WorkId, result, currentLoad)

	// Incrementar el contador de solicitudes manejadas
	atomic.AddInt32(&s.totalHandled, 1)

	return &pb.Response{Result: result}, nil
}

// Función para guardar los resultados en un archivo CSV de manera concurrente
func (s *Server) saveToCSV(workId int32, result string, load int32) {
	// Abrir archivo CSV
	file, err := os.OpenFile("responses.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error al abrir el archivo CSV: %v", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Si el archivo está vacío, escribir encabezado
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Error al obtener información del archivo CSV: %v", err)
		return
	}

	if fileInfo.Size() == 0 {
		writer.Write([]string{"Timestamp", "TrabajoID", "Resultado", "CargaActiva"})
	}

	// Escribir en el archivo CSV
	record := []string{
		time.Now().Format("2006-01-02 15:04:05"),
		fmt.Sprintf("%d", workId),
		result,
		fmt.Sprintf("%d", load),
	}

	if err := writer.Write(record); err != nil {
		log.Printf("Error al escribir en el archivo CSV: %v", err)
	}
}

func main() {
	// Verificar argumentos
	if len(os.Args) < 2 {
		log.Fatal("Debes proporcionar un puerto (ejemplo: ./server :50051)")
	}
	port := os.Args[1]

	// Asegurar que el puerto comience con ':'
	if port[0] != ':' {
		port = ":" + port
	}

	// Crear e inicializar el servidor
	server := &Server{
		activeLoads:  0,
		totalHandled: 0,
		port:         port,
	}

	// Iniciar el servidor gRPC
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error al iniciar el servidor en puerto %s: %v", port, err)
	}

	// Crear un servidor gRPC
	s := grpc.NewServer()
	pb.RegisterLoadBalancerServiceServer(s, server)

	// Log de inicio del servidor
	log.Printf("Servidor iniciado en puerto %s (Carga inicial: 0)", port)

	// Iniciar el servidor
	if err := s.Serve(listener); err != nil {
		log.Fatalf("Error en el servidor %s: %v", port, err)
	}
}
