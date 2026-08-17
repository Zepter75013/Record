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
import ExcelJS from 'exceljs'
import { Document, Packer, Paragraph, HeadingLevel, Table, TableRow, TableCell, TextRun, ImageRun, WidthType, PageOrientation } from 'docx'

import { useApi } from '@/composables/useApi'
import { formatCurrency, formatDate } from '@/utils/format'
import { fetchLatestReportMetadata, uploadGeneratedReport, viewLatestReportFile } from '@/services/reports'
import { fetchTracks } from '@/services/tracks'

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

// ✅ Préférences de mise en forme du rapport (police/taille/pistes) —
// persistées en localStorage pour ne pas devoir les reparamétrer à chaque
// connexion. Les filtres (période/genre/artiste...) ne le sont pas : ils
// décrivent la portée d'UN rapport ponctuel, pas une préférence d'affichage.
const REPORT_PREFS_KEY = 'recordManagerReportPreferences'
const OUTPUT_FORMATS = ['pdf', 'xlsx', 'docx', 'csv']
const ORIENTATIONS = ['portrait', 'landscape']
const SORT_FIELDS = ['title', 'artist_name', 'genre_name', 'format_name', 'release_year', 'price', 'created_at']
const SORT_DIRECTIONS = ['asc', 'desc']
const DEFAULT_SORT_RULES = [{ field: 'created_at', direction: 'desc' }]
const REPORT_PREFS_DEFAULTS = {
  fontFamily: 'helvetica',
  fontSize: 8,
  includeTracklist: false,
  outputFormat: 'pdf',
  orientation: 'portrait',
  sortRules: DEFAULT_SORT_RULES,
}

function sanitizeSortRules(rules) {
  if (!Array.isArray(rules)) return [...DEFAULT_SORT_RULES]
  const cleaned = rules.filter(
    (rule) => rule && SORT_FIELDS.includes(rule.field) && SORT_DIRECTIONS.includes(rule.direction)
  )
  return cleaned.length > 0 ? cleaned : [...DEFAULT_SORT_RULES]
}

function loadReportPrefs() {
  try {
    const raw = localStorage.getItem(REPORT_PREFS_KEY)
    if (!raw) return { ...REPORT_PREFS_DEFAULTS }
    const parsed = JSON.parse(raw)
    return {
      fontFamily: ['helvetica', 'times', 'courier'].includes(parsed.fontFamily)
        ? parsed.fontFamily
        : REPORT_PREFS_DEFAULTS.fontFamily,
      fontSize: Number.isFinite(parsed.fontSize) ? parsed.fontSize : REPORT_PREFS_DEFAULTS.fontSize,
      includeTracklist: !!parsed.includeTracklist,
      outputFormat: OUTPUT_FORMATS.includes(parsed.outputFormat) ? parsed.outputFormat : REPORT_PREFS_DEFAULTS.outputFormat,
      orientation: ORIENTATIONS.includes(parsed.orientation) ? parsed.orientation : REPORT_PREFS_DEFAULTS.orientation,
      sortRules: sanitizeSortRules(parsed.sortRules),
    }
  } catch {
    return { ...REPORT_PREFS_DEFAULTS }
  }
}

const reportPrefs = loadReportPrefs()
const reportIncludeTracklist = ref(reportPrefs.includeTracklist)
const reportFontFamily = ref(reportPrefs.fontFamily)
const reportFontSize = ref(reportPrefs.fontSize)
// Format de sortie du rapport généré — nommé "output" pour ne pas être
// confondu avec reportFormat, qui est le FILTRE sur le format du disque
// (Vinyle/CD/...), une notion complètement différente.
const reportOutputFormat = ref(reportPrefs.outputFormat)
const reportOrientation = ref(reportPrefs.orientation)
// Tri multi-critères du tableau détaillé du rapport (ex: Genre puis Artiste
// puis Titre) — une liste ordonnée de règles appliquées comme un ORDER BY
// SQL à plusieurs colonnes : la 1ère règle départage en premier, la 2e ne
// s'applique qu'en cas d'égalité sur la 1ère, etc.
const reportSortRules = ref(reportPrefs.sortRules.map((rule) => ({ ...rule })))

const sortFieldOptions = [
  { value: 'title', label: 'Titre' },
  { value: 'artist_name', label: 'Artiste' },
  { value: 'genre_name', label: 'Genre' },
  { value: 'format_name', label: 'Format' },
  { value: 'release_year', label: 'Année' },
  { value: 'price', label: 'Prix' },
  { value: 'created_at', label: "Date d'ajout" },
]

