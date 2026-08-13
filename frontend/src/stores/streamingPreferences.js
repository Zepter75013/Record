import { defineStore } from 'pinia'
import { ref } from 'vue'

const STORAGE_KEY = 'streamingPreferences'
const VALID_PLATFORMS = ['apple_music', 'spotify', 'deezer', 'youtube']

function loadStored() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY)) || {}
  } catch {
    return {}
  }
}

export const useStreamingPreferencesStore = defineStore('streamingPreferences', () => {
  const stored = loadStored()

  const preferredPlatform = ref(VALID_PLATFORMS.includes(stored.preferredPlatform) ? stored.preferredPlatform : null)

  function setPreferredPlatform(value) {
    if (value !== null && !VALID_PLATFORMS.includes(value)) return
    preferredPlatform.value = value
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ preferredPlatform: value }))
  }

  return { preferredPlatform, setPreferredPlatform }
})
