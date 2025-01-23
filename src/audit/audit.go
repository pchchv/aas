package audit

import (
	"encoding/json"
	"log/slog"

	"github.com/pchchv/aas/src/config"
)

type AuditEvent struct {
	AuditEvent string                 `json:"audit_event"`
	Details    map[string]interface{} `json:"details"`
}

type AuditLogger struct {
}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}

func (al *AuditLogger) Log(auditEvent string, details map[string]interface{}) {
	if !config.Get().AuditLogsInConsole {
		return
	}

	evt := AuditEvent{
		AuditEvent: auditEvent,
		Details:    details,
	}

	if eventJSON, err := json.Marshal(evt); err != nil {
		slog.Error("failed to marshal audit event", "error", err, "event", auditEvent)
		return
	} else {
		slog.Info(string(eventJSON))
	}
}
