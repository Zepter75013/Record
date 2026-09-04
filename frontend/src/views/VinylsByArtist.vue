<!-- VinylsByArtist.vue -->
<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '@/composables/useApi'
import StreamingButtons from '@/components/StreamingButtons.vue'
import TracklistModal from '@/components/TracklistModal.vue'
import { fetchTracks } from '@/services/tracks'
import { formatCurrency } from '@/utils/format'

const router = useRouter()
const { apiFetch } = useApi()

// États
const artists = ref([])
const vinyls = ref([])
const isLoading = ref(true)
const error = ref(null)
const searchQuery = ref('')
const selectedLetter = ref('All')
const sortOrder = ref('asc')

// États pour redimensionnement et toggle
const sidebarWidth = ref(400)
const isResizing = ref(false)
const isSidebarCollapsed = ref(false)

// Alphabet pour le filtrage
const alphabet = ['All', '0-9', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z']

// Artistes filtrés
const filteredArtists = computed(() => {
  let filtered = [...artists.value]
  
  // Filtrage par lettre
  if (selectedLetter.value !== 'All') {
    if (selectedLetter.value === '0-9') {
      filtered = filtered.filter(artist => /^[0-9]/.test(artist.name))
    } else {
      filtered = filtered.filter(artist => 
        artist.name.toUpperCase().startsWith(selectedLetter.value)
      )
    }
  }
  
  // Filtrage par recherche
  if (searchQuery.value) {
    filtered = filtered.filter(artist =>
      artist.name.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  }
  
  // Tri
  if (sortOrder.value === 'asc') {
    filtered.sort((a, b) => a.name.localeCompare(b.name))
  } else {
    filtered.sort((a, b) => b.name.localeCompare(a.name))
  }
  
  return filtered
})

// Disques filtrés selon les artistes visibles
const filteredVinyls = computed(() => {
  if (selectedLetter.value === 'All' && !searchQuery.value) {
    return vinyls.value
  }
  
  const artistIds = filteredArtists.value.map(artist => artist.id)
  return vinyls.value.filter(vinyl => artistIds.includes(vinyl.artist_id))
})

// Vérifier si une lettre a des artistes
const hasArtistsForLetter = (letter) => {
  if (letter === 'All') return true
  
  if (letter === '0-9') {
    return artists.value.some(artist => /^[0-9]/.test(artist.name))
  }
  
  return artists.value.some(artist => 
    artist.name.toUpperCase().startsWith(letter)
  )
}

const gridStyle = computed(() => ({
  gridTemplateColumns: sidebarWidth.value + 'px 1fr'
}))

// Disques d'un artiste sélectionné
const selectedArtist = ref(null)
const artistVinyls = computed(() => {
  if (!selectedArtist.value) return []
  return vinyls.value.filter(vinyl => vinyl.artist_id === selectedArtist.value.id)
})
const sortedArtistVinyls = computed(() => {
  return [...artistVinyls.value].sort((a, b) => {
    const ay = a.release_year || 0
    const by = b.release_year || 0
    if (ay !== by) return ay - by
    return (a.title || '').localeCompare(b.title || '')
  })
})

// Album sélectionné dans la fiche détaillée + ses pistes
const selectedVinyl = ref(null)
const vinylTracks = ref([])
const tracksLoading = ref(false)

const loadVinylTracks = async (vinyl) => {
  if (!vinyl?.id) {
    vinylTracks.value = []
    return
  }
  tracksLoading.value = true
  try {
    vinylTracks.value = await fetchTracks(vinyl.id)
  } catch (err) {
    console.error('❌ Erreur chargement pistes:', err)
    vinylTracks.value = []
  } finally {
    tracksLoading.value = false
  }
}

const selectVinyl = (vinyl) => {
  selectedVinyl.value = vinyl
  loadVinylTracks(vinyl)
}

// Répartit les pistes par face à partir du préfixe de position (A1, B2…) ;
// si aucune position n'est renseignée, on coupe simplement la liste en deux.
const tracksByFace = computed(() => {
  const tracks = vinylTracks.value
  const hasPositions = tracks.some(t => /^[AB]/i.test(t.position || ''))
  if (hasPositions) {
    return {
      A: tracks.filter(t => /^A/i.test(t.position || '')),
      B: tracks.filter(t => /^B/i.test(t.position || ''))
    }
  }
  const mid = Math.ceil(tracks.length / 2)
  return { A: tracks.slice(0, mid), B: tracks.slice(mid) }
})
const faceATracks = computed(() => tracksByFace.value.A)
const faceBTracks = computed(() => tracksByFace.value.B)

const formatPrice = (value) => {
  if (value === null || value === undefined || value === '') return '—'
  return formatCurrency(value)
}

// Charger les artistes
const fetchArtists = async () => {
  try {
    isLoading.value = true
    error.value = null
    
    console.log('🔄 Chargement des artistes...')
    
    const [artistsData, vinylsData] = await Promise.all([
      apiFetch('artists'),
      apiFetch('discs')
    ])
    
    vinyls.value = vinylsData
    
    artists.value = artistsData.map(artist => ({
      ...artist,
      vinyl_count: vinyls.value.filter(v => v.artist_id === artist.id).length
    })).filter(artist => artist.vinyl_count > 0)
    
    console.log('✅ Artistes chargés:', artists.value.length)
  } catch (err) {
    error.value = 'Erreur lors du chargement des artistes'
    console.error('❌ Erreur fetchArtists:', err)
  } finally {
    isLoading.value = false
  }
}

// Sélectionner un artiste
const selectArtist = (artist) => {
  selectedArtist.value = artist
  selectVinyl(sortedArtistVinyls.value[0] || null)
  console.log('🎤 Artiste sélectionné:', artist.name, '(', artist.vinyl_count, 'disques )')
}

// Retour à la liste
const backToList = () => {
  selectedArtist.value = null
  selectedVinyl.value = null
  vinylTracks.value = []
}

// Voir les détails d'un disque
const viewVinylDetails = (vinyl) => {
  router.push(`/dashboard/vinyls`)
}

const isTracklistModalOpen = ref(false)
const vinylForTracklist = ref(null)
const openTracklistModal = (vinyl) => {
  vinylForTracklist.value = vinyl
  isTracklistModalOpen.value = true
}
const handleTracksUpdated = ({ discId, hasTracks }) => {
  const vinyl = vinyls.value.find((v) => v.id === discId)
  if (vinyl) vinyl.has_tracks = hasTracks
}

// Normaliser l'URL de la pochette — même logique que getImageUrl de
// DiscsView.vue : en production VITE_SERVER_BASE_URL est vide et nginx
// proxifie /uploads/ sur le même domaine (voir nginx.conf), donc une URL
// relative suffit. Un port codé en dur ici pointerait dans le vide.
const normalizeCoverUrl = (url) => {
  if (!url) return ''

  // Si l'URL est déjà complète avec protocole
  if (url.startsWith('https://') || url.startsWith('http://')) {
    return url.replace('http://', 'https://') // Forcer HTTPS en production
  }

  const SERVER_BASE_URL = import.meta.env.VITE_SERVER_BASE_URL || ''

  if (url.startsWith('/')) {
    return `${SERVER_BASE_URL}${url}`
  }

  return `${SERVER_BASE_URL}/uploads/${url}`
}

// Gestion erreur image
const handleImageError = (e) => {
  e.target.style.display = 'none'
  const fallback = e.target.parentElement?.querySelector('.patchwork-fallback, .cover-fallback, .mosaic-fallback')
  if (fallback) fallback.style.display = 'flex'
}

// Tri
const toggleSort = () => {
  sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
}

const getSortIcon = () => {
  return sortOrder.value === 'asc' ? '👆' : '👇'
}

const toggleSidebarMobile = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

// Calculer la position pour un patchwork dense et naturel
const getCircularPosition = (index, total) => {
  // Plus de cartes, plus d'anneaux
  const itemsPerRing = [1, 6, 10, 14, 18]; // 5 anneaux concentriques
  let ring = 0;
  let positionInRing = index;
  let totalInPreviousRings = 0;
  
  for (let i = 0; i < itemsPerRing.length; i++) {
    if (index < totalInPreviousRings + itemsPerRing[i]) {
      ring = i;
      positionInRing = index - totalInPreviousRings;
      break;
    }
    totalInPreviousRings += itemsPerRing[i];
  }
  
  // Si on dépasse les anneaux prévus, continuer sur le dernier
  if (ring >= itemsPerRing.length) {
    ring = itemsPerRing.length - 1;
    positionInRing = (index - totalInPreviousRings) % 18;
  }
  
  let x, y, rotation;
  
  if (ring === 0) {
    // Centre exact
    x = 50;
    y = 50;
    rotation = -12 + Math.random() * 24;
  } else {
    // Anneaux concentriques avec rayons plus proches
    const radiusMultipliers = [0, 12, 22, 32, 42];
    const radius = radiusMultipliers[ring];
    const itemsInThisRing = itemsPerRing[ring];
    
    // Ajouter un offset aléatoire pour un look plus naturel
    const angleOffset = (Math.random() - 0.5) * 0.3;
    const angle = ((positionInRing / itemsInThisRing) * 2 * Math.PI) + angleOffset - Math.PI / 2;
    
    // Variation du rayon pour plus de naturel
    const radiusVariation = -3 + Math.random() * 6;
    const finalRadius = radius + radiusVariation;
    
    x = 50 + finalRadius * Math.cos(angle);
    y = 50 + finalRadius * Math.sin(angle);
    rotation = -18 + Math.random() * 36;
  }
  
  return {
    left: `${x}%`,
    top: `${y}%`,
    transform: `translate(-50%, -50%) rotate(${rotation}deg)`,
    animationDelay: `${index * 0.03}s`,
    zIndex: ring === 0 ? 10 : 5 - ring
  }
}

const startResize = (e) => {
  e.preventDefault()
  isResizing.value = true
  const sx = e.clientX
  const sw = sidebarWidth.value

  const move = (me) => {
    sidebarWidth.value = Math.max(250, Math.min(600, sw + (me.clientX - sx)))
  }

  const up = () => {
    isResizing.value = false
    document.removeEventListener('mousemove', move)
    document.removeEventListener('mouseup', up)
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
  }

  document.addEventListener('mousemove', move)
  document.addEventListener('mouseup', up)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

onMounted(() => {
  fetchArtists()
})
</script>

<template>
  <div class="vinyls-by-artist">
    <!-- Header -->
    <header class="page-header">
      <div class="header-top">
        <div class="title-section">
          <button @click="router.push('/dashboard')" class="back-button ghost-btn">
          <span class="icon">🏠</span>
          </button>
          <span class="title-icon">🎤</span>
          <h1>{{ selectedArtist ? selectedArtist.name : 'Disques par Artistes' }}</h1>
        </div>
        <div class="header-stats" v-if="!selectedArtist">
          <span class="stat-badge">{{ filteredArtists.length }} artistes</span>
          <span class="stat-badge">{{ filteredVinyls.length }} disques</span>
        </div>
        <button v-else @click="backToList" class="back-to-list-button">
          ← Retour à la liste
        </button>
      </div>
    </header>

    <!-- Vue liste des artistes -->
    <div v-if="!selectedArtist" class="artists-view">
      <!-- Filtres alphabet -->
      <div class="alphabet-filter">
        <button
          v-for="letter in alphabet"
          :key="letter"
          @click="selectedLetter = letter"
          class="letter-button"
          :class="{ 
            active: selectedLetter === letter,
            disabled: !hasArtistsForLetter(letter)
          }"
          :disabled="!hasArtistsForLetter(letter) && letter !== 'All'"
        >
          {{ letter }}
        </button>
      </div>

      <!-- Bouton toggle mobile (visible seulement en mobile) -->
      <button 
        class="mobile-toggle-btn"
        @click="toggleSidebarMobile"
        v-if="!selectedArtist"
        :title="isSidebarCollapsed ? 'Afficher la liste' : 'Masquer la liste'"
      >
        <span v-if="!isSidebarCollapsed">✖</span>
        <span v-else>📋</span>
      </button>

      <!-- Sidebar avec liste des artistes -->
      <div class="content-layout" :style="gridStyle">
        <aside class="artists-sidebar" :class="{ collapsed: isSidebarCollapsed }">
          <div class="search-box" v-show="!isSidebarCollapsed">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search artist..."
              class="search-input"
            />
            <span class="search-icon">🔍</span>
          </div>

          <div class="artists-list" v-show="!isSidebarCollapsed">
            <div class="list-header" @click="toggleSort" role="button" tabindex="0" :title="sortOrder === 'asc' ? 'Tri A→Z' : 'Tri Z→A'">
              <span class="list-title">[All Music]</span>
              <span class="list-count-with-sort">
                <span class="count-number">{{ filteredArtists.length }}</span>
                <span class="sort-icon" v-html="getSortIcon()"></span>
              </span>
            </div>

            <button
              v-for="artist in filteredArtists"
              :key="artist.id"
              @click="selectArtist(artist)"
              class="artist-item"
            >
              <span class="artist-name">{{ artist.name }}</span>
              <span class="artist-count">{{ artist.vinyl_count }}</span>
            </button>
          </div>

          <div 
            class="resize-handle"
            @mousedown="startResize"
            :class="{ 'is-resizing': isResizing }"
          ></div>
        </aside>

        <!-- Patchwork circulaire dense et naturel -->
        <main class="vinyls-grid-preview">
          <div class="empty-state" v-if="filteredVinyls.length === 0">
            <span class="empty-icon">🎵</span>
            <p>Aucun disque trouvé</p>
          </div>
          <div v-else class="circular-patchwork">
            <div
              v-for="(vinyl, index) in filteredVinyls.slice(0, 49)"
              :key="vinyl.id"
              class="polaroid-card"
              :class="{ 'center-card': index === 0 }"
              :style="getCircularPosition(index, Math.min(filteredVinyls.length, 49))"
            >
              <div class="polaroid-inner">
                <div class="polaroid-image">
                  <img
                    v-if="vinyl.cover_url"
                    :src="normalizeCoverUrl(vinyl.cover_url)"
                    :alt="vinyl.title"
                    @error="handleImageError"
                  />
                  <div class="mosaic-fallback">
                    <span class="fallback-icon">💿</span>
                  </div>
                </div>
                <div class="polaroid-caption">
                  <p class="polaroid-title">{{ vinyl.title }}</p>
                </div>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>

    <!-- Vue détails artiste : liste des albums + fiche détaillée -->
    <div v-else class="artist-detail-view">
      <div class="detail-layout">
        <!-- Liste des albums de l'artiste -->
        <aside class="albums-sidebar">
          <div class="albums-sidebar-header">
            <span>Albums</span>
            <span class="albums-count">{{ sortedArtistVinyls.length }}</span>
          </div>
          <div class="albums-list">
            <button
              v-for="vinyl in sortedArtistVinyls"
              :key="vinyl.id"
              class="album-list-item"
              :class="{ active: selectedVinyl?.id === vinyl.id }"
              @click="selectVinyl(vinyl)"
            >
              <span class="album-year">{{ vinyl.release_year || '—' }}</span>
              <span class="album-title">{{ vinyl.title }}</span>
            </button>
            <div v-if="!sortedArtistVinyls.length" class="empty-state small">
              <span class="empty-icon">💿</span>
              <p>Aucun disque pour cet artiste</p>
            </div>
          </div>
        </aside>

        <!-- Fiche détaillée de l'album sélectionné -->
        <section class="vinyl-detail-panel" v-if="selectedVinyl">
          <div class="detail-content">
            <div class="detail-fields">
              <div class="field-row">
                <div class="field">
                  <label>Artiste</label>
                  <div class="field-value">{{ selectedArtist.name }}</div>
                </div>
                <div class="field">
                  <label>Titre</label>
                  <div class="field-value">{{ selectedVinyl.title }}</div>
                </div>
              </div>
              <div class="field-row">
                <div class="field">
                  <label>Genre</label>
                  <div class="field-value">{{ selectedVinyl.genre_name || '—' }}</div>
                </div>
                <div class="field">
                  <label>Format</label>
                  <div class="field-value">{{ selectedVinyl.format_name || '—' }}</div>
                </div>
              </div>
              <div class="field-row">
                <div class="field">
                  <label>Éditeur</label>
                  <div class="field-value">{{ selectedVinyl.label_name || '—' }}</div>
                </div>
                <div class="field">
                  <label>Pays</label>
                  <div class="field-value">{{ selectedVinyl.country_name || '—' }}</div>
                </div>
              </div>
              <div class="field-row">
                <div class="field">
                  <label>Année</label>
                  <div class="field-value">{{ selectedVinyl.release_year || '—' }}</div>
                </div>
                <div class="field">
                  <label>Code-barres</label>
                  <div class="field-value">{{ selectedVinyl.barcode || '—' }}</div>
                </div>
              </div>
              <div class="field-row">
                <div class="field">
                  <label>Prix</label>
                  <div class="field-value">{{ formatPrice(selectedVinyl.price) }}</div>
                </div>
                <div class="field">
                  <label>Quantité</label>
                  <div class="field-value">{{ selectedVinyl.quantity || 1 }}</div>
                </div>
              </div>
              <div class="field-row single">
                <div class="field full">
                  <label>Commentaires</label>
                  <div class="field-value textarea">{{ selectedVinyl.notes || '—' }}</div>
                </div>
              </div>

              <div class="tracklist-columns">
                <div class="tracklist-face">
                  <h4>Face A</h4>
                  <p v-if="tracksLoading" class="tracklist-loading">Chargement…</p>
                  <ol v-else-if="faceATracks.length">
                    <li v-for="(t, i) in faceATracks" :key="t.id ?? i">
                      <span class="track-pos">{{ t.position || i + 1 }}</span>
                      <span class="track-title">{{ t.title }}</span>
                      <span class="track-duration" v-if="t.duration">{{ t.duration }}</span>
                    </li>
                  </ol>
                  <p v-else class="tracklist-empty">Aucune piste</p>
                </div>
                <div class="tracklist-face">
                  <h4>Face B</h4>
                  <p v-if="tracksLoading" class="tracklist-loading">Chargement…</p>
                  <ol v-else-if="faceBTracks.length">
                    <li v-for="(t, i) in faceBTracks" :key="t.id ?? i">
                      <span class="track-pos">{{ t.position || i + 1 }}</span>
                      <span class="track-title">{{ t.title }}</span>
                      <span class="track-duration" v-if="t.duration">{{ t.duration }}</span>
                    </li>
                  </ol>
                  <p v-else class="tracklist-empty">Aucune piste</p>
                </div>
              </div>
            </div>

            <div class="detail-side">
              <div class="detail-cover">
                <img
                  v-if="selectedVinyl.cover_url"
                  :src="normalizeCoverUrl(selectedVinyl.cover_url)"
                  :alt="selectedVinyl.title"
                  @error="handleImageError"
                />
                <div class="cover-fallback" :style="{ display: selectedVinyl.cover_url ? 'none' : 'flex' }">
                  <span class="fallback-icon">💿</span>
                </div>
              </div>
              <StreamingButtons :disc="selectedVinyl" />
              <div class="detail-actions">
                <button type="button" class="detail-action-btn" @click="openTracklistModal(selectedVinyl)">
                  🎵 Gérer les pistes
                </button>
                <button type="button" class="detail-action-btn primary" @click="viewVinylDetails(selectedVinyl)">
                  ✏️ Modifier
                </button>
              </div>
            </div>
          </div>
        </section>

        <div v-else class="empty-state">
          <span class="empty-icon">💿</span>
          <p>Sélectionnez un album</p>
        </div>
      </div>
    </div>

    <!-- Loading overlay -->
    <div v-if="isLoading" class="loading-overlay">
      <div class="spinner">💿</div>
      <p>Chargement des artistes...</p>
    </div>

    <!-- Error message -->
    <div v-if="error" class="error-message">
      {{ error }}
      <button @click="fetchArtists" class="retry-button">Réessayer</button>
    </div>

    <TracklistModal
      v-model="isTracklistModalOpen"
      :disc="vinylForTracklist"
      @tracks-updated="handleTracksUpdated"
    />
  </div>
</template>

<style scoped>
.vinyls-by-artist {
  min-height: 100vh;
  background: transparent;
  padding: 20px;
}

/* ============================================
   HEADER
   ============================================ */
.page-header {
  margin-bottom: 20px;
}

.header-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.title-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-button {
  display: none; /* Masqué par défaut en desktop */
  padding: 0;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  align-items: center;
  justify-content: center;
  font-size: 1.3em;
}

@media (max-width: 767px) {
  .back-button {
    display: flex; /* Visible en mobile */
  }
}

.title-icon {
  font-size: 2em;
}

.page-header h1 {
  color: var(--text);
  font-size: 2em;
  margin: 0;
  font-weight: bold;
}

.header-stats {
  display: flex;
  gap: 12px;
}

.stat-badge {
  background: rgba(59, 130, 246, 0.14);
  color: var(--accent-soft);
  padding: 8px 16px;
  border-radius: 20px;
  font-weight: 600;
  font-size: 0.9em;
}

.back-to-list-button {
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  padding: 10px 20px;
  font-weight: 600;
  transition: all 0.3s;
}

.back-to-list-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

/* ============================================
   ALPHABET FILTER
   ============================================ */
.alphabet-filter {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(40px, 1fr));
  gap: 4px;
  margin-bottom: 20px;
  background: var(--panel-bg);
  backdrop-filter: blur(10px);
  padding: 10px;
  border-radius: 12px;
  border: 1px solid var(--line-soft);
  position: relative;
}

.letter-button {
  flex: 1 1 auto;
  background: rgba(var(--tint-rgb), 0.05);
  border: none;
  border-radius: 6px;
  color: var(--text);
  cursor: pointer;
  padding: 10px 6px;
  font-weight: 600;
  transition: all 0.2s;
  min-width: 36px;
  text-align: center;
}

.letter-button:hover {
  background: rgba(var(--tint-rgb), 0.1);
}

.letter-button.active {
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  color: white;
}

.letter-button.disabled {
  opacity: 0.3;
  cursor: not-allowed;
  background: rgba(var(--tint-rgb), 0.06);
  color: var(--text-dim);
}

.letter-button.disabled:hover {
  background: rgba(var(--tint-rgb), 0.06);
  transform: none;
}

/* ============================================
   CONTENT LAYOUT
   ============================================ */
.content-layout {
  display: grid;
  grid-template-columns: 400px 1fr;
  gap: 20px;
  min-height: 600px;
}

/* ============================================
   ARTISTS SIDEBAR
   ============================================ */
.artists-sidebar {
  background: var(--panel-bg);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px solid var(--line-soft);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: fit-content;
  max-height: 80vh;
  overflow: hidden;
}

.search-box {
  position: relative;
}

.search-input {
  width: 100%;
  padding: 12px 40px 12px 12px;
  background: rgba(var(--tint-rgb), 0.05);
  border: 1px solid var(--line);
  border-radius: 8px;
  color: var(--text);
  font-size: 1em;
}

.search-input::placeholder {
  color: var(--text-dim);
}

.search-icon {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 1.2em;
}

.sort-button:hover,
.view-button:hover {
  background: rgba(var(--tint-rgb), 0.1);
}

.artists-list {
  overflow-y: auto;
  max-height: calc(80vh - 180px);
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  border-radius: 8px;
  margin-bottom: 8px;
  font-weight: 600;
  cursor: pointer;
  user-select: none;
  transition: all 0.2s ease;
}

.list-header:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.list-title {
  color: white;
}

.list-count {
  background: rgba(255, 255, 255, 0.3);
  padding: 4px 12px;
  border-radius: 12px;
  color: white;
  font-size: 0.9em;
}

.list-count-with-sort {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.3);
  padding: 4px 12px;
  border-radius: 12px;
  color: white;
  font-size: 0.9em;
}