const sortDirectionOptions = [
  { value: 'asc', label: 'Croissant' },
  { value: 'desc', label: 'Décroissant' },
]

function availableSortFields(index) {
  const usedElsewhere = new Set(
    reportSortRules.value.filter((_, i) => i !== index).map((rule) => rule.field)
  )
  return sortFieldOptions.filter((option) => !usedElsewhere.has(option.value))
}

function addSortRule() {
  const unused = sortFieldOptions.find(
    (option) => !reportSortRules.value.some((rule) => rule.field === option.value)
  )
  if (!unused) return
  reportSortRules.value.push({ field: unused.value, direction: 'asc' })
}

function removeSortRule(index) {
  if (reportSortRules.value.length <= 1) return
  reportSortRules.value.splice(index, 1)
}

const outputFormatOptions = [
  { value: 'pdf', label: 'PDF' },
  { value: 'xlsx', label: 'Excel (XLSX)' },
  { value: 'docx', label: 'Word (DOCX)' },
  { value: 'csv', label: 'CSV' },
]

const orientationOptions = [
  { value: 'portrait', label: 'Portrait' },
  { value: 'landscape', label: 'Paysage' },
]

watch(
  [reportFontFamily, reportFontSize, reportIncludeTracklist, reportOutputFormat, reportOrientation, reportSortRules],
  () => {
    localStorage.setItem(
      REPORT_PREFS_KEY,
      JSON.stringify({
        fontFamily: reportFontFamily.value,
        fontSize: reportFontSize.value,
        includeTracklist: reportIncludeTracklist.value,
        outputFormat: reportOutputFormat.value,
        orientation: reportOrientation.value,
        sortRules: reportSortRules.value,
      })
    )
  },
  { deep: true }
)

// ⚠️ jsPDF n'embarque que ces 3 familles sans avoir à fournir un fichier de
// police (.ttf) à intégrer au document — Helvetica/Times/Courier sont les
// seules polices "standard PDF" disponibles nativement.
const fontFamilyOptions = [
  { value: 'helvetica', label: 'Helvetica' },
  { value: 'times', label: 'Times' },
  { value: 'courier', label: 'Courier' },
]

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

function compareSortField(a, b, field) {
  if (field === 'release_year') {
    return (a.release_year || 0) - (b.release_year || 0)
  }
  if (field === 'price') {
    const priceA = a.price != null && a.price !== '' ? parseFloat(a.price) : -Infinity
    const priceB = b.price != null && b.price !== '' ? parseFloat(b.price) : -Infinity
    return priceA - priceB
  }
  if (field === 'created_at') {
    return new Date(a.created_at) - new Date(b.created_at)
  }
  return String(a[field] || '').localeCompare(String(b[field] || ''), 'fr', { sensitivity: 'base' })
}

const reportRows = computed(() => {
  const rules = reportSortRules.value.length > 0 ? reportSortRules.value : DEFAULT_SORT_RULES
  return [...filteredDiscs.value].sort((a, b) => {
    for (const rule of rules) {
      const cmp = compareSortField(a, b, rule.field)
      if (cmp !== 0) return rule.direction === 'desc' ? -cmp : cmp
    }
    return 0
  })
})

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

  const isDefaultSort =
    reportSortRules.value.length === 1 &&
    reportSortRules.value[0].field === DEFAULT_SORT_RULES[0].field &&
    reportSortRules.value[0].direction === DEFAULT_SORT_RULES[0].direction
  if (!isDefaultSort) {
    const sortLabel = reportSortRules.value
      .map((rule) => {
        const field = sortFieldOptions.find((option) => option.value === rule.field)
        const direction = rule.direction === 'desc' ? '↓' : '↑'
        return `${field ? field.label : rule.field} ${direction}`
      })
      .join(', ')
    parts.push(`Tri : ${sortLabel}`)
  }

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
  reportIncludeTracklist.value = false
  reportFontFamily.value = 'helvetica'
  reportFontSize.value = 8
  reportOutputFormat.value = 'pdf'
  reportOrientation.value = 'portrait'
  reportSortRules.value = DEFAULT_SORT_RULES.map((rule) => ({ ...rule }))
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

