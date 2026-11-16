package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"order-billing/config"
	"order-billing/internal/order/billingclient"
	"order-billing/internal/order/database"
	"order-billing/internal/order/handler"
	"order-billing/internal/order/repository"
	"order-billing/internal/order/service"
	"order-billing/proto/billing"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg, err := config.Load()

	//logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	//grpc client
	target := fmt.Sprintf("%s:%s", cfg.BillingHost, cfg.BillingPort)
	credentials := insecure.NewCredentials()

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(credentials))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	grpcClient := billing.NewBillingServiceClient(conn)

	//db conn
	db, err := database.NewPostgresDB(cfg.OrderDSN)
	if err != nil {
		log.Fatal("cannot connect to db")
	}
	defer db.Close()

	billingClient := billingclient.New(grpcClient)
	repo := repository.NewPostgresRepo(db)
	svc := service.NewOrderService(billingClient, repo)
	handler := handler.NewOrderHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", handler.CreateOrder)
	mux.HandleFunc("GET /orders/{id}", handler.GetByID)

	server := &http.Server{
		Addr:    ":" + cfg.OrderPort,
		Handler: mux,
	}

	go func() {
		log.Printf("Сервер успешно запущен на порте: %s", server.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server crash: %v", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Print("Не получилось выключить нормально")
		if err := server.Close(); err != nil {
			log.Print("Не получилось выключить никак")
		}
		return
	}

	log.Print("Сервер успешно остановлен")
}
