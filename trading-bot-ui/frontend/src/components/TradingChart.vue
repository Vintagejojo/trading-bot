<template>
  <v-card>
    <v-card-title class="d-flex align-center">
      <v-icon icon="mdi-chart-candlestick" class="mr-2"></v-icon>
      Multi-Timeframe Chart

      <!-- Timeframe Selector -->
      <v-chip-group
        v-model="selectedTimeframe"
        mandatory
        color="primary"
        class="ml-4"
      >
        <v-chip
          v-for="tf in timeframes"
          :key="tf.value"
          :value="tf.value"
          size="small"
          variant="outlined"
        >
          {{ tf.label }}
        </v-chip>
      </v-chip-group>

      <v-spacer></v-spacer>

      <!-- Indicator Toggles -->
      <v-menu>
        <template v-slot:activator="{ props }">
          <v-btn
            icon="mdi-tune"
            variant="text"
            size="small"
            v-bind="props"
          ></v-btn>
        </template>
        <v-list density="compact">
          <v-list-subheader>Indicators</v-list-subheader>
          <v-list-item>
            <v-checkbox
              v-model="indicators.rsi"
              label="RSI"
              density="compact"
              hide-details
            ></v-checkbox>
          </v-list-item>
          <v-list-item>
            <v-checkbox
              v-model="indicators.macd"
              label="MACD"
              density="compact"
              hide-details
            ></v-checkbox>
          </v-list-item>
          <v-list-item>
            <v-checkbox
              v-model="indicators.bbands"
              label="Bollinger Bands"
              density="compact"
              hide-details
            ></v-checkbox>
          </v-list-item>
        </v-list>
      </v-menu>

      <!-- Refresh Button -->
      <v-btn
        icon="mdi-refresh"
        variant="text"
        size="small"
        @click="loadChartData"
        :loading="loading"
      ></v-btn>
    </v-card-title>

    <v-card-text>
      <!-- Chart Container -->
      <div ref="chartContainer" style="position: relative; height: 500px; width: 100%;"></div>

      <!-- RSI Chart -->
      <div v-if="indicators.rsi" ref="rsiContainer" style="position: relative; height: 150px; width: 100%; margin-top: 10px;"></div>

      <!-- MACD Chart -->
      <div v-if="indicators.macd" ref="macdContainer" style="position: relative; height: 150px; width: 100%; margin-top: 10px;"></div>

      <!-- Status Bar -->
      <v-sheet class="mt-4 pa-2" color="surface-variant" rounded>
        <div class="d-flex justify-space-between text-caption">
          <div>
            <strong>Timeframe:</strong> {{ currentTimeframeLabel }}
          </div>
          <div v-if="chartData">
            <strong>Candles:</strong> {{ chartData.candles.length }}
          </div>
          <div v-if="chartData && chartData.is_ready">
            <v-chip size="x-small" color="success">Ready</v-chip>
          </div>
          <div v-else>
            <v-chip size="x-small" color="warning">Collecting Data</v-chip>
          </div>
          <div v-if="chartData && chartData.indicators">
            <strong>RSI:</strong> {{ chartData.indicators.rsi.toFixed(2) }} |
            <strong>MACD:</strong> {{ chartData.indicators.macd.toFixed(6) }}
          </div>
        </div>
      </v-sheet>
    </v-card-text>
  </v-card>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, computed } from 'vue'
import { createChart } from 'lightweight-charts'
import { GetTimeframeData } from '../../wailsjs/go/main/App'

const chartContainer = ref(null)
const rsiContainer = ref(null)
const macdContainer = ref(null)

const selectedTimeframe = ref('1h')
const loading = ref(false)
const chartData = ref(null)

const timeframes = [
  { label: '5m', value: '5m' },
  { label: '1h', value: '1h' },
  { label: '1d', value: '1d' },
]

const indicators = ref({
  rsi: true,
  macd: true,
  bbands: true,
})

const currentTimeframeLabel = computed(() => {
  const tf = timeframes.find(t => t.value === selectedTimeframe.value)
  return tf ? tf.label : ''
})

let mainChart = null
let candleSeries = null
let bbUpperLine = null
let bbMiddleLine = null
let bbLowerLine = null
let rsiChart = null
let rsiLine = null
let macdChart = null
let macdLine = null
let macdSignalLine = null
let macdHistogram = null
let updateInterval = null

