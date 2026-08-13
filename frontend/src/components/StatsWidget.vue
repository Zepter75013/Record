<!-- components/StatsWidget.vue - VERSION RESPONSIVE MOBILE -->
<script setup>
import { computed, watch, ref, onMounted } from 'vue';
import { useCounter } from '@/composables/useCounter';

const props = defineProps({
  type: {
    type: String,
    required: true,
    validator: (value) => ['vinyls', 'artists', 'genres', 'formats', 'countries', 'labels'].includes(value)
  },
  value: {
    type: Number,
    default: 0
  },
  config: {
    type: Object,
    default: () => ({
      icon: '💿',
      color: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      title: 'Disques'
    })
  },
  isMain: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['click']);

// Initialiser le compteur avec la valeur initiale
const counter = useCounter(1000); // 1 seconde d'animation
const prevValue = ref(props.value);

// Démarrer avec la valeur initiale au montage
onMounted(() => {
  console.log(`🔢 [StatsWidget/${props.type}] Monté avec valeur: ${props.value}`);
  counter.start(props.value);
  prevValue.value = props.value;
});

// Observer les changements de la valeur
watch(() => props.value, (newValue, oldValue) => {
  console.log(`🔄 [StatsWidget/${props.type}] ${oldValue} → ${newValue}`);
  if (newValue !== prevValue.value && newValue >= 0) {
    counter.start(newValue);
    prevValue.value = newValue;
  }
});

// Classes CSS conditionnelles
const widgetClass = computed(() => ({
  'vinyls-widget': props.isMain,
  'thematic-widget': !props.isMain,
  'stats-widget': true
}));

const contentClass = computed(() => ({
  'vinyls-content': props.isMain,
  'thematic-content': !props.isMain,
  'stats-content': true
}));
</script>

<template>
  <div 
    :class="widgetClass"
    :style="{ background: config.color }"
    @click="emit('click')"
  >
    <div :class="contentClass">
      <div class="stats-icon">{{ config.icon }}</div>
      <div class="stats-text">
        <div class="stats-main">
          <p class="stats-count">{{ counter.current }}</p>
          <h3 class="stats-title">{{ config.title }}</h3>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.stats-widget {
  background: white;
  padding: 14px 18px;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
  border: none;
  color: white;
  position: relative;
  overflow: hidden;
  cursor: pointer;
  /* IMPORTANT pour le responsive grid */
  width: 100%;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
}

.vinyls-widget {
  min-height: 80px;
}

.thematic-widget {
  min-height: 70px;
}

.stats-widget::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0) 100%);
  z-index: 1;
}

.stats-widget:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15);
}

.stats-content {
  display: flex;
  align-items: center;
  z-index: 2;
  position: relative;
  height: 100%;
  min-width: 0;
}

.vinyls-content {
  align-items: center;
}

.thematic-content {
  align-items: center;
}

.stats-icon {
  font-size: 1.8em;
  filter: drop-shadow(0 2px 6px rgba(0, 0, 0, 0.3));
  flex-shrink: 0;
  margin-right: 15px;
}

.vinyls-content .stats-icon {
  font-size: 2em;
}

.thematic-content .stats-icon {
  font-size: 1.8em;
  margin-right: 12px;
}

.stats-text {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.stats-main {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.vinyls-content .stats-main {
  gap: 12px;
}

.thematic-content .stats-main {
  gap: 8px;
}

.stats-count {
  font-size: 1.5em;
  margin: 0;
  font-weight: bold;
  line-height: 1;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  flex-shrink: 0;
}

.vinyls-content .stats-count {
  font-size: 1.7em;
}

.thematic-content .stats-count {
  font-size: 1.5em;
}

.stats-title {
  font-size: 0.78em;
  margin: 0;
  font-weight: 600;
  opacity: 0.9;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.vinyls-content .stats-title {
  font-size: 0.8em;
}

.thematic-content .stats-title {
  font-size: 0.75em;
}

/* ========== RESPONSIVE MOBILE ========== */

@media (max-width: 1200px) {
  .stats-widget {
    padding: 12px 14px;
  }
  
  .thematic-widget {
    min-height: 65px;
  }
  
  .stats-icon {
    font-size: 1.6em;
    margin-right: 12px;
  }
  
  .thematic-content .stats-icon {
    font-size: 1.5em;
    margin-right: 10px;
  }
  
  .stats-count {
    font-size: 1.4em;
  }
  
  .thematic-content .stats-count {
    font-size: 1.3em;
  }
  
  .stats-title {
    font-size: 0.7em;
  }
}

@media (max-width: 767px) {
  .stats-widget {
    padding: 10px 12px;
    border-radius: 10px;
  }
  
  .vinyls-widget {
    min-height: 70px;
  }
  
  .thematic-widget {
    min-height: 60px;
  }
  
  .stats-icon {
    font-size: 1.4em;
    margin-right: 10px;
  }
  
  .vinyls-content .stats-icon {
    font-size: 1.6em;
  }
  
  .thematic-content .stats-icon {
    font-size: 1.3em;
    margin-right: 8px;
  }
  
  .stats-main {
    gap: 6px;
  }
  
  .stats-count {
    font-size: 1.25em;
  }
  
  .vinyls-content .stats-count {
    font-size: 1.4em;
  }
  
  .thematic-content .stats-count {
    font-size: 1.2em;
  }
  
  .stats-title {
    font-size: 0.6em;
    letter-spacing: 0.3px;
  }
  
  .vinyls-content .stats-title {
    font-size: 0.7em;
  }
  
  .thematic-content .stats-title {
    font-size: 0.55em;
  }
}

@media (max-width: 480px) {
  .stats-widget {
    padding: 8px 10px;
    border-radius: 8px;
  }
  
  .vinyls-widget {
    min-height: 60px;
  }
  
  .thematic-widget {
    min-height: 55px;
  }
  
  .stats-icon {
    font-size: 1.2em;
    margin-right: 8px;
  }
  
  .vinyls-content .stats-icon {
    font-size: 1.4em;
  }
  
  .thematic-content .stats-icon {
    font-size: 1.1em;
    margin-right: 6px;
  }
  
  .stats-main {
    gap: 4px;
  }
  
  .stats-count {
    font-size: 1.1em;
  }
  
  .vinyls-content .stats-count {
    font-size: 1.25em;
  }
  
  .thematic-content .stats-count {
    font-size: 1em;
  }
  
  .stats-title {
    font-size: 0.5em;
    letter-spacing: 0;
  }
  
  .vinyls-content .stats-title {
    font-size: 0.6em;
  }
  
  .thematic-content .stats-title {
    font-size: 0.45em;
  }
}

@media (max-width: 380px) {
  .stats-widget {
    padding: 6px 8px;
  }
  
  .thematic-widget {
    min-height: 50px;
  }
  
  .stats-icon {
    font-size: 1em;
    margin-right: 6px;
  }
  
  .thematic-content .stats-icon {
    font-size: 0.95em;
    margin-right: 5px;
  }
  
  .stats-count {
    font-size: 1em;
  }
  
  .thematic-content .stats-count {
    font-size: 0.9em;
  }
  
  .stats-title {
    font-size: 0.45em;
  }
  
  .thematic-content .stats-title {
    font-size: 0.4em;
  }
}
</style>
