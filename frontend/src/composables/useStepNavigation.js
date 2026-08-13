// frontend/src/composables/useStepNavigation.js
import { ref, computed } from 'vue'

export function useStepNavigation(options = {}) {
  const currentStep = ref(options.initialStep || 1)
  const totalSteps = options.totalSteps || 3
  const stepsHistory = ref([currentStep.value])
  const canNavigate = ref(true)
  
  const steps = [
    { number: 1, label: 'Recherche', icon: '🔍' },
    { number: 2, label: 'Informations', icon: '📝' },
    { number: 3, label: 'Confirmation', icon: '✅' }
  ]
  
  const currentStepInfo = computed(() => {
    return steps.find(s => s.number === currentStep.value) || steps[0]
  })
  
  const isFirstStep = computed(() => currentStep.value === 1)
  const isLastStep = computed(() => currentStep.value === totalSteps)
  
  const canGoNext = computed(() => {
    return canNavigate.value && !isLastStep.value
  })
  
  const canGoPrevious = computed(() => {
    return canNavigate.value && !isFirstStep.value
  })
  
  const progress = computed(() => {
    return Math.round((currentStep.value / totalSteps) * 100)
  })
  
  const nextStep = () => {
    if (!canGoNext.value) {
      return { success: false, error: 'Impossible d\'avancer' }
    }
    
    currentStep.value++
    stepsHistory.value.push(currentStep.value)
    
    return { success: true, step: currentStep.value }
  }
  
  const previousStep = () => {
    if (!canGoPrevious.value) {
      return { success: false, error: 'Impossible de reculer' }
    }
    
    currentStep.value--
    
    return { success: true, step: currentStep.value }
  }
  
  const setStep = (step) => {
    if (!canNavigate.value) {
      return { success: false, error: 'Navigation désactivée' }
    }
    
    if (step < 1 || step > totalSteps) {
      return { success: false, error: 'Étape invalide' }
    }
    
    currentStep.value = step
    stepsHistory.value.push(step)
    
    return { success: true, step }
  }
  
  const goToStep = (stepNumber) => {
    return setStep(stepNumber)
  }
  
  const resetSteps = () => {
    currentStep.value = options.initialStep || 1
    stepsHistory.value = [currentStep.value]
    canNavigate.value = true
  }
  
  const disableNavigation = () => {
    canNavigate.value = false
  }
  
  const enableNavigation = () => {
    canNavigate.value = true
  }
  
  const hasVisitedStep = (step) => {
    return stepsHistory.value.includes(step)
  }
  
  return {
    currentStep,
    currentStepInfo,
    totalSteps,
    steps,
    stepsHistory,
    canNavigate,
    isFirstStep,
    isLastStep,
    canGoNext,
    canGoPrevious,
    progress,
    nextStep,
    previousStep,
    setStep,
    goToStep,
    resetSteps,
    disableNavigation,
    enableNavigation,
    hasVisitedStep
  }
}

export default useStepNavigation
