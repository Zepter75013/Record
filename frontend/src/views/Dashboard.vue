<!-- Dashboard.vue - FINAL avec sous-menus déroulants + Auto-logout -->
<script setup>
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { ref, computed, onMounted, onActivated, watch, onBeforeUnmount } from 'vue'
import { useAuthStore } from '@/stores/auth'

// Composants
import StatsWidget from '@/components/StatsWidget.vue'
import PieChart from '@/components/PieChart.vue'
import RecentDiscs from '@/components/RecentDiscs.vue'
import RecentGames from '@/components/RecentGames.vue'
import MapWidget from '@/components/MapWidget.vue'
import AboutModal from '@/components/AboutModal.vue'

// Composables
import { useStats } from '@/composables/useStats'
import { useGameStats } from '@/composables/useGameStats'
import { useTooltips } from '@/composables/useTooltips'
import { useIdleTimeout } from '@/composables/useIdleTimeout'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// États
const sidebarOpenMobile = ref(false)
const windowWidth = ref(window.innerWidth)
const showArtistsMenu = ref(false)
const isAboutModalOpen = ref(false)

// --- Sidebar : largeur, réduction et redimensionnement (persistés) ---
const WIDTH_STORAGE_KEY = 'sidebarWidth'
const COLLAPSED_STORAGE_KEY = 'sidebarCollapsed'
const MIN_WIDTH = 220
const MAX_WIDTH = 340
const DEFAULT_WIDTH = 280
const COLLAPSED_WIDTH = 80

function clampWidth(width) {
  return Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, width))
}

const storedWidth = Number(localStorage.getItem(WIDTH_STORAGE_KEY))
const sidebarWidth = ref(clampWidth(storedWidth > 0 ? storedWidth : DEFAULT_WIDTH))
const isSidebarCollapsed = ref(localStorage.getItem(COLLAPSED_STORAGE_KEY) === '1')
const isResizing = ref(false)

// Tiroir mobile : la sidebar bascule en tiroir en recouvrement en dessous de
// 920px, indépendamment du seuil 768px utilisé par le contenu (tableaux/
// cartes des vues CRUD, non concerné par cette passe de restylage).
const SIDEBAR_MOBILE_QUERY = '(max-width: 920px)'
const sidebarMobileQuery = window.matchMedia(SIDEBAR_MOBILE_QUERY)
const isMobile = ref(sidebarMobileQuery.matches)

function handleSidebarMobileQueryChange(event) {
  isMobile.value = event.matches
  if (!isMobile.value) {
    sidebarOpenMobile.value = false
  }
}

const effectiveCollapsed = computed(() => isSidebarCollapsed.value && !isMobile.value)

const currentSidebarWidth = computed(() => {
  if (isMobile.value) return sidebarWidth.value
  return isSidebarCollapsed.value ? COLLAPSED_WIDTH : sidebarWidth.value
})

function handleResizeMove(event) {
  sidebarWidth.value = clampWidth(event.clientX)
}

function stopResize() {
  if (!isResizing.value) return

  isResizing.value = false
  document.body.style.removeProperty('cursor')
  document.body.style.removeProperty('user-select')
  window.removeEventListener('mousemove', handleResizeMove)
  window.removeEventListener('mouseup', stopResize)
  localStorage.setItem(WIDTH_STORAGE_KEY, String(sidebarWidth.value))
}

function resetSidebarWidth() {
  sidebarWidth.value = DEFAULT_WIDTH
  localStorage.setItem(WIDTH_STORAGE_KEY, String(DEFAULT_WIDTH))
}

function startResize(event) {
  if (isSidebarCollapsed.value) return

  isResizing.value = true
  event.preventDefault()
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
  window.addEventListener('mousemove', handleResizeMove)
  window.addEventListener('mouseup', stopResize)
}

// État pour les sous-menus
const openSubmenus = ref({
  artists: false,
  settings: false
})

// État pour les sections repliables de la barre latérale (Disques / Jeux),
// mémorisé comme le collapse global de la sidebar
const SECTIONS_STORAGE_KEY = 'sidebarSectionsOpen'
function loadSectionsOpen() {
  try {
    const raw = localStorage.getItem(SECTIONS_STORAGE_KEY)
    if (!raw) return { discs: true, games: true }
    const parsed = JSON.parse(raw)
    return { discs: parsed.discs !== false, games: parsed.games !== false }
  } catch {
    return { discs: true, games: true }
  }
}
const sectionsOpen = ref(loadSectionsOpen())
const toggleSection = (key) => {
  sectionsOpen.value[key] = !sectionsOpen.value[key]
  localStorage.setItem(SECTIONS_STORAGE_KEY, JSON.stringify(sectionsOpen.value))
}

// 🆕 Auto-déconnexion après 1h d'inactivité avec avertissement à 55min
const { idle, warning, remainingTime, reset } = useIdleTimeout({
 timeout: 60 * 60 * 1000,      // 1 heure
  warningTime: 55 * 60 * 1000   // Avertissement à 55 minutes
})

// 🆕 Surveiller l'état idle pour déconnexion automatique
watch(idle, (isIdle) => {
  if (isIdle) {
    console.log('🔒 Déconnexion automatique après 1h d\'inactivité')
    handleLogout()
  }
})

// 🆕 Fonction pour rester connecté (appelée depuis le modal)
const stayConnected = () => {
  reset()
}

// Composables
const {
  stats,
  genreDistribution,
  formatDistribution,
  artistDistribution,
  recentDiscs,
  countries,
  isLoading,
  error,
  fetchStats,
  fetchGenreDistribution,
  fetchFormatDistribution,
  fetchArtistDistribution,
  fetchRecentDiscs,
  fetchCountries,
  normalizeCoverUrl
} = useStats()

const {
  stats: gameStats,
  platformDistribution,
  genreDistribution: gameGenreDistribution,
  publisherDistribution,
  recentGames,
  fetchGameStats,
  fetchPlatformDistribution,
  fetchGameGenreDistribution,
  fetchPublisherDistribution,
  fetchRecentGames
} = useGameStats()

const {
  tooltip,
  showTooltip,
  hideTooltip,
  updateTooltipPosition
} = useTooltips()

// Choix du tableau de bord affiché sur l'écran Accueil (Disques ou Jeux),
// mémorisé comme les autres préférences d'affichage de la sidebar
const DASHBOARD_MODE_KEY = 'dashboardMode'
const dashboardMode = ref(
  localStorage.getItem(DASHBOARD_MODE_KEY) === 'games' ? 'games' : 'discs'
)
const setDashboardMode = (mode) => {
  dashboardMode.value = mode
  localStorage.setItem(DASHBOARD_MODE_KEY, mode)
}

