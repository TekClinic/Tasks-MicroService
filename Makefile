PROTO_SRC = tasks_protobuf/tasks_service.proto

generate-proto:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $(PROTO_SRC)
	@echo "Done generating protobuf code!"

.PHONY: generate-proto
