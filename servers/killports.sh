#!/bin/bash

# Verificar que se proporcionaron los argumentos necesarios
if [ "$#" -ne 2 ]; then
  echo "Uso: $0 <puerto_inicio> <puerto_final>"
  exit 1
fi

# Asignar los valores de los puertos de inicio y fin
puerto_inicio=$1
puerto_final=$2

# Asegurarse de que los puertos ingresados son números
if ! [[ "$puerto_inicio" =~ ^[0-9]+$ ]] || ! [[ "$puerto_final" =~ ^[0-9]+$ ]]; then
  echo "Error: Los puertos deben ser números."
  exit 1
fi

# Verificar que el puerto de inicio no sea mayor que el puerto final
if [ "$puerto_inicio" -gt "$puerto_final" ]; then
  echo "Error: El puerto de inicio no puede ser mayor que el puerto final."
  exit 1
fi

# Detener servidores en el rango de puertos especificado
for ((puerto=$puerto_inicio; puerto<=$puerto_final; puerto++)); do
  pid=$(lsof -t -i :$puerto)  # Obtener el PID del proceso en el puerto
  if [ ! -z "$pid" ]; then
    kill $pid  # Matar el proceso
    echo "Servidor en el puerto $puerto detenido exitosamente."
  else
    echo "No se encontró ningún servidor corriendo en el puerto $puerto."
  fi
done
