<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useApi } from '@/composables/useApi'
import rawgApi from '@/services/rawgApi'
import BarcodeScanner from '@/components/BarcodeScanner.vue'

const props = defineProps({
  isOpen: { type: Boolean, default: false },
  gameData: { type: Object, default: null },
  apiError: { type: String, default: null },
})

const emit = defineEmits(['close', 'game-saved', 'lookups-changed'])

const { apiFetch, upload } = useApi()

const isEditing = computed(() => !!props.gameData?.id)
const showSearchStep = ref(true)
const isSaving = ref(false)
const localApiError = ref(null)

const platforms = ref([])
const gameGenres = ref([])
const publishers = ref([])

async function fetchLookups() {
  try {
    const [p, g, pub] = await Promise.all([
      apiFetch('/platforms'),
      apiFetch('/game-genres'),
      apiFetch('/publishers'),
    ])
    platforms.value = Array.isArray(p) ? p : []
    gameGenres.value = Array.isArray(g) ? g : []
    publishers.value = Array.isArray(pub) ? pub : []
  } catch (error) {
    console.error('Erreur chargement listes de référence:', error)
  }
}

function defaultFormData() {
  return {
    id: null,
    title: '',
    platform_id: '',
    genre_id: '',
    publisher_id: '',
    release_year: '',
    barcode: '',
    price: '',
    quantity: 1,
    notes: '',
    cover_image: '',
    rawg_id: null,
  }
}

const formData = ref(defaultFormData())
const pendingGenreName = ref('')
const pendingPublisherName = ref('')

// === Étape recherche ===
const barcodeInput = ref(null)
const barcode = ref('')
const searchTitle = ref('')
const isSearching = ref(false)
const searchResults = ref([])
const searchMessage = ref('')
const selectedResult = ref(null)
const isBarcodeScannerOpen = ref(false)
const fileImportInput = ref(null)

// === Gestion de la jaquette (modifier / supprimer / upload) ===
const fileInput = ref(null)
const uploadPreview = ref({ url: null, file: null, size: 0, width: 0, height: 0, loading: false })

function resetSearchStep() {
  barcode.value = ''
  searchTitle.value = ''
  searchResults.value = []
  searchMessage.value = ''
  selectedResult.value = null
}

watch(
  () => props.isOpen,
  async (open) => {
    if (!open) return
    localApiError.value = null
    await fetchLookups()
    if (isEditing.value) {
      showSearchStep.value = false
      formData.value = {
        id: props.gameData.id,
        title: props.gameData.title || '',
        platform_id: props.gameData.platform_id || '',
        genre_id: props.gameData.genre_id || '',
        publisher_id: props.gameData.publisher_id || '',
        release_year: props.gameData.release_year || '',
        barcode: props.gameData.barcode || '',
        price: props.gameData.price ?? '',
        quantity: props.gameData.quantity ?? 1,
        notes: props.gameData.notes || '',
        cover_image: props.gameData.cover_url || '',
        rawg_id: props.gameData.rawg_id || null,
      }
    } else {
      showSearchStep.value = true
      formData.value = defaultFormData()
      resetSearchStep()
      // Le champ code-barres doit avoir le focus dès l'ouverture : une
      // douchette USB se comporte comme un clavier qui tape très vite +
      // Entrée — sans focus sur ce champ, la saisie part dans le vide (ou
      // dans un autre champ) et "le numéro ne se met pas dans la case".
      await nextTick()
      barcodeInput.value?.focus()
    }
    pendingGenreName.value = ''
    pendingPublisherName.value = ''
  }
)

function handleBarcodeScanned(code) {
  isBarcodeScannerOpen.value = false
  barcode.value = code
}

function triggerFileImport() {
  fileImportInput.value?.click()
}

async function handleFileImport(event) {
  const file = event.target.files?.[0]
  event.target.value = ''
  if (!file) return
  try {
    const text = await file.text()
    const line = text
      .split('\n')
      .map((l) => l.trim())
      .find((l) => l.length >= 6)
    if (line) barcode.value = line
  } catch (error) {
    console.error('Erreur lecture fichier:', error)
  }
}

