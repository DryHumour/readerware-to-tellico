package normalize

// Result contains the outcome of a normalization pipeline step.
type Result struct {
	Value         string   // Value the current text.
	Roles         []string // Roles are the extracted canonical roles (e.g., "editor")
	Markers       []string // Makers are the extracted canonical marker (e.g., "signed")
	Literal       bool     // Literal indicates no further transforms should be applied.
	Discarded     bool     // Discarded indicates the result is explicitly discarded.
	RequiresAudit bool     // RequiresAudit indicates the Tellico entry should be marked for human review.
	AuditReasons  []string // AuditReasons are the reasons the entry needs review.
}

// Keep marks the result as literal (i.e. no further transforms are to be applied).
func (r Result) Keep() Result {
	r.Literal = true
	return r
}

// Discard marks the result as discarded and sets it to literal, discarding all
// roles and markers.
func (r Result) Discard() Result {
	r.Discarded = true
	r.Literal = true
	r.Value = ""
	r.Roles = nil
	r.Markers = nil
	return r
}

// Update returns a new Result with a modified value, perfectly preserving
// all previously extracted roles, markers, and audit flags.
// Intended for injecting Sprig template modifications back into the pipeline.
func (r Result) Update(newVal string) Result {
	r.Value = newVal
	return r
}

// AddAudit returns a new Result that has been manually flagged for Tellico review.
func (r Result) AddAudit(reason string) Result {
	r.RequiresAudit = true
	// Force a new allocation using Go's native append-growth math
	// by clamping the capacity of the original slice to its length.
	r.AuditReasons = append(r.AuditReasons[:len(r.AuditReasons):len(r.AuditReasons)], reason)
	return r
}

// isImmutable returns true if the result is immutable (literal or discarded).
// This is used to prevent further transformations on the result.
func (r Result) isImmutable() bool {
	return r.Literal || r.Discarded
}
