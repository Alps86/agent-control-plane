package datenbereich

import (
	"context"
	"errors"

	domainagent "agentcontrolplane/app/internal/domain/agent"
	domaindatenbereich "agentcontrolplane/app/internal/domain/datenbereich"
	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
)

var ErrNotFound = errors.New("not_found")

// Store prüft die Datenrechte bei jeder konkreten Projektoperation erneut.
type Store interface {
	ReplaceScopes(context.Context, string, string, string, []domaindatenbereich.Scope) error
	ListScopes(context.Context, string, string, string) ([]domaindatenbereich.Scope, error)
	ReadProjectForAgent(context.Context, string, string, string) (domainprojekt.Project, error)
	UpdateProjectDescriptionForAgent(context.Context, string, string, string, string) (domainprojekt.Project, error)
	ListProjectsForOperator(context.Context, string, string, string) ([]domainprojekt.Project, error)
	FindAgentForOperator(context.Context, string, string, string) (domainagent.Agent, error)
	AuditDenied(context.Context, string, string, string, domaindatenbereich.Action) error
	ListDenials(context.Context, string, string, string) ([]domaindatenbereich.Denial, error)
}
