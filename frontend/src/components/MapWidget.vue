<!-- components/MapWidget.vue - VERSION RESPONSIVE MOBILE OPTIMISÉE -->
<script setup>
import { onMounted, onUnmounted, watch, ref } from 'vue';
import { useMap } from '@/composables/useMap';

const props = defineProps({
  countries: {
    type: Array,
    default: () => []
  },
  title: {
    type: String,
    default: '🌍 Carte des Pays'
  },
  mapElementId: {
    type: String,
    default: 'map-widget'
  }
});

const emit = defineEmits(['map-loaded', 'map-error']);

// Utiliser le composable de carte
const { 
  mapInitialized, 
  mapLoading, 
  mapError, 
  initMap, 
  destroyMap, 
  resizeMap 
} = useMap();

// États locaux
const countriesWithDiscs = ref(0);

// Calculer le nombre de pays avec disques
const calculateCountriesWithDiscs = () => {
  if (!props.countries || !Array.isArray(props.countries)) return 0;
  return props.countries.filter(country => (country.disc_count || 0) > 0).length;
};

// Initialiser la carte
const initializeMap = async () => {
  if (props.countries.length === 0) {
    console.warn('⚠️ MapWidget: Aucun pays à afficher');
    return;
  }
  
  await initMap(props.mapElementId, props.countries);
  
  if (mapInitialized.value) {
    emit('map-loaded', { 
      countriesCount: props.countries.length,
      countriesWithDiscs: countriesWithDiscs.value
    });
  } else if (mapError.value) {
    emit('map-error', { error: mapError.value });
  }
};

// Redimensionner la carte quand la sidebar change
const handleSidebarChange = () => {
  setTimeout(() => {
    resizeMap();
  }, 300);
};

// Watch pour les changements de pays
watch(() => props.countries, (newCountries) => {
  if (newCountries && Array.isArray(newCountries)) {
    countriesWithDiscs.value = calculateCountriesWithDiscs();
    console.log(`🗺️ ${countriesWithDiscs.value} pays avec disques`);
  }
}, { deep: true });

// Lifecycle
onMounted(async () => {
  console.log('🗺️ MapWidget monté');
  countriesWithDiscs.value = calculateCountriesWithDiscs();
  
  setTimeout(() => {
    initializeMap();
  }, 500);
});

onUnmounted(() => {
  console.log('🗺️ MapWidget démonté');
  destroyMap();
});

// Exposer des méthodes si nécessaire
defineExpose({
  refreshMap: initializeMap,
  resizeMap: handleSidebarChange
});
</script>

<template>
  <div class="map-widget">
    <div class="map-header">
      <h3>{{ title }}</h3>
      <div class="map-status">
        <span v-if="mapInitialized" class="status-badge status-success">
          ✅ {{ countriesWithDiscs }} pays
        </span>
        <span v-else-if="mapLoading" class="status-badge status-loading">
          ⏳ Chargement...
        </span>
        <span v-else-if="mapError" class="status-badge status-error">
          ❌ Erreur
        </span>
        <span v-else class="status-badge status-idle">
          🔄 Prêt
        </span>
      </div>
    </div>
    
    <div 
      :id="mapElementId" 
      class="map-container"
    ></div>
    
    <div v-if="!mapInitialized && !mapLoading" class="map-placeholder">
      <div class="placeholder-content">
        <div class="placeholder-icon">🗺️</div>
        <p>La carte sera chargée ici</p>
        <small v-if="countries.length > 0">
          {{ countries.length }} pays disponibles, {{ countriesWithDiscs }} avec des disques
        </small>
        <button 
          v-if="!mapLoading" 
          @click="initializeMap" 
          class="retry-button"
        >
          Charger la carte
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.map-widget {
  background: var(--panel-bg);
  padding: 20px;
  padding-bottom: 25px;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
  position: relative;
  margin-top: 15px;
  margin-bottom: 30px;
  border: 1px solid var(--line-soft);
  overflow: visible;
}

.map-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  flex-wrap: wrap;
  gap: 10px;
}

.map-header h3 {
  color: var(--text);
  margin: 0;
  font-weight: 700;
  font-size: 1.05em;
  border-bottom: 1px solid var(--line-soft);
  padding-bottom: 10px;
  flex-grow: 1;
}

.map-status {
  display: flex;
  align-items: center;
}

.status-badge {
  font-size: 0.85em;
  font-weight: 500;
  padding: 6px 12px;
  border-radius: 20px;
  white-space: nowrap;
}

.status-success {
  color: #27ae60;
  background: #e8f8f0;
  border: 1px solid #a3e4bc;
}

