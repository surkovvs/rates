package main

import (
	"context"
	"fmt"
	"log"
	"rates_service/pkg/proto/gen/ratesservicepb"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	conn, err := grpc.NewClient("localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second/2)
	hcResp, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		log.Fatal("remote service health check failed", err)
	}
	if hcResp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		log.Fatal("unexpected server status", grpc_health_v1.HealthCheckResponse_ServingStatus_name[int32(hcResp.GetStatus())])
	} else {
		fmt.Println("remote server healthy", hcResp)
	}
	client := ratesservicepb.NewRatesServiceClient(conn)

	for {
		go func() {
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			resp, err := client.GetRates(ctx, &ratesservicepb.GetRatesRequest{})
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Println(time.Now().Format("15:04:05"), resp)
		}()
		time.Sleep(time.Second * 5)
	}
}
