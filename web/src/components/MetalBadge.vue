<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    serviceId: string
    name?: string
    status?: 'available' | 'limited' | 'blocked' | 'untested' | 'unknown' | string
    region?: string
    quality?: string
    latencyMs?: number
    size?: 'sm' | 'md' | 'lg' | 'pill'
    showLabel?: boolean
    interactive?: boolean
  }>(),
  {
    name: '',
    status: 'untested',
    region: '',
    quality: '',
    latencyMs: 0,
    size: 'md',
    showLabel: true,
    interactive: true,
  }
)

const emit = defineEmits<{
  (e: 'click', serviceId: string): void
}>()

const statusNormalized = computed(() => {
  const s = props.status?.toLowerCase() || 'untested'
  if (['available', 'unlocked', 'good', 'online'].includes(s)) return 'available'
  if (['limited', 'warn', 'warning'].includes(s)) return 'limited'
  if (['blocked', 'danger', 'offline', 'failed'].includes(s)) return 'blocked'
  return 'untested'
})

const statusLabel = computed(() => {
  switch (statusNormalized.value) {
    case 'available':
      return '解锁'
    case 'limited':
      return '受限'
    case 'blocked':
      return '封锁'
    default:
      return '未测'
  }
})
</script>

<template>
  <div
    class="metal-badge"
    :class="[
      `badge-${size}`,
      `status-${statusNormalized}`,
      { 'is-interactive': interactive },
    ]"
    :title="`${name || serviceId}: ${statusLabel} ${region ? `(${region})` : ''} ${latencyMs ? `· ${latencyMs}ms` : ''}`"
    @click="interactive && emit('click', serviceId)"
  >
    <!-- Metallic Beveled Rim Background -->
    <div class="metal-foil"></div>
    <div class="metal-light-sweep"></div>

    <!-- Official Brand Logo Container -->
    <div class="brand-logo-wrap">
      <!-- Netflix -->
      <svg v-if="serviceId === 'netflix'" viewBox="0 0 24 24" class="brand-svg svg-netflix">
        <path fill="#E50914" d="M5.398 0v24c2.816-.364 5.372-.94 7.644-1.704V0H5.398zm7.644 0v14.47c1.86-.484 3.732-1.128 5.56-1.928V0h-5.56zM0 0v24c1.8-.236 3.6-.54 5.398-.908V0H0z"/>
      </svg>

      <!-- Disney+ -->
      <svg v-else-if="serviceId === 'disney'" viewBox="0 0 24 24" class="brand-svg svg-disney">
        <path fill="#113CCF" d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.894 13.914c-1.32.748-3.08 1.188-5.068 1.188-4.474 0-6.82-2.182-6.82-5.102 0-2.81 2.222-4.908 5.76-4.908 3.498 0 5.578 1.956 5.578 4.636 0 .426-.048.832-.14 1.218-.846-.436-1.898-.684-3.036-.684-2.88 0-4.464 1.488-4.464 3.238 0 1.258.85 2.15 2.378 2.15 1.764 0 3.328-.84 4.184-2.128.84.192 1.47.392 1.628.392z"/>
      </svg>

      <!-- YouTube Premium -->
      <svg v-else-if="serviceId === 'youtube'" viewBox="0 0 24 24" class="brand-svg svg-youtube">
        <path fill="#FF0000" d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"/>
      </svg>

      <!-- Amazon Prime Video -->
      <svg v-else-if="serviceId === 'prime'" viewBox="0 0 24 24" class="brand-svg svg-prime">
        <path fill="#00A8E1" d="M22.5 16.5c-3.5 2.5-8.5 3.5-13.5 2-1-.3-1.8-.7-2.5-1.2-.3-.2-.1-.6.2-.5 4.5 1.5 9.8.7 13.8-1.5.5-.3.9.6.4.9l1.6.3zM23.7 14.8c-.3-.4-1.9-.2-2.6 0-.2 0-.3-.2-.1-.4.9-1.2 2.4-.6 2.8-.1.4.5-.1 2.3-1 3.4-.2.2-.4.1-.3-.1.4-.7.6-2.2.2-2.8h1zM11.8 4C7.5 4 4 7.5 4 11.8c0 2.2.9 4.2 2.3 5.6 1.4-1.5 3.3-2.6 5.5-2.6 4.3 0 7.8-3.5 7.8-7.8S16.1 4 11.8 4z"/>
      </svg>

      <!-- Max (HBO) -->
      <svg v-else-if="serviceId === 'max'" viewBox="0 0 24 24" class="brand-svg svg-max">
        <path fill="#002BE7" d="M0 4h5.2l3.4 8.5L12 4h5.2v16h-4.4V9.8l-3.3 8.2H8.3L5 9.8V20H0V4zm19.6 0H24v16h-4.4V4z"/>
      </svg>

      <!-- Spotify -->
      <svg v-else-if="serviceId === 'spotify'" viewBox="0 0 24 24" class="brand-svg svg-spotify">
        <path fill="#1DB954" d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.49 17.305c-.215.352-.676.463-1.028.247-2.818-1.722-6.365-2.112-10.543-1.157-.402.092-.8-.16-.893-.563-.092-.402.16-.8.563-.893 4.576-1.045 8.498-.598 11.654 1.338.352.216.463.677.247 1.028zm1.464-3.255c-.27.44-.848.578-1.287.308-3.226-1.983-8.143-2.557-11.958-1.4-.495.15-1.023-.133-1.173-.628-.15-.495.133-1.023.628-1.173 4.364-1.324 9.774-.683 13.482 1.606.44.27.578.848.308 1.287zm.126-3.41c-3.868-2.297-10.248-2.508-13.94-1.387-.593.18-1.223-.158-1.403-.75-.18-.593.158-1.223.75-1.403 4.24-1.288 11.284-1.042 15.74 1.604.533.316.708 1.008.392 1.541-.316.533-1.008.708-1.541.392l.002.003z"/>
      </svg>

      <!-- Hulu -->
      <svg v-else-if="serviceId === 'hulu'" viewBox="0 0 24 24" class="brand-svg svg-hulu">
        <path fill="#1CE783" d="M6.2 3H2v18h4.2v-6.3h4.6V21H15V3h-4.2v6.4H6.2V3zm11.6 6.3h4.2V21h-4.2V9.3z"/>
      </svg>

      <!-- Bahamut 巴哈姆特动画疯 -->
      <svg v-else-if="serviceId === 'bahamut'" viewBox="0 0 24 24" class="brand-svg svg-bahamut">
        <path fill="#00B4D8" d="M12 2L2 7l10 5 10-5-10-5zm0 8.5L4.5 7 12 3.2 19.5 7 12 10.5zM2 17l10 5 10-5v-3.5L12 18.5 2 13.5V17zm0-5l10 5 10-5v-3.5L12 13.5 2 8.5V12z"/>
      </svg>

      <!-- AbemaTV -->
      <svg v-else-if="serviceId === 'abema'" viewBox="0 0 24 24" class="brand-svg svg-abema">
        <path fill="#22C55E" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 14.5v-9l7 4.5-7 4.5z"/>
      </svg>

      <!-- TikTok -->
      <svg v-else-if="serviceId === 'tiktok'" viewBox="0 0 24 24" class="brand-svg svg-tiktok">
        <path fill="#25F4EE" d="M19.589 6.686a4.793 4.793 0 0 1-3.77-4.245V2h-3.445v13.672a2.896 2.896 0 0 1-2.887 2.87 2.896 2.896 0 0 1-2.887-2.87 2.896 2.896 0 0 1 2.887-2.87c.307 0 .604.045.885.129V9.45a6.34 6.34 0 0 0-.885-.062 6.343 6.343 0 0 0-6.344 6.344 6.343 6.343 0 0 0 6.344 6.344 6.343 6.343 0 0 0 6.344-6.344V8.995a8.21 8.21 0 0 0 4.887 1.6v-3.445a4.814 4.814 0 0 1-1.229-.464z"/>
        <path fill="#FE2C55" d="M19.589 6.686v-1a4.814 4.814 0 0 1-1.229-.464A4.793 4.793 0 0 1 14.59 1h-1.445v14.672a2.896 2.896 0 0 1-2.887 2.87 2.896 2.896 0 0 1-2.887-2.87 2.896 2.896 0 0 1 2.887-2.87c.307 0 .604.045.885.129V9.45a6.34 6.34 0 0 0-.885-.062 6.343 6.343 0 0 0-6.344 6.344 6.343 6.343 0 0 0 6.344 6.344 6.343 6.343 0 0 0 6.344-6.344V8.995a8.21 8.21 0 0 0 4.887 1.6v-3.445a4.814 4.814 0 0 1-1.229-.464z" opacity="0.75"/>
      </svg>

      <!-- DAZN -->
      <svg v-else-if="serviceId === 'dazn'" viewBox="0 0 24 24" class="brand-svg svg-dazn">
        <path fill="#FACC15" d="M3 4h6v16H3V4zm9 0h6l3 8-3 8h-6l3-8-3-8z"/>
      </svg>

      <!-- ChatGPT (OpenAI) -->
      <svg v-else-if="serviceId === 'chatgpt'" viewBox="0 0 24 24" class="brand-svg svg-chatgpt">
        <path fill="#10A37F" d="M22.282 9.821a5.985 5.985 0 0 0-.516-4.91 6.046 6.046 0 0 0-6.51-2.9A6.065 6.065 0 0 0 4.98 4.181a5.984 5.984 0 0 0-3.998 3.778 6.046 6.046 0 0 0 .743 7.097 5.98 5.98 0 0 0 .51 4.911 6.051 6.051 0 0 0 6.515 2.9A5.985 5.985 0 0 0 13.26 24a6.056 6.056 0 0 0 5.772-4.206 5.99 5.99 0 0 0 3.997-3.778 6.057 6.057 0 0 0-.747-6.195zM13.26 22.43a4.476 4.476 0 0 1-2.876-1.04l.141-.08 4.779-2.758a.795.795 0 0 0 .392-.681v-6.737l2.02 1.168a.071.071 0 0 1 .038.052v5.583a4.504 4.504 0 0 1-4.494 4.493zm-8.88-4.27a4.473 4.473 0 0 1-.535-3.014l.142.085 4.783 2.759a.771.771 0 0 0 .78 0l5.843-3.369v2.332a.08.08 0 0 1-.033.062L10.5 19.49a4.5 4.5 0 0 1-6.12-1.33zm-1.04-8.878A4.477 4.477 0 0 1 5.68 6.94l-.003.16v5.52a.792.792 0 0 0 .392.681l5.842 3.368-2.02 1.168a.076.076 0 0 1-.071 0l-4.839-2.793a4.504 4.504 0 0 1-1.64-5.762zm15.426 3.205l-5.843-3.37 2.02-1.164a.08.08 0 0 1 .071 0l4.839 2.791a4.496 4.496 0 0 1-.692 8.1l-.395-.228v-5.448a.792.792 0 0 0-.392-.681zm2.02-3.372a4.473 4.473 0 0 1-.535 3.014l-.142-.085-4.779-2.76a.775.775 0 0 0-.784 0l-5.84 3.369v-2.332a.08.08 0 0 1 .033-.062L13.5 4.51a4.5 4.5 0 0 1 6.286 1.716zm-7.986 4.67l-2.67-1.541 2.67-1.542 2.67 1.542-2.67 1.541z"/>
      </svg>

      <!-- Claude (Anthropic) -->
      <svg v-else-if="serviceId === 'claude'" viewBox="0 0 24 24" class="brand-svg svg-claude">
        <path fill="#D97706" d="M12 2l2.4 6.6L21 11l-6.6 2.4L12 20l-2.4-6.6L3 11l6.6-2.4L12 2zm0 4.2L10.7 10 7 11.3l3.7 1.3L12 16.3l1.3-3.7 3.7-1.3-3.7-1.3L12 6.2z"/>
      </svg>

      <!-- Gemini (Google AI) -->
      <svg v-else-if="serviceId === 'gemini'" viewBox="0 0 24 24" class="brand-svg svg-gemini">
        <path fill="#38BDF8" d="M12 0C12 6.627 6.627 12 0 12c6.627 0 12 5.373 12 12 0-6.627 5.373-12 12-12-6.627 0-12-5.373-12-12z"/>
      </svg>

      <!-- DeepSeek -->
      <svg v-else-if="serviceId === 'deepseek'" viewBox="0 0 24 24" class="brand-svg svg-deepseek">
        <path fill="#0284C7" d="M12 3c-4.97 0-9 4.03-9 9 0 2.12.74 4.07 1.97 5.61L3.5 21l3.6-1.37C8.58 20.4 10.22 21 12 21c4.97 0 9-4.03 9-9s-4.03-9-9-9zm-1 14h-2v-2h2v2zm4 0h-2v-2h2v2zm2-5H7V9h10v3z"/>
      </svg>

      <!-- Midjourney -->
      <svg v-else-if="serviceId === 'midjourney'" viewBox="0 0 24 24" class="brand-svg svg-midjourney">
        <path fill="#E0E7FF" d="M12 2L4 18h6l2-4 2 4h6L12 2zm0 5l4.5 9h-9L12 7z"/>
      </svg>

      <!-- Microsoft Copilot -->
      <svg v-else-if="serviceId === 'copilot'" viewBox="0 0 24 24" class="brand-svg svg-copilot">
        <path fill="#A855F7" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z"/>
      </svg>

      <!-- Grok (xAI) -->
      <svg v-else-if="serviceId === 'grok'" viewBox="0 0 24 24" class="brand-svg svg-grok">
        <path fill="#F8FAFC" d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
      </svg>

      <!-- Perplexity AI -->
      <svg v-else-if="serviceId === 'perplexity'" viewBox="0 0 24 24" class="brand-svg svg-perplexity">
        <path fill="#2DD4BF" d="M12 2v20M2 12h20M4.93 4.93l14.14 14.14M4.93 19.07L19.07 4.93" stroke="#2DD4BF" stroke-width="2.5" stroke-linecap="round"/>
      </svg>

      <!-- GitHub Copilot -->
      <svg v-else-if="serviceId === 'github_cop'" viewBox="0 0 24 24" class="brand-svg svg-github">
        <path fill="#F8FAFC" d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
      </svg>

      <!-- Generic Fallback Badge -->
      <div v-else class="brand-generic-char">
        {{ (name || serviceId).slice(0, 2).toUpperCase() }}
      </div>
    </div>

    <!-- Metadata / Status Capsule -->
    <div v-if="showLabel" class="metal-badge-meta">
      <div class="meta-top-row">
        <span class="badge-title">{{ name || serviceId }}</span>
        <span v-if="region" class="badge-region-tag">{{ region }}</span>
      </div>
      <div class="meta-bottom-row">
        <span class="status-gem"></span>
        <span class="status-desc">{{ quality || statusLabel }}</span>
        <span v-if="latencyMs > 0" class="badge-latency">{{ latencyMs }}ms</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.metal-badge {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 5px 10px;
  background: linear-gradient(145deg, #1b222c, #0e1319);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 8px;
  box-shadow:
    0 4px 12px rgba(0, 0, 0, 0.4),
    inset 0 1px 0 rgba(255, 255, 255, 0.15),
    inset 0 -1px 0 rgba(0, 0, 0, 0.4);
  transition: all 0.22s cubic-bezier(0.16, 1, 0.3, 1);
  overflow: hidden;
  user-select: none;
}

.metal-badge.is-interactive {
  cursor: pointer;
}
.metal-badge.is-interactive:hover {
  transform: translateY(-2px);
  border-color: rgba(56, 189, 248, 0.4);
  box-shadow:
    0 8px 20px rgba(0, 0, 0, 0.6),
    0 0 12px rgba(56, 189, 248, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.25);
}

/* Foil and Light Sweeps */
.metal-foil {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 50% 0%, rgba(255, 255, 255, 0.06), transparent 70%);
  pointer-events: none;
}
.metal-light-sweep {
  position: absolute;
  top: 0;
  left: -100%;
  width: 60%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.08), transparent);
  transform: skewX(-20deg);
  transition: left 0.5s ease;
  pointer-events: none;
}
.metal-badge:hover .metal-light-sweep {
  left: 140%;
}

