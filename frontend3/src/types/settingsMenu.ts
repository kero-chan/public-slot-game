export type SettingsMenuContentType = 'sound' | 'paytable' | 'guides' | 'histories' | null

export interface SettingsMenuState {
  isOpen: boolean
  activeContent: SettingsMenuContentType
}
