<script setup lang="ts">
const props = defineProps<{ value: string; kind?: 'node' | 'media' }>()
const labels: Record<string, string> = { online: '在线', warning: '注意', alert: '告警', offline: '离线', pending: '待上报', available: '可用', limited: '受限', blocked: '不可用', unknown: '未知' }
</script>

<template>
  <span class="status-badge" :class="`is-${props.value}`"><i aria-hidden="true"></i>{{ labels[props.value] ?? props.value }}</span>
</template>
