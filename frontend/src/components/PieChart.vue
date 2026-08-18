<!-- components/PieChart.vue - VERSION RESPONSIVE MOBILE OPTIMISÉE -->
<script setup>
import { computed } from 'vue';

const props = defineProps({
  data: {
    type: Array,
    default: () => []
  },
  colors: {
    type: Array,
    default: () => []
  },
  total: {
    type: Number,
    default: 0
  },
  title: {
    type: String,
    default: 'Répartition'
  },
  type: {
    type: String,
    default: 'genres' // genres, formats, artists
  },
  unitLabel: {
    type: String,
    default: 'disques'
  }
});

const emit = defineEmits(['slice-hover', 'mouseleave']);

// Calculer les données du camembert
const pieData = computed(() => {
  if (!props.data.length || props.total === 0) {
    return [];
  }
  
  let cumulativePercentage = 0;
  return props.data.map((item, index) => {
    const percentage = (item.count / props.total) * 100;
    const startAngle = (cumulativePercentage / 100) * 360;
    const endAngle = ((cumulativePercentage + percentage) / 100) * 360;
    cumulativePercentage += percentage;
    
    return {
      ...item,
      percentage: percentage.toFixed(1),
      color: props.colors[index % props.colors.length]?.solid || '#667eea',
      gradientId: `grad-${props.type}-${(index % 8) + 1}`,
      startAngle,
      endAngle
    };
  });
});

// Limiter le nombre d'éléments dans la légende (5 max)
const maxLegendItems = 5;

const limitedLegendData = computed(() => {
  return pieData.value.slice(0, maxLegendItems);
});

// Nombre d'éléments restants
const remainingCount = computed(() => {
  return pieData.value.length - maxLegendItems;
});

// Générer le chemin SVG d'une part
const getSlicePath = (startAngle, endAngle, radius = 100, centerX = 110, centerY = 110) => {
  const startRad = (startAngle - 90) * Math.PI / 180;
  const endRad = (endAngle - 90) * Math.PI / 180;
  const x1 = centerX + radius * Math.cos(startRad);
  const y1 = centerY + radius * Math.sin(startRad);
  const x2 = centerX + radius * Math.cos(endRad);
  const y2 = centerY + radius * Math.sin(endRad);
  const largeArc = endAngle - startAngle > 180 ? 1 : 0;
  
  return `M ${centerX} ${centerY} L ${x1} ${y1} A ${radius} ${radius} 0 ${largeArc} 1 ${x2} ${y2} Z`;
};

// Gestion des événements hover
const handleSliceHover = (event, slice) => {
  emit('slice-hover', {
    event,
    name: slice.name,
    count: slice.count,
    percentage: slice.percentage
  });
};

// Gestion du mouseleave
const handleMouseLeave = () => {
  emit('mouseleave');
};
</script>

