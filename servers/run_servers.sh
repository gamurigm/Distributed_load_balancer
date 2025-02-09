#!/bin/bash

echo "Ingrese el número de servidores:"
read num_servers

start_port=50051
servers_file="servers.txt"  # Ruta al archivo en el directorio raíz

> "$servers_file"  # Limpiar el archivo si ya existe

for ((i=0; i<num_servers; i++))
do
    port=$((start_port + i))
    go run ./servers/server.go ":$port" &
    echo "localhost:$port" >> "$servers_file"  # Guardar la dirección del servidor
    echo "Servidor iniciado en el puerto $port"
done

echo "Todos los servidores han sido iniciados."
echo "Las direcciones de los servidores se han guardado en $servers_file."