.status-loading {
  color: #f39c12;
  background: #fef9e7;
  border: 1px solid #f8c471;
}

.status-error {
  color: #e74c3c;
  background: #fdeaea;
  border: 1px solid #f5b7b1;
}

.status-idle {
  color: #3498db;
  background: #e8f4fc;
  border: 1px solid #aed6f1;
}

.map-container {
  height: 480px;
  min-height: 480px;
  width: 100%;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border: 1px solid #e1e5e9;
  position: relative;
  z-index: 0;
  background: #e8f4f8;
}

/* Permettre aux popups de sortir */
:deep(.leaflet-popup-pane) {
  z-index: 700 !important;
}

:deep(.leaflet-top),
:deep(.leaflet-bottom) {
  z-index: 500 !important;
}

.map-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(var(--tint-rgb), 0.06);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  z-index: 2;
}

.placeholder-content {
  text-align: center;
  padding: 30px;
}

.placeholder-icon {
  font-size: 3em;
  margin-bottom: 15px;
  opacity: 0.5;
}

.placeholder-content p {
  color: var(--text-soft);
  margin-bottom: 10px;
  font-weight: 500;
}

.placeholder-content small {
  color: var(--text-dim);
  display: block;
  margin-bottom: 20px;
}

.retry-button {
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  border: none;
  border-radius: 6px;
  color: white;
  cursor: pointer;
  padding: 10px 20px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.retry-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

/* Styles pour les icônes radar personnalisées */
:deep(.custom-radar-icon) {
  background: transparent !important;
  border: none !important;
}

:deep(.radar-marker) {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  filter: drop-shadow(0 4px 12px rgba(147, 51, 234, 0.5));
  cursor: pointer;
  transition: all 0.3s ease;
}

:deep(.radar-marker:hover) {
  transform: scale(1.15);
  filter: drop-shadow(0 6px 16px rgba(147, 51, 234, 0.7));
}

:deep(.radar-marker svg) {
  overflow: visible;
}

/* Styles pour les popups radar */
:deep(.radar-popup) {
  text-align: center;
  min-width: 180px;
  padding: 5px;
}

:deep(.radar-popup h4) {
  margin: 0 0 10px 0;
  color: #2c3e50;
  font-weight: 700;
  font-size: 1.1em;
}

:deep(.radar-popup p) {
  margin: 6px 0;
  font-size: 0.95em;
  color: #555;
}

/* ——— RESPONSIVE ——— */
@media (max-width: 992px) {
  .map-widget {
    padding: 16px;
    margin-top: 12px;
    margin-bottom: 20px;
  }
  
  .map-container {
    height: 400px;
    min-height: 400px;
  }
}

@media (max-width: 767px) {
  .map-widget {
    padding: 12px;
    padding-bottom: 16px;
    margin-top: 10px;
    margin-bottom: 16px;
  }
  
  .map-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    margin-bottom: 12px;
  }
  
  .map-header h3 {
    font-size: 0.95em;
    padding-bottom: 8px;
    width: 100%;
  }
  
  .status-badge {
    font-size: 0.8em;
    padding: 4px 10px;
  }
  
  .map-container {
    height: 300px;
    min-height: 300px;
    border-radius: 6px;
  }
  
  .placeholder-icon {
    font-size: 2.5em;
  }
  
  .placeholder-content {
    padding: 20px;
  }
  
  .placeholder-content p {
    font-size: 0.9em;
  }
  
  .retry-button {
    padding: 8px 16px;
    font-size: 0.9em;
  }
}

@media (max-width: 480px) {
  .map-widget {
    padding: 10px;
    margin-top: 8px;
    margin-bottom: 12px;
  }
  
  .map-header h3 {
    font-size: 0.9em;
  }
  
  .status-badge {
    font-size: 0.75em;
    padding: 3px 8px;
  }
  
  .map-container {
    height: 250px;
    min-height: 250px;
  }
  
  .placeholder-icon {
    font-size: 2em;
    margin-bottom: 10px;
  }
  
  .placeholder-content {
    padding: 15px;
  }
  
  .placeholder-content p {
    font-size: 0.85em;
    margin-bottom: 8px;
  }
  
  .placeholder-content small {
    font-size: 0.75em;
    margin-bottom: 15px;
  }
  
  :deep(.radar-popup) {
    min-width: 150px;
    padding: 3px;
  }
  
  :deep(.radar-popup h4) {
    font-size: 1em;
  }
  
  :deep(.radar-popup p) {
    font-size: 0.85em;
  }
}

@media (max-width: 360px) {
  .map-container {
    height: 220px;
    min-height: 220px;
  }
}
</style>
