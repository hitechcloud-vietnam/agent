package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hitechcloud-vietnam/agent/pkg/config"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type Logger struct {
	level     Level
	stdout    bool
	logFile   string
	maxSizeMB int64
	maxFiles  int
	mu        sync.Mutex
	file      *os.File
}

func New(cfg *config.Config) (*Logger, io.Closer, error) {
	logger := &Logger{
		level:     parseLevel(cfg.LogLevel),
		stdout:    cfg.LogToStdout,
		logFile:   cfg.LogFile,
		maxSizeMB: cfg.LogMaxSize,
		maxFiles:  cfg.LogMaxFiles,
	}

	if logger.logFile != "" {
		if err := os.MkdirAll(filepath.Dir(logger.logFile), 0755); err != nil {
			return nil, nil, err
		}

		if err := logger.openFile(); err != nil {
			return nil, nil, err
		}
	}

	return logger, logger, nil
}

func (l *Logger) Debugf(format string, args ...any) {
	l.logf(DebugLevel, format, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.logf(InfoLevel, format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.logf(WarnLevel, format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.logf(ErrorLevel, format, args...)
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}

	return nil
}

func (l *Logger) logf(level Level, format string, args ...any) {
	if level < l.level {
		return
	}

	line := fmt.Sprintf("agent: %s [%s] %s\n", time.Now().UTC().Format("2006-01-02 15:04:05"), level.String(), fmt.Sprintf(format, args...))

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.stdout {
		_, _ = os.Stdout.WriteString(line)
	}

	if l.logFile == "" {
		return
	}

	if err := l.rotateIfNeeded(int64(len(line))); err != nil {
		_, _ = os.Stdout.WriteString(fmt.Sprintf("agent: %s [ERROR] failed rotating logs: %v\n", time.Now().UTC().Format("2006-01-02 15:04:05"), err))
		return
	}

	if l.file == nil {
		if err := l.openFile(); err != nil {
			_, _ = os.Stdout.WriteString(fmt.Sprintf("agent: %s [ERROR] failed opening log file: %v\n", time.Now().UTC().Format("2006-01-02 15:04:05"), err))
			return
		}
	}

	_, _ = l.file.WriteString(line)
}

func (l *Logger) openFile() error {
	file, err := os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	l.file = file

	return nil
}

func (l *Logger) rotateIfNeeded(nextWriteSize int64) error {
	if l.file == nil || l.maxSizeMB <= 0 {
		return nil
	}

	info, err := l.file.Stat()
	if err != nil {
		return err
	}

	maxBytes := l.maxSizeMB * 1024 * 1024
	if info.Size()+nextWriteSize < maxBytes {
		return nil
	}

	if err := l.file.Close(); err != nil {
		return err
	}
	l.file = nil

	maxFiles := l.maxFiles
	if maxFiles < 1 {
		maxFiles = 1
	}

	oldest := fmt.Sprintf("%s.%d", l.logFile, maxFiles)
	_ = os.Remove(oldest)

	for i := maxFiles - 1; i >= 1; i-- {
		source := fmt.Sprintf("%s.%d", l.logFile, i)
		target := fmt.Sprintf("%s.%d", l.logFile, i+1)
		if _, err := os.Stat(source); err == nil {
			if err := os.Rename(source, target); err != nil {
				return err
			}
		}
	}

	if _, err := os.Stat(l.logFile); err == nil {
		if err := os.Rename(l.logFile, l.logFile+".1"); err != nil {
			return err
		}
	}

	return l.openFile()
}

func parseLevel(value string) Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return DebugLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "INFO"
	}
}