// frontend/src/composables/useFormValidation.js
import { ref, computed, watch } from 'vue'
import { discValidation } from '@/services/discValidation'

export function useFormValidation(formData) {
  const errors = ref({})
  const touched = ref({})
  const isValidating = ref(false)
  
  const fields = {
    artist: { required: true, validator: discValidation.validateArtist },
    title: { required: true, validator: discValidation.validateTitle },
    year: { required: false, validator: discValidation.validateYear },
    barcode: { required: false, validator: discValidation.validateBarcode },
    genre: { required: false, validator: discValidation.validateGenre },
    format: { required: false, validator: discValidation.validateFormat },
    label: { required: false, validator: discValidation.validateLabel },
    country: { required: false, validator: discValidation.validateCountry }
  }
  
  const validateField = (fieldName) => {
    const field = fields[fieldName]
    if (!field) return true
    
    const value = formData[fieldName]
    const validation = field.validator(value)
    
    if (!validation.valid) {
      errors.value[fieldName] = validation.error
      return false
    }
    
    errors.value[fieldName] = null
    return true
  }
  
  const validateAll = () => {
    isValidating.value = true
    let isValid = true
    
    for (const fieldName in fields) {
      if (!validateField(fieldName)) {
        isValid = false
      }
    }
    
    isValidating.value = false
    return isValid
  }
  
  const markTouched = (fieldName) => {
    touched.value[fieldName] = true
    validateField(fieldName)
  }
  
  const markAllTouched = () => {
    for (const fieldName in fields) {
      touched.value[fieldName] = true
    }
  }
  
  const clearError = (fieldName) => {
    errors.value[fieldName] = null
  }
  
  const clearAllErrors = () => {
    errors.value = {}
  }
  
  const resetValidation = () => {
    errors.value = {}
    touched.value = {}
    isValidating.value = false
  }
  
  const getFieldError = (fieldName) => {
    if (!touched.value[fieldName]) return null
    return errors.value[fieldName]
  }
  
  const hasErrors = computed(() => {
    return Object.values(errors.value).some(error => error !== null)
  })
  
  const hasRequiredFields = computed(() => {
    return discValidation.hasRequiredFields(formData)
  })
  
  const isFormValid = computed(() => {
    return hasRequiredFields.value && !hasErrors.value
  })
  
  // Watchers pour validation en temps réel
  watch(() => formData.artist, () => {
    if (touched.value.artist) validateField('artist')
  })
  
  watch(() => formData.title, () => {
    if (touched.value.title) validateField('title')
  })
  
  watch(() => formData.year, () => {
    if (touched.value.year) validateField('year')
  })
  
  watch(() => formData.barcode, () => {
    if (touched.value.barcode) validateField('barcode')
  })
  
  return {
    errors,
    touched,
    isValidating,
    hasErrors,
    hasRequiredFields,
    isFormValid,
    validateField,
    validateAll,
    markTouched,
    markAllTouched,
    clearError,
    clearAllErrors,
    resetValidation,
    getFieldError
  }
}

export default useFormValidation
