package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"alikhan.practice3/internal/handler"
	"alikhan.practice3/internal/middleware"
	"alikhan.practice3/internal/repository"
	_postgres "alikhan.practice3/internal/repository/_postgres"
	"alikhan.practice3/internal/usecase"
	"alikhan.practice3/pkg/modules"
)

func main() {
	dbConfig := &modules.PostgreConfig{
		Host:        "localhost",
		Port:        "5432",
		Username:    "postgres",
		Password:    "alikhan",
		DBName:      "mydb",
		SSLMode:     "disable",
		ExecTimeout: 5 * time.Second,
	}

	db := _postgres.NewPGXDialect(context.Background(), dbConfig)
	log.Println("Database connected and migrated successfully!")

	repos := repository.NewRepositories(db)
	userUC := usecase.NewUserUsecase(repos.UserRepository)
	userH := handler.NewUserHandler(userUC)

	mux := http.NewServeMux()

	mux.Handle("/users", middleware.APIKeyAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userH.GetUsers(w, r)
		case http.MethodPost:
			userH.CreateUser(w, r)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/users/", middleware.APIKeyAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userH.GetUserByID(w, r)
		case http.MethodPut, http.MethodPatch:
			userH.UpdateUser(w, r)
		case http.MethodDelete:
			userH.DeleteUser(w, r)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})))

	server := &http.Server{
		Addr:    ":8080",
		Handler: middleware.Logger(mux),
	}

	fmt.Println("Server is running on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
