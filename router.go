package main

import (
	"database/sql"
	"net/http"
	"reflect"

	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
)

func CreateRouter(db *sql.DB, rdb *redis.Client) *httprouter.Router {
	router := httprouter.New()

	// controllers
	controller := CreateController(db, rdb)

	// middleware
	auth := CreateRequireTokenMiddleware(rdb).Middleware
	body := CreateBodyParserMiddleware().Middleware

	// shorthand reflect TypeOf
	typeOf := reflect.TypeOf

	// app routes
	router.GET("/", controller.HealthCheck)
	router.GET("/version", controller.Version)
	router.GET("/robots.txt", controller.Robots)

	// unit routes
	router.PUT("/unit/:id/toggle-lock", auth(controller.UnitToggleLock))

	// user routes
	router.POST("/user/sign-up", body(typeOf(SignUpReq{}), controller.SignUp))
	router.POST("/user/sign-in", body(typeOf(SignInReq{}), controller.SignIn))
	router.PUT("/user/rename", auth(body(typeOf(UserRenameReq{}), controller.UserRename)))

	// method not allowed
	router.MethodNotAllowed = http.HandlerFunc(controller.MethodNotAllowed)

	// not found handler
	router.NotFound = http.HandlerFunc(controller.NotFound)

	return router
}
