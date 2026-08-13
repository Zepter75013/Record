// frontend/src/composables/useClipboardWatch.js
import { ref, onUnmounted } from 'vue'
import { discValidation } from '@/services/discValidation'

export function useClipboardWatch(onBarcodeDetected) {
  const isWatching = ref(false)
  const lastClipboardContent = ref('')
  const intervalId = ref(null)
  const watchInterval = 500 // ms
  
  const isClipboardSupported = () => {
    return !!(navigator.clipboard && navigator.clipboard.readText)
  }
  
  const startWatching = async () => {
    if (!isClipboardSupported()) {
      console.error('API Clipboard non supportée par ce navigateur')
      return { success: false, error: 'Clipboard API non supporté' }
    }
    
    try {
      // Lire le contenu initial
      const initialContent = await navigator.clipboard.readText()
      lastClipboardContent.value = initialContent
      isWatching.value = true
      
      // Commencer la surveillance
      intervalId.value = setInterval(async () => {
        try {
          const currentContent = await navigator.clipboard.readText()
          
          // Si le contenu a changé
          if (currentContent !== lastClipboardContent.value) {
            lastClipboardContent.value = currentContent
            
            // Vérifier si c'est un code-barres valide
            const cleanContent = currentContent.trim()
            const validation = discValidation.validateBarcode(cleanContent)
            
            if (validation.valid) {
              // Code-barres détecté !
              if (typeof onBarcodeDetected === 'function') {
                onBarcodeDetected(cleanContent)
              }
              
              // Arrêter la surveillance après détection
              stopWatching()
            }
          }
        } catch (error) {
          // Erreur de permission silencieuse
          if (error.name === 'NotAllowedError') {
            console.warn('Permission refusée pour lire le presse-papier')
            stopWatching()
          }
        }
      }, watchInterval)
      
      return { success: true }
    } catch (error) {
      console.error('Erreur démarrage surveillance:', error)
      return { success: false, error: error.message }
    }
  }
  
  const stopWatching = () => {
    if (intervalId.value) {
      clearInterval(intervalId.value)
      intervalId.value = null
    }
    isWatching.value = false
  }
  
  const toggleWatching = async () => {
    if (isWatching.value) {
      stopWatching()
      return { success: true, watching: false }
    } else {
      const result = await startWatching()
      return { ...result, watching: result.success }
    }
  }
  
  const readClipboardOnce = async () => {
    if (!isClipboardSupported()) {
      return { success: false, error: 'Clipboard API non supporté' }
    }
    
    try {
      const content = await navigator.clipboard.readText()
      const cleanContent = content.trim()
      
      const validation = discValidation.validateBarcode(cleanContent)
      
      if (validation.valid) {
        return { success: true, barcode: cleanContent }
      } else {
        return { success: false, error: 'Le presse-papier ne contient pas un code-barres valide' }
      }
    } catch (error) {
      console.error('Erreur lecture presse-papier:', error)
      return { success: false, error: error.message }
    }
  }
  
  const writeToClipboard = async (text) => {
    if (!navigator.clipboard || !navigator.clipboard.writeText) {
      return { success: false, error: 'Clipboard write non supporté' }
    }
    
    try {
      await navigator.clipboard.writeText(text)
      return { success: true }
    } catch (error) {
      console.error('Erreur écriture presse-papier:', error)
      return { success: false, error: error.message }
    }
  }
  
  // Cleanup
  onUnmounted(() => {
    stopWatching()
  })
  
  return {
    isWatching,
    lastClipboardContent,
    isClipboardSupported: isClipboardSupported(),
    startWatching,
    stopWatching,
    toggleWatching,
    readClipboardOnce,
    writeToClipboard
  }
}

export default useClipboardWatch
