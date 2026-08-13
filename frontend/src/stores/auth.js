// src/stores/auth.js
import { defineStore } from 'pinia';
import router from '@/router';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('user_token') || null,
    user: JSON.parse(localStorage.getItem('user_data')) || null,
    isLoggedIn: !!localStorage.getItem('user_token'),
  }),

  getters: {
    isAuthenticated: (state) => !!state.token,
  },

  actions: {
    // ✅ Plus besoin de setToken/setAuthHeader — useApi.js gère tout seul

    async login(credentials) {
      try {
        // On utilise useApi (pas axios)
        const { post } = await import('@/composables/useApi').then(m => m.useApi());
        const response = await post('/login', credentials);

        const { token, user } = response;

        this.token = token;
        this.user = user;
        this.isLoggedIn = true;

        localStorage.setItem('user_token', token);
        localStorage.setItem('user_data', JSON.stringify(user));

        router.push({ name: 'Dashboard' });
      } catch (error) {
        this.logout();
        throw new Error(error.message || 'Identifiants invalides');
      }
    },

    logout() {
      this.token = null;
      this.user = null;
      this.isLoggedIn = false;

      localStorage.removeItem('user_token');
      localStorage.removeItem('user_data');

      router.push({ name: 'Login' });
    },

    // ✅ initialize() devient optionnel — mais utile pour restauration
    initialize() {
      // Rien à faire : useApi.js lit directement this.token
      // Le token est déjà dans le state grâce à localStorage
    },
  },
});