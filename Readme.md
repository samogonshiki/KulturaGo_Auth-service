# Auth-service

![intro](src/auth-klg-intro.png)

### Архитектура

```mermaid
graph TD
  subgraph Edge_infra
    Nginx -.-> API_GW[API-Gateway]
  end
  API_GW -->|REST| Auth[Auth-service]
  Auth -->|Kafka events| Kafka[(Kafka)]
  Auth -->|PostgreSQL| PG[(DB auth)]
  Auth -->|Redis black-list| Redis
```

> [!NOTE]
>### REST‑API
>
> | HTTP  | Путь                           | Описание                                        | Токен      |
> |-------|--------------------------------|-------------------------------------------------|------------|
> | POST  | /api/v1/auth/signup            | Регистрация нового пользователя                 | —          |
> | POST  | /api/v1/auth/signin            | Логин, выдача access + refresh                  | —          |
> | POST  | /api/v1/auth/refresh           | Обновление access-токена по refresh             | refresh    |
> | POST  | /api/v1/auth/logout            | Инвалидация пары токенов                        | access     |
> | GET   | /api/v1/me                     | Короткая карточка «Я»                           | access     |
> | GET   | /api/v1/profile                | Полный профиль                                  | access     |
> | PUT   | /api/v1/profile                | Сохранение профиля                              | access     |
> | GET   | /api/v1/avatar/presign         | Presigned-URL для загрузки аватара в S3         | access     |



### Вход через Yandex ID, VK ID, APPLE ID

```mermaid
sequenceDiagram
  participant SPA
  participant Edge as Nginx
  participant GW  as API‑Gateway
  participant AS  as Auth‑service
  participant VK

  SPA->>Edge: GET /api/v1/auth/oauth/vk/login
  Edge->>AS: 302 redirect
  AS-->>VK:  authorize
  VK-->>SPA: redirect code
  SPA->>Edge: /callback?code
  Edge->>AS: exchange code
  AS->>AS: find/create user
  AS-->>SPA: JSON {access, refresh}
  AS-->>Kafka: user.signed_in
```


### Подключение к `KulturaGo_infostructure`

```bash
docker network inspect backend >/dev/null 2>&1 || \
  docker compose -f ../KulturaGo_infostructure/docker-compose.yml up -d

make dev
```

```
git clone https://…/KulturaGo_Auth-service.git
cd KulturaGo_Auth-service
cp .env.example .env                    
make dev                                 
open http://localhost:8080/swagger/index.html
```

### migrations

```shell
docker exec -i kulturago_auth-service-postgres-1 psql \
      -U root \
      -d postgres \
      < ./db/migrations/0001_init.up.sql
```

## Redactor:
- **Finnik**