async function pasteFromClipboard() {
  try {
    const text = await navigator.clipboard.readText()
    if (text) barcode.value = text.trim()
  } catch (error) {
    console.error('Presse-papiers indisponible:', error)
  }
}

async function doSearch() {
  if (!searchTitle.value.trim()) return
  isSearching.value = true
  searchMessage.value = ''
  searchResults.value = []
  try {
    const preview = await rawgApi.searchByTitle(searchTitle.value)
    if (preview.found && preview.results?.length) {
      searchResults.value = preview.results
      searchMessage.value = `${preview.results.length} résultat(s) trouvé(s)`
    } else {
      searchMessage.value = 'Aucun résultat trouvé'
    }
  } catch (error) {
    searchMessage.value = error.message
  } finally {
    isSearching.value = false
  }
}

function applyPreviewToForm(preview) {
  formData.value.title = preview.title || searchTitle.value
  formData.value.release_year = preview.year ? String(preview.year).split('-')[0] : ''
  formData.value.cover_image = preview.cover_url || ''
  formData.value.rawg_id = preview.rawg_id || null
  formData.value.barcode = barcode.value || ''

  pendingGenreName.value = ''
  if (preview.genres?.length) {
    const match = gameGenres.value.find((g) => g.name.toLowerCase() === preview.genres[0].toLowerCase())
    formData.value.genre_id = match ? match.id : ''
    if (!match) pendingGenreName.value = preview.genres[0]
  }

  pendingPublisherName.value = ''
  if (preview.publisher) {
    const match = publishers.value.find((p) => p.name.toLowerCase() === preview.publisher.toLowerCase())
    formData.value.publisher_id = match ? match.id : ''
    if (!match) pendingPublisherName.value = preview.publisher
  }
}

async function selectResult(result) {
  isSearching.value = true
  try {
    const details = await rawgApi.getGameDetails(result.id)
    applyPreviewToForm(details)
  } catch (error) {
    // Repli : on utilise les infos basiques du résultat de recherche
    applyPreviewToForm({
      title: result.name,
      year: result.released,
      cover_url: result.background_image,
      rawg_id: result.id,
      genres: (result.genres || []).map((g) => g.name),
    })
  } finally {
    isSearching.value = false
    showSearchStep.value = false
  }
}

function skipToManualEntry() {
  if (isEditing.value) {
    // Recherche relancée depuis l'édition (researchOnWeb) : on revient au
    // formulaire tel quel, sans perdre les données déjà saisies —
    // contrairement à la création, où le formulaire démarre vide.
    showSearchStep.value = false
    return
  }
  formData.value = defaultFormData()
  formData.value.barcode = barcode.value || ''
  showSearchStep.value = false
}

function backToSearch() {
  showSearchStep.value = true
  nextTick(() => barcodeInput.value?.focus())
}

// Permet, en édition, de relancer une recherche RAWG pour remplacer la
// jaquette (ou d'autres infos) sans perdre les données déjà saisies dans le
// formulaire — même principe que researchOnDiscogs côté disques.
function researchOnWeb() {
  searchResults.value = []
  searchMessage.value = ''
  barcode.value = formData.value.barcode || ''
  searchTitle.value = formData.value.title || ''
  showSearchStep.value = true
  nextTick(() => barcodeInput.value?.focus())
}

function openFilePicker() {
  fileInput.value?.click()
}

function handleFileUpload(event) {
  const file = event.target.files?.[0]
  event.target.value = ''
  if (!file) return
  const validTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp']
  if (!validTypes.includes(file.type) || file.size > 5 * 1024 * 1024) return
  const reader = new FileReader()
  reader.onload = (e) => {
    const img = new Image()
    img.onload = () => {
      Object.assign(uploadPreview.value, {
        url: e.target.result,
        file,
        size: file.size,
        width: img.width,
        height: img.height,
      })
    }
    img.src = e.target.result
  }
  reader.readAsDataURL(file)
}

