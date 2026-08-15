<script setup>
import { ref, watch } from 'vue'
import { fetchTracks, fetchTracklistOnDemand, updateTracks } from '../services/tracks'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  disc: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['update:modelValue', 'tracks-updated'])

const tracks = ref([])
const editableTracks = ref([])
const isEditing = ref(false)
const isLoading = ref(false)
const isFetching = ref(false)
const isSaving = ref(false)
const error = ref(null)
const hasLoaded = ref(false)

watch(
  () => props.modelValue,
  (isOpen) => {
    if (isOpen) {
      loadTracks()
    } else {
      tracks.value = []
      error.value = null
      hasLoaded.value = false
      isEditing.value = false
    }
  }
)

async function loadTracks() {
  if (!props.disc?.id) return
  isLoading.value = true
  error.value = null
  try {
    const result = await fetchTracks(props.disc.id)
    tracks.value = result || []
  } catch (e) {
    error.value = e.message
  } finally {
    isLoading.value = false
    hasLoaded.value = true
  }
}

async function fetchFromDiscogs() {
  if (!props.disc?.id) return
  isFetching.value = true
  error.value = null
  try {
    const result = await fetchTracklistOnDemand(props.disc.id)
    tracks.value = result || []
    emit('tracks-updated', { discId: props.disc.id, hasTracks: tracks.value.length > 0 })
  } catch (e) {
    error.value = e.message
  } finally {
    isFetching.value = false
  }
}

function startEditing() {
  editableTracks.value = tracks.value.length
    ? tracks.value.map((t) => ({ position: t.position || '', title: t.title || '', duration: t.duration || '' }))
    : [{ position: '', title: '', duration: '' }]
  isEditing.value = true
}

function cancelEditing() {
  isEditing.value = false
}

function addTrackRow() {
  editableTracks.value.push({ position: '', title: '', duration: '' })
}

function removeTrackRow(index) {
  editableTracks.value.splice(index, 1)
}

async function saveTracks() {
  if (!props.disc?.id) return
  const cleaned = editableTracks.value.filter((t) => t.title && t.title.trim() !== '')
  isSaving.value = true
  error.value = null
  try {
    const result = await updateTracks(props.disc.id, cleaned)
    tracks.value = result || []
    isEditing.value = false
    emit('tracks-updated', { discId: props.disc.id, hasTracks: tracks.value.length > 0 })
  } catch (e) {
    error.value = e.message
  } finally {
    isSaving.value = false
  }
}

function closeModal() {
  emit('update:modelValue', false)
}

function getImageUrl(path, cacheBuster = null) {
  if (!path || path === 'null' || path === 'undefined') return ''
  if (path.startsWith('http')) return path
  const BASE = import.meta.env.VITE_SERVER_BASE_URL || ''
  let url = path.startsWith('/uploads') ? BASE + path : BASE + '/uploads/' + (path.startsWith('/') ? path.slice(1) : path)
  if (cacheBuster) url += (url.includes('?') ? '&' : '?') + `t=${cacheBuster}`
  return url
}
</script>

<template>
  <div v-if="modelValue && disc" class="modal-overlay" @click.self="closeModal">
    <section class="modal-card tracklist-modal-card" role="dialog" aria-modal="true" aria-labelledby="tracklist-title">
      <div class="modal-header">
        <div class="tracklist-heading">
          <img
            v-if="disc.cover_url || disc.cover_image"
            class="tracklist-cover"
            :src="getImageUrl(disc.cover_url || disc.cover_image, disc.updated_at)"
            :alt="`Pochette de ${disc.title}`"
          />
          <span v-else class="tracklist-icon" aria-hidden="true">🎵</span>
          <div>
            <p class="eyebrow">Pistes</p>
            <h2 id="tracklist-title">{{ disc.artist_name || disc.title }}</h2>
            <p v-if="disc.artist_name" class="tracklist-album">{{ disc.title }}</p>
          </div>
        </div>
      </div>

      <div v-if="isLoading" class="tracklist-state">Chargement des pistes…</div>

      <div v-else-if="error" class="tracklist-state tracklist-error">{{ error }}</div>

      <template v-else-if="isEditing">
        <ol class="tracklist-list tracklist-edit-list">
          <li v-for="(track, index) in editableTracks" :key="index" class="tracklist-edit-item">
            <input
              v-model="track.position"
              class="tracklist-input tracklist-input-position"
              type="text"
              placeholder="A1"
              aria-label="Position"
            />
            <input
              v-model="track.title"
              class="tracklist-input tracklist-input-title"
              type="text"
              placeholder="Titre de la piste"
              aria-label="Titre"
            />
            <input
              v-model="track.duration"
              class="tracklist-input tracklist-input-duration"
              type="text"
              placeholder="--:--"
              aria-label="Durée"
            />
            <button
              type="button"
              class="tracklist-remove-btn"
              :aria-label="`Supprimer la piste ${index + 1}`"
              @click="removeTrackRow(index)"
            >
              🗑️
            </button>
          </li>
        </ol>
        <button type="button" class="ghost-btn tracklist-add-btn" @click="addTrackRow">
          ➕ Ajouter une piste
        </button>
      </template>

      <ol v-else-if="tracks.length" class="tracklist-list">
        <li v-for="(track, index) in tracks" :key="index" class="tracklist-item">
          <span class="tracklist-position">{{ track.position || index + 1 }}</span>
          <span class="tracklist-title">{{ track.title }}</span>
          <span v-if="track.duration" class="tracklist-duration">{{ track.duration }}</span>
        </li>
      </ol>

      <div v-else-if="hasLoaded" class="tracklist-state tracklist-empty">
        <p>Aucune piste enregistrée pour ce disque.</p>
        <div class="tracklist-empty-actions">
          <button class="primary-btn" type="button" :disabled="isFetching" @click="fetchFromDiscogs">
            {{ isFetching ? 'Recherche…' : '🎵 Charger les pistes' }}
          </button>
          <button class="ghost-btn" type="button" @click="startEditing">✏️ Saisir manuellement</button>
        </div>
      </div>

      <div class="modal-actions">
        <template v-if="isEditing">
          <button class="ghost-btn" type="button" :disabled="isSaving" @click="cancelEditing">Annuler</button>
          <button class="primary-btn" type="button" :disabled="isSaving" @click="saveTracks">
            {{ isSaving ? 'Enregistrement…' : 'Enregistrer' }}
          </button>
        </template>
        <template v-else>
          <button class="ghost-btn" type="button" @click="closeModal">Fermer</button>
          <button v-if="hasLoaded && !error" class="primary-btn" type="button" @click="startEditing">
            ✏️ Modifier
          </button>
        </template>
      </div>
    </section>
  </div>
