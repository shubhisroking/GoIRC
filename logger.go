package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger handles application logging with size limits
type Logger struct {
	config      *Config
	logFile     *os.File
	debugFile   *os.File
	logger      *log.Logger
	debugLogger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(config *Config) (*Logger, error) {
	if !config.Logging.Enabled {
		return &Logger{config: config}, nil
	}

	// Ensure log directory exists
	if err := os.MkdirAll(config.Logging.LogPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logger := &Logger{config: config}

	// Setup main log file
	logPath := config.GetLogFilePath()
	if err := logger.rotateLogIfNeeded(logPath); err != nil {
		return nil, fmt.Errorf("failed to setup log rotation: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	logger.logFile = logFile
	logger.logger = log.New(logFile, "", log.LstdFlags)

	// Setup debug log file if debug mode is enabled
	if config.Logging.DebugMode {
		debugPath := config.GetDebugLogFilePath()
		if err := logger.rotateLogIfNeeded(debugPath); err != nil {
			return nil, fmt.Errorf("failed to setup debug log rotation: %w", err)
		}

		debugFile, err := os.OpenFile(debugPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open debug log file: %w", err)
		}
		logger.debugFile = debugFile
		logger.debugLogger = log.New(debugFile, "DEBUG: ", log.LstdFlags|log.Lshortfile)
	}

	return logger, nil
}

// rotateLogIfNeeded checks if a log file needs rotation based on size
func (l *Logger) rotateLogIfNeeded(logPath string) error {
	info, err := os.Stat(logPath)
	if os.IsNotExist(err) {
		return nil // File doesn't exist yet, no rotation needed
	}
	if err != nil {
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	maxSizeBytes := int64(l.config.Logging.MaxSizeKB * 1024)
	if info.Size() >= maxSizeBytes {
		return l.rotateLog(logPath)
	}

	return nil
}

// rotateLog rotates a log file by renaming it with a timestamp
func (l *Logger) rotateLog(logPath string) error {
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(logPath)
	base := logPath[:len(logPath)-len(ext)]
	rotatedPath := fmt.Sprintf("%s.%s%s", base, timestamp, ext)

	if err := os.Rename(logPath, rotatedPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	return nil
}

// Log logs a message to the main log file
func (l *Logger) Log(format string, v ...interface{}) {
	if l.config == nil || !l.config.Logging.Enabled || l.logger == nil {
		return
	}

	message := fmt.Sprintf(format, v...)
	l.logger.Println(message)

	// Check if rotation is needed after writing
	logPath := l.config.GetLogFilePath()
	if err := l.rotateLogIfNeeded(logPath); err != nil {
		// If rotation fails, log to stderr but continue
		fmt.Fprintf(os.Stderr, "Failed to rotate log: %v\n", err)
	}
}

// Debug logs a debug message to the debug log file
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.config == nil || !l.config.Logging.Enabled || !l.config.Logging.DebugMode || l.debugLogger == nil {
		return
	}

	message := fmt.Sprintf(format, v...)
	l.debugLogger.Println(message)

	// Check if rotation is needed after writing
	debugPath := l.config.GetDebugLogFilePath()
	if err := l.rotateLogIfNeeded(debugPath); err != nil {
		// If rotation fails, log to stderr but continue
		fmt.Fprintf(os.Stderr, "Failed to rotate debug log: %v\n", err)
	}
}

// LogIRCMessage logs an IRC message with proper formatting
func (l *Logger) LogIRCMessage(channel, user, message string) {
	if channel == "" {
		l.Log("<%s> %s", user, message)
	} else {
		l.Log("[%s] <%s> %s", channel, user, message)
	}
}

// LogIRCEvent logs an IRC event (joins, parts, etc.)
func (l *Logger) LogIRCEvent(event string, args ...interface{}) {
	l.Log("* "+event, args...)
}

// LogError logs an error message
func (l *Logger) LogError(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	l.Log("ERROR: %s", message)
	l.Debug("ERROR: %s", message)
}

// Close closes all log files
func (l *Logger) Close() error {
	var lastErr error

	if l.logFile != nil {
		if err := l.logFile.Close(); err != nil {
			lastErr = err
		}
	}

	if l.debugFile != nil {
		if err := l.debugFile.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// GetLogWriter returns an io.Writer for the main log
func (l *Logger) GetLogWriter() io.Writer {
	if l.logFile != nil {
		return l.logFile
	}
	return os.Stdout
}

// GetDebugWriter returns an io.Writer for debug logs
func (l *Logger) GetDebugWriter() io.Writer {
	if l.debugFile != nil {
		return l.debugFile
	}
	return os.Stderr
}

// LogStartup logs application startup information
func (l *Logger) LogStartup(version string) {
	l.Log("=== GoIRC Client Started ===")
	if version != "" {
		l.Log("Version: %s", version)
	}
	l.Log("Config loaded from: %s", l.config.FilePath)
	l.Log("Logs directory: %s", l.config.Logging.LogPath)
	l.Log("Max log size: %d KB", l.config.Logging.MaxSizeKB)
	l.Debug("Debug logging enabled")
}

// LogShutdown logs application shutdown information
func (l *Logger) LogShutdown() {
	l.Log("=== GoIRC Client Shutdown ===")
}
