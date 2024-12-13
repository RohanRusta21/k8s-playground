package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/google/uuid"
    "github.com/gorilla/mux"
    "github.com/rs/cors"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type Todo struct {
    gorm.Model
    UUID        string `json:"uuid" gorm:"unique"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Completed   bool   `json:"completed"`
    FilePath    string `json:"file_path,omitempty"`
}

var db *gorm.DB

func connectToDatabase() *gorm.DB {
    maxRetries := 5
    for attempt := 1; attempt <= maxRetries; attempt++ {
        dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
            os.Getenv("DB_HOST"),
            os.Getenv("DB_USER"),
            os.Getenv("DB_PASSWORD"),
            os.Getenv("DB_NAME"),
            os.Getenv("DB_PORT"),
        )

        database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
        if err == nil {
            log.Println("Successfully connected to database")
            return database
        }

        log.Printf("Database connection attempt %d/%d failed: %v", attempt, maxRetries, err)
        time.Sleep(time.Duration(attempt*2) * time.Second)
    }

    log.Fatal("Failed to connect to database after multiple attempts")
    return nil
}

func main() {
    // Retry database connection
    db = connectToDatabase()

    // Auto migrate the schema
    err := db.AutoMigrate(&Todo{})
    if err != nil {
        log.Fatalf("Failed to migrate database: %v", err)
    }

    // Ensure uploads directory exists
    uploadDir := "/app/uploads"
    if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
        log.Fatalf("Failed to create uploads directory: %v", err)
    }

    // Create router
    r := mux.NewRouter()

    // CRUD Routes for Todos
    r.HandleFunc("/todos", createTodo).Methods("POST")
    r.HandleFunc("/todos", getAllTodos).Methods("GET")
    r.HandleFunc("/todos/{uuid}", getTodo).Methods("GET")
    r.HandleFunc("/todos/{uuid}", updateTodo).Methods("PUT")
    r.HandleFunc("/todos/{uuid}", deleteTodo).Methods("DELETE")

    // File system routes
    r.HandleFunc("/files/upload", uploadFile).Methods("POST")
    r.HandleFunc("/files/list", listFiles).Methods("GET")
    r.HandleFunc("/files/download/{filename}", downloadFile).Methods("GET")
    r.HandleFunc("/files/{filename}", deleteFile).Methods("DELETE")

    // CORS and server setup
    // handler := cors.Default().Handler(r)
	// allow all origins and headers
	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type"},
	}).Handler(r)
    log.Println("Server starting on :8080")
    if err := http.ListenAndServe(":8080", handler); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func createTodo(w http.ResponseWriter, r *http.Request) {
    var todo Todo
    err := json.NewDecoder(r.Body).Decode(&todo)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Generate a unique UUID for the todo
    todo.UUID = uuid.New().String()

    result := db.Create(&todo)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(todo)
}

func getAllTodos(w http.ResponseWriter, r *http.Request) {
    var todos []Todo
    result := db.Find(&todos)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(todos)
}

func getTodo(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    uuid := vars["uuid"]

    var todo Todo
    result := db.Where("uuid = ?", uuid).First(&todo)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(todo)
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    uuid := vars["uuid"]

    var updatedTodo Todo
    err := json.NewDecoder(r.Body).Decode(&updatedTodo)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    result := db.Model(&Todo{}).Where("uuid = ?", uuid).Updates(map[string]interface{}{
        "completed": updatedTodo.Completed,
    })    
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    var todo Todo
    db.Where("uuid = ?", uuid).First(&todo)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(todo)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    uuid := vars["uuid"]

    result := db.Where("uuid = ?", uuid).Delete(&Todo{})
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()

    uploadDir := "/app/uploads"
    filePath := filepath.Join(uploadDir, fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(header.Filename)))
    outFile, err := os.Create(filePath)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer outFile.Close()

    _, err = io.Copy(outFile, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"file_path": filePath})
}

func listFiles(w http.ResponseWriter, r *http.Request) {
    uploadDir := "/app/uploads"
    files, err := os.ReadDir(uploadDir)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var fileNames []string
    for _, file := range files {
        if !file.IsDir() {
            fileNames = append(fileNames, file.Name())
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(fileNames)
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    fileName := vars["filename"]
    filePath := filepath.Join("/app/uploads", fileName)

    file, err := os.Open(filePath)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }
    defer file.Close()

    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
    w.Header().Set("Content-Type", "application/octet-stream")
    io.Copy(w, file)
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    fileName := vars["filename"]
    filePath := filepath.Join("/app/uploads", fileName)

    err := os.Remove(filePath)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}