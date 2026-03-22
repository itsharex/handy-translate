package config

// FindTemplate 在配置中查找指定的提示词模板。
// 查找优先级：指定 ID → 默认模板 → 第一个可用模板。
// 返回模板内容字符串，如果未找到则返回空字符串。
func FindTemplate(templates *ExplainTemplatesConfig, templateID string) string {
	if templates == nil || len(templates.Templates) == 0 {
		return ""
	}

	// 1. 如果 templateID 为空，使用默认模板 ID
	if templateID == "" {
		templateID = templates.DefaultTemplate
	}

	// 2. 如果默认模板 ID 也为空，使用第一个可用模板
	if templateID == "" {
		for id := range templates.Templates {
			templateID = id
			break
		}
	}

	// 3. 按 ID 查找
	if tmpl, exists := templates.Templates[templateID]; exists {
		return tmpl.Template
	}

	// 4. 回退到默认模板
	if templates.DefaultTemplate != "" {
		if tmpl, exists := templates.Templates[templates.DefaultTemplate]; exists {
			return tmpl.Template
		}
	}

	// 5. 最后尝试使用第一个可用模板
	for _, tmpl := range templates.Templates {
		return tmpl.Template
	}

	return ""
}
