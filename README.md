
# Dynamic Load Balancer

This repository contains a Go implementation of a dynamic load balancer that distributes HTTP traffic across multiple backend servers. The load balancer can be used with a configurable load balancing strategy (e.g., Round Robin, Hashed).

## Project Structure

```
.
├── README.md
├── common
│   └── common.go
├── go.mod
├── go.sum
├── main.go
├── servers
│   └── server.go
├── strategy
│   └── strategy.go
└── vendor
    ├── github.com
    │   └── google
    └── modules.txt
```

## Features

- **Round Robin Balancing**: Distributes traffic evenly across all backend servers.
- **Hashed Balancing**: Consistent hashing to balance requests based on request content.
- **Multiple Backends**: You can configure multiple backend servers and the load balancer will route traffic to them.

## How to Run

1. Clone the repository:

   ```bash
   git clone <repo-url>
   cd dynamic-load-balancer
   ```

2. Run the backend servers:

   ```bash
   go run servers/server.go -port=8081
   go run servers/server.go -port=8082
   go run servers/server.go -port=8083
   go run servers/server.go -port=8084
   ```

3. Start the load balancer:

   ```bash
   go run main.go
   ```

4. Test the load balancer by sending HTTP requests to `http://localhost:9090`:

   ```bash
   curl http://localhost:9090
   ```

## Automation Script

To spawn multiple backend servers automatically, you can use the provided `start_servers.sh` script:

```bash
#!/bin/bash

# Start servers on a range of ports
start_port=8081
end_port=8084

for port in $(seq $start_port $end_port); do
    echo "Starting server on port $port"
    go run servers/server.go -port=$port &
done

echo "Servers started on ports from $start_port to $end_port"
```

Run the script with:

```bash
./start_servers.sh
```

## License

This project is licensed under the MIT License.
