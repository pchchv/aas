package audit

type AuditEvent struct {
	AuditEvent string                 `json:"audit_event"`
	Details    map[string]interface{} `json:"details"`
}
