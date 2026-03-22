// Package app 提供应用核心结构（依赖注入模式）。
package app

import (
	"sync"
)

// State 集中管理应用的可变状态，替代散落各处的全局变量。
// 所有字段通过方法访问，保证并发安全。
type State struct {
	mu             sync.RWMutex
	fromLang       string
	toLang         string
	currentQuery   string
	toolbarMode    string
}

// NewState 创建默认状态。
func NewState() *State {
	return &State{
		fromLang:    "auto",
		toLang:      "zh",
		toolbarMode: "translate",
	}
}

// SetLangs 设置翻译语言。
func (s *State) SetLangs(from, to string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fromLang = from
	s.toLang = to
}

// GetLangs 获取翻译语言。
func (s *State) GetLangs() (string, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fromLang, s.toLang
}

// SetCurrentQuery 设置当前查询文本。
func (s *State) SetCurrentQuery(query string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentQuery = query
}

// GetCurrentQuery 获取当前查询文本。
func (s *State) GetCurrentQuery() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentQuery
}

// SetToolbarMode 设置工具栏模式。
func (s *State) SetToolbarMode(mode string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toolbarMode = mode
}

// GetToolbarMode 获取工具栏模式。
func (s *State) GetToolbarMode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.toolbarMode
}
