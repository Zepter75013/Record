<script setup>
import { ref, nextTick } from 'vue'
import { useBarcodeScanner } from '@/composables/useBarcodeScanner'
import { useClipboardWatch } from '@/composables/useClipboardWatch'
import BarcodeScanner from '@/components/BarcodeScanner.vue'

const result = ref(null)

function handleResult(outcome) {
  const bc = barcode.value
  if (outcome.exists) {
    result.value = { barcode: bc, exists: true, disc: outcome.disc }
  } else if (outcome.data) {
    // Pas dans la collection, mais Discogs reconnaît ce code-barres — on
    // affiche l'info trouvée pour aider à identifier le disque en rayon.
    result.value = { barcode: bc, exists: false, discogs: outcome.data }
  } else {
    result.value = { barcode: bc, exists: false, discogs: null }
  }
}

const {
  barcode,
  scannerDetected,
  scanInProgress,
  searchByBarcode,
  handleBarcodeFocus,
  handleBarcodeEnter,
  handleFileImport,
} = useBarcodeScanner(handleResult)

const barcodeInput = ref(null)
const fileImportInput = ref(null)
const isBarcodeScannerOpen = ref(false)

const clipboard = useClipboardWatch((detected) => {
  barcode.value = detected
  searchByBarcode(detected)
})

function openBarcodeScanner() {
  isBarcodeScannerOpen.value = true
}

function handleBarcodeScanned(code) {
  isBarcodeScannerOpen.value = false
  barcode.value = code
  searchByBarcode(code)
}

function triggerFileImport() {
  fileImportInput.value?.click()
}

function copyBarcode() {
  clipboard.writeToClipboard(result.value?.barcode || '')
}

function resetSearch() {
  result.value = null
  barcode.value = ''
  nextTick(() => barcodeInput.value?.focus())
}
</script>

<template>
  <div class="check-disc-view">
    <header class="page-header">
      <div class="header-wrapper">
        <div class="header-top-row">
          <div class="title-section">
            <button @click="$router.push('/dashboard')" class="back-button ghost-btn">
              <span class="icon">🏠</span>
            </button>
            <span class="title-icon">🔎</span>
            <h1>Vérifier un disque</h1>
          </div>
        </div>
        <p class="subtitle">
          Scannez ou saisissez un code-barres pour savoir si ce disque est déjà dans votre collection —
          avant de l'acheter en double.
        </p>
      </div>
    </header>

    <div class="check-disc-content">
      <section class="panel check-card">
        <div class="barcode-input-container">
          <div class="barcode-input-group">
            <input
              ref="barcodeInput"
              v-model="barcode"
              type="text"
              placeholder="Ex: 0196587842024 ou scannez avec une douchette"
              class="barcode-input"
              :class="{ 'scanner-input': scannerDetected || clipboard.isWatching.value }"
              @focus="handleBarcodeFocus"
              @keyup.enter="handleBarcodeEnter"
              :disabled="scanInProgress"
            />
            <button type="button" @click="openBarcodeScanner" class="scan-button" title="Scanner avec la caméra" :disabled="scanInProgress">
              📷
            </button>
            <button type="button" @click="triggerFileImport" class="scan-button" title="Importer depuis un fichier" :disabled="scanInProgress">
              📄
            </button>
            <button
              type="button"
              @click="clipboard.toggleWatching"
              :class="['scan-button', { active: clipboard.isWatching.value }]"
              :title="clipboard.isWatching.value ? 'Arrêter la surveillance' : 'Mode Scan Auto (Presse-papiers)'"
              :disabled="scanInProgress"
            >
              {{ clipboard.isWatching.value ? '⏸️' : '📋' }}
            </button>
            <input ref="fileImportInput" type="file" accept=".txt" style="display: none" @change="handleFileImport" />
            <button @click="handleBarcodeEnter" :disabled="!barcode || scanInProgress" class="primary-btn search-button">
              <span v-if="scanInProgress" class="btn-spinner"></span>
              <span v-else>🔍 Vérifier</span>
            </button>
          </div>
          <p class="scan-hint">📷 = Caméra&nbsp;&nbsp;|&nbsp;&nbsp;📄 = Fichier&nbsp;&nbsp;|&nbsp;&nbsp;📋 = Scan auto&nbsp;&nbsp;|&nbsp;&nbsp;Douchette = automatique</p>
        </div>

        <div v-if="result" class="check-result" :class="result.exists ? 'is-found' : 'is-missing'">
          <div class="result-icon">{{ result.exists ? '✅' : '🆕' }}</div>

          <div class="result-body">
            <template v-if="result.exists">
              <h3>Vous possédez déjà ce disque</h3>
              <p class="result-title">{{ result.disc?.title }}</p>
              <p class="result-artist">{{ result.disc?.artist_name }}</p>
            </template>
            <template v-else-if="result.discogs">
              <h3>Pas encore dans votre collection</h3>
              <p class="result-title">{{ result.discogs.title }}</p>
              <p class="result-artist">{{ result.discogs.artist }}</p>
              <p class="result-note">Identifié via Discogs — vous pouvez l'ajouter depuis "Liste des disques".</p>
            </template>
            <template v-else>
              <h3>Ce disque n'est pas dans votre collection</h3>
              <p class="result-note">Discogs ne reconnaît pas non plus ce code-barres.</p>
            </template>
            <p class="result-barcode">Code-barres : {{ result.barcode }}</p>

            <div class="result-actions">
              <button type="button" class="ghost-btn" @click="copyBarcode">📋 Copier le code-barres</button>
              <RouterLink to="/dashboard/vinyls" class="ghost-btn">💿 Liste des disques</RouterLink>
              <button type="button" class="primary-btn" @click="resetSearch">🔄 Vérifier un autre disque</button>
            </div>
          </div>
        </div>
      </section>
    </div>

    <BarcodeScanner :is-open="isBarcodeScannerOpen" @close="isBarcodeScannerOpen = false" @barcode-detected="handleBarcodeScanned" />
  </div>
