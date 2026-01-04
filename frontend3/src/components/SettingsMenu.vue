<template>
  <Teleport to="body">
    <Transition name="fade">
      <div v-if="isOpen" class="settings-overlay" @click.self="handleClose">
        <!-- Settings Menu with Background Image -->
        <div class="settings-menu" :style="menuStyle">
          <!-- Close Button -->
          <img
            :src="menuClose"
            alt="Close"
            class="menu-close-btn"
            @click="handleClose"
          />

          <!-- Menu Item Buttons -->
          <div class="menu-buttons-container">
            <template v-for="button in menuButtons" :key="button.id">
              <!-- Sound button with icon -->
              <div
                v-if="button.id === 'sound'"
                class="menu-button-wrapper menu-button--sound"
              >
                <div class="sound-container-item">
                  <img
                    :src="getButtonImage(button.id)"
                    :alt="button.label"
                    class="menu-button sound-button"
                    @click="handleButtonClick(button.id)"
                  />
                  <img
                    :src="gameSound ? iconSoundOn : iconSoundOff"
                    :alt="gameSound ? 'Sound On' : 'Sound Off'"
                    class="sound-icon"
                    @click="handleButtonClick(button.id)"
                  />
                </div>
              </div>

              <!-- Other buttons -->
              <img
                v-else
                :src="getButtonImage(button.id)"
                :alt="button.label"
                class="menu-button"
                :class="`menu-button--${button.id}`"
                @click="handleButtonClick(button.id)"
              />
            </template>
          </div>
        </div>

        <!-- Content Layer - Same dimensions as overlay -->
        <Transition name="slide-fade">
          <div
            v-if="activeContent"
            class="content-layer"
            @click.self="closeContent"
          >
            <div class="content-container">
              <!-- Paytable Content -->
              <RulesContent v-if="activeContent === 'paytable'" />

              <!-- Guides Content -->
              <GuideContent v-if="activeContent === 'guides'" />

              <!-- Game Histories Content -->
              <HistoryContent v-if="activeContent === 'histories'" />

              <!-- Close Button for Content -->
              <button
                class="content-close-btn"
                @click="closeContent"
                aria-label="Close"
              >
                ×
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useUIStore } from "@/stores";
import { useAuthStore, useSettingsStore } from "@/stores/user";
import menuFrame from "@/assets/images/japaneseOmakase/menu/menu_frame.webp";
import menuClose from "@/assets/images/japaneseOmakase/icons/icon_close.webp";
import menuSound from "@/assets/images/japaneseOmakase/menu/menu_sound_item.webp";
import menuPaytable from "@/assets/images/japaneseOmakase/menu/menu_paytable_item.webp";
import menuHowToPlay from "@/assets/images/japaneseOmakase/menu/menu_how_to_play_item.webp";
import menuGameHistory from "@/assets/images/japaneseOmakase/menu/menu_game_history_item.webp";
import menuLogout from "@/assets/images/japaneseOmakase/menu/menu_logout_item.webp";
import iconSoundOn from "@/assets/images/japaneseOmakase/menu/icon_sound_on.webp";
import iconSoundOff from "@/assets/images/japaneseOmakase/menu/icon_sound_off.webp";
import type { SettingsMenuContentType } from "@/types/settingsMenu";
import RulesContent from "@/components/footer/RulesContent.vue";
import GuideContent from "@/components/footer/GuideContent.vue";
import HistoryContent from "@/components/footer/HistoryContent.vue";
import { useRouter } from "vue-router";
import { audioEvents, AUDIO_EVENTS } from "@/composables/audioEventBus";
import { howlerAudio } from "@/composables/useHowlerAudio";

const uiStore = useUIStore();
const authStore = useAuthStore();
const settingsStore = useSettingsStore();
const router = useRouter();

/**
 * Play generic UI sound with pitch randomization (0.6-1.4)
 */
