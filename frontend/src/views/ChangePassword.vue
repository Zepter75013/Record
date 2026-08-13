<template>
  <div class="settings-password-view">
    <header class="page-header">
      <div class="header-wrapper">
        <div class="header-top-row">
          <div class="title-section">
            <button @click="$router.push('/dashboard')" class="back-button ghost-btn">
              <span class="icon">🏠</span>
            </button>
            <span class="title-icon">🔐</span>
            <h1>Changer mon mot de passe</h1>
          </div>
        </div>
        <p class="subtitle">Modifiez le mot de passe de votre compte en toute sécurité.</p>
        <p class="user-hint">
          <span>💡</span>
          <strong>Astuce :</strong> Choisissez un mot de passe fort avec majuscules, chiffres et caractères spéciaux
        </p>
      </div>
    </header>

    <div class="content-panel panel">
      <form @submit.prevent="handleSubmit" class="password-form">
        <label class="form-field">
          <span>Ancien mot de passe *</span>
          <div class="password-field">
            <input
              :type="showOldPassword ? 'text' : 'password'"
              v-model="oldPassword"
              required
              placeholder="Entrez votre ancien mot de passe"
              :disabled="loading"
            />
            <button
              type="button"
              class="password-toggle"
              @click="showOldPassword = !showOldPassword"
              :aria-label="showOldPassword ? 'Masquer' : 'Afficher'"
            >
              {{ showOldPassword ? '👁️' : '👁️‍🗨️' }}
            </button>
          </div>
        </label>

        <label class="form-field">
          <span>Nouveau mot de passe *</span>
          <div class="password-field">
            <input
              :type="showNewPassword ? 'text' : 'password'"
              v-model="passwordValidation.password.value"
              required
              placeholder="Entrez votre nouveau mot de passe"
              :disabled="loading"
            />
            <button
              type="button"
              class="password-toggle"
              @click="showNewPassword = !showNewPassword"
              :aria-label="showNewPassword ? 'Masquer' : 'Afficher'"
            >
              {{ showNewPassword ? '👁️' : '👁️‍🗨️' }}
            </button>
          </div>

          <PasswordStrengthIndicator
            :password="passwordValidation.password.value"
            :strength="passwordValidation.strength.value"
            :requirements="passwordValidation.requirements.value"
          />
        </label>

        <label class="form-field">
          <span>Confirmer le nouveau mot de passe *</span>
          <div class="password-field">
            <input
              :type="showConfirmPassword ? 'text' : 'password'"
              v-model="confirmPassword"
              required
              placeholder="Confirmez votre nouveau mot de passe"
              :disabled="loading"
            />
            <button
              type="button"
              class="password-toggle"
              @click="showConfirmPassword = !showConfirmPassword"
              :aria-label="showConfirmPassword ? 'Masquer' : 'Afficher'"
            >
              {{ showConfirmPassword ? '👁️' : '👁️‍🗨️' }}
            </button>
          </div>
          <p v-if="confirmPassword && confirmPassword !== passwordValidation.password.value" class="form-error">
            Les mots de passe ne correspondent pas
          </p>
        </label>

        <p v-if="error" class="form-error">{{ error }}</p>
        <p v-if="success" class="form-success">{{ success }}</p>

        <div class="button-group">
          <button
            type="button"
            class="ghost-btn"
            @click="goBack"
            :disabled="loading"
          >
            <span class="icon">↩️</span>
            Annuler
          </button>
          <button
            type="submit"
            class="primary-btn"
            :disabled="!canSubmit || loading"
          >
            <span class="icon">✓</span>
            {{ loading ? 'Modification en cours...' : 'Changer le mot de passe' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '@/composables/useApi'
import { usePasswordValidation } from '@/composables/usePasswordValidation'
import { useAuthStore } from '@/stores/auth'
import PasswordStrengthIndicator from '@/components/PasswordStrengthIndicator.vue'

const router = useRouter()
const { post } = useApi()
const authStore = useAuthStore()
const passwordValidation = usePasswordValidation()

const oldPassword = ref('')
const confirmPassword = ref('')
const showOldPassword = ref(false)
const showNewPassword = ref(false)
const showConfirmPassword = ref(false)
const loading = ref(false)
const error = ref('')
const success = ref('')

const canSubmit = computed(() => {
  return oldPassword.value &&
         passwordValidation.password.value &&
         confirmPassword.value &&
         passwordValidation.isValid.value &&
         passwordValidation.password.value === confirmPassword.value
})

const handleSubmit = async () => {
  if (!canSubmit.value) return

  error.value = ''
  success.value = ''
  loading.value = true

  try {
    const response = await post('/user/change-password', {
      old_password: oldPassword.value,
      new_password: passwordValidation.password.value
    })

    success.value = response.message

    // Déconnexion après 2 secondes
    setTimeout(() => {
      authStore.logout()
    }, 2000)

  } catch (err) {
    error.value = err.message || 'Une erreur est survenue'
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.push('/dashboard')
}
</script>

<style scoped>
.settings-password-view {
  padding: 20px;
  min-height: 100vh;
  box-sizing: border-box;
}

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

.title-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-button {
  display: none;
  padding: 0;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  align-items: center;
  justify-content: center;
}

@media (max-width: 767px) {
  .back-button {
    display: flex;
  }

  .settings-password-view {
    padding: 10px;
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

.subtitle {
  color: var(--text-soft);
  margin: 0 0 10px 0;
  font-size: 1.1em;
}

.user-hint {
  color: var(--text-soft);
  margin: 0 0 20px 0;
  background: rgba(var(--tint-rgb), 0.04);
  padding: 8px 12px;
  border-radius: 10px;
  border-left: 4px solid var(--accent);
}

.content-panel {
  padding: 30px;
  max-width: 600px;
  margin: 0 auto;
}

.password-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
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
  font-size: 1.1rem;
  cursor: pointer;
  opacity: 0.7;
  transition: opacity 0.2s ease, background 0.2s ease;
}

.password-toggle:hover {
  opacity: 1;
  background: rgba(var(--tint-rgb), 0.06);
}

.button-group {
  display: flex;
  gap: 12px;
  margin-top: 10px;
}

.button-group .ghost-btn,
.button-group .primary-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

@media (max-width: 480px) {
  .content-panel {
    padding: 20px;
  }

  .button-group {
    flex-direction: column;
  }
}
</style>
