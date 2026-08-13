<script setup>
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const avatarLetter = () => (authStore.user?.email || '?').charAt(0).toUpperCase()
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
        <div class="profile-avatar">{{ avatarLetter() }}</div>
        <div class="profile-info">
          <strong>{{ authStore.user?.email || 'Compte' }}</strong>
          <p>Compte principal</p>
        </div>
      </div>

      <div class="profile-actions">
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
}
</style>