<template>
  <div 
    class="pie-chart-widget"
    @mouseleave="handleMouseLeave"
  >
    <h3 class="widget-title">{{ title }}</h3>
    <div v-if="pieData.length === 0" class="placeholder-text">
      Aucune donnée disponible
    </div>
    <div v-else class="pie-chart-container">
      <!-- Camembert -->
      <div class="pie-svg-container">
        <svg 
          viewBox="0 0 220 220" 
          class="pie-chart-svg"
          @mouseleave="handleMouseLeave"
        >
          <defs>
            <linearGradient 
              v-for="i in 8" 
              :key="i" 
              :id="`grad-${type}-${i}`" 
              x1="0%" y1="0%" x2="100%" y2="100%"
            >
              <stop offset="0%" :style="{ 'stop-color': colors[i-1]?.solid || '#667eea', 'stop-opacity': 1 }" />
              <stop offset="100%" :style="{ 'stop-color': colors[(i+2) % colors.length]?.solid || '#764ba2', 'stop-opacity': 1 }" />
            </linearGradient>
            <filter :id="`shadow-${type}`" x="-50%" y="-50%" width="200%" height="200%">
              <feDropShadow dx="0" dy="4" stdDeviation="6" flood-opacity="0.25"/>
            </filter>
          </defs>
          <g v-for="(slice, index) in pieData" :key="index">
            <path
              :d="getSlicePath(slice.startAngle, slice.endAngle)"
              :fill="`url(#grad-${type}-${(index % 8) + 1})`"
              class="pie-slice"
              :filter="`url(#shadow-${type})`"
              @mouseenter="(event) => handleSliceHover(event, slice)"
              @mousemove="(event) => handleSliceHover(event, slice)"
            />
          </g>
          <circle cx="110" cy="110" r="32" fill="white" class="pie-center" :filter="`url(#shadow-${type})`"/>
          <text x="110" y="105" text-anchor="middle" class="pie-center-text">
            {{ total }}
          </text>
          <text x="110" y="122" text-anchor="middle" class="pie-center-subtext">
            {{ unitLabel }}
          </text>
        </svg>
      </div>
      
      <!-- Légende compacte -->
      <div 
        class="legend-container"
        @mouseleave="handleMouseLeave"
      >
        <div class="legend-scroll" :class="`${type}-legend`">
          <div 
            v-for="(slice, index) in limitedLegendData" 
            :key="index"
            class="legend-item"
            @mouseenter="(event) => handleSliceHover(event, slice)"
          >
            <span class="legend-color" :style="{ background: slice.color }"></span>
            <span class="legend-name">{{ slice.name }}</span>
            <span class="legend-count">{{ slice.count }}</span>
          </div>
        </div>
        <div v-if="remainingCount > 0" class="legend-more">
          + {{ remainingCount }} autre{{ remainingCount > 1 ? 's' : '' }}
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.pie-chart-widget {
  background: var(--panel-bg);
  padding: 16px;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
  border: 1px solid var(--line-soft);
  transition: all 0.3s ease;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  min-height: 0; /* Important pour flexbox */
}

.pie-chart-widget:hover {
  transform: translateY(-2px);
}

.widget-title {
  color: var(--text);
  margin: 0 0 12px 0;
  font-weight: 700;
  font-size: 0.95em;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--line-soft);
  flex-shrink: 0;
}

.placeholder-text {
  color: var(--text-dim);
  font-style: italic;
  text-align: center;
  margin: 30px 0;
  font-size: 0.95em;
}

.pie-chart-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  flex: 1;
  min-height: 0;
}

/* Camembert */
.pie-svg-container {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-shrink: 0;
}

.pie-chart-svg {
  width: 180px;
  height: 180px;
  flex-shrink: 0;
  filter: drop-shadow(0 6px 12px rgba(0, 0, 0, 0.15));
}

.pie-slice {
  transition: all 0.3s ease;
  cursor: pointer;
  animation: sliceGrow 0.8s cubic-bezier(0.4, 0, 0.2, 1) forwards;
  transform-origin: center;
}

.pie-slice:hover {
  opacity: 0.85;
  filter: brightness(1.15) drop-shadow(0 4px 8px rgba(0, 0, 0, 0.3));
  transform: scale(1.05);
}