.count-number { 
  font-weight: 600; 
}

.sort-icon { 
  display: inline-flex; 
  align-items: center; 
  font-size: 1.1em; 
  opacity: 0.9; 
}

.artist-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: rgba(var(--tint-rgb), 0.04);
  border: none;
  border-radius: 8px;
  color: var(--text);
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 4px;
  width: 100%;
  text-align: left;
}

.artist-item:hover {
  background: rgba(var(--tint-rgb), 0.08);
  transform: translateX(4px);
}

.artist-name {
  font-weight: 500;
}

.artist-count {
  background: rgba(var(--tint-rgb), 0.12);
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 0.85em;
  color: var(--text);
}

/* ============================================
   PATCHWORK CIRCULAIRE STYLE POLAROID
   ============================================ */
.vinyls-grid-preview {
  background: transparent;
  border-radius: 12px;
  padding: 20px;
  min-height: 700px;
  overflow: visible;
  position: relative;
}

.circular-patchwork {
  position: relative;
  width: 100%;
  height: 700px;
}

.polaroid-card {
  position: absolute;
  width: 130px;
  opacity: 0;
  animation: fadeInRotate 0.8s ease-out forwards;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  cursor: pointer;
}

/* Carte centrale plus grande */
.polaroid-card.center-card {
  width: 160px;
}

