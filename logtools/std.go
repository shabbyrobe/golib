package logtools

import "log"

func Std() Logger {
	return stdLogger
}

var stdLogger = &std{}

type std struct{}

func (s *std) Print(v ...interface{}) {
	log.Print(v...)
}
func (s *std) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
func (s *std) Println(v ...interface{}) {
	log.Println(v...)
}

func (s *std) Fatal(v ...interface{}) {
	log.Fatal(v...)
}
func (s *std) Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}
func (s *std) Fatalln(v ...interface{}) {
	log.Fatalln(v...)
}

func (s *std) Panic(v ...interface{}) {
	log.Panic(v...)
}
func (s *std) Panicf(format string, v ...interface{}) {
	log.Panicf(format, v...)
}
func (s *std) Panicln(v ...interface{}) {
	log.Panicln(v...)
}
