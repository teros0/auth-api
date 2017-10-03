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

	client := getRedis()

	client.Set(token, value, 3*time.Hour)

	responser.Encode(map[string]string{
		"Access Key": key,
		"Token":      token,
	})
}

func VerToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Token")

	client := getRedis()

	exists := client.Exists(token).Val()
	if exists == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DelToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Token")

	client := getRedis()

	client.Del(token)
	w.WriteHeader(http.StatusOK)
}

func getRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func getToken(key string) (token, value string) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Print(err.Error())
	}
	defer db.Close()

	query := `select token, value from auth.gettokentrue($1)`
	err = db.QueryRow(query, key).Scan(&token, &value)

	if err != nil {
		log.Print(err.Error())
	}

	return
}
