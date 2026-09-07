<!-- frontend/src/components/DiscogsTracklistRefresh.vue
     Bouton "Mettre à jour depuis Discogs" + modale de validation
     (tracklist proposée + suggestion pour le champ Notes Discogs) —
     extrait de VinylsByArtist.vue pour être réutilisable ailleurs
     (DiscsModal.vue notamment), rien n'est jamais écrasé sans validation
     explicite. -->
<template>
  <button type="button" class="discogs-refresh-btn" :disabled="isRefetching" @click="refetchFromDiscogs">
    {{ isRefetching ? '⏳ Récupération…' : '🔄 Mettre à jour depuis Discogs' }}
  </button>

  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="isPreviewModalOpen" class="modal-overlay preview-tracklist-overlay" @click.self="cancelPreview">
        <div class="modal-card preview-tracklist-card">
          <div class="modal-header">
            <h2>🔄 Tracklist trouvée sur Discogs</h2>
          </div>
          <p class="preview-intro">
            Pour « {{ disc?.title }} » — {{ previewTracklistTracks.length }} piste(s),
            {{ previewTotalDuration }}. Vérifiez avant de remplacer la tracklist actuelle.
          </p>

          <div class="preview-tracklist">
            <div v-for="d in previewGroups.discs" :key="d.disc" class="preview-disc-group">
              <h4 v-if="previewGroups.discs.length > 1" class="preview-disc-title">Disque {{ d.disc }}</h4>
              <div class="preview-sides">
                <div v-for="s in d.sides" :key="s.letter" class="preview-side">
                  <h5 v-if="s.tracks.length">Face {{ s.letter }}</h5>
                  <ol v-if="s.tracks.length">
                    <li v-for="(t, i) in s.tracks" :key="i">
                      <span class="preview-track-pos">{{ t.position || i + 1 }}</span>
                      <span class="preview-track-title">{{ t.title }}</span>
                      <span class="preview-track-duration" v-if="t.duration">{{ t.duration }}</span>
                    </li>
                  </ol>
                </div>
              </div>
            </div>
            <ol v-if="previewGroups.noFace.length" class="preview-noface">
              <li v-for="(t, i) in previewGroups.noFace" :key="i">
                <span class="preview-track-pos">{{ t.position || i + 1 }}</span>
                <span class="preview-track-title">{{ t.title }}</span>
                <span class="preview-track-duration" v-if="t.duration">{{ t.duration }}</span>
              </li>
            </ol>
          </div>

          <label v-if="previewNotesSuggestion" class="preview-notes-suggestion">
            <input type="checkbox" v-model="applyNotesSuggestion" />
            <span class="preview-notes-text">
              <strong>Notes Discogs trouvées :</strong> « {{ previewNotesSuggestion }} »
            </span>
          </label>

          <div class="modal-actions">
            <button type="button" class="ghost-btn" :disabled="isApplyingPreview" @click="cancelPreview">
              Annuler
            </button>
            <button type="button" class="danger-btn" :disabled="isApplyingPreview" @click="applyPreview">
              {{ isApplyingPreview ? 'Enregistrement…' : 'Valider et remplacer' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useApi } from '@/composables/useApi'
import { previewTracklist, updateTracks } from '@/services/tracks'
import { groupTracksByDiscSide, formatDuration, sumDuration } from '@/utils/discSides'

const props = defineProps({
  // Le disque courant — a minima id, title, artist_id, genre_id,
  // format_id, country_id, label_id, release_year, barcode, price,
  // quantity, notes, isrc, discogs_notes.
  disc: { type: Object, required: true }
})

// tracks-updated: nouvelle tracklist appliquée ([]Track).
// disc-updated: disque renvoyé par l'API après application de la
// suggestion Notes Discogs (émis seulement si la case était cochée).
const emit = defineEmits(['tracks-updated', 'disc-updated'])

const { apiFetch } = useApi()

const isRefetching = ref(false)
const isPreviewModalOpen = ref(false)
const isApplyingPreview = ref(false)
const previewTracklistTracks = ref([])
const previewGroups = computed(() => groupTracksByDiscSide(previewTracklistTracks.value))
const previewTotalDuration = computed(() => formatDuration(sumDuration(previewTracklistTracks.value)))
// Coché par défaut seulement si le disque n'a pas déjà des Notes Discogs,
// pour ne jamais remplacer une valeur existante sans action explicite.
const previewNotesSuggestion = ref('')
const applyNotesSuggestion = ref(false)

const refetchFromDiscogs = async () => {
  if (!props.disc?.id) return
  isRefetching.value = true
  try {
    const result = await previewTracklist(props.disc.id)
    previewTracklistTracks.value = result?.tracks || []
    previewNotesSuggestion.value = (result?.notes || '').trim()
    applyNotesSuggestion.value = !!previewNotesSuggestion.value && !props.disc.discogs_notes?.trim()
    if (!previewTracklistTracks.value.length) {
      alert('Aucune piste trouvée sur Discogs pour ce disque.')
      return
    }
    isPreviewModalOpen.value = true
  } catch (err) {
    alert(`Échec de la recherche sur Discogs : ${err.message}`)
  } finally {
    isRefetching.value = false
  }
}

const cancelPreview = () => {
  isPreviewModalOpen.value = false
  previewTracklistTracks.value = []
  previewNotesSuggestion.value = ''
}

const applyPreview = async () => {
  if (!props.disc?.id) return
  isApplyingPreview.value = true
  try {
    const result = await updateTracks(props.disc.id, previewTracklistTracks.value)
    emit('tracks-updated', result || [])

    if (applyNotesSuggestion.value && previewNotesSuggestion.value) {
      const disc = props.disc
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
          discogs_notes: previewNotesSuggestion.value
        }),
        headers: { 'Content-Type': 'application/json' }
      })
      emit('disc-updated', updated)
    }

    isPreviewModalOpen.value = false
    previewTracklistTracks.value = []
    previewNotesSuggestion.value = ''
  } catch (err) {
    alert(`Échec de l'enregistrement des pistes : ${err.message}`)
  } finally {
    isApplyingPreview.value = false
  }
}
</script>

