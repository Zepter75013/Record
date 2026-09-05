<!-- frontend/src/components/ArtistsBulkBiographyModal.vue -->
<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="isOpen" class="modal-overlay" @click.self="handleClose">
        <div class="modal-card bulk-card">
          <div class="modal-header">
            <div class="header-content">
              <h2>🔄 Mettre à jour les biographies</h2>
              <p class="modal-subtitle">
                Cochez les artistes à mettre à jour, la biographie sera recherchée automatiquement sur Discogs.
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
            <div class="bulk-count">{{ selectedCount }} / {{ artists.length }} sélectionné{{ selectedCount > 1 ? 's' : '' }}</div>
            <div class="bulk-toolbar-actions">
              <button type="button" @click="selectAll" class="ghost-btn small" :disabled="running">Tout cocher</button>
              <button type="button" @click="selectNone" class="ghost-btn small" :disabled="running">Tout décocher</button>
              <button type="button" @click="selectMissingOnly" class="ghost-btn small" :disabled="running">Sans biographie uniquement</button>
            </div>
          </div>

          <div class="bulk-list">
            <div v-if="artists.length === 0" class="bulk-empty">Aucun artiste.</div>
            <label v-for="artist in artists" :key="artist.id" class="bulk-row" :class="{ 'is-running': results[artist.id]?.status === 'loading' }">
              <input
                type="checkbox"
                v-model="selected[artist.id]"
                :disabled="running"
              />
              <span class="bulk-name">{{ artist.name }}</span>
              <span v-if="artist.biography" class="bulk-badge">📝 déjà renseignée</span>
              <span class="bulk-status">
                <span v-if="results[artist.id]?.status === 'loading'" class="spinner-small"></span>
                <span v-else-if="results[artist.id]?.status === 'success'" class="status-ok">✅</span>
                <span v-else-if="results[artist.id]?.status === 'not-found'" class="status-not-found">➖ aucune donnée</span>
                <span v-else-if="results[artist.id]?.status === 'error'" class="status-error" :title="results[artist.id]?.message">⚠️ {{ results[artist.id]?.message }}</span>
                <span v-else-if="results[artist.id]?.status === 'skipped'" class="status-skipped">⏭️ annulé</span>
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
import { ref, computed, watch } from 'vue';
import { useApi } from '@/composables/useApi';

const props = defineProps({
  isOpen: Boolean,
  artists: { type: Array, default: () => [] }
});
const emit = defineEmits(['close', 'artist-updated']);

const api = useApi();

const selected = ref({});
const results = ref({});
const running = ref(false);
const stopRequested = ref(false);
const hasRun = ref(false);

const initSelection = () => {
  const sel = {};
  for (const artist of props.artists) {
    sel[artist.id] = !artist.biography;
  }
  selected.value = sel;
  results.value = {};
  hasRun.value = false;
  stopRequested.value = false;
};

watch(() => props.isOpen, (open) => {
  if (open) initSelection();
});

const selectedCount = computed(() => Object.values(selected.value).filter(Boolean).length);

const selectAll = () => {
  for (const artist of props.artists) selected.value[artist.id] = true;
};
const selectNone = () => {
  for (const artist of props.artists) selected.value[artist.id] = false;
};
const selectMissingOnly = () => {
  for (const artist of props.artists) selected.value[artist.id] = !artist.biography;
};

const totalToProcess = ref(0);
const processedCount = computed(() => {
  return Object.values(results.value).filter((r) => r.status === 'success' || r.status === 'not-found' || r.status === 'error' || r.status === 'skipped').length;
});
const progressPercent = computed(() => {
  if (totalToProcess.value === 0) return 0;
  return Math.round((processedCount.value / totalToProcess.value) * 100);
});
const successCount = computed(() => Object.values(results.value).filter((r) => r.status === 'success').length);
const notFoundCount = computed(() => Object.values(results.value).filter((r) => r.status === 'not-found').length);
const errorCount = computed(() => Object.values(results.value).filter((r) => r.status === 'error').length);
const skippedCount = computed(() => Object.values(results.value).filter((r) => r.status === 'skipped').length);

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const startRun = async () => {
  const targets = props.artists.filter((artist) => selected.value[artist.id]);
  if (targets.length === 0) return;

  running.value = true;
  hasRun.value = true;
  stopRequested.value = false;
  results.value = {};
  totalToProcess.value = targets.length;

  for (const artist of targets) {
    if (stopRequested.value) {
      results.value[artist.id] = { status: 'skipped' };
      continue;
    }

    results.value[artist.id] = { status: 'loading' };
    try {
      const suggestion = await api.post(`/artists/${artist.id}/biography/suggest`);
      const biography = suggestion?.biography;
      if (!biography) {
        results.value[artist.id] = { status: 'not-found' };
      } else {
        // On préserve le pays existant : UpdateArtist remplace tous les
        // champs envoyés, un country_id absent effacerait le pays déjà
        // renseigné sur cet artiste.
        await api.put(`/artists/${artist.id}`, { name: artist.name, biography, country_id: artist.country_id ?? null });
        results.value[artist.id] = { status: 'success' };
        emit('artist-updated', { id: artist.id, biography });
      }
    } catch (error) {
      results.value[artist.id] = { status: 'error', message: error.message || 'Échec de la mise à jour' };
    }

    // Pause entre chaque artiste pour rester sous la limite de débit de
    // Discogs (60 requêtes/min authentifiées ; chaque suggestion fait 2
    // appels Discogs côté backend, qui retente lui-même en cas de 429 —
    // cette pause est une marge de sécurité en plus de ce retry).
    if (!stopRequested.value) {
      await sleep(1500);
    }
  }

  running.value = false;
};

const stopRun = () => {
  stopRequested.value = true;
};

const handleClose = () => {
  if (running.value) return;
  emit('close');
};
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
  font-size: 0.78em;
  color: var(--text-dim);
  background: rgba(var(--tint-rgb), 0.06);
  padding: 2px 8px;
  border-radius: 8px;
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
