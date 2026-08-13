import { ref, computed, watch } from 'vue'

export function usePasswordValidation() {
  const password = ref('')
  
  const strength = computed(() => {
    if (!password.value) return { score: 0, label: '', color: '' }
    
    let score = 0
    const checks = {
      length: password.value.length >= 8,
      upper: /[A-Z]/.test(password.value),
      lower: /[a-z]/.test(password.value),
      number: /[0-9]/.test(password.value),
      special: /[!@#$%^&*(),.?":{}|<>]/.test(password.value)
    }
    
    if (checks.length) score++
    if (checks.upper) score++
    if (checks.lower) score++
    if (checks.number) score++
    if (checks.special) score++
    
    if (score <= 2) return { score, label: 'Faible', color: '#ef4444' }
    if (score === 3) return { score, label: 'Moyen', color: '#f97316' }
    if (score === 4) return { score, label: 'Bon', color: '#eab308' }
    return { score, label: 'Excellent', color: '#22c55e' }
  })
  
  const requirements = computed(() => [
    { text: 'Au moins 8 caractères', met: password.value.length >= 8 },
    { text: 'Une lettre majuscule', met: /[A-Z]/.test(password.value) },
    { text: 'Une lettre minuscule', met: /[a-z]/.test(password.value) },
    { text: 'Un chiffre', met: /[0-9]/.test(password.value) },
    { text: 'Un caractère spécial', met: /[!@#$%^&*(),.?":{}|<>]/.test(password.value) }
  ])
  
  const isValid = computed(() => {
    return requirements.value.every(req => req.met)
  })
  
  return {
    password,
    strength,
    requirements,
    isValid
  }
}
