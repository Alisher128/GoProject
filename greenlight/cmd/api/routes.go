package main

import (
	"expvar" // New import
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/games", app.requirePermission("games:read", app.listGamesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/games", app.requirePermission("games:write", app.createGameHandler))
	router.HandlerFunc(http.MethodGet, "/v1/games/:id", app.requirePermission("games:read", app.showGameHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/games/:id", app.requirePermission("games:write", app.updateGameHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/games/:id", app.requirePermission("games:write", app.deleteGameHandler))
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())
	// Use the new metrics() middleware at the start of the chain.
	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
