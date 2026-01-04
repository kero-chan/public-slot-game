// @ts-nocheck
import { Container, Texture, Rectangle, Sprite } from 'pixi.js'
import type { Application } from 'pixi.js'
import { ASSETS } from '@/config/assets'
import { TILE_SLICES, BASE_RECT } from './config'
import type { TileConfig, Rectangle as ConfigRectangle, SourceSpriteIcon } from './config'

let composing = false

export function composeTilesTextures(app: Application): void {
  if (composing) return

  const sheetTex = ASSETS.loadedImages?.tiles_50
  const source = sheetTex?.source || (sheetTex as any)?.baseTexture
  if (!source) return

  composing = true
  try {
    Object.entries(TILE_SLICES).forEach(([symbol, cfg]: [string, TileConfig]) => {
      if (ASSETS.loadedImages?.[symbol]) return

      const outSize = cfg.outSize || Math.max((BASE_RECT?.w || 256), (BASE_RECT?.h || 256))
      const container = new Container()
      const makeSubTex = (source: any, r: ConfigRectangle) => new Texture({ source, frame: new Rectangle(r.x, r.y, r.w, r.h) })

      // Draw base only if the tile defines one
      if (cfg.base) {
        let baseAlias: string
        let baseRect: ConfigRectangle | undefined
        let baseScale = 1
        let baseOffsetX = 0
        let baseOffsetY = 0

        if (typeof cfg.base === 'object' && 'icon' in cfg.base) {
          const sourceSpriteIcon = cfg.base as SourceSpriteIcon
          baseAlias = sourceSpriteIcon.sourceSprite || 'tiles_50'
          baseRect = sourceSpriteIcon.icon
          baseScale = sourceSpriteIcon.scale ?? 1
          baseOffsetX = sourceSpriteIcon.offsetX ?? 0
          baseOffsetY = sourceSpriteIcon.offsetY ?? 0
        } else {
          baseAlias = cfg.baseSourceSprite || (cfg.layers?.[0]?.sourceSprite) || 'tiles_50'
          baseRect = cfg.base as ConfigRectangle
          baseScale = cfg.baseScale ?? 1
          baseOffsetX = cfg.baseOffsetX ?? 0
          baseOffsetY = cfg.baseOffsetY ?? 0
        }

        const baseSheetTex = ASSETS.loadedImages?.[baseAlias]
        const baseSource = baseSheetTex?.source || (baseSheetTex as any)?.baseTexture
        if (baseSource && baseRect) {
          const baseSp = new Sprite(new Texture({ source: baseSource, frame: new Rectangle(baseRect.x, baseRect.y, baseRect.w, baseRect.h) }))
          baseSp.anchor.set(0.5)
          baseSp.width = Math.floor(outSize * baseScale)
          baseSp.height = Math.floor(outSize * baseScale)
          baseSp.x = Math.floor(outSize / 2) + Math.floor(outSize * baseOffsetX)
          baseSp.y = Math.floor(outSize / 2) + Math.floor(outSize * baseOffsetY)
          container.addChild(baseSp)
        }
      }

      // Layered composition
      const addLayer = (layer: any): void => {
        const layerAlias = layer.sourceSprite || 'tiles_50'
        const sheetTex = ASSETS.loadedImages?.[layerAlias]
        const source = sheetTex?.source || (sheetTex as any)?.baseTexture
        if (!source || !layer.icon) return

        const sprite = new Sprite(makeSubTex(source, layer.icon))
        sprite.anchor.set(0.5)

        const scale = layer.scale ?? cfg.iconScale ?? 0.78
        sprite.width = Math.floor(outSize * scale)
        sprite.height = Math.floor(outSize * scale)

        const offsetX = Math.floor(outSize * (layer.offsetX || 0))
        const offsetY = Math.floor(outSize * (layer.offsetY || 0))
        sprite.x = Math.floor(outSize / 2) + offsetX
        sprite.y = Math.floor(outSize / 2) + offsetY

        container.addChild(sprite)
      }

      if (Array.isArray(cfg.layers) && cfg.layers.length) {
        cfg.layers.forEach(addLayer)
      } else if (cfg.icon) {
        addLayer({ icon: cfg.icon, sourceSprite: cfg.sourceSprite })
      } else if (cfg.iconAsset) {
        const iconTex = ASSETS.loadedImages?.[cfg.iconAsset]
        if (iconTex) {
          const sprite = new Sprite(iconTex)
          sprite.anchor.set(0.5)
          const scale = cfg.iconScale ?? 0.78
          sprite.width = Math.floor(outSize * scale)
          sprite.height = Math.floor(outSize * scale)
          sprite.x = Math.floor(outSize / 2)
          sprite.y = Math.floor(outSize / 2)
          container.addChild(sprite)
        }
      }

      const composedTex = app.renderer.generateTexture({
        target: container,
        resolution: window.devicePixelRatio || 2
      })
      if (composedTex && composedTex.width > 0 && composedTex.height > 0) {
        // Set texture quality for crisp rendering
        if (composedTex.source) {
          composedTex.source.scaleMode = 'linear'
          composedTex.source.autoGenerateMipmaps = true
        }
        ASSETS.loadedImages[symbol] = composedTex
      }

      container.removeChildren()
    })
  } finally {
    composing = false
  }
}
