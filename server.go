package main

import (
	"context"
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DataBase struct {
	client *mongo.Client
}

func (d *DataBase) connect() {
	mongodbURI := os.Getenv("MONGODB_URI")
	if mongodbURI == "" {
		err := errors.New("MONGODB_URI variable is not defined")
		logger(err)
		panic(err)
	}
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongodbURI).SetServerAPIOptions(serverAPI)
	var err error
	d.client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		logger(err)
		panic(err)
	}
}

// Should be defered
func (d *DataBase) disconnect() {
	err := d.client.Disconnect(context.TODO())
	if err != nil {
		logger(err)
		panic(err)
	}
}

var db DataBase

func main() {
	// customLogger = setLog(zerolog.DebugLevel)
	setLog(zerolog.DebugLevel)

	// Load env variables from file
	err := godotenv.Load()
	if err != nil {
		logger(err)
	}

	// Start server
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	// AllowOrigins: []string{"http://localhost:3000", "http://10.0.0.11:3000"},
	// 	// AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	// 	// AllowMethods: []string{http.MethodPost, http.MethodGet, http.MethodPut, http.MethodDelete},
	// 	AllowOrigins: []string{"*"},
	// 	AllowHeaders: []string{"*"},
	// 	AllowMethods: []string{"*"},
	// }))

	// Requests logs
	e.Use(middleware.RequestLoggerWithConfig(RequestLoggerConfig))

	// Routes
	e.POST("/tenants", postTenant)
	e.GET("/tenants", getTenant)
	e.PUT("/tenants/:id", putTenant)
	e.DELETE("/tenants/:id", deleteTenant)

	// Data base connection
	// mongodbURI := os.Getenv("MONGODB_URI")
	// serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	// opts := options.Client().ApplyURI(mongodbURI).SetServerAPIOptions(serverAPI)
	// client, err = mongo.Connect(context.TODO(), opts)
	// if err != nil {
	// 	logger(err)
	// 	panic(err)
	// }

	// defer func() {
	// 	if err = client.Disconnect(context.TODO()); err != nil {
	// 		logger(err)
	// 		panic(err)
	// 	}
	// }()
	// defer connect()()
	db = DataBase{}
	db.connect()

	defer db.disconnect()

	logger("Pinged your deployment. You successfully connected to MongoDB!")

	port := os.Getenv("PORT")
	e.Logger.Fatal(e.Start(port))
}

// func connect() func() {
// 	mongodbURI := os.Getenv("MONGODB_URI")
// 	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
// 	opts := options.Client().ApplyURI(mongodbURI).SetServerAPIOptions(serverAPI)
// 	var err error
// 	client, err = mongo.Connect(context.TODO(), opts)
// 	if err != nil {
// 		logger(err)
// 		panic(err)
// 	}

// 	return func() {
// 		if err = client.Disconnect(context.TODO()); err != nil {
// 			logger(err)
// 			panic(err)
// 		}
// 	}
// }
