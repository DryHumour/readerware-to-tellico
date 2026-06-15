package collection

import "github.com/DryHumour/readerware-to-tellico/internal/normalize"

var (
	_ Auditor = (*auditor)(nil)
)

// auditor implements Auditor.
type auditor struct {
	audit   bool
	reasons []string
}

// RequiresAudit reports whether this entry was flagged for human review.
func (a *auditor) RequiresAudit() bool {
	return a.audit
}

// AuditReasons returns the list of reasons why this entry needs human review.
func (a *auditor) AuditReasons() []string {
	return a.reasons
}

// AddAudit flags the entry for audit with optional reasons.
func (a *auditor) AddAudit(reasons ...string) (_ Nothing) {
	a.audit = true
	a.reasons = append(a.reasons, reasons...)
	return
}

// AddAuditResult adds audit information from a normalize.Result.
func (a *auditor) AddAuditResult(r normalize.Result) (_ Nothing) {
	a.audit = a.audit || r.RequiresAudit
	a.reasons = append(a.reasons, r.AuditReasons...)
	return
}