// Volontairement une URL RELATIVE (pas de VITE_SERVER_BASE_URL) : Vite (dev)
// et Nginx (prod) proxient tous les deux /uploads sous la même origine que
// la page. En passant par l'URL absolue http://localhost:8080/..., le
// navigateur voit une requête cross-origin et "tainte" le canvas dès qu'on
// essaie de le relire (toDataURL) pour le redimensionner — même si l'image
// s'affiche très bien dans une simple balise <img>, qui elle ne lit jamais
// les pixels.
function getThumbnailSourceUrl(path) {
  if (path.startsWith('http')) return path
  return path.startsWith('/uploads') ? path : '/uploads/' + (path.startsWith('/') ? path.slice(1) : path)
}

// Charge la pochette d'un disque et la redimensionne côté client (canvas) en
// petite miniature JPEG avant de l'embarquer dans le rapport — sinon chaque
// pochette (parfois 500 Ko+ telle qu'uploadée) alourdirait démesurément un
// rapport qui en contient potentiellement des centaines.
async function loadCoverThumbnail(disc, maxDim = 64) {
  const path = disc.cover_url || disc.cover_image
  if (!path) return null

  try {
    const image = await new Promise((resolve, reject) => {
      const img = new Image()
      img.crossOrigin = 'anonymous'
      img.onload = () => resolve(img)
      img.onerror = () => reject(new Error('Échec de chargement de la pochette'))
      img.src = getThumbnailSourceUrl(path)
    })

    const scale = Math.min(1, maxDim / Math.max(image.naturalWidth, image.naturalHeight))
    const width = Math.max(1, Math.round(image.naturalWidth * scale))
    const height = Math.max(1, Math.round(image.naturalHeight * scale))

    const canvas = document.createElement('canvas')
    canvas.width = width
    canvas.height = height
    canvas.getContext('2d').drawImage(image, 0, 0, width, height)

    return { dataUrl: canvas.toDataURL('image/jpeg', 0.82), width, height }
  } catch {
    return null
  }
}

// Miniatures pour tous les disques du rapport, indexées par id — chargées en
// parallèle une seule fois et réutilisées par les 3 formats qui les intègrent
// (PDF/XLSX/DOCX ; le CSV est purement tabulaire et ne peut pas en contenir).
async function buildCoverThumbnails() {
  const entries = await Promise.all(
    reportRows.value.map(async (disc) => [disc.id, await loadCoverThumbnail(disc)])
  )
  return Object.fromEntries(entries)
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

async function addChartToDoc(doc, chart, x, y, maxWidth, maxHeight) {
  if (!chart) return y

  const dataUrl = chart.toBase64Image()
  const props = doc.getImageProperties(dataUrl)
  let width = maxWidth
  let height = width * (props.height / props.width)
  // En paysage la page est bien moins haute qu'en portrait — sans ce
  // plafond, un graphique dimensionné pour toute la largeur de la page
  // déborderait verticalement avant même d'atteindre le tableau détaillé.
  if (maxHeight && height > maxHeight) {
    height = maxHeight
    width = height * (props.width / props.height)
  }

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
    await viewLatestReportFile(latestReportMeta.value?.format)
  } catch (err) {
    latestReportError.value = err instanceof Error ? err.message : 'Impossible d’ouvrir le rapport.'
  }
}

