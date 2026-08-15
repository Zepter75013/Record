<script setup>
import { computed, ref } from 'vue'
import { useStreamingPreferencesStore } from '@/stores/streamingPreferences'
import { getAuthToken } from '@/utils/authToken'

const props = defineProps({
  disc: {
    type: Object,
    required: true
  }
})

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'

const streamingPrefs = useStreamingPreferencesStore()
const isResolving = ref(false)

// Recherche + sélection automatique de l'album exact via une API publique du service
// (pas besoin de lien direct stocké en base). Renvoie l'URL à ouvrir, ou null si rien trouvé.
async function resolveAppleMusic(query) {
  const res = await fetch(`https://itunes.apple.com/search?entity=album&limit=1&country=FR&term=${query}`)
  const data = await res.json()
  return data.results?.[0]?.collectionViewUrl || null
}

async function resolveDeezer(query) {
  const token = getAuthToken()
  const res = await fetch(`${API_BASE_URL}/streaming/deezer-search?q=${query}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  if (!res.ok) return null
  const data = await res.json()
  return data.id ? `https://widget.deezer.com/widget/dark/album/${data.id}?autoplay=true` : null
}

const PLATFORMS = {
  apple_music: {
    label: 'Apple Music',
    class: 'apple-music',
    searchUrl: (query) => `https://music.apple.com/fr/search?term=${query}`,
    resolve: resolveAppleMusic,
  },
  spotify: {
    label: 'Spotify',
    class: 'spotify',
    searchUrl: (query) => `https://open.spotify.com/search/${query}`,
  },
  deezer: {
    label: 'Deezer',
    class: 'deezer',
    searchUrl: (query) => `https://www.deezer.com/fr/search/${query}`,
    resolve: resolveDeezer,
  },
  youtube: {
    label: 'YouTube',
    class: 'youtube',
    searchUrl: (query) => `https://www.youtube.com/results?search_query=${query}`,
  },
}

const platform = computed(() => PLATFORMS[streamingPrefs.preferredPlatform] || null)

// Toujours dérivé du disque (artiste + titre) — plus de lien stocké en base à maintenir.
const query = computed(() => {
  if (!props.disc.artist_name || !props.disc.title) return null
  return encodeURIComponent(`${props.disc.artist_name} ${props.disc.title}`)
})

const searchUrl = computed(() => {
  if (!platform.value || !query.value) return null
  return platform.value.searchUrl(query.value)
})

async function openStreaming() {
  if (!platform.value.resolve) {
    window.open(searchUrl.value, '_blank')
    return
  }

  isResolving.value = true
  let url = searchUrl.value
  try {
    url = (await platform.value.resolve(query.value)) || searchUrl.value
  } catch {
    // garde l'URL de recherche générique en repli
  } finally {
    isResolving.value = false
  }

  // Si le navigateur bloque l'ouverture d'onglet après l'attente asynchrone
  // (bloqueur de pop-up), on navigue dans l'onglet courant plutôt que de perdre l'action.
  const win = window.open(url, '_blank')
  if (!win) window.location.href = url
}
</script>

<template>
  <button
    v-if="platform && query"
    @click.stop="openStreaming"
    class="cover-play-btn"
    :class="platform.class"
    :disabled="isResolving"
    :title="`Écouter sur ${platform.label}`"
    :aria-label="`Écouter sur ${platform.label}`"
  >
    <span class="play-glyph">{{ isResolving ? '⏳' : '▶' }}</span>
  </button>
</template>

<style scoped>
/* Ce bouton est prévu pour flotter sur une pochette (parent en position: relative). */
.cover-play-btn {
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  color: white;
  box-shadow: 0 3px 12px rgba(0, 0, 0, 0.4);
  transition: opacity 0.25s ease, transform 0.2s ease;
  z-index: 2;
}

.cover-play-btn:disabled {
  cursor: wait;
}

.play-glyph {
  line-height: 1;
  margin-left: 2px; /* recentrage optique du triangle ▶ */
}

/* Écrans avec un vrai survol (desktop) : bouton centré, révélé au survol de la carte */
@media (hover: hover) {
  .cover-play-btn {
    inset: 0;
    margin: auto;
    width: 56px;
    height: 56px;
    font-size: 1.4em;
    opacity: 0;
  }

  .vinyl-card:hover .cover-play-btn {
    opacity: 1;
  }

  .cover-play-btn:hover {
    transform: scale(1.1);
  }
}

/* Tactile (mobile/tablette, pas de survol) : toujours visible, discret, coin bas-droit */
@media (hover: none) {
  .cover-play-btn {
    bottom: 8px;
    right: 8px;
    width: 38px;
    height: 38px;
    font-size: 1.1em;
    opacity: 1;
  }
}

/* Couleurs de marque par plateforme */
.apple-music { background: linear-gradient(135deg, #fc3c44 0%, #ff0033 100%); }
.spotify { background: linear-gradient(135deg, #1DB954 0%, #1ed760 100%); }
.deezer { background: linear-gradient(135deg, #FF0092 0%, #ff3cb4 100%); }
.youtube { background: linear-gradient(135deg, #FF0000 0%, #cc0000 100%); }
</style>
