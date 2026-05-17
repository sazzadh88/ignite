package logging

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// baseChannel provides common functionality for channel implementations.
type baseChannel struct {
	name      string
	level     Level
	formatter Formatter
}

func (b *baseChannel) shouldLog(level Level) bool {
	return level >= b.level
}

func (b *baseChannel) createEntry(level Level, msg string, ctx ...map[string]any) *Entry {
	entry := &Entry{
		Level:     level,
		Message:   msg,
		Timestamp: time.Now(),
		Channel:   b.name,
		Context:   make(map[string]any),
	}

	if len(ctx) > 0 && ctx[0] != nil {
		entry.Context = ctx[0]
	}

	return entry
}

// FileChannel writes log entries to a file.
type FileChannel struct {
	baseChannel
	writer *Writer
}

// NewFileChannel creates a new file channel that writes to the specified path.
func NewFileChannel(name, path string, level Level, formatter Formatter) (*FileChannel, error) {
	writer, err := NewWriter(path)
	if err != nil {
		return nil, err
	}

	return &FileChannel{
		baseChannel: baseChannel{
			name:      name,
			level:     level,
			formatter: formatter,
		},
		writer: writer,
	}, nil
}

func (f *FileChannel) Emergency(msg string, ctx ...map[string]any) {
	f.Log(EMERGENCY, msg, ctx...)
}

func (f *FileChannel) Alert(msg string, ctx ...map[string]any) {
	f.Log(ALERT, msg, ctx...)
}

func (f *FileChannel) Critical(msg string, ctx ...map[string]any) {
	f.Log(CRITICAL, msg, ctx...)
}

func (f *FileChannel) Error(msg string, ctx ...map[string]any) {
	f.Log(ERROR, msg, ctx...)
}

func (f *FileChannel) Warning(msg string, ctx ...map[string]any) {
	f.Log(WARNING, msg, ctx...)
}

func (f *FileChannel) Notice(msg string, ctx ...map[string]any) {
	f.Log(NOTICE, msg, ctx...)
}

func (f *FileChannel) Info(msg string, ctx ...map[string]any) {
	f.Log(INFO, msg, ctx...)
}

func (f *FileChannel) Debug(msg string, ctx ...map[string]any) {
	f.Log(DEBUG, msg, ctx...)
}

func (f *FileChannel) Log(level Level, msg string, ctx ...map[string]any) {
	if !f.shouldLog(level) {
		return
	}

	entry := f.createEntry(level, msg, ctx...)
	formatted := f.formatter.Format(entry)
	f.writer.Write([]byte(formatted))
}

// Close closes the file channel's writer.
func (f *FileChannel) Close() error {
	return f.writer.Close()
}

// DailyChannel writes log entries to a file with daily rotation.
// The filename includes the current date (e.g., app-2024-01-01.log).
type DailyChannel struct {
	baseChannel
	mu           sync.Mutex
	basePath     string
	currentDate  string
	currentWriter *Writer
}

// NewDailyChannel creates a new daily rotating file channel.
// The basePath should not include the date suffix (e.g., "logs/app.log").
func NewDailyChannel(name, basePath string, level Level, formatter Formatter) *DailyChannel {
	return &DailyChannel{
		baseChannel: baseChannel{
			name:      name,
			level:     level,
			formatter: formatter,
		},
		basePath: basePath,
	}
}

func (d *DailyChannel) getWriter() (*Writer, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	today := time.Now().Format("2006-01-02")

	if d.currentWriter != nil && d.currentDate == today {
		return d.currentWriter, nil
	}

	// Close old writer if exists
	if d.currentWriter != nil {
		d.currentWriter.Close()
	}

	// Create new writer with date in filename
	path := fmt.Sprintf("%s-%s", d.basePath, today)
	writer, err := NewWriter(path)
	if err != nil {
		return nil, err
	}

	d.currentWriter = writer
	d.currentDate = today

	return writer, nil
}

func (d *DailyChannel) Emergency(msg string, ctx ...map[string]any) {
	d.Log(EMERGENCY, msg, ctx...)
}

func (d *DailyChannel) Alert(msg string, ctx ...map[string]any) {
	d.Log(ALERT, msg, ctx...)
}

func (d *DailyChannel) Critical(msg string, ctx ...map[string]any) {
	d.Log(CRITICAL, msg, ctx...)
}

func (d *DailyChannel) Error(msg string, ctx ...map[string]any) {
	d.Log(ERROR, msg, ctx...)
}

func (d *DailyChannel) Warning(msg string, ctx ...map[string]any) {
	d.Log(WARNING, msg, ctx...)
}

func (d *DailyChannel) Notice(msg string, ctx ...map[string]any) {
	d.Log(NOTICE, msg, ctx...)
}

func (d *DailyChannel) Info(msg string, ctx ...map[string]any) {
	d.Log(INFO, msg, ctx...)
}

func (d *DailyChannel) Debug(msg string, ctx ...map[string]any) {
	d.Log(DEBUG, msg, ctx...)
}

