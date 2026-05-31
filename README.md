# Go Microservices: Auth, Blog, Notification

Three Go microservices communicating via gRPC and Kafka.

## Services

- **Auth** – registration/login, JWT issuing, token validation via gRPC
- **Blog** – posts CRUD, Redis caching, authorization via Auth gRPC
- **Notification** – listens to Kafka `user_created` topic, logs new users

## Stack

- Go 1.22+
- PostgreSQL (separate DB per service)
- Redis (cache for Blog)
- Kafka (user created events)
- gRPC (Auth ↔ Blog token validation)
- Docker / Docker Compose

## Quick Start

### 1. Clone
git clone github.com:rubbenc1/integrated_microservices.git
cd integrated_microservices

### 2. Configure environment
cp internal/auth/.env.example internal/auth/.env
cp internal/blog/.env.example internal/blog/.env
# Fill in passwords and secrets in the .env files

### 3. Run migrations
make migrate

### 4. Start service
make run-auth
make run-blog
make run-notification

## Ports

| Service      | Port  |
|--------------|-------|
| Auth HTTP    | 8080  |
| Auth gRPC    | 50051 |
| Blog HTTP    | 8081  |
| Kafka UI     | 8082  |
| Auth DB      | 5433  |
| Blog DB      | 5434  |
| Redis        | 6379  |

## API

### Auth
- `POST /register` – create account
- `POST /login` – get JWT token

### Blog
- `GET    /posts` – list posts (cached)
- `GET    /posts/{id}` – get post by id
- `POST   /posts` – create post (requires JWT)
- `PUT    /posts/{id}` – update post (requires JWT)
- `DELETE /posts/{id}` – delete post (requires JWT)

