package socketsrv

import "time"

type ConnConfig struct {
	IncomingBuffer     int
	OutgoingBuffer     int
	ReadBufferInitial  int
	WriteBufferInitial int
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	ResponseTimeout    time.Duration
	CleanupInterval    time.Duration
}

func (c ConnConfig) IsZero() bool {
	return c == ConnConfig{}
}

func DefaultConnConfig() ConnConfig {
	return ConnConfig{
		IncomingBuffer:     1024,
		OutgoingBuffer:     1024,
		WriteTimeout:       10 * time.Second,
		ReadTimeout:        10 * time.Second,
		ReadBufferInitial:  2048,
		WriteBufferInitial: 2048,
		CleanupInterval:    60 * time.Second,
		ResponseTimeout:    10 * time.Second,
	}
}

type ServerConfig struct {
	Conn ConnConfig
}

func (s ServerConfig) IsZero() bool {
	return s == ServerConfig{}
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Conn: DefaultConnConfig(),
	}
}

type ConnectorConfig struct {
	Conn        ConnConfig
	HaltTimeout time.Duration
}

func (c ConnectorConfig) IsZero() bool {
	return c == ConnectorConfig{}
}

func DefaultConnectorConfig() ConnectorConfig {
	return ConnectorConfig{
		Conn:        DefaultConnConfig(),
		HaltTimeout: 10 * time.Second,
	}
}
