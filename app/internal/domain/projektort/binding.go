package projektort

// Valid erlaubt ausschließlich den derzeit belegten privaten Ortstyp.
func (b Binding) Valid() bool {
	return b.OrganizationID != "" && b.ProjectID != "" && b.Kind == ManagedDirectory
}

// Validate prüft den vom Client wählbaren Ortstyp.
func (b Binding) Validate() error {
	if !b.Valid() {
		return ErrKindInvalid
	}

	return nil
}
