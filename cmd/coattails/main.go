package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/routes"
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/sundress"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func initDeps() time.Duration {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var (
		wait        time.Duration
		pgHost      string
		pgPort      int
		pgUser      string
		pgPwd       string
		pgDb        string
		cacheHost   string
		debugNoDeps bool
	)

	flag.DurationVar(&wait,
		"graceful-timeout",
		time.Second*15,
		"the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.StringVar(&pgHost, "pg-host", "localhost", "postgresql host name")
	flag.IntVar(&pgPort, "pg-port", 5432, "postgresql port")
	flag.StringVar(&pgUser, "pg-user", "postgres", "postgresql user")
	flag.StringVar(&pgPwd, "pg-pwd", "bdc", "postgresql password")
	flag.StringVar(&pgDb, "pg-db", "wardrobe", "postgresql db")
	flag.StringVar(&cacheHost, "redis-host", "localhost", "redis host")
	flag.BoolVar(&debugNoDeps, "run-without-deps", false, "debug setting")
	//first arg is a pointer, second arg is the value we are checking for, third value is what we set if we don't see the flag, fourth is description
	flag.Parse()
	// Initialize singleton instances after parsing flag

	sundress.InitSecret()
	stockings.InitKeygen()
	if debugNoDeps {
		log.Println("Warning: You are starting a server without a Database and Cache")
		log.Println("Calls to functions that use a Database or Cache will segfault")
	} else {
		wardrobe.InitDB(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			pgHost, pgPort, pgUser, pgPwd, pgDb))
		wardrobe.InitCache(cacheHost)
	}

	return wait
}

func main() {
	wait := initDeps()
	r := mux.NewRouter().StrictSlash(true)
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://bdc-web:38001"},
		AllowCredentials: true,
	}).Handler(r)

	routes.RegisterAllRoutes(r)

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handler, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	err := srv.Shutdown(ctx)
	if err != nil {
		panic(err)
	}

	err = wardrobe.CloseDB()
	if err != nil {
		panic(err)
	}
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}