// Config
const widgetConfig = {
  vinyls: { icon: '💿', color: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', title: 'Disques' },
  artists: { icon: '🎤', color: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)', title: 'Artistes' },
  genres: { icon: '🎵', color: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)', title: 'Genres' },
  formats: { icon: '💽', color: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)', title: 'Formats' },
  countries: { icon: '🌍', color: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)', title: 'Pays' },
  labels: { icon: '🏷️', color: 'linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)', title: 'Labels' },
  games: { icon: '🎮', color: 'linear-gradient(135deg, #f6416c 0%, #ff9a44 100%)', title: 'Jeux' },
  platforms: { icon: '🕹️', color: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)', title: 'Plateformes' },
  publishers: { icon: '🏢', color: 'linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)', title: 'Éditeurs' }
}

// Format ORIGINAL des couleurs (objets, pas strings)
const pieColors = [
  { solid: '#667eea' },
  { solid: '#f093fb' },
  { solid: '#4facfe' },
  { solid: '#43e97b' },
  { solid: '#fa709a' },
  { solid: '#a8edea' },
  { solid: '#764ba2' },
  { solid: '#f5576c' },
  { solid: '#00f2fe' },
  { solid: '#38f9d7' }
]

// --- Chips dégradées pour les icônes de la sidebar (même style que Finance) ---
const NAV_ICON_COLORS = {
  home: '#7aa2f7',
  discs: '#e0a15c',
  games: '#e2685c',
  platforms: '#6c9bd1',
  gamegenres: '#c97fd4',
  publishers: '#f0b429',
  artists: '#b18cf0',
  'artists-manage': '#9aa5b1',
  'artists-by-disc': '#e0a15c',
  genres: '#5ecb8f',
  formats: '#5cc8a0',
  countries: '#f07aa5',
  labels: '#e0c15c',
  preferences: '#9aa5b1',
  profile: '#5cc8d1',
  reports: '#e0c15c',
  about: '#8fa8a0'
}

function shade(hex, percent) {
  const r = parseInt(hex.slice(1, 3), 16)
  const g = parseInt(hex.slice(3, 5), 16)
  const b = parseInt(hex.slice(5, 7), 16)
  const target = percent < 0 ? 0 : 255
  const p = Math.abs(percent)

  const mix = (channel) => Math.round((target - channel) * p) + channel

  return `rgb(${mix(r)}, ${mix(g)}, ${mix(b)})`
}

function iconGradient(name) {
  const base = NAV_ICON_COLORS[name] || '#8fa8a0'
  const light = shade(base, 0.4)
  const dark = shade(base, -0.3)
  return `linear-gradient(150deg, ${light} 0%, ${base} 55%, ${dark} 100%)`
}

// Actions de navigation
const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
  localStorage.setItem(COLLAPSED_STORAGE_KEY, isSidebarCollapsed.value ? '1' : '0')
}

const toggleSidebarMobile = () => {
  sidebarOpenMobile.value = !sidebarOpenMobile.value
}

// Toggle sous-menu
const toggleSubmenu = (menuKey) => {
  openSubmenus.value[menuKey] = !openSubmenus.value[menuKey]
}

const handleLogout = () => {
  console.log('Déconnexion...')
  authStore.logout()
}

// Navigation pour les widgets
const navigateToSection = (section) => {
  console.log('Dashboard: Navigation vers', section)

  // Si c'est le widget artistes sur mobile, afficher le menu contextuel
  if (section === 'artists' && windowWidth.value < 768) {
    showArtistsMenu.value = !showArtistsMenu.value
    return
  }

  const routes = {
    vinyls: '/dashboard/vinyls',
    artists: '/dashboard/settings/artists',
    genres: '/dashboard/settings/genres',
    formats: '/dashboard/settings/formats',
    countries: '/dashboard/settings/countries',
    labels: '/dashboard/settings/labels'
  }

  const targetRoute = routes[section]
  if (targetRoute) {
    router.push(targetRoute)
  } else {
    console.warn('Dashboard: Route inconnue pour', section)
  }
}

const isLinkActive = (path) => {
  return route.path === path || route.path.startsWith(path + '/')
}

const refreshAll = async () => {
  console.log('Dashboard: Rafraîchissement des données...')
  try {
    await Promise.all([
      fetchStats(),
      fetchCountries(),
      fetchRecentDiscs(),
      fetchGenreDistribution(),
      fetchFormatDistribution(),
      fetchArtistDistribution(),
      fetchGameStats(),
      fetchRecentGames(),
      fetchPlatformDistribution(),
      fetchGameGenreDistribution(),
      fetchPublisherDistribution()
    ])
  } catch (err) {
    console.error('Erreur lors du rafraîchissement', err)
  }
}

const handleImageError = ({ event }) => {
  event.target.style.display = 'none'
  const fallback = event.target.parentElement?.querySelector('.cover-fallback')
  if (fallback) fallback.style.display = 'flex'
}

// Gestion resize
const updateWindowWidth = () => {
  windowWidth.value = window.innerWidth
}

onMounted(async () => {
  await refreshAll()
  window.addEventListener('resize', updateWindowWidth)
  sidebarMobileQuery.addEventListener('change', handleSidebarMobileQueryChange)
})

onActivated(async () => {
  await refreshAll()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateWindowWidth)
  window.removeEventListener('mousemove', handleResizeMove)
  window.removeEventListener('mouseup', stopResize)
  sidebarMobileQuery.removeEventListener('change', handleSidebarMobileQueryChange)
})

watch(() => route.path, async (newPath) => {
  if (newPath === '/dashboard') {
    await refreshAll()
  }
  if (isMobile.value) {
    sidebarOpenMobile.value = false
  }
})
</script>

