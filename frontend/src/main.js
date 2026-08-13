// records-manager/frontend/src/main.js

import './assets/base.css';
import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';

// 1. Créer l'application
const app = createApp(App);

// 2. Créer et installer Pinia
const pinia = createPinia();
app.use(pinia);

// 3. Installer le routeur
app.use(router);

// 4. Monter l'application → ⚠️ indispensable avant d'utiliser les stores
app.mount('#app');

// ✅ 5. Initialisation DIFFÉRÉE du store (après montage)
import { nextTick } from 'vue';
nextTick(async () => {
  // ✅ CORRECTION : Utiliser import() dynamique au lieu de require()
  const { useAuthStore } = await import('@/stores/auth');
  const authStore = useAuthStore();
  authStore.initialize(); // ← safe maintenant
});