func (d *DailyChannel) Log(level Level, msg string, ctx ...map[string]any) {
	if !d.shouldLog(level) {
		return
	}

	writer, err := d.getWriter()
	if err != nil {
		// Fallback to stderr if file writing fails
		fmt.Fprintf(os.Stderr, "Failed to get log writer: %v\n", err)
		return
	}

	entry := d.createEntry(level, msg, ctx...)
	formatted := d.formatter.Format(entry)
	writer.Write([]byte(formatted))
}

// Close closes the daily channel's current writer.
func (d *DailyChannel) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.currentWriter != nil {
		return d.currentWriter.Close()
	}
	return nil
}

// StderrChannel writes log entries to stderr.
type StderrChannel struct {
	baseChannel
	mu sync.Mutex
}

// NewStderrChannel creates a new stderr channel.
func NewStderrChannel(name string, level Level, formatter Formatter) *StderrChannel {
	return &StderrChannel{
		baseChannel: baseChannel{
			name:      name,
			level:     level,
			formatter: formatter,
		},
	}
}

func (s *StderrChannel) Emergency(msg string, ctx ...map[string]any) {
	s.Log(EMERGENCY, msg, ctx...)
}

func (s *StderrChannel) Alert(msg string, ctx ...map[string]any) {
	s.Log(ALERT, msg, ctx...)
}

func (s *StderrChannel) Critical(msg string, ctx ...map[string]any) {
	s.Log(CRITICAL, msg, ctx...)
}

func (s *StderrChannel) Error(msg string, ctx ...map[string]any) {
	s.Log(ERROR, msg, ctx...)
}

func (s *StderrChannel) Warning(msg string, ctx ...map[string]any) {
	s.Log(WARNING, msg, ctx...)
}

func (s *StderrChannel) Notice(msg string, ctx ...map[string]any) {
	s.Log(NOTICE, msg, ctx...)
}

func (s *StderrChannel) Info(msg string, ctx ...map[string]any) {
	s.Log(INFO, msg, ctx...)
}

func (s *StderrChannel) Debug(msg string, ctx ...map[string]any) {
	s.Log(DEBUG, msg, ctx...)
}

func (s *StderrChannel) Log(level Level, msg string, ctx ...map[string]any) {
	if !s.shouldLog(level) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry := s.createEntry(level, msg, ctx...)
	formatted := s.formatter.Format(entry)
	os.Stderr.Write([]byte(formatted))
}

// NullChannel discards all log entries.
type NullChannel struct {
	baseChannel
}

// NewNullChannel creates a new null channel that discards all logs.
func NewNullChannel(name string) *NullChannel {
	return &NullChannel{
		baseChannel: baseChannel{
			name:      name,
			level:     EMERGENCY + 1, // Never log anything
			formatter: NewLineFormatter(),
		},
	}
}

func (n *NullChannel) Emergency(msg string, ctx ...map[string]any) {}
func (n *NullChannel) Alert(msg string, ctx ...map[string]any)     {}
func (n *NullChannel) Critical(msg string, ctx ...map[string]any)  {}
func (n *NullChannel) Error(msg string, ctx ...map[string]any)     {}
func (n *NullChannel) Warning(msg string, ctx ...map[string]any)   {}
func (n *NullChannel) Notice(msg string, ctx ...map[string]any)    {}
func (n *NullChannel) Info(msg string, ctx ...map[string]any)      {}
func (n *NullChannel) Debug(msg string, ctx ...map[string]any)     {}
func (n *NullChannel) Log(level Level, msg string, ctx ...map[string]any) {}

// StackChannel writes log entries to multiple channels.
type StackChannel struct {
	loggers []Logger
}

// NewStackChannel creates a new stack channel that writes to multiple loggers.
func NewStackChannel(loggers []Logger) *StackChannel {
	return &StackChannel{
		loggers: loggers,
	}
}

func (s *StackChannel) Emergency(msg string, ctx ...map[string]any) {
	for _, logger := range s.loggers {
		logger.Emergency(msg, ctx...)
	}
}

func (s *StackChannel) Alert(msg string, ctx ...map[string]any) {
	for _, logger := range s.loggers {
		logger.Alert(msg, ctx...)
	}
}

func (s *StackChannel) Critical(msg string, ctx ...map[string]any) {
	for _, logger := range s.loggers {
		logger.Critical(msg, ctx...)
	}
}

func (s *StackChannel) Error(msg string, ctx ...map[string]any) {
	for _, logger := range s.loggers {
		logger.Error(msg, ctx...)
	}
}

func (s *StackChannel) Warning(msg string, ctx ...map[string]any) {
	for _, logger := range s.loggers {
		logger.Warning(msg, ctx...)
	}
}

func (s *StackChannel) Notice(msg string, ctx ...map[string]any) {
	for _, logger := range s.loggers {
		logger.Notice(msg, ctx...)
	}
}

func (s *StackChannel) Info(msg string, ctx ...map[string]any) {
	for _, logger := range s.loggers {
		logger.Info(msg, ctx...)
	}
}

func (s *StackChannel) Debug(msg string, ctx ...map[string]any) {
	for _, logger := range s.loggers {
		logger.Debug(msg, ctx...)
	}
}

func (s *StackChannel) Log(level Level, msg string, ctx ...map[string]any) {
	for _, logger := range s.loggers {
		logger.Log(level, msg, ctx...)
	}
}
