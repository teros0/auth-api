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

	token, value, err := getToken(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if token == "" {
		http.Error(w, "Incorrect access key", http.StatusUnauthorized)
		return
	}

	client, err := getRedis()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = client.Set(token, value, 3*time.Hour).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	responser.Encode(map[string]string{
		"Access Key": key,
		"Token":      token,
	})
}

func VerToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Token")

	client, err := getRedis()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	exists := client.Exists(token).Val()
	if exists == 0 {
		http.Error(w, "Incorrect token", http.StatusBadRequest)
		return
	}
}

func DelToken(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Token")

	client, err := getRedis()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	exists := client.Exists(token).Val()
	if exists == 0 {
		http.Error(w, "Incorrect token", http.StatusBadRequest)
		return
	}
	client.Del(token)
}

func getRedis() (*redis.Client, error) {
	cl := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := cl.Ping().Result()
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	return cl, nil
}

func getToken(key string) (token, value string, err error) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer db.Close()

	query := `select token, value from auth.gettokentrue($1)`
	err = db.QueryRow(query, key).Scan(&token, &value)
	if err != nil {
		log.Print(err.Error())
		return
	}
	return
}
