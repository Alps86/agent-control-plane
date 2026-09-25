package codexprofil

// Validate verhindert Schreibrechte ohne aktivierten Arbeitsbereich.
func (i Input) Validate() error {
	if i.WriteEnabled && !i.WorkspaceEnabled {
		return ErrWriteWithoutWorkspace
	}

	return nil
}
