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
    nodeRegion?: string
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
    nodeRegion: '',
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
  if (['available', 'unlocked', 'good', 'online', 'yes', '原生', '解锁'].includes(s)) return 'available'
  if (['limited', 'warn', 'warning', '部分解锁', '仅自制'].includes(s)) return 'limited'
  if (['blocked', 'danger', 'offline', 'failed', 'no', '封锁', '未解锁'].includes(s)) return 'blocked'
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

const isYouTubeCN = computed(() =>
  props.serviceId === 'youtube' &&
  (props.region?.toUpperCase() === 'CN' || props.quality?.includes('送中')),
)
const regionMatchesNode = computed(() =>
  Boolean(props.region && props.nodeRegion) && props.region.trim().toUpperCase() === props.nodeRegion.trim().toUpperCase(),
)
</script>

<template>
  <div
    class="metal-badge"
    :class="[
      `badge-${size}`,
      `status-${statusNormalized}`,
      { 'is-interactive': interactive, 'has-cn-route': isYouTubeCN },
    ]"
    :title="`${name || serviceId}: ${quality || statusLabel} ${region ? `(${region})` : ''} ${latencyMs ? `· ${latencyMs}ms` : ''}`"
    :role="interactive ? 'button' : undefined"
    :tabindex="interactive ? 0 : undefined"
    :aria-label="interactive ? `${name || serviceId}：${quality || statusLabel}` : undefined"
    @click="interactive && emit('click', serviceId)"
    @keydown.enter.prevent="interactive && emit('click', serviceId)"
    @keydown.space.prevent="interactive && emit('click', serviceId)"
  >
    <!-- Metallic Beveled Rim Background -->
    <div class="metal-foil"></div>
    <div class="metal-light-sweep"></div>

    <!-- Official Brand Logo Container -->
    <div class="brand-logo-wrap" :class="`brand-logo-${statusNormalized}`">
      <!-- 1. Netflix Official Red N Ribbon -->
      <svg v-if="serviceId === 'netflix'" viewBox="0 0 24 24" class="brand-svg svg-netflix">
        <path fill="#E50914" d="M4 2.5v19h4.2V8.4l5.3 13.1H18V2.5h-4.2v13.1L8.5 2.5H4z"/>
      </svg>

      <!-- 2. Disney+ Official Star Arc & Plus -->
      <svg v-else-if="serviceId === 'disney'" viewBox="0 0 24 24" class="brand-svg svg-disney">
        <path fill="#113CCF" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.8 11.8c-1.1.6-2.5 1-4.2 1-3.6 0-5.5-1.8-5.5-4.2 0-2.3 1.8-4 4.7-4 2.8 0 4.5 1.6 4.5 3.8 0 .4 0 .7-.1 1-.7-.4-1.6-.6-2.5-.6-2.4 0-3.7 1.2-3.7 2.7 0 1 .7 1.8 1.9 1.8 1.4 0 2.7-.7 3.4-1.7.7.1 1.2.3 1.5.2v-.2z"/>
        <path fill="#00D2FF" d="M19 8.5v3h3v1.5h-3v3h-1.5v-3h-3V11.5h3v-3H19z"/>
      </svg>

      <!-- 3. YouTube Premium Official Red Badge -->
      <svg v-else-if="serviceId === 'youtube'" viewBox="0 0 24 24" class="brand-svg svg-youtube">
        <path fill="#FF0000" d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814z"/>
        <polygon fill="#FFFFFF" points="9.75,15.02 15.75,12 9.75,8.98"/>
      </svg>

      <!-- 4. Amazon Prime Video Official Smile Arrow -->
      <svg v-else-if="serviceId === 'prime'" viewBox="0 0 24 24" class="brand-svg svg-prime">
        <path fill="#00A8E1" d="M21.5 15.8c-3.1 2.2-7.6 3.1-12.1 1.8-.9-.3-1.6-.6-2.2-1.1-.3-.2-.1-.5.2-.4 4 1.3 8.7.6 12.3-1.3.4-.2.8.5.3.8l1.5.2zm1-1.5c-.3-.4-1.7-.2-2.3 0-.2 0-.3-.2-.1-.4.8-1.1 2.1-.5 2.5-.1.4.4-.1 2-1 3-.2.2-.4.1-.3-.1.4-.6.5-1.9.2-2.4h1z"/>
        <path fill="#00A8E1" d="M11 4C7.1 4 4 7.1 4 11c0 2 .8 3.8 2.1 5.1 1.3-1.4 3-2.4 4.9-2.4 3.9 0 7-3.1 7-7S14.9 4 11 4z"/>
      </svg>

      <!-- 5. Max (HBO) Official Blue Geometry -->
      <svg v-else-if="serviceId === 'max'" viewBox="0 0 24 24" class="brand-svg svg-max">
        <path fill="#002BE7" d="M2 4h4.5l3.2 7.8L13 4h4.5v16h-3.8v-8.5l-3.1 7.5H8.4L5.3 11.5V20H2V4zm17 0h4v16h-4V4z"/>
      </svg>

      <!-- 6. Spotify Official Emerald Sound Waves -->
      <svg v-else-if="serviceId === 'spotify'" viewBox="0 0 24 24" class="brand-svg svg-spotify">
        <path fill="#1DB954" d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.49 17.305c-.215.352-.676.463-1.028.247-2.818-1.722-6.365-2.112-10.543-1.157-.402.092-.8-.16-.893-.563-.092-.402.16-.8.563-.893 4.576-1.045 8.498-.598 11.654 1.338.352.216.463.677.247 1.028zm1.464-3.255c-.27.44-.848.578-1.287.308-3.226-1.983-8.143-2.557-11.958-1.4-.495.15-1.023-.133-1.173-.628-.15-.495.133-1.023.628-1.173 4.364-1.324 9.774-.683 13.482 1.606.44.27.578.848.308 1.287zm.126-3.41c-3.868-2.297-10.248-2.508-13.94-1.387-.593.18-1.223-.158-1.403-.75-.18-.593.158-1.223.75-1.403 4.24-1.288 11.284-1.042 15.74 1.604.533.316.708 1.008.392 1.541-.316.533-1.008.708-1.541.392l.002.003z"/>
      </svg>

      <!-- 7. Hulu Official Mint Green -->
      <svg v-else-if="serviceId === 'hulu'" viewBox="0 0 24 24" class="brand-svg svg-hulu">
        <path fill="#1CE783" d="M6 3H2v18h4v-6h5v6h4V3h-4v6H6V3zm12 6h4v12h-4V9z"/>
      </svg>

      <!-- 8. Bahamut (巴哈姆特动画疯) Crest -->
      <svg v-else-if="serviceId === 'bahamut'" viewBox="0 0 24 24" class="brand-svg svg-bahamut">
        <path fill="#00B4D8" d="M12 2L2 7l10 5 10-5-10-5zm0 8.5L4.5 7 12 3.2 19.5 7 12 10.5zM2 17l10 5 10-5v-3.5L12 18.5 2 13.5V17zm0-5l10 5 10-5v-3.5L12 13.5 2 8.5V12z"/>
      </svg>

      <!-- 9. AbemaTV Official Green -->
      <svg v-else-if="serviceId === 'abema'" viewBox="0 0 24 24" class="brand-svg svg-abema">
        <path fill="#22C55E" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 14.5v-9l7 4.5-7 4.5z"/>
      </svg>

      <!-- 10. TikTok Official Cyan & Crimson Duo -->
      <svg v-else-if="serviceId === 'tiktok'" viewBox="0 0 24 24" class="brand-svg svg-tiktok">
        <path fill="#25F4EE" d="M19.589 6.686a4.793 4.793 0 0 1-3.77-4.245V2h-3.445v13.672a2.896 2.896 0 0 1-2.887 2.87 2.896 2.896 0 0 1-2.887-2.87 2.896 2.896 0 0 1 2.887-2.87c.307 0 .604.045.885.129V9.45a6.34 6.34 0 0 0-.885-.062 6.343 6.343 0 0 0-6.344 6.344 6.343 6.343 0 0 0 6.344 6.344 6.343 6.343 0 0 0 6.344-6.344V8.995a8.21 8.21 0 0 0 4.887 1.6v-3.445a4.814 4.814 0 0 1-1.229-.464z"/>
        <path fill="#FE2C55" d="M19.589 6.686v-1a4.814 4.814 0 0 1-1.229-.464A4.793 4.793 0 0 1 14.59 1h-1.445v14.672a2.896 2.896 0 0 1-2.887 2.87 2.896 2.896 0 0 1-2.887-2.87 2.896 2.896 0 0 1 2.887-2.87c.307 0 .604.045.885.129V9.45a6.34 6.34 0 0 0-.885-.062 6.343 6.343 0 0 0-6.344 6.344 6.343 6.343 0 0 0 6.344 6.344 6.343 6.343 0 0 0 6.344-6.344V8.995a8.21 8.21 0 0 0 4.887 1.6v-3.445a4.814 4.814 0 0 1-1.229-.464z" opacity="0.8"/>
      </svg>

      <!-- 11. DAZN Official Yellow -->
      <svg v-else-if="serviceId === 'dazn'" viewBox="0 0 24 24" class="brand-svg svg-dazn">
        <path fill="#FACC15" d="M3 4h6v16H3V4zm9 0h6l3 8-3 8h-6l3-8-3-8z"/>
      </svg>

      <!-- 12. Apple TV+ -->
      <svg v-else-if="serviceId === 'appletv'" viewBox="0 0 24 24" class="brand-svg svg-apple">
        <path fill="#F8FAFC" d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M15.97 6.35c.62-.75 1.04-1.8 0.93-2.85-.9.04-1.99.6-2.63 1.35-.57.65-1.06 1.72-.93 2.74 1.01.08 2.02-.49 2.63-1.24z"/>
      </svg>

      <!-- 13. Bilibili (哔哩哔哩) -->
      <svg v-else-if="serviceId === 'bilibili'" viewBox="0 0 24 24" class="brand-svg svg-bilibili">
        <path fill="#00AEEC" d="M17.8 2.8l2.2 2.2-3.5 3.5h3c1.4 0 2.5 1.1 2.5 2.5v9c0 1.4-1.1 2.5-2.5 2.5H4.5C3.1 22.5 2 21.4 2 20V11c0-1.4 1.1-2.5 2.5-2.5h3L4 5l2.2-2.2L10.2 6.8h3.6L17.8 2.8zM8 13.5a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3zm8 0a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3z"/>
      </svg>

      <!-- 14. ChatGPT (OpenAI) Official Emerald Spiral -->
      <svg v-else-if="serviceId === 'chatgpt'" viewBox="0 0 24 24" class="brand-svg svg-chatgpt">
        <path fill="#10A37F" d="M22.282 9.821a5.985 5.985 0 0 0-.516-4.91 6.046 6.046 0 0 0-6.51-2.9A6.065 6.065 0 0 0 4.98 4.181a5.984 5.984 0 0 0-3.998 3.778 6.046 6.046 0 0 0 .743 7.097 5.98 5.98 0 0 0 .51 4.911 6.051 6.051 0 0 0 6.515 2.9A5.985 5.985 0 0 0 13.26 24a6.056 6.056 0 0 0 5.772-4.206 5.99 5.99 0 0 0 3.997-3.778 6.057 6.057 0 0 0-.747-6.195zM13.26 22.43a4.476 4.476 0 0 1-2.876-1.04l.141-.08 4.779-2.758a.795.795 0 0 0 .392-.681v-6.737l2.02 1.168a.071.071 0 0 1 .038.052v5.583a4.504 4.504 0 0 1-4.494 4.493zm-8.88-4.27a4.473 4.473 0 0 1-.535-3.014l.142.085 4.783 2.759a.771.771 0 0 0 .78 0l5.843-3.369v2.332a.08.08 0 0 1-.033.062L10.5 19.49a4.5 4.5 0 0 1-6.12-1.33zm-1.04-8.878A4.477 4.477 0 0 1 5.68 6.94l-.003.16v5.52a.792.792 0 0 0 .392.681l5.842 3.368-2.02 1.168a.076.076 0 0 1-.071 0l-4.839-2.793a4.504 4.504 0 0 1-1.64-5.762zm15.426 3.205l-5.843-3.37 2.02-1.164a.08.08 0 0 1 .071 0l4.839 2.791a4.496 4.496 0 0 1-.692 8.1l-.395-.228v-5.448a.792.792 0 0 0-.392-.681zm2.02-3.372a4.473 4.473 0 0 1-.535 3.014l-.142-.085-4.779-2.76a.775.775 0 0 0-.784 0l-5.84 3.369v-2.332a.08.08 0 0 1 .033-.062L13.5 4.51a4.5 4.5 0 0 1 6.286 1.716zm-7.986 4.67l-2.67-1.541 2.67-1.542 2.67 1.542-2.67 1.541z"/>
      </svg>

      <!-- 15. Claude (Anthropic) Official Terracotta Spark -->
      <svg v-else-if="serviceId === 'claude'" viewBox="0 0 24 24" class="brand-svg svg-claude">
        <path fill="#D97706" d="M12 2l2.4 6.6L21 11l-6.6 2.4L12 20l-2.4-6.6L3 11l6.6-2.4L12 2zm0 4.2L10.7 10 7 11.3l3.7 1.3L12 16.3l1.3-3.7 3.7-1.3-3.7-1.3L12 6.2z"/>
      </svg>

      <!-- 16. Gemini (Google AI) Quad Sparkle -->
      <svg v-else-if="serviceId === 'gemini'" viewBox="0 0 24 24" class="brand-svg svg-gemini">
        <path fill="#38BDF8" d="M12 0C12 6.627 6.627 12 0 12c6.627 0 12 5.373 12 12 0-6.627 5.373-12 12-12-6.627 0-12-5.373-12-12z"/>
      </svg>

      <!-- 17. DeepSeek Official Whale & Wave -->
      <svg v-else-if="serviceId === 'deepseek'" viewBox="0 0 24 24" class="brand-svg svg-deepseek">
        <path fill="#0284C7" d="M12 3c-4.97 0-9 4.03-9 9 0 2.12.74 4.07 1.97 5.61L3.5 21l3.6-1.37C8.58 20.4 10.22 21 12 21c4.97 0 9-4.03 9-9s-4.03-9-9-9zm-1 14h-2v-2h2v2zm4 0h-2v-2h2v2zm2-5H7V9h10v3z"/>
      </svg>

      <!-- 18. Midjourney Sailboat Star -->
      <svg v-else-if="serviceId === 'midjourney'" viewBox="0 0 24 24" class="brand-svg svg-midjourney">
        <path fill="#818CF8" d="M12 2L4 18h6l2-4 2 4h6L12 2zm0 5l4.5 9h-9L12 7z"/>
      </svg>

      <!-- 19. Microsoft Copilot -->
      <svg v-else-if="serviceId === 'copilot'" viewBox="0 0 24 24" class="brand-svg svg-copilot">
        <path fill="#A855F7" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z"/>
      </svg>

      <!-- 20. Grok (xAI) Official X -->
      <svg v-else-if="serviceId === 'grok'" viewBox="0 0 24 24" class="brand-svg svg-grok">
        <path fill="#F8FAFC" d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
      </svg>

      <!-- 21. Perplexity AI -->
      <svg v-else-if="serviceId === 'perplexity'" viewBox="0 0 24 24" class="brand-svg svg-perplexity">
        <path fill="#2DD4BF" d="M12 2v20M2 12h20M4.93 4.93l14.14 14.14M4.93 19.07L19.07 4.93" stroke="#2DD4BF" stroke-width="2.5" stroke-linecap="round"/>
      </svg>

      <!-- 22. GitHub Copilot -->
      <svg v-else-if="serviceId === 'github_cop'" viewBox="0 0 24 24" class="brand-svg svg-github">
        <path fill="#F8FAFC" d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
      </svg>

      <!-- 23. Reddit Official Snoo Face -->
      <svg v-else-if="serviceId === 'reddit'" viewBox="0 0 24 24" class="brand-svg svg-reddit">
        <path fill="#FF4500" d="M12 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0zm5.01 4.744c.688 0 1.25.561 1.25 1.249a1.25 1.25 0 0 1-2.498.056l-2.597-.547-.8 3.747c1.824.07 3.48.632 4.674 1.488.308-.309.73-.491 1.207-.491.968 0 1.754.786 1.754 1.754 0 .716-.435 1.333-1.01 1.614a3.111 3.111 0 0 1 .042.52c0 2.694-3.13 4.87-7.004 4.87-3.874 0-7.004-2.176-7.004-4.87 0-.183.015-.366.043-.534A1.748 1.748 0 0 1 4.028 12c0-.968.786-1.754 1.754-1.754.463 0 .898.196 1.207.49 1.207-.883 2.878-1.43 4.744-1.487l.885-4.182a.342.342 0 0 1 .14-.197.35.35 0 0 1 .238-.042l2.906.617a1.214 1.214 0 0 1 1.108-.701zM9.25 12C8.56 12 8 12.56 8 13.25c0 .688.56 1.25 1.25 1.25.688 0 1.25-.562 1.25-1.25 0-.69-.562-1.25-1.25-1.25zm5.5 0c-.69 0-1.25.56-1.25 1.25 0 .688.56 1.25 1.25 1.25.688 0 1.25-.562 1.25-1.25 0-.69-.562-1.25-1.25-1.25zm-5.465 4.382a.498.498 0 0 0-.083.7.502.502 0 0 0 .7.084c.895-.623 2.148-.966 2.098-.966.05 0 1.203.343 2.098.966a.5.5 0 0 0 .7-.084.498.498 0 0 0-.083-.7c-1.077-.751-2.45-1.157-2.715-1.157-.266 0-1.638.406-2.715 1.157z"/>
      </svg>

      <!-- Generic Fallback Badge -->
      <div v-else class="brand-generic-char">
        {{ (name || serviceId).slice(0, 2).toUpperCase() }}
      </div>
      <span v-if="isYouTubeCN" class="compact-cn-tag">送中</span>
    </div>
    <span v-if="!showLabel" class="compact-status-dot" aria-hidden="true"></span>

    <!-- Metadata / Status Capsule -->
    <div v-if="showLabel" class="metal-badge-meta">
      <div class="meta-top-row">
        <span class="badge-title">{{ name || serviceId }}</span>
        <span v-if="region" class="badge-region-tag" :class="{ 'is-local': regionMatchesNode }">{{ region }}</span>
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
  border-radius: 8px;
  background: rgba(7, 10, 13, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.07);
  box-shadow: inset 2px 0 0 var(--status-accent, rgba(100, 116, 139, 0.45));
  transition: transform 0.18s ease, border-color 0.18s ease, background 0.18s ease;
  overflow: hidden;
  user-select: none;
}