function cancelUpload() {
  uploadPreview.value = { url: null, file: null, size: 0, width: 0, height: 0, loading: false }
}

async function confirmUpload() {
  if (!uploadPreview.value.file) return
  uploadPreview.value.loading = true
  try {
    const fd = new FormData()
    fd.append('cover', uploadPreview.value.file)
    fd.append('title', formData.value.title || 'Jeu')
    const platformName = platforms.value.find((p) => p.id === Number(formData.value.platform_id))?.name
    fd.append('artist', platformName || 'Jeu')
    const res = await upload('/upload-cover', fd)
    if (res.filePath) {
      formData.value.cover_image = res.filePath
      cancelUpload()
    }
  } catch (error) {
    console.error('Erreur upload jaquette:', error)
  } finally {
    uploadPreview.value.loading = false
  }
}

function removeCover() {
  formData.value.cover_image = ''
}

function formatFileSize(bytes) {
  if (!bytes) return '0 Ko'
  const k = 1024
  const sizes = ['o', 'Ko', 'Mo']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(k)), 2)
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

// === Création rapide plateforme / genre / éditeur ===
const quickCreate = ref({ type: null, name: '' })

function openQuickCreate(type, prefillName = '') {
  quickCreate.value = { type, name: prefillName }
}

function closeQuickCreate() {
  quickCreate.value = { type: null, name: '' }
}

const quickCreateLabel = computed(() => {
  const labels = { platform: 'une plateforme', genre: 'un genre', publisher: 'un éditeur' }
  return labels[quickCreate.value.type] || ''
})

async function saveQuickCreate() {
  const { type, name } = quickCreate.value
  if (!name.trim()) return
  const endpointMap = { platform: '/platforms', genre: '/game-genres', publisher: '/publishers' }
  try {
    const created = await apiFetch(endpointMap[type], {
      method: 'POST',
      body: JSON.stringify({ name: name.trim(), description: '' }),
    })
    if (type === 'platform') {
      platforms.value.push(created)
      formData.value.platform_id = created.id
    } else if (type === 'genre') {
      gameGenres.value.push(created)
      formData.value.genre_id = created.id
      pendingGenreName.value = ''
    } else if (type === 'publisher') {
      publishers.value.push(created)
      formData.value.publisher_id = created.id
      pendingPublisherName.value = ''
    }
    emit('lookups-changed')
    closeQuickCreate()
  } catch (error) {
    console.error('Erreur création rapide:', error)
  }
}

// === Sauvegarde ===
async function handleSubmit() {
  if (!formData.value.title.trim() || !formData.value.platform_id) return

  const payload = {
    title: formData.value.title.trim(),
    platform_id: Number(formData.value.platform_id),
    genre_id: formData.value.genre_id ? Number(formData.value.genre_id) : null,
    publisher_id: formData.value.publisher_id ? Number(formData.value.publisher_id) : null,
    release_year: formData.value.release_year ? parseInt(formData.value.release_year) : null,
    barcode: formData.value.barcode?.trim() || null,
    notes: formData.value.notes?.trim() || null,
    price: formData.value.price !== '' && formData.value.price != null ? parseFloat(formData.value.price) : null,
    quantity: formData.value.quantity ? parseInt(formData.value.quantity) : 1,
  }

  if (!isEditing.value) {
    payload.cover_image = formData.value.cover_image || null
    payload.rawg_id = formData.value.rawg_id || null
  } else {
    // Toujours envoyer le champ en édition (jamais l'omettre) : une chaîne
    // vide signale explicitement une suppression au backend, qu'il y ait eu
    // ou non une jaquette avant — omettre le champ serait interprété comme
    // "aucun changement demandé" et laisserait l'ancienne jaquette en place.
    payload.cover_image = formData.value.cover_image || ''
  }

  isSaving.value = true
  localApiError.value = null
  try {
    const url = isEditing.value ? `/games/${formData.value.id}` : '/games'
    const result = await apiFetch(url, {
      method: isEditing.value ? 'PUT' : 'POST',
      body: JSON.stringify(payload),
    })
    emit('game-saved', result)
  } catch (error) {
    localApiError.value = error.message
  } finally {
    isSaving.value = false
  }
}

