package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Data config
var Data Config

// configFilePath 存储配置文件的绝对路径，Init 和 Save 共用。
var configFilePath string

type (
	Config struct {
		Appname          string                 `toml:"appname"`
		Keyboards        map[string][]string    `toml:"keyboards"`
		TranslateWay     string                 `toml:"translate_way"`
		Translate        map[string]Translate   `toml:"translate"`
		ExplainTemplates ExplainTemplatesConfig `toml:"explain_templates"`
		History          HistoryConfig          `toml:"history"`
		ToolbarMode      string                 `toml:"toolbar_mode"`
		ToolbarPinned    bool                   `toml:"toolbar_pinned"`
	}

	Translate struct {
		Name    string `toml:"name" json:"name,omitempty"`
		AppID   string `toml:"appID" json:"appID,omitempty"`
		Key     string `toml:"key" json:"key,omitempty"`
		BaseURL string `toml:"base_url" json:"base_url,omitempty"`
		Model   string `toml:"model" json:"model,omitempty"`
	}

	ExplainTemplatesConfig struct {
		DefaultTemplate string                     `toml:"default_template"`
		Templates       map[string]ExplainTemplate `toml:"templates"`
	}

	ExplainTemplate struct {
		Name        string `toml:"name" json:"name"`
		Description string `toml:"description" json:"description"`
		Template    string `toml:"template" json:"template"`
	}

	HistoryConfig struct {
		Enabled     bool   `toml:"enabled"`
		StoragePath string `toml:"storage_path"`
	}
)

// Init 初始化配置。
func Init(projectName string) {
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("获取工作目录失败", slog.Any("error", err))
		return
	}

	idx := strings.Index(cwd, projectName)
	if idx == -1 {
		slog.Error("工作目录中未找到项目名，将使用默认配置",
			slog.String("cwd", cwd),
			slog.String("projectName", projectName))
		return
	}

	configDir := cwd[:idx+len(projectName)]
	configFilePath = filepath.Join(configDir, "config.toml")

	configFile, err := os.Open(configFilePath)
	if err != nil {
		slog.Error("打开配置文件失败，将使用默认配置", slog.Any("error", err))
		return
	}
	defer configFile.Close()

	fd, err := io.ReadAll(configFile)
	if err != nil {
		slog.Error("读取配置文件失败", slog.Any("error", err))
		return
	}
	err = toml.Unmarshal(fd, &Data)
	if err != nil {
		slog.Error("解析配置文件失败", slog.Any("error", err))
		return
	}

	// 只在调试模式下打印配置（减少启动时 I/O）
	if os.Getenv("DEBUG") == "true" {
		fmt.Printf("配置已加载: %+v\n", Data)
	}
}

// Save 保存配置到文件（原子写入）。
func Save() error {
	if configFilePath == "" {
		configFilePath = "./config.toml"
	}

	data, err := toml.Marshal(&Data)
	if err != nil {
		slog.Error("Marshal config failed", slog.Any("error", err))
		return fmt.Errorf("marshal config: %w", err)
	}

	// 使用原子写入: 先写临时文件，然后重命名
	// 这样可以避免写入失败导致配置文件损坏
	tempFilePath := configFilePath + ".tmp"

	// 创建临时文件
	file, err := os.Create(tempFilePath)
	if err != nil {
		slog.Error("Create temp config file failed", slog.Any("error", err))
		return fmt.Errorf("create temp file: %w", err)
	}

	// 写入数据
	_, err = file.Write(data)
	if err != nil {
		file.Close()
		os.Remove(tempFilePath) // 清理临时文件
		slog.Error("Write config failed", slog.Any("error", err))
		return fmt.Errorf("write config: %w", err)
	}

	// 确保数据写入磁盘
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(tempFilePath)
		slog.Error("Sync config failed", slog.Any("error", err))
		return fmt.Errorf("sync config: %w", err)
	}

	file.Close()

	// 原子性地重命名临时文件为目标文件
	if err := os.Rename(tempFilePath, configFilePath); err != nil {
		os.Remove(tempFilePath) // 清理临时文件
		slog.Error("Rename config file failed", slog.Any("error", err))
		return fmt.Errorf("rename config file: %w", err)
	}

	slog.Debug("Config saved successfully")
	return nil
}