@keyframes sliceGrow {
  from {
    opacity: 0;
    transform: scale(0.3);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.pie-center-text {
  font-size: 22px;
  font-weight: 700;
  fill: #2d3748;
}

.pie-center-subtext {
  font-size: 9px;
  font-weight: 600;
  fill: #718096;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Légende ultra compacte */
.legend-container {
  width: 100%;
  flex-shrink: 0;
}

.legend-scroll {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: center;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(var(--tint-rgb), 0.03);
  border: 1px solid var(--line-soft);
  border-radius: 6px;
  padding: 5px 8px;
  font-size: 0.75em;
  transition: all 0.2s ease;
  cursor: pointer;
  max-width: 140px;
}

.legend-item:hover {
  background: rgba(var(--tint-rgb), 0.07);
  transform: translateY(-1px);
}

.legend-color {
  width: 10px;
  height: 10px;
  border-radius: 3px;
  flex-shrink: 0;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.15);
}

.legend-name {
  font-weight: 600;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  min-width: 0;
}

.legend-count {
  font-weight: 700;
  color: var(--text-soft);
  background: transparent;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.9em;
  flex-shrink: 0;
}

.legend-more {
  text-align: center;
  color: var(--text-dim);
  font-size: 0.75em;
  font-style: italic;
  margin-top: 8px;
  padding: 4px 8px;
  background: transparent;
  border-radius: 4px;
}

/* ——— RESPONSIVE ——— */
@media (max-width: 1400px) {
  .pie-chart-svg {
    width: 200px;
    height: 200px;
  }
  
  .pie-center-text {
    font-size: 24px;
  }
}

@media (max-width: 992px) {
  .pie-chart-widget {
    padding: 14px;
  }
  
  .pie-chart-container {
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
    gap: 20px;
  }
  
  .pie-svg-container {
    flex-shrink: 0;
  }
  
  .pie-chart-svg {
    width: 160px;
    height: 160px;
  }
  
  .pie-center-text {
    font-size: 20px;
  }
  
  .pie-center-subtext {
    font-size: 8px;
  }
  
  .legend-container {
    flex: 1;
    min-width: 0;
    overflow: hidden;
  }
  
  .legend-scroll {
    flex-direction: column;
    flex-wrap: nowrap;
    justify-content: flex-start;
    max-height: 180px;
    overflow-y: auto;
    overflow-x: hidden;
    gap: 5px;
  }
  
  .legend-item {
    max-width: 100%;
    width: 100%;
    flex: 0 0 auto;
    box-sizing: border-box;
  }
}

@media (max-width: 767px) {
  .pie-chart-widget {
    padding: 12px;
  }
  
  .widget-title {
    font-size: 0.9em;
    margin-bottom: 10px;
    padding-bottom: 6px;
  }
  
  .pie-chart-container {
    flex-direction: row;
    gap: 12px;
  }
  
  .pie-svg-container {
    flex-shrink: 0;
  }
  
  .pie-chart-svg {
    width: 120px;
    height: 120px;
  }
  
  .pie-center-text {
    font-size: 16px;
  }
  
  .pie-center-subtext {
    font-size: 7px;
  }
  
  .legend-container {
    flex: 1;
    min-width: 0;
    overflow: hidden;
  }
  
  .legend-scroll {
    flex-direction: column;
    flex-wrap: nowrap;  /* Force une seule colonne */
    gap: 4px;
    max-height: 160px;
    overflow-y: auto;
    overflow-x: hidden;
    align-items: stretch;
  }
  
  .legend-item {
    padding: 4px 6px;
    font-size: 0.7em;
    max-width: 100%;  /* Prend toute la largeur */
    width: 100%;
    box-sizing: border-box;
  }
  
  .legend-color {
    width: 8px;
    height: 8px;
  }
  
  .legend-name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .legend-count {
    padding: 1px 4px;
    font-size: 0.85em;
  }
  
  .legend-more {
    font-size: 0.7em;
    margin-top: 4px;
    padding: 3px 6px;
  }
}

@media (max-width: 480px) {
  .pie-chart-svg {
    width: 90px;
    height: 90px;
  }
  
  .pie-center-text {
    font-size: 13px;
  }
  
  .pie-center-subtext {
    font-size: 5px;
  }
  
  .pie-chart-container {
    gap: 10px;
  }
  
  .legend-scroll {
    max-height: 140px;
    gap: 3px;
  }
  
  .legend-item {
    font-size: 0.65em;
    padding: 3px 5px;
    gap: 4px;
  }
  
  .legend-color {
    width: 7px;
    height: 7px;
  }
  
  .legend-count {
    padding: 1px 3px;
    font-size: 0.8em;
  }
}
</style>