function closeModal() {
  emit('close')
}

// Évite la fermeture accidentelle de la modale quand l'utilisateur
// sélectionne du texte (ex: pour le copier) et que le relâchement de la
// souris finit hors de la carte : sans cette garde, @click.self suffisait
// à fermer même si le mousedown avait démarré à l'intérieur.
const overlayMouseDownTarget = ref(null)
const onOverlayMouseDown = (e) => {
  overlayMouseDownTarget.value = e.target
}
const onOverlayClick = (e) => {
  if (e.target === e.currentTarget && overlayMouseDownTarget.value === e.currentTarget) {
    closeModal()
  }
  overlayMouseDownTarget.value = null
}

const quickCreateOverlayMouseDownTarget = ref(null)
const onQuickCreateOverlayMouseDown = (e) => {
  quickCreateOverlayMouseDownTarget.value = e.target
}
const onQuickCreateOverlayClick = (e) => {
  if (e.target === e.currentTarget && quickCreateOverlayMouseDownTarget.value === e.currentTarget) {
    closeQuickCreate()
  }
  quickCreateOverlayMouseDownTarget.value = null
}

const displayedError = computed(() => localApiError.value || props.apiError)
</script>

<template>
  <div v-if="isOpen" class="modal-overlay" @mousedown="onOverlayMouseDown" @click="onOverlayClick">
    <section class="modal-card game-modal-card" role="dialog" aria-modal="true" aria-labelledby="game-modal-title">
      <div class="modal-header">
        <div>
          <p class="eyebrow">{{ isEditing ? 'Modifier le jeu' : 'Ajouter un jeu' }}</p>
          <h2 id="game-modal-title">{{ isEditing ? formData.title : 'Nouveau jeu vidéo' }}</h2>
        </div>
        <button type="button" class="icon-action-btn" @click="closeModal" aria-label="Fermer">
          <span>&times;</span>
        </button>
      </div>

      <p v-if="displayedError" class="form-error">{{ displayedError }}</p>

      <!-- === Étape recherche === -->
      <div v-if="showSearchStep" class="game-search-step">
        <div class="barcode-input-container">
          <div class="barcode-input-group">
            <input
              ref="barcodeInput"
              v-model="barcode"
              type="text"
              placeholder="Code-barres (scanner, douchette ou saisie)"
              class="barcode-input"
            />
            <button type="button" @click="isBarcodeScannerOpen = true" class="scan-button" title="Scanner avec la caméra">📷</button>
            <button type="button" @click="triggerFileImport" class="scan-button" title="Importer depuis un fichier">📄</button>
            <button type="button" @click="pasteFromClipboard" class="scan-button" title="Coller depuis le presse-papiers">📋</button>
            <input ref="fileImportInput" type="file" accept=".txt" style="display: none" @change="handleFileImport" />
          </div>
          <p class="scan-hint">
            ℹ️ Le code-barres est capturé et conservé, mais la recherche se fait par <strong>titre</strong> ci-dessous — RAWG (base de jeux)
            ne permet pas la recherche par code-barres.
          </p>
        </div>

        <div class="game-search-divider">OU</div>

        <div class="game-manual-search">
          <label class="form-field">
            <span>Titre du jeu</span>
            <div class="search-row">
              <input v-model="searchTitle" type="text" placeholder="Ex: Zelda, God of War..." @keyup.enter="doSearch" />
              <button type="button" class="primary-btn" :disabled="isSearching || !searchTitle.trim()" @click="doSearch">
                <span v-if="isSearching" class="btn-spinner"></span>
                <span v-else>🔍 Rechercher</span>
              </button>
            </div>
          </label>
          <p v-if="searchMessage" class="reports-muted">{{ searchMessage }}</p>
        </div>

        <div v-if="searchResults.length" class="game-results-grid">
          <div v-for="result in searchResults" :key="result.id" class="game-result-card">
            <img v-if="result.background_image" :src="result.background_image" :alt="result.name" class="game-result-cover" />
            <div v-else class="game-result-cover game-result-cover-fallback">🎮</div>
            <div class="game-result-info">
              <strong>{{ result.name }}</strong>
              <span class="game-result-meta">{{ result.released ? result.released.split('-')[0] : '—' }}</span>
            </div>
            <button type="button" class="ghost-btn" :disabled="isSearching" @click="selectResult(result)">Choisir</button>
          </div>
        </div>

        <div class="modal-actions game-search-actions">
          <button type="button" class="ghost-btn" @click="skipToManualEntry">✏️ Saisir manuellement</button>
        </div>
      </div>

      <!-- === Étape formulaire === -->
      <form v-else class="game-form" @submit.prevent="handleSubmit">
        <div class="game-cover-container">
          <img v-if="formData.cover_image" :src="formData.cover_image" alt="Jaquette" class="game-cover-preview-img" />
          <div v-else class="game-cover-preview-img game-cover-placeholder">🎮</div>
          <div class="game-cover-actions">
            <button type="button" @click="openFilePicker" class="game-cover-action-btn" title="Changer la jaquette">📁</button>
            <button type="button" @click="researchOnWeb" class="game-cover-action-btn" title="Rechercher une jaquette sur le web">🔍</button>
            <button
              type="button"
              v-if="formData.cover_image"
              @click="removeCover"
              class="game-cover-action-btn remove"
              title="Supprimer la jaquette"
            >
              🗑️
            </button>
          </div>
        </div>
        <input ref="fileInput" type="file" accept="image/*" style="display: none" @change="handleFileUpload" />

        <div v-if="uploadPreview.url" class="game-upload-preview">
          <div class="game-upload-preview-header">
            <h5>📷 Aperçu de l'image</h5>
            <button type="button" class="ghost-btn" @click="cancelUpload">❌ Annuler</button>
          </div>
          <div class="game-upload-preview-content">
            <img :src="uploadPreview.url" alt="Aperçu upload" class="game-upload-preview-img" />
            <div class="game-upload-preview-info">
              <span>{{ formatFileSize(uploadPreview.size) }} · {{ uploadPreview.width }}×{{ uploadPreview.height }}px</span>
              <button type="button" class="primary-btn" @click="confirmUpload" :disabled="uploadPreview.loading">
                <span v-if="uploadPreview.loading" class="btn-spinner"></span>
                <span v-else>✅ Utiliser cette image</span>
              </button>
            </div>
          </div>
        </div>

        <label class="form-field">
          <span class="required">Titre</span>
          <input v-model="formData.title" type="text" required placeholder="Ex: The Legend of Zelda" :disabled="isSaving" />
        </label>

        <div class="game-form-grid">
          <label class="form-field">
            <span class="required">Plateforme</span>
            <div class="select-with-button">
              <select v-model="formData.platform_id" required :disabled="isSaving">
                <option value="">Sélectionnez une plateforme</option>
                <option v-for="p in platforms" :key="p.id" :value="p.id">{{ p.name }}</option>
              </select>
              <button type="button" class="create-quick-button" title="Créer une nouvelle plateforme" @click="openQuickCreate('platform')">
                <span class="icon">+</span>
              </button>
            </div>
          </label>

          <label class="form-field">
            <span>Genre</span>
            <div class="select-with-button">
              <select v-model="formData.genre_id" :disabled="isSaving">
                <option value="">Sélectionnez un genre</option>
                <option v-for="g in gameGenres" :key="g.id" :value="g.id">{{ g.name }}</option>
              </select>
              <button type="button" class="create-quick-button" title="Créer un nouveau genre" @click="openQuickCreate('genre')">
                <span class="icon">+</span>
              </button>
            </div>
            <div v-if="pendingGenreName" class="form-help">
              <button type="button" class="inline-button" @click="openQuickCreate('genre', pendingGenreName)">
                🎮 Créer "{{ pendingGenreName }}"
              </button>
            </div>
          </label>

          <label class="form-field">
            <span>Éditeur</span>
            <div class="select-with-button">
              <select v-model="formData.publisher_id" :disabled="isSaving">
                <option value="">Sélectionnez un éditeur</option>
                <option v-for="p in publishers" :key="p.id" :value="p.id">{{ p.name }}</option>
              </select>
              <button type="button" class="create-quick-button" title="Créer un nouvel éditeur" @click="openQuickCreate('publisher')">
                <span class="icon">+</span>
              </button>
            </div>
            <div v-if="pendingPublisherName" class="form-help">
              <button type="button" class="inline-button" @click="openQuickCreate('publisher', pendingPublisherName)">
                🏢 Créer "{{ pendingPublisherName }}"
              </button>
            </div>
          </label>

          <label class="form-field">
            <span>Année</span>
            <input v-model="formData.release_year" type="number" placeholder="Ex: 2023" :disabled="isSaving" />
          </label>

          <label class="form-field">
            <span>Code-barres</span>
            <input v-model="formData.barcode" type="text" placeholder="Ex: 0045496510960" :disabled="isSaving" />
          </label>

          <label class="form-field">
            <span>Quantité</span>
            <input v-model="formData.quantity" type="number" min="1" :disabled="isSaving" />
          </label>
        </div>

        <label class="form-field">
          <span>Notes</span>
          <textarea v-model="formData.notes" rows="3" placeholder="État, édition spéciale, anecdotes..." :disabled="isSaving"></textarea>
        </label>

        <div class="modal-actions">
          <button v-if="!isEditing" type="button" class="ghost-btn" @click="backToSearch">← Retour à la recherche</button>
          <button type="button" class="ghost-btn" @click="closeModal">Annuler</button>
          <button type="submit" class="primary-btn" :disabled="isSaving || !formData.title.trim() || !formData.platform_id">
            <span v-if="isSaving" class="btn-spinner"></span>
            <span v-else>Enregistrer</span>
          </button>
        </div>
      </form>
    </section>

    <BarcodeScanner :is-open="isBarcodeScannerOpen" @close="isBarcodeScannerOpen = false" @barcode-detected="handleBarcodeScanned" />

    <!-- === Sous-modale de création rapide (plateforme/genre/éditeur) === -->
    <Teleport to="body">
      <div
        v-if="quickCreate.type"
        class="modal-overlay"
        @mousedown="onQuickCreateOverlayMouseDown"
        @click="onQuickCreateOverlayClick"
      >
        <section class="modal-card quick-create-card" role="dialog" aria-modal="true">
          <div class="modal-header">
            <h2>Créer {{ quickCreateLabel }}</h2>
            <button type="button" class="icon-action-btn" @click="closeQuickCreate" aria-label="Fermer">
              <span>&times;</span>
            </button>
          </div>
          <label class="form-field">
            <span>Nom</span>
            <input v-model="quickCreate.name" type="text" autofocus @keyup.enter="saveQuickCreate" />
          </label>
          <div class="modal-actions">
            <button type="button" class="ghost-btn" @click="closeQuickCreate">Annuler</button>
            <button type="button" class="primary-btn" :disabled="!quickCreate.name.trim()" @click="saveQuickCreate">Créer</button>
          </div>
        </section>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.game-modal-card {
  max-width: 640px;
  width: 100%;
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

.scan-button:hover {
  background: rgba(var(--tint-rgb), 0.1);
}

.scan-hint {
  margin: 0;
  color: var(--text-dim);
  font-size: 0.8rem;
}

.game-search-divider {
  text-align: center;
  color: var(--text-dim);
  font-size: 0.8rem;
  letter-spacing: 0.08em;
  margin: 1.2rem 0;
}

.game-manual-search .search-row {
  display: flex;
  gap: 0.6rem;
}

.game-manual-search input {
  flex: 1;
  padding: 0.7rem 0.9rem;
  border-radius: 10px;
  border: 1px solid var(--line);
  background: rgba(var(--tint-rgb), 0.04);
  color: var(--text);
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
  to {
    transform: rotate(360deg);
  }
}

.game-results-grid {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  margin-top: 1rem;
  max-height: 320px;
  overflow-y: auto;
}

.game-result-card {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  padding: 0.6rem;
  border: 1px solid var(--line-soft);
  border-radius: 12px;
  background: rgba(var(--tint-rgb), 0.03);
}

.game-result-cover {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  object-fit: cover;
  flex-shrink: 0;
}

.game-result-cover-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(var(--tint-rgb), 0.08);
  font-size: 1.3em;
}

