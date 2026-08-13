<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  CategoryScale,
  LinearScale,
  BarElement,
  ArcElement,
  BarController,
  DoughnutController,
} from 'chart.js'
import jsPDF from 'jspdf'
import autoTable from 'jspdf-autotable'

import { useApi } from '@/composables/useApi'
import { formatCurrency, formatDate } from '@/utils/format'
import { fetchLatestReportMetadata, uploadGeneratedReport, viewLatestReportPdf } from '@/services/reports'

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  CategoryScale,
  LinearScale,
  BarElement,
  ArcElement,
  BarController,
  DoughnutController
)

const CHART_COLORS = ['#3b82f6', '#f59e0b', '#22c55e', '#ec4899', '#8b5cf6', '#14b8a6', '#f97316']

const { get } = useApi()

const discs = ref([])
const isLoadingDiscs = ref(true)
const loadError = ref('')

async function loadDiscs() {
  isLoadingDiscs.value = true
  loadError.value = ''

  try {
    const data = await get('/discs')
    discs.value = Array.isArray(data) ? data : []
  } catch (err) {
    loadError.value = err instanceof Error ? err.message : 'Impossible de charger la collection.'
  } finally {
    isLoadingDiscs.value = false
  }
}

const reportPeriod = ref('all')
const reportCustomStart = ref('')
const reportCustomEnd = ref('')
const reportGenre = ref('')
const reportArtist = ref('')
const reportFormat = ref('')
const reportYear = ref('')

const periodOptions = [
  { value: '30d', label: '30 jours' },
  { value: '90d', label: '90 jours' },
  { value: '12m', label: '12 mois' },
  { value: 'ytd', label: 'Cette année' },
  { value: 'all', label: 'Tout' },
  { value: 'custom', label: 'Période personnalisée' },
]

function isWithinRange(dateValue, range) {
  if (range === 'all') return true

  const date = new Date(dateValue)
  if (Number.isNaN(date.getTime())) return false

  if (range === 'custom') {
    if (reportCustomStart.value) {
      const start = new Date(reportCustomStart.value)
      start.setHours(0, 0, 0, 0)
      if (date < start) return false
    }

    if (reportCustomEnd.value) {
      const end = new Date(reportCustomEnd.value)
      end.setHours(23, 59, 59, 999)
      if (date > end) return false
    }

    return true
  }

  const now = new Date()
  const startDate = new Date(now)

  if (range === '30d') startDate.setDate(now.getDate() - 30)
  else if (range === '90d') startDate.setDate(now.getDate() - 90)
  else if (range === '12m') startDate.setMonth(now.getMonth() - 12)
  else if (range === 'ytd') {
    startDate.setMonth(0, 1)
    startDate.setHours(0, 0, 0, 0)
  }

  return date >= startDate && date <= now
}

function uniqueSortedValues(items, keyOf) {
  const values = new Set()
  for (const item of items) {
    const value = keyOf(item)
    if (value) values.add(value)
  }
  return Array.from(values).sort((a, b) => String(a).localeCompare(String(b), 'fr', { sensitivity: 'base' }))
}

const genreOptions = computed(() => uniqueSortedValues(discs.value, (d) => d.genre_name))
const artistOptions = computed(() => uniqueSortedValues(discs.value, (d) => d.artist_name))
const formatOptions = computed(() => uniqueSortedValues(discs.value, (d) => d.format_name))
const yearOptions = computed(() =>
  uniqueSortedValues(discs.value, (d) => d.release_year).sort((a, b) => b - a)
)

const filteredDiscs = computed(() =>
  discs.value.filter((disc) => {
    if (!isWithinRange(disc.created_at, reportPeriod.value)) return false
    if (reportGenre.value && (disc.genre_name || '') !== reportGenre.value) return false
    if (reportArtist.value && (disc.artist_name || '') !== reportArtist.value) return false
    if (reportFormat.value && (disc.format_name || '') !== reportFormat.value) return false
    if (reportYear.value && String(disc.release_year || '') !== String(reportYear.value)) return false
    return true
  })
)

const discCount = computed(() => filteredDiscs.value.length)

const pricedDiscs = computed(() =>
  filteredDiscs.value.filter((disc) => disc.price != null && disc.price !== '' && !Number.isNaN(parseFloat(disc.price)))
)

const totalValue = computed(() =>
  pricedDiscs.value.reduce((sum, disc) => sum + parseFloat(disc.price), 0)
)