function formatLabel(format) {
  return outputFormatOptions.find((option) => option.value === format)?.label || (format || '').toUpperCase()
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

// ✅ Détail des pistes, optionnel — uniquement pour les disques déjà dotés
// d'une tracklist enregistrée (has_tracks) : pas de fetch Discogs à la volée
// ici (trop lent/risqué pour une génération de rapport). Partagé par les 4
// formats de sortie.
async function buildTracksByDiscId() {
  if (!reportIncludeTracklist.value) return {}
  const discsWithTracks = reportRows.value.filter((disc) => disc.has_tracks)
  const entries = await Promise.all(
    discsWithTracks.map(async (disc) => {
      try {
        return [disc.id, await fetchTracks(disc.id)]
      } catch {
        return [disc.id, []]
      }
    })
  )
  return Object.fromEntries(entries)
}

function trackLine(track, index) {
  const position = track.position || String(index + 1)
  const duration = track.duration ? ` — ${track.duration}` : ''
  return `${position}. ${track.title || '—'}${duration}`
}

async function generatePdfBlob() {
  const doc = new jsPDF({ unit: 'mm', format: 'a4', orientation: reportOrientation.value })
  const pageWidth = doc.internal.pageSize.getWidth()
  const marginX = 14

  const fontFamily = reportFontFamily.value

  doc.setFillColor(59, 130, 246)
  doc.rect(0, 0, pageWidth, 32, 'F')
  doc.setTextColor(255, 255, 255)
  doc.setFont(fontFamily, 'bold')
  doc.setFontSize(20)
  doc.text('Rapport de collection', marginX, 18)
  doc.setFont(fontFamily, 'normal')
  doc.setFontSize(10)
  const generatedAtLabel = new Intl.DateTimeFormat('fr-FR', {
    dateStyle: 'long',
    timeStyle: 'short',
  }).format(new Date())
  doc.text(`Généré le ${generatedAtLabel}`, marginX, 26)

  let cursorY = 42
  doc.setTextColor(58, 63, 51)
  doc.setFont(fontFamily, 'bold')
  doc.setFontSize(11)
  doc.text('Critères appliqués', marginX, cursorY)
  cursorY += 6
  doc.setFont(fontFamily, 'normal')
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
    doc.setFont(fontFamily, 'normal')
    doc.setFontSize(7.5)
    doc.text(kpi.label, x + 3, y + 6)
    doc.setTextColor(40, 46, 38)
    doc.setFont(fontFamily, 'bold')
    doc.setFontSize(11)
    doc.text(kpi.value, x + 3, y + 15)
  })

  const kpiRows = Math.ceil(kpis.length / cardsPerRow)
  cursorY += kpiRows * (cardHeight + cardGap) + 6

  const halfWidth = (pageWidth - marginX * 2 - 6) / 2
  // En paysage la page fait bien moins de hauteur qu'en portrait — sans ce
  // plafond, des graphiques dimensionnés pour toute la largeur disponible
  // déborderaient verticalement de la page avant même le tableau détaillé.
  const pageHeight = doc.internal.pageSize.getHeight()
  const bottomMargin = 15

  if (genreBreakdownChart && countryBreakdownChart) {
    const pieMaxHeight = pageHeight - cursorY - bottomMargin
    const bottom1 = await addChartToDoc(doc, genreBreakdownChart, marginX, cursorY, halfWidth, pieMaxHeight)
    const bottom2 = await addChartToDoc(doc, countryBreakdownChart, marginX + halfWidth + 6, cursorY, halfWidth, pieMaxHeight)
    cursorY = Math.max(bottom1, bottom2) + 8
  } else if (genreBreakdownChart) {
    cursorY = (await addChartToDoc(doc, genreBreakdownChart, marginX, cursorY, pageWidth - marginX * 2, pageHeight - cursorY - bottomMargin)) + 8
  } else if (countryBreakdownChart) {
    cursorY = (await addChartToDoc(doc, countryBreakdownChart, marginX, cursorY, pageWidth - marginX * 2, pageHeight - cursorY - bottomMargin)) + 8
  }

  if (evolutionChart) {
    cursorY = (await addChartToDoc(doc, evolutionChart, marginX, cursorY, pageWidth - marginX * 2, pageHeight - cursorY - bottomMargin)) + 8
  }

  doc.addPage()
  doc.setTextColor(40, 46, 38)
  doc.setFont(fontFamily, 'bold')
  doc.setFontSize(13)
  doc.text('Détail des disques', marginX, 16)

  // ✅ Détail des pistes, optionnel — insérées comme lignes indentées en
  // italique juste sous la ligne de leur disque (pas dans un tableau à part).
  const tracksByDiscId = await buildTracksByDiscId()
  const coverThumbnails = await buildCoverThumbnails()

  const tableBody = []
  reportRows.value.forEach((disc) => {
    const row = [
      '',
      disc.title || '—',
      disc.artist_name || '—',
      disc.genre_name || '—',
      disc.format_name || '—',
      disc.release_year ? String(disc.release_year) : '—',
      disc.price != null && disc.price !== '' ? formatCurrencyForPdf(disc.price) : '—',
      formatTableDate(disc.created_at),
    ]
    // autoTable transmet la même référence de tableau à didDrawCell (data.row.raw)
    // — on s'en sert pour retrouver la miniature à dessiner sur cette ligne.
    row._discId = disc.id
    tableBody.push(row)

    const tracks = tracksByDiscId[disc.id]
    if (tracks && tracks.length > 0) {
      tracks.forEach((track, index) => {
        tableBody.push([
          {
            content: trackLine(track, index),
            colSpan: 8,
            styles: {
              font: fontFamily,
              fontStyle: 'italic',
              fontSize: Math.max(reportFontSize.value - 1, 4),
              textColor: [120, 126, 118],
              halign: 'left',
              cellPadding: { top: 1, right: 3, bottom: 1, left: 10 },
            },
          },
        ])
      })
    }
  })

  autoTable(doc, {
    startY: 22,
    head: [['Jaquette', 'Titre', 'Artiste', 'Genre', 'Format', 'Année', 'Prix', "Date d'ajout"]],
    body: tableBody,
    styles: { font: fontFamily, fontSize: reportFontSize.value, cellPadding: 2.4 },
    headStyles: { fillColor: [59, 130, 246], textColor: 255, halign: 'center' },
    alternateRowStyles: { fillColor: [244, 245, 242] },
    columnStyles: {
      0: { cellWidth: 20, minCellWidth: 20, minCellHeight: 14 },
      6: { halign: 'right', cellWidth: 20, minCellWidth: 20 },
    },
    margin: { left: marginX, right: marginX },
    didDrawCell: (data) => {
      if (data.section !== 'body' || data.column.index !== 0) return
      const discId = data.row.raw?._discId
      const thumb = discId != null ? coverThumbnails[discId] : null
      if (!thumb) return

      const scale = Math.min((data.cell.width - 2) / thumb.width, (data.cell.height - 2) / thumb.height)
      const w = thumb.width * scale
      const h = thumb.height * scale
      const x = data.cell.x + (data.cell.width - w) / 2
      const y = data.cell.y + (data.cell.height - h) / 2
      try {
        doc.addImage(thumb.dataUrl, 'JPEG', x, y, w, h)
      } catch {
        // pochette illisible pour jsPDF (format inattendu) — la cellule reste vide
      }
    },
  })

  const pageCount = doc.internal.getNumberOfPages()
  for (let i = 1; i <= pageCount; i += 1) {
    doc.setPage(i)
    doc.setFont(fontFamily, 'normal')
    doc.setFontSize(8)
    doc.setTextColor(150, 150, 150)
    doc.text(`Page ${i} / ${pageCount}`, pageWidth - marginX, doc.internal.pageSize.getHeight() - 8, { align: 'right' })
    doc.text('Disques Manager', marginX, doc.internal.pageSize.getHeight() - 8)
  }

  return doc.output('blob')
}

