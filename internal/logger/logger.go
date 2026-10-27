package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	mu    sync.RWMutex
	infoL *log.Logger
	errL  *log.Logger
	warnL *log.Logger
	file  *os.File
)

func Init(logPath string) error {
	mu.Lock()
	defer mu.Unlock()
	if logPath == "" {
		infoL = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
		errL = log.New(os.Stderr, "[ERR]  ", log.LstdFlags)
		warnL = log.New(os.Stdout, "[WARN] ", log.LstdFlags)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	file = f
	mw := io.MultiWriter(os.Stdout, f)
	mew := io.MultiWriter(os.Stderr, f)
	infoL = log.New(mw, "[INFO] ", log.LstdFlags)
	errL = log.New(mew, "[ERR]  ", log.LstdFlags)
	warnL = log.New(mw, "[WARN] ", log.LstdFlags)
	return nil
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		_ = file.Close()
		file = nil
	}
}

func Info(format string, v ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if infoL != nil {
		infoL.Printf(format, v...)
	} else {
		log.Printf("[INFO] "+format, v...)
	}
}

func Error(format string, v ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if errL != nil {
		errL.Printf(format, v...)
	} else {
		log.Printf("[ERR] "+format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if warnL != nil {
		warnL.Printf(format, v...)
	} else {
		log.Printf("[WARN] "+format, v...)
	}
}

func Fatal(format string, v ...interface{}) {
	Error(format, v...)
	os.Exit(1)
}

func Printf(format string, v ...interface{}) { fmt.Printf(format, v...) }
