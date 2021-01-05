package logger

import (
	"io"
	"log"
	"os"
)

func NewLogger(suiteName string) (*log.Logger, *os.File) {
	logpath := "./logs"
	if _, err := os.Stat("./storage/logs"); os.IsNotExist(err) {
		os.Mkdir(logpath, os.ModePerm|os.ModeDir)
	}

	f, err := os.OpenFile("./storage/logs/"+suiteName+".log", os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalf("error while opening file: %v", err)
	}
	mw := io.MultiWriter(os.Stdout, f)
	l := log.New(mw, suiteName, log.LstdFlags)
	l.SetOutput(mw)
	return l, f
}