async function generateXlsxBlob() {
  const workbook = new ExcelJS.Workbook()
  workbook.creator = 'Disques Manager'
  workbook.created = new Date()
  // N'affecte que la mise en page à l'impression (une feuille Excel n'a pas
  // de "page" au sens visuel comme un PDF/DOCX tant qu'on ne l'imprime pas).
  const addWorksheet = (name) => {
    const sheet = workbook.addWorksheet(name)
    sheet.pageSetup = { orientation: reportOrientation.value, fitToPage: true }
    return sheet
  }

  const summarySheet = addWorksheet('Résumé')
  summarySheet.columns = [{ width: 28 }, { width: 42 }]
  summarySheet.addRow(['Rapport de collection']).font = { bold: true, size: 16 }
  summarySheet.addRow([
    `Généré le ${new Intl.DateTimeFormat('fr-FR', { dateStyle: 'long', timeStyle: 'short' }).format(new Date())}`,
  ])
  summarySheet.addRow([])
  summarySheet.addRow(['Critères', criteriaDescription.value])
  summarySheet.addRow([])
  summarySheet.addRow(['Indicateur', 'Valeur']).font = { bold: true }
  summarySheet.addRow(['Nombre de disques', discCount.value])
  summarySheet.addRow(['Valeur de la collection (€)', Number(totalValue.value.toFixed(2))])
  summarySheet.addRow(['Prix moyen (€)', Number(averagePrice.value.toFixed(2))])
  summarySheet.addRow(["Nombre d'artistes", artistCount.value])

  const addBreakdownSheet = (name, columnLabel, breakdown) => {
    const sheet = addWorksheet(name)
    sheet.columns = [{ header: columnLabel, width: 30 }, { header: 'Nombre de disques', width: 20 }]
    sheet.getRow(1).font = { bold: true }
    breakdown.labels.forEach((label, i) => sheet.addRow([label, breakdown.values[i]]))
  }
  addBreakdownSheet('Répartition genre', 'Genre', genreBreakdown.value)
  addBreakdownSheet('Répartition pays', 'Pays', countryBreakdown.value)

  const evolutionSheet = addWorksheet('Évolution mensuelle')
  evolutionSheet.columns = [{ header: 'Mois', width: 15 }, { header: 'Disques ajoutés', width: 20 }]
  evolutionSheet.getRow(1).font = { bold: true }
  monthlyAcquisitions.value.labels.forEach((label, i) => evolutionSheet.addRow([label, monthlyAcquisitions.value.values[i]]))

  const tracksByDiscId = await buildTracksByDiscId()
  const coverThumbnails = await buildCoverThumbnails()
  const includeTracks = reportIncludeTracklist.value

  const detailHeaders = ['Jaquette', 'Titre', 'Artiste', 'Genre', 'Format', 'Année', 'Prix (€)', "Date d'ajout"]
  if (includeTracks) detailHeaders.push('Pistes')
  const detailSheet = addWorksheet('Détail')
  detailSheet.columns = detailHeaders.map((header) => ({
    header,
    width: header === 'Jaquette' ? 10 : header === 'Pistes' ? 60 : 22,
  }))
  detailSheet.getRow(1).font = { bold: true }

  reportRows.value.forEach((disc) => {
    const row = [
      '',
      disc.title || '—',
      disc.artist_name || '—',
      disc.genre_name || '—',
      disc.format_name || '—',
      disc.release_year || '—',
      disc.price != null && disc.price !== '' ? Number(parseFloat(disc.price).toFixed(2)) : null,
      formatTableDate(disc.created_at),
    ]
    if (includeTracks) {
      const tracks = tracksByDiscId[disc.id] || []
      row.push(tracks.map((track, index) => trackLine(track, index)).join('\n'))
    }
    const addedRow = detailSheet.addRow(row)
    addedRow.height = 50
    if (includeTracks) addedRow.getCell(detailHeaders.length).alignment = { wrapText: true, vertical: 'top' }

    const thumb = coverThumbnails[disc.id]
    if (thumb) {
      const imageId = workbook.addImage({ base64: thumb.dataUrl, extension: 'jpeg' })
      detailSheet.addImage(imageId, {
        tl: { col: 0.08, row: addedRow.number - 1 + 0.08 },
        ext: { width: thumb.width, height: thumb.height },
      })
    }
  })

  const buffer = await workbook.xlsx.writeBuffer()
  return new Blob([buffer], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  })
}