const averagePrice = computed(() =>
  pricedDiscs.value.length > 0 ? totalValue.value / pricedDiscs.value.length : 0
)

const artistCount = computed(() => {
  const names = new Set()
  for (const disc of filteredDiscs.value) {
    if (disc.artist_name) names.add(disc.artist_name)
  }
  return names.size
})

function buildBreakdown(items, keyOf, emptyLabel) {
  const totals = new Map()

  for (const item of items) {
    const key = keyOf(item)?.trim() || emptyLabel
    totals.set(key, (totals.get(key) || 0) + 1)
  }

  const sorted = Array.from(totals.entries()).sort((a, b) => b[1] - a[1])
  const top = sorted.slice(0, 6)
  const others = sorted.slice(6)
  const othersTotal = others.reduce((sum, [, count]) => sum + count, 0)

  const labels = top.map(([label]) => label)
  const values = top.map(([, value]) => value)

  if (othersTotal > 0) {
    labels.push('Autres')
    values.push(othersTotal)
  }

  return { labels, values }
}

const genreBreakdown = computed(() =>
  buildBreakdown(filteredDiscs.value, (d) => d.genre_name, 'Sans genre')
)
const countryBreakdown = computed(() =>
  buildBreakdown(filteredDiscs.value, (d) => d.country_name, 'Sans pays')
)

function monthKey(date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
}

function formatMonthLabel(date) {
  const formatted = new Intl.DateTimeFormat('fr-FR', { month: 'short', year: '2-digit' }).format(date)
  return formatted.charAt(0).toUpperCase() + formatted.slice(1)
}

const monthlyAcquisitions = computed(() => {
  const counts = new Map()

  for (const disc of filteredDiscs.value) {
    if (!disc.created_at) continue
    const date = new Date(disc.created_at)
    if (Number.isNaN(date.getTime())) continue
    const key = monthKey(date)
    counts.set(key, (counts.get(key) || 0) + 1)
  }

  const keys = Array.from(counts.keys()).sort()
  const labels = keys.map((key) => {
    const [year, month] = key.split('-').map(Number)
    return formatMonthLabel(new Date(year, month - 1, 1))
  })
  const values = keys.map((key) => counts.get(key))

  return { labels, values }
})

const reportRows = computed(() =>
  [...filteredDiscs.value].sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
)

function formatFilterDate(value) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat('fr-FR').format(date)
}

const criteriaDescription = computed(() => {
  const parts = []

  if (reportPeriod.value === 'custom') {
    const start = reportCustomStart.value ? formatFilterDate(reportCustomStart.value) : '…'
    const end = reportCustomEnd.value ? formatFilterDate(reportCustomEnd.value) : '…'
    parts.push(`Période : du ${start} au ${end}`)
  } else {
    const found = periodOptions.find((option) => option.value === reportPeriod.value)
    parts.push(`Période : ${found ? found.label : 'Tout'}`)
  }

  parts.push(reportGenre.value ? `Genre : ${reportGenre.value}` : 'Tous les genres')
  if (reportArtist.value) parts.push(`Artiste : ${reportArtist.value}`)
  if (reportFormat.value) parts.push(`Format : ${reportFormat.value}`)
  if (reportYear.value) parts.push(`Année : ${reportYear.value}`)

  return parts.join(' • ')
})

function resetFilters() {
  reportPeriod.value = 'all'
  reportCustomStart.value = ''
  reportCustomEnd.value = ''
  reportGenre.value = ''
  reportArtist.value = ''
  reportFormat.value = ''
  reportYear.value = ''
}

// Les polices standard de jsPDF (Helvetica) n'ont pas les glyphes des espaces
// insécable/fine insécable utilisées par Intl pour séparer les milliers —
// elles s'affichaient comme des "/" ou cassaient l'alignement des chiffres.
function formatCurrencyForPdf(value) {
  return formatCurrency(value).replace(/[  ]/g, ' ')
}

function formatTableDate(value) {
  return formatDate(value) || ''
}

const genreBreakdownCanvas = ref(null)
const countryBreakdownCanvas = ref(null)
const evolutionCanvas = ref(null)

let genreBreakdownChart = null
let countryBreakdownChart = null
let evolutionChart = null

function destroyCharts() {
  genreBreakdownChart?.destroy()
  countryBreakdownChart?.destroy()
  evolutionChart?.destroy()
  genreBreakdownChart = null
  countryBreakdownChart = null
  evolutionChart = null
}

