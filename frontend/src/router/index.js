// records-manager/frontend/src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'
import Login from '@/views/Login.vue'
import Dashboard from '@/views/Dashboard.vue'
import GenresView from '@/views/GenresView.vue'
import ArtistsView from '@/views/ArtistsView.vue'
import FormatsView from '@/views/FormatsView.vue'
import CountriesView from '@/views/CountriesView.vue'
import LabelsView from '@/views/LabelsView.vue'
import DiscsView from '@/views/DiscsView.vue'
import VinylsByArtist from '@/views/VinylsByArtist.vue'
import GamesView from '@/views/GamesView.vue'
import PlatformsView from '@/views/PlatformsView.vue'
import GameGenresView from '@/views/GameGenresView.vue'
import PublishersView from '@/views/PublishersView.vue'
import ChangePassword from '@/views/ChangePassword.vue'
import ForgotPassword from '@/views/ForgotPassword.vue'
import ResetPassword from '@/views/ResetPassword.vue' // 🆕 AJOUTÉ
import PreferencesView from '@/views/PreferencesView.vue'
import ProfileView from '@/views/ProfileView.vue'
import ReportsView from '@/views/ReportsView.vue'

import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/',
    redirect: '/login'
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresAuth: false }
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: ForgotPassword,
    meta: { requiresAuth: false }
  },
  {
    path: '/reset-password', // 🆕 AJOUTÉ
    name: 'ResetPassword',
    component: ResetPassword,
    meta: { requiresAuth: false }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard,
    meta: { requiresAuth: true },
    children: [
      // Routes de paramétrage
      {
        path: 'settings/formats',
        name: 'Formats',
        component: FormatsView,
        meta: { requiresAuth: true }
      },
      {
        path: 'settings/artists',
        name: 'Artists',
        component: ArtistsView,
        meta: { requiresAuth: true }
      },
      {
        path: 'settings/genres',
        name: 'Genres',
        component: GenresView,
        meta: { requiresAuth: true }
      },
      {
        path: 'settings/countries',
        name: 'Countries',
        component: CountriesView,
        meta: { requiresAuth: true }
      },
      {
        path: 'settings/labels',
        name: 'Labels',
        component: LabelsView,
        meta: { requiresAuth: true }
      },
      {
        path: 'settings/preferences',
        name: 'Preferences',
        component: PreferencesView,
        meta: { requiresAuth: true }
      },
      {
        path: 'settings/profile',
        name: 'Profile',
        component: ProfileView,
        meta: { requiresAuth: true }
      },
      {
        path: 'settings/reports',
        name: 'Reports',
        component: ReportsView,
        meta: { requiresAuth: true }
      },
      // Routes des disques
      {
        path: 'vinyls',
        name: 'Discs',
        component: DiscsView,
        meta: { requiresAuth: true }
      },
      {
        path: 'vinyls/by-artist',
        name: 'VinylsByArtist',
        component: VinylsByArtist,
        meta: { requiresAuth: true }
      },
      // Routes des jeux vidéo
      {
        path: 'games',
        name: 'Games',
        component: GamesView,
        meta: { requiresAuth: true }
      },
      {
        path: 'settings/platforms',
        name: 'Platforms',
        component: PlatformsView,
        meta: { requiresAuth: true }
      },
      {
        path: 'settings/game-genres',
        name: 'GameGenres',
        component: GameGenresView,
        meta: { requiresAuth: true }
      },
      {
        path: 'settings/publishers',
        name: 'Publishers',
        component: PublishersView,
        meta: { requiresAuth: true }
      },
      // Route changement de mot de passe (DANS le Dashboard)
      {
        path: 'change-password',
        name: 'ChangePassword',
        component: ChangePassword,
        meta: { requiresAuth: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guard pour vérifier l'authentification
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  // Si la route nécessite l'authentification
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    // Rediriger vers login
    next({ name: 'Login' })
  } 
  // Si l'utilisateur est connecté et essaie d'accéder au login ou forgot-password
  else if ((to.name === 'Login' || to.name === 'ForgotPassword') && authStore.isAuthenticated) {
    // Rediriger vers dashboard
    next({ name: 'Dashboard' })
  } 
  else {
    // Autoriser la navigation
    next()
  }
})

export default router
