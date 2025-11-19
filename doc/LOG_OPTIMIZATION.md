# 流式数据翻译日志优化文档

## 优化概述

对流式数据翻译时的日志打印进行了全面优化，使日志更清晰、更便于人类阅读和调试。

## 主要改进

### 1. **日志级别优化**
- **Info 级别**: 用于重要操作的开始和完成
  - 使用 `🔍` 表示操作开始
  - 使用 `✅` 表示操作成功完成
  - 使用 `⚠️` 表示警告信息
  - 使用 `❌` 表示错误信息
  
- **Debug 级别**: 用于详细的流式数据块接收信息
  - 使用 `📥` 表示接收数据块
  - 用于追踪流式过程中的每个块，但不会污染主日志

### 2. **关键指标统计**

#### 流式翻译时记录的指标：
```
- 📤 开始流式翻译
  - service: 翻译服务名称
  - source_lang: 源语言
  - target_lang: 目标语言
  - input_length: 输入文本长度

- 📥 接收流式数据块 (Debug 级别)
  - chunk_count: 当前是第几个块
  - chunk_size: 当前块的大小
  - total_length: 累积的总长度

- ✅ 流式翻译完成
  - service: 翻译服务名称
  - input_length: 输入长度
  - output_length: 输出长度
  - total_chunks: 总块数
  - duration: 耗时

- ❌ 流式翻译失败
  - error: 错误信息
  - chunks_received: 已接收的块数
  - duration: 耗时
```

### 3. **文本预览优化**

新增 `truncateText()` 辅助函数，用于日志中显示文本预览：
- 限制显示长度为 50 个字符
- 超长文本用 `...` 表示截断
- 正确处理中文字符（按字符数而非字节数截断）

### 4. **错误日志改进**

错误信息现在更清晰：
```
❌ 流式翻译失败
  service: deepseek
  error: connection timeout
  chunks_received: 5
  duration: 30.5s
```

### 5. **支持的操作类型**

#### 流式翻译 (`TranslateStream`)
- 显示流式进度信息
- 记录接收块数和耗时
- 清晰的错误提示

#### 流式含义翻译 (`TranslateMeanings`)
- 区别于普通翻译
- 记录含义特定的统计数据

#### 流式解释 (`processExplain`)
- 支持模板 ID 记录
- 详细的解释过程指标

#### 普通翻译 (`processTranslate`)
- 自动检测是否支持流式
- 流式和普通模式的日志明确区分

## 日志示例

### 成功的流式翻译
```
INFO 📤 开始流式翻译 service=deepseek source_lang=auto target_lang=zh input_length=123
DEBUG 📥 接收流式数据块 chunk_count=1 chunk_size=45 total_length=45
DEBUG 📥 接收流式数据块 chunk_count=2 chunk_size=52 total_length=97
DEBUG 📥 接收流式数据块 chunk_count=3 chunk_size=38 total_length=135
INFO ✅ 流式翻译完成 service=deepseek input_length=123 output_length=135 total_chunks=3 duration=2.5s
```

### 流式翻译失败
```
INFO 📤 开始流式翻译 service=deepseek source_lang=auto target_lang=zh input_length=50
DEBUG 📥 接收流式数据块 chunk_count=1 chunk_size=30 total_length=30
ERROR ❌ 流式翻译失败 error=connection timeout chunks_received=1 duration=30.5s
```

### 不支持流式的服务
```
INFO ⚠️  不支持流式翻译，使用普通模式 service=baidu
INFO 📝 开始普通翻译 service=baidu source_lang=auto target_lang=zh
INFO ✅ 普通翻译完成 service=baidu results_count=1
```

## 技术细节

### 时间统计
- 使用 `time.Now()` 和 `time.Since()` 精确记录操作耗时
- 所有耗时都以易读的格式显示（如 `2.5s`、`150ms`）

### 数据大小跟踪
- 记录输入和输出的字符长度
- 统计总块数和平均块大小
- 帮助识别性能问题

### 文本安全
- 不在日志中打印完整的翻译内容
- 只在开始时显示输入文本预览（前 50 字符）
- 防止日志文件过大

## 调试建议

1. **性能分析**: 查看 `duration` 字段判断翻译速度
2. **流式质量**: 查看 `total_chunks` 判断流式粒度（块数越多越细致）
3. **错误追踪**: 查看 `chunks_received` 判断在第几块时出错
4. **容量规划**: 查看 `output_length` 判断输出数据量

## 后续扩展

- [ ] 可选的详细日志模式（显示完整内容）
- [ ] 日志统计汇总功能
- [ ] 性能热点分析
