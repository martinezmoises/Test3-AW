package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/martinezmoises/Test3/internal/mailer"

	_ "github.com/lib/pq"
	"github.com/martinezmoises/Test3/internal/data"
)

const appVersion = "1.0.0"

type serverConfig struct {
	port        int
	environment string
	db          struct {
		dsn string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}

	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}

	cors struct {
		trustedOrigins []string
	}
}

type applicationDependencies struct {
	config           serverConfig
	logger           *slog.Logger
	bookModel        data.BookModel
	readingListModel data.ReadingListModel
	reviewModel      data.ReviewModel // Add reviewModel
	userModel        data.UserModel
	mailer           mailer.Mailer
	wg               sync.WaitGroup
	tokenModel       data.TokenModel
}

func main() {
	var settings serverConfig

	flag.IntVar(&settings.port, "port", 4000, "Server port")
	flag.StringVar(&settings.environment, "env", "development", "Environment(development|staging|production)")
	flag.StringVar(&settings.db.dsn, "db-dsn", "postgres://bookclub:fishsticks@localhost/bookclub?sslmode=disable", "PostgreSQL DSN")
	flag.StringVar(&settings.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	// We have port 25, 465, 587, 2525. If 25 doesn't work choose another
	flag.IntVar(&settings.smtp.port, "smtp-port", 2525, "SMTP port")
	// Use your Username value provided by Mailtrap
	flag.StringVar(&settings.smtp.username, "smtp-username", "acc22bc2a9b178", "SMTP username")
	// Use your Password value provided by Mailtrap
	flag.StringVar(&settings.smtp.password, "smtp-password", "9dc66a20f67fb7", "SMTP password")
	flag.StringVar(&settings.smtp.sender, "smtp-sender", "Books Club <no-reply@booksclub.example.com>", "SMTP sender")

	flag.Float64Var(&settings.limiter.rps, "limiter-rps", 2, "Rate Limiter maximum requests per second")
	flag.IntVar(&settings.limiter.burst, "limiter-burst", 5, "Rate Limiter maximum burst")
	flag.BoolVar(&settings.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)",
		func(val string) error {
			settings.cors.trustedOrigins = strings.Fields(val)
			return nil
		})

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// the call to openDB() sets up our connection pool
	db, err := openDB(settings)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// release the database resources before exiting
	defer db.Close()

	logger.Info("database connection pool established")

	appInstance := &applicationDependencies{
		config:           settings,
		logger:           logger,
		bookModel:        data.BookModel{DB: db},        // Initialize BookModel
		readingListModel: data.ReadingListModel{DB: db}, // Initialize ReadingListModel
		reviewModel:      data.ReviewModel{DB: db},      // Initialize ReviewModel
		userModel:        data.UserModel{DB: db},        // Initialize UserModel
		mailer:           mailer.New(settings.smtp.host, settings.smtp.port, settings.smtp.username, settings.smtp.password, settings.smtp.sender),
		tokenModel:       data.TokenModel{DB: db}, // Initialize TokenModel
	}

	//router := http.NewServeMux()
	//router.HandleFunc("/v1/healthcheck", appInstance.healthCheckHandler)

	err = appInstance.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(settings serverConfig) (*sql.DB, error) {
	// open a connection pool
	db, err := sql.Open("postgres", settings.db.dsn)
	if err != nil {
		return nil, err
	}

	// set a context to ensure DB operations don't take too long
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// return the connection pool (sql.DB)
	return db, nil
}
