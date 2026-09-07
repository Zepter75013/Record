<!-- frontend/src/components/DiscsBulkDiscogsModal.vue
     Mise à jour groupée (tracklist + Notes Discogs) depuis Discogs — même
     principe que LabelsBulkInfoModal.vue / ArtistsBulkBiographyModal.vue :
     sélection, lancement, un statut par ligne, rien n'écrase le champ
     Commentaires personnel (seul discogs_notes est touché), et Notes
     Discogs n'est mis à jour que s'il est encore vide, pour ne jamais
     remplacer une valeur déjà validée sans action explicite. -->
<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="isOpen" class="modal-overlay" @click.self="handleClose">
        <div class="modal-card bulk-card">
          <div class="modal-header">
            <div class="header-content">
              <h2>🔄 Mettre à jour pistes / Notes Discogs</h2>
              <p class="modal-subtitle">
                Cochez les disques à mettre à jour, la tracklist et les notes de pressage seront recherchées automatiquement sur Discogs.
              </p>
            </div>
            <button
              @click="handleClose"
              class="icon-action-btn"
              :disabled="running"
              aria-label="Fermer"
            >
              ✕
            </button>
          </div>

          <div class="bulk-toolbar">
            <div class="bulk-count">{{ selectedCount }} / {{ discs.length }} sélectionné{{ selectedCount > 1 ? 's' : '' }}</div>
            <div class="bulk-toolbar-actions">
              <button type="button" @click="selectAll" class="ghost-btn small" :disabled="running">Tout cocher</button>
              <button type="button" @click="selectNone" class="ghost-btn small" :disabled="running">Tout décocher</button>
              <button type="button" @click="selectIncompleteOnly" class="ghost-btn small" :disabled="running">Incomplets uniquement</button>
            </div>
          </div>

          <div class="bulk-list">
            <div v-if="discs.length === 0" class="bulk-empty">Aucun disque.</div>
            <label v-for="disc in discs" :key="disc.id" class="bulk-row" :class="{ 'is-running': results[disc.id]?.status === 'loading' }">
              <input
                type="checkbox"
                v-model="selected[disc.id]"
                :disabled="running"
              />
              <span class="bulk-name">{{ disc.artist_name }} — {{ disc.title }}</span>
              <span v-if="disc.has_tracks" class="bulk-badge" title="Pistes déjà renseignées">🎵</span>
              <span v-if="disc.discogs_notes" class="bulk-badge" title="Notes Discogs déjà renseignées">📝</span>
              <span class="bulk-status">
                <span v-if="results[disc.id]?.status === 'loading'" class="spinner-small"></span>
                <span v-else-if="results[disc.id]?.status === 'success'" class="status-ok" :title="results[disc.id]?.detail">✅</span>
                <span v-else-if="results[disc.id]?.status === 'not-found'" class="status-not-found">➖ aucune donnée</span>
                <span v-else-if="results[disc.id]?.status === 'error'" class="status-error" :title="results[disc.id]?.message">⚠️ {{ results[disc.id]?.message }}</span>
                <span v-else-if="results[disc.id]?.status === 'skipped'" class="status-skipped">⏭️ annulé</span>
              </span>
            </label>
          </div>

          <div v-if="running" class="bulk-progress">
            <div class="bulk-progress-bar">
              <div class="bulk-progress-fill" :style="{ width: progressPercent + '%' }"></div>
            </div>
            <div class="bulk-progress-text">{{ processedCount }} / {{ totalToProcess }} traités…</div>
          </div>
          <div v-else-if="hasRun" class="bulk-summary">
            {{ successCount }} mise{{ successCount > 1 ? 's' : '' }} à jour, {{ notFoundCount }} sans donnée, {{ errorCount }} échec{{ errorCount > 1 ? 's' : '' }}<span v-if="skippedCount"> , {{ skippedCount }} annulé{{ skippedCount > 1 ? 's' : '' }}</span>.
          </div>

          <div class="modal-actions">
            <button type="button" @click="running ? stopRun() : handleClose()" class="ghost-btn">
              {{ running ? '⏹️ Arrêter' : 'Fermer' }}
            </button>
            <button type="button" @click="startRun" class="primary-btn" :disabled="selectedCount === 0 || running">
              <span v-if="running" class="spinner"></span>
              <span v-else>Lancer la mise à jour ({{ selectedCount }})</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useApi } from '@/composables/useApi'
