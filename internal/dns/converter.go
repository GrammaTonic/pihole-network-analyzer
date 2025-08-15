package dns

import (
	"pihole-analyzer/internal/types"
	"time"
)

// ConvertConfig converts types.DNSConfig to dns.Config
func ConvertConfig(typesConfig types.DNSConfig) *Config {
	return &Config{
		Enabled:    typesConfig.Enabled,
		Host:       typesConfig.Host,
		Port:       typesConfig.Port,
		TCPEnabled: typesConfig.TCPEnabled,
		UDPEnabled: typesConfig.UDPEnabled,

		ReadTimeout:  time.Duration(typesConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(typesConfig.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(typesConfig.IdleTimeout) * time.Second,

		Cache: CacheConfig{
			Enabled:         typesConfig.Cache.Enabled,
			MaxSize:         typesConfig.Cache.MaxSize,
			DefaultTTL:      time.Duration(typesConfig.Cache.DefaultTTL) * time.Second,
			MaxTTL:          time.Duration(typesConfig.Cache.MaxTTL) * time.Second,
			MinTTL:          time.Duration(typesConfig.Cache.MinTTL) * time.Second,
			CleanupInterval: time.Duration(typesConfig.Cache.CleanupInterval) * time.Second,
			EvictionPolicy:  typesConfig.Cache.EvictionPolicy,
			MaxMemoryMB:     typesConfig.Cache.MaxMemoryMB,
		},

		Forwarder: ForwarderConfig{
			Enabled:        typesConfig.Forwarder.Enabled,
			Upstreams:      typesConfig.Forwarder.Upstreams,
			Timeout:        time.Duration(typesConfig.Forwarder.Timeout) * time.Second,
			Retries:        typesConfig.Forwarder.Retries,
			HealthCheck:    typesConfig.Forwarder.HealthCheck,
			HealthInterval: time.Duration(typesConfig.Forwarder.HealthInterval) * time.Second,
			LoadBalancing:  typesConfig.Forwarder.LoadBalancing,
			EDNS0Enabled:   typesConfig.Forwarder.EDNS0Enabled,
			UDPSize:        typesConfig.Forwarder.UDPSize,
		},

		LogQueries:           typesConfig.LogQueries,
		LogLevel:             typesConfig.LogLevel,
		MaxConcurrentQueries: typesConfig.MaxConcurrentQueries,
		BufferSize:           typesConfig.BufferSize,
	}
}

// ConvertToTypesConfig converts dns.Config to types.DNSConfig
func ConvertToTypesConfig(dnsConfig *Config) types.DNSConfig {
	return types.DNSConfig{
		Enabled:    dnsConfig.Enabled,
		Host:       dnsConfig.Host,
		Port:       dnsConfig.Port,
		TCPEnabled: dnsConfig.TCPEnabled,
		UDPEnabled: dnsConfig.UDPEnabled,

		ReadTimeout:  int(dnsConfig.ReadTimeout.Seconds()),
		WriteTimeout: int(dnsConfig.WriteTimeout.Seconds()),
		IdleTimeout:  int(dnsConfig.IdleTimeout.Seconds()),

		Cache: types.DNSCacheConfig{
			Enabled:         dnsConfig.Cache.Enabled,
			MaxSize:         dnsConfig.Cache.MaxSize,
			DefaultTTL:      int(dnsConfig.Cache.DefaultTTL.Seconds()),
			MaxTTL:          int(dnsConfig.Cache.MaxTTL.Seconds()),
			MinTTL:          int(dnsConfig.Cache.MinTTL.Seconds()),
			CleanupInterval: int(dnsConfig.Cache.CleanupInterval.Seconds()),
			EvictionPolicy:  dnsConfig.Cache.EvictionPolicy,
			MaxMemoryMB:     dnsConfig.Cache.MaxMemoryMB,
		},

		Forwarder: types.DNSForwarderConfig{
			Enabled:        dnsConfig.Forwarder.Enabled,
			Upstreams:      dnsConfig.Forwarder.Upstreams,
			Timeout:        int(dnsConfig.Forwarder.Timeout.Seconds()),
			Retries:        dnsConfig.Forwarder.Retries,
			HealthCheck:    dnsConfig.Forwarder.HealthCheck,
			HealthInterval: int(dnsConfig.Forwarder.HealthInterval.Seconds()),
			LoadBalancing:  dnsConfig.Forwarder.LoadBalancing,
			EDNS0Enabled:   dnsConfig.Forwarder.EDNS0Enabled,
			UDPSize:        dnsConfig.Forwarder.UDPSize,
		},

		LogQueries:           dnsConfig.LogQueries,
		LogLevel:             dnsConfig.LogLevel,
		MaxConcurrentQueries: dnsConfig.MaxConcurrentQueries,
		BufferSize:           dnsConfig.BufferSize,
	}
}

// GetDefaultTypesConfig returns a default DNS configuration for types.DNSConfig
func GetDefaultTypesConfig() types.DNSConfig {
	defaultConfig := DefaultConfig()
	return ConvertToTypesConfig(defaultConfig)
}
