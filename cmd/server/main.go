package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/noellimx/hepmilserver/src/config"
	"github.com/noellimx/hepmilserver/src/httplog"

	"github.com/noellimx/hepmilserver/src/controller/middlewares"

	"github.com/noellimx/hepmilserver/src/controller/mux/ping"

	taskmux "github.com/noellimx/hepmilserver/src/controller/mux/task"
	taskrepo "github.com/noellimx/hepmilserver/src/infrastructure/repositories/task"
	taskservice "github.com/noellimx/hepmilserver/src/service/task"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"

	_ "github.com/noellimx/hepmilserver/docs"
	"github.com/swaggo/http-swagger"
)

var Config config.Config
var DbConnPool *pgxpool.Pool

func main() {
	err := Init()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Starting hepmilserver::main().")

	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, syscall.SIGINT /*keyboard input*/, syscall.SIGTERM /*process kill*/)
	mux := http.NewServeMux()
	defaultMiddlewares := middlewares.MiddewareStack{}

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	mux.Handle("/ping", defaultMiddlewares.Finalize(ping.PingHandler{}.ServeHTTP))

	taskRepo := taskrepo.New(DbConnPool)
	taskService := taskservice.New(taskRepo)

	taskHandlers := taskmux.NewHandlers(taskService)
	mux.Handle("POST /task", defaultMiddlewares.Finalize(taskHandlers.Create))

	c := cors.New(cors.Options{
		AllowedOrigins:   append(Config.ServerConfig.Cors.AllowedOrigins, "http://localhost:5173", "http://localhost:4173"),
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	}).Handler(mux)

	go func() {
		log.Println("Listening on " + Config.ServerConfig.Port)
		http.ListenAndServe(Config.ServerConfig.Port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = httplog.ContextualizeHttpRequest(r)
			log.Printf("%s [middleware 0]\n", httplog.SPrintHttpRequestPrefix(r))
			c.ServeHTTP(w, r)
		}))
	}()

	recvSig := <-interruptSignal
	log.Println("Received signal: " + recvSig.String() + " ; exiting...")
}

func Init() (err error) {
	Config, err = config.InitConfig()
	if err != nil {
		return
	}

	// Db Connection Pool
	if Config.ConnString == "" {
		return errors.New("no db connection string provided")
	}
	config, err := pgxpool.ParseConfig(Config.ConnString)
	if err != nil {
		return err
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = 5 * time.Minute
	ctx := context.Background()

	// Create the pool
	DbConnPool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}

	err = DbConnPool.Ping(ctx)
	if err != nil {
		panic(err)
	}
	return nil
}
