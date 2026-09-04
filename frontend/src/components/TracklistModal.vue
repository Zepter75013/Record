<script setup>
import { ref, computed, watch } from 'vue'
import { fetchTracks, fetchTracklistOnDemand, updateTracks } from '../services/tracks'
import { groupTracksByDiscSide, discSideToLetter } from '../utils/discSides'
import StreamingButtons from './StreamingButtons.vue'

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
// Édition structurée par disque puis par face (A/B) — voir startEditing().
// editableOther recueille les pistes existantes dont la position ne suit
// pas la convention lettre de face (évite de les perdre en cas d'édition).
const editableDiscs = ref([])
const editableOther = ref([])
const isEditing = ref(false)
const isLoading = ref(false)
const isFetching = ref(false)
const isSaving = ref(false)
const error = ref(null)
const hasLoaded = ref(false)

// Regroupement pour l'affichage (lecture seule) : Disque N > Face A/B.
const groupedTracks = computed(() => groupTracksByDiscSide(tracks.value))

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

function emptyDisc(discNumber) {
  return { disc: discNumber, sides: { A: [{ title: '', duration: '' }], B: [] } }
}

function startEditing() {
  const { discs, noFace } = groupedTracks.value
  editableDiscs.value = discs.length
    ? discs.map((d) => ({
        disc: d.disc,
        sides: {
          A: d.sides
            .find((s) => s.side === 'A')
            .tracks.map((t) => ({ title: t.title || '', duration: t.duration || '' })),
          B: d.sides
            .find((s) => s.side === 'B')
            .tracks.map((t) => ({ title: t.title || '', duration: t.duration || '' })),
        },
      }))
    : [emptyDisc(1)]
  editableOther.value = noFace.map((t) => ({
    position: t.position || '',
    title: t.title || '',
    duration: t.duration || '',
  }))
  isEditing.value = true
}

function cancelEditing() {
  isEditing.value = false
}

function addDisc() {
  const nextNumber = editableDiscs.value.reduce((max, d) => Math.max(max, d.disc), 0) + 1
  editableDiscs.value.push(emptyDisc(nextNumber))
}

function removeDisc(discIndex) {
  const d = editableDiscs.value[discIndex]
  const hasContent = [...d.sides.A, ...d.sides.B].some((t) => t.title && t.title.trim() !== '')
  if (hasContent && !window.confirm(`Supprimer le disque ${d.disc} et toutes ses pistes ?`)) return
  editableDiscs.value.splice(discIndex, 1)
}

function addTrack(discIndex, side) {
  editableDiscs.value[discIndex].sides[side].push({ title: '', duration: '' })
}

function removeTrack(discIndex, side, trackIndex) {
  editableDiscs.value[discIndex].sides[side].splice(trackIndex, 1)
}

function addOtherTrack() {
  editableOther.value.push({ position: '', title: '', duration: '' })
}

function removeOtherTrack(index) {
  editableOther.value.splice(index, 1)
}

