package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"

	"example.com/go-chat/internal/drivers"
	"example.com/go-chat/internal/server"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()

	pg, err := drivers.NewPostgres(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("pg connect: %v", err)
	}
	defer pg.Close(ctx)

	rds := drivers.NewRedis(os.Getenv("REDIS_ADDR"))
	defer rds.Close()

	jwtMgr := drivers.NewJWTManager(os.Getenv("JWT_SECRET"))

	repos := drivers.NewRepositories(pg)

	h := server.NewHandler(repos, rds, jwtMgr)

	r := gin.Default()

	r.POST("/api/auth/signup", h.SignUp)
	r.POST("/api/auth/login", h.Login)

	// protected
	auth := r.Group("/api")
	auth.Use(server.AuthMiddleware(jwtMgr))
	{
		auth.GET("/me", h.Me)
		auth.POST("/groups", h.CreateGroup)
		auth.POST("/groups/:id/join", h.JoinGroup)
		auth.GET("/groups/:id/members", h.ListGroupMembers)
		auth.GET("/messages", h.GetPrivateHistory)
		auth.GET("/groups/:id/messages", h.GetGroupHistory)
	}

	r.GET("/ws", server.WSHandler(rds, jwtMgr, repos))

	log.Printf("listening on :%s", port)
	r.Run(":" + port)
}
