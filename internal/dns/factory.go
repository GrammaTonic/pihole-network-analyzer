package dns

import (
	"pihole-analyzer/internal/logger"
)

// Factory implements the DNSServerFactory interface
type Factory struct {
	logger *logger.Logger
}

// NewFactory creates a new DNS server factory
func NewFactory(logger *logger.Logger) DNSServerFactory {
	return &Factory{
		logger: logger,
	}
}

// CreateServer creates a DNS server instance
func (f *Factory) CreateServer(config *Config) (DNSServer, error) {
	return NewServer(config, f.logger), nil
}

// CreateCache creates a DNS cache instance
func (f *Factory) CreateCache(config *CacheConfig) (DNSCache, error) {
	return NewCache(*config), nil
}

// CreateForwarder creates a DNS forwarder instance
func (f *Factory) CreateForwarder(config *ForwarderConfig) (DNSForwarder, error) {
	return NewForwarder(*config), nil
}

// CreateParser creates a DNS parser instance
func (f *Factory) CreateParser() DNSParser {
	return NewParser()
}
