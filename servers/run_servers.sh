#!/bin/bash

# Preguntar al usuario el número de servidores a ejecutar
read -p "Ingresa el número de servidores a ejecutar: " NUM_SERVERS

# Navegar al directorio de servidores
cd servers/

# Compilar el servidor (si no está compilado)
echo "Compilando el servidor..."
go build -o server server.go
if [ $? -ne 0 ]; then
    echo "Error al compilar el servidor."
    exit 1
fi
echo "Servidor compilado correctamente."

# Ejecutar los servidores
for ((i = 0; i < NUM_SERVERS; i++)); do
    PORT=$((50051 + i))
    echo "Ejecutando servidor en el puerto :$PORT..."
    ./server :$PORT &
done

echo "Todos los servidores han sido iniciados."
