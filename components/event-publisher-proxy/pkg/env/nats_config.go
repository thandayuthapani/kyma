package env

import (
	"fmt"
	"time"
)

// compile time check
var _ fmt.Stringer = &NatsConfig{}

// NatsConfig represents the environment config for the Event Publisher to NATS.
type NatsConfig struct {
	Port                 int           `envconfig:"INGRESS_PORT" default:"8080"`
	URL                  string        `envconfig:"NATS_URL" default:"nats.nats.svc.cluster.local"`
	RetryOnFailedConnect bool          `envconfig:"RETRY_ON_FAILED_CONNECT" default:"true"`
	MaxReconnects        int           `envconfig:"MAX_RECONNECTS" default:"-1"` // Negative means keep try reconnecting.
	ReconnectWait        time.Duration `envconfig:"RECONNECT_WAIT" default:"5s"`
	RequestTimeout       time.Duration `envconfig:"REQUEST_TIMEOUT" default:"5s"`

	// Legacy Namespace is used as the event source for legacy events
	LegacyNamespace string `envconfig:"LEGACY_NAMESPACE" default:"kyma"`
	// EventTypePrefix is the prefix of each event as per the eventing specification
	// It follows the eventType format: <eventTypePrefix>.<appName>.<event-name>.<version>
	EventTypePrefix string `envconfig:"EVENT_TYPE_PREFIX" default:"kyma"`

	// JetStream-specific configs
	JSStreamName          string `envconfig:"JS_STREAM_NAME" default:"kyma"`
	JSStreamSubjectPrefix string `envconfig:"JS_STREAM_SUBJECT_PREFIX" required:"true"`
}

// ToConfig converts to a default BEB BebConfig
func (c *NatsConfig) ToConfig() *BebConfig {
	cfg := &BebConfig{
		BEBNamespace:    c.LegacyNamespace,
		EventTypePrefix: c.EventTypePrefix,
	}
	return cfg
}

// String implements the fmt.Stringer interface
func (c *NatsConfig) String() string {
	return fmt.Sprintf("%#v", c)
}
