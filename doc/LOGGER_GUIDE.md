# 统一日志组件使用指南

## 概述

创建了 `logger` 包提供统一的日志打印接口，确保整个项目的日志风格一致，并简化日志的调用方式。

## 组件结构

### 1. 日志级别常量

```go
const (
	LevelStart     LogLevel = "🔍" // 操作开始
	LevelSuccess   LogLevel = "✅" // 操作成功
	LevelError     LogLevel = "❌" // 操作失败
	LevelWarn      LogLevel = "⚠️" // 警告
	LevelSend      LogLevel = "📤" // 发送/开始
	LevelReceive   LogLevel = "📥" // 接收/数据块
	LevelExplain   LogLevel = "📚" // 解释/释义
	LevelTranslate LogLevel = "📝" // 翻译
)
```

### 2. StreamLogger 类型

用于记录流式操作的详细信息，自动计算耗时和统计数据。

```go
type StreamLogger struct {
	operationName string     // 操作名称
	service       string     // 服务名称
	startTime     time.Time  // 开始时间
	chunkCount    int        // 接收块数
	totalLength   int        // 总长度
}
```

## 核心函数

### 翻译相关

#### LogTranslateStart
记录流式翻译开始

```go
logger.LogTranslateStart(
	service,     // 翻译服务名称（如 "deepseek"）
	sourceLang,  // 源语言
	targetLang,  // 目标语言
	inputLen,    // 输入文本长度
)
```

#### LogTranslateSuccess
记录流式翻译成功完成

```go
logger.LogTranslateSuccess(
	service,    // 翻译服务名称
	inputLen,   // 输入长度
	outputLen,  // 输出长度
	chunks,     // 总块数
	duration,   // 耗时
)
```

#### LogTranslateError
记录流式翻译失败

```go
logger.LogTranslateError(
	service,   // 翻译服务名称
	err,       // 错误对象
	chunks,    // 已接收块数
	duration,  // 耗时
)
```

#### LogNormalTranslateStart / LogNormalTranslateSuccess / LogNormalTranslateError
用于非流式翻译的日志记录

### 解释相关

#### LogExplainStart / LogExplainSuccess / LogExplainError
与翻译函数类似，但用于解释操作，额外包含 `templateID` 参数

```go
logger.LogExplainStart(service, templateID, inputLen)
logger.LogExplainSuccess(service, templateID, inputLen, outputLen, chunks, duration)
logger.LogExplainError(service, templateID, err, chunks, duration)
```

### 通用函数

#### LogChunkReceived
记录接收到的数据块（Debug级别）

```go
logger.LogChunkReceived(
	chunkCount,   // 当前块序号
	chunkSize,    // 当前块大小
	totalLength,  // 累积总长度
)
```

#### LogStreamNotSupported
记录不支持流式的情况

```go
logger.LogStreamNotSupported(service)
```

#### LogStreamExplainNotSupported
记录不支持流式解释

```go
logger.LogStreamExplainNotSupported(service, templateID)
```

## 使用示例

### 示例 1: 流式翻译

```go
// 记录开始
logger.LogTranslateStart("deepseek", "auto", "zh", 100)

// 处理数据流
for chunk := range translateStream {
	logger.LogChunkReceived(chunkCount, len(chunk), totalLen)
	chunkCount++
	totalLen += len(chunk)
}

// 记录成功或失败
if err != nil {
	logger.LogTranslateError("deepseek", err, chunkCount, elapsed)
} else {
	logger.LogTranslateSuccess("deepseek", 100, totalLen, chunkCount, elapsed)
}
```

### 示例 2: 使用 StreamLogger

```go
// 创建流式日志记录器
sl := logger.NewStreamLogger("translation", "deepseek")

// 记录开始
sl.LogStart(logger.LevelSend, "开始翻译")

// 处理每个数据块
for chunk := range stream {
	sl.LogChunk(len(chunk))
}

// 记录结束
if err != nil {
	sl.LogError("翻译失败", err)
} else {
	sl.LogSuccess("翻译完成")
}
```

## 日志输出示例

### 流式翻译成功
```
2024-11-17 15:30:45 INFO  📤 开始流式翻译 service=deepseek source_lang=auto target_lang=zh input_length=50
2024-11-17 15:30:45 DEBUG 📥 接收流式数据块 chunk_count=1 chunk_size=20 total_length=20
2024-11-17 15:30:45 DEBUG 📥 接收流式数据块 chunk_count=2 chunk_size=25 total_length=45
2024-11-17 15:30:47 INFO  ✅ 流式翻译完成 service=deepseek input_length=50 output_length=45 total_chunks=2 duration=2.1s
```

### 流式翻译失败
```
2024-11-17 15:30:45 INFO  📤 开始流式翻译 service=deepseek source_lang=auto target_lang=zh input_length=50
2024-11-17 15:30:45 DEBUG 📥 接收流式数据块 chunk_count=1 chunk_size=20 total_length=20
2024-11-17 15:30:47 ERROR ❌ 流式翻译失败 service=deepseek error=connection timeout chunks_received=1 duration=2.1s
```

### 不支持流式
```
2024-11-17 15:30:45 INFO  ⚠️ 不支持流式翻译，使用普通模式 service=baidu
2024-11-17 15:30:45 INFO  📝 开始普通翻译 service=baidu source_lang=auto target_lang=zh
2024-11-17 15:30:47 INFO  ✅ 普通翻译完成 service=baidu results_count=1
```

## 集成步骤

1. **导入包**
   ```go
   import "handy-translate/logger"
   ```

2. **替换直接的 slog 调用**
   ```go
   // 旧方式
   slog.Info("📤 开始流式翻译",
       slog.String("service", service),
       slog.String("source_lang", fromLang),
       ...)
   
   // 新方式
   logger.LogTranslateStart(service, fromLang, toLang, len(queryText))
   ```

3. **确保参数一致**
   - 使用正确的服务名称
   - 正确计算输入/输出长度
   - 准确记录块数和耗时

## 优势

✅ **一致性**: 统一的日志格式和风格  
✅ **简洁性**: 减少重复代码，调用更简洁  
✅ **可维护性**: 统一管理日志逻辑，修改时只需改一个地方  
✅ **易用性**: 提供高级函数，无需手动传递所有参数  
✅ **灵活性**: StreamLogger 提供面向对象的日志记录方式  

## 后续扩展

- [ ] 添加日志级别配置（如只输出 Warning 及以上）
- [ ] 添加日志输出到文件的功能
- [ ] 添加性能指标分析（平均速率等）
- [ ] 支持自定义日志格式
