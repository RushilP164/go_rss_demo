package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/rushilp164/go_rss_demo/internal/database"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	godotenv.Load()

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT not found")
	}

	dbUrl := os.Getenv("DB_URL")
	log.Println(dbUrl)
	if dbUrl == "" {
		log.Fatal("DB URL not found")
	}

	conn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Can't connect to database")
	}

	apiConf := apiConfig{
		DB: database.New(conn),
	}

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()
	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", handlerError)
	v1Router.Post("/users", apiConf.handlerUsersCreate)
	v1Router.Get("/users/by-apikey", apiConf.middlewareAuth(apiConf.handlerGetUser))
	v1Router.Post("/feeds", apiConf.middlewareAuth(apiConf.handlerFeedsCreate))
	v1Router.Get("/feeds", apiConf.handlerGetFeeds)
	v1Router.Post("/feed-follows", apiConf.middlewareAuth(apiConf.handlerCreateFeedFollows))
	v1Router.Get("/feed-follows", apiConf.middlewareAuth(apiConf.handlerGetFeedFollows))
	v1Router.Delete("/feed-follows/{feedFollowID}", apiConf.middlewareAuth(apiConf.handlerDeleteFeedFollows))
	v1Router.Get("/posts", apiConf.middlewareAuth(apiConf.handlerGetPosts))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	// Start scrapping
	const collectionConcurrency = 10
	const collectionInterval = 5 * time.Minute
	go startScraping(apiConf.DB, collectionConcurrency, collectionInterval)

	log.Printf("Sever started runnign on %v", portString)
	srvErr := srv.ListenAndServe()
	if srvErr != nil {
		log.Fatal(err)
	}
}
