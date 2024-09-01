
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

4. Test the load balancer by sending HTTP requests to \`http://localhost:9090\`:

   ```bash
   curl http://localhost:9090
   ```

## Simulating a Backend Node Failure

In a real-world scenario, backend servers may go down due to various reasons such as network issues, hardware failures, or software crashes. To simulate a backend node going down during testing, you should specifically kill the process running on a particular port.

### Steps to Simulate Node Down

1. Identify the process running on the port of the backend server you want to simulate as down. You can use the following command to find the Process ID (PID) of the server:

   ```bash
   ps aux | grep <port_number>
   ```

   Replace \`<port_number>\` with the port number of the backend server you want to kill. For example, if you want to kill the server running on port 8083:

   ```bash
   ps aux | grep 8083
   ```

   This command will output a list of processes. Look for the line that includes your \`go run servers/server.go -port=8083\` command. The second column in the output is the PID of the process.

2. Once you've identified the PID, you can kill that specific process using the \`kill\` command:

   ```bash
   kill <PID>
   ```

   For example, if the PID is \`59396\`, you would run:

   ```bash
   kill 59396
   ```

   This will stop the server on port 8083 without affecting other processes, including the load balancer.

3. Observe how the load balancer handles the failure by checking the logs or sending new requests to the load balancer:

   ```bash
   curl http://localhost:9090
   ```

   The load balancer should automatically reroute traffic to the remaining healthy backend servers.
