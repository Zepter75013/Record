<script setup>
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { useApi } from '@/composables/useApi'

const authStore = useAuthStore()
const router = useRouter()
const { upload, del } = useApi()

const avatarLetter = () => (authStore.user?.email || '?').charAt(0).toUpperCase()

const getAvatarUrl = (path) => {
  if (!path) return ''
  if (path.startsWith('http')) return path
  const SERVER_BASE_URL = import.meta.env.VITE_SERVER_BASE_URL || ''
  return `${SERVER_BASE_URL}${path}`
}

const avatarInput = ref(null)
const isUploadingAvatar = ref(false)
const avatarError = ref(null)

function openAvatarPicker() {
  avatarInput.value?.click()
}

async function handleAvatarUpload(event) {
  const file = event.target.files?.[0]
  event.target.value = ''
  if (!file) return

  const validTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp']
  if (!validTypes.includes(file.type) || file.size > 5 * 1024 * 1024) {
    avatarError.value = 'Format non supporté ou fichier trop volumineux (max 5 Mo).'
    return
  }

  avatarError.value = null
  isUploadingAvatar.value = true
  try {
    const fd = new FormData()
    fd.append('avatar', file)
    const res = await upload('/user/avatar', fd)
    authStore.setAvatarPath(res.avatar_url)
  } catch (error) {
    avatarError.value = error.message
  } finally {
    isUploadingAvatar.value = false
  }
}

async function removeAvatar() {
  avatarError.value = null
  isUploadingAvatar.value = true
  try {
    await del('/user/avatar')
    authStore.setAvatarPath(null)
  } catch (error) {
    avatarError.value = error.message
  } finally {
    isUploadingAvatar.value = false
  }
}
</script>

<template>
  <div class="settings-profile-view">
    <header class="page-header">
      <div class="header-wrapper">
        <div class="header-top-row">
          <div class="title-section">
            <button @click="$router.push('/dashboard')" class="back-button ghost-btn">
              <span class="icon">🏠</span>
            </button>
            <span class="title-icon">👤</span>
            <h1>Mon compte</h1>
          </div>
        </div>
        <p class="subtitle">Informations de ton compte et sécurité.</p>
      </div>
    </header>

    <div class="content-panel panel">
      <div class="profile-summary">
        <button
          type="button"
          class="profile-avatar-button"
          :disabled="isUploadingAvatar"
          @click="openAvatarPicker"
          aria-label="Changer la photo de profil"
        >
          <img v-if="authStore.user?.avatar_path" :src="getAvatarUrl(authStore.user.avatar_path)" alt="Photo de profil" class="profile-avatar-img" />
          <div v-else class="profile-avatar">{{ avatarLetter() }}</div>
          <span class="profile-avatar-overlay">📷</span>
        </button>
        <input ref="avatarInput" type="file" accept="image/*" style="display: none" @change="handleAvatarUpload" />
        <div class="profile-info">
          <strong>{{ authStore.user?.email || 'Compte' }}</strong>
          <p>Compte principal</p>
        </div>
      </div>

      <p v-if="avatarError" class="avatar-error">{{ avatarError }}</p>

      <div class="profile-actions">
        <button class="ghost-btn" type="button" :disabled="isUploadingAvatar" @click="openAvatarPicker">
          📷 Changer la photo
        </button>
        <button v-if="authStore.user?.avatar_path" class="ghost-btn" type="button" :disabled="isUploadingAvatar" @click="removeAvatar">
          🗑️ Supprimer la photo
        </button>
        <button class="primary-btn" type="button" @click="router.push('/dashboard/change-password')">
          🔐 Changer le mot de passe
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.settings-profile-view {
  padding: 20px;
  min-height: 100vh;
  box-sizing: border-box;
}

.header-wrapper { padding: 0; margin: 0; }
.header-top-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 5px; flex-wrap: wrap; gap: 16px; }
.title-section { display: flex; align-items: center; gap: 12px; }
.back-button { display: none; padding: 0; width: 38px; height: 38px; border-radius: 50%; align-items: center; justify-content: center; }

@media (max-width: 767px) {
  .back-button { display: flex; }
  .settings-profile-view { padding: 10px; }
}

.title-icon { font-size: 2em; }
.page-header h1 { color: var(--text); font-size: 2em; margin: 0; font-weight: bold; }
.subtitle { color: var(--text-soft); margin: 0 0 20px 0; font-size: 1.1em; }

.content-panel {
  padding: 24px;
  max-width: 520px;
}

.profile-summary {
  display: flex;
  align-items: center;
  gap: 16px;
  padding-bottom: 20px;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--line-soft);
}

.profile-avatar {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  color: white;
  font-weight: 700;
  font-size: 1.4rem;
  flex-shrink: 0;
}

.profile-avatar-button {
  position: relative;
  width: 56px;
  height: 56px;
  flex-shrink: 0;
  padding: 0;
  border: none;
  border-radius: 16px;
  background: none;
  cursor: pointer;
  overflow: hidden;
}

.profile-avatar-button:disabled {
  cursor: default;
  opacity: 0.7;
}

.profile-avatar-img {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  object-fit: cover;
  display: block;
}

.profile-avatar-overlay {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  background: rgba(0, 0, 0, 0.45);
  color: white;
  font-size: 1.1rem;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.profile-avatar-button:hover .profile-avatar-overlay,
.profile-avatar-button:focus-visible .profile-avatar-overlay {
  opacity: 1;
}

.avatar-error {
  color: var(--negative-text, #e05555);
  font-size: 0.88rem;
  margin: -10px 0 16px;
}

.profile-info strong {
  display: block;
  color: var(--text);
  font-size: 1.1rem;
}

.profile-info p {
  margin-top: 4px;
  color: var(--text-dim);
  font-size: 0.88rem;
}

.profile-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
</style>
