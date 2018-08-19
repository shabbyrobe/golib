package nettools

func HostDefaultPort(host string, defaultPort string) string {
	hl := len(host)
	if hl > 0 && host[hl-1] == ':' {
		host = host[:hl-1]
		hl -= 1
	}

	i := hl
	for i--; i >= 0; i-- {
		if host[i] == ':' {
			break
		}
	}

	if i < 0 {
		if len(defaultPort) > 0 {
			if defaultPort[0] == ':' {
				defaultPort = defaultPort[1:]
			}
			host = host + ":" + defaultPort
		}
	}
	return host
}