.polaroid-card.center-card .polaroid-inner {
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
}

@keyframes fadeInRotate {
  from {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.5) rotate(0deg);
  }
  to {
    opacity: 1;
  }
}

.polaroid-card:hover {
  z-index: 200 !important;
  transform: translate(-50%, -50%) scale(1.2) rotate(0deg) !important;
  filter: drop-shadow(0 15px 30px rgba(0, 0, 0, 0.5));
}

.polaroid-card.center-card:hover {
  transform: translate(-50%, -50%) scale(1.25) rotate(0deg) !important;
}

.polaroid-inner {
  background: white;
  padding: 8px;
  padding-bottom: 32px;
  border-radius: 3px;
  box-shadow: 0 5px 20px rgba(0, 0, 0, 0.25);
}

.polaroid-image {
  position: relative;
  width: 100%;
  aspect-ratio: 1;
  background: #f5f5f5;
  overflow: hidden;
  border-radius: 2px;
}

.polaroid-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.mosaic-fallback {
  position: absolute;
  inset: 0;
  display: none;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
}

.fallback-icon {
  font-size: 2em;
  opacity: 0.7;
}

.polaroid-caption {
  padding: 6px 4px 0;
  text-align: center;
}

.polaroid-title {
  margin: 0;
  font-size: 0.7em;
  color: #333;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: 'Courier New', monospace;
}

