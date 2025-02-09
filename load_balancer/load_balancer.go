package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	pb "Distributed_load_balancer/proto"

	"google.golang.org/grpc"
)

type LoadBalancer struct {
	pb.UnimplementedLoadBalancerServiceServer // Implementa la interfaz generada
	servers                                   []string
	mu                                        sync.Mutex
}

// Lee las direcciones de los servidores desde un archivo
func readServersFromFile(filename string) ([]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error al leer el archivo de servidores: %v", err)
	}

	// Divide el contenido en líneas y elimina espacios en blanco
	servers := strings.Split(strings.TrimSpace(string(content)), "\n")
	return servers, nil
}

// Obtiene la carga de un servidor
func getServerLoad(server string) (int32, error) {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	client := pb.NewLoadBalancerServiceClient(conn)
	res, err := client.GetLoad(context.Background(), &pb.LoadRequest{})
	if err != nil {
		return 0, err
	}
	return res.Load, nil
}

// Selecciona el servidor con menos carga
func (lb *LoadBalancer) selectServer() (string, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	var selectedServer string
	var minLoad int32 = 1 << 30 // Un número grande

	for _, server := range lb.servers {
		load, err := getServerLoad(server)
		if err != nil {
			log.Printf("Error al obtener carga del servidor %s: %v", server, err)
			continue
		}

		if load < minLoad {
			minLoad = load
			selectedServer = server
		}
	}

	if selectedServer == "" {
		return "", fmt.Errorf("no hay servidores disponibles")
	}

	return selectedServer, nil
}

// Procesa una solicitud del cliente
func (lb *LoadBalancer) ProcessRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	server, err := lb.selectServer()
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewLoadBalancerServiceClient(conn)
	return client.ProcessRequest(ctx, req)
}

func main() {
	// Leer las direcciones de los servidores desde el archivo
	servers, err := readServersFromFile("servers.txt")
	if err != nil {
		log.Fatalf("Error al leer las direcciones de los servidores: %v", err)
	}

	lb := &LoadBalancer{servers: servers}

	listener, err := net.Listen("tcp", ":4000")
	if err != nil {
		log.Fatalf("Error al iniciar el balanceador de carga: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterLoadBalancerServiceServer(s, lb)

	log.Printf("Balanceador de carga corriendo en el puerto :4000")
	if err := s.Serve(listener); err != nil {
		log.Fatalf("Error en el balanceador de carga: %v", err)
	}
}
