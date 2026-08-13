// frontend/src/composables/useCoverImage.js
import { ref, computed } from 'vue'
import { imageHelpers } from '@/services/imageHelpers'
import { discogsApi } from '@/services/discogsApi'

export function useCoverImage() {
  const imageUrl = ref(null)
  const imageFile = ref(null)
  const discogsImages = ref([])
  const selectedImageIndex = ref(null)
  const activeTab = ref('preview')
  const uploadProgress = ref(0)
  const uploadError = ref(null)
  const loading = ref(false)
  
  const hasDiscogsImages = computed(() => discogsImages.value.length > 0)
  const hasImage = computed(() => !!imageUrl.value)
  
  const handleUpload = async (event) => {
    const file = event.target.files[0]
    if (!file) return { success: false, error: 'Aucun fichier sélectionné' }
    
    const validation = imageHelpers.validateImageFile(file)
    if (!validation.valid) {
      uploadError.value = validation.error
      return { success: false, error: validation.error }
    }
    
    try {
      // Redimensionner l'image si nécessaire
      let processedFile = file
      const dimensions = await imageHelpers.getImageDimensions(file)
      
      if (dimensions.width > 800 || dimensions.height > 800) {
        processedFile = await imageHelpers.resizeImage(file, 800, 800, 0.85)
      }
      
      imageFile.value = processedFile
      imageUrl.value = imageHelpers.createPreviewUrl(processedFile)
      uploadError.value = null
      activeTab.value = 'preview'
      
      return { success: true }
    } catch (error) {
      console.error('Erreur traitement image:', error)
      uploadError.value = 'Erreur lors du traitement de l\'image'
      return { success: false, error: uploadError.value }
    }
  }
  
  const captureFromCamera = async () => {
    // Cette méthode ouvrira le composant BarcodeScanner en mode caméra
    return { success: true, action: 'open_camera' }
  }
  
  const selectDiscogsImage = (index) => {
    if (index < 0 || index >= discogsImages.value.length) {
      return { success: false, error: 'Index invalide' }
    }
    
    selectedImageIndex.value = index
    const selectedImage = discogsImages.value[index]
    imageUrl.value = imageHelpers.getOptimizedDiscogsUrl(selectedImage.uri, '500')
    imageFile.value = null // Pas de fichier local pour une image Discogs
    activeTab.value = 'preview'
    
    return { success: true }
  }
  
  const removeImage = () => {
    if (imageUrl.value && imageUrl.value.startsWith('blob:')) {
      imageHelpers.revokePreviewUrl(imageUrl.value)
    }
    
    imageUrl.value = null
    imageFile.value = null
    selectedImageIndex.value = null
    uploadError.value = null
  }
  
  const loadDiscogsImages = (images) => {
    if (!Array.isArray(images)) {
      console.error('Images Discogs invalides:', images)
      return
    }
    
    discogsImages.value = images
    
    // Sélectionner automatiquement la première image "primary" s'il y en a une
    const primaryIndex = images.findIndex(img => img.type === 'primary')
    if (primaryIndex !== -1) {
      selectDiscogsImage(primaryIndex)
    } else if (images.length > 0) {
      // Sinon prendre la première image
      selectDiscogsImage(0)
    }
  }
  
  const uploadToServer = async (discId = null) => {
    if (!imageFile.value) {
      return { success: false, error: 'Aucune image à uploader' }
    }
    
    loading.value = true
    uploadProgress.value = 0
    
    try {
      const result = await discogsApi.uploadCoverImage(imageFile.value, discId)
      loading.value = false
      uploadProgress.value = 100
      
      return { success: true, data: result }
    } catch (error) {
      loading.value = false
      uploadError.value = error.message || 'Erreur lors de l\'upload'
      console.error('Erreur upload:', error)
      
      return { success: false, error: uploadError.value }
    }
  }
  
  const changeImage = () => {
    activeTab.value = 'upload'
  }
  
  const resetImages = () => {
    removeImage()
    discogsImages.value = []
    selectedImageIndex.value = null
    activeTab.value = 'preview'
    uploadError.value = null
    uploadProgress.value = 0
  }
  
  const getImageData = () => {
    return {
      url: imageUrl.value,
      file: imageFile.value,
      isDiscogs: selectedImageIndex.value !== null,
      discogsIndex: selectedImageIndex.value
    }
  }
  
  return {
    imageUrl,
    imageFile,
    discogsImages,
    selectedImageIndex,
    activeTab,
    uploadProgress,
    uploadError,
    loading,
    hasDiscogsImages,
    hasImage,
    handleUpload,
    captureFromCamera,
    selectDiscogsImage,
    removeImage,
    loadDiscogsImages,
    uploadToServer,
    changeImage,
    resetImages,
    getImageData
  }
}

export default useCoverImage
