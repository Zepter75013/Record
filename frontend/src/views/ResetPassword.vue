<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApi } from '@/composables/useApi'
import { usePasswordValidation } from '@/composables/usePasswordValidation'
import PasswordStrengthIndicator from '@/components/PasswordStrengthIndicator.vue'

const route = useRoute()
const router = useRouter()
const { post } = useApi()
const passwordValidation = usePasswordValidation()

const token = ref('')
const confirmPassword = ref('')
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const loading = ref(false)
const success = ref(false)
const error = ref('')

const canSubmit = computed(() => {
  return passwordValidation.password.value &&
         confirmPassword.value &&
         passwordValidation.isValid.value &&
         passwordValidation.password.value === confirmPassword.value
})

onMounted(() => {
  token.value = route.query.token || ''
  if (!token.value) {
    error.value = 'Token de réinitialisation manquant'
  }
})

const handleSubmit = async () => {
  if (!canSubmit.value || loading.value) return

  error.value = ''
  loading.value = true

  try {
    await post('/password-reset/reset', {
      token: token.value,
      new_password: passwordValidation.password.value
    })

    success.value = true

    setTimeout(() => {
      router.push('/login')
    }, 3000)

  } catch (err) {
    error.value = err.message || 'Une erreur est survenue'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="login-shell">
    <div class="login-glow" aria-hidden="true"></div>

    <section class="login-card">
      <div class="login-brand">
        <div class="login-logo-wrap">
          <img src="/mediavault-logo.png" alt="MediaVault" class="login-logo-image" />
        </div>
      </div>

      <template v-if="!success">
        <div class="login-heading">
          <p class="eyebrow">Sécurité</p>
          <h1>Nouveau mot de passe</h1>
          <p class="login-subtitle">Choisissez un nouveau mot de passe sécurisé pour votre compte.</p>
        </div>

        <form class="login-form" @submit.prevent="handleSubmit">
          <label class="form-field">
            <span>Nouveau mot de passe</span>
            <div class="password-field">
              <input
                v-model="passwordValidation.password.value"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="new-password"
                placeholder="••••••••"
                required
                :disabled="loading"
              />
              <button
                type="button"
                class="password-toggle"
                :aria-label="showPassword ? 'Masquer le mot de passe' : 'Afficher le mot de passe'"
                @click="showPassword = !showPassword"
              >
                <svg v-if="showPassword" viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    d="M3 3l18 18M10.58 10.58a2 2 0 0 0 2.83 2.83M9.88 5.09A9.77 9.77 0 0 1 12 5c5 0 9 4 10 7-.32.99-1.06 2.35-2.17 3.6M6.15 6.44C3.9 7.9 2.32 9.9 2 12c1 3 5 7 10 7 1.2 0 2.35-.23 3.4-.64"
                    fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"
                  />
                </svg>
                <svg v-else viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z"
                    fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"
                  />
                  <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" stroke-width="1.8" />
                </svg>
              </button>
            </div>

            <PasswordStrengthIndicator
              :password="passwordValidation.password.value"
              :strength="passwordValidation.strength.value"
              :requirements="passwordValidation.requirements.value"
            />
          </label>

          <label class="form-field">
            <span>Confirmer le mot de passe</span>
            <div class="password-field">
              <input
                v-model="confirmPassword"
                :type="showConfirmPassword ? 'text' : 'password'"
                autocomplete="new-password"
                placeholder="••••••••"
                required
                :disabled="loading"
              />
              <button
                type="button"
                class="password-toggle"
                :aria-label="showConfirmPassword ? 'Masquer le mot de passe' : 'Afficher le mot de passe'"
                @click="showConfirmPassword = !showConfirmPassword"
              >
                <svg v-if="showConfirmPassword" viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    d="M3 3l18 18M10.58 10.58a2 2 0 0 0 2.83 2.83M9.88 5.09A9.77 9.77 0 0 1 12 5c5 0 9 4 10 7-.32.99-1.06 2.35-2.17 3.6M6.15 6.44C3.9 7.9 2.32 9.9 2 12c1 3 5 7 10 7 1.2 0 2.35-.23 3.4-.64"
                    fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"
                  />
                </svg>
                <svg v-else viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z"
                    fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"
                  />
                  <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" stroke-width="1.8" />
                </svg>
              </button>
            </div>
            <p v-if="confirmPassword && confirmPassword !== passwordValidation.password.value" class="form-error">
              Les mots de passe ne correspondent pas
            </p>
          </label>

          <p v-if="error" class="form-error">{{ error }}</p>

          <button class="primary-btn login-submit" type="submit" :disabled="!canSubmit || loading">
            {{ loading ? 'Réinitialisation...' : 'Réinitialiser le mot de passe' }}
          </button>

          <button class="login-link-btn" type="button" @click="router.push('/login')">
            Retour à la connexion
          </button>
        </form>
      </template>

      <template v-else>
        <div class="login-heading">
          <p class="eyebrow">Sécurité</p>
          <h1>Mot de passe réinitialisé !</h1>
        </div>

        <p class="form-success">Votre mot de passe a été modifié avec succès.</p>
        <p class="login-subtitle" style="text-align: center; margin-top: 0.6rem;">
          Redirection vers la page de connexion dans 3 secondes...
        </p>

        <button class="login-link-btn" type="button" @click="router.push('/login')">
          ← Retour à la connexion
        </button>
      </template>
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
  width: min(100%, 460px);
  padding: 2rem 1.9rem;
  border-radius: var(--radius-xl, 28px);
  background: rgba(22, 29, 46, 0.62);
  border: 1px solid rgba(255, 255, 255, 0.14);
  box-shadow: 0 30px 70px rgba(5, 8, 16, 0.45);
  backdrop-filter: blur(22px);
  -webkit-backdrop-filter: blur(22px);
}

.login-brand {
  display: flex;
  justify-content: center;
  margin-bottom: 0.6rem;
}

.login-logo-wrap {
  position: relative;
  width: 96px;
  height: 96px;
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
  height: 100%;
  object-fit: contain;
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
  font-size: 1.5rem;
  line-height: 1.2;
}

.login-subtitle {
  margin: 0.6rem 0 0;
  color: var(--text-soft, #b3bbc4);
  font-size: 0.9rem;
  line-height: 1.5;
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
