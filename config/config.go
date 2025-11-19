package config

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/sirupsen/logrus"
)

// Data config
var Data config

type (
	config struct {
		Appname          string                 `toml:"appname"`
		Keyboards        map[string][]string    `toml:"keyboards"`
		TranslateWay     string                 `toml:"translate_way"`
		Translate        map[string]Translate   `toml:"translate"`
		ExplainTemplates ExplainTemplatesConfig `toml:"explain_templates"`
		History          HistoryConfig          `toml:"history"`
		ToolbarMode      string                 `toml:"toolbar_mode"`
	}

	Translate struct {
		Name  string `toml:"name" json:"name,omitempty"`
		AppID string `toml:"appID" json:"appID,omitempty"`
		Key   string `toml:"key" json:"key,omitempty"`
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

// Init  config
func Init(projectName string) {
	filePath, _ := os.Getwd()
	b := strings.Index(filePath, projectName)
	configPath := filePath[:b+len(projectName)]

	configFile, err := os.Open(configPath + "/config.toml")
	if err != nil {
		logrus.WithError(err).Error("打开配置文件失败，将使用默认配置")
		return // ← 改为 return，而不是 os.Exit(1)，允许应用继续运行
	}
	defer configFile.Close()

	fd, err := io.ReadAll(configFile)
	if err != nil {
		logrus.WithError(err).Error("读取配置文件失败")
		return // ← 改为 return
	}
	err = toml.Unmarshal(fd, &Data)
	if err != nil {
		logrus.WithError(err).Error("解析配置文件失败")
		return // ← 改为 return
	}

	// 只在调试模式下打印配置（减少启动时 I/O）
	if os.Getenv("DEBUG") == "true" {
		fmt.Printf("配置已加载: %+v\n", Data)
	}
}

func Save() error {
	filePath := "./config.toml"
	data, err := toml.Marshal(&Data)
	if err != nil {
		logrus.WithError(err).Error("Marshal config failed")
		return fmt.Errorf("marshal config: %w", err)
	}

	// 使用原子写入: 先写临时文件，然后重命名
	// 这样可以避免写入失败导致配置文件损坏
	tempFilePath := filePath + ".tmp"

	// 创建临时文件
	file, err := os.Create(tempFilePath)
	if err != nil {
		logrus.WithError(err).Error("Create temp config file failed")
		return fmt.Errorf("create temp file: %w", err)
	}

	// 写入数据
	_, err = file.Write(data)
	if err != nil {
		file.Close()
		os.Remove(tempFilePath) // 清理临时文件
		logrus.WithError(err).Error("Write config failed")
		return fmt.Errorf("write config: %w", err)
	}

	// 确保数据写入磁盘
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(tempFilePath)
		logrus.WithError(err).Error("Sync config failed")
		return fmt.Errorf("sync config: %w", err)
	}

	file.Close()

	// 原子性地重命名临时文件为目标文件
	if err := os.Rename(tempFilePath, filePath); err != nil {
		os.Remove(tempFilePath) // 清理临时文件
		logrus.WithError(err).Error("Rename config file failed")
		return fmt.Errorf("rename config file: %w", err)
	}

	logrus.Info("Config saved successfully")
	return nil
}