<template>
  <div class="dashboard-layout" :class="{ 'sidebar-collapsed': effectiveCollapsed }">

    <!-- Bouton hamburger (tiroir mobile, <920px) -->
    <button
      v-if="isMobile"
      class="mobile-menu-toggle"
      type="button"
      aria-label="Ouvrir le menu"
      @click="sidebarOpenMobile = true"
    >
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path d="M4 6h16M4 12h16M4 18h16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </svg>
    </button>

    <div v-if="isMobile && sidebarOpenMobile" class="mobile-backdrop" @click="sidebarOpenMobile = false"></div>

    <!-- Sidebar -->
    <aside
      class="sidebar"
      :class="{
        'is-collapsed': effectiveCollapsed,
        'is-resizing': isResizing,
        'is-mobile': isMobile,
        'is-mobile-open': isMobile && sidebarOpenMobile
      }"
      :style="{ width: currentSidebarWidth + 'px', minWidth: currentSidebarWidth + 'px' }"
      @click.stop
    >
      <div class="sidebar-scroll">
        <div class="brand">
          <RouterLink to="/dashboard" class="brand-mark" aria-label="Accueil" @click="sidebarOpenMobile = false">💿</RouterLink>
          <div v-if="!effectiveCollapsed">
            <strong>Disques Manager</strong>
            <p>Collection vinyle</p>
          </div>

          <button
            v-if="isMobile"
            class="mobile-close-btn"
            type="button"
            aria-label="Fermer le menu"
            @click="sidebarOpenMobile = false"
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M6 6l12 12M18 6L6 18" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
            </svg>
          </button>
        </div>

        <nav class="sidebar-nav" aria-label="Navigation principale">
          <RouterLink to="/dashboard" class="nav-item" :class="{ active: route.path === '/dashboard' }" @click="sidebarOpenMobile = false">
            <span class="nav-icon-chip" :style="{ background: iconGradient('home') }"><span class="nav-emoji">🏠</span></span>
            <span v-if="!effectiveCollapsed">Accueil</span>
            <span v-else class="nav-tooltip">Accueil</span>
          </RouterLink>

          <button
            v-if="!effectiveCollapsed"
            type="button"
            class="nav-section-toggle"
            :aria-expanded="sectionsOpen.discs"
            aria-controls="nav-section-discs"
            @click="toggleSection('discs')"
          >
            <span class="nav-section-label">Disques</span>
            <svg viewBox="0 0 24 24" aria-hidden="true" class="nav-group-chevron" :class="{ 'is-open': sectionsOpen.discs }">
              <path d="M9 6l6 6-6 6" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>
          <div v-else class="nav-section-divider" role="separator"></div>

          <div id="nav-section-discs" v-show="effectiveCollapsed || sectionsOpen.discs" class="nav-section-body">
            <RouterLink to="/dashboard/vinyls" class="nav-item nav-subitem" :class="{ active: isLinkActive('/dashboard/vinyls') }" @click="sidebarOpenMobile = false">
              <span class="nav-icon-chip" :style="{ background: iconGradient('discs') }"><span class="nav-emoji">📀</span></span>
              <span v-if="!effectiveCollapsed">Liste des disques</span>
              <span v-else class="nav-tooltip">Liste des disques</span>
            </RouterLink>

            <div class="nav-group" :class="{ 'nav-group-collapsed': effectiveCollapsed }">
              <button
                class="nav-item nav-subitem nav-group-toggle"
                :class="{ active: isLinkActive('/dashboard/settings/artists') || isLinkActive('/dashboard/vinyls/by-artist') }"
                type="button"
                aria-label="Artistes"
                @click="toggleSubmenu('artists')"
              >
                <span class="nav-icon-chip" :style="{ background: iconGradient('artists') }"><span class="nav-emoji">🎤</span></span>
                <span v-if="!effectiveCollapsed" class="nav-group-toggle-label">Artistes</span>
                <span v-else class="nav-tooltip">Artistes</span>
                <svg
                  v-if="!effectiveCollapsed"
                  viewBox="0 0 24 24"
                  aria-hidden="true"
                  class="nav-group-chevron"
                  :class="{ 'is-open': openSubmenus.artists }"
                >
                  <path d="M9 6l6 6-6 6" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </button>

              <transition name="submenu">
                <div v-show="openSubmenus.artists && !effectiveCollapsed" class="nav-children-rail has-rail">
                  <RouterLink
                    to="/dashboard/settings/artists"
                    class="nav-item nav-item-child"
                    :class="{ active: isLinkActive('/dashboard/settings/artists') }"
                    @click="sidebarOpenMobile = false"
                  >
                    <span class="nav-icon-chip" :style="{ background: iconGradient('artists-manage') }"><span class="nav-emoji">⚙️</span></span>
                    <span>Gérer les Artistes</span>
                  </RouterLink>

                  <RouterLink
                    to="/dashboard/vinyls/by-artist"
                    class="nav-item nav-item-child"
                    :class="{ active: isLinkActive('/dashboard/vinyls/by-artist') }"
                    @click="sidebarOpenMobile = false"
                  >
                    <span class="nav-icon-chip" :style="{ background: iconGradient('artists-by-disc') }"><span class="nav-emoji">💿</span></span>
                    <span>Disques par Artistes</span>
                  </RouterLink>
                </div>
              </transition>
            </div>

            <RouterLink to="/dashboard/settings/genres" class="nav-item nav-subitem" :class="{ active: isLinkActive('/dashboard/settings/genres') }" @click="sidebarOpenMobile = false">
              <span class="nav-icon-chip" :style="{ background: iconGradient('genres') }"><span class="nav-emoji">🎵</span></span>
              <span v-if="!effectiveCollapsed">Genres</span>
              <span v-else class="nav-tooltip">Genres</span>
            </RouterLink>

            <RouterLink to="/dashboard/settings/formats" class="nav-item nav-subitem" :class="{ active: isLinkActive('/dashboard/settings/formats') }" @click="sidebarOpenMobile = false">
              <span class="nav-icon-chip" :style="{ background: iconGradient('formats') }"><span class="nav-emoji">💽</span></span>
              <span v-if="!effectiveCollapsed">Formats</span>
              <span v-else class="nav-tooltip">Formats</span>
            </RouterLink>

            <RouterLink to="/dashboard/settings/countries" class="nav-item nav-subitem" :class="{ active: isLinkActive('/dashboard/settings/countries') }" @click="sidebarOpenMobile = false">
              <span class="nav-icon-chip" :style="{ background: iconGradient('countries') }"><span class="nav-emoji">🌍</span></span>
              <span v-if="!effectiveCollapsed">Pays</span>
              <span v-else class="nav-tooltip">Pays</span>
            </RouterLink>

            <RouterLink to="/dashboard/settings/labels" class="nav-item nav-subitem" :class="{ active: isLinkActive('/dashboard/settings/labels') }" @click="sidebarOpenMobile = false">
              <span class="nav-icon-chip" :style="{ background: iconGradient('labels') }"><span class="nav-emoji">🏷️</span></span>
              <span v-if="!effectiveCollapsed">Labels</span>
              <span v-else class="nav-tooltip">Labels</span>
            </RouterLink>
          </div>

          <button
            v-if="!effectiveCollapsed"
            type="button"
            class="nav-section-toggle"
            :aria-expanded="sectionsOpen.games"
            aria-controls="nav-section-games"
            @click="toggleSection('games')"
          >
            <span class="nav-section-label">Jeux</span>
            <svg viewBox="0 0 24 24" aria-hidden="true" class="nav-group-chevron" :class="{ 'is-open': sectionsOpen.games }">
              <path d="M9 6l6 6-6 6" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>
          <div v-else class="nav-section-divider" role="separator"></div>

          <div id="nav-section-games" v-show="effectiveCollapsed || sectionsOpen.games" class="nav-section-body">
            <RouterLink to="/dashboard/games" class="nav-item nav-subitem" :class="{ active: isLinkActive('/dashboard/games') }" @click="sidebarOpenMobile = false">
              <span class="nav-icon-chip" :style="{ background: iconGradient('games') }"><span class="nav-emoji">🎮</span></span>
              <span v-if="!effectiveCollapsed">Liste des jeux</span>
              <span v-else class="nav-tooltip">Liste des jeux</span>
            </RouterLink>

            <RouterLink to="/dashboard/settings/platforms" class="nav-item nav-subitem" :class="{ active: isLinkActive('/dashboard/settings/platforms') }" @click="sidebarOpenMobile = false">
              <span class="nav-icon-chip" :style="{ background: iconGradient('platforms') }"><span class="nav-emoji">🕹️</span></span>
              <span v-if="!effectiveCollapsed">Plateformes</span>
              <span v-else class="nav-tooltip">Plateformes</span>
            </RouterLink>

            <RouterLink to="/dashboard/settings/game-genres" class="nav-item nav-subitem" :class="{ active: isLinkActive('/dashboard/settings/game-genres') }" @click="sidebarOpenMobile = false">
              <span class="nav-icon-chip" :style="{ background: iconGradient('gamegenres') }"><span class="nav-emoji">🎯</span></span>
              <span v-if="!effectiveCollapsed">Genres de jeux</span>
              <span v-else class="nav-tooltip">Genres de jeux</span>
            </RouterLink>

            <RouterLink to="/dashboard/settings/publishers" class="nav-item nav-subitem" :class="{ active: isLinkActive('/dashboard/settings/publishers') }" @click="sidebarOpenMobile = false">
              <span class="nav-icon-chip" :style="{ background: iconGradient('publishers') }"><span class="nav-emoji">🏢</span></span>
              <span v-if="!effectiveCollapsed">Éditeurs</span>
              <span v-else class="nav-tooltip">Éditeurs</span>
            </RouterLink>
          </div>

          <p v-if="!effectiveCollapsed" class="nav-section-label">Général</p>
          <div v-else class="nav-section-divider" role="separator"></div>

          <RouterLink to="/dashboard/settings/reports" class="nav-item nav-subitem" :class="{ active: isLinkActive('/dashboard/settings/reports') }" @click="sidebarOpenMobile = false">
            <span class="nav-icon-chip" :style="{ background: iconGradient('reports') }"><span class="nav-emoji">📈</span></span>
            <span v-if="!effectiveCollapsed">Éditions</span>
            <span v-else class="nav-tooltip">Éditions</span>
          </RouterLink>

          <RouterLink to="/dashboard/settings/preferences" class="nav-item nav-subitem" :class="{ active: isLinkActive('/dashboard/settings/preferences') }" @click="sidebarOpenMobile = false">
            <span class="nav-icon-chip" :style="{ background: iconGradient('preferences') }"><span class="nav-emoji">⚙️</span></span>
            <span v-if="!effectiveCollapsed">Préférences</span>
            <span v-else class="nav-tooltip">Préférences</span>
          </RouterLink>

          <RouterLink to="/dashboard/settings/profile" class="nav-item nav-subitem" :class="{ active: isLinkActive('/dashboard/settings/profile') }" @click="sidebarOpenMobile = false">
            <span class="nav-icon-chip" :style="{ background: iconGradient('profile') }"><span class="nav-emoji">👤</span></span>
            <span v-if="!effectiveCollapsed">Mon compte</span>
            <span v-else class="nav-tooltip">Mon compte</span>
          </RouterLink>

          <button type="button" class="nav-item nav-subitem" @click="isAboutModalOpen = true; sidebarOpenMobile = false">
            <span class="nav-icon-chip" :style="{ background: iconGradient('about') }"><span class="nav-emoji">ℹ️</span></span>
            <span v-if="!effectiveCollapsed">À propos</span>
            <span v-else class="nav-tooltip">À propos</span>
          </button>
        </nav>
      </div>

      <div class="sidebar-footer">
        <div class="profile-card" :class="{ 'profile-card-collapsed': effectiveCollapsed }">
          <div class="avatar" :aria-label="authStore.user?.email || 'Compte'">
            {{ (authStore.user?.email || '?').charAt(0).toUpperCase() }}
            <span v-if="effectiveCollapsed" class="nav-tooltip">{{ authStore.user?.email || 'Compte' }}</span>
          </div>

          <template v-if="!effectiveCollapsed">
            <div class="profile-info">
              <strong>{{ authStore.user?.email || 'Compte' }}</strong>
              <p>Compte principal</p>
            </div>
            <div class="profile-actions">
              <RouterLink
                to="/dashboard/change-password"
                class="icon-action-btn"
                title="Changer le mot de passe"
                aria-label="Changer le mot de passe"
                @click="sidebarOpenMobile = false"
              >
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    d="M12 17a2 2 0 1 0 0-4 2 2 0 0 0 0 4zM12 13v-2M7 9V7a5 5 0 0 1 10 0v2"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.8"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                  <rect x="5" y="9" width="14" height="11" rx="2.5" fill="none" stroke="currentColor" stroke-width="1.8" />
                </svg>
              </RouterLink>

              <button
                class="icon-action-btn icon-action-btn-danger"
                type="button"
                title="Se déconnecter"
                aria-label="Se déconnecter"
                @click="handleLogout"
              >
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    d="M15 17l5-5-5-5M20 12H9M12 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h6"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.8"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </button>
            </div>
          </template>
        </div>

        <button
          v-if="!isMobile"
          class="collapse-toggle"
          type="button"
          :title="isSidebarCollapsed ? 'Afficher la barre latérale' : 'Masquer la barre latérale'"
          :aria-label="isSidebarCollapsed ? 'Afficher la barre latérale' : 'Masquer la barre latérale'"
          @click="toggleSidebar"
        >
          <svg viewBox="0 0 24 24" aria-hidden="true" :class="{ 'is-flipped': isSidebarCollapsed }">
            <path d="M15 5l-7 7 7 7" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <span v-if="isSidebarCollapsed">Afficher</span>
          <span v-else>Masquer</span>
        </button>
      </div>

      <div
        v-if="!effectiveCollapsed && !isMobile"
        class="resize-handle"
        role="separator"
        aria-orientation="vertical"
        aria-label="Redimensionner la barre latérale"
        @mousedown="startResize"
        @dblclick="resetSidebarWidth"
      ></div>
    </aside>

    <!-- Contenu principal -->
    <main class="content-area">
      <template v-if="route.name === 'Dashboard'">
        <header class="page-header">
          <div class="header-wrapper">
            <div class="header-top-row">
              <div class="header-title-container">
                <span class="title-icon">📊</span>
                <h1>Tableau de bord</h1>
              </div>

              <!-- Bouton déconnexion visible uniquement en mobile (<768px) -->
              <div class="toolbar" v-if="windowWidth < 768">
                <button @click="handleLogout" class="logout-header-button" title="Déconnexion">
                  <span class="icon">🚪</span>
                  <span class="text">Déconnexion</span>
                </button>
              </div>
            </div>

            <div class="dashboard-mode-toggle" role="tablist" aria-label="Choix du tableau de bord">
              <button
                type="button"
                role="tab"
                class="dashboard-mode-tab"
                :class="{ active: dashboardMode === 'discs' }"
                :aria-selected="dashboardMode === 'discs'"
                @click="setDashboardMode('discs')"
              >
                <span class="icon">📀</span> Disques
              </button>
              <button
                type="button"
                role="tab"
                class="dashboard-mode-tab"
                :class="{ active: dashboardMode === 'games' }"
                :aria-selected="dashboardMode === 'games'"
                @click="setDashboardMode('games')"
              >
                <span class="icon">🎮</span> Jeux
              </button>
            </div>
          </div>
        </header>

        <div v-if="error" class="error-message">
          {{ error }}
          <button @click="refreshAll" class="retry-button">Réessayer</button>
        </div>

        <!-- Widgets : Disques -->
        <template v-if="dashboardMode === 'discs'">
          <section class="dashboard-widgets">
            <!-- Widget principal des disques -->
            <StatsWidget
              type="vinyls"
              :value="stats.vinyls"
              :config="widgetConfig.vinyls"
              is-main
              @click="navigateToSection('vinyls')"
            />

            <!-- Grille des 5 thématiques -->
            <div class="thematics-row">
              <StatsWidget
                v-for="thematic in ['artists', 'genres', 'formats', 'countries', 'labels']"
                :key="thematic"
                :type="thematic"
                :value="stats[thematic]"
                :config="widgetConfig[thematic]"
                @click="navigateToSection(thematic)"
              />
            </div>

            <!-- Grille des 4 widgets : RecentDiscs + 3 PieCharts -->
            <div class="charts-row">
              <RecentDiscs
                :discs="recentDiscs"
                @image-error="handleImageError"
                class="recent-widget"
              />

              <PieChart
                :data="artistDistribution"
                :colors="pieColors"
                :total="stats.vinyls"
                title="Par artistes"
                type="artists"
                @slice-hover="showTooltip"
                @mouseleave="hideTooltip"
                class="pie-widget"
              />

              <PieChart
                :data="genreDistribution"
                :colors="pieColors"
                :total="stats.vinyls"
                title="Par genres"
                type="genres"
                @slice-hover="showTooltip"
                @mouseleave="hideTooltip"
                class="pie-widget"
              />

              <PieChart
                :data="formatDistribution"
                :colors="pieColors"
                :total="stats.vinyls"
                title="Par formats"
                type="formats"
                @slice-hover="showTooltip"
                @mouseleave="hideTooltip"
                class="pie-widget"
              />
            </div>
          </section>

          <!-- Carte -->
          <MapWidget
            :countries="countries"
            title="Carte des Pays"
            map-element-id="map-widget"
          />
        </template>

        <!-- Widgets : Jeux -->
        <template v-else>
          <section class="dashboard-widgets">
            <!-- Widget principal des jeux -->
            <StatsWidget
              type="games"
              :value="gameStats.games"
              :config="widgetConfig.games"
              is-main
              @click="router.push('/dashboard/games')"
            />

            <!-- Grille des 3 thématiques -->
            <div class="thematics-row">
              <StatsWidget
                v-for="thematic in ['platforms', 'genres', 'publishers']"
                :key="thematic"
                :type="thematic === 'genres' ? 'genres' : thematic"
                :value="thematic === 'genres' ? gameStats.genres : gameStats[thematic]"
                :config="widgetConfig[thematic]"
                @click="router.push('/dashboard/games')"
              />
            </div>

            <!-- Grille des 3 widgets : RecentGames + 2 PieCharts -->
            <div class="charts-row">
              <RecentGames
                :games="recentGames"
                @image-error="handleImageError"
                class="recent-widget"
              />

              <PieChart
                :data="platformDistribution"
                :colors="pieColors"
                :total="gameStats.games"
                title="Par plateforme"
                type="platforms"
                unit-label="jeux"
                @slice-hover="showTooltip"
                @mouseleave="hideTooltip"
                class="pie-widget"
              />

              <PieChart
                :data="gameGenreDistribution"
                :colors="pieColors"
                :total="gameStats.games"
                title="Par genre"
                type="genres"
                unit-label="jeux"
                @slice-hover="showTooltip"
                @mouseleave="hideTooltip"
                class="pie-widget"
              />

              <PieChart
                :data="publisherDistribution"
                :colors="pieColors"
                :total="gameStats.games"
                title="Par éditeur"
                type="publishers"
                unit-label="jeux"
                @slice-hover="showTooltip"
                @mouseleave="hideTooltip"
                class="pie-widget"
              />
            </div>
          </section>
        </template>

        <!-- Overlay de chargement -->
        <div v-show="isLoading" class="loading-overlay">
          <div class="spinner">💿</div>
          <p>Chargement des statistiques...</p>
        </div>
      </template>

      <RouterView v-else />
    </main>

    <!-- 🆕 MODAL D'AVERTISSEMENT D'INACTIVITÉ -->
    <Teleport to="body">
      <div v-if="warning" class="idle-warning-overlay">
        <div class="idle-warning-modal">
          <div class="warning-icon">⏰</div>
          <h3>Session expirée bientôt</h3>
          <p class="warning-text">Vous serez déconnecté dans</p>
          <div class="countdown">{{ remainingTime }}</div>
          <p class="countdown-label">secondes</p>
          <p class="hint">Cliquez sur le bouton ci-dessous pour rester connecté</p>
          <button @click="stayConnected" class="stay-button">
            Rester connecté
          </button>
        </div>
      </div>
    </Teleport>

    <!-- Tooltips -->
    <div v-if="tooltip.visible" class="pie-tooltip" :style="{ left: tooltip.x + 25 + 'px', top: tooltip.y - 80 + 'px' }">
      <div class="tooltip-content">
        <strong>{{ tooltip.name }}</strong>
        <div class="tooltip-count">{{ tooltip.count }} disque{{ tooltip.count > 1 ? 's' : '' }}</div>
        <div class="tooltip-percentage">{{ tooltip.percentage }}%</div>
      </div>
      <div class="tooltip-arrow"></div>
    </div>

    <!-- Menu contextuel artistes (mobile uniquement) -->
    <Teleport to="body">
      <div v-if="showArtistsMenu && windowWidth < 768" class="artists-context-menu-overlay" @click="showArtistsMenu = false">
        <div class="artists-context-menu" @click.stop>
          <h3>🎤 Artistes</h3>
          <p class="menu-subtitle">Choisissez une action</p>
          <button @click="router.push('/dashboard/settings/artists'); showArtistsMenu = false" class="menu-option">
            <span class="icon">⚙️</span>
            <div class="option-content">
              <span class="text">Gérer les Artistes</span>
              <span class="subtext">Ajouter, modifier, supprimer</span>
            </div>
          </button>
          <button @click="router.push('/dashboard/vinyls/by-artist'); showArtistsMenu = false" class="menu-option">
            <span class="icon">💿</span>
            <div class="option-content">
              <span class="text">Disques par Artistes</span>
              <span class="subtext">Parcourir par artiste</span>
            </div>
          </button>
          <button @click="showArtistsMenu = false" class="menu-close">✕ Fermer</button>
        </div>
      </div>
    </Teleport>

    <AboutModal v-model="isAboutModalOpen" />
  </div>