function chartImageBytes(chart) {
  if (!chart) return null
  const dataUrl = chart.toBase64Image()
  return dataUrlToBytes(dataUrl)
}

function dataUrlToBytes(dataUrl) {
  const base64 = dataUrl.split(',')[1]
  const binary = atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i += 1) bytes[i] = binary.charCodeAt(i)
  return bytes
}

async function generateDocxBlob() {
  const children = []

  children.push(new Paragraph({ text: 'Rapport de collection', heading: HeadingLevel.TITLE }))
  children.push(
    new Paragraph({
      text: `Généré le ${new Intl.DateTimeFormat('fr-FR', { dateStyle: 'long', timeStyle: 'short' }).format(new Date())}`,
    })
  )
  children.push(new Paragraph({ text: `Critères : ${criteriaDescription.value}`, spacing: { after: 200 } }))

  const kpiRows = [
    ['Nombre de disques', String(discCount.value)],
    ['Valeur de la collection', formatCurrency(totalValue.value)],
    ['Prix moyen', formatCurrency(averagePrice.value)],
    ["Nombre d'artistes", String(artistCount.value)],
  ].map(
    ([label, value]) =>
      new TableRow({
        children: [
          new TableCell({ children: [new Paragraph({ children: [new TextRun({ text: label, bold: true })] })] }),
          new TableCell({ children: [new Paragraph(value)] }),
        ],
      })
  )
  children.push(new Table({ rows: kpiRows, width: { size: 100, type: WidthType.PERCENTAGE } }))

  for (const chart of [genreBreakdownChart, countryBreakdownChart, evolutionChart]) {
    const bytes = chartImageBytes(chart)
    if (!bytes) continue
    children.push(
      new Paragraph({
        spacing: { before: 200 },
        children: [new ImageRun({ data: bytes, type: 'png', transformation: { width: 420, height: 250 } })],
      })
    )
  }

  children.push(
    new Paragraph({ text: 'Détail des disques', heading: HeadingLevel.HEADING_1, spacing: { before: 300, after: 120 } })
  )

  const headers = ['Jaquette', 'Titre', 'Artiste', 'Genre', 'Format', 'Année', 'Prix', "Date d'ajout"]
  const headerRow = new TableRow({
    children: headers.map(
      (header) => new TableCell({ children: [new Paragraph({ children: [new TextRun({ text: header, bold: true })] })] })
    ),
  })

  const tracksByDiscId = await buildTracksByDiscId()
  const coverThumbnails = await buildCoverThumbnails()
  const bodyRows = []
  reportRows.value.forEach((disc) => {
    const thumb = coverThumbnails[disc.id]
    const coverCell = new TableCell({
      children: [
        new Paragraph(
          thumb
            ? {
                children: [
                  new ImageRun({
                    data: dataUrlToBytes(thumb.dataUrl),
                    type: 'jpg',
                    transformation: { width: thumb.width, height: thumb.height },
                  }),
                ],
              }
            : {}
        ),
      ],
    })

    bodyRows.push(
      new TableRow({
        children: [
          coverCell,
          ...[
            disc.title || '—',
            disc.artist_name || '—',
            disc.genre_name || '—',
            disc.format_name || '—',
            disc.release_year ? String(disc.release_year) : '—',
            disc.price != null && disc.price !== '' ? formatCurrency(disc.price) : '—',
            formatTableDate(disc.created_at),
          ].map((text) => new TableCell({ children: [new Paragraph(text)] })),
        ],
      })
    )

    const tracks = tracksByDiscId[disc.id]
    if (tracks && tracks.length > 0) {
      tracks.forEach((track, index) => {
        bodyRows.push(
          new TableRow({
            children: [
              new TableCell({
                columnSpan: headers.length,
                children: [new Paragraph({ children: [new TextRun({ text: trackLine(track, index), italics: true, size: 18 })] })],
              }),
            ],
          })
        )
      })
    }
  })

  children.push(new Table({ rows: [headerRow, ...bodyRows], width: { size: 100, type: WidthType.PERCENTAGE } }))

  const document = new Document({
    sections: [
      {
        properties: {
          page: {
            size: {
              orientation: reportOrientation.value === 'landscape' ? PageOrientation.LANDSCAPE : PageOrientation.PORTRAIT,
            },
          },
        },
        children,
      },
    ],
  })
  return Packer.toBlob(document)
}