async function saveTracks() {
  if (!props.disc?.id) return
  const flattened = []
  for (const d of editableDiscs.value) {
    for (const side of ['A', 'B']) {
      d.sides[side].forEach((t, i) => {
        if (t.title && t.title.trim() !== '') {
          flattened.push({
            position: discSideToLetter(d.disc, side) + (i + 1),
            title: t.title.trim(),
            duration: t.duration || '',
          })
        }
      })
    }
  }
  for (const t of editableOther.value) {
    if (t.title && t.title.trim() !== '') {
      flattened.push({ position: t.position || '', title: t.title.trim(), duration: t.duration || '' })
    }
  }
  isSaving.value = true
  error.value = null
  try {
    const result = await updateTracks(props.disc.id, flattened)
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
        <div class="disc-groups">
          <div v-for="(d, discIndex) in editableDiscs" :key="discIndex" class="disc-group">
            <div class="disc-group-header">
              <h3>Disque {{ d.disc }}</h3>
              <button
                type="button"
                class="tracklist-remove-btn"
                :aria-label="`Supprimer le disque ${d.disc}`"
                @click="removeDisc(discIndex)"
              >
                🗑️
              </button>
            </div>
            <div class="side-columns">
              <div v-for="side in ['A', 'B']" :key="side" class="side-column">
                <h4>Face {{ discSideToLetter(d.disc, side) }}</h4>
                <ol class="tracklist-list tracklist-edit-list">
                  <li v-for="(track, trackIndex) in d.sides[side]" :key="trackIndex" class="tracklist-edit-item">
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
                      :aria-label="`Supprimer la piste ${trackIndex + 1}`"
                      @click="removeTrack(discIndex, side, trackIndex)"
                    >
                      🗑️
                    </button>
                  </li>
                </ol>
                <button type="button" class="ghost-btn tracklist-add-btn small" @click="addTrack(discIndex, side)">
                  ➕ Piste
                </button>
              </div>
            </div>
          </div>
        </div>
        <button type="button" class="ghost-btn tracklist-add-btn" @click="addDisc">
          ➕ Ajouter un disque
        </button>

        <template v-if="editableOther.length">
          <h3 class="other-tracks-title">Autres pistes (position libre)</h3>
          <ol class="tracklist-list tracklist-edit-list">
            <li v-for="(track, index) in editableOther" :key="index" class="tracklist-edit-item tracklist-edit-item-other">
              <input
                v-model="track.position"
                class="tracklist-input tracklist-input-position"
                type="text"
                placeholder="Ex: 1"
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
                @click="removeOtherTrack(index)"
              >
                🗑️
              </button>
            </li>
          </ol>
        </template>
      </template>

      <template v-else-if="tracks.length">
        <div v-if="groupedTracks.discs.length > 1" class="disc-groups">
          <div v-for="d in groupedTracks.discs" :key="d.disc" class="disc-group">
            <h3>Disque {{ d.disc }}</h3>
            <div class="side-columns">
              <div v-for="s in d.sides" :key="s.letter" class="side-column">
                <h4 v-if="s.tracks.length">Face {{ s.letter }}</h4>
                <ol v-if="s.tracks.length" class="tracklist-list">
                  <li v-for="(track, index) in s.tracks" :key="index" class="tracklist-item">
                    <span class="tracklist-position">{{ track.position || index + 1 }}</span>
                    <span class="tracklist-title">{{ track.title }}</span>
                    <span v-if="track.duration" class="tracklist-duration">{{ track.duration }}</span>
                    <StreamingButtons :disc="disc" :track="track" inline />
                  </li>
                </ol>
              </div>
            </div>
          </div>
        </div>
        <ol v-else class="tracklist-list">
          <li v-for="(track, index) in tracks" :key="index" class="tracklist-item">
            <span class="tracklist-position">{{ track.position || index + 1 }}</span>
            <span class="tracklist-title">{{ track.title }}</span>
            <span v-if="track.duration" class="tracklist-duration">{{ track.duration }}</span>
            <StreamingButtons :disc="disc" :track="track" inline />
          </li>
        </ol>
      </template>

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
  grid-template-columns: 2.5rem 1fr auto auto;
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
  grid-template-columns: 1fr 4.5rem auto;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem;
  border-radius: 10px;
  background: rgba(var(--tint-rgb), 0.03);
}

.tracklist-edit-item-other {
  grid-template-columns: 3.5rem 1fr 4.5rem auto;
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

.tracklist-add-btn.small {
  margin-top: 0.4rem;
  padding: 0.35rem;
  font-size: 0.82rem;
}

.disc-groups {
  display: grid;
  gap: 0.9rem;
  margin: 1rem 0;
  max-height: 55vh;
  overflow-y: auto;
}

.disc-group {
  border: 1px solid var(--line-soft);
  border-radius: 12px;
  padding: 0.75rem;
  background: rgba(var(--tint-rgb), 0.02);
}

.disc-group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.disc-group h3 {
  margin: 0 0 0.5rem 0;
  font-size: 1rem;
  color: var(--text);
}

.disc-group-header h3 {
  margin: 0;
}

.side-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.side-column h4 {
  margin: 0 0 0.4rem 0;
  font-size: 0.85rem;
  color: var(--text-dim);
}

.other-tracks-title {
  margin: 1rem 0 0.4rem 0;
  font-size: 0.95rem;
  color: var(--text);
}

@media (max-width: 640px) {
  .side-columns {
    grid-template-columns: 1fr;
  }
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