/* ============================================
   ARTIST DETAIL — FICHE MAÎTRE/DÉTAIL
   ============================================ */
.artist-detail-view {
  margin-top: 20px;
}

.detail-layout {
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: 20px;
  align-items: start;
}

.albums-sidebar {
  background: var(--panel-bg);
  backdrop-filter: blur(10px);
  border: 1px solid var(--line-soft);
  border-radius: 12px;
  padding: 12px;
  max-height: 80vh;
  overflow-y: auto;
}

.albums-sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 10px;
  font-weight: 600;
  color: var(--text-dim);
  text-transform: uppercase;
  font-size: 0.8em;
  letter-spacing: 0.5px;
  margin-bottom: 8px;
}

.albums-count {
  background: rgba(var(--tint-rgb), 0.12);
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 0.9em;
}

.album-list-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  text-align: left;
  padding: 10px;
  background: rgba(var(--tint-rgb), 0.04);
  border: none;
  border-radius: 8px;
  color: var(--text);
  cursor: pointer;
  margin-bottom: 4px;
  transition: all 0.2s;
}

.album-list-item:hover {
  background: rgba(var(--tint-rgb), 0.08);
}

.album-list-item.active {
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  color: white;
}

.album-year {
  font-size: 0.85em;
  color: var(--text-dim);
  min-width: 36px;
}

