<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    value: string
    kind?: 'node' | 'media' | 'ai'
    pill?: boolean
    region?: string
    showDot?: boolean
  }>(),
  {
    kind: 'node',
    pill: false,
    region: '',
    showDot: true,
  }
)

const labels: Record<string, string> = {
  online: '在线',
  warning: '注意',
  alert: '告警',
  offline: '离线',
  pending: '待上报',
  available: '可用',
  limited: '受限',
  blocked: '不可用',
  unknown: '未知',
}

const statusClass = computed(() => {
  const v = props.value?.toLowerCase() || 'unknown'
  if (['online', 'available', 'good'].includes(v)) return 'is-good'
  if (['warning', 'limited', 'warn'].includes(v)) return 'is-warn'
  if (['alert', 'blocked', 'danger', 'critical'].includes(v)) return 'is-danger'
  if (['offline'].includes(v)) return 'is-offline'
  return 'is-neutral'
})
</script>

<template>
  <span class="status-pill" :class="[statusClass, { 'as-capsule': pill }]">
    <i v-if="showDot" class="status-pulse-dot" aria-hidden="true"></i>
    <span class="status-text">{{ labels[props.value] ?? props.value }}</span>
    <span v-if="props.region" class="status-region-tag">{{ props.region }}</span>
  </span>
</template>

<style scoped>
.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 7px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.4;
  white-space: nowrap;
  transition: all 0.18s ease;
  user-select: none;
}

.as-capsule {
  border-radius: 9999px;
  padding: 3px 9px;
  font-size: 10px;
  border: 1px solid currentColor;
}

.status-pulse-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  position: relative;
}

.is-good {
  color: var(--good, #10b981);
  background: color-mix(in srgb, var(--good, #10b981) 12%, transparent);
}
.is-good .status-pulse-dot {
  background: var(--good, #10b981);
  box-shadow: 0 0 6px var(--good, #10b981);
}

.is-warn {
  color: var(--warning, #f59e0b);
  background: color-mix(in srgb, var(--warning, #f59e0b) 12%, transparent);
}
.is-warn .status-pulse-dot {
  background: var(--warning, #f59e0b);
  box-shadow: 0 0 6px var(--warning, #f59e0b);
}

.is-danger {
  color: var(--danger, #ef4444);
  background: color-mix(in srgb, var(--danger, #ef4444) 12%, transparent);
}
.is-danger .status-pulse-dot {
  background: var(--danger, #ef4444);
  box-shadow: 0 0 6px var(--danger, #ef4444);
}

.is-offline {
  color: var(--muted, #64748b);
  background: color-mix(in srgb, var(--muted, #64748b) 12%, transparent);
}
.is-offline .status-pulse-dot {
  background: var(--muted, #64748b);
}

.is-neutral {
  color: var(--faint, #94a3b8);
  background: rgba(148, 163, 184, 0.1);
}
.is-neutral .status-pulse-dot {
  background: var(--faint, #94a3b8);
}

.status-region-tag {
  display: inline-block;
  padding: 0 4px;
  border-radius: 2px;
  font-family: 'Fira Code', monospace;
  font-size: 9px;
  font-weight: 600;
  opacity: 0.85;
  background: rgba(0, 0, 0, 0.15);
}
</style>