.game-result-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.game-result-meta {
  color: var(--text-dim);
  font-size: 0.85em;
}

.game-search-actions {
  justify-content: center;
  margin-top: 1.2rem;
}

.game-cover-container {
  position: relative;
  display: flex;
  justify-content: center;
  margin-bottom: 1rem;
}

.game-cover-preview-img {
  /* Format portrait (~0.8), comme les vignettes de GamesView.vue : les
     jaquettes de jeux (box art) sont le plus souvent verticales, pas
     carrées comme les pochettes de disques. */
  width: 140px;
  height: 175px;
  border-radius: 10px;
  object-fit: cover;
}

.game-cover-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(var(--tint-rgb), 0.06);
  border: 1px dashed var(--line);
  font-size: 2.2em;
  color: var(--text-dim);
}

.game-cover-actions {
  position: absolute;
  top: 4px;
  right: calc(50% - 70px + 4px);
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.game-cover-container:hover .game-cover-actions,
.game-cover-container:focus-within .game-cover-actions {
  opacity: 1;
}

.game-cover-action-btn {
  background: rgba(0, 0, 0, 0.7);
  border: none;
  border-radius: 6px;
  color: white;
  cursor: pointer;
  padding: 4px 8px;
  font-size: 0.95em;
  transition: all 0.2s;
}

.game-cover-action-btn:hover {
  background: rgba(0, 0, 0, 0.9);
  transform: scale(1.08);
}

.game-cover-action-btn.remove {
  background: rgba(220, 53, 69, 0.75);
}

.game-cover-action-btn.remove:hover {
  background: rgba(220, 53, 69, 0.95);
}

.game-upload-preview {
  border: 1px solid var(--line-soft);
  border-radius: 12px;
  padding: 0.8rem;
  margin-bottom: 1rem;
  background: rgba(var(--tint-rgb), 0.03);
}

.game-upload-preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.6rem;
}

.game-upload-preview-header h5 {
  margin: 0;
  font-size: 0.9em;
  color: var(--text);
}

.game-upload-preview-content {
  display: flex;
  align-items: center;
  gap: 0.9rem;
  flex-wrap: wrap;
}

.game-upload-preview-img {
  width: 80px;
  height: 80px;
  border-radius: 8px;
  object-fit: cover;
  flex-shrink: 0;
}

.game-upload-preview-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  font-size: 0.85em;
  color: var(--text-soft);
}

.game-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.9rem;
  margin-top: 0.9rem;
}

@media (max-width: 560px) {
  .game-form-grid {
    grid-template-columns: 1fr;
  }
}

.select-with-button {
  display: flex;
  gap: 0.5rem;
}

.select-with-button select {
  flex: 1;
}

.create-quick-button {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  border: none;
  border-radius: 10px;
  background: var(--accent);
  color: white;
  font-size: 1.1em;
  cursor: pointer;
}

.form-help {
  margin-top: 0.3rem;
}

.inline-button {
  background: none;
  border: none;
  color: var(--accent);
  cursor: pointer;
  font-size: 0.85em;
  padding: 0;
}

.quick-create-card {
  max-width: 400px;
  width: 100%;
}
</style>
