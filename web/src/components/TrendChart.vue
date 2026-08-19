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
  const labels = props.points.map((p) => new Intl.DateTimeFormat('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit' }).format(new Date(p.at)))
  chart.setOption({
    animationDuration: matchMedia('(prefers-reduced-motion: reduce)').matches ? 0 : 380,
    grid: { top: 28, right: 16, bottom: 34, left: 42 },
    tooltip: { trigger: 'axis', backgroundColor: '#1f2937', borderWidth: 0, textStyle: { color: '#fff', fontSize: 12 } },
    legend: { top: 0, right: 8, itemWidth: 16, itemHeight: 3, textStyle: { color: '#667085', fontSize: 12 }, data: ['综合风险', 'IPQS', 'Scamalytics'] },
    xAxis: { type: 'category', data: labels, boundaryGap: false, axisLine: { lineStyle: { color: '#dfe3e8' } }, axisLabel: { color: '#7b8492', interval: 4, fontSize: 11 }, axisTick: { show: false } },
    yAxis: { type: 'value', min: 0, max: 100, interval: 25, splitLine: { lineStyle: { color: '#eef0f2' } }, axisLabel: { color: '#8a919c', fontSize: 11 } },
    series: [
      { name: '综合风险', type: 'line', data: props.points.map((p) => p.risk), smooth: 0.25, symbol: 'circle', symbolSize: 5, showSymbol: false, lineStyle: { color: '#a16207', width: 2.5 }, itemStyle: { color: '#a16207' }, markPoint: { symbol: 'circle', symbolSize: 9, data: [{ type: 'max' }], label: { show: false }, itemStyle: { color: '#c2410c', borderColor: '#fff', borderWidth: 2 } } },
      { name: 'IPQS', type: 'line', data: props.points.map((p) => p.ipqs), showSymbol: false, lineStyle: { color: '#475569', width: 1.5, type: 'dashed' } },
      { name: 'Scamalytics', type: 'line', data: props.points.map((p) => p.scamalytics), showSymbol: false, lineStyle: { color: '#0f766e', width: 1.5, type: 'dotted' } },
    ],
  })
}

onMounted(() => { render(); observer = new ResizeObserver(() => chart?.resize()); if (host.value) observer.observe(host.value) })
watch(() => props.points, render, { deep: true })
onBeforeUnmount(() => { observer?.disconnect(); chart?.dispose() })
</script>

<template><div ref="host" class="trend-chart" role="img" aria-label="最近十天风险分趋势图"></div></template>