</template>

<style scoped>
/* ============================================
   LAYOUT GLOBAL
   ============================================ */
.dashboard-layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  height: 100vh;
  width: 100vw;
  overflow-x: hidden;
  overflow-y: hidden;
  transition: grid-template-columns 0.3s ease;
  position: relative;
  background: var(--body-bg);
}

.dashboard-layout.sidebar-collapsed {
  grid-template-columns: 80px 1fr;
}

@media (max-width: 920px) {
  .dashboard-layout {
    grid-template-columns: 1fr !important;
  }
}

/* ============================================
   SIDEBAR (alignée sur AppSidebar.vue de Finance)
   ============================================ */
.sidebar {
  position: relative;
  height: 100vh;
  padding: 1.2rem 1rem;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  background: var(--sidebar-bg);
  border-right: 1px solid rgba(var(--tint-rgb), 0.05);
  z-index: 200;
}

.sidebar:not(.is-resizing) {
  transition: width 160ms ease, min-width 160ms ease;
}

.sidebar.is-collapsed {
  padding-inline: 0.6rem;
}

.sidebar-scroll {
  overflow-y: auto;
  overflow-x: hidden;
  flex: 1;
}

.sidebar.is-collapsed .sidebar-scroll {
  overflow: visible;
}

.mobile-menu-toggle,
.mobile-close-btn,
.mobile-backdrop {
  display: none;
}

