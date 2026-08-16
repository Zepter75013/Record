<script setup>
import { ref, nextTick, watch } from 'vue'
import { useBarcodeScanner } from '@/composables/useBarcodeScanner'
import { useClipboardWatch } from '@/composables/useClipboardWatch'
import BarcodeScanner from '@/components/BarcodeScanner.vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:modelValue'])

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

function closeModal() {
  emit('update:modelValue', false)
}

watch(
  () => props.modelValue,
  (isOpen) => {
    if (isOpen) {
      resetSearch()
    } else {
      clipboard.stopWatching()
    }
  }
)
</script>

<template>
  <div v-if="modelValue" class="modal-overlay" @click.self="closeModal">
    <section class="modal-card check-disc-modal-card" role="dialog" aria-modal="true" aria-labelledby="check-disc-title">
      <div class="modal-header">
        <div class="check-disc-heading">
          <span class="check-disc-icon" aria-hidden="true">🔎</span>
          <div>
            <p class="eyebrow">Vérification</p>
            <h2 id="check-disc-title">Ce disque est-il déjà dans ma collection ?</h2>
          </div>
        </div>
        <button type="button" class="icon-action-btn" @click="closeModal" aria-label="Fermer">
          <span>&times;</span>
        </button>
      </div>

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
            <p class="result-note">Identifié via Discogs.</p>
          </template>
          <template v-else>
            <h3>Ce disque n'est pas dans votre collection</h3>
            <p class="result-note">Discogs ne reconnaît pas non plus ce code-barres.</p>
          </template>
          <p class="result-barcode">Code-barres : {{ result.barcode }}</p>

          <div class="result-actions">
            <button type="button" class="ghost-btn" @click="copyBarcode">📋 Copier le code-barres</button>
            <button type="button" class="ghost-btn" @click="resetSearch">🔄 Vérifier un autre disque</button>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button type="button" class="ghost-btn" @click="closeModal">Fermer</button>
      </div>
    </section>

    <BarcodeScanner :is-open="isBarcodeScannerOpen" @close="isBarcodeScannerOpen = false" @barcode-detected="handleBarcodeScanned" />
  </div>
</template>

<style scoped>
.check-disc-modal-card {
  max-width: 560px;
  width: 100%;
}

.check-disc-heading {
  display: flex;
  align-items: flex-start;
  gap: 0.9rem;
}

.check-disc-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 999px;
  background: rgba(var(--tint-rgb), 0.1);
  font-size: 1.3rem;
  flex-shrink: 0;
}

.eyebrow {
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 0.74rem;
}

.modal-header h2 {
  margin-top: 0.35rem;
  font-size: 1.15rem;
  line-height: 1.3;
  color: var(--text);
}

.barcode-input-container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-top: 1rem;
}

.barcode-input-group {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.barcode-input {
  flex: 1;
  min-width: 200px;
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
  font-size: 0.8rem;
}

.check-result {
  margin-top: 1.2rem;
  padding-top: 1.2rem;
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
  font-size: 1.05em;
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