/* Brand Logo */
.brand-logo-wrap {
  width: 26px;
  height: 26px;
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.08);
  display: grid;
  place-items: center;
  flex-shrink: 0;
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.6);
}
.brand-svg {
  width: 17px;
  height: 17px;
}
.brand-generic-char {
  font-family: 'Fira Code', monospace;
  font-size: 11px;
  font-weight: 700;
  color: #94a3b8;
}

/* Sizes */
.badge-sm {
  padding: 3px 6px;
  gap: 5px;
}
.badge-sm .brand-logo-wrap {
  width: 18px;
  height: 18px;
}
.badge-sm .brand-svg {
  width: 12px;
  height: 12px;
}
.badge-sm .badge-title {
  font-size: 10px;
}
.badge-sm .status-desc {
  font-size: 8.5px;
}

.badge-lg {
  padding: 8px 14px;
  gap: 12px;
}
.badge-lg .brand-logo-wrap {
  width: 34px;
  height: 34px;
}
.badge-lg .brand-svg {
  width: 22px;
  height: 22px;
}
.badge-lg .badge-title {
  font-size: 13px;
}

.badge-pill {
  padding: 4px 8px;
  border-radius: 20px;
}

/* Status Jewels and Colors */
.status-gem {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-available .status-gem {
  background: #10b981;
  box-shadow: 0 0 8px #10b981;
}
.status-available {
  border-color: rgba(16, 185, 129, 0.3);
}

.status-limited .status-gem {
  background: #f59e0b;
  box-shadow: 0 0 8px #f59e0b;
}
.status-limited {
  border-color: rgba(245, 158, 11, 0.3);
}

.status-blocked .status-gem {
  background: #ef4444;
  box-shadow: 0 0 8px #ef4444;
}
.status-blocked {
  border-color: rgba(239, 68, 68, 0.3);
  opacity: 0.85;
}

.status-untested .status-gem {
  background: #64748b;
}
.status-untested {
  border-style: dashed;
  opacity: 0.6;
}

/* Metadata */
.metal-badge-meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.meta-top-row {
  display: flex;
  align-items: center;
  gap: 4px;
}
.badge-title {
  font-size: 11.5px;
  font-weight: 600;
  color: #f8fafc;
  white-space: nowrap;
}
.badge-region-tag {
  font-family: 'Fira Code', monospace;
  font-size: 8.5px;
  font-weight: 700;
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
  padding: 0 3px;
  border-radius: 3px;
  border: 1px solid rgba(56, 189, 248, 0.3);
}

.meta-bottom-row {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 9.5px;
  color: #94a3b8;
  white-space: nowrap;
}
.status-desc {
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 95px;
}
.badge-latency {
  font-family: 'Fira Code', monospace;
  font-size: 8.5px;
  color: #38bdf8;
  background: rgba(0, 0, 0, 0.3);
  padding: 0 3px;
  border-radius: 2px;
}
</style>
