module kulturago/auth-service

go 1.23.3

require (
    github.com/admpub/goth v0.0.0-20250212-123abc     // yandex provider
    github.com/confluentinc/confluent-kafka-go/v2 v2.4.0
    github.com/go-chi/chi/v5 v5.0.10
    github.com/go-chi/cors v1.2.1
    github.com/golang-jwt/jwt/v5 v5.2.1
    github.com/google/uuid v1.6.0
    github.com/joho/godotenv v1.5.1
    github.com/markbates/goth v1.81.0
    github.com/pressly/goose/v3 v3.15.0
    github.com/swaggo/swag v1.16.2
    golang.org/x/crypto v0.23.0
    github.com/jackc/pgx/v5 v5.5.4
)

replace github.com/markbates/goth => github.com/markbates/goth v1.81.0