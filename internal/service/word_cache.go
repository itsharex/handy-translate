// Package service 提供应用层业务编排。
// WordCache 实现单词查询的文件持久化缓存。
package service

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// WordCache 单词查询结果的文件缓存。
// 缓存目录: data/word_cache/<word>.json
type WordCache struct {
	dir string
	mu  sync.RWMutex
	mem map[string]string // 内存热缓存，避免重复读磁盘
}

// NewWordCache 创建单词缓存实例。
func NewWordCache(cacheDir string) *WordCache {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		slog.Error("创建缓存目录失败", slog.String("dir", cacheDir), slog.Any("err", err))
	}
	return &WordCache{
		dir: cacheDir,
		mem: make(map[string]string),
	}
}

// Get 查询缓存，返回 JSON 字符串和是否命中。
func (c *WordCache) Get(word string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(word))

	// 1. 内存缓存
	c.mu.RLock()
	if val, ok := c.mem[key]; ok {
		c.mu.RUnlock()
		return val, true
	}
	c.mu.RUnlock()

	// 2. 文件缓存
	filePath := filepath.Join(c.dir, key+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", false
	}

	jsonStr := string(data)

	// 校验 JSON 合法性
	if !json.Valid(data) {
		slog.Warn("缓存文件 JSON 无效，删除", slog.String("file", filePath))
		os.Remove(filePath)
		return "", false
	}

	// 加载到内存缓存
	c.mu.Lock()
	c.mem[key] = jsonStr
	c.mu.Unlock()

	return jsonStr, true
}

// Set 写入缓存（内存 + 文件）。
func (c *WordCache) Set(word, jsonResult string) {
	key := strings.ToLower(strings.TrimSpace(word))

	// 校验 JSON
	if !json.Valid([]byte(jsonResult)) {
		slog.Warn("缓存写入跳过：非法 JSON", slog.String("word", key))
		return
	}

	// 内存缓存
	c.mu.Lock()
	c.mem[key] = jsonResult
	c.mu.Unlock()

	// 文件缓存（异步写入，不阻塞主流程）
	go func() {
		filePath := filepath.Join(c.dir, key+".json")
		if err := os.WriteFile(filePath, []byte(jsonResult), 0644); err != nil {
			slog.Error("缓存写入失败", slog.String("file", filePath), slog.Any("err", err))
		} else {
			slog.Info("📦 单词缓存已保存", slog.String("word", key))
		}
	}()
}
