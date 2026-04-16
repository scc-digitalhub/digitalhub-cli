package utils

import (
	"log"
	"runtime"
)

const (
	Reset     = "\033[0m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Cyan      = "\033[36m"
	LightGray = "\033[37m"
	DarkGray  = "\033[90m"
)

// LogMode defines the verbosity level
type LogMode int

const (
	ModNormal LogMode = iota
	ModeQuiet
	ModeVerbose
)

type StepLogger struct {
	step int
	mode LogMode
}

// Global logger instance - defaults to quiet mode
var globalLogger = &StepLogger{step: 0, mode: ModeQuiet}

func CreateStepLogger() *StepLogger {
	return &StepLogger{step: 0, mode: ModeQuiet}
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *StepLogger {
	return globalLogger
}

func (s *StepLogger) SetMode(mode LogMode) {
	s.mode = mode
}

func (s *StepLogger) GetMode() LogMode {
	return s.mode
}

func (s *StepLogger) IsVerbose() bool {
	return s.mode == ModeVerbose
}

func (s *StepLogger) Step(msg string) {
	// Skip in quiet mode
	if s.mode == ModeQuiet {
		return
	}
	s.step++
	if supportsColor() {
		log.Printf("%s[%d] %s%s\n", DarkGray, s.step, msg, Reset)
	} else {
		log.Printf("[%d] %s\n", s.step, msg)
	}
}

func (s *StepLogger) Info(msg string) {
	// Skip in quiet mode
	if s.mode == ModeQuiet {
		return
	}
	if supportsColor() {
		log.Println(DarkGray + "ℹ " + msg + Reset)
	} else {
		log.Println("ℹ " + msg)
	}
}

func (s *StepLogger) Debug(msg string) {
	// Skip in quiet mode
	if s.mode == ModeQuiet {
		return
	}
	if supportsColor() {
		log.Println(DarkGray + "DEBUG " + msg + Reset)
	} else {
		log.Println("DEBUG " + msg)
	}
}

func (s *StepLogger) Success(msg string) {
	// Always show
	if supportsColor() {
		log.Println(Green + "✔ " + msg + Reset)
	} else {
		log.Println("✔ " + msg)
	}
}

func (s *StepLogger) Warn(msg string) {
	// Skip in quiet mode
	if s.mode == ModeQuiet {
		return
	}
	if supportsColor() {
		log.Println(Yellow + "⚠ " + msg + Reset)
	} else {
		log.Println("⚠ " + msg)
	}
}

func (s *StepLogger) Error(msg string) {
	// Always show
	if supportsColor() {
		log.Println(Red + "✖ " + msg + Reset)
	} else {
		log.Println("✖ " + msg)
	}
}

func supportsColor() bool {
	return runtime.GOOS != "windows" // simplified
}
