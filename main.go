package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "123post"
	dbname   = "postgres"
)

var psqlInfo = fmt.Sprintf("host=%s port=%d user=%s "+
	"password=%s dbname=%s sslmode=disable",
	host, port, user, password, dbname)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/get-token/", GetToken).Methods("GET")
	router.HandleFunc("/ver-token/", VerToken).Methods("GET")
	router.HandleFunc("/del-token/", DelToken).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8000", router))
}

// GetToken
func GetToken(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Key")
	responser := json.NewEncoder(w)
	token, value := getToken(key)

	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	client.Set(token, value, 3*time.Hour)
	responser.Encode(map[string]string{
		"Access Key": key,
		"Token":      token,
	})
}

func VerToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Token")

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	exists := client.Exists(token).Val()
	if exists == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DelToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Token")

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	client.Del(token)
	w.WriteHeader(http.StatusOK)
}

func getToken(key string) (token, value string) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	query := `select token, value from auth.gettokentrue($1)`
	err = db.QueryRow(query, key).Scan(&token, &value)

	if err != nil {
		fmt.Println("Query error", err)
	}

	return
}
