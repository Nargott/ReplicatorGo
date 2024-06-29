package main

import (
	"log"
	"os"
)

type RLog struct {
	IsDebugEnabled bool
}

var Rlog = new(RLog)

func (l *RLog) SetDebugEnabled(enabled bool) {
	l.IsDebugEnabled = enabled
}

func (l *RLog) Debug(v ...any) {
	if !l.IsDebugEnabled {
		return
	}
	log.SetOutput(os.Stdout)
	log.Println(v)
}

func (l *RLog) Debugf(format string, v ...any) {
	if !l.IsDebugEnabled {
		return
	}
	log.SetOutput(os.Stdout)
	log.Printf(format, v...)
}

func (l *RLog) Info(v ...any) {
	log.SetOutput(os.Stdout)
	log.Println(v)
}

func (l *RLog) Infof(format string, v ...any) {
	log.SetOutput(os.Stdout)
	log.Printf(format, v...)
}

func (l *RLog) Error(v ...any) {
	log.SetOutput(os.Stderr)
	log.Println(v)
}

func (l *RLog) Errorf(format string, v ...any) {
	log.SetOutput(os.Stderr)
	log.Printf(format, v...)
}

func (l *RLog) Fatal(v ...any) {
	log.SetOutput(os.Stderr)
	log.Fatal(v)
}

func (l *RLog) Fatalf(format string, v ...any) {
	log.SetOutput(os.Stderr)
	log.Fatalf(format, v...)
}
