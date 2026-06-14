package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func Init(logDir string) error {
	if logDir == "" {
		logDir = "./logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	infoWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "info.log"),
		MaxSize:    100,
		MaxBackups: 30,
		MaxAge:     30,
		Compress:   true,
	}

	errorWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "error.log"),
		MaxSize:    100,
		MaxBackups: 30,
		MaxAge:     30,
		Compress:   true,
	}

	infoMulti := io.MultiWriter(os.Stdout, infoWriter)
	errorMulti := io.MultiWriter(os.Stderr, errorWriter)

	InfoLogger = log.New(infoMulti, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(errorMulti, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

func Info(format string, v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Printf(format, v...)
	} else {
		log.Printf("INFO: "+format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Printf(format, v...)
	} else {
		log.Printf("ERROR: "+format, v...)
	}
}

func Now() time.Time {
	return time.Now().Local()
}