@media (max-width: 920px) {
  .mobile-menu-toggle {
    display: grid;
    place-items: center;
    position: fixed;
    top: 14px;
    left: 14px;
    z-index: 210;
    width: 42px;
    height: 42px;
    border: 1px solid rgba(var(--tint-rgb), 0.1);
    border-radius: 12px;
    background: var(--sidebar-bg);
    color: var(--text);
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.28);
    cursor: pointer;
  }

  .mobile-menu-toggle svg {
    width: 20px;
    height: 20px;
  }

  .mobile-backdrop {
    display: block;
    position: fixed;
    inset: 0;
    z-index: 199;
    background: rgba(0, 0, 0, 0.5);
  }

  .mobile-close-btn {
    display: grid;
    place-items: center;
    margin-left: auto;
    flex-shrink: 0;
    width: 32px;
    height: 32px;
    border: none;
    border-radius: 10px;
    background: rgba(var(--tint-rgb), 0.06);
    color: var(--text);
    cursor: pointer;
  }

  .mobile-close-btn svg {
    width: 16px;
    height: 16px;
  }

  .sidebar.is-mobile {
    position: fixed;
    top: 0;
    left: 0;
    z-index: 205;
    height: 100dvh;
    max-width: 85vw;
    transform: translateX(-100%);
    box-shadow: 12px 0 40px rgba(0, 0, 0, 0.35);
    transition: transform 220ms ease;
  }

  .sidebar.is-mobile.is-mobile-open {
    transform: translateX(0);
  }
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  margin-bottom: 1.6rem;
}

