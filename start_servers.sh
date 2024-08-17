#!/bin/bash

# Start servers on a range of ports
start_port=8081
end_port=8084

for port in $(seq $start_port $end_port); do
    echo "Starting server on port $port"
    go run servers/server.go -port=$port &
done

echo "Servers started on ports from $start_port to $end_port"