.album-list-item.active .album-year {
  color: rgba(255, 255, 255, 0.85);
}

.album-title {
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.vinyl-detail-panel {
  background: var(--panel-bg);
  backdrop-filter: blur(10px);
  border: 1px solid var(--line-soft);
  border-radius: 12px;
  padding: 20px;
}

.detail-content {
  display: grid;
  grid-template-columns: 1fr 220px;
  gap: 24px;
}

.field-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 12px;
}

.field-row.single {
  grid-template-columns: 1fr;
}

.field label {
  display: block;
  font-size: 0.75em;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 4px;
}

.field-value {
  background: rgba(var(--tint-rgb), 0.05);
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 8px 12px;
  color: var(--text);
  font-size: 0.95em;
  min-height: 20px;
}

.field-value.textarea {
  min-height: 48px;
  white-space: pre-wrap;
}

.tracklist-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-top: 8px;
}

.tracklist-face h4 {
  margin: 0 0 8px 0;
  color: var(--text);
  font-size: 0.95em;
}

.tracklist-face ol {
  list-style: none;
  margin: 0;
  padding: 0;
  background: rgba(var(--tint-rgb), 0.04);
  border: 1px solid var(--line-soft);
  border-radius: 8px;
  overflow: hidden;
}

.tracklist-face li {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-bottom: 1px solid var(--line-soft);
  font-size: 0.88em;
}