async function renderCharts() {
  destroyCharts()
  await nextTick()

  if (genreBreakdown.value.labels.length && genreBreakdownCanvas.value) {
    genreBreakdownChart = new ChartJS(genreBreakdownCanvas.value.getContext('2d'), {
      type: 'doughnut',
      data: {
        labels: genreBreakdown.value.labels,
        datasets: [{ data: genreBreakdown.value.values, backgroundColor: CHART_COLORS, borderWidth: 0 }],
      },
      options: {
        responsive: false,
        animation: false,
        plugins: {
          legend: { position: 'right', labels: { color: '#3a3f33', font: { size: 11 }, boxWidth: 12 } },
          title: { display: true, text: 'Répartition par genre', color: '#22262a', font: { size: 13, weight: '600' } },
        },
      },
    })
  }

  if (countryBreakdown.value.labels.length && countryBreakdownCanvas.value) {
    countryBreakdownChart = new ChartJS(countryBreakdownCanvas.value.getContext('2d'), {
      type: 'doughnut',
      data: {
        labels: countryBreakdown.value.labels,
        datasets: [{ data: countryBreakdown.value.values, backgroundColor: CHART_COLORS, borderWidth: 0 }],
      },
      options: {
        responsive: false,
        animation: false,
        plugins: {
          legend: { position: 'right', labels: { color: '#3a3f33', font: { size: 11 }, boxWidth: 12 } },
          title: { display: true, text: 'Répartition par pays', color: '#22262a', font: { size: 13, weight: '600' } },
        },
      },
    })
  }

  if (monthlyAcquisitions.value.labels.length && evolutionCanvas.value) {
    evolutionChart = new ChartJS(evolutionCanvas.value.getContext('2d'), {
      type: 'bar',
      data: {
        labels: monthlyAcquisitions.value.labels,
        datasets: [{ label: 'Disques ajoutés', data: monthlyAcquisitions.value.values, backgroundColor: '#3b82f6' }],
      },
      options: {
        responsive: false,
        animation: false,
        scales: {
          x: { ticks: { color: '#3a3f33' }, grid: { display: false } },
          y: { ticks: { color: '#3a3f33', precision: 0 }, grid: { color: '#e5e7eb' } },
        },
        plugins: {
          legend: { display: false },
          title: { display: true, text: 'Acquisitions par mois', color: '#22262a', font: { size: 13, weight: '600' } },
        },
      },
    })
  }
}

watch(
  [reportPeriod, reportCustomStart, reportCustomEnd, reportGenre, reportArtist, reportFormat, reportYear, discs],
  renderCharts,
  { deep: true }
)

// Le watch ci-dessus (dépendance "discs") déclenche déjà le premier rendu
// dès que loadDiscs() peuple la liste — un second appel explicite ici
// entrerait en course avec lui sur les mêmes canvas.
onMounted(loadDiscs)
onBeforeUnmount(destroyCharts)

async function addChartToDoc(doc, chart, x, y, maxWidth) {
  if (!chart) return y

  const dataUrl = chart.toBase64Image()
  const props = doc.getImageProperties(dataUrl)
  const width = maxWidth
  const height = width * (props.height / props.width)

  doc.addImage(dataUrl, 'PNG', x, y, width, height)

  return y + height
}

const isGenerating = ref(false)
const generationError = ref('')
const generationSuccess = ref('')

const latestReportMeta = ref(null)
const isLoadingLatest = ref(true)
const latestReportError = ref('')

async function loadLatestReport() {
  isLoadingLatest.value = true
  latestReportError.value = ''

  try {
    latestReportMeta.value = await fetchLatestReportMetadata()
  } catch (err) {
    latestReportError.value = err instanceof Error ? err.message : 'Erreur inconnue.'
  } finally {
    isLoadingLatest.value = false
  }
}

onMounted(loadLatestReport)

async function openLatestReport() {
  latestReportError.value = ''

  try {
    await viewLatestReportPdf()
  } catch (err) {
    latestReportError.value = err instanceof Error ? err.message : 'Impossible d’ouvrir le rapport.'
  }
}

function formatGeneratedAt(value) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat('fr-FR', { dateStyle: 'long', timeStyle: 'short' }).format(date)
}

function formatSize(bytes) {
  const size = Number(bytes || 0)
  if (size < 1024) return `${size} o`
  const kb = size / 1024
  if (kb < 1024) return `${kb.toFixed(0)} Ko`
  return `${(kb / 1024).toFixed(1)} Mo`
}

