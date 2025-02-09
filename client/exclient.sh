#!/bin/bash

# Verifica que se pase un argumento
if [ -z "$1" ]; then
    echo "Uso: ./exclient.sh <número de veces>"
    exit 1
fi

# Número de veces a ejecutar el cliente
N=$1

# Ejecutar el cliente N veces
for ((i=1; i<=N; i++)); do
    go run client/client.go $1 &
done

echo "Se han lanzado $N instancias de client.go"
