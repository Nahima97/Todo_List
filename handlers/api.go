package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"todo/auth"
	"todo/database"
	"todo/models"
	"todo/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// POST /register: Register a new user.
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	// collect request details
	var signUp models.Users
	err := json.NewDecoder(r.Body).Decode(&signUp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//check if they exist in the db
	var user models.Users
	err = database.DB.Where("username = ?", signUp.Username).First(&user).Error
	if err == nil {
		http.Error(w, "user already exists", http.StatusBadRequest)
		return
	}
	//hash password
	hashedPass, err := utils.HashPassword(signUp.Password)
	if err != nil {
		http.Error(w, "unable to hash password", http.StatusInternalServerError)
		return
	}
	signUp.Password = hashedPass

	myuuid := uuid.NewString()
	signUp.ID = myuuid

	//put into db
	err = database.DB.Create(&signUp).Error
	if err != nil {
		http.Error(w, "unable to create user", http.StatusInternalServerError)
		return
	}
	//response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(signUp)
}

// POST /login: Log in and receive a JWT token.
func Login(w http.ResponseWriter, r *http.Request) {
	// get login data from request body
	var login models.Users
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// check that the user does exist in the database
	var user models.Users
	err = database.DB.Where("username = ?", login.Username).First(&user).Error
	if err != nil {
		http.Error(w, "user not found", http.StatusBadRequest)
		return
	}
	//check password matches hashed password in database
	err = utils.ComparePassword(user.Password, login.Password)
	if err != nil {
		http.Error(w, "passwords do not match", http.StatusBadRequest)
		return
	}
	//generate JWT token
	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "unable to generate token", http.StatusBadRequest)
		return
	}
	//response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
}

// POST /todos: Create a todo (authenticated).
func CreateToDo(w http.ResponseWriter, r *http.Request) {
	//get todo from request body
	var todo models.ToDoList
	err := json.NewDecoder(r.Body).Decode(&todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//check it doesn't exist in the db already
	var existingTodo models.ToDoList
	err = database.DB.Where("name = ?", todo.Name).First(&existingTodo).Error
	if err == nil {
		http.Error(w, "todo already exists", http.StatusBadRequest)
		return
	}

	myuuid := uuid.NewString()
	todo.ID = myuuid

	// add to db using encoder
	err = database.DB.Create(&todo).Error
	if err != nil {
		http.Error(w, "unable to create todo", http.StatusInternalServerError)
		return
	}
	//response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

// GET /todos: Get all todos belonging to the logged-in user.
func GetToDo(w http.ResponseWriter, r *http.Request) {
	//retrieve todolist of that user
	var todos []models.ToDoList

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	// tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := auth.VerifyJWT(tokenString)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["userID"].(string)

	// err = database.DB.Preload("todolists").First("user_id =?", user.ID).Error
	err = database.DB.Where("user_id = ?", userID).Find(&todos).Error
	if err != nil {
		http.Error(w, "failed to retrieve todos", http.StatusInternalServerError)
		return
	}
	//send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(todos)
}

// PUT /todos/:id: Update a specific todo (only if it belongs to the user).
func UpdateToDo(w http.ResponseWriter, r *http.Request) {
	var toDo models.ToDoList
	err := json.NewDecoder(r.Body).Decode(&toDo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//getting id of todo
	id := strings.TrimPrefix(r.URL.Path, "/todos/")
	toDo.ID = id
	err = database.DB.Model(&toDo).Where("id = ?", id).Update("name", toDo.Name).Error
	if err != nil {
		http.Error(w, "unable to update todo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /todos/:id: Delete a specific todo (only if it belongs to the user).
func DeleteToDo(w http.ResponseWriter, r *http.Request) {
	var toDo models.ToDoList
	//getting id of todo
	id := strings.TrimPrefix(r.URL.Path, "/todos/")

	err := database.DB.Where("id =?", id).Delete(&toDo).Error
	if err != nil {
		http.Error(w, "failed to delete todo", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
