package logger

import (
	"log/slog"
	"time"
)

// LogLevel 定义日志级别对应的表情符号和文本
type LogLevel string

const (
	// 操作状态相关
	LevelStart   LogLevel = "🔍"  // 操作开始
	LevelSuccess LogLevel = "✅"  // 操作成功
	LevelError   LogLevel = "❌"  // 操作失败
	LevelWarn    LogLevel = "⚠️" // 警告

	// 数据流相关
	LevelSend    LogLevel = "📤" // 发送/开始
	LevelReceive LogLevel = "📥" // 接收/数据块

	// 功能相关
	LevelExplain   LogLevel = "📚" // 解释/释义
	LevelTranslate LogLevel = "📝" // 翻译
)

// StreamLogger 流式操作日志记录器
type StreamLogger struct {
	operationName string
	service       string
	startTime     time.Time
	chunkCount    int
	totalLength   int
}

// NewStreamLogger 创建新的流式日志记录器
func NewStreamLogger(operationName, service string) *StreamLogger {
	return &StreamLogger{
		operationName: operationName,
		service:       service,
		startTime:     time.Now(),
		chunkCount:    0,
		totalLength:   0,
	}
}

// LogStart 记录操作开始
func (sl *StreamLogger) LogStart(level LogLevel, message string) {
	slog.Info(string(level)+" "+message,
		slog.String("service", sl.service))
}

// LogChunk 记录接收到的数据块
func (sl *StreamLogger) LogChunk(chunkSize int) {
	sl.chunkCount++
	sl.totalLength += chunkSize

	slog.Debug(string(LevelReceive)+" 接收数据块",
		slog.Int("chunk_count", sl.chunkCount),
		slog.Int("chunk_size", chunkSize),
		slog.Int("total_length", sl.totalLength),
		slog.String("service", sl.service))
}

// LogSuccess 记录操作成功完成
func (sl *StreamLogger) LogSuccess(message string) {
	elapsed := time.Since(sl.startTime)

	slog.Info(string(LevelSuccess)+" "+message,
		slog.String("service", sl.service),
		slog.Int("total_chunks", sl.chunkCount),
		slog.Int("total_length", sl.totalLength),
		slog.String("duration", elapsed.String()))
}

// LogError 记录操作失败
func (sl *StreamLogger) LogError(message string, err error) {
	elapsed := time.Since(sl.startTime)

	slog.Error(string(LevelError)+" "+message,
		slog.String("service", sl.service),
		slog.String("error", err.Error()),
		slog.Int("chunks_received", sl.chunkCount),
		slog.String("duration", elapsed.String()))
}

// ==================== 便捷函数 ====================

// LogTranslateStart 记录翻译开始
func LogTranslateStart(service, sourceLang, targetLang string, inputLen int) {
	slog.Info(string(LevelSend)+" 开始流式翻译",
		slog.String("service", service),
		slog.String("source_lang", sourceLang),
		slog.String("target_lang", targetLang),
		slog.Int("input_length", inputLen))
}

// LogTranslateSuccess 记录翻译成功
func LogTranslateSuccess(service string, inputLen, outputLen, chunks int, duration time.Duration) {
	slog.Info(string(LevelSuccess)+" 流式翻译完成",
		slog.String("service", service),
		slog.Int("input_length", inputLen),
		slog.Int("output_length", outputLen),
		slog.Int("total_chunks", chunks),
		slog.String("duration", duration.String()))
}

// LogTranslateError 记录翻译失败
func LogTranslateError(service string, err error, chunks int, duration time.Duration) {
	slog.Error(string(LevelError)+" 流式翻译失败",
		slog.String("service", service),
		slog.String("error", err.Error()),
		slog.Int("chunks_received", chunks),
		slog.String("duration", duration.String()))
}

// LogExplainStart 记录解释开始
func LogExplainStart(service, templateID string, inputLen int) {
	slog.Info(string(LevelExplain)+" 开始流式解释",
		slog.String("service", service),
		slog.String("template_id", templateID),
		slog.Int("input_length", inputLen))
}

// LogExplainSuccess 记录解释成功
func LogExplainSuccess(service, templateID string, inputLen, outputLen, chunks int, duration time.Duration) {
	slog.Info(string(LevelSuccess)+" 流式解释完成",
		slog.String("service", service),
		slog.String("template_id", templateID),
		slog.Int("input_length", inputLen),
		slog.Int("output_length", outputLen),
		slog.Int("total_chunks", chunks),
		slog.String("duration", duration.String()))
}

// LogExplainError 记录解释失败
func LogExplainError(service, templateID string, err error, chunks int, duration time.Duration) {
	slog.Error(string(LevelError)+" 流式解释失败",
		slog.String("service", service),
		slog.String("template_id", templateID),
		slog.String("error", err.Error()),
		slog.Int("chunks_received", chunks),
		slog.String("duration", duration.String()))
}

// LogNormalTranslateStart 记录普通翻译开始
func LogNormalTranslateStart(service, sourceLang, targetLang string) {
	slog.Info(string(LevelTranslate)+" 开始普通翻译",
		slog.String("service", service),
		slog.String("source_lang", sourceLang),
		slog.String("target_lang", targetLang))
}

// LogNormalTranslateSuccess 记录普通翻译成功
func LogNormalTranslateSuccess(service string, resultCount int) {
	slog.Info(string(LevelSuccess)+" 普通翻译完成",
		slog.String("service", service),
		slog.Int("results_count", resultCount))
}

// LogNormalTranslateError 记录普通翻译失败
func LogNormalTranslateError(service string, err error) {
	slog.Error(string(LevelError)+" PostQuery 失败",
		slog.String("service", service),
		slog.String("error", err.Error()))
}

// LogStreamNotSupported 记录不支持流式的情况
func LogStreamNotSupported(service string) {
	slog.Info(string(LevelWarn)+" 不支持流式翻译，使用普通模式",
		slog.String("service", service))
}

// LogStreamExplainNotSupported 记录不支持流式解释
func LogStreamExplainNotSupported(service, templateID string) {
	slog.Error(string(LevelError)+" 不支持流式解释",
		slog.String("service", service),
		slog.String("template_id", templateID))
}

// LogChunkReceived 记录接收到的数据块（Debug级别）
func LogChunkReceived(chunkCount, chunkSize, totalLength int) {
	slog.Debug(string(LevelReceive)+" 接收流式数据块",
		slog.Int("chunk_count", chunkCount),
		slog.Int("chunk_size", chunkSize),
		slog.Int("total_length", totalLength))
}

// LogOperationStart 通用操作开始日志
func LogOperationStart(level LogLevel, message string, service string) {
	slog.Info(string(level)+" "+message,
		slog.String("service", service))
}

// LogOperationSuccess 通用操作成功日志
func LogOperationSuccess(message string, service string) {
	slog.Info(string(LevelSuccess)+" "+message,
		slog.String("service", service))
}

// LogOperationError 通用操作失败日志
func LogOperationError(message string, service string, err error) {
	slog.Error(string(LevelError)+" "+message,
		slog.String("service", service),
		slog.String("error", err.Error()))
}
