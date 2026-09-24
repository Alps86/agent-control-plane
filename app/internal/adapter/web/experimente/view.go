package experimente

func (i inventory) viewMap() map[string]any {
	groups := make([]map[string]any, 0, len(i.Groups))
	for _, group := range i.Groups {
		groups = append(groups, group.viewMap())
	}

	return map[string]any{
		"PageTitle":  "Paperclip-Experimente",
		"Navigation": []any{}, "Notice": "", "Errors": []any{}, "AllowedActions": []any{},
		"View": map[string]any{"Kind": "experimente", "SourceAsOf": i.SourceAsOf, "Groups": groups},
	}
}

func (g group) viewMap() map[string]any {
	entries := make([]map[string]any, 0, len(g.Entries))
	for _, entry := range g.Entries {
		entries = append(entries, entry.viewMap())
	}

	return map[string]any{"Key": g.Key, "Heading": g.Heading, "Entries": entries}
}

func (e entry) viewMap() map[string]any {
	links := make([]map[string]string, 0, len(e.SourceLinks))
	for _, link := range e.SourceLinks {
		links = append(links, map[string]string{"Label": link.Label, "URL": link.URL})
	}

	parts := make([]map[string]string, 0, len(e.DecisionParts))
	for _, part := range e.DecisionParts {
		parts = append(parts, map[string]string{"Kind": part.Kind, "Reason": part.Reason})
	}

	return map[string]any{"SourceKey": e.SourceKey, "Name": e.Name, "SourceLinks": links,
		"PaperclipMaturity": e.PaperclipMaturity, "Benefit": e.Benefit,
		"ACPDependencies": e.ACPDependencies, "DecisionParts": parts,
		"ACPImplementationStatus": e.ACPImplementationStatus, "RelatedSourceKey": e.RelatedSourceKey}
}
