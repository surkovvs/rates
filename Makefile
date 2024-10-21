shell=/bin/sh
M_PATH=./rates_service/cmd/main.go
BIN_PATH=./rates_service/bin/usdt
build: tidy		# сборка приложения
	go build -o $(BIN_PATH) $(M_PATH)
test: tidy		# прогон unit тестов
	go test ./rates_service/...
docker-build:	# сборка образа с приложением
run: 			# запуск приложения
	cd ./rates_service && go run ./cmd/main.go
lint:			# запуск линтера
	golangci-lint run ./rates_service/...
tidy:
	cd ./rates_service && go mod tidy

all: 
PROTO_DIR=./rates_service/pkg/proto
gen_proto:
	protoc -I=$(PROTO_DIR) \
  	--go_out=. \
  	--go-grpc_out=. \
  	$(PROTO_DIR)/*.proto
# gen_gateway:
# 	protoc -I=$(PROTO_DIR) \
#   	--grpc-gateway_out=. \
#   	--grpc-gateway_opt generate_unbound_methods=true \
#   	--openapiv2_out ./rates_service/docs/ \
#   	$(PROTO_DIR)/service.proto
create_migr_stage:
	migrate create -ext sql -dir ./rates_service/infrastructure/db/migration -seq -digits 2 init
clean:
	rm -r ./rates_service/bin &
	rm -r $(PROTO_DIR)/gen/* &
	rm -r ./rates_service/docs/