// frontend/src/composables/useBarcodeScanner.js
import { ref, onMounted, onUnmounted } from 'vue'
import { discogsApi } from '@/services/discogsApi'
import { discValidation } from '@/services/discValidation'

// onResult(outcome) est appelé après CHAQUE tentative de recherche, quel que
// soit le déclencheur (saisie/Entrée, douchette physique, import fichier) —
// utile pour un composant qui veut afficher un résultat unique peu importe
// comment le code-barres a été obtenu, sans dupliquer cette logique.
export function useBarcodeScanner(onResult) {
  const barcode = ref('')
  const scanBuffer = ref('')
  const scannerDetected = ref(false)
  const scanInProgress = ref(false)
  const scanServerActive = ref(false)
  const lastKeyTime = ref(0)
  const scanTimeout = ref(null)
  
  const SCAN_THRESHOLD_MS = 50
  const MIN_SCAN_LENGTH = 8
  const BUFFER_TIMEOUT_MS = 100
  
  const handleKeypress = (event) => {
    const currentTime = Date.now()
    const timeDiff = currentTime - lastKeyTime.value
    
    if (event.ctrlKey || event.altKey || event.metaKey) return
    
    if (timeDiff < SCAN_THRESHOLD_MS && event.key.length === 1) {
      scannerDetected.value = true
      scanBuffer.value += event.key
      
      clearTimeout(scanTimeout.value)
      scanTimeout.value = setTimeout(() => finalizeScan(), BUFFER_TIMEOUT_MS)
    } else if (scannerDetected.value && event.key === 'Enter') {
      clearTimeout(scanTimeout.value)
      finalizeScan()
    } else {
      resetScannerBuffer()
    }
    
    lastKeyTime.value = currentTime
  }
  
  const finalizeScan = async () => {
    if (scanBuffer.value.length >= MIN_SCAN_LENGTH) {
      barcode.value = scanBuffer.value
      if (isValidBarcode(scanBuffer.value)) {
        await searchByBarcode(scanBuffer.value)
      }
    }
    resetScannerBuffer()
  }
  
  const resetScannerBuffer = () => {
    scanBuffer.value = ''
    scannerDetected.value = false
    clearTimeout(scanTimeout.value)
  }
  
  const isValidBarcode = (code) => {
    const validation = discValidation.validateBarcode(code)
    return validation.valid
  }
  
  const searchByBarcode = async (code) => {
    if (!code || code.trim().length === 0) {
      return { success: false, error: 'Code-barres vide' }
    }
    
    scanInProgress.value = true

    try {
      const existingDisc = await discogsApi.checkBarcodeExists(code)

      if (existingDisc) {
        scanInProgress.value = false
        const outcome = { success: true, exists: true, disc: existingDisc }
        onResult?.(outcome)
        return outcome
      }

      const result = await discogsApi.searchByBarcode(code)
      scanInProgress.value = false

      let outcome
      if (result) {
        outcome = { success: true, exists: false, data: result }
      } else {
        outcome = { success: false, error: 'Aucun résultat trouvé sur Discogs' }
      }
      onResult?.(outcome)
      return outcome
    } catch (error) {
      scanInProgress.value = false
      console.error('Erreur recherche code-barres:', error)
      const outcome = { success: false, error: error.message || 'Erreur lors de la recherche' }
      onResult?.(outcome)
      return outcome
    }
  }
  
  const startAutomaticScan = async () => {
    scanServerActive.value = true
    scanInProgress.value = true
    
    try {
      const response = await fetch('http://localhost:8081/scan')
      const data = await response.json()
      
      if (data.barcode) {
        barcode.value = data.barcode
        await searchByBarcode(data.barcode)
      }
    } catch (error) {
      console.error('Erreur scan automatique:', error)
      return { success: false, error: 'Impossible de se connecter au serveur de scan' }
    } finally {
      scanServerActive.value = false
      scanInProgress.value = false
    }
  }
  
  const cancelAutomaticScan = () => {
    scanServerActive.value = false
    scanInProgress.value = false
    barcode.value = ''
  }
  
  const handleBarcodeFocus = () => {
    resetScannerBuffer()
  }
  
  const handleBarcodeEnter = async () => {
    if (barcode.value && barcode.value.trim().length > 0) {
      return await searchByBarcode(barcode.value)
    }
  }
  
  const handleFileImport = async (event) => {
    const file = event.target.files[0]
    if (!file) return
    
    try {
      const text = await file.text()
      const lines = text.split('\n')
      
      for (const line of lines) {
        const cleanLine = line.trim()
        if (isValidBarcode(cleanLine)) {
          barcode.value = cleanLine
          return await searchByBarcode(cleanLine)
        }
      }
      
      return { success: false, error: 'Aucun code-barres valide trouvé dans le fichier' }
    } catch (error) {
      console.error('Erreur lecture fichier:', error)
      return { success: false, error: 'Erreur lors de la lecture du fichier' }
    }
  }
  
  const checkBarcodeExists = async () => {
    if (!barcode.value || barcode.value.trim().length === 0) return null
    
    try {
      const existingDisc = await discogsApi.checkBarcodeExists(barcode.value)
      return existingDisc
    } catch (error) {
      console.error('Erreur vérification:', error)
      return null
    }
  }
  
  onMounted(() => {
    window.addEventListener('keypress', handleKeypress)
  })
  
  onUnmounted(() => {
    window.removeEventListener('keypress', handleKeypress)
    clearTimeout(scanTimeout.value)
  })
  
  return {
    barcode,
    scanBuffer,
    scannerDetected,
    scanInProgress,
    scanServerActive,
    searchByBarcode,
    startAutomaticScan,
    cancelAutomaticScan,
    handleBarcodeFocus,
    handleBarcodeEnter,
    handleFileImport,
    checkBarcodeExists,
    resetScannerBuffer,
    isValidBarcode
  }
}

export default useBarcodeScanner
