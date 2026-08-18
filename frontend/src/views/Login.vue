<script setup>
// records-manager/frontend/src/views/Login.vue
import { ref } from 'vue';
import { useAuthStore } from '@/stores/auth';
import router from '@/router';
import charlieLogo from '@/assets/charlie-digital-logo.png';
import { APP_VERSION } from '@/version';

// --------------------------------------------------------
// LOGIQUE D'AUTHENTIFICATION
// --------------------------------------------------------
const authStore = useAuthStore();

// Redirection si déjà authentifié
if (authStore.isAuthenticated) {
  router.push({ name: 'Dashboard' });
}

// Modèles de données
const email = ref('');
const password = ref('');
const rememberMe = ref(false);
const errorMessage = ref(null);
const showPassword = ref(false);

const handleLogin = async () => {
  errorMessage.value = null;
  try {
    await authStore.login({
      email: email.value,
      password: password.value,
      remember: rememberMe.value
    });
  } catch (error) {
    console.error("Échec de la connexion:", error.message);
    errorMessage.value = "Identifiants invalides ou erreur de connexion au serveur.";
  }
};
</script>

<template>
  <main class="login-shell">
    <div class="login-glow" aria-hidden="true"></div>

    <section class="login-card">
      <img :src="charlieLogo" alt="Charlie digital" class="login-parent-logo" />
      <p class="login-version">v{{ APP_VERSION }}</p>

      <div class="login-brand">
        <div class="login-logo-wrap">
          <img src="/mediavault-logo.png" alt="MediaVault" class="login-logo-image" />
        </div>
      </div>

      <div class="login-heading">
        <p class="eyebrow">Espace privé</p>
        <h1>MediaVault</h1>
        <p class="login-subtitle">Gérez votre collection de disques avec style 🎶</p>
      </div>

      <div v-if="errorMessage" class="form-error login-error-banner">
        <span class="error-icon">⚠️</span>
        {{ errorMessage }}
      </div>

      <form class="login-form" @submit.prevent="handleLogin">
        <label class="form-field">
          <span>Utilisateur</span>
          <input
            v-model="email"
            type="email"
            autocomplete="username"
            placeholder="votre@email.com"
            required
          />
        </label>

        <label class="form-field">
          <span>Mot de passe</span>
          <div class="password-field">
            <input
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              autocomplete="current-password"
              placeholder="••••••••"
              required
            />
            <button
              type="button"
              class="password-toggle"
              :aria-label="showPassword ? 'Masquer le mot de passe' : 'Afficher le mot de passe'"
              :title="showPassword ? 'Masquer le mot de passe' : 'Afficher le mot de passe'"
              @click="showPassword = !showPassword"
            >
              <svg v-if="showPassword" viewBox="0 0 24 24" aria-hidden="true">
                <path
                  d="M3 3l18 18M10.58 10.58a2 2 0 0 0 2.83 2.83M9.88 5.09A9.77 9.77 0 0 1 12 5c5 0 9 4 10 7-.32.99-1.06 2.35-2.17 3.6M6.15 6.44C3.9 7.9 2.32 9.9 2 12c1 3 5 7 10 7 1.2 0 2.35-.23 3.4-.64"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.8"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
              <svg v-else viewBox="0 0 24 24" aria-hidden="true">
                <path
                  d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.8"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
                <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" stroke-width="1.8" />
              </svg>
            </button>
          </div>
        </label>

        <label class="form-field form-field-checkbox">
          <input type="checkbox" v-model="rememberMe" />
          <span>Se souvenir de moi</span>
        </label>

        <button class="primary-btn login-submit" type="submit" :disabled="authStore.loading">
          {{ authStore.loading ? 'Connexion en cours...' : 'Connexion' }}
        </button>

        <router-link to="/forgot-password" class="login-link-btn">
          Mot de passe oublié ?
        </router-link>
      </form>
    </section>
  </main>
</template>

<style scoped>
.login-shell {
  position: relative;
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 1.5rem;
  --text: #f3f5fa;
  --text-soft: #c9d2e6;
  --text-dim: #93a1c2;
  --tint-rgb: 255, 255, 255;
  --line: rgba(255, 255, 255, 0.18);
  background: var(--hero-gradient);
  overflow: hidden;
}

