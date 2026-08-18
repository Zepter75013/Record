<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '@/composables/useApi'

const router = useRouter()
const { post } = useApi()

const email = ref('')
const loading = ref(false)
const success = ref(false)
const error = ref('')

const handleSubmit = async () => {
  if (!email.value || loading.value) return

  error.value = ''
  loading.value = true

  try {
    await post('/password-reset/request', {
      email: email.value
    })

    success.value = true
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
          <h1>Mot de passe oublié</h1>
          <p class="login-subtitle">
            Entrez votre adresse email pour recevoir un lien de réinitialisation.
          </p>
        </div>

        <form class="login-form" @submit.prevent="handleSubmit">
          <label class="form-field">
            <span>Adresse email</span>
            <input
              v-model="email"
              type="email"
              autocomplete="email"
              placeholder="votre@email.com"
              required
              :disabled="loading"
              autofocus
            />
          </label>

          <p v-if="error" class="form-error">{{ error }}</p>

          <button class="primary-btn login-submit" type="submit" :disabled="loading || !email">
            {{ loading ? 'Envoi en cours...' : 'Envoyer le lien de réinitialisation' }}
          </button>

          <button class="login-link-btn" type="button" @click="router.push('/login')">
            Retour à la connexion
          </button>
        </form>

        <div class="login-info-box">
          <p class="login-info-title">Contactez l'administrateur</p>
          <p class="login-info-text">Si vous n'avez pas accès à votre email, contactez :</p>
          <p class="login-info-email">admin@vinylmanager.com</p>
        </div>
      </template>

      <template v-else>
        <div class="login-heading">
          <p class="eyebrow">Sécurité</p>
          <h1>Email envoyé !</h1>
        </div>

        <p class="form-success">
          Si votre adresse email est enregistrée, vous recevrez un lien de réinitialisation dans
          quelques instants. Vérifiez également votre dossier spam.
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
  width: min(100%, 420px);
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

.login-info-box {
  margin-top: 1.4rem;
  padding: 1rem 1.1rem;
  border-radius: 14px;
  background: rgba(var(--tint-rgb), 0.04);
  border: 1px solid rgba(var(--tint-rgb), 0.08);
}

.login-info-title {
  margin: 0 0 4px;
  color: var(--text, #eef1f3);
  font-size: 0.84rem;
  font-weight: 700;
}

.login-info-text {
  margin: 0 0 4px;
  color: var(--text-soft, #b3bbc4);
  font-size: 0.82rem;
}

.login-info-email {
  margin: 0;
  color: var(--text, #eef1f3);
  font-size: 0.82rem;
  font-weight: 600;
}

@media (max-width: 480px) {
  .login-card {
    padding: 1.6rem 1.3rem;
  }
}
</style>
