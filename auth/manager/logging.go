package manager

import (
	"encoding/json"
	"log"
	"time"

	"aigateway/auth/errors"
)

// StateLogger logs state changes for observability
type StateLogger struct {
	enabled bool
	prefix  string
}

// NewStateLogger creates a new state logger
func NewStateLogger(enabled bool) *StateLogger {
	return &StateLogger{
		enabled: enabled,
		prefix:  "[AuthManager]",
	}
}

// LogAccountLoaded logs when accounts are loaded
func (l *StateLogger) LogAccountLoaded(providerID string, count int) {
	if !l.enabled {
		return
	}
	log.Printf("%s Loaded %d accounts for provider %s", l.prefix, count, providerID)
}

// LogAccountSelected logs when an account is selected
func (l *StateLogger) LogAccountSelected(accountID, providerID, model string) {
	if !l.enabled {
		return
	}
	log.Printf("%s Selected account %s for %s/%s", l.prefix, accountID, providerID, model)
}

// LogAccountBlocked logs when an account becomes blocked
func (l *StateLogger) LogAccountBlocked(accountID, model string, reason BlockReason, until time.Time) {
	if !l.enabled {
		return
	}
	log.Printf("%s Account %s blocked for model %s, reason=%s, until=%s",
		l.prefix, accountID, model, reason, until.Format(time.RFC3339))
}

// LogAccountUnblocked logs when an account becomes available again
func (l *StateLogger) LogAccountUnblocked(accountID, model string) {
	if !l.enabled {
		return
	}
	log.Printf("%s Account %s unblocked for model %s", l.prefix, accountID, model)
}

// LogAccountDisabled logs when an account is disabled
func (l *StateLogger) LogAccountDisabled(accountID string, reason string) {
	if !l.enabled {
		return
	}
	log.Printf("%s Account %s disabled: %s", l.prefix, accountID, reason)
}

// LogAllBlocked logs when all accounts are blocked
func (l *StateLogger) LogAllBlocked(providerID, model string, retryAt time.Time) {
	if !l.enabled {
		return
	}
	log.Printf("%s All accounts blocked for %s/%s, retry at %s",
		l.prefix, providerID, model, retryAt.Format(time.RFC3339))
}

// LogSuccess logs a successful request
func (l *StateLogger) LogSuccess(accountID, model string) {
	if !l.enabled {
		return
	}
	log.Printf("%s Success: account=%s model=%s", l.prefix, accountID, model)
}

// LogFailure logs a failed request with error details
func (l *StateLogger) LogFailure(accountID, model string, err *errors.ParsedError) {
	if !l.enabled {
		return
	}
	log.Printf("%s Failure: account=%s model=%s type=%s code=%d cooldown=%s",
		l.prefix, accountID, model, err.Type, err.StatusCode, err.CooldownDur)
}

// LogTokenRefresh logs token refresh events
func (l *StateLogger) LogTokenRefresh(accountID string, success bool, err error) {
	if !l.enabled {
		return
	}
	if success {
		log.Printf("%s Token refreshed for account %s", l.prefix, accountID)
	} else {
		log.Printf("%s Token refresh failed for account %s: %v", l.prefix, accountID, err)
	}
}

// LogRetry logs retry attempts
func (l *StateLogger) LogRetry(providerID, model string, attempt int, reason string) {
	if !l.enabled {
		return
	}
	log.Printf("%s Retry attempt %d for %s/%s: %s", l.prefix, attempt, providerID, model, reason)
}

// LogStateChange logs generic state changes as JSON
func (l *StateLogger) LogStateChange(event string, data map[string]interface{}) {
	if !l.enabled {
		return
	}
	jsonData, _ := json.Marshal(data)
	log.Printf("%s %s: %s", l.prefix, event, string(jsonData))
}

// SetEnabled enables or disables logging
func (l *StateLogger) SetEnabled(enabled bool) {
	l.enabled = enabled
}