.login-glow {
  position: absolute;
  inset: -20%;
  background:
    radial-gradient(circle at 20% 20%, rgba(59, 130, 246, 0.22), transparent 45%),
    radial-gradient(circle at 80% 75%, rgba(242, 168, 120, 0.22), transparent 40%);
  pointer-events: none;
}

.login-card {
  position: relative;
  width: min(100%, 420px);
  padding: 2rem 1.9rem;
  border-radius: var(--radius-xl, 28px);
  background: rgba(22, 29, 46, 0.62);
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow: 0 30px 70px rgba(5, 8, 16, 0.45);
  backdrop-filter: blur(22px);
  -webkit-backdrop-filter: blur(22px);
}

.login-version {
  position: absolute;
  top: 1.4rem;
  right: 1.9rem;
  margin: 0;
  color: var(--text-dim, #8a939d);
  font-size: 0.76rem;
  letter-spacing: 0.02em;
}

.login-brand {
  display: flex;
  justify-content: center;
  margin-bottom: 0.6rem;
}

.login-parent-logo {
  position: absolute;
  top: 1.4rem;
  left: 1.9rem;
  width: 36px;
  height: auto;
  opacity: 0.85;
  filter: drop-shadow(0 4px 10px rgba(0, 0, 0, 0.3));
}

.login-logo-wrap {
  position: relative;
  width: 100%;
  max-width: 210px;
  margin: 0 auto;
}

.login-logo-wrap::before {
  content: '';
  position: absolute;
  inset: -12%;
  background: radial-gradient(circle, var(--accent, #3b82f6) 0%, var(--accent-soft, #60a5fa) 45%, transparent 72%);
  opacity: 0.32;
  filter: blur(18px);
  pointer-events: none;
}

.login-logo-image {
  position: relative;
  display: block;
  width: 100%;
  height: auto;
  filter: drop-shadow(0 10px 24px rgba(0, 0, 0, 0.35));
  -webkit-mask-image: radial-gradient(ellipse 62% 62% at center, black 45%, transparent 78%);
  mask-image: radial-gradient(ellipse 62% 62% at center, black 45%, transparent 78%);
}

.login-heading {
  margin-bottom: 1.6rem;
  text-align: center;
}

.eyebrow {
  color: var(--text-dim, #8a939d);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 0.74rem;
}

.login-heading h1 {
  margin: 0.3rem 0 0;
  color: var(--text, #eef1f3);
  font-size: 1.6rem;
  font-weight: 800;
  line-height: 1.2;
}

.login-subtitle {
  margin: 0.6rem 0 0;
  color: var(--text-soft, #b3bbc4);
  font-size: 0.9rem;
  line-height: 1.5;
}

.login-error-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: 14px;
  background: rgba(239, 68, 68, 0.14);
  border: 1px solid rgba(239, 68, 68, 0.28);
  margin-bottom: 1rem;
  font-size: 0.875rem;
}

.error-icon {
  font-size: 1.1em;
}

.login-form {
  display: grid;
  gap: 1rem;
}

.password-field {
  position: relative;
}

.password-field input {
  padding-right: 44px;
}

.password-toggle {
  position: absolute;
  top: 50%;
  right: 6px;
  transform: translateY(-50%);
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--text-dim, #8a939d);
  cursor: pointer;
  transition: background 140ms ease, color 140ms ease;
}

.password-toggle:hover {
  background: rgba(var(--tint-rgb), 0.06);
  color: var(--text, #eef1f3);
}

.password-toggle svg {
  width: 18px;
  height: 18px;
}

.login-submit {
  width: 100%;
  height: 46px;
  margin-top: 0.3rem;
  font-size: 0.95rem;
}

.login-link-btn {
  display: block;
  width: 100%;
  border: none;
  background: transparent;
  color: var(--text-dim, #8a939d);
  font-size: 0.84rem;
  font-weight: 600;
  cursor: pointer;
  padding: 0.3rem;
  text-align: center;
  transition: color 140ms ease;
}

.login-link-btn:hover {
  color: var(--text-soft, #b3bbc4);
  text-decoration: underline;
  background: transparent;
}

@media (max-width: 480px) {
  .login-card {
    padding: 1.6rem 1.3rem;
  }
}
</style>
