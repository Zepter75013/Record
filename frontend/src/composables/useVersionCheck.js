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

  // Portable en veille (couvercle fermé) : le minuteur ne tourne plus
  // pendant la veille, et visibilitychange ne se déclenche pas toujours
  // au réveil (l'onglet peut rester "visible" tout du long). Le focus de
  // la fenêtre est un signal plus fiable pour "l'utilisateur revient".
  function handleWindowFocus() {
    checkForUpdate()
  }

  // Restauration depuis le cache arrière/avant du navigateur (bfcache,
  // fréquent sur Safari/mobile) : la page ne recharge rien, donc aucun des
  // autres déclencheurs ne se déclenche sans ce cas particulier.
  function handlePageShow(event) {
    if (event.persisted) checkForUpdate()
  }

  function reloadApp() {
    window.location.reload()
  }

  onMounted(() => {
    checkForUpdate()
    intervalId = setInterval(checkForUpdate, CHECK_INTERVAL_MS)
    document.addEventListener('visibilitychange', handleVisibilityChange)
    window.addEventListener('focus', handleWindowFocus)
    window.addEventListener('pageshow', handlePageShow)
  })

  onUnmounted(() => {
    clearInterval(intervalId)
    document.removeEventListener('visibilitychange', handleVisibilityChange)
    window.removeEventListener('focus', handleWindowFocus)
    window.removeEventListener('pageshow', handlePageShow)
  })

  return { updateAvailable, reloadApp }
}