function playGenericUISound(): void {
  const howl = howlerAudio.getHowl('generic_ui');
  if (!howl) return;

  // Apply pitch randomization 0.6-1.4
  const randomPitch = 0.6 + Math.random() * 0.8;
  howl.rate(randomPitch);

  audioEvents.emit(AUDIO_EVENTS.EFFECT_PLAY, { audioKey: 'generic_ui', volume: 0.6 });
}

const isOpen = computed(() => uiStore.isSettingsMenuOpen);
const activeContent = computed(() => uiStore.settingsMenuContent);
const gameSound = computed(() => settingsStore.gameSound);

// Menu button images mapping
const menuButtonImages: Record<string, string> = {
  sound: menuSound,
  paytable: menuPaytable,
  guides: menuHowToPlay,
  histories: menuGameHistory,
  logout: menuLogout,
};

const menuStyle = computed(() => ({
  backgroundImage: `url(${menuFrame})`,
}));

const menuButtons = [
  {
    id: "sound",
    label: "Sound Setting",
  },
  {
    id: "paytable",
    label: "Paytable",
  },
  {
    id: "guides",
    label: "Guides",
  },
  {
    id: "histories",
    label: "Game Histories",
  },
  {
    id: "logout",
    label: "Logout",
  },
];

function getButtonImage(buttonId: string): string {
  return menuButtonImages[buttonId] || "";
}

function handleClose(): void {
  // Play generic UI sound
  playGenericUISound();
  uiStore.closeSettings();
}

function handleButtonClick(buttonId: string): void {
  // Play generic UI sound for all button clicks
  playGenericUISound();

  if (buttonId === "logout") {
    // Handle logout
    authStore.logout().then(() => {
      handleClose();
      router.push("/login");
    });
  } else if (buttonId === "sound") {
    // Toggle sound settings
    settingsStore.toggleGameSound();
  } else {
    // Show content for paytable, guides, histories
    uiStore.setSettingsMenuContent(buttonId as SettingsMenuContentType);
  }
}

function closeContent(): void {
  // Play generic UI sound
  playGenericUISound();
  uiStore.clearSettingsMenuContent();
}
</script>

<style scoped lang="scss">
.settings-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.settings-menu {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  max-width: min(90vw, 600px);
  aspect-ratio: 820 / 1062;
  background-size: 100% 100%;
  background-position: center;
  background-repeat: no-repeat;

  @media (max-width: 768px) {
    max-width: 85vw;
  }

  @media (max-width: 480px) {
    max-width: 90vw;
  }
}

.menu-buttons-container {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;

  .sound-button {
    left: 0;
  }
}

.menu-button-wrapper {
  position: absolute;
  width: 60%;
  left: 20%;
  pointer-events: auto;

  &.menu-button--sound {
    top: 14.4%;
  }

  .menu-button {
    position: relative;
    width: 100%;
    height: auto;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover {
      transform: scale(1.05);
      filter: brightness(1.15);
    }

    &:active {
      transform: scale(0.98);
      filter: brightness(0.9);
    }
  }

  .sound-container-item:hover .sound-icon {
    transform: translateY(-50%) translateX(-10%) scale(1.1);
    filter: brightness(1.2);
  }

  // When hovering sound-icon, apply effect to menu-button
  .sound-container-item:has(.sound-icon:hover) .menu-button {
    transform: scale(1.05);
    filter: brightness(1.15);
  }

  .sound-icon {
    position: absolute;
    left: 7%;
    top: 50%;
    transform: translateY(-50%);
    width: 12%;
    height: auto;
    cursor: pointer;
    transition: all 0.2s ease;
    z-index: 5;

    &:hover {
      transform: translateY(-50%) scale(1.1);
      filter: brightness(1.2);
    }

    &:active {
      transform: translateY(-50%) scale(0.95);
    }
  }

  @media (max-width: 768px) {
    width: 62%;
    left: 19%;

    &.menu-button--sound {
      top: 14.2%;
    }

    .sound-icon {
      left: 7%;
      width: 14%;
    }
  }

  @media (max-width: 480px) {
    width: 64%;
    left: 18%;

    &.menu-button--sound {
      top: 14%;
    }

    .sound-icon {
      left: 6%;
      width: 16%;
    }
  }
}