async function generateReport() {
  isGenerating.value = true
  generationError.value = ''
  generationSuccess.value = ''

  try {
    await renderCharts()
    await nextTick()

    const doc = new jsPDF({ unit: 'mm', format: 'a4' })
    const pageWidth = doc.internal.pageSize.getWidth()
    const marginX = 14

    doc.setFillColor(59, 130, 246)
    doc.rect(0, 0, pageWidth, 32, 'F')
    doc.setTextColor(255, 255, 255)
    doc.setFont('helvetica', 'bold')
    doc.setFontSize(20)
    doc.text('Rapport de collection', marginX, 18)
    doc.setFont('helvetica', 'normal')
    doc.setFontSize(10)
    const generatedAtLabel = new Intl.DateTimeFormat('fr-FR', {
      dateStyle: 'long',
      timeStyle: 'short',
    }).format(new Date())
    doc.text(`Généré le ${generatedAtLabel}`, marginX, 26)

    let cursorY = 42
    doc.setTextColor(58, 63, 51)
    doc.setFont('helvetica', 'bold')
    doc.setFontSize(11)
    doc.text('Critères appliqués', marginX, cursorY)
    cursorY += 6
    doc.setFont('helvetica', 'normal')
    doc.setFontSize(9.5)
    doc.setTextColor(90, 96, 88)
    doc.text(criteriaDescription.value, marginX, cursorY)
    cursorY += 10

    const kpis = [
      { label: 'Nombre de disques', value: String(discCount.value) },
      { label: 'Valeur de la collection', value: formatCurrencyForPdf(totalValue.value) },
      { label: 'Prix moyen', value: formatCurrencyForPdf(averagePrice.value) },
      { label: "Nombre d'artistes", value: String(artistCount.value) },
    ]

    const cardsPerRow = 4
    const cardGap = 4
    const cardHeight = 20
    const cardWidth = (pageWidth - marginX * 2 - cardGap * (cardsPerRow - 1)) / cardsPerRow

    kpis.forEach((kpi, index) => {
      const row = Math.floor(index / cardsPerRow)
      const col = index % cardsPerRow
      const x = marginX + col * (cardWidth + cardGap)
      const y = cursorY + row * (cardHeight + cardGap)

      doc.setFillColor(240, 242, 238)
      doc.roundedRect(x, y, cardWidth, cardHeight, 2, 2, 'F')
      doc.setTextColor(120, 128, 114)
      doc.setFont('helvetica', 'normal')
      doc.setFontSize(7.5)
      doc.text(kpi.label, x + 3, y + 6)
      doc.setTextColor(40, 46, 38)
      doc.setFont('helvetica', 'bold')
      doc.setFontSize(11)
      doc.text(kpi.value, x + 3, y + 15)
    })

    const kpiRows = Math.ceil(kpis.length / cardsPerRow)
    cursorY += kpiRows * (cardHeight + cardGap) + 6

    const halfWidth = (pageWidth - marginX * 2 - 6) / 2

    if (genreBreakdownChart && countryBreakdownChart) {
      const bottom1 = await addChartToDoc(doc, genreBreakdownChart, marginX, cursorY, halfWidth)
      const bottom2 = await addChartToDoc(
        doc,
        countryBreakdownChart,
        marginX + halfWidth + 6,
        cursorY,
        halfWidth
      )
      cursorY = Math.max(bottom1, bottom2) + 8
    } else if (genreBreakdownChart) {
      cursorY = (await addChartToDoc(doc, genreBreakdownChart, marginX, cursorY, pageWidth - marginX * 2)) + 8
    } else if (countryBreakdownChart) {
      cursorY = (await addChartToDoc(doc, countryBreakdownChart, marginX, cursorY, pageWidth - marginX * 2)) + 8
    }

    if (evolutionChart) {
      cursorY = (await addChartToDoc(doc, evolutionChart, marginX, cursorY, pageWidth - marginX * 2)) + 8
    }

    doc.addPage()
    doc.setTextColor(40, 46, 38)
    doc.setFont('helvetica', 'bold')
    doc.setFontSize(13)
    doc.text('Détail des disques', marginX, 16)

    autoTable(doc, {
      startY: 22,
      head: [['Titre', 'Artiste', 'Genre', 'Format', 'Année', 'Prix', "Date d'ajout"]],
      body: reportRows.value.map((disc) => [
        disc.title || '—',
        disc.artist_name || '—',
        disc.genre_name || '—',
        disc.format_name || '—',
        disc.release_year ? String(disc.release_year) : '—',
        disc.price != null && disc.price !== '' ? formatCurrencyForPdf(disc.price) : '—',
        formatTableDate(disc.created_at),
      ]),
      styles: { fontSize: 8, cellPadding: 2.4 },
      headStyles: { fillColor: [59, 130, 246], textColor: 255 },
      alternateRowStyles: { fillColor: [244, 245, 242] },
      columnStyles: { 5: { halign: 'right', cellWidth: 20, minCellWidth: 20 } },
      margin: { left: marginX, right: marginX },
    })

    const pageCount = doc.internal.getNumberOfPages()
    for (let i = 1; i <= pageCount; i += 1) {
      doc.setPage(i)
      doc.setFont('helvetica', 'normal')
      doc.setFontSize(8)
      doc.setTextColor(150, 150, 150)
      doc.text(
        `Page ${i} / ${pageCount}`,
        pageWidth - marginX,
        doc.internal.pageSize.getHeight() - 8,
        { align: 'right' }
      )
      doc.text('Disques Manager', marginX, doc.internal.pageSize.getHeight() - 8)
    }

    const blob = doc.output('blob')
    const meta = await uploadGeneratedReport(blob, criteriaDescription.value)
    latestReportMeta.value = meta
    generationSuccess.value = 'Le rapport a été généré avec succès.'
  } catch (err) {
    generationError.value = err instanceof Error ? err.message : 'Échec de la génération du rapport.'
  } finally {
    isGenerating.value = false
  }
}
</script>