onMounted(() => {
  initCharts()
  loadChartData()

  // Auto-update every 5 seconds
  updateInterval = setInterval(loadChartData, 5000)
})

onBeforeUnmount(() => {
  if (updateInterval) {
    clearInterval(updateInterval)
  }
  destroyCharts()
})

watch(selectedTimeframe, () => {
  loadChartData()
})

watch(() => indicators.value.rsi, (newVal) => {
  if (newVal && !rsiChart) {
    initRSIChart()
    updateRSIChart()
  } else if (!newVal && rsiChart) {
    rsiChart.remove()
    rsiChart = null
    rsiLine = null
  }
})

watch(() => indicators.value.macd, (newVal) => {
  if (newVal && !macdChart) {
    initMACDChart()
    updateMACDChart()
  } else if (!newVal && macdChart) {
    macdChart.remove()
    macdChart = null
    macdLine = null
    macdSignalLine = null
    macdHistogram = null
  }
})

watch(() => indicators.value.bbands, () => {
  updateMainChart()
})

function initCharts() {
  if (!chartContainer.value) return

  // Main candlestick chart
  mainChart = createChart(chartContainer.value, {
    layout: {
      background: { color: '#1E1E1E' },
      textColor: '#D9D9D9',
    },
    grid: {
      vertLines: { color: '#2B2B43' },
      horzLines: { color: '#2B2B43' },
    },
    crosshair: {
      mode: 1,
    },
    rightPriceScale: {
      borderColor: '#2B2B43',
    },
    timeScale: {
      borderColor: '#2B2B43',
      timeVisible: true,
      secondsVisible: false,
    },
    width: chartContainer.value.clientWidth,
    height: 500,
  })

  // Candlestick series
  candleSeries = mainChart.addCandlestickSeries({
    upColor: '#26a69a',
    downColor: '#ef5350',
    borderVisible: false,
    wickUpColor: '#26a69a',
    wickDownColor: '#ef5350',
  })

  // Bollinger Bands
  bbUpperLine = mainChart.addLineSeries({
    color: 'rgba(255, 82, 82, 0.5)',
    lineWidth: 1,
    priceScaleId: 'right',
  })

  bbMiddleLine = mainChart.addLineSeries({
    color: 'rgba(255, 255, 255, 0.5)',
    lineWidth: 1,
    lineStyle: 2, // dashed
    priceScaleId: 'right',
  })

  bbLowerLine = mainChart.addLineSeries({
    color: 'rgba(33, 150, 243, 0.5)',
    lineWidth: 1,
    priceScaleId: 'right',
  })

  // Init RSI and MACD if enabled
  if (indicators.value.rsi) initRSIChart()
  if (indicators.value.macd) initMACDChart()

  // Handle window resize
  window.addEventListener('resize', handleResize)
}

function initRSIChart() {
  if (!rsiContainer.value || rsiChart) return

  rsiChart = createChart(rsiContainer.value, {
    layout: {
      background: { color: '#1E1E1E' },
      textColor: '#D9D9D9',
    },
    grid: {
      vertLines: { color: '#2B2B43' },
      horzLines: { color: '#2B2B43' },
    },
    rightPriceScale: {
      borderColor: '#2B2B43',
      scaleMargins: {
        top: 0.1,
        bottom: 0.1,
      },
    },
    timeScale: {
      borderColor: '#2B2B43',
      visible: false,
    },
    width: rsiContainer.value.clientWidth,
    height: 150,
  })

  rsiLine = rsiChart.addLineSeries({
    color: '#2962FF',
    lineWidth: 2,
  })

  // Add reference lines at 30 and 70
  const oversoldLine = rsiChart.addLineSeries({
    color: 'rgba(239, 83, 80, 0.5)',
    lineWidth: 1,
    lineStyle: 2,
    lastValueVisible: false,
    priceLineVisible: false,
  })
  oversoldLine.setData([{ time: 0, value: 30 }])

  const overboughtLine = rsiChart.addLineSeries({
    color: 'rgba(239, 83, 80, 0.5)',
    lineWidth: 1,
    lineStyle: 2,
    lastValueVisible: false,
    priceLineVisible: false,
  })
  overboughtLine.setData([{ time: 0, value: 70 }])
}

