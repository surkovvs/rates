FROM golang:1.22.6-alpine AS builder
COPY ./rates_service/ /build/
WORKDIR /build
# RUN go mod tidy
# RUN go build -o /builded/binary ./cmd/main.go
RUN go build -mod=vendor -o /builded/binary ./cmd/main.go
FROM alpine:3.16
COPY ./rates_service/infrastructure/db/migrations/* /rates_service/migrations/
COPY --from=builder /builded/ /rates_service/
WORKDIR /rates_service
CMD [ "./binary" ]