package main

import (
	"log"
	"os"
)

const (
	debugLevel = iota
	infoLevel
	errorLevel
	silentLevel
)

type LevelLogger struct {
	*log.Logger
	level int
}

func NewLogger(level int) *LevelLogger {
	return &LevelLogger{
		Logger: log.New(os.Stderr, "", log.LstdFlags),
		level:  level,
	}
}

func (l *LevelLogger) Debug(v ...interface{}) {
	if l.level <= debugLevel {
		log.Print(v)
	}
}

func (l *LevelLogger) Debugln(v ...interface{}) {
	if l.level <= debugLevel {
		log.Println(v...)
	}
}

func (l *LevelLogger) Debugf(format string, v ...interface{}) {
	if l.level <= debugLevel {
		log.Printf(format, v...)
	}
}

func (l *LevelLogger) Info(v ...interface{}) {
	if l.level <= infoLevel {
		log.Print(v)
	}
}

func (l *LevelLogger) Infoln(v ...interface{}) {
	if l.level <= infoLevel {
		log.Println(v...)
	}
}

func (l *LevelLogger) Infof(format string, v ...interface{}) {
	if l.level <= infoLevel {
		log.Printf(format, v...)
	}
}

func (l *LevelLogger) Error(v ...interface{}) {
	if l.level <= errorLevel {
		log.Print(v)
	}
}

func (l *LevelLogger) Errorln(v ...interface{}) {
	if l.level <= errorLevel {
		log.Println(v...)
	}
}

func (l *LevelLogger) Errorf(format string, v ...interface{}) {
	if l.level <= errorLevel {
		log.Printf(format, v...)
	}
}

func (l *LevelLogger) Fatal(v ...interface{}) {
	if l.level <= errorLevel {
		log.Print(v)
	}
	os.Exit(1)
}

func (l *LevelLogger) Fatalln(v ...interface{}) {
	if l.level <= errorLevel {
		log.Println(v...)
	}
	os.Exit(1)
}

func (l *LevelLogger) Fatalf(format string, v ...interface{}) {
	if l.level <= errorLevel {
		log.Printf(format, v...)
	}
	os.Exit(1)
}