import { previewTracklist, updateTracks } from '@/services/tracks'

const props = defineProps({
  isOpen: Boolean,
  discs: { type: Array, default: () => [] }
})
const emit = defineEmits(['close', 'disc-updated'])

const { apiFetch } = useApi()

const isIncomplete = (disc) => !disc.has_tracks || !disc.discogs_notes

const selected = ref({})
const results = ref({})
const running = ref(false)
const stopRequested = ref(false)
const hasRun = ref(false)

const initSelection = () => {
  const sel = {}
  for (const disc of props.discs) {
    sel[disc.id] = isIncomplete(disc)
  }
  selected.value = sel
  results.value = {}
  hasRun.value = false
  stopRequested.value = false
}

watch(() => props.isOpen, (open) => {
  if (open) initSelection()
})

const selectedCount = computed(() => Object.values(selected.value).filter(Boolean).length)

const selectAll = () => {
  for (const disc of props.discs) selected.value[disc.id] = true
}
const selectNone = () => {
  for (const disc of props.discs) selected.value[disc.id] = false
}
const selectIncompleteOnly = () => {
  for (const disc of props.discs) selected.value[disc.id] = isIncomplete(disc)
}

const totalToProcess = ref(0)
const processedCount = computed(() => {
  return Object.values(results.value).filter((r) => r.status === 'success' || r.status === 'not-found' || r.status === 'error' || r.status === 'skipped').length
})
const progressPercent = computed(() => {
  if (totalToProcess.value === 0) return 0
  return Math.round((processedCount.value / totalToProcess.value) * 100)
})
const successCount = computed(() => Object.values(results.value).filter((r) => r.status === 'success').length)
const notFoundCount = computed(() => Object.values(results.value).filter((r) => r.status === 'not-found').length)
const errorCount = computed(() => Object.values(results.value).filter((r) => r.status === 'error').length)
const skippedCount = computed(() => Object.values(results.value).filter((r) => r.status === 'skipped').length)

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

// "Aucune piste trouvée sur Discogs" (ou release introuvable) est un cas
// normal — pas une panne — à afficher comme "aucune donnée", jamais comme
// un échec.
const isNotFoundMessage = (message) => /aucune piste trouv|aucun résultat|introuvable/i.test(message || '')

const startRun = async () => {
  const targets = props.discs.filter((disc) => selected.value[disc.id])
  if (targets.length === 0) return

  running.value = true
  hasRun.value = true
  stopRequested.value = false
  results.value = {}
  totalToProcess.value = targets.length

  for (const disc of targets) {
    if (stopRequested.value) {
      results.value[disc.id] = { status: 'skipped' }
      continue
    }

    results.value[disc.id] = { status: 'loading' }
    try {
      const preview = await previewTracklist(disc.id)
      const tracks = preview?.tracks || []
      const notes = (preview?.notes || '').trim()

      if (!tracks.length) {
        results.value[disc.id] = { status: 'not-found' }
      } else {
        const found = [`${tracks.length} piste${tracks.length > 1 ? 's' : ''}`]
        await updateTracks(disc.id, tracks)

        let discogsNotesApplied = null
        if (notes && !disc.discogs_notes?.trim()) {
          const updated = await apiFetch(`discs/${disc.id}`, {
            method: 'PUT',
            body: JSON.stringify({
              title: disc.title,
              artist_id: disc.artist_id,
              genre_id: disc.genre_id || null,
              format_id: disc.format_id || null,
              country_id: disc.country_id || null,
              label_id: disc.label_id || null,
              release_year: disc.release_year || null,
              barcode: disc.barcode || null,
              price: disc.price ?? null,
              quantity: disc.quantity || 1,
              notes: disc.notes || null,
              isrc: disc.isrc || null,
              discogs_notes: notes
            }),
            headers: { 'Content-Type': 'application/json' }
          })
          discogsNotesApplied = updated.discogs_notes
          found.push('notes Discogs')
        }

        results.value[disc.id] = { status: 'success', detail: found.join(', ') }
        emit('disc-updated', { discId: disc.id, hasTracks: true, discogsNotes: discogsNotesApplied })
      }
    } catch (error) {
      const message = error.message || 'Échec de la mise à jour'
      if (isNotFoundMessage(message)) {
        results.value[disc.id] = { status: 'not-found' }
      } else {
        results.value[disc.id] = { status: 'error', message }
      }
    }

    // Pause entre chaque disque pour rester sous la limite de débit de
    // Discogs (le backend retente déjà lui-même en cas de 429) — même
    // marge que pour les labels/artistes.
    if (!stopRequested.value) {
      await sleep(1500)
    }
  }

  running.value = false
}

