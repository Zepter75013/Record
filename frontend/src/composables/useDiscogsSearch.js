// frontend/src/composables/useDiscogsSearch.js
import { ref, computed } from 'vue'
import { discogsApi } from '@/services/discogsApi'
import { discValidation } from '@/services/discValidation'

export function useDiscogsSearch() {
  const title = ref('')
  const artist = ref('')
  const country = ref('')
  const year = ref('')
  const loading = ref(false)
  const searchAttempted = ref(false)
  const results = ref([])
  const selectedId = ref(null)
  const selectedResult = ref(null)
  const searchError = ref(null)
  const lastSearchType = ref('') // 'manual', 'barcode', or ''
  
  // ✅ CORRECTION: Utiliser la même variable que useApi.js
  const API_URL = import.meta.env.VITE_API_BASE_URL || '/api'
  
  const isValid = computed(() => {
    const validation = discValidation.validateSearchTerms(title.value, artist.value)
    return validation.valid
  })
  
  const validationMessage = computed(() => {
    if (!title.value && !artist.value) return null
    const validation = discValidation.validateSearchTerms(title.value, artist.value)
    return validation.error || null
  })
  
  const searchManual = async () => {
    if (!isValid.value) {
      searchError.value = 'Saisissez au moins 2 caractères dans le titre ou l\'artiste'
      return { success: false, error: searchError.value }
    }
    
    loading.value = true
    searchError.value = null
    searchAttempted.value = true
    lastSearchType.value = 'manual'
    
    try {
      const searchResults = await discogsApi.searchByTitleArtist(
        title.value, 
        artist.value,
        country.value,
        year.value
      )
      results.value = searchResults
      loading.value = false
      
      if (searchResults.length === 1) {
        selectResult(searchResults[0])
      }
      
      return { success: true, count: searchResults.length }
    } catch (error) {
      loading.value = false
      
      // Gestion spécifique des erreurs HTTP
      if (error.message.includes('405')) {
        searchError.value = 'Le serveur ne supporte pas cette méthode de recherche. Veuillez contacter l\'administrateur.'
      } else {
        searchError.value = error.message || 'Erreur lors de la recherche'
      }
      
      console.error('Erreur recherche manuelle:', error)
      return { success: false, error: searchError.value }
    }
  }
  
  const searchByBarcode = async (barcode) => {
    if (!barcode || barcode.trim().length === 0) {
      searchError.value = 'Code-barres vide'
      return { success: false, error: searchError.value }
    }
    
    const barcodeTrimmed = barcode.trim()
    const validation = discValidation.validateBarcode(barcodeTrimmed)
    if (!validation.valid) {
      searchError.value = validation.error
      return { success: false, error: searchError.value }
    }
    
    loading.value = true
    searchError.value = null
    searchAttempted.value = true
    lastSearchType.value = 'barcode'
    
    try {
      const result = await discogsApi.searchByBarcode(barcodeTrimmed)
      loading.value = false
      
      if (result) {
        results.value = [result]
        selectResult(result)
        return { success: true, data: result }
      } else {
        searchError.value = 'Aucun résultat trouvé pour ce code-barres'
        return { success: false, error: searchError.value }
      }
    } catch (error) {
      loading.value = false
      
      // Gestion spécifique des erreurs HTTP
      if (error.message.includes('405') || error.message.includes('Method Not Allowed')) {
        searchError.value = 'Le serveur ne supporte pas la recherche par code-barres. Essayez la recherche manuelle.'
        console.error('Erreur 405 - Configuration backend incorrecte:', error)
      } else if (error.message.includes('404')) {
        searchError.value = 'Aucun disque trouvé pour ce code-barres'
      } else if (error.message.includes('timeout') || error.message.includes('Timeout')) {
        searchError.value = 'Le serveur met trop de temps à répondre. Veuillez réessayer.'
      } else {
        searchError.value = error.message || 'Erreur lors de la recherche par code-barres'
      }
      
      console.error('Erreur recherche code-barres:', error)
      return { success: false, error: searchError.value }
    }
  }
  
  // Fonction de secours pour la recherche par code-barres
  const searchByBarcodeFallback = async (barcode) => {
    const barcodeTrimmed = barcode.trim()
    loading.value = true
    searchError.value = null
    
    try {
      // Essayer différentes méthodes si la première échoue
      let result = null
      
      // ✅ CORRECTION: Utiliser l'URL relative
      // Méthode 1: Essayer avec POST
      try {
        const response = await fetch(`${API_URL}/search-discogs`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ barcode: barcodeTrimmed })
        })
        
        if (response.ok) {
          result = await response.json()
        }
      } catch (e) {
        console.log('POST method failed, trying GET...', e)
      }
      
      // Méthode 2: Essayer avec GET
      if (!result) {
        try {
          // ✅ CORRECTION: Utiliser l'URL relative
          const response = await fetch(
            `${API_URL}/search-discogs?barcode=${encodeURIComponent(barcodeTrimmed)}`
          )
          
          if (response.ok) {
            result = await response.json()
          }
        } catch (e) {
          console.log('GET method also failed:', e)
        }
      }
      
      loading.value = false
      
      if (result) {
        results.value = [result]
        selectResult(result)
        return { success: true, data: result }
      } else {
        searchError.value = 'Impossible de contacter le serveur de recherche'
        return { success: false, error: searchError.value }
      }
    } catch (error) {
      loading.value = false
      searchError.value = 'Erreur technique lors de la recherche'
      console.error('Fallback search error:', error)
      return { success: false, error: searchError.value }
    }
  }
  
  const selectResult = (result) => {
    selectedId.value = result.id
    selectedResult.value = result
  }
  
  const clearSelection = () => {
    selectedId.value = null
    selectedResult.value = null
  }
  
  const fetchReleaseDetails = async (releaseId) => {
    if (!releaseId) {
      return { success: false, error: 'ID de release manquant' }
    }
    
    loading.value = true
    try {
      const release = await discogsApi.getRelease(releaseId)
      loading.value = false
      return { success: true, data: release }
    } catch (error) {
      loading.value = false
      console.error('Erreur récupération détails:', error)
      
      let errorMessage = error.message || 'Erreur lors de la récupération'
      if (error.message.includes('405')) {
        errorMessage = 'Le serveur ne supporte pas cette opération'
      }
      
      return { success: false, error: errorMessage }
    }
  }
  
  const fetchReleaseImages = async (releaseId) => {
    if (!releaseId) {
      return { success: false, error: 'ID de release manquant', images: [] }
    }
    
    try {
      const images = await discogsApi.getReleaseImages(releaseId)
      return { success: true, images }
    } catch (error) {
      console.error('Erreur récupération images:', error)
      return { success: false, error: error.message, images: [] }
    }
  }
  
  const retryLastSearch = async () => {
    if (lastSearchType.value === 'manual') {
      return await searchManual()
    } else if (lastSearchType.value === 'barcode') {
      // Pour retenter avec une méthode différente
      if (searchError.value?.includes('405')) {
        // Essayer avec la méthode de secours
        return await searchByBarcodeFallback(title.value || artist.value)
      }
      return await searchByBarcode(title.value || artist.value)
    }
    return { success: false, error: 'Aucune recherche précédente' }
  }
  
  const resetSearch = () => {
    title.value = ''
    artist.value = ''
    country.value = ''
    year.value = ''
    results.value = []
    selectedId.value = null
    selectedResult.value = null
    searchError.value = null
    searchAttempted.value = false
    loading.value = false
    lastSearchType.value = ''
  }
  
  const clearResults = () => {
    results.value = []
    selectedId.value = null
    selectedResult.value = null
  }
  
  return {
    title,
    artist,
    country,
    year,
    loading,
    searchAttempted,
    searchError,
    results,
    selectedId,
    selectedResult,
    lastSearchType,
    isValid,
    validationMessage,
    searchManual,
    searchByBarcode,
    searchByBarcodeFallback,
    selectResult,
    clearSelection,
    fetchReleaseDetails,
    fetchReleaseImages,
    retryLastSearch,
    resetSearch,
    clearResults
  }
}

export default useDiscogsSearch