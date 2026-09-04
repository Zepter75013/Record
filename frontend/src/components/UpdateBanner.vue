<script setup>
import { useVersionCheck } from '@/composables/useVersionCheck'

const { updateAvailable, reloadApp } = useVersionCheck()
</script>

<template>
  <Transition name="update-banner-slide">
    <div v-if="updateAvailable" class="update-banner" role="status">
      <span class="update-banner-text">
        🔄 Une nouvelle version de l'application est disponible.
      </span>
      <button type="button" class="update-banner-btn" @click="reloadApp">
        Recharger
      </button>
    </div>
  </Transition>
</template>

<style scoped>
.update-banner {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 12px 20px;
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  color: white;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
}

.update-banner-text {
  font-weight: 500;
  font-size: 0.95em;
}

.update-banner-btn {
  padding: 8px 18px;
  border: 1px solid rgba(255, 255, 255, 0.5);
  border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, 0.15);
  color: white;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s ease;
  flex-shrink: 0;
}

.update-banner-btn:hover {
  background: rgba(255, 255, 255, 0.28);
}

.update-banner-slide-enter-active,
.update-banner-slide-leave-active {
  transition: transform 0.25s ease, opacity 0.25s ease;
}

.update-banner-slide-enter-from,
.update-banner-slide-leave-to {
  transform: translateY(-100%);
  opacity: 0;
}

@media (max-width: 640px) {
  .update-banner {
    flex-direction: column;
    gap: 8px;
    text-align: center;
  }
}
</style>