</template>

<style scoped>
.check-disc-view {
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
  .check-disc-view { padding: 10px; }
}

.title-icon { font-size: 2em; }
.page-header h1 { color: var(--text); font-size: 2em; margin: 0; font-weight: bold; }
.subtitle { color: var(--text-soft); margin: 0 0 20px 0; font-size: 1.1em; }

.check-disc-content {
  display: grid;
  gap: 0.9rem;
  max-width: 640px;
}

.check-card {
  padding: 1.4rem;
}

.barcode-input-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.barcode-input-group {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.barcode-input {
  flex: 1;
  min-width: 220px;
  padding: 0.7rem 0.9rem;
  border-radius: 10px;
  border: 1px solid var(--line);
  background: rgba(var(--tint-rgb), 0.04);
  color: var(--text);
  font-size: 1em;
}

.barcode-input:focus {
  outline: 2px solid var(--accent);
  outline-offset: 1px;
}

.barcode-input.scanner-input {
  border-color: var(--accent);
  background: rgba(var(--tint-rgb), 0.08);
}

.scan-button {
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  border: 1px solid var(--line);
  border-radius: 10px;
  background: rgba(var(--tint-rgb), 0.04);
  font-size: 1.1em;
  cursor: pointer;
  transition: all 0.2s;
}

.scan-button:hover:not(:disabled) {
  background: rgba(var(--tint-rgb), 0.1);
}

.scan-button.active {
  border-color: var(--accent);
  background: rgba(var(--tint-rgb), 0.12);
}

.scan-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.search-button {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  white-space: nowrap;
}

.btn-spinner {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.4);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.scan-hint {
  margin: 0;
  color: var(--text-dim);
  font-size: 0.82rem;
}

.check-result {
  margin-top: 1.4rem;
  padding-top: 1.4rem;
  border-top: 1px solid var(--line-soft);
  display: flex;
  gap: 1rem;
}

.result-icon {
  font-size: 2.2em;
  line-height: 1;
  flex-shrink: 0;
}

.result-body {
  flex: 1;
  min-width: 0;
}

.result-body h3 {
  margin: 0 0 0.4rem;
  color: var(--text);
  font-size: 1.15em;
}

.check-result.is-found .result-body h3 { color: var(--positive-text); }
.check-result.is-missing .result-body h3 { color: var(--accent); }

.result-title {
  margin: 0;
  font-weight: 700;
  color: var(--text);
}

.result-artist {
  margin: 0.15rem 0 0;
  color: var(--text-soft);
}

.result-note {
  margin: 0.4rem 0 0;
  color: var(--text-dim);
  font-size: 0.9em;
}

.result-barcode {
  margin: 0.6rem 0 0;
  color: var(--text-dim);
  font-size: 0.85em;
  font-variant-numeric: tabular-nums;
}

.result-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  margin-top: 1rem;
}
</style>
