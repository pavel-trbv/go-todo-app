package main

import (
	"context"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pavel-trbv/go-todo-app/internal/handler"
	"github.com/pavel-trbv/go-todo-app/internal/repository"
	"github.com/pavel-trbv/go-todo-app/internal/server"
	"github.com/pavel-trbv/go-todo-app/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variables: %s", err.Error())
	}

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: os.Getenv("DB_PASSWORD"),
	})
	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	srv := new(server.Server)
	go func() {
		err := srv.Run(viper.GetString("port"), handlers.InitRoutes())
		if err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("error occured while running http server: %s", err)
		}
	}()

	logrus.Println("Todo App started")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Println("Todo App is shutting down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured when server was shutting down: %s", err.Error())
	}
	if err := db.Close(); err != nil {
		logrus.Errorf("error occured when db was closing connection: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
