package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
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

var redisOptions = &redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
}

var redisPool = sync.Pool{
	New: func() interface{} { return getRedis() },
}

type redisConn struct {
	client *redis.Client
	err    error
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/get-token/", makeHandler(GetToken)).Methods("GET")
	router.HandleFunc("/ver-token/", makeHandler(VerToken)).Methods("GET")
	router.HandleFunc("/del-token/", makeHandler(DelToken)).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, redisConn)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn := redisPool.Get().(redisConn)
		if conn.err != nil {
			http.Error(w, conn.err.Error(), http.StatusInternalServerError)
			return
		}
		defer redisPool.Put(conn)
		fn(w, r, conn)
	}
}

// GetToken checks if an Access Key is in db and, is so, writes it to Redis
func GetToken(w http.ResponseWriter, r *http.Request, conn redisConn) {
	key := r.Header.Get("Key")

	token, value, err := getToken(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if token == "" {
		http.Error(w, "Incorrect access key", http.StatusUnauthorized)
		return
	}

	err = conn.client.Set(token, value, 3*time.Hour).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(map[string]string{
		"Access Key": key,
		"Token":      token,
	})
}

// VerToken verifies given token by checking it's existence in Redis
func VerToken(w http.ResponseWriter, r *http.Request, conn redisConn) {
	token := r.Header.Get("Token")
	exists := conn.client.Exists(token).Val()
	if exists == 0 {
		http.Error(w, "No such token", http.StatusBadRequest)
		return
	}
}

// DelToken deletes given token from Redis
func DelToken(w http.ResponseWriter, r *http.Request, conn redisConn) {
	token := r.Header.Get("Token")
	exists := conn.client.Exists(token).Val()
	if exists == 0 {
		http.Error(w, "No such token", http.StatusBadRequest)
		return
	}
	conn.client.Del(token)
}

func getRedis() redisConn {
	var conn redisConn
	conn.client = redis.NewClient(redisOptions)

	_, conn.err = conn.client.Ping().Result()
	if conn.err != nil {
		log.Print(conn.err.Error())
		return conn
	}
	return conn
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
