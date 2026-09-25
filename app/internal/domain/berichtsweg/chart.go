package berichtsweg

// ParentOf findet den unmittelbaren Vorgesetzten, ohne weitere Rechte abzuleiten.
func (c Chart) ParentOf(agentID string) (string, bool) {
	for _, line := range c.Lines {
		if line.AgentID == agentID {
			return line.ParentID, true
		}
	}

	return "", false
}

// ChildrenOf zeigt direkte Berichtsempfänger als mögliche, nicht freigegebene Kandidaten.
func (c Chart) ChildrenOf(parentID string) []Line {
	children := []Line{}
	for _, line := range c.Lines {
		if line.ParentID == parentID && line.AgentID != parentID {
			children = append(children, line)
		}
	}

	return children
}