.tracklist-face li:last-child {
  border-bottom: none;
}

.track-pos {
  color: var(--text-dim);
  min-width: 24px;
  font-weight: 600;
}

.track-title {
  flex: 1;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.track-duration {
  color: var(--text-dim);
  font-size: 0.9em;
}

.tracklist-empty,
.tracklist-loading {
  color: var(--text-dim);
  font-size: 0.85em;
  padding: 10px 0;
  margin: 0;
}

.detail-side {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 1;
  border-radius: 10px;
  overflow: hidden;
  background: rgba(var(--tint-rgb), 0.06);
}

.detail-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-fallback {
  position: absolute;
  inset: 0;
  display: none;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
}

.detail-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-action-btn {
  padding: 10px;
  border-radius: 8px;
  border: 1px solid var(--line);
  background: rgba(var(--tint-rgb), 0.05);
  color: var(--text);
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.detail-action-btn:hover {
  background: rgba(var(--tint-rgb), 0.1);
}

.detail-action-btn.primary {
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  border: none;
  color: white;
}

.detail-action-btn.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.empty-state.small {
  padding: 30px 10px;
}

/* ============================================
   EMPTY STATE
   ============================================ */
.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-soft);
}

.empty-icon {
  font-size: 4em;
  display: block;
  margin-bottom: 16px;
}

