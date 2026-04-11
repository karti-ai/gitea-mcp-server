// Package log provides structured logging with context support for request tracing.
package log

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// contextKey is a type for context keys to avoid collisions.
type contextKey string

const (
	// correlationIDKey is the context key for correlation ID.
	correlationIDKey contextKey = "correlation_id"
	// operationKey is the context key for operation name.
	operationKey contextKey = "operation"
	// startTimeKey is the context key for operation start time.
	startTimeKey contextKey = "start_time"
)

var (
	// correlationIDGenerator provides thread-safe ID generation.
	correlationIDGenerator = &idGenerator{}
)

// idGenerator generates unique correlation IDs.
type idGenerator struct {
	mu  sync.Mutex
	seq uint64
}

// Generate creates a new unique correlation ID.
func (g *idGenerator) Generate() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.seq++
	return time.Now().Format("20060102-150405") + "-" + string(rune(g.seq))
}

// WithCorrelationID adds a correlation ID to the context for request tracing.
// If no ID is provided, a new one will be generated.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	if id == "" {
		id = correlationIDGenerator.Generate()
	}
	return context.WithValue(ctx, correlationIDKey, id)
}

// WithOperation adds an operation name to the context.
func WithOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, operationKey, operation)
}

// WithStartTime adds operation start time to the context.
func WithStartTime(ctx context.Context) context.Context {
	return context.WithValue(ctx, startTimeKey, time.Now())
}

// GetCorrelationID retrieves the correlation ID from context.
func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}

// GetOperation retrieves the operation name from context.
func GetOperation(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if op, ok := ctx.Value(operationKey).(string); ok {
		return op
	}
	return ""
}

// GetStartTime retrieves the operation start time from context.
func GetStartTime(ctx context.Context) time.Time {
	if ctx == nil {
		return time.Time{}
	}
	if t, ok := ctx.Value(startTimeKey).(time.Time); ok {
		return t
	}
	return time.Time{}
}

// Duration returns the elapsed time since the operation started.
// Returns 0 if no start time is set in context.
func Duration(ctx context.Context) time.Duration {
	start := GetStartTime(ctx)
	if start.IsZero() {
		return 0
	}
	return time.Since(start)
}

// Logger provides request-scoped logging with context fields.
type Logger struct {
	*zap.Logger
	ctx context.Context
}

// WithContext creates a new Logger with context fields.
func WithContext(ctx context.Context) *Logger {
	return &Logger{
		Logger: Default(),
		ctx:    ctx,
	}
}

// WithLogger creates a new Logger with a specific zap logger and context.
func WithLogger(logger *zap.Logger, ctx context.Context) *Logger {
	return &Logger{
		Logger: logger,
		ctx:    ctx,
	}
}

// contextFields returns zap fields from context values.
func (l *Logger) contextFields() []zap.Field {
	if l.ctx == nil {
		return nil
	}

	var fields []zap.Field

	if id := GetCorrelationID(l.ctx); id != "" {
		fields = append(fields, zap.String("correlation_id", id))
	}

	if op := GetOperation(l.ctx); op != "" {
		fields = append(fields, zap.String("operation", op))
	}

	if start := GetStartTime(l.ctx); !start.IsZero() {
		fields = append(fields, zap.Duration("duration", time.Since(start)))
	}

	return fields
}

// Debug logs a message at debug level with context fields.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	allFields := append(l.contextFields(), fields...)
	l.Logger.Debug(msg, allFields...)
}

// Info logs a message at info level with context fields.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	allFields := append(l.contextFields(), fields...)
	l.Logger.Info(msg, allFields...)
}

// Warn logs a message at warn level with context fields.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	allFields := append(l.contextFields(), fields...)
	l.Logger.Warn(msg, allFields...)
}

// Error logs a message at error level with context fields.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	allFields := append(l.contextFields(), fields...)
	l.Logger.Error(msg, allFields...)
}

// Fatal logs a message at fatal level with context fields.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	allFields := append(l.contextFields(), fields...)
	l.Logger.Fatal(msg, allFields...)
}

// Panic logs a message at panic level with context fields.
func (l *Logger) Panic(msg string, fields ...zap.Field) {
	allFields := append(l.contextFields(), fields...)
	l.Logger.Panic(msg, allFields...)
}

// Debugf logs a formatted message at debug level with context fields.
func (l *Logger) Debugf(format string, args ...any) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Infof logs a formatted message at info level with context fields.
func (l *Logger) Infof(format string, args ...any) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted message at warn level with context fields.
func (l *Logger) Warnf(format string, args ...any) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted message at error level with context fields.
func (l *Logger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}

// Fatalf logs a formatted message at fatal level with context fields.
func (l *Logger) Fatalf(format string, args ...any) {
	l.Fatal(fmt.Sprintf(format, args...))
}

// Panicf logs a formatted message at panic level with context fields.
func (l *Logger) Panicf(format string, args ...any) {
	l.Panic(fmt.Sprintf(format, args...))
}

// With creates a child logger with additional fields.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		ctx:    l.ctx,
	}
}

// WithError creates an error log entry with error details.
func (l *Logger) WithError(err error) *ErrorEntry {
	return &ErrorEntry{
		Logger: l,
		err:    err,
	}
}

// ErrorEntry provides structured error logging.
type ErrorEntry struct {
	*Logger
	err error
}

// Log logs the error with additional context.
func (e *ErrorEntry) Log(msg string) {
	if e.err == nil {
		e.Error(msg)
		return
	}

	fields := []zap.Field{
		zap.Error(e.err),
		zap.String("error_type", fmt.Sprintf("%T", e.err)),
	}

	e.Error(msg, fields...)
}

// LogWithStatus logs the error with HTTP status code.
func (e *ErrorEntry) LogWithStatus(msg string, statusCode int) {
	if e.err == nil {
		e.Error(msg, zap.Int("status_code", statusCode))
		return
	}

	fields := []zap.Field{
		zap.Error(e.err),
		zap.String("error_type", fmt.Sprintf("%T", e.err)),
		zap.Int("status_code", statusCode),
	}

	e.Error(msg, fields...)
}

// OperationLogger provides a convenient way to log operation execution.
type OperationLogger struct {
	logger    *Logger
	operation string
	start     time.Time
}

// StartOperation begins logging an operation with timing.
func StartOperation(ctx context.Context, operation string) *OperationLogger {
	logger := WithContext(WithOperation(ctx, operation))
	logger.ctx = WithStartTime(logger.ctx)

	return &OperationLogger{
		logger:    logger,
		operation: operation,
		start:     time.Now(),
	}
}

// Start begins the operation logging with a start message.
func (o *OperationLogger) Start(msg string) {
	o.logger.Info(msg, zap.String("phase", "start"))
}

// Success logs a successful operation completion.
func (o *OperationLogger) Success(msg string) {
	duration := time.Since(o.start)
	o.logger.Info(msg,
		zap.String("phase", "complete"),
		zap.String("status", "success"),
		zap.Duration("duration", duration),
	)
}

// Failure logs a failed operation.
func (o *OperationLogger) Failure(msg string, err error) {
	duration := time.Since(o.start)
	fields := []zap.Field{
		zap.String("phase", "complete"),
		zap.String("status", "failure"),
		zap.Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	o.logger.Error(msg, fields...)
}

// Duration returns the elapsed time since operation started.
func (o *OperationLogger) Duration() time.Duration {
	return time.Since(o.start)
}
