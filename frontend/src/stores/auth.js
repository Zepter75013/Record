// src/stores/auth.js
import { defineStore } from 'pinia';
import router from '@/router';

const TOKEN_KEY = 'user_token';
const USER_KEY = 'user_data';

// Le token est dans sessionStorage par défaut (effacé à la fermeture du
// navigateur/onglet — l'app "oublie" la connexion) ou dans localStorage si
// l'utilisateur a coché "Se souvenir de moi" à la connexion (survit à la
// fermeture). On lit les deux ici pour retrouver une session existante quel
// que soit l'endroit où elle a été écrite.
function readToken() {
  return sessionStorage.getItem(TOKEN_KEY) || localStorage.getItem(TOKEN_KEY) || null;
}
function readUser() {
  const raw = sessionStorage.getItem(USER_KEY) || localStorage.getItem(USER_KEY);
  try {
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: readToken(),
    user: readUser(),
    isLoggedIn: !!readToken(),
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

        const storage = credentials.remember ? localStorage : sessionStorage;
        const other = credentials.remember ? sessionStorage : localStorage;
        other.removeItem(TOKEN_KEY);
        other.removeItem(USER_KEY);
        storage.setItem(TOKEN_KEY, token);
        storage.setItem(USER_KEY, JSON.stringify(user));

        router.push({ name: 'Dashboard' });
      } catch (error) {
        this.logout();
        throw new Error(error.message || 'Identifiants invalides');
      }
    },

    // Met à jour la photo de profil dans le state ET dans le storage où vit
    // déjà la session (localStorage si "se souvenir de moi", sinon
    // sessionStorage) — sans ça la photo redeviendrait la lettre par défaut
    // au prochain rechargement de page.
    setAvatarPath(avatarPath) {
      this.user = { ...(this.user || {}), avatar_path: avatarPath };
      const storage = localStorage.getItem(TOKEN_KEY) ? localStorage : sessionStorage;
      storage.setItem(USER_KEY, JSON.stringify(this.user));
    },

    logout() {
      this.token = null;
      this.user = null;
      this.isLoggedIn = false;

      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
      sessionStorage.removeItem(TOKEN_KEY);
      sessionStorage.removeItem(USER_KEY);

      router.push({ name: 'Login' });
    },

    // ✅ initialize() devient optionnel — mais utile pour restauration
    initialize() {
      // Rien à faire : useApi.js lit directement this.token
      // Le token est déjà dans le state grâce à session/localStorage
    },
  },
});