function initMACDChart() {
  if (!macdContainer.value || macdChart) return

  macdChart = createChart(macdContainer.value, {
    layout: {
      background: { color: '#1E1E1E' },
      textColor: '#D9D9D9',
    },
    grid: {
      vertLines: { color: '#2B2B43' },
      horzLines: { color: '#2B2B43' },
    },
    rightPriceScale: {
      borderColor: '#2B2B43',
      scaleMargins: {
        top: 0.1,
        bottom: 0.1,
      },
    },
    timeScale: {
      borderColor: '#2B2B43',
      visible: false,
    },
    width: macdContainer.value.clientWidth,
    height: 150,
  })

  macdHistogram = macdChart.addHistogramSeries({
    color: '#26a69a',
    priceFormat: {
      type: 'price',
      precision: 8,
      minMove: 0.00000001,
    },
  })

  macdLine = macdChart.addLineSeries({
    color: '#2962FF',
    lineWidth: 2,
  })

  macdSignalLine = macdChart.addLineSeries({
    color: '#FF6D00',
    lineWidth: 2,
  })
}

function destroyCharts() {
  if (mainChart) {
    mainChart.remove()
    mainChart = null
  }
  if (rsiChart) {
    rsiChart.remove()
    rsiChart = null
  }
  if (macdChart) {
    macdChart.remove()
    macdChart = null
  }
  window.removeEventListener('resize', handleResize)
}

function handleResize() {
  if (mainChart && chartContainer.value) {
    mainChart.applyOptions({ width: chartContainer.value.clientWidth })
  }
  if (rsiChart && rsiContainer.value) {
    rsiChart.applyOptions({ width: rsiContainer.value.clientWidth })
  }
  if (macdChart && macdContainer.value) {
    macdChart.applyOptions({ width: macdContainer.value.clientWidth })
  }
}

async function loadChartData() {
  if (loading.value) return

  loading.value = true
  try {
    const data = await GetTimeframeData(selectedTimeframe.value)
    chartData.value = data

    updateMainChart()
    if (indicators.value.rsi) updateRSIChart()
    if (indicators.value.macd) updateMACDChart()
  } catch (error) {
    console.error('Error loading chart data:', error)
  } finally {
    loading.value = false
  }
}

function updateMainChart() {
  if (!chartData.value || !candleSeries) return

  // Convert candles to lightweight-charts format
  const candles = chartData.value.candles.map(c => ({
    time: c.timestamp / 1000, // Convert ms to seconds
    open: c.open,
    high: c.high,
    low: c.low,
    close: c.close,
  }))

  candleSeries.setData(candles)

  // Update Bollinger Bands if enabled
  if (indicators.value.bbands && chartData.value.indicators) {
    const bbUpper = candles.map(c => ({
      time: c.time,
      value: chartData.value.indicators.bb_upper || c.close,
    }))

    const bbMiddle = candles.map(c => ({
      time: c.time,
      value: chartData.value.indicators.bb_middle || c.close,
    }))

    const bbLower = candles.map(c => ({
      time: c.time,
      value: chartData.value.indicators.bb_lower || c.close,
    }))

    bbUpperLine.setData(bbUpper)
    bbMiddleLine.setData(bbMiddle)
    bbLowerLine.setData(bbLower)
  }
}

function updateRSIChart() {
  if (!chartData.value || !rsiLine || !chartData.value.indicators) return

  const candles = chartData.value.candles
  if (candles.length === 0) return

  // For now, just show the last RSI value
  // In a real implementation, you'd track RSI history
  const lastCandle = candles[candles.length - 1]
  const rsiData = [{
    time: lastCandle.timestamp / 1000,
    value: chartData.value.indicators.rsi || 50,
  }]

  rsiLine.setData(rsiData)
}

function updateMACDChart() {
  if (!chartData.value || !macdLine || !chartData.value.indicators) return

  const candles = chartData.value.candles
  if (candles.length === 0) return

  const lastCandle = candles[candles.length - 1]
  const time = lastCandle.timestamp / 1000

  // MACD line
  macdLine.setData([{
    time,
    value: chartData.value.indicators.macd || 0,
  }])

  // Signal line
  macdSignalLine.setData([{
    time,
    value: chartData.value.indicators.signal || 0,
  }])

  // Histogram
  const histValue = chartData.value.indicators.histogram || 0
  macdHistogram.setData([{
    time,
    value: histValue,
    color: histValue >= 0 ? '#26a69a' : '#ef5350',
  }])
}
</script>

<style scoped>
/* Chart styling is handled by lightweight-charts */
</style>
