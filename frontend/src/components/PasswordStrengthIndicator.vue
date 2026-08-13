<template>
  <div class="password-strength" v-if="password">
    <div class="strength-bar">
      <div 
        class="strength-fill" 
        :style="{ width: `${(strength.score / 5) * 100}%`, backgroundColor: strength.color }"
      ></div>
    </div>
    <div class="strength-label" :style="{ color: strength.color }">
      {{ strength.label }}
    </div>
    
    <ul class="requirements-list">
      <li 
        v-for="(req, index) in requirements" 
        :key="index"
        :class="{ 'met': req.met }"
      >
        <span class="icon">{{ req.met ? '✓' : '○' }}</span>
        {{ req.text }}
      </li>
    </ul>
  </div>
</template>

<script setup>
defineProps({
  password: String,
  strength: Object,
  requirements: Array
})
</script>

<style scoped>
.password-strength {
  margin-top: 0.5rem;
}

.strength-bar {
  height: 6px;
  background-color: #e5e7eb;
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 0.5rem;
}

.strength-fill {
  height: 100%;
  transition: width 0.3s ease, background-color 0.3s ease;
}

.strength-label {
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
}

.requirements-list {
  list-style: none;
  padding: 0;
  margin: 0;
  font-size: 0.875rem;
}

.requirements-list li {
  padding: 0.25rem 0;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.requirements-list li.met {
  color: #22c55e;
}

.icon {
  font-weight: bold;
  width: 1rem;
  display: inline-block;
}
</style>
