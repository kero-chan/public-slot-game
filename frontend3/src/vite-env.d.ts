/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string
  readonly VITE_GAME_ID?: string
  readonly VITE_TEST_WIN_ANIMATIONS?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
