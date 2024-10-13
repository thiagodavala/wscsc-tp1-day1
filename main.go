package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-ini/ini"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	AppPort     string
	RedisHost   string
	RedisPort   string
	PostgresDSN string
}

type Input struct {
	ID   uint
	Hash string
}

var (
	rdb     *redis.Client
	db      *gorm.DB
	config  Config
	logFile *os.File
)

func init() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Failed to load config file: %s\n", err)
		os.Exit(1)
	}

	config = Config{
		AppPort:     cfg.Section("app").Key("port").MustString("8080"),
		RedisHost:   cfg.Section("redis").Key("host").MustString("localhost"),
		RedisPort:   cfg.Section("redis").Key("port").MustString("6379"),
		PostgresDSN: cfg.Section("postgres").Key("dsn").MustString(""), // Certifique-se de definir a DSN completa no arquivo INI
	}

	// Conectar ao Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		Password: "", // replace with your password
		DB:       0,  // use default DB
	})

	// Conectar ao PostgreSQL
	db, err = gorm.Open(postgres.Open(config.PostgresDSN))
	if err != nil {
		panic(err)
	}

	// Criar o arquivo de log
	logFile, err = os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf(fmt.Sprintf("Running on %s!\n", config.AppPort))
}

func calcHandler(w http.ResponseWriter, r *http.Request) {
	input := r.URL.Query().Get("input")
	exists, err := rdb.Exists(r.Context(), input).Result()
	if err != nil {
		panic(err)
	}
	if exists == 1 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{`status`:`Registro em cache, operação já realizada!`}"))
		return
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	_, err = logFile.WriteString(fmt.Sprintf("%s - %s\n", time.Now().Format(time.RFC3339), encoded))
	if err != nil {
		panic(err)
	}
	db.Create(&Input{Hash: encoded})
	err = rdb.Set(r.Context(), input, "processed", 0).Err()
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{`status`:`Cadastro realizado!`}"))
}

func cancellationHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	_, err := logFile.WriteString(fmt.Sprintf("%s - 222-cancellation-id-%s\n", time.Now().Format(time.RFC3339), id))
	if err != nil {
		// ... (tratar o erro)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{`status`:`Cancelamento realizado!`}"))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/calc", calcHandler)
	http.HandleFunc("/cancellation", cancellationHandler)

	http.ListenAndServe(config.AppPort, nil)
}
