package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"kulturago/auth-service/internal/handler/routes"
	"kulturago/auth-service/internal/kafka"
	"kulturago/auth-service/internal/repository"
	"kulturago/auth-service/internal/service"
	"kulturago/auth-service/internal/tokens"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	brokersCSV := os.Getenv("KAFKA_BROKERS")
	secret := []byte(os.Getenv("JWT_SECRET"))
	if dsn == "" || brokersCSV == "" || len(secret) < 16 {
		log.Fatal("env not set: DATABASE_URL / KAFKA_BROKERS / JWT_SECRET")
	}

	pg, err := repository.New(dsn)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	brokers := strings.Split(brokersCSV, ",")
	kprod := kafka.New(brokers)

	tokenMgr := tokens.NewManager(secret, 15*60, 30*24*60*60)
	authSvc := service.New(pg, kprod, tokenMgr)

	base := chi.NewRouter()
	base.Mount("/", routes.NewRouter(authSvc, tokenMgr))
	base.Get("/swagger/*", httpSwagger.WrapHandler)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           base,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("auth-service listening on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
