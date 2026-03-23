/**
 * 主题配置文件 — 定义多个预设配色方案
 * 每个主题包含 CSS 变量 + 组件动态 class
 */

export const themes = {
  warm: {
    id: 'warm',
    name: '暖琥珀',
    dot: '#D97706',  // 主题预览色块
    cssVars: {
      '--color-primary': '#C2772E',
      '--color-bg-page': '#FDFBF7',
      '--color-text-main': '#3D2E1F',
      '--color-text-secondary': '#8B7355',
      '--color-bg-tab': '#F5F0E8',
      '--color-bg-example': '#FFF8F0',
      '--color-highlight-text': '#B45309',
      '--color-highlight-bg': '#FEF3C7',
      '--color-border-light': '#F0E6D8',
      '--color-border-section': '#E8DDD0',
      '--color-skeleton-from': '#F5F0E8',
      '--color-skeleton-mid': '#EDE4D4',
      '--color-heading': '#2D1F10',
      '--color-body': '#4A3728',
      '--color-muted': '#8B7355',
      '--color-code-bg': '#FFF5E6',
      '--color-code-text': '#B45309',
      '--color-code-border': 'rgba(180, 83, 9, 0.1)',
      '--color-pre-bg': '#FDF8F0',
      '--color-pre-border': '#EDE4D4',
      '--color-quote-border': '#D4A574',
      '--color-cursor': '#D97706',
      '--color-empty': '#B8A08A',
      '--color-example-bg': '#FFF8F0',
      '--color-example-border': '#D4A574',
      '--color-example-block-bg': 'linear-gradient(135deg, #FFF8F0 0%, #FFF3E6 100%)',
    },
    classes: {
      header: 'bg-gradient-to-r from-[#FFF3E0] to-[#FFE8CC] border-b border-[#F5D5A8]/50',
      logoBg: 'bg-white/70 backdrop-blur-sm border border-amber-100/60',
      tabList: 'bg-amber-50/80 backdrop-blur-md rounded-[10px] p-[4px] border border-amber-200/50',
      tabCursor: 'bg-gradient-to-r from-amber-500 to-orange-500 shadow-md shadow-amber-500/20 rounded-[8px]',
      tabContent: 'text-amber-700/60 group-data-[selected=true]:text-white group-data-[selected=true]:font-[500]',
      pinActive: 'bg-gradient-to-tr from-amber-500 to-orange-500 text-white shadow-md shadow-amber-600/30 pinned-glow border border-amber-400',
      pinInactive: 'bg-white/70 backdrop-blur-sm text-amber-700/50 hover:bg-amber-50 hover:text-amber-600 active:bg-amber-100 active:text-amber-700 border border-amber-100/60 hover:border-amber-200 shadow-sm',
      copyActive: 'bg-gradient-to-tr from-amber-400 to-orange-500 text-white shadow-md shadow-amber-500/30 border border-amber-300',
      copyInactive: 'bg-white/70 backdrop-blur-sm text-amber-700/50 hover:bg-amber-50 hover:text-amber-600 active:bg-amber-100 active:text-amber-700 border border-amber-100/60 hover:border-amber-200 shadow-sm',
      speakBtn: 'text-amber-600 hover:bg-amber-50 hover:text-amber-700 active:bg-amber-100',
      speakBtnExample: 'bg-white/50 text-amber-600 hover:bg-white hover:text-amber-700 hover:shadow-sm',
      themeBtn: 'bg-white/70 backdrop-blur-sm text-amber-700/50 hover:bg-amber-50 hover:text-amber-600 active:bg-amber-100 active:text-amber-700 border border-amber-100/60 hover:border-amber-200 shadow-sm',
    }
  },

  cool: {
    id: 'cool',
    name: '清凉蓝',
    dot: '#3B82F6',
    cssVars: {
      '--color-primary': '#2E75B6',
      '--color-bg-page': '#F9FAFB',
      '--color-text-main': '#1F2937',
      '--color-text-secondary': '#6B7280',
      '--color-bg-tab': '#F3F4F6',
      '--color-bg-example': '#F0F7FF',
      '--color-highlight-text': '#1D4ED8',
      '--color-highlight-bg': '#FEF9C3',
      '--color-border-light': '#F3F4F6',
      '--color-border-section': '#E5E7EB',
      '--color-skeleton-from': '#f1f5f9',
      '--color-skeleton-mid': '#e2e8f0',
      '--color-heading': '#0f172a',
      '--color-body': '#334155',
      '--color-muted': '#64748b',
      '--color-code-bg': '#f1f5f9',
      '--color-code-text': '#dc2626',
      '--color-code-border': 'rgba(0, 0, 0, 0.06)',
      '--color-pre-bg': '#f8fafc',
      '--color-pre-border': '#e2e8f0',
      '--color-quote-border': '#3b82f6',
      '--color-cursor': '#3b82f6',
      '--color-empty': '#9ca3af',
      '--color-example-bg': '#F0F7FF',
      '--color-example-border': '#2E75B6',
      '--color-example-block-bg': 'linear-gradient(135deg, #f8faff 0%, #f0f4ff 100%)',
    },
    classes: {
      header: 'bg-gradient-to-r from-[#d4f0f7] to-[#dbeafe] border-b border-[#bae6fd]/50',
      logoBg: 'bg-white/60 backdrop-blur-sm border border-white/40',
      tabList: 'bg-slate-100/80 backdrop-blur-md rounded-[10px] p-[4px] border border-slate-200/50',
      tabCursor: 'bg-gradient-to-r from-blue-500 to-indigo-500 shadow-md shadow-indigo-500/20 rounded-[8px]',
      tabContent: 'text-slate-500 group-data-[selected=true]:text-white group-data-[selected=true]:font-[500]',
      pinActive: 'bg-gradient-to-tr from-blue-500 to-indigo-500 text-white shadow-md shadow-blue-600/30 pinned-glow border border-blue-400',
      pinInactive: 'bg-white/60 backdrop-blur-sm text-slate-500 hover:bg-blue-50 hover:text-blue-500 active:bg-blue-100 active:text-blue-600 border border-white/40 hover:border-blue-200 shadow-sm',
      copyActive: 'bg-gradient-to-tr from-emerald-400 to-teal-500 text-white shadow-md shadow-emerald-500/30 border border-emerald-300',
      copyInactive: 'bg-white/60 backdrop-blur-sm text-slate-500 hover:bg-emerald-50 hover:text-emerald-500 active:bg-emerald-100 active:text-emerald-600 border border-white/40 hover:border-emerald-200 shadow-sm',
      speakBtn: 'text-indigo-500 hover:bg-indigo-50 hover:text-indigo-600 active:bg-indigo-100',
      speakBtnExample: 'bg-white/50 text-indigo-500 hover:bg-white hover:text-indigo-600 hover:shadow-sm',
      themeBtn: 'bg-white/60 backdrop-blur-sm text-slate-500 hover:bg-blue-50 hover:text-blue-500 active:bg-blue-100 active:text-blue-600 border border-white/40 hover:border-blue-200 shadow-sm',
    }
  },

  violet: {
    id: 'violet',
    name: '优雅紫',
    dot: '#8B5CF6',
    cssVars: {
      '--color-primary': '#7C3AED',
      '--color-bg-page': '#FAFAFF',
      '--color-text-main': '#1E1B3A',
      '--color-text-secondary': '#7C728E',
      '--color-bg-tab': '#F3F0F9',
      '--color-bg-example': '#F5F0FF',
      '--color-highlight-text': '#7C3AED',
      '--color-highlight-bg': '#EDE9FE',
      '--color-border-light': '#E8E0F5',
      '--color-border-section': '#DDD5ED',
      '--color-skeleton-from': '#F3F0F9',
      '--color-skeleton-mid': '#E8E0F5',
      '--color-heading': '#1E1035',
      '--color-body': '#3D3556',
      '--color-muted': '#7C728E',
      '--color-code-bg': '#F5F0FF',
      '--color-code-text': '#7C3AED',
      '--color-code-border': 'rgba(124, 58, 237, 0.1)',
      '--color-pre-bg': '#FAF8FF',
      '--color-pre-border': '#E8E0F5',
      '--color-quote-border': '#A78BFA',
      '--color-cursor': '#8B5CF6',
      '--color-empty': '#B0A8C0',
      '--color-example-bg': '#F5F0FF',
      '--color-example-border': '#A78BFA',
      '--color-example-block-bg': 'linear-gradient(135deg, #FAF8FF 0%, #F0EBFF 100%)',
    },
    classes: {
      header: 'bg-gradient-to-r from-[#F3EEFF] to-[#E8DFFF] border-b border-[#D8CCFA]/50',
      logoBg: 'bg-white/70 backdrop-blur-sm border border-violet-100/60',
      tabList: 'bg-violet-50/80 backdrop-blur-md rounded-[10px] p-[4px] border border-violet-200/50',
      tabCursor: 'bg-gradient-to-r from-violet-500 to-purple-500 shadow-md shadow-violet-500/20 rounded-[8px]',
      tabContent: 'text-violet-600/60 group-data-[selected=true]:text-white group-data-[selected=true]:font-[500]',
      pinActive: 'bg-gradient-to-tr from-violet-500 to-purple-500 text-white shadow-md shadow-violet-600/30 pinned-glow border border-violet-400',
      pinInactive: 'bg-white/70 backdrop-blur-sm text-violet-500/50 hover:bg-violet-50 hover:text-violet-600 active:bg-violet-100 active:text-violet-700 border border-violet-100/60 hover:border-violet-200 shadow-sm',
      copyActive: 'bg-gradient-to-tr from-violet-400 to-purple-500 text-white shadow-md shadow-violet-500/30 border border-violet-300',
      copyInactive: 'bg-white/70 backdrop-blur-sm text-violet-500/50 hover:bg-violet-50 hover:text-violet-600 active:bg-violet-100 active:text-violet-700 border border-violet-100/60 hover:border-violet-200 shadow-sm',
      speakBtn: 'text-violet-500 hover:bg-violet-50 hover:text-violet-600 active:bg-violet-100',
      speakBtnExample: 'bg-white/50 text-violet-500 hover:bg-white hover:text-violet-600 hover:shadow-sm',
      themeBtn: 'bg-white/70 backdrop-blur-sm text-violet-500/50 hover:bg-violet-50 hover:text-violet-600 active:bg-violet-100 active:text-violet-700 border border-violet-100/60 hover:border-violet-200 shadow-sm',
    }
  },

  rose: {
    id: 'rose',
    name: '温柔粉',
    dot: '#F43F5E',
    cssVars: {
      '--color-primary': '#E11D48',
      '--color-bg-page': '#FFFBFC',
      '--color-text-main': '#2D1A24',
      '--color-text-secondary': '#8E6B7A',
      '--color-bg-tab': '#FDF0F3',
      '--color-bg-example': '#FFF5F7',
      '--color-highlight-text': '#BE123C',
      '--color-highlight-bg': '#FFE4E6',
      '--color-border-light': '#F5DDE3',
      '--color-border-section': '#EDCDD6',
      '--color-skeleton-from': '#FDF0F3',
      '--color-skeleton-mid': '#F5DDE3',
      '--color-heading': '#2D1019',
      '--color-body': '#4A2838',
      '--color-muted': '#8E6B7A',
      '--color-code-bg': '#FFF5F7',
      '--color-code-text': '#BE123C',
      '--color-code-border': 'rgba(190, 18, 60, 0.1)',
      '--color-pre-bg': '#FFFAFC',
      '--color-pre-border': '#F5DDE3',
      '--color-quote-border': '#FB7185',
      '--color-cursor': '#F43F5E',
      '--color-empty': '#C4A0AE',
      '--color-example-bg': '#FFF5F7',
      '--color-example-border': '#FB7185',
      '--color-example-block-bg': 'linear-gradient(135deg, #FFFAFC 0%, #FFF0F3 100%)',
    },
    classes: {
      header: 'bg-gradient-to-r from-[#FFF0F3] to-[#FFE4E8] border-b border-[#FBB6C5]/40',
      logoBg: 'bg-white/70 backdrop-blur-sm border border-rose-100/60',
      tabList: 'bg-rose-50/80 backdrop-blur-md rounded-[10px] p-[4px] border border-rose-200/50',
      tabCursor: 'bg-gradient-to-r from-rose-500 to-pink-500 shadow-md shadow-rose-500/20 rounded-[8px]',
      tabContent: 'text-rose-600/60 group-data-[selected=true]:text-white group-data-[selected=true]:font-[500]',
      pinActive: 'bg-gradient-to-tr from-rose-500 to-pink-500 text-white shadow-md shadow-rose-600/30 pinned-glow border border-rose-400',
      pinInactive: 'bg-white/70 backdrop-blur-sm text-rose-500/50 hover:bg-rose-50 hover:text-rose-600 active:bg-rose-100 active:text-rose-700 border border-rose-100/60 hover:border-rose-200 shadow-sm',
      copyActive: 'bg-gradient-to-tr from-rose-400 to-pink-500 text-white shadow-md shadow-rose-500/30 border border-rose-300',
      copyInactive: 'bg-white/70 backdrop-blur-sm text-rose-500/50 hover:bg-rose-50 hover:text-rose-600 active:bg-rose-100 active:text-rose-700 border border-rose-100/60 hover:border-rose-200 shadow-sm',
      speakBtn: 'text-rose-500 hover:bg-rose-50 hover:text-rose-600 active:bg-rose-100',
      speakBtnExample: 'bg-white/50 text-rose-500 hover:bg-white hover:text-rose-600 hover:shadow-sm',
      themeBtn: 'bg-white/70 backdrop-blur-sm text-rose-500/50 hover:bg-rose-50 hover:text-rose-600 active:bg-rose-100 active:text-rose-700 border border-rose-100/60 hover:border-rose-200 shadow-sm',
    }
  },
}

export const themeIds = Object.keys(themes)
export const defaultThemeId = 'warm'

/**
 * 应用主题 — 将 CSS 变量写入 :root
 */
export function applyThemeCssVars(themeId) {
  const theme = themes[themeId]
  if (!theme) return
  const root = document.documentElement
  Object.entries(theme.cssVars).forEach(([key, value]) => {
    root.style.setProperty(key, value)
  })
}

/**
 * 从 localStorage 读取已保存的主题 ID
 */
export function getSavedThemeId() {
  try {
    return localStorage.getItem('handy-translate-theme') || defaultThemeId
  } catch {
    return defaultThemeId
  }
}

/**
 * 保存主题 ID 到 localStorage
 */
export function saveThemeId(themeId) {
  try {
    localStorage.setItem('handy-translate-theme', themeId)
  } catch {
    // ignore
  }
}
