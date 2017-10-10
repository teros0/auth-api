package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Config struct {
	Host          string        `json:"host"`
	Port          int           `json:"port"`
	User          string        `json:"user"`
	Password      string        `json:"password"`
	DBname        string        `json:"dbname"`
	LiveTimeHours time.Duration `json:"liveTimeHours"`
}

type JResp struct {
	Token string          `json:"token"`
	Value json.RawMessage `json:"value"`
}

var redisPool = sync.Pool{
	New: func() interface{} { return getRedis() },
}
var conf = initializeConfig()

func initializeConfig() *Config {
	raw, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println("Cannot read file", err)
	}

	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		fmt.Println(err)
	}
	return &c
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/get-token/", makeHandler(GetToken)).Methods("GET")
	router.HandleFunc("/ver-token/", makeHandler(VerToken)).Methods("GET")
	router.HandleFunc("/del-token/", makeHandler(DelToken)).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, *redis.Client)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn := redisPool.Get().(*redis.Client)
		if conn == nil {
			http.Error(w, "Redis problems", http.StatusInternalServerError)
			return
		}
		defer redisPool.Put(conn)
		fn(w, r, conn)
	}
}

// GetToken retrieves a token corresponding to an Access Key
// and writes this token to Redis
func GetToken(w http.ResponseWriter, r *http.Request, client *redis.Client) {
	key := r.Header.Get("Key")

	resp, err := getToken(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if resp.Token == "" {
		http.Error(w, "Incorrect access key", http.StatusUnauthorized)
		return
	}

	if err = client.Set(resp.Token, string(resp.Value), conf.LiveTimeHours*time.Hour).Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// VerToken verifies given token by checking it's existence in Redis
func VerToken(w http.ResponseWriter, r *http.Request, client *redis.Client) {
	token := r.Header.Get("Token")
	val := client.Get(token).Val()
	if val == "" {
		http.Error(w, "No such token", http.StatusBadRequest)
		return
	}
	w.Write([]byte(val))
	w.Header().Set("Content-Type", "application/json")
}

// DelToken deletes given token from Redis
func DelToken(w http.ResponseWriter, r *http.Request, client *redis.Client) {
	token := r.Header.Get("Token")
	if exists := client.Exists(token).Val(); exists == 0 {
		http.Error(w, "No such token", http.StatusBadRequest)
		return
	}
	client.Del(token)
}

func getRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Print(err.Error())
		return nil
	}
	return client
}

func getToken(key string) (resp JResp, err error) {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		conf.Host, conf.Port, conf.User, conf.Password, conf.DBname))
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer db.Close()

	query := `select token, value from auth.gettokentrue($1)`
	if err = db.QueryRow(query, key).Scan(&resp.Token, &resp.Value); err != nil {
		log.Print(err.Error())
		return
	}
	return
}
