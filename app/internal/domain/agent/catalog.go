package agent

import "strings"

// NewCatalog stellt die kleine, geprüfte POC-Auswahl bereit.
func NewCatalog() *Catalog {
	return &Catalog{
		templates: []Template{{
			ID: "recherche", Name: "Recherche", Role: "Rechercheagentin",
			Instructions: "Recherchiere und fasse belegte Ergebnisse zusammen.",
			Kind:         Eino, Capabilities: []string{"Wissensrecherche", "Zusammenfassung"},
		}},
		teams: []TeamTemplate{{ID: "rechercheteam", Name: "Rechercheteam", TemplateIDs: []string{"recherche"}}},
	}
}

// Templates kopiert Vorlagen, damit Aufrufer den Katalog nicht verändern.
func (c *Catalog) Templates() []Template {
	result := make([]Template, 0, len(c.templates))
	for _, template := range c.templates {
		template.Capabilities = append([]string(nil), template.Capabilities...)
		result = append(result, template)
	}

	return result
}

// TeamTemplates kopiert alle auswählbaren Teamvorlagen.
func (c *Catalog) TeamTemplates() []TeamTemplate {
	result := make([]TeamTemplate, 0, len(c.teams))
	for _, team := range c.teams {
		team.TemplateIDs = append([]string(nil), team.TemplateIDs...)
		result = append(result, team)
	}

	return result
}

// Find sucht ausschließlich nach einer veröffentlichten Vorlage.
func (c *Catalog) Find(id string) (Template, bool) {
	for _, template := range c.Templates() {
		if template.ID == strings.TrimSpace(id) {
			return template, true
		}
	}

	return Template{}, false
}

// Register ergänzt beim Bootstrap eine ausdrücklich kuratierte Adaptervorlage.
func (c *Catalog) Register(template Template) bool {
	if template.ID == "" || template.Name == "" || template.Kind == "" || !c.validCapabilities(template) {
		return false
	}

	if _, exists := c.Find(template.ID); exists {
		return false
	}

	template.Capabilities = append([]string(nil), template.Capabilities...)
	c.templates = append(c.templates, template)
	return true
}

func (c *Catalog) validCapabilities(template Template) bool {
	for _, capability := range template.Capabilities {
		if !c.safeCapability(capability) {
			return false
		}
	}

	return true
}

func (c *Catalog) safeCapability(capability string) bool {
	return capability == "Wissensrecherche" || capability == "Zusammenfassung"
}