.is-collapsed .brand {
  justify-content: center;
}

.brand-mark {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  box-shadow: 0 4px 14px rgba(59, 130, 246, 0.35);
  flex-shrink: 0;
  font-size: 18px;
  text-decoration: none;
}

.brand strong {
  display: block;
  color: var(--text);
  font-size: 0.98rem;
}

.brand p {
  margin-top: 0.2rem;
  color: rgba(var(--tint-rgb), 0.56);
  font-size: 0.84rem;
}

.sidebar-nav {
  display: grid;
  gap: 0.3rem;
}

.nav-section-body {
  display: contents;
}

.nav-section-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.4rem;
  width: 100%;
  margin: 0.5rem 0 0.05rem;
  padding: 0.15rem 0.85rem;
  background: none;
  border: none;
  cursor: pointer;
  border-radius: 8px;
}

.nav-section-toggle:hover {
  background: rgba(var(--tint-rgb), 0.05);
}

.nav-section-toggle .nav-group-chevron {
  color: rgba(var(--tint-rgb), 0.35);
}

.nav-section-label {
  margin: 0;
  padding: 0;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-dim);
}

/* Le libellé "Général" est un <p> autonome (pas de bouton repliable comme
   Disques/Jeux) — sans cette règle il n'a pas la même marge que
   .nav-section-toggle et se retrouve visuellement plus proche du groupe
   précédent que les deux autres en-têtes de section. */
p.nav-section-label {
  margin: 0.5rem 0 0.05rem;
  padding: 0.15rem 0.85rem;
}

.nav-section-divider {
  margin: 0.5rem 0.6rem;
  border-top: 1px solid var(--line-soft);
}

.nav-group {
  display: contents;
}

.nav-group.nav-group-collapsed {
  display: grid;
  gap: 0.35rem;
  padding: 0.35rem 0.2rem;
  border-radius: 14px;
  background: rgba(var(--tint-rgb), 0.05);
  border: 1px solid rgba(var(--tint-rgb), 0.08);
}

.nav-group-toggle-label {
  flex: 1;
}

.nav-group-chevron {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  color: rgba(var(--tint-rgb), 0.4);
  transition: transform 160ms ease;
}

.nav-group-chevron.is-open {
  transform: rotate(90deg);
}

.nav-children-rail {
  position: relative;
  display: grid;
  gap: 0.45rem;
}

.nav-children-rail.has-rail {
  padding-left: 2.5rem;
  margin-left: 0.35rem;
  margin-top: 0.3rem;
}

.nav-children-rail.has-rail::before {
  content: '';
  position: absolute;
  top: 0.2rem;
  bottom: 0.2rem;
  left: 0;
  width: 1px;
  background: rgba(var(--tint-rgb), 0.1);
}

.nav-item-child {
  gap: 0.5rem;
  color: rgba(var(--tint-rgb), 0.5);
  font-size: 0.72rem;
  font-weight: 500;
}

.nav-item-child .nav-icon-chip {
  width: 22px;
  height: 22px;
}

.nav-item-child .nav-emoji {
  font-size: 11px;
}

.nav-item {
  position: relative;
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.6rem;
  border: none;
  border-radius: 12px;
  padding: 0.6rem 0.85rem;
  background: transparent;
  color: rgba(var(--tint-rgb), 0.68);
  text-align: left;
  text-decoration: none;
  font-size: 0.86rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 160ms ease, color 160ms ease;
}

/* Décale les items enfants (sous Disques/Jeux/Général) vers la droite pour
   marquer clairement leur rattachement au libellé de section au-dessus. */
.nav-subitem {
  padding-left: 1.7rem;
}

/* En mode réduit (icônes seules), la barre latérale est trop étroite pour
   ce décalage — sans ce reset les icônes se retrouveraient désaxées. */
.sidebar.is-collapsed .nav-subitem {
  padding-left: 0.85rem;
}