.metal-badge.is-interactive {
  cursor: pointer;
}
.metal-badge.is-interactive:focus-visible {
  outline: 2px solid rgba(125, 211, 252, 0.8);
  outline-offset: 2px;
}
.metal-badge.is-interactive:hover {
  transform: translateY(-1px);
  border-color: rgba(255, 255, 255, 0.14);
  background: rgba(10, 14, 18, 0.64);
}
.metal-badge.has-cn-route {
  overflow: visible;
}

/* Status color is confined to the inset edge and status dot. */
.status-available {
  --status-accent: rgba(34, 197, 94, 0.55);
}
.status-available .brand-logo-wrap {
  background: rgba(255, 255, 255, 0.035);
  filter: none;
}
.status-available .status-gem {
  background: #10b981;
}

.status-limited {
  --status-accent: rgba(245, 158, 11, 0.55);
}
.status-limited .brand-logo-wrap {
  background: rgba(255, 255, 255, 0.035);
}
.status-limited .status-gem {
  background: #f59e0b;
}

.status-blocked {
  --status-accent: rgba(239, 68, 68, 0.5);
}
.status-blocked .brand-logo-wrap {
  background: rgba(255, 255, 255, 0.035);
}
.status-blocked .status-gem {
  background: #ef4444;
}

.status-untested {
  --status-accent: rgba(100, 116, 139, 0.45);
}
.status-untested .brand-logo-wrap {
  background: rgba(255, 255, 255, 0.025);
}
.status-untested .status-gem {
  background: #64748b;
}

