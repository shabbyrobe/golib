package socketsrv

import "time"

type ConnConfig struct {
	IncomingBuffer     int
	OutgoingBuffer     int
	ReadBufferInitial  int
	WriteBufferInitial int
	HeartbeatInterval  time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration

	// ResponseTimeout declares how long the connection will wait for a
	// response to the MessageID before yielding an error. The effective
	// ResponseTimeout is 'ResponseTimeout <= n <= ResponseTimeout +
	// CleanupInterval', so make sure you balance those two intervals to suit
	// your needs.
	ResponseTimeout time.Duration

	// CleanupInterval determines how frequently to check for responses that
	// have timed out. Cleanup blocks the connection and considers all messages
	// currently in-flight.
	CleanupInterval time.Duration
}

func (c ConnConfig) IsZero() bool {
	return c == ConnConfig{}
}

func DefaultConnConfig() ConnConfig {
	return ConnConfig{
		IncomingBuffer:     1024,
		OutgoingBuffer:     1024,
		HeartbeatInterval:  2 * time.Second,
		WriteTimeout:       10 * time.Second,
		ReadTimeout:        10 * time.Second,
		ReadBufferInitial:  2048,
		WriteBufferInitial: 2048,
		CleanupInterval:    5 * time.Second,
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
