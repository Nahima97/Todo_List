package main

import (
	"fmt"
	"log"
	"net/http"
	"todo/auth"
	"todo/database"
	"todo/handlers"

	"github.com/gorilla/mux"
)

/*You are to build a simple Todo List API that allows users to register, log in, and manage their own todos.
The goal of this assignment is to demonstrate your understanding of middleware in a real-world application.
Requirements:
Implement user registration and login functionality. DONE
Use JWT authentication to protect the todo endpoints. DONE
Create middleware that:
Authenticates requests using JWT. DONE
Core Features:
POST /register: Register a new user. DONE
POST /login: Log in and receive a JWT token. DONE
POST /todos: Create a todo (authenticated). DONE
GET /todos: Get all todos belonging to the logged-in user. DONE
PUT /todos/:id: Update a specific todo (only if it belongs to the user). DONE 
DELETE /todos/:id: Delete a specific todo (only if it belongs to the user). DONE */

func main() {

	database.InitDb()

	//define routes
	r := mux.NewRouter()

	//public routes
	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.HandleFunc("/register", handlers.RegisterUser).Methods("POST")

	//sub router for protected routes

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(auth.AuthMiddleware)

	//authenticated routes
	protected.HandleFunc("/todos", handlers.CreateToDo).Methods("POST")
	protected.HandleFunc("/todos", handlers.GetToDo).Methods("GET")
	protected.HandleFunc("/todos/{id}", handlers.UpdateToDo).Methods("PUT")
	protected.HandleFunc("/todos/{id}", handlers.DeleteToDo).Methods("DELETE")

	// start server
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("failed to start server", err)
	}
	fmt.Println("server started in :8080")
}