<template>
  <div class="settings-reports-view">
    <header class="page-header">
      <div class="header-wrapper">
        <div class="header-top-row">
          <div class="title-section">
            <button @click="$router.push('/dashboard')" class="back-button ghost-btn">
              <span class="icon">🏠</span>
            </button>
            <span class="title-icon">📈</span>
            <h1>Éditions</h1>
          </div>
        </div>
        <p class="subtitle">Génère un état PDF complet de ta collection : indicateurs, répartitions et détail des disques.</p>
      </div>
    </header>

    <div class="reports-view">
      <p v-if="loadError" class="form-error">{{ loadError }}</p>

      <section class="panel reports-card">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Dernier état généré</p>
            <h2>Rapport disponible</h2>
          </div>
        </div>

        <p v-if="isLoadingLatest" class="reports-muted">Chargement…</p>

        <template v-else-if="latestReportMeta">
          <div class="latest-report-details">
            <div>
              <p class="eyebrow">Généré le</p>
              <p class="detail-value">{{ formatGeneratedAt(latestReportMeta.generated_at) }}</p>
            </div>
            <div>
              <p class="eyebrow">Critères</p>
              <p class="detail-value">{{ latestReportMeta.criteria || '—' }}</p>
            </div>
            <div>
              <p class="eyebrow">Taille</p>
              <p class="detail-value">{{ formatSize(latestReportMeta.size_bytes) }}</p>
            </div>
          </div>

          <button class="primary-btn" type="button" @click="openLatestReport">Ouvrir le PDF</button>
        </template>

        <p v-else class="reports-muted">Aucun état n'a encore été généré.</p>

        <p v-if="latestReportError" class="form-error">{{ latestReportError }}</p>
      </section>

      <section class="panel reports-card">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Nouveau rapport</p>
            <h2>Critères de génération</h2>
          </div>
          <button class="ghost-btn" type="button" @click="resetFilters">Réinitialiser</button>
        </div>

        <div class="reports-field-grid">
          <label class="form-field">
            <span>Période</span>
            <select v-model="reportPeriod">
              <option v-for="option in periodOptions" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>

          <label class="form-field">
            <span>Genre</span>
            <select v-model="reportGenre">
              <option value="">Tous les genres</option>
              <option v-for="genre in genreOptions" :key="genre" :value="genre">{{ genre }}</option>
            </select>
          </label>

          <label class="form-field">
            <span>Artiste</span>
            <select v-model="reportArtist">
              <option value="">Tous les artistes</option>
              <option v-for="artist in artistOptions" :key="artist" :value="artist">{{ artist }}</option>
            </select>
          </label>

          <label class="form-field">
            <span>Format</span>
            <select v-model="reportFormat">
              <option value="">Tous les formats</option>
              <option v-for="format in formatOptions" :key="format" :value="format">{{ format }}</option>
            </select>
          </label>

          <label class="form-field">
            <span>Année</span>
            <select v-model="reportYear">
              <option value="">Toutes les années</option>
              <option v-for="year in yearOptions" :key="year" :value="year">{{ year }}</option>
            </select>
          </label>

          <template v-if="reportPeriod === 'custom'">
            <label class="form-field">
              <span>Du</span>
              <input v-model="reportCustomStart" type="date" />
            </label>
            <label class="form-field">
              <span>Au</span>
              <input v-model="reportCustomEnd" type="date" />
            </label>
          </template>
        </div>

        <p v-if="isLoadingDiscs" class="reports-muted">Chargement de la collection…</p>

        <template v-else>
          <div class="reports-kpi-grid">
            <div class="reports-kpi">
              <p class="eyebrow">Nombre de disques</p>
              <p class="reports-kpi-value">{{ discCount }}</p>
            </div>
            <div class="reports-kpi">
              <p class="eyebrow">Valeur de la collection</p>
              <p class="reports-kpi-value">{{ formatCurrency(totalValue) }}</p>
            </div>
            <div class="reports-kpi">
              <p class="eyebrow">Prix moyen</p>
              <p class="reports-kpi-value">{{ formatCurrency(averagePrice) }}</p>
            </div>
            <div class="reports-kpi">
              <p class="eyebrow">Nombre d'artistes</p>
              <p class="reports-kpi-value">{{ artistCount }}</p>
            </div>
          </div>

          <div class="reports-charts">
            <canvas ref="genreBreakdownCanvas" width="320" height="220"></canvas>
            <canvas ref="countryBreakdownCanvas" width="320" height="220"></canvas>
            <canvas ref="evolutionCanvas" width="660" height="220" class="reports-chart-wide"></canvas>
          </div>

          <p v-if="generationError" class="form-error">{{ generationError }}</p>
          <p v-if="generationSuccess" class="form-success">{{ generationSuccess }}</p>

          <button class="primary-btn" type="button" :disabled="isGenerating || discCount === 0" @click="generateReport">
            {{ isGenerating ? 'Génération…' : '📄 Générer le rapport PDF' }}
          </button>
        </template>
      </section>
    </div>
  </div>