</template>

<style scoped>
.tracklist-modal-card {
  animation: modal-pop-in 180ms ease;
  transform-origin: center;
}

.modal-header {
  margin-bottom: 0.4rem;
}

.tracklist-heading {
  display: flex;
  align-items: flex-start;
  gap: 0.9rem;
}

.tracklist-icon {
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

.tracklist-cover {
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 8px;
  object-fit: cover;
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
  font-size: 1.35rem;
  line-height: 1.15;
  color: var(--text);
}

.tracklist-album {
  margin-top: 0.2rem;
  font-size: 0.88rem;
  color: var(--text-soft);
}

.tracklist-state {
  margin: 1rem 0;
  padding: 1rem;
  border-radius: 14px;
  background: rgba(var(--tint-rgb), 0.04);
  border: 1px solid var(--line-soft);
  color: var(--text-soft);
  text-align: center;
}

.tracklist-error {
  color: var(--negative-text);
}

.tracklist-empty {
  display: grid;
  gap: 0.85rem;
  justify-items: center;
}

.tracklist-list {
  list-style: none;
  margin: 1rem 0;
  padding: 0;
  display: grid;
  gap: 0.3rem;
  max-height: 50vh;
  overflow-y: auto;
}

.tracklist-item {
  display: grid;
  grid-template-columns: 2.5rem 1fr auto;
  align-items: center;
  gap: 0.75rem;
  padding: 0.55rem 0.7rem;
  border-radius: 10px;
  background: rgba(var(--tint-rgb), 0.03);
}

.tracklist-item:nth-child(odd) {
  background: rgba(var(--tint-rgb), 0.055);
}

.tracklist-position {
  color: var(--text-dim);
  font-size: 0.85rem;
  font-variant-numeric: tabular-nums;
}

.tracklist-title {
  color: var(--text);
}

.tracklist-duration {
  color: var(--text-dim);
  font-size: 0.85rem;
  font-variant-numeric: tabular-nums;
}

.tracklist-empty-actions {
  display: flex;
  gap: 0.6rem;
  flex-wrap: wrap;
  justify-content: center;
}

.tracklist-edit-list {
  gap: 0.4rem;
}

.tracklist-edit-item {
  display: grid;
  grid-template-columns: 3.5rem 1fr 4.5rem auto;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem;
  border-radius: 10px;
  background: rgba(var(--tint-rgb), 0.03);
}

.tracklist-input {
  width: 100%;
  border-radius: 8px;
  border: 1px solid var(--line);
  background: rgba(var(--tint-rgb), 0.04);
  color: var(--text);
  padding: 0.45rem 0.6rem;
  font-size: 0.88rem;
}

.tracklist-input:focus {
  outline: 2px solid var(--accent);
  outline-offset: 1px;
}

.tracklist-input-position,
.tracklist-input-duration {
  text-align: center;
  font-variant-numeric: tabular-nums;
}

.tracklist-remove-btn {
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 1rem;
  padding: 0.3rem;
  border-radius: 8px;
  line-height: 1;
}

.tracklist-remove-btn:hover {
  background: rgba(var(--tint-rgb), 0.08);
}

.tracklist-add-btn {
  margin-top: 0.75rem;
  width: 100%;
}

@keyframes modal-pop-in {
  from {
    opacity: 0;
    transform: translateY(8px) scale(0.985);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@media (prefers-reduced-motion: reduce) {
  .tracklist-modal-card {
    animation: none;
  }
}
</style>
