package main

import (
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

func CreateRouter(controller *Controller) *httprouter.Router {
	router := httprouter.New()

	// middleware
	auth := CreateRequireTokenMiddleware(controller.rdb).Middleware
	body := CreateBodyParserMiddleware().Middleware

	// shorthand reflect TypeOf
	typeOf := reflect.TypeOf

	// app routes
	router.GET("/", controller.HealthCheck)
	router.GET("/version", controller.Version)
	router.GET("/robots.txt", controller.Robots)

	// campaign routes
	router.PUT("/campaign/collect", auth(controller.CampaignCollect))

	// chat routes
	router.GET("/chat/message/history", auth(controller.ChatMessageHistory))
	router.POST("/chat/message/send", auth(body(typeOf(ChatMessageSendReq{}), controller.ChatMessageSend)))

	// daily quest routes
	router.PUT("/daily-quest/:id/complete", auth(controller.DailyQuestComplete))

	// summon routes
	router.PUT("/summon/unit", auth(controller.SummonUnit))

	// unit routes
	router.PUT("/unit/:id/toggle-lock", auth(controller.UnitToggleLock))

	// user routes
	router.POST("/user/sign-up", body(typeOf(SignUpReq{}), controller.SignUp))
	router.POST("/user/sign-in", body(typeOf(SignInReq{}), controller.SignIn))
	router.PUT("/user/rename", auth(body(typeOf(UserRenameReq{}), controller.UserRename)))

	// WebSocket upgrade route
	router.GET("/ws", auth(controller.WebSocketConnectionHandler))

	// method not allowed
	router.MethodNotAllowed = http.HandlerFunc(controller.MethodNotAllowed)

	// not found handler
	router.NotFound = http.HandlerFunc(controller.NotFound)

	return router
}
