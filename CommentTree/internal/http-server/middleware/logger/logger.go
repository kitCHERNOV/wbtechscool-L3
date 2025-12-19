package logger

import "os"

type localLogger struct {
	file *os.File
}

func MustNewLocalLogger(filePath string) *localLogger {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return &localLogger{file: f}
}

func (l *localLogger) Close() {
	l.file.Close()
}

func (l *localLogger) Write(p []byte) (n int, err error) {
	return l.file.Write(p)
}
