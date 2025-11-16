# Go Chat App

A simple chat backend in Go using **Gin**, **GORM**, **PostgreSQL**, **Redis**, and **WebSockets**.
Supports:

- User signup/login (REST)
- Private and group chat (WebSocket)
- Chat history (REST)
- Group creation/join (REST)

---

## 🛠 Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Postman (for testing APIs)
- Make

---

## 🚀 Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/NatiG100/go-chat
```

```bash
cd go-chat
```

### 2. Environment variables

Create a `.env` file in the root directory:

Copy the content of `.env.example` to `.env`

Modify the variables as desired

### 3. Start Docker (Postgres + Redis)

```bash
docker-compose up -d
```

Check containers:

```bash
docker ps
```

### 4. Create the database

1. Connect to Postgres running on Docker:

```bash
docker exec -it < container-name > psql -U postgres -d chatdb
```

Password: `password` (from `docker-compose.yml`)

2. Create the `chatdb` database:

```sql
CREATE DATABASE chatdb;
```

### 5. Install dependencies

```bash
make deps
```

### 6. Run the server

```bash
make run
```

The server should start on `http://localhost:8080`.

---

## 🧪 Postman Collection

You can import the Postman collection to test the API endpoints:

**Features included:**

- User Signup/Login → returns Bearer token
- Groups → Create/Join/List
- Chat History → List messages
- WebSocket test for private and group messages

## 🔧 WebSocket Testing

- Connect using a WebSocket client (Postman, wscat, or frontend)
- URL: `ws://localhost:8080/ws`
- Authentication: make sure to include the token as Bearer authentication
- Send JSON messages:

**Private message:**

```json
{
  "type": "private_message",
  "to": "<recipient_user_id>",
  "content": "Hello!"
}
```

**Group message:**

```json
{
  "type": "group_message",
  "group_id": "<group_id>",
  "content": "Hello group!"
}
```
