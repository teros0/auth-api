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
	token := getToken(key)

	fmt.Println("Token", token)
	if len(token) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	responser.Encode(map[string]string{
		"Access Key": key,
		"Token":      token,
	})

	client.Set(key, token, 3*time.Hour)
}

func VerToken(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Key")
	token := getToken(key)
	responser := json.NewEncoder(w)

	fmt.Println("Token", token)
	if len(token) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		responser.Encode(map[string]bool{
			"Verified": false,
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	responser.Encode(map[string]bool{
		"Verified": true,
	})
}

func DelToken(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Key")
	token := getToken(key)

	fmt.Println("Token", token)
	if len(token) == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	client.Del(key)
	w.WriteHeader(http.StatusOK)
}

func getToken(key string) (token string) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var val string
	fmt.Println("Executing true gettoken")
	query := `SELECT auth.gettokentrue($1)`
	rows, err := db.Query(query, key)
	for rows.Next() {
		rows.Scan(&token, &val)
	}
	fmt.Println("Results", token, val)

	query = `SELECT auth.gettoken($1)`
	db.QueryRow(query, key).Scan(&token)

	return
}
