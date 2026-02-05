package security

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// AuditEventType represents the type of security audit event
type AuditEventType string

const (
	AuditSSHSignAllowed  AuditEventType = "ssh_sign_allowed"
	AuditSSHSignDenied   AuditEventType = "ssh_sign_denied"
	AuditSSHKeyListed    AuditEventType = "ssh_key_listed"
	AuditSSHKeyFiltered  AuditEventType = "ssh_key_filtered"
	AuditGPGSignAllowed  AuditEventType = "gpg_sign_allowed"
	AuditGPGSignDenied   AuditEventType = "gpg_sign_denied"
	AuditGPGDecryptAllow AuditEventType = "gpg_decrypt_allowed"
	AuditGPGDecryptDeny  AuditEventType = "gpg_decrypt_denied"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
	Timestamp time.Time      `json:"timestamp"`
	Type      AuditEventType `json:"type"`
	KeyID     string         `json:"key_id,omitempty"`
	Comment   string         `json:"comment,omitempty"`
	Allowed   bool           `json:"allowed"`
	Reason    string         `json:"reason,omitempty"`
}

// AuditLogger handles security audit logging
type AuditLogger struct {
	mu      sync.Mutex
	file    *os.File
	enabled bool
}

var (
	defaultAuditLogger *AuditLogger
	auditLoggerOnce    sync.Once
)

// GetAuditLogger returns the singleton audit logger instance
func GetAuditLogger() *AuditLogger {
	auditLoggerOnce.Do(func() {
		defaultAuditLogger = &AuditLogger{enabled: false}
	})
	return defaultAuditLogger
}

// EnableAuditLog enables audit logging to the specified file
func EnableAuditLog(path string) error {
	logger := GetAuditLogger()
	logger.mu.Lock()
	defer logger.mu.Unlock()

	if logger.file != nil {
		logger.file.Close()
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open audit log: %w", err)
	}

	logger.file = file
	logger.enabled = true
	return nil
}

// DisableAuditLog disables audit logging
func DisableAuditLog() {
	logger := GetAuditLogger()
	logger.mu.Lock()
	defer logger.mu.Unlock()

	if logger.file != nil {
		logger.file.Close()
		logger.file = nil
	}
	logger.enabled = false
}

// LogEvent logs an audit event
func (l *AuditLogger) LogEvent(event AuditEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.enabled || l.file == nil {
		return
	}

	event.Timestamp = time.Now().UTC()
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	l.file.Write(data)
	l.file.Write([]byte("\n"))
}

// LogSSHSign logs an SSH signing operation
func LogSSHSign(keyComment string, allowed bool, reason string) {
	eventType := AuditSSHSignAllowed
	if !allowed {
		eventType = AuditSSHSignDenied
	}

	GetAuditLogger().LogEvent(AuditEvent{
		Type:    eventType,
		Comment: keyComment,
		Allowed: allowed,
		Reason:  reason,
	})
}

// LogSSHKeyAccess logs SSH key listing/filtering
func LogSSHKeyAccess(keyComment string, allowed bool) {
	eventType := AuditSSHKeyListed
	if !allowed {
		eventType = AuditSSHKeyFiltered
	}

	GetAuditLogger().LogEvent(AuditEvent{
		Type:    eventType,
		Comment: keyComment,
		Allowed: allowed,
	})
}

// LogGPGSign logs a GPG signing operation
func LogGPGSign(keyID string, allowed bool, reason string) {
	eventType := AuditGPGSignAllowed
	if !allowed {
		eventType = AuditGPGSignDenied
	}

	GetAuditLogger().LogEvent(AuditEvent{
		Type:    eventType,
		KeyID:   keyID,
		Allowed: allowed,
		Reason:  reason,
	})
}

// LogGPGDecrypt logs a GPG decryption operation
func LogGPGDecrypt(keyID string, allowed bool, reason string) {
	eventType := AuditGPGDecryptAllow
	if !allowed {
		eventType = AuditGPGDecryptDeny
	}

	GetAuditLogger().LogEvent(AuditEvent{
		Type:    eventType,
		KeyID:   keyID,
		Allowed: allowed,
		Reason:  reason,
	})
}