<style scoped>
.discogs-refresh-btn {
  padding: 10px;
  border-radius: 8px;
  border: 1px solid var(--line);
  background: rgba(var(--tint-rgb), 0.05);
  color: var(--text);
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
  width: 100%;
}

.discogs-refresh-btn:hover:not(:disabled) {
  background: rgba(var(--tint-rgb), 0.1);
}

.discogs-refresh-btn:disabled {
  opacity: 0.6;
  cursor: default;
}

/* Ce composant peut s'ouvrir depuis DiscsModal.vue (bouton dans son champ
   "Notes Discogs"), dont l'overlay est à z-index: 9999 — sans ce z-index
   dédié plus élevé, cette modale hérite du .modal-overlay partagé
   (z-index: 50, App.vue) et se retrouve visuellement DERRIÈRE celui de
   DiscsModal : visible en transparence, mais tous les clics (dont
   "Valider et remplacer") sont interceptés par l'overlay du dessus, sans
   jamais déclencher la moindre requête. Même piège déjà documenté et
   corrigé dans DiscsModal.vue pour ses propres modales imbriquées
   (.barcode-conflict-overlay, .unsaved-changes-overlay, toutes deux à
   10000). */
.preview-tracklist-overlay {
  z-index: 10000;
}

.preview-tracklist-card {
  max-width: 560px;
}

.modal-header h2 {
  margin: 0 0 14px 0;
  font-size: 1.3em;
  color: var(--text);
}

.preview-intro {
  margin: 0 0 14px 0;
  color: var(--text-soft);
  line-height: 1.5;
}

.preview-tracklist {
  max-height: 45vh;
  overflow-y: auto;
  display: grid;
  gap: 14px;
}

.preview-disc-title {
  margin: 0 0 8px 0;
  color: var(--text);
  font-size: 0.95em;
  font-weight: 600;
}

.preview-sides {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.preview-side h5 {
  margin: 0 0 6px 0;
  color: var(--text-dim);
  font-size: 0.85em;
}

.preview-tracklist ol {
  list-style: none;
  margin: 0;
  padding: 0;
}

.preview-tracklist li {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  font-size: 0.88em;
  color: var(--text);
  border-bottom: 1px solid var(--line-soft);
}

.preview-tracklist li:last-child {
  border-bottom: none;
}

.preview-track-pos {
  color: var(--text-dim);
  min-width: 24px;
  font-weight: 600;
}

.preview-track-title {
  flex: 1;
  min-width: 0;
}

.preview-track-duration {
  color: var(--text-dim);
  font-size: 0.9em;
}

.preview-notes-suggestion {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-top: 14px;
  padding: 10px 12px;
  background: rgba(var(--tint-rgb), 0.05);
  border: 1px solid var(--line-soft);
  border-radius: 10px;
  cursor: pointer;
}

.preview-notes-suggestion input[type='checkbox'] {
  margin-top: 3px;
  flex-shrink: 0;
}

.preview-notes-text {
  color: var(--text-soft);
  font-size: 0.9em;
  line-height: 1.4;
}

.modal-actions {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>
