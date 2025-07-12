package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"kulturago/auth-service/internal/handler/routes"
	"kulturago/auth-service/internal/kafka"
	"kulturago/auth-service/internal/logger"
	"kulturago/auth-service/internal/redis"
	"kulturago/auth-service/internal/repository"
	"kulturago/auth-service/internal/service"
	"kulturago/auth-service/internal/storage"
	"kulturago/auth-service/internal/tokens"
	"kulturago/auth-service/internal/util"
)

func main() {
	logger.Init()
	logger.Log.Info("auth-service starting…")

	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	brokersCSV := os.Getenv("KAFKA_BROKERS")
	secret := []byte(os.Getenv("JWT_SECRET"))
	redisAddr := os.Getenv("REDIS_ADDR")

	awsRegion := os.Getenv("AWS_REGION")
	awsKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	s3Bucket := os.Getenv("S3_BUCKET")
	publicURL := os.Getenv("S3_PUBLIC_ENDPOINT")
	endpoint := os.Getenv("S3_ENDPOINT")

	if dsn == "" || brokersCSV == "" || len(secret) < 16 {
		log.Fatal("missing DATABASE_URL / KAFKA_BROKERS / JWT_SECRET")
	}
	if awsRegion == "" || awsKey == "" || awsSecret == "" || s3Bucket == "" {
		log.Fatal("missing AWS_* or S3_BUCKET")
	}

	pg, err := repository.New(dsn)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	rdb, _ := redis.New(redisAddr)
	rtStore := redis.NewRefresh(rdb.Client)

	brokers := strings.Split(brokersCSV, ",")
	kprod := kafka.New(brokers)

	store, err := storage.New(
		context.Background(),
		s3Bucket,
		awsRegion,
		endpoint,
		publicURL,
		awsKey, awsSecret,
	)
	if err != nil {
		log.Fatal(err)
	}

	accessTTL := util.EnvInt("ACCESS_TTL_SECONDS", 60*60)
	refreshTTL := util.EnvInt("REFRESH_TTL_SECONDS", 30*24*60*60)
	tokenMgr := tokens.NewManager(secret, accessTTL, refreshTTL)
	authSvc := service.New(pg, kprod, tokenMgr, rtStore, store)

	r := chi.NewRouter()
	r.Mount("/", routes.NewRouter(authSvc, tokenMgr))
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Log.Info("listening on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
