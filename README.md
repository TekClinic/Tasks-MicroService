# Tasks-Microservice

This repository contains a gRPC service for managing task information within TekClinic.
The service is implemented in Go and uses Protobuf for defining message types and service methods.

Please note that the provided code assumes the existence of a `TekClinic/MicroService-Lib `
library for authentication and environment variable handling,
and setting up the environment variables found in `TekClinic/MicroService-Lib` is a prerequisite.

## Table of Contents

- [Installation](#installation)
- [gRPC Functions](docs/grpc.md#grpc-functions)
    - [GetTask](docs/grpc.md#gettask)
    - [GetTasksIDs](docs/grpc.md#gettasksids)
    - [CreateTask](docs/grpc.md#createtask)
    - [DeleteTask](docs/grpc.md#deletetask)
    - [UpdateTask](docs/grpc.md#updatetask)

## Installation

1. Clone the repository:

```bash
git clone https://github.com/TekClinic/Tasks-MicroService.git
```

2. Set up the required environment variables for database connection:

```
DB_ADDR=<database_address>
DB_USER=<database_user>
DB_PASSWORD=<database_password>
DB_DATABASE=<database_name>
```

3. This microservice uses the `TekClinic/MicroService-Lib` library for base configuration,
   therefore, you have to set up environment variables for the library.
   For further information, please refer to
   the [MicroService-Lib repository](https://github.com/TekClinic/MicroService-Lib)

4. Run the server:

```bash
go run server.go
```

## Protobuf

Protobuf generates Go code. You must setup the protobuf compiler with the Go and the gRPC plugins: https://grpc.io/docs/languages/go/quickstart/.

To generate the Go code, we have a Makefile command. After doing the above
setup, you can just run:

```bash
make generate-proto
```
