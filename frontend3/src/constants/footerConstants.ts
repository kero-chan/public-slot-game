/**
 * Footer Panel Constants
 */

export const FOOTER_CONTENT_TYPES = {
  RULES: 'rules',
  HISTORY: 'history',
  GUIDE: 'guide'
} as const

export const FOOTER_CONTENT_TITLES = {
  [FOOTER_CONTENT_TYPES.RULES]: '赔付表',
  [FOOTER_CONTENT_TYPES.HISTORY]: '历史记录',
  [FOOTER_CONTENT_TYPES.GUIDE]: '游戏规则'
} as const

export type FooterContentType = typeof FOOTER_CONTENT_TYPES[keyof typeof FOOTER_CONTENT_TYPES]