function csvEscape(value) {
  const text = String(value ?? '')
  return /[",;\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text
}

async function generateCsvBlob() {
  const headers = ['Titre', 'Artiste', 'Genre', 'Format', 'Année', 'Prix', "Date d'ajout"]
  const lines = [headers.map(csvEscape).join(';')]

  const tracksByDiscId = await buildTracksByDiscId()

  reportRows.value.forEach((disc) => {
    lines.push(
      [
        disc.title || '',
        disc.artist_name || '',
        disc.genre_name || '',
        disc.format_name || '',
        disc.release_year || '',
        disc.price != null && disc.price !== '' ? disc.price : '',
        formatTableDate(disc.created_at),
      ]
        .map(csvEscape)
        .join(';')
    )

    const tracks = tracksByDiscId[disc.id]
    if (tracks && tracks.length > 0) {
      tracks.forEach((track, index) => {
        lines.push(['', '', '', '', '', '', trackLine(track, index)].map(csvEscape).join(';'))
      })
    }
  })

  // BOM UTF-8 pour qu'Excel détecte l'encodage correctement à l'ouverture.
  return new Blob(['﻿' + lines.join('\r\n')], { type: 'text/csv;charset=utf-8' })
}

async function generateReport() {
  isGenerating.value = true
  generationError.value = ''
  generationSuccess.value = ''

  try {
    await renderCharts()
    await nextTick()

    const format = reportOutputFormat.value
    let blob
    if (format === 'xlsx') blob = await generateXlsxBlob()
    else if (format === 'docx') blob = await generateDocxBlob()
    else if (format === 'csv') blob = await generateCsvBlob()
    else blob = await generatePdfBlob()

    const meta = await uploadGeneratedReport(blob, criteriaDescription.value, format)
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
        <p class="subtitle">Génère un état complet de ta collection (PDF, Excel, Word ou CSV) : indicateurs, répartitions et détail des disques.</p>
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
            <div>
              <p class="eyebrow">Format</p>
              <p class="detail-value">{{ formatLabel(latestReportMeta.format) }}</p>
            </div>
          </div>

          <button class="primary-btn" type="button" @click="openLatestReport">
            {{ latestReportMeta.format === 'pdf' ? 'Ouvrir le PDF' : `Télécharger (${formatLabel(latestReportMeta.format)})` }}
          </button>
        </template>

        <p v-else class="reports-muted">Aucun état n'a encore été généré.</p>

        <p v-if="latestReportError" class="form-error">{{ latestReportError }}</p>
      </section>

      <section class="panel reports-card">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Nouveau rapport</p>
            <h2>Filtres des disques</h2>
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
      </section>

      <section class="panel reports-card">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Nouveau rapport</p>
            <h2>Modalités d'impression</h2>
          </div>
        </div>

        <div class="reports-field-grid">
          <label class="form-field">
            <span>Format d'export</span>
            <select v-model="reportOutputFormat">
              <option v-for="option in outputFormatOptions" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>

          <label v-if="reportOutputFormat !== 'csv'" class="form-field">
            <span>Orientation</span>
            <select v-model="reportOrientation">
              <option v-for="option in orientationOptions" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>

          <template v-if="reportOutputFormat === 'pdf'">
            <label class="form-field">
              <span>Police</span>
              <select v-model="reportFontFamily">
                <option v-for="option in fontFamilyOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </label>

            <label class="form-field">
              <span>Taille de police (pt)</span>
              <input v-model.number="reportFontSize" type="number" min="4" max="72" step="0.5" />
            </label>
          </template>

          <label class="form-field form-field-checkbox reports-tracklist-field">
            <input v-model="reportIncludeTracklist" type="checkbox" />
            <span>Inclure le détail des pistes</span>
          </label>
        </div>

        <p v-if="reportOutputFormat === 'csv'" class="reports-muted reports-format-note">
          ℹ️ Le format CSV ne contient que le tableau détaillé des disques (indicateurs et graphiques non inclus, format tabulaire).
        </p>

        <div class="reports-sort-section">
          <p class="eyebrow">Tri du tableau détaillé</p>
          <div class="reports-sort-rules">
            <div v-for="(rule, index) in reportSortRules" :key="index" class="reports-sort-rule">
              <span class="reports-sort-rank">{{ index + 1 }}</span>
              <select v-model="rule.field" class="reports-sort-field">
                <option v-for="option in availableSortFields(index)" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
              <select v-model="rule.direction" class="reports-sort-direction">
                <option v-for="option in sortDirectionOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
              <button
                type="button"
                class="icon-action-btn"
                :disabled="reportSortRules.length <= 1"
                title="Retirer ce critère"
                @click="removeSortRule(index)"
              >
                <span>&times;</span>
              </button>
            </div>
          </div>
          <button
            type="button"
            class="ghost-btn reports-sort-add"
            :disabled="reportSortRules.length >= sortFieldOptions.length"
            @click="addSortRule"
          >
            + Ajouter un critère de tri
          </button>
        </div>
      </section>

      <section class="panel reports-card">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Nouveau rapport</p>
            <h2>Aperçu et génération</h2>
          </div>
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
            {{ isGenerating ? 'Génération…' : `📄 Générer le rapport (${formatLabel(reportOutputFormat)})` }}
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
  grid-template-columns: repeat(4, minmax(0, 1fr));
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

.reports-sort-section {
  margin-top: 1.2rem;
  padding-top: 1.2rem;
  border-top: 1px solid var(--line-soft);
}

.reports-sort-rules {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  margin-top: 0.6rem;
}

.reports-sort-rule {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.reports-sort-rank {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.6rem;
  height: 1.6rem;
  flex-shrink: 0;
  border-radius: 999px;
  background: rgba(var(--tint-rgb), 0.08);
  color: var(--text-dim);
  font-size: 0.8rem;
  font-weight: 700;
}

.reports-sort-field {
  flex: 2;
  min-width: 140px;
}

.reports-sort-direction {
  flex: 1;
  min-width: 120px;
}

.reports-sort-field,
.reports-sort-direction {
  padding: 0.55rem 0.7rem;
  border-radius: 10px;
  border: 1px solid var(--line);
  background: rgba(var(--tint-rgb), 0.04);
  color: var(--text);
}

.reports-sort-add {
  margin-top: 0.8rem;
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