</template>

<style scoped>
.settings-reports-view {
  padding: 20px;
  min-height: 100vh;
  box-sizing: border-box;
}

.header-wrapper { padding: 0; margin: 0; }
.header-top-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 5px; flex-wrap: wrap; gap: 16px; }
.title-section { display: flex; align-items: center; gap: 12px; }
.back-button { display: none; padding: 0; width: 38px; height: 38px; border-radius: 50%; align-items: center; justify-content: center; }

@media (max-width: 767px) {
  .back-button { display: flex; }
  .settings-reports-view { padding: 10px; }
}

.title-icon { font-size: 2em; }
.page-header h1 { color: var(--text); font-size: 2em; margin: 0; font-weight: bold; }
.subtitle { color: var(--text-soft); margin: 0 0 20px 0; font-size: 1.1em; }

.reports-view {
  display: grid;
  gap: 0.9rem;
}

.reports-card {
  padding: 1.1rem;
}

.reports-muted {
  margin: 0.9rem 0 0;
  color: var(--text-dim);
  font-size: 0.88rem;
}

.latest-report-details {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.9rem;
  margin: 0.9rem 0;
}

.detail-value {
  margin: 0.15rem 0 0;
  color: var(--text);
  font-weight: 600;
}

@media (max-width: 640px) {
  .latest-report-details {
    grid-template-columns: 1fr;
  }
}

.reports-field-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.9rem;
  margin-top: 0.9rem;
}

@media (max-width: 900px) {
  .reports-field-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .reports-field-grid {
    grid-template-columns: 1fr;
  }
}

.reports-kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
  margin-top: 1.2rem;
}

@media (max-width: 900px) {
  .reports-kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.reports-kpi {
  padding: 0.9rem 1rem;
  border-radius: 14px;
  background: rgba(var(--tint-rgb), 0.03);
  border: 1px solid var(--line-soft);
}

.reports-kpi-value {
  margin: 0.3rem 0 0;
  color: var(--text);
  font-size: 1.4rem;
  font-weight: 700;
}

.reports-charts {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
  margin-top: 1.2rem;
}

.reports-charts canvas {
  max-width: 100%;
  height: auto;
  background: #f4f5f2;
  border-radius: 12px;
  padding: 8px;
  box-sizing: border-box;
}

.reports-chart-wide {
  grid-column: 1 / -1;
}

@media (max-width: 760px) {
  .reports-charts {
    grid-template-columns: 1fr;
  }
}

.reports-view .primary-btn {
  margin-top: 1.2rem;
}

.reports-view .primary-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