/* Foil and Light Sweeps */
.metal-foil {
  position: absolute;
  inset: 0;
  background: linear-gradient(120deg, transparent 20%, rgba(255, 255, 255, 0.025) 48%, transparent 68%);
  pointer-events: none;
}
.metal-light-sweep {
  display: none;
}

/* Brand Logo Container */
.brand-logo-wrap {
  position: relative;
  width: 26px;
  height: 26px;
  border-radius: 6px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.035);
  transition: background 0.18s ease;
}
.compact-cn-tag {
  position: absolute;
  right: -7px;
  bottom: -6px;
  z-index: 3;
  padding: 0 3px;
  border-radius: 4px;
  background: #f59e0b;
  color: #111827;
  font-size: 7px;
  font-weight: 700;
  line-height: 12px;
  white-space: nowrap;
}
.compact-status-dot {
  position: absolute;
  right: 2px;
  bottom: 2px;
  z-index: 4;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #64748b;
  border: 1px solid rgba(8, 12, 16, 0.9);
}
.status-available .compact-status-dot { background: #34d399; }
.status-limited .compact-status-dot { background: #fbbf24; }
.status-blocked .compact-status-dot { background: #f87171; }
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

/* Status Jewels */
.status-gem {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

/* Metadata */
.metal-badge-meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}
.meta-top-row {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
}
.badge-title {
  min-width: 0;
  overflow: hidden;
  font-size: 11.5px;
  font-weight: 600;
  color: #f8fafc;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.badge-region-tag {
  flex: none;
  margin-left: auto;
  font-family: 'Fira Code', monospace;
  font-size: 8.5px;
  font-weight: 600;
  color: #7dd3fc;
}
.badge-region-tag.is-local {
  color: #697582;
}

.meta-bottom-row {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 9.5px;
  color: #94a3b8;
  white-space: nowrap;
  min-width: 0;
}
.status-desc {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 95px;
}
.badge-latency {
  margin-left: auto;
  flex: none;
  font-family: 'Fira Code', monospace;
  font-size: 8.5px;
  color: #8b97a4;
  font-variant-numeric: tabular-nums;
}
</style>
