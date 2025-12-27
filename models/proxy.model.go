package models

import "time"

// ProxyProtocol represents proxy protocol type
type ProxyProtocol string

const (
	ProxyProtocolHTTP   ProxyProtocol = "http"
	ProxyProtocolHTTPS  ProxyProtocol = "https"
	ProxyProtocolSOCKS5 ProxyProtocol = "socks5"
)

// HealthStatus represents proxy health state
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "healthy"
	HealthStatusDegraded HealthStatus = "degraded"
	HealthStatusDown     HealthStatus = "down"
)

// Proxy represents a proxy server
type Proxy struct {
	ID                  int           `gorm:"primaryKey;autoIncrement" json:"id"`
	URL                 string        `gorm:"size:255;not null;unique" json:"url"`
	Protocol            ProxyProtocol `gorm:"type:enum('http','https','socks5');default:'http'" json:"protocol"`
	IsActive            bool          `gorm:"default:true;index:idx_active_health" json:"is_active"`
	HealthStatus        HealthStatus  `gorm:"type:enum('healthy','degraded','down');default:'healthy';index:idx_active_health" json:"health_status"`
	MaxAccounts         int           `gorm:"default:0" json:"max_accounts"`
	CurrentAccounts     int           `gorm:"default:0;index:idx_capacity" json:"current_accounts"`
	LastUsedAt          *time.Time    `json:"last_used_at"`
	UsageCount          int64         `gorm:"default:0" json:"usage_count"`
	Priority            int           `gorm:"default:0;index:idx_priority" json:"priority"`
	Weight              int           `gorm:"default:1" json:"weight"`
	ConsecutiveFailures int           `gorm:"default:0" json:"consecutive_failures"`
	MaxFailures         int           `gorm:"default:3" json:"max_failures"`
	SuccessRate         float64       `gorm:"type:decimal(5,2);default:100.00" json:"success_rate"`
	AvgLatencyMs        int           `gorm:"default:0" json:"avg_latency_ms"`
	LastCheckedAt       *time.Time    `json:"last_checked_at"`
	MarkedDownAt        *time.Time    `gorm:"index" json:"marked_down_at"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

func (Proxy) TableName() string {
	return "proxy_pool"
}

// ProxyStats represents daily aggregated statistics
type ProxyStats struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProxyID        int       `gorm:"not null;index:idx_proxy_provider_date" json:"proxy_id"`
	ProviderID     *string   `gorm:"size:50;index:idx_proxy_provider_date" json:"provider_id"`
	RequestCount   int       `gorm:"default:0" json:"request_count"`
	SuccessCount   int       `gorm:"default:0" json:"success_count"`
	ErrorCount     int       `gorm:"default:0" json:"error_count"`
	TotalLatencyMs int64     `gorm:"default:0" json:"total_latency_ms"`
	Date           time.Time `gorm:"type:date;not null;index:idx_proxy_provider_date,idx_date" json:"date"`

	Proxy    *Proxy    `gorm:"foreignKey:ProxyID" json:"proxy,omitempty"`
	Provider *Provider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
}

func (ProxyStats) TableName() string {
	return "proxy_stats"
}

// RequestLog represents audit trail
type RequestLog struct {
	ID                   int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProviderID           *string   `gorm:"size:50;index:idx_provider_account" json:"provider_id"`
	AccountID            *string   `gorm:"size:36;index:idx_provider_account,idx_account" json:"account_id"`
	ProxyID              *int      `json:"proxy_id"`
	Model                string    `gorm:"size:100" json:"model"`
	StatusCode           int       `json:"status_code"`
	LatencyMs            int       `json:"latency_ms"`
	RetryCount           int       `gorm:"default:0" json:"retry_count"`
	SwitchedFromAccountID *string  `gorm:"size:36" json:"switched_from_account_id,omitempty"`
	Error                string    `gorm:"type:text" json:"error"`
	CreatedAt            time.Time `gorm:"index:idx_created" json:"created_at"`
}

func (RequestLog) TableName() string {
	return "request_logs"
}
