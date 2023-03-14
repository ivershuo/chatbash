package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var mutex sync.Mutex

func setLog(logPath string, discard bool) {
	if discard {
		log.SetOutput(io.Discard)
		return
	}
	if logPath == "" {
		homePath, err := os.UserHomeDir()
		if err != nil {
			logPath = "/tmp/chatbash"
		}
		if runtime.GOOS == "darwin" {
			logPath = filepath.Join(homePath, "Library/Logs/chatbash")
		} else if runtime.GOOS == "linux" {
			logPath = filepath.Join(homePath, ".log/chatbash")
		}
	}
	if err := os.MkdirAll(logPath, 0755); err != nil {
		log.Fatal("Failed to create log dir: " + err.Error())
	}
	logFile, err := os.OpenFile(filepath.Join(logPath, time.Now().Format("20060102")+".log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open log file: " + err.Error())
	}
	log.SetOutput(logFile)
}

func writeLog(level string, format string, v ...any) {
	mutex.Lock()
	defer mutex.Unlock()
	msg := fmt.Sprintf(format, v...)
	log.Printf("[%s] %s", level, msg)
}

func InfoLog(format string, v ...any) {
	go writeLog("info", format, v...)
}

func ErrLog(format string, v ...any) {
	go writeLog("error", format, v...)
}
