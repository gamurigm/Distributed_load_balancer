package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	pb "Distributed_load_balancer/proto" // Asegúrate de que la ruta del paquete sea correcta

	"google.golang.org/grpc"
)

type LoadBalancer struct {
	pb.UnimplementedLoadBalancerServiceServer
	servers []string
	mu      sync.Mutex
}

type ServerLoad struct {
	address string
	load    int32
	err     error
}

var csvMutex sync.Mutex

// Lee la lista de servidores desde el archivo
func readServersFromFile(filename string) ([]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error al leer el archivo de servidores: %v", err)
	}
	servers := strings.Split(strings.TrimSpace(string(content)), "\n")
	return servers, nil
}

// Obtiene la carga de un servidor específico
func (lb *LoadBalancer) getServerLoad(server string) (int32, error) {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return 0, fmt.Errorf("error al conectar con servidor %s: %v", server, err)
	}
	defer conn.Close()

	client := pb.NewLoadBalancerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := client.GetLoad(ctx, &pb.LoadRequest{})
	if err != nil {
		return 0, fmt.Errorf("error al obtener carga de %s: %v", server, err)
	}
	return res.Load, nil
}

// Selecciona el servidor con menor carga
func (lb *LoadBalancer) selectServer() (string, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Canal para recibir las cargas de los servidores
	loadChan := make(chan ServerLoad, len(lb.servers))

	// Obtener la carga de cada servidor de forma concurrente
	var wg sync.WaitGroup
	for _, server := range lb.servers {
		wg.Add(1)
		go func(serverAddr string) {
			defer wg.Done()
			load, err := lb.getServerLoad(serverAddr)
			loadChan <- ServerLoad{
				address: serverAddr,
				load:    load,
				err:     err,
			}
		}(server)
	}

	// Esperar en una goroutine separada a que terminen todas las consultas
	go func() {
		wg.Wait()
		close(loadChan)
	}()

	// Encontrar el servidor con menor carga
	var selectedServer string
	minLoad := int32(1<<31 - 1) // Máximo valor de int32
	availableServers := 0

	for serverLoad := range loadChan {
		if serverLoad.err == nil {
			availableServers++
			log.Printf("Servidor %s tiene carga: %d", serverLoad.address, serverLoad.load)
			if serverLoad.load < minLoad {
				minLoad = serverLoad.load
				selectedServer = serverLoad.address
			}
		} else {
			log.Printf("Error al obtener carga de %s: %v", serverLoad.address, serverLoad.err)
		}
	}

	if availableServers == 0 {
		return "", fmt.Errorf("no hay servidores disponibles")
	}

	log.Printf("Seleccionado servidor %s con carga %d", selectedServer, minLoad)
	return selectedServer, nil
}

// Procesa la solicitud de un cliente
func (lb *LoadBalancer) ProcessRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	log.Printf("Recibida solicitud para trabajo %d", req.WorkId)

	server, err := lb.selectServer()
	if err != nil {
		return nil, fmt.Errorf("error al seleccionar servidor: %v", err)
	}

	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("error al conectar con servidor %s: %v", server, err)
	}
	defer conn.Close()

	client := pb.NewLoadBalancerServiceClient(conn)
	res, err := client.ProcessRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error al procesar solicitud en servidor %s: %v", server, err)
	}

	log.Printf("Respuesta del servidor %s: %s", server, res.Result)

	// Guardar la respuesta en un archivo CSV
	go func() {
		csvMutex.Lock()
		defer csvMutex.Unlock()

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
			writer.Write([]string{"Timestamp", "TrabajoID", "Servidor", "Carga", "Resultado"})
		}

		// Escribir en el archivo CSV
		record := []string{
			time.Now().Format("2006/01/02 15:04:05"), // Fecha en formato YYYY/MM/DD HH:MM:SS
			fmt.Sprintf("%d", req.WorkId),            // ID del trabajo
			server,                                   // Servidor
			//fmt.Sprintf("%d", carga),                 // Carga del servidor
			res.Result, // Resultado del trabajo
		}

		if err := writer.Write(record); err != nil {
			log.Printf("Error al escribir en el archivo CSV: %v", err)
		}
	}()

	return res, nil
}

func main() {
	// Leer la lista de servidores desde el archivo
	servers, err := readServersFromFile("servers.txt")
	if err != nil {
		log.Fatalf("Error al leer las direcciones de los servidores: %v", err)
	}

	log.Printf("Servidores cargados: %v", servers)
	lb := &LoadBalancer{servers: servers}

	// Crear un servidor GRPC
	listener, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatalf("Error al iniciar el balanceador de carga: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterLoadBalancerServiceServer(s, lb)

	// Iniciar el servidor
	log.Printf("Balanceador de carga corriendo en el puerto :4000")
	if err := s.Serve(listener); err != nil {
		log.Fatalf("Error en el balanceador de carga: %v", err)
	}
}
