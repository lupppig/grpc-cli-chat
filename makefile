.SILENT:

gen-go:
	protoc --proto_path=proto \
    --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    proto/*.proto

clean:
	rm -r pb/*.go
server:
	go run cmd/server/*.go --port 8080
client:
	go run cmd/client/*.go --address 0.0.0.0:8080


.PHONY: redis

redis:
	echo "Starting Redis via Docker..."
	./setup



redis-down:
	echo "Tearing down Redis container..."
	if docker ps -a --format '{{.Names}}' | grep -q "^redis$$"; then \
		docker stop redis >/dev/null 2>&1 || true; \
		docker rm redis >/dev/null 2>&1 || true; \
		echo "🧹 Redis container removed"; \
	else \
		echo "Redis container does not exist"; \
	fi