.empty-state p {
  font-size: 1.2em;
  margin: 0;
}

/* ============================================
   LOADING & ERROR
   ============================================ */
.loading-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  color: white;
}

.loading-overlay p {
  color: white;
  margin-top: 20px;
}

.spinner {
  font-size: 4em;
  animation: spin 2s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-message {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.28);
  border-radius: 8px;
  color: var(--negative-text);
  padding: 12px 16px;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.retry-button {
  background: rgba(239, 68, 68, 0.85);
  border: none;
  border-radius: 4px;
  color: white;
  cursor: pointer;
  padding: 8px 12px;
  font-size: 0.9em;
}

.retry-button:hover {
  background: rgba(220, 38, 38, 0.95);
}

/* ============================================
   RESPONSIVE
   ============================================ */
@media (max-width: 1024px) {
  .resize-handle { display: none !important; }
  .content-layout {
    grid-template-columns: 300px 1fr;
  }

  .detail-layout {
    grid-template-columns: 220px 1fr;
  }

  .detail-content {
    grid-template-columns: 1fr 180px;
  }

  .polaroid-card {
    width: 100px;
  }
  
  .polaroid-card.center-card {
    width: 130px;
  }
  
  .circular-patchwork {
    height: 600px;
  }
  
  .vinyls-grid-preview {
    min-height: 600px;
  }
}

@media (max-width: 768px) {
  /* Bouton toggle visible en mobile */
  .mobile-toggle-btn { 
    display: flex !important;
    top: 330px !important;
    right: 12px !important;
  }

  .resize-handle { 
    display: none !important; 
  }

  .content-layout {
    grid-template-columns: 1fr !important;
    gap: 12px;
  }

  .artists-sidebar {
    max-height: 400px;
    transition: all 0.3s ease;
  }

  /* Sidebar collapsed = disparaît */
  .artists-sidebar.collapsed {
    display: none !important;
  }

  .detail-layout {
    grid-template-columns: 1fr;
  }

  .detail-content {
    grid-template-columns: 1fr;
  }

  .field-row {
    grid-template-columns: 1fr;
  }

  .tracklist-columns {
    grid-template-columns: 1fr;
  }

  .polaroid-card {
    width: 70px;
  }

  .polaroid-card.center-card {
    width: 100px;
  }

  .circular-patchwork {
    height: 500px;
  }

  .vinyls-grid-preview {
    min-height: 500px;
  }

  .page-header h1 {
    font-size: 1.5em;
  }
}

/* Les boutons de streaming utilisent désormais le composant partagé
   StreamingButtons.vue (voir plus haut dans le template) — ses styles
   vivent dans ce composant, plus ici. */

/* REDIMENSIONNEMENT */
.artists-sidebar { 
  position: relative; 
}

.resize-handle {
  position: absolute;
  top: 0;
  right: -10px;
  width: 20px;
  height: 100%;
  cursor: col-resize;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.resize-handle::after {
  content: '';
  position: absolute;
  width: 4px;
  height: 50px;
  background: rgba(59, 130, 246, 0.4);
  border-radius: 2px;
  transition: all 0.2s;
}

.resize-handle:hover::after {
  background: rgba(59, 130, 246, 0.7);
  height: 80px;
  width: 5px;
}

.resize-handle.is-resizing::after {
  background: var(--accent);
  height: 100%;
  width: 3px;
}

/* BOUTON TOGGLE MOBILE */
.mobile-toggle-btn {
  display: none;
  position: fixed;
  top: 196px;
  right: 18px;
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  border: 2px solid rgba(255, 255, 255, 0.5);
  border-radius: 10px;
  width: 34px;
  height: 34px;
  font-size: 1em;
  cursor: pointer;
  z-index: 1001;
  transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  box-shadow:
    0 3px 10px rgba(59, 130, 246, 0.5),
    0 2px 4px rgba(0, 0, 0, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(10px);
  color: white;
  font-weight: bold;
}

.mobile-toggle-btn:hover {
  transform: scale(1.1) rotate(6deg);
  box-shadow:
    0 5px 16px rgba(59, 130, 246, 0.7),
    0 3px 6px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.4);
  background: linear-gradient(135deg, var(--accent-blue), var(--accent));
  border-color: rgba(255, 255, 255, 0.7);
}

.mobile-toggle-btn:active {
  transform: scale(0.92) rotate(0deg);
  box-shadow:
    0 2px 6px rgba(59, 130, 246, 0.6),
    0 1px 2px rgba(0, 0, 0, 0.25),
    inset 0 -1px 0 rgba(0, 0, 0, 0.15);
}

.mobile-toggle-btn span {
  display: flex;
  align-items: center;
  justify-content: center;
  filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.4));
}
</style>
