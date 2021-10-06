package impltests

import "log"

var noplog = &NopLogger{
	log.New(NullWriter(1), "", log.LstdFlags),
}

// NullWriter implements the io.Write interface but doesn't do anything.
type NullWriter int

// Write implements the io.Write interface but is a noop.
func (NullWriter) Write([]byte) (int, error) { return 0, nil }

// NopLogger is a noop logger for passing to grpclog to minimize spew.
type NopLogger struct {
	*log.Logger
}

// Fatal is a noop
func (l *NopLogger) Fatal(args ...interface{}) {}

// Fatalf is a noop
func (l *NopLogger) Fatalf(format string, args ...interface{}) {}

// Fatalln is a noop
func (l *NopLogger) Fatalln(args ...interface{}) {}

// Print is a noop
func (l *NopLogger) Print(args ...interface{}) {}

// Printf is a noop
func (l *NopLogger) Printf(format string, args ...interface{}) {}

// Println is a noop
func (l *NopLogger) Println(v ...interface{}) {}
