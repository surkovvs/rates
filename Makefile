shell=/bin/sh
SERVICE_PATH=./rates_service

M_PATH=/cmd/main.go
BIN_PATH=./rates_service/bin/usdt
build: tidy		# сборка приложения
	go build -o $(BIN_PATH) "$(SERVICE_PATH)$(M_PATH)"
test: tidy		# прогон unit тестов
	go test ./rates_service/...
docker-build:	# сборка образа с приложением
run: tidy		# запуск приложения
	go run ./rates_service/cmd/main.go -c=./testdata/.env
lint:			# запуск линтера
	golangci-lint run ./rates_service/...

cd_to_service:
	cd  
tidy:
	cd $(SERVICE_PATH) && go mod tidy


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

POSTGRESQL_URL := 'postgres://vladimir:sandbox@localhost:5433/usdt?sslmode=disable'
PATH_TO_MIGR:='./rates_service/infrastructure/db/migrations'
migrate_up1:
	migrate -database ${POSTGRESQL_URL} -path ${PATH_TO_MIGR} up 1

migrate_down1:
	migrate -database ${POSTGRESQL_URL} -path ${PATH_TO_MIGR} down 1

migrate_drop:
	migrate -database ${POSTGRESQL_URL} -path ${PATH_TO_MIGR} drop -f