.menu-button {
  position: absolute;
  cursor: pointer;
  pointer-events: auto;
  transition: all 0.2s ease;
  display: block;
  width: 60%;
  height: auto;
  left: 20%;

  &--paytable {
    top: 28.8%;
  }

  &--guides {
    top: 43.4%;
  }

  &--histories {
    top: 57.9%;
  }

  &--logout {
    top: 73%;
  }

  &:hover {
    transform: scale(1.05);
    filter: brightness(1.15);
  }

  &:active {
    transform: scale(0.98);
    filter: brightness(0.9);
  }

  @media (max-width: 768px) {
    width: 62%;
    left: 19%;

    &--paytable {
      top: 28.6%;
    }

    &--guides {
      top: 43.2%;
    }

    &--histories {
      top: 57.7%;
    }

    &--logout {
      top: 72.8%;
    }
  }

  @media (max-width: 480px) {
    width: 64%;
    left: 18%;

    &--paytable {
      top: 28.4%;
    }

    &--guides {
      top: 43%;
    }

    &--histories {
      top: 57.5%;
    }

    &--logout {
      top: 72.5%;
    }
  }
}

.menu-close-btn {
  position: absolute;
  top: 3%;
  right: 4%;
  width: 8%;
  max-width: 50px;
  min-width: 30px;
  height: auto;
  pointer-events: auto;
  transition: all 0.2s ease;
  z-index: 10;
  cursor: pointer;

  &:hover {
    transform: scale(1.1);
    filter: brightness(1.2);
  }

  &:active {
    transform: scale(0.95);
  }

  @media (max-width: 768px) {
    top: 2.5%;
    right: 3.5%;
    width: 10%;
  }

  @media (max-width: 480px) {
    top: 2%;
    right: 3%;
    width: 12%;
  }
}

// Content Layer
.content-layer {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1001;
}

.content-container {
  position: relative;
  background: rgba(20, 20, 20, 0.95);
  border-radius: 12px;
  overflow-y: auto;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
  height: 100vh;
  max-width: 768px;
  width: 100%;
  padding: 24px 32px;

  // Custom scrollbar
  &::-webkit-scrollbar {
    width: 8px;
  }

  &::-webkit-scrollbar-track {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
  }

  &::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.3);
    border-radius: 4px;

    &:hover {
      background: rgba(255, 255, 255, 0.5);
    }
  }
}

.content-close-btn {
  position: absolute;
  top: 1rem;
  right: 1rem;
  width: 2rem;
  height: 2rem;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  border: 2px solid rgba(255, 255, 255, 0.3);
  color: #fff;
  font-size: 1.2rem;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  padding-bottom: 2px;

  &:hover {
    background: rgba(255, 255, 255, 0.2);
    border-color: rgba(255, 255, 255, 0.5);
    transform: scale(1.1);
  }

  &:active {
    transform: scale(0.95);
  }
}

// Transition animations
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.fade-enter-active .settings-menu,
.fade-leave-active .settings-menu {
  transition: transform 0.3s ease;
}

.fade-enter-from .settings-menu {
  transform: scale(0.9);
}

.fade-leave-to .settings-menu {
  transform: scale(0.9);
}

// Content layer transitions
.slide-fade-enter-active,
.slide-fade-leave-active {
  transition: all 0.3s ease;
}

.slide-fade-enter-from {
  opacity: 0;
  transform: translateY(-20px);
}

.slide-fade-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

.slide-fade-enter-active .content-container,
.slide-fade-leave-active .content-container {
  transition: transform 0.3s ease;
}

.slide-fade-enter-from .content-container {
  transform: scale(0.95);
}

.slide-fade-leave-to .content-container {
  transform: scale(0.95);
}
</style>