.nav-tooltip {
  position: absolute;
  left: calc(100% + 10px);
  top: 50%;
  transform: translateY(-50%) translateX(-4px);
  z-index: 30;
  padding: 0.4rem 0.7rem;
  border-radius: 8px;
  background: var(--modal-bg);
  color: var(--text);
  font-size: 0.8rem;
  font-weight: 600;
  white-space: nowrap;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.28);
  opacity: 0;
  pointer-events: none;
  transition: opacity 140ms ease, transform 140ms ease;
}

.nav-item:hover .nav-tooltip,
.nav-item:focus-visible .nav-tooltip {
  opacity: 1;
  transform: translateY(-50%) translateX(0);
}

.nav-icon-chip {
  flex-shrink: 0;
  width: 26px;
  height: 26px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.3),
    inset 0 -6px 10px rgba(0, 0, 0, 0.18),
    0 2px 5px rgba(0, 0, 0, 0.28);
  transition: transform 160ms ease;
}

.nav-item:hover .nav-icon-chip {
  transform: scale(1.06) translateY(-1px);
}

.nav-emoji {
  font-size: 13px;
  line-height: 1;
  filter: drop-shadow(0 1px 1px rgba(0, 0, 0, 0.25));
}

.nav-item:hover {
  background: rgba(var(--tint-rgb), 0.06);
  color: var(--text);
}

.nav-item.active {
  background: linear-gradient(90deg, rgba(59, 130, 246, 0.18), rgba(59, 130, 246, 0.06));
  color: var(--text);
  box-shadow: inset 0 0 0 1px rgba(59, 130, 246, 0.3);
}

.sidebar-footer {
  padding-top: 1rem;
  display: grid;
  gap: 0.6rem;
}

.profile-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.85rem 0.9rem;
  border-radius: 14px;
  background: rgba(var(--tint-rgb), 0.04);
  border: 1px solid rgba(var(--tint-rgb), 0.05);
}

.profile-card-collapsed {
  justify-content: center;
  padding: 0.6rem;
}

.avatar {
  position: relative;
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  background: rgba(var(--tint-rgb), 0.1);
  color: var(--text);
  font-weight: 700;
  font-size: 0.88rem;
  flex-shrink: 0;
}

.profile-info {
  min-width: 0;
  flex: 1;
}

.profile-card strong {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text);
  font-size: 0.92rem;
}

.profile-card p {
  margin-top: 0.15rem;
  color: rgba(var(--tint-rgb), 0.54);
  font-size: 0.78rem;
}

.profile-actions {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  flex-shrink: 0;
}

.icon-action-btn {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border: 1px solid rgba(var(--tint-rgb), 0.08);
  border-radius: 9px;
  display: grid;
  place-items: center;
  background: rgba(var(--tint-rgb), 0.03);
  color: rgba(var(--tint-rgb), 0.6);
  cursor: pointer;
  text-decoration: none;
  transition: background 140ms ease, color 140ms ease, transform 140ms ease;
}

.icon-action-btn:hover {
  background: rgba(var(--tint-rgb), 0.09);
  color: var(--text);
  transform: translateY(-1px);
}

.icon-action-btn-danger:hover {
  background: rgba(220, 38, 38, 0.14);
  color: var(--negative-text);
}

.icon-action-btn svg {
  width: 14px;
  height: 14px;
}

.collapse-toggle {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  width: 100%;
  height: 34px;
  border: 1px solid rgba(var(--tint-rgb), 0.07);
  border-radius: 10px;
  background: rgba(var(--tint-rgb), 0.03);
  color: rgba(var(--tint-rgb), 0.6);
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 140ms ease, color 140ms ease;
}

.collapse-toggle:hover {
  background: rgba(var(--tint-rgb), 0.08);
  color: var(--text);
}

.collapse-toggle svg {
  width: 14px;
  height: 14px;
  transition: transform 160ms ease;
  flex-shrink: 0;
}

.collapse-toggle svg.is-flipped {
  transform: rotate(180deg);
}

.resize-handle {
  position: absolute;
  top: 0;
  right: -3px;
  width: 6px;
  height: 100%;
  cursor: col-resize;
  z-index: 5;
}

.resize-handle:hover,
.is-resizing .resize-handle {
  background: rgba(59, 130, 246, 0.35);
}

/* Animation du sous-menu Artistes */
.submenu-enter-active,
.submenu-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.submenu-enter-from,
.submenu-leave-to {
  opacity: 0;
  max-height: 0;
  transform: translateY(-10px);
}

.submenu-enter-to,
.submenu-leave-from {
  opacity: 1;
  max-height: 200px;
  transform: translateY(0);
}

/* ============================================
   CONTENT AREA
   ============================================ */
.content-area {
  background: transparent;
  padding: 20px;
  overflow-y: auto;
  overflow-x: hidden;
  height: 100vh;
  width: 100%;
  box-sizing: border-box;
}

@media (max-width: 767px) {
  .content-area {
    padding: 12px;
    width: 100vw;
    max-width: 100vw;
  }
}

/* ============================================
   PAGE HEADER
   ============================================ */
.page-header {
  margin-bottom: 20px;
}

.header-wrapper {
  padding: 0;
  margin: 0;
}

.header-top-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 5px;
  flex-wrap: wrap;
  gap: 16px;
}

.header-title-container {
  display: flex;
  align-items: center;
}

.dashboard-mode-toggle {
  display: inline-flex;
  gap: 0.3rem;
  padding: 0.3rem;
  margin-top: 0.9rem;
  background: rgba(var(--tint-rgb), 0.05);
  border: 1px solid var(--line-soft);
  border-radius: 999px;
}

.dashboard-mode-tab {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 1.1rem;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: var(--text-soft);
  font-weight: 600;
  font-size: 0.9em;
  cursor: pointer;
  transition: all 0.2s ease;
}

.dashboard-mode-tab:hover {
  color: var(--text);
}

.dashboard-mode-tab.active {
  background: var(--accent);
  color: white;
}

.title-icon {
  font-size: 2em;
  margin-right: 8px;
}

.page-header h1 {
  color: var(--text);
  font-size: 2em;
  margin: 0;
  font-weight: bold;
}

.toolbar {
  margin: 8px 0 0 0;
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.logout-header-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  padding: 10px 16px;
  font-weight: 600;
  font-size: 1em;
  box-shadow: 0 2px 8px rgba(231, 76, 60, 0.3);
  transition: all 0.3s ease;
  min-height: 44px;
}

.logout-header-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(231, 76, 60, 0.4);
}

.logout-header-button .icon {
  font-size: 1.2em;
}

@media (max-width: 480px) {
  .logout-header-button .text {
    display: none;
  }
  
  .logout-header-button {
    padding: 10px 12px;
  }
  
  .page-header h1 {
    font-size: 1.5em;
  }
}

/* ============================================
   ERROR MESSAGE
   ============================================ */
.error-message {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.28);
  border-radius: 14px;
  color: var(--negative-text);
  padding: 12px 16px;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 10px;
}

.retry-button {
  background: rgba(239, 68, 68, 0.85);
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  padding: 8px 12px;
  font-size: 0.9em;
}

