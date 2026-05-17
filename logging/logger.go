package logging

import "sync"

// Level represents the severity level of a log message.
type Level int

const (
	// DEBUG is the lowest severity level for detailed debugging information.
	DEBUG Level = iota
	// INFO is for general informational messages.
	INFO
	// NOTICE is for normal but significant conditions.
	NOTICE
	// WARNING is for warning conditions that should be addressed.
	WARNING
	// ERROR is for error conditions that should be investigated.
	ERROR
	// CRITICAL is for critical conditions requiring immediate attention.
	CRITICAL
	// ALERT is for conditions requiring immediate action.
	ALERT
	// EMERGENCY is the highest severity level for system-wide failures.
	EMERGENCY
)

// String returns the string representation of a log level.
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case NOTICE:
		return "NOTICE"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case CRITICAL:
		return "CRITICAL"
	case ALERT:
		return "ALERT"
	case EMERGENCY:
		return "EMERGENCY"
	default:
		return "UNKNOWN"
	}
}

// Logger defines the interface for logging messages at various severity levels.
type Logger interface {
	// Emergency logs a message at EMERGENCY level.
	Emergency(msg string, ctx ...map[string]any)
	// Alert logs a message at ALERT level.
	Alert(msg string, ctx ...map[string]any)
	// Critical logs a message at CRITICAL level.
	Critical(msg string, ctx ...map[string]any)
	// Error logs a message at ERROR level.
	Error(msg string, ctx ...map[string]any)
	// Warning logs a message at WARNING level.
	Warning(msg string, ctx ...map[string]any)
	// Notice logs a message at NOTICE level.
	Notice(msg string, ctx ...map[string]any)
	// Info logs a message at INFO level.
	Info(msg string, ctx ...map[string]any)
	// Debug logs a message at DEBUG level.
	Debug(msg string, ctx ...map[string]any)
	// Log logs a message at the specified level.
	Log(level Level, msg string, ctx ...map[string]any)
}

// Manager manages logging channels and provides access to loggers.
type Manager struct {
	mu           sync.RWMutex
	channels     map[string]Logger
	defaultName  string
	config       map[string]any
	contextData  map[string]any
}

// NewManager creates a new logging manager with the provided configuration.
func NewManager(config map[string]any) *Manager {
	return &Manager{
		channels:    make(map[string]Logger),
		config:      config,
		defaultName: "default",
		contextData: make(map[string]any),
	}
}

// Channel returns the logger for the specified channel name.
// If the channel doesn't exist, it creates a default stderr channel.
func (m *Manager) Channel(name string) Logger {
	m.mu.RLock()
	logger, exists := m.channels[name]
	m.mu.RUnlock()

	if exists {
		return logger
	}

	// Create default stderr channel if not found
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if logger, exists := m.channels[name]; exists {
		return logger
	}

	logger = NewStderrChannel(name, DEBUG, NewLineFormatter())
	m.channels[name] = logger
	return logger
}

// Stack creates a logger that writes to multiple channels.
func (m *Manager) Stack(channels []string) Logger {
	loggers := make([]Logger, 0, len(channels))
	for _, name := range channels {
		loggers = append(loggers, m.Channel(name))
	}
	return NewStackChannel(loggers)
}

// WithContext returns a logger that includes the provided context in all log entries.
func (m *Manager) WithContext(ctx map[string]any) Logger {
	return &contextLogger{
		logger:  m.Channel(m.defaultName),
		context: ctx,
	}
}

// RegisterChannel adds a named channel to the manager.
func (m *Manager) RegisterChannel(name string, logger Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[name] = logger
}

// Emergency logs a message at EMERGENCY level using the default channel.
func (m *Manager) Emergency(msg string, ctx ...map[string]any) {
	m.Channel(m.defaultName).Emergency(msg, ctx...)
}

// Alert logs a message at ALERT level using the default channel.
func (m *Manager) Alert(msg string, ctx ...map[string]any) {
	m.Channel(m.defaultName).Alert(msg, ctx...)
}

// Critical logs a message at CRITICAL level using the default channel.
func (m *Manager) Critical(msg string, ctx ...map[string]any) {
	m.Channel(m.defaultName).Critical(msg, ctx...)
}

// Error logs a message at ERROR level using the default channel.
func (m *Manager) Error(msg string, ctx ...map[string]any) {
	m.Channel(m.defaultName).Error(msg, ctx...)
}

// Warning logs a message at WARNING level using the default channel.
func (m *Manager) Warning(msg string, ctx ...map[string]any) {
	m.Channel(m.defaultName).Warning(msg, ctx...)
}

// Notice logs a message at NOTICE level using the default channel.
func (m *Manager) Notice(msg string, ctx ...map[string]any) {
	m.Channel(m.defaultName).Notice(msg, ctx...)
}

// Info logs a message at INFO level using the default channel.
func (m *Manager) Info(msg string, ctx ...map[string]any) {
	m.Channel(m.defaultName).Info(msg, ctx...)
}

// Debug logs a message at DEBUG level using the default channel.
func (m *Manager) Debug(msg string, ctx ...map[string]any) {
	m.Channel(m.defaultName).Debug(msg, ctx...)
}

// Log logs a message at the specified level using the default channel.
func (m *Manager) Log(level Level, msg string, ctx ...map[string]any) {
	m.Channel(m.defaultName).Log(level, msg, ctx...)
}

// contextLogger wraps a Logger and adds persistent context data to all log entries.
type contextLogger struct {
	logger  Logger
	context map[string]any
}

func (c *contextLogger) Emergency(msg string, ctx ...map[string]any) {
	c.Log(EMERGENCY, msg, ctx...)
}

func (c *contextLogger) Alert(msg string, ctx ...map[string]any) {
	c.Log(ALERT, msg, ctx...)
}

func (c *contextLogger) Critical(msg string, ctx ...map[string]any) {
	c.Log(CRITICAL, msg, ctx...)
}

func (c *contextLogger) Error(msg string, ctx ...map[string]any) {
	c.Log(ERROR, msg, ctx...)
}

func (c *contextLogger) Warning(msg string, ctx ...map[string]any) {
	c.Log(WARNING, msg, ctx...)
}

func (c *contextLogger) Notice(msg string, ctx ...map[string]any) {
	c.Log(NOTICE, msg, ctx...)
}

func (c *contextLogger) Info(msg string, ctx ...map[string]any) {
	c.Log(INFO, msg, ctx...)
}

func (c *contextLogger) Debug(msg string, ctx ...map[string]any) {
	c.Log(DEBUG, msg, ctx...)
}

func (c *contextLogger) Log(level Level, msg string, ctx ...map[string]any) {
	merged := make(map[string]any)
	for k, v := range c.context {
		merged[k] = v
	}
	if len(ctx) > 0 {
		for k, v := range ctx[0] {
			merged[k] = v
		}
	}
	c.logger.Log(level, msg, merged)
}

// Log is the default package-level logger instance.
var Log Logger = NewManager(nil)
