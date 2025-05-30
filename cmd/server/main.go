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

	"github.com/noellimx/hepmilserver/src/infrastructure/reddit_miner"
	statisticsrepo "github.com/noellimx/hepmilserver/src/infrastructure/repositories/statistics"
	statisticsservice "github.com/noellimx/hepmilserver/src/service/statistics"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
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
	mux.Handle("DELETE /task", defaultMiddlewares.Finalize(taskHandlers.Delete))

	c := cors.New(cors.Options{
		AllowedOrigins:   append(Config.ServerConfig.Cors.AllowedOrigins, "http://localhost:5173", "http://localhost:4173"),
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	}).Handler(mux)

	statisticsRepo := statisticsrepo.NewAAA(DbConnPool)
	statisticService := statisticsservice.NewWWW(statisticsRepo)

	go func() {
		log.Println("Listening on " + Config.ServerConfig.Port)
		http.ListenAndServe(Config.ServerConfig.Port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = httplog.ContextualizeHttpRequest(r)
			log.Printf("%s [middleware 0]\n", httplog.SPrintHttpRequestPrefix(r))
			c.ServeHTTP(w, r)
		}))
	}()

	cron := NewWorker(taskService, statisticService)
	cron.Start()

	recvSig := <-interruptSignal
	log.Println("Received signal: " + recvSig.String() + " ; tearing down...")
	<-cron.Stop().Done()

	log.Println("Terminating hepmilserver::main()...")
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

func NewWorker(taskService *taskservice.Service, statisticsService *statisticsservice.Service) *cron.Cron {
	c := cron.New(cron.WithChain(
		cron.Recover(cron.DefaultLogger),
	))
	/* robfig/cron
	Entry                  | Description                                | Equivalent To
	-----                  | -----------                                | -------------
	@yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 1 1 *
	@monthly               | Run once a month, midnight, first of month | 0 0 1 * *
	@weekly                | Run once a week, midnight between Sat/Sun  | 0 0 * * 0
	@daily (or @midnight)  | Run once a day, midnight                   | 0 0 * * *
	@hourly                | Run once an hour, beginning of hour        | 0 * * * *
	*/
	c.AddFunc("@every 1m", func() {
		tasks, err := taskService.GetTasks(taskrepo.IntervalHour)
		if err != nil {
			log.Println(err)
		}

		for _, task := range tasks {
			go statisticsService.Scrape(task.SubRedditName, reddit_miner.CreatedWithinPast(task.PostsCreatedWithinPast))
		}
		log.Printf("Tasks: %#v\n", tasks)
	})
	return c
}
