package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"

	pb "Distributed_load_balancer/proto"

	"google.golang.org/grpc"
)

// sendRequest envía una solicitud al balanceador de carga
func sendRequest(client pb.LoadBalancerServiceClient, workId int32, wg *sync.WaitGroup) {
	defer wg.Done()

	// Crear una solicitud con el ID de trabajo
	req := &pb.Request{WorkId: workId}

	// Enviar la solicitud al balanceador de carga
	res, err := client.ProcessRequest(context.Background(), req)
	if err != nil {
		log.Printf("Error al procesar la solicitud %d: %v", workId, err)
		return
	}

	// Imprimir la respuesta del servidor
	log.Printf("Trabajo %d - Respuesta del servidor: %s", workId, res.Result)
}

func main() {
	// Verificar que se proporcione el número de clientes como argumento
	if len(os.Args) < 2 {
		log.Fatal("Debes proporcionar el número de clientes (ejemplo: ./client 10)")
	}

	// Convertir el argumento a un número entero
	numClients, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Número de clientes inválido: %v", err)
	}

	// Conectar con el balanceador de carga
	conn, err := grpc.Dial("localhost:4000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error al conectar con el balanceador de carga: %v", err)
	}
	defer conn.Close()

	// Crear un cliente gRPC
	client := pb.NewLoadBalancerServiceClient(conn)

	// Usar un WaitGroup para esperar a que todas las goroutines terminen
	var wg sync.WaitGroup

	// Iniciar los clientes
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go sendRequest(client, int32(i+1), &wg)
	}

	// Esperar a que todas las solicitudes se completen
	wg.Wait()
}
