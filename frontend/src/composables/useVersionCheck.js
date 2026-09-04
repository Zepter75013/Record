import { ref, onMounted, onUnmounted } from 'vue'

// Vérifie périodiquement si une nouvelle version de l'app a été déployée,
// en comparant le index.html servi (jamais mis en cache — voir nginx.conf)
// à celui chargé au démarrage de cet onglet. S'il a changé, les fichiers
// JS/CSS hashés qu'il référence ont changé aussi : nouvelle version.
const CHECK_INTERVAL_MS = 5 * 60 * 1000 // 5 minutes

export function useVersionCheck() {
  const updateAvailable = ref(false)
  let loadedHtml = null
  let intervalId = null

  async function fetchCurrentIndexHtml() {
    try {
      const res = await fetch('/', { cache: 'no-store' })
      if (!res.ok) return null
      return await res.text()
    } catch {
      return null
    }
  }

  async function checkForUpdate() {
    const html = await fetchCurrentIndexHtml()
    if (!html) return
    if (loadedHtml === null) {
      loadedHtml = html
      return
    }
    if (html !== loadedHtml) {
      updateAvailable.value = true
    }
  }

  function handleVisibilityChange() {
    if (document.visibilityState === 'visible') checkForUpdate()
  }

  function reloadApp() {
    window.location.reload()
  }

  onMounted(() => {
    checkForUpdate()
    intervalId = setInterval(checkForUpdate, CHECK_INTERVAL_MS)
    document.addEventListener('visibilitychange', handleVisibilityChange)
  })

  onUnmounted(() => {
    clearInterval(intervalId)
    document.removeEventListener('visibilitychange', handleVisibilityChange)
  })

  return { updateAvailable, reloadApp }
}
