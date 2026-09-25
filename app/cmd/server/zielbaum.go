package main

import (
	"agentcontrolplane/app/internal/adapter/web"
	webziel "agentcontrolplane/app/internal/adapter/web/ziel"
)

func (b *Bootstrap) mountGoalTree(server *web.Server, goals *webziel.Handler) {
	for _, pattern := range []string{
		"GET /api/organisationen/{id}/zielbaum",
		"POST /api/organisationen/{id}/ziele/{zielID}/kinder",
		"POST /api/organisationen/{id}/ziele/{zielID}/status",
		"POST /api/organisationen/{id}/ziele/{zielID}/verschieben",
		"POST /organisationen/{id}/ziele/{zielID}/kinder",
		"POST /organisationen/{id}/ziele/{zielID}/status",
		"POST /organisationen/{id}/ziele/{zielID}/verschieben",
	} {
		server.Handle(pattern, goals)
	}
}