const stopRun = () => {
  stopRequested.value = true
}

const handleClose = () => {
  if (running.value) return
  emit('close')
}
</script>

<style scoped>
.modal-overlay {
  z-index: 10000;
}

.bulk-card {
  max-width: 640px;
  width: 100%;
}

.header-content {
  flex: 1;
}

.modal-header h2 {
  margin: 0 0 6px;
  font-size: 1.4em;
}

.modal-subtitle {
  margin: 0;
  color: var(--text-soft);
  font-size: 0.9em;
}

.bulk-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  margin: 16px 0 10px;
}

.bulk-count {
  font-weight: 600;
  color: var(--text-soft);
  font-size: 0.9em;
}

.bulk-toolbar-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.ghost-btn.small {
  padding: 6px 10px;
  font-size: 0.82em;
}

.bulk-list {
  max-height: 360px;
  overflow-y: auto;
  border: 1px solid var(--line-soft);
  border-radius: 12px;
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.bulk-empty {
  padding: 20px;
  text-align: center;
  color: var(--text-dim);
  font-style: italic;
}

.bulk-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
}

.bulk-row:hover {
  background: rgba(var(--tint-rgb), 0.06);
}

.bulk-row.is-running {
  background: rgba(var(--tint-rgb), 0.09);
}

.bulk-row input[type='checkbox'] {
  flex-shrink: 0;
  width: 16px;
  height: 16px;
}

.bulk-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text);
}

.bulk-badge {
  flex-shrink: 0;
  font-size: 0.85em;
}

.bulk-status {
  flex-shrink: 0;
  font-size: 0.85em;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-ok {
  color: var(--positive-text, #22c55e);
}

.status-not-found {
  color: var(--text-dim);
}

.status-error {
  color: var(--negative-text);
}

.status-skipped {
  color: var(--text-dim);
}

.spinner-small {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid rgba(var(--tint-rgb), 0.25);
  border-top: 2px solid var(--accent);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.bulk-progress {
  margin-top: 14px;
}

.bulk-progress-bar {
  width: 100%;
  height: 8px;
  background: rgba(var(--tint-rgb), 0.08);
  border-radius: 8px;
  overflow: hidden;
}

.bulk-progress-fill {
  height: 100%;
  background: var(--accent);
  transition: width 0.3s ease;
}

.bulk-progress-text {
  margin-top: 6px;
  font-size: 0.85em;
  color: var(--text-soft);
  text-align: center;
}

.bulk-summary {
  margin-top: 14px;
  padding: 10px 14px;
  background: rgba(var(--tint-rgb), 0.05);
  border-radius: 10px;
  color: var(--text-soft);
  font-size: 0.9em;
  text-align: center;
}

.modal-actions {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.35);
  border-top: 2px solid white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  display: inline-block;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: all 0.3s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-from .modal-card,
.modal-fade-leave-to .modal-card {
  transform: scale(0.95) translateY(20px);
}
</style>
