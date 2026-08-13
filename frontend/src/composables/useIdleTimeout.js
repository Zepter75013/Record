// frontend/src/composables/useIdleTimeout.js
import { ref, onMounted, onBeforeUnmount } from 'vue'

export function useIdleTimeout(config = {}) {
  const {
    timeout = 60 * 60 * 1000,
    warningTime = 55 * 60 * 1000
  } = config

  const idle = ref(false)
  const warning = ref(false)
  const lastActive = ref(Date.now())
  const remainingTime = ref(0)
  
  let timer = null
  let warningTimer = null
  let countdownInterval = null
  
  const events = [
    'mousedown',
    'mousemove',
    'keypress',
    'scroll',
    'touchstart',
    'click',
    'wheel'
  ]

  const clearTimers = () => {
    if (timer) clearTimeout(timer)
    if (warningTimer) clearTimeout(warningTimer)
    if (countdownInterval) clearInterval(countdownInterval)
  }

  const resetTimer = () => {
    idle.value = false
    warning.value = false
    lastActive.value = Date.now()
    
    clearTimers()
    
    warningTimer = setTimeout(() => {
      warning.value = true
      remainingTime.value = Math.floor((timeout - warningTime) / 1000)
      
      countdownInterval = setInterval(() => {
        remainingTime.value--
        if (remainingTime.value <= 0) {
          clearInterval(countdownInterval)
        }
      }, 1000)
    }, warningTime)
    
    timer = setTimeout(() => {
      idle.value = true
      warning.value = false
      clearTimers()
    }, timeout)
  }

  const start = () => {
    events.forEach(event => {
      window.addEventListener(event, resetTimer, true)
    })
    resetTimer()
  }

  const stop = () => {
    events.forEach(event => {
      window.removeEventListener(event, resetTimer, true)
    })
    clearTimers()
  }

  const reset = () => {
    resetTimer()
  }

  onMounted(() => {
    start()
  })

  onBeforeUnmount(() => {
    stop()
  })

  return {
    idle,
    warning,
    remainingTime,
    lastActive,
    reset,
    start,
    stop
  }
}
