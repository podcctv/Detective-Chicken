<script setup lang="ts">
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, LegendComponent, MarkPointComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { TrendPoint } from '../types'

const props = defineProps<{ points: TrendPoint[] }>()
const host = ref<HTMLDivElement>()
let chart: echarts.ECharts | undefined
let observer: ResizeObserver | undefined

echarts.use([LineChart, GridComponent, LegendComponent, MarkPointComponent, TooltipComponent, CanvasRenderer])

const render = () => {
  if (!host.value) return
  chart ??= echarts.init(host.value)
  const isDark = document.documentElement.dataset.theme === 'dark'
  const labels = props.points.map((p) =>
    new Intl.DateTimeFormat('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit' }).format(new Date(p.at))
  )

  const textMuted = isDark ? '#94a3b8' : '#64748b'
  const splitLineColor = isDark ? 'rgba(255, 255, 255, 0.05)' : '#eef0f2'
  const axisLineColor = isDark ? 'rgba(255, 255, 255, 0.1)' : '#dfe3e8'

  chart.setOption({
    animationDuration: matchMedia('(prefers-reduced-motion: reduce)').matches ? 0 : 380,
    grid: { top: 28, right: 16, bottom: 30, left: 40 },
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark ? 'rgba(15, 23, 42, 0.92)' : '#ffffff',
      borderColor: isDark ? 'rgba(56, 189, 248, 0.3)' : '#e2e8f0',
      borderWidth: 1,
      textStyle: { color: isDark ? '#f8fafc' : '#1e293b', fontSize: 12 },
    },
    legend: {
      top: 0,
      right: 8,
      itemWidth: 14,
      itemHeight: 4,
      textStyle: { color: textMuted, fontSize: 11 },
      data: ['综合风险', 'IPQS', 'Scamalytics'],
    },
    xAxis: {
      type: 'category',
      data: labels,
      boundaryGap: false,
      axisLine: { lineStyle: { color: axisLineColor } },
      axisLabel: { color: textMuted, interval: 4, fontSize: 10 },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value',
      min: 0,
      max: 100,
      interval: 25,
      splitLine: { lineStyle: { color: splitLineColor } },
      axisLabel: { color: textMuted, fontSize: 10 },
    },
    series: [
      {
        name: '综合风险',
        type: 'line',
        data: props.points.map((p) => p.risk),
        smooth: 0.3,
        symbol: 'circle',
        symbolSize: 6,
        showSymbol: false,
        lineStyle: { color: '#f59e0b', width: 2.5 },
        itemStyle: { color: '#f59e0b' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(245, 158, 11, 0.28)' },
            { offset: 1, color: 'rgba(245, 158, 11, 0.0)' },
          ]),
        },
        markPoint: {
          symbol: 'circle',
          symbolSize: 9,
          data: [{ type: 'max' }],
          label: { show: false },
          itemStyle: { color: '#ef4444', borderColor: '#fff', borderWidth: 2 },
        },
      },
      {
        name: 'IPQS',
        type: 'line',
        data: props.points.map((p) => p.ipqs),
        showSymbol: false,
        lineStyle: { color: '#38bdf8', width: 1.5, type: 'dashed' },
      },
      {
        name: 'Scamalytics',
        type: 'line',
        data: props.points.map((p) => p.scamalytics),
        showSymbol: false,
        lineStyle: { color: '#10b981', width: 1.5, type: 'dotted' },
      },
    ],
  })
}

onMounted(() => {
  render()
  observer = new ResizeObserver(() => chart?.resize())
  if (host.value) observer.observe(host.value)
})

watch(() => props.points, render, { deep: true })
onBeforeUnmount(() => {
  observer?.disconnect()
  chart?.dispose()
})
</script>

<template>
  <div ref="host" class="trend-chart" role="img" aria-label="最近十天风险分趋势图"></div>
</template>
