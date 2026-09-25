package aktivitaet

import (
	"bytes"
	"errors"
	"net/http"
	"strings"

	appaktivitaet "agentcontrolplane/app/internal/app/aktivitaet"
	apporganisation "agentcontrolplane/app/internal/app/organisation"
	domainaktivitaet "agentcontrolplane/app/internal/domain/aktivitaet"
	domainorganisation "agentcontrolplane/app/internal/domain/organisation"
)

func (h *Handler) pageList(w http.ResponseWriter, r *http.Request) {
	events, err := h.service.List(r.Context(), r.PathValue("id"))
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	organization, err := h.organizations.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		h.pageError(w, r, err)
		return
	}

	data := h.pageData(organization, "list")
	data["View"].(map[string]any)["Events"] = h.eventViews(events)
	h.render(w, r, http.StatusOK, data)
}

func (h *Handler) eventViews(events []domainaktivitaet.Event) []map[string]string {
	items := make([]map[string]string, 0, len(events))
	for _, event := range events {
		items = append(items, h.eventView(event))
	}

	return items
}

func (h *Handler) eventView(event domainaktivitaet.Event) map[string]string {
	return map[string]string{"ID": event.ID, "KindLabel": h.kindLabel(event.Kind),
		"ActorLabel": h.actorLabel(event.Actor), "SourceLabel": h.sourceLabel(event.Source),
		"OccurredAtISO":   event.OccurredAt.UTC().Format("2006-01-02T15:04:05Z"),
		"OccurredAtLabel": event.OccurredAt.UTC().Format("02.01.2006 15:04 UTC"),
		"ObjectTitle":     event.ObjectTitle, "AssigneeName": event.AssigneeName, "DeepLink": event.DeepLink}
}

func (h *Handler) kindLabel(kind string) string {
	if kind == "created" {
		return "Aufgabe angelegt"
	}

	if kind == "assigned" {
		return "Aufgabe zugewiesen"
	}

	return kind
}

func (h *Handler) actorLabel(actor string) string {
	if actor == "local-operator" {
		return "Lokaler Betreiber"
	}

	return actor
}

func (h *Handler) sourceLabel(source string) string {
	if source == "api" {
		return "API"
	}

	if source == "browser" {
		return "Browser"
	}

	return source
}

func (h *Handler) pageData(organization domainorganisation.Organization, kind string) map[string]any {
	return map[string]any{"PageTitle": "Aktivitätsverlauf", "Organization": map[string]string{"ID": organization.ID, "Name": organization.Name},
		"Navigation": []map[string]string{{"Key": "organisation", "Label": "Organisationen", "Href": "/organisationen", "Icon": "✳"},
			{"Key": "aktivitaet", "Label": "Aktivität", "Href": "/organisationen/" + organization.ID + "/aktivitaet", "Icon": "◷"}},
		"Errors": []string{}, "View": map[string]any{"Kind": kind}}
}

func (h *Handler) pageError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, appaktivitaet.ErrAccessDenied) || errors.Is(err, apporganisation.ErrAccessDenied) {
		h.pageFailure(w, r, http.StatusForbidden, "Zugriff verweigert")
		return
	}

	if errors.Is(err, appaktivitaet.ErrNotFound) || errors.Is(err, apporganisation.ErrNotFound) {
		h.pageFailure(w, r, http.StatusNotFound, "Organisation nicht gefunden")
		return
	}

	h.pageFailure(w, r, http.StatusInternalServerError, "Der Aktivitätsverlauf ist derzeit nicht verfügbar")
}

func (h *Handler) pageFailure(w http.ResponseWriter, r *http.Request, status int, message string) {
	data := h.pageData(domainorganisation.Organization{}, "error")
	data["Errors"] = []string{message}
	data["Navigation"] = []map[string]string{{"Key": "organisation", "Label": "Organisationen", "Href": "/organisationen", "Icon": "✳"}}
	h.render(w, r, status, data)
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, status int, data map[string]any) {
	if h.bridge == nil {
		http.Error(w, "Oberfläche nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	var output bytes.Buffer
	if err := h.bridge.Render(&output, h.templateName(r), data); err != nil {
		http.Error(w, "Oberfläche nicht verfügbar", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = output.WriteTo(w)
}

func (h *Handler) templateName(r *http.Request) string {
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		return "aktivitaet/content"
	}

	return "aktivitaet/page"
}