.retry-button:hover {
  background: rgba(220, 38, 38, 0.95);
}

/* ============================================
   LOADING OVERLAY
   ============================================ */
.loading-overlay {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--bg);
  backdrop-filter: blur(8px);
  z-index: 999;
  text-align: center;
  padding-left: 280px;
  transition: padding-left 0.3s ease;
}

.dashboard-layout.sidebar-collapsed .loading-overlay {
  padding-left: 80px;
}

@media (max-width: 767px) {
  .loading-overlay {
    padding-left: 0 !important;
  }
}

.spinner {
  font-size: 4em;
  margin-bottom: 20px;
  animation: spin 2s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* ============================================
   DASHBOARD WIDGETS
   ============================================ */
.dashboard-widgets {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 20px;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

@media (max-width: 767px) {
  .dashboard-widgets {
    gap: 12px;
    margin-bottom: 16px;
  }
}

.thematics-row {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

@media (max-width: 1100px) {
  .thematics-row {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 767px) {
  .thematics-row {
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }
}

@media (max-width: 340px) {
  .thematics-row {
    grid-template-columns: 1fr;
    gap: 6px;
  }
}

.charts-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

@media (max-width: 1400px) {
  .charts-row {
    grid-template-columns: repeat(2, 1fr);
    gap: 14px;
  }
}

@media (max-width: 992px) {
  .charts-row {
    grid-template-columns: 1fr;
    gap: 12px;
  }
}

@media (max-width: 767px) {
  .charts-row {
    gap: 10px;
  }
}

/* ============================================
   🆕 MODAL D'AVERTISSEMENT D'INACTIVITÉ
   ============================================ */
.idle-warning-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(8px);
  z-index: 99999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  animation: overlayFadeIn 0.3s ease-out;
}

@keyframes overlayFadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.idle-warning-modal {
  background: white;
  border-radius: 20px;
  padding: 40px 32px;
  max-width: 420px;
  width: 100%;
  text-align: center;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
  animation: modalBounceIn 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

@keyframes modalBounceIn {
  0% {
    opacity: 0;
    transform: scale(0.3) translateY(-50px);
  }
  50% {
    transform: scale(1.05);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.warning-icon {
  font-size: 4rem;
  margin-bottom: 20px;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
}

.idle-warning-modal h3 {
  font-size: 1.75rem;
  color: #dc3545;
  margin: 0 0 16px 0;
  font-weight: 700;
}

.warning-text {
  color: #6b7280;
  margin: 0 0 12px 0;
  font-size: 1rem;
}

.countdown {
  font-size: 4rem;
  font-weight: 700;
  color: #dc3545;
  margin: 16px 0 8px 0;
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.countdown-label {
  color: #9ca3af;
  font-size: 0.875rem;
  margin: 0 0 20px 0;
}

.hint {
  color: #6b7280;
  font-size: 0.875rem;
  font-style: italic;
  margin: 20px 0 24px 0;
}

.stay-button {
  width: 100%;
  padding: 16px 32px;
  background: linear-gradient(135deg, #d87d3a 0%, #e89b5a 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 1.125rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 16px rgba(216, 125, 58, 0.4);
}

.stay-button:hover {
  background: linear-gradient(135deg, #c66a2a 0%, #d87d3a 100%);
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(216, 125, 58, 0.5);
}

.stay-button:active {
  transform: translateY(0);
}

@media (max-width: 480px) {
  .idle-warning-modal {
    padding: 32px 24px;
  }

  .warning-icon {
    font-size: 3rem;
  }

  .idle-warning-modal h3 {
    font-size: 1.5rem;
  }

  .countdown {
    font-size: 3rem;
  }

  .stay-button {
    font-size: 1rem;
    padding: 14px 28px;
  }
}

/* ============================================
   TOOLTIPS
   ============================================ */
.pie-tooltip {
  pointer-events: none;
  position: fixed;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.85);
  color: white;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 0.9em;
  max-width: 250px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  animation: tooltipFadeIn 0.2s ease-out;
}

@keyframes tooltipFadeIn {
  from {
    opacity: 0;
    transform: translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.tooltip-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.tooltip-count {
  font-size: 0.95em;
  opacity: 0.9;
}

.tooltip-percentage {
  font-size: 0.95em;
  color: #a0d8f1;
}

.tooltip-arrow {
  position: absolute;
  bottom: -6px;
  left: 20px;
  width: 0;
  height: 0;
  border-left: 6px solid transparent;
  border-right: 6px solid transparent;
  border-top: 6px solid rgba(0, 0, 0, 0.85);
}

/* ============================================
   MENU CONTEXTUEL ARTISTES (MOBILE)
   ============================================ */
.artists-context-menu-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  animation: overlayFadeIn 0.2s ease-out;
}

.artists-context-menu {
  background: white;
  border-radius: 16px;
  padding: 24px;
  max-width: 340px;
  width: 100%;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.3);
  animation: menuSlideUp 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes menuSlideUp {
  from {
    opacity: 0;
    transform: translateY(30px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.artists-context-menu h3 {
  margin: 0 0 8px 0;
  color: #1a1a2e;
  font-size: 1.5em;
  text-align: center;
  font-weight: 700;
}

.menu-subtitle {
  text-align: center;
  color: #666;
  font-size: 0.9em;
  margin: 0 0 20px 0;
}

.menu-option {
  display: flex;
  align-items: center;
  gap: 14px;
  width: 100%;
  padding: 16px 18px;
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  border: none;
  border-radius: 12px;
  color: white;
  font-weight: 600;
  cursor: pointer;
  margin-bottom: 12px;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow: 0 4px 12px rgba(240, 147, 251, 0.3);
  text-align: left;
}

.menu-option:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 20px rgba(240, 147, 251, 0.5);
}

.menu-option:active {
  transform: translateY(-1px);
}

.menu-option .icon {
  font-size: 1.6em;
  flex-shrink: 0;
}

.option-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
}

.option-content .text {
  font-size: 1em;
  font-weight: 600;
  line-height: 1.2;
}

.option-content .subtext {
  font-size: 0.8em;
  opacity: 0.9;
  font-weight: 400;
}

.menu-close {
  display: block;
  width: 100%;
  padding: 12px;
  background: rgba(0, 0, 0, 0.05);
  border: none;
  border-radius: 10px;
  color: #666;
  cursor: pointer;
  margin-top: 8px;
  transition: all 0.2s;
  font-weight: 500;
  font-size: 0.95em;
}

.menu-close:hover {
  background: rgba(0, 0, 0, 0.1);
  color: #333;
}

.menu-close:active {
  transform: scale(0.98);
}

/* Animation au clic sur le widget */
@media (max-width: 767px) {
  .dashboard-widgets :deep(.stats-widget[data-type="artists"]) {
    position: relative;
  }

  .dashboard-widgets :deep(.stats-widget[data-type="artists"]:active) {
    transform: scale(0.97);
  }
}
</style>
