package main

import (
	"fmt"
	"log"
	"net"
	"order-billing/config"
	"order-billing/internal/billing/database"
	"order-billing/internal/billing/repository"
	"order-billing/internal/billing/server"
	"order-billing/internal/billing/service"
	"order-billing/proto/billing"

	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		//пока ошибки не может быть
		log.Printf("config load err: %v", err)
	}
	db, err := database.NewPostgresDB(cfg.BillingDSN)
	if err != nil {
		log.Fatal("cannot connect to db")
	}
	defer db.Close()

	repo := repository.NewPostgresRepo(db)
	svc := service.NewBillingService(repo)
	adapter := server.NewServer(svc)

	grpcEngine := grpc.NewServer()
	billing.RegisterBillingServiceServer(grpcEngine, adapter)

	port := fmt.Sprintf(":%s", cfg.BillingPort)
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("Error starting server, err: %v", err)
		return
	}
	log.Printf("Server running at port: %s", listener.Addr())

	if err = grpcEngine.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
