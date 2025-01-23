package audit

type AuditEvent struct {
	AuditEvent string                 `json:"audit_event"`
	Details    map[string]interface{} `json:"details"`
}

type AuditLogger struct {
}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}
