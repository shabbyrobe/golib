package httpagent

type Logger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

type nilLogger struct{}

func (n *nilLogger) Println(v ...interface{})               {}
func (n *nilLogger) Printf(format string, v ...interface{}) {}
