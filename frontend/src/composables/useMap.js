// composables/useMap.js
import { ref, onMounted } from 'vue';

export function useMap() {
  const map = ref(null);
  const mapInitialized = ref(false);
  const mapLoading = ref(false);
  const mapError = ref(null);
  const resizeObserver = ref(null);

  // Charger le CSS Leaflet dynamiquement
  const loadLeafletCSS = () => {
    const leafletCssLoaded = document.querySelector('link[href*="leaflet"]');
    if (!leafletCssLoaded) {
      const link = document.createElement('link');
      link.rel = 'stylesheet';
      link.href = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css';
      link.integrity = 'sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY=';
      link.crossOrigin = '';
      document.head.appendChild(link);
      console.log('📦 Leaflet CSS chargé dynamiquement');
    }
  };

  // Observer les changements de taille du conteneur
  const setupResizeObserver = (mapElementId) => {
    if (typeof ResizeObserver === 'undefined') return;
    
    const mapElement = document.getElementById(mapElementId);
    if (!mapElement) return;
    
    // Détruire l'observateur existant
    if (resizeObserver.value) {
      resizeObserver.value.disconnect();
    }
    
    // Créer un nouvel observateur
    resizeObserver.value = new ResizeObserver(() => {
      if (map.value) {
        console.log('📏 Redimensionnement détecté, ajustement de la carte...');
        setTimeout(() => {
          forceMapResize();
        }, 50);
      }
    });
    
    resizeObserver.value.observe(mapElement);
  };

  // Forcer le redimensionnement de la carte
  const forceMapResize = () => {
    if (!map.value) return;
    
    const mapElement = map.value.getContainer();
    if (!mapElement) return;
    
    console.log('🗺️ Forcer le redimensionnement de la carte');
    
    // Technique 1: invalidateSize avec plusieurs tentatives
    map.value.invalidateSize(false);
    
    // Technique 2: Forcer un recalcul de style
    void mapElement.offsetWidth;
    
    // Technique 3: Petite animation pour déclencher le rendu
    const currentCenter = map.value.getCenter();
    const currentZoom = map.value.getZoom();
    
    setTimeout(() => {
      if (map.value) {
        map.value.setView(currentCenter, currentZoom - 0.001, { animate: false });
        setTimeout(() => {
          if (map.value) {
            map.value.setView(currentCenter, currentZoom, { animate: false });
          }
        }, 10);
      }
    }, 100);
    
    // Technique 4: Recalcul après un délai
    setTimeout(() => {
      if (map.value) {
        map.value.invalidateSize(true);
      }
    }, 300);
  };

  // Attendre que Leaflet soit disponible
  const waitForLeaflet = (maxAttempts = 10, interval = 300) => {
    return new Promise((resolve, reject) => {
      let attempts = 0;
      
      const checkLeaflet = () => {
        attempts++;
        console.log(`🗺️ Tentative ${attempts}/${maxAttempts} - Leaflet disponible:`, typeof window.L !== 'undefined');
        
        if (typeof window.L !== 'undefined') {
          console.log('✅ Leaflet trouvé!');
          resolve(true);
        } else if (attempts >= maxAttempts) {
          reject(new Error('Leaflet non disponible après plusieurs tentatives'));
        } else {
          setTimeout(checkLeaflet, interval);
        }
      };
      
      checkLeaflet();
    });
  };

  // Créer une icône radar
  const createRadarIconHTML = (count) => {
    const size = Math.max(70, Math.min(120, 70 + count * 5));
    const fontSize = Math.max(22, Math.min(36, 22 + count * 1.2));
    
    return `
<div class="radar-marker" style="width:${size}px;height:${size}px;">
  <svg width="100%" height="100%" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
    <circle cx="50" cy="50" r="45" fill="#B86684" opacity="0.2"><animate attributeName="opacity" values="0.2;0.1;0.2" dur="2s" repeatCount="indefinite"/></circle>
    <circle cx="50" cy="50" r="35" fill="#B86684" opacity="0.35"><animate attributeName="opacity" values="0.35;0.2;0.35" dur="2s" repeatCount="indefinite"/></circle>
    <circle cx="50" cy="50" r="25" fill="#B86684" opacity="0.55"><animate attributeName="opacity" values="0.55;0.35;0.55" dur="2s" repeatCount="indefinite"/></circle>
    <circle cx="50" cy="50" r="15" fill="#B86684" opacity="0.75"><animate attributeName="opacity" values="0.75;0.55;0.75" dur="2s" repeatCount="indefinite"/></circle>
    <circle cx="50" cy="50" r="8" fill="#B86684" opacity="0.95"><animate attributeName="opacity" values="0.95;0.75;0.95" dur="2s" repeatCount="indefinite"/></circle>
    <text x="50" y="50" text-anchor="middle" dominant-baseline="central"
          fill="white" font-family="Arial" font-weight="bold" font-size="${fontSize}"
          style="text-shadow:0 2px 4px rgba(0,0,0,0.7);">
      ${count}
    </text>
  </svg>
</div>`;
  };

  // Coordonnées par défaut pour les pays
  const defaultCoords = {
    'FR': [48.8566, 2.3522],
    'DE': [52.5200, 13.4050],
    'GB': [51.5074, -0.1278],
    'US': [38.8951, -77.0364],
    'JP': [35.6895, 139.6917],
    'ES': [40.4168, -3.7038],
    'IT': [41.9028, 12.4964],
    'CA': [45.4215, -75.6972],
    'AU': [-33.8688, 151.2093],
    'BR': [-15.7939, -47.8828],
    'CZ': [50.0755, 14.4378],
    'NL': [52.3676, 4.9041],
    'EU': [50.8505, 4.3488],
    'WW': [20, 0],
    'XX': [0, 0]
  };

  // Initialiser la carte
  const initMap = async (mapElementId, countries) => {
    console.log('🗺️ Début initialisation carte...');
    
    if (mapInitialized.value && map.value) {
      console.log('🗺️ Carte déjà initialisée, mise à jour...');
      updateMapData(countries);
      return;
    }

    mapLoading.value = true;
    mapError.value = null;

    try {
      loadLeafletCSS();
      await waitForLeaflet();
      
      const mapElement = document.getElementById(mapElementId);
      if (!mapElement) {
        throw new Error(`Élément #${mapElementId} non trouvé`);
      }

      console.log('🗺️ Élément trouvé, dimensions:', mapElement.offsetWidth, 'x', mapElement.offsetHeight);

      // Nettoyer l'ancienne carte
      if (map.value) {
        map.value.remove();
        map.value = null;
      }

      // Créer la carte avec des options optimisées
      map.value = window.L.map(mapElement, {
        center: [48.8566, 2.3522],
        zoom: 4,
        scrollWheelZoom: true,
        zoomControl: true,
        fadeAnimation: false,
        zoomAnimation: false,
        markerZoomAnimation: false,
        transform3DLimit: 0, // Désactiver les transformations 3D problématiques
        preferCanvas: true // Utiliser canvas pour de meilleures performances
      });

      // Ajouter les tuiles
      window.L.tileLayer('https://server.arcgisonline.com/ArcGIS/rest/services/World_Physical_Map/MapServer/tile/{z}/{y}/{x}', {
        attribution: 'Tiles &copy; Esri',
        maxZoom: 8,
        minZoom: 2,
        updateWhenIdle: false, // Mettre à jour immédiatement
        updateWhenZooming: false
      }).addTo(map.value);

      // Ajouter les marqueurs
      addMarkersToMap(countries);

      mapInitialized.value = true;
      
      // Configurer l'observateur de redimensionnement
      setupResizeObserver(mapElementId);
      
      // Forcer un redimensionnement initial
      setTimeout(() => {
        forceMapResize();
      }, 500);

      console.log('✅ Carte initialisée avec succès');

    } catch (err) {
      console.error('❌ Erreur lors de l\'initialisation de la carte:', err);
      mapError.value = err.message;
      mapInitialized.value = false;
    } finally {
      mapLoading.value = false;
    }
  };

  // Ajouter des marqueurs à la carte
  const addMarkersToMap = (countries) => {
    if (!map.value || !countries || countries.length === 0) return;
    
    const countriesWithDiscs = countries.filter(country => (country.disc_count || 0) > 0);
    console.log(`🗺️ Ajout de ${countriesWithDiscs.length} marqueurs`);
    
    countriesWithDiscs.forEach(country => {
      const coords = defaultCoords[country.code] || 
                    [country.latitude || 48.8566, country.longitude || 2.3522];
      const discCount = country.disc_count;
      const size = Math.max(70, Math.min(120, 70 + discCount * 5));
      
      const radarIcon = window.L.divIcon({
        className: 'custom-radar-icon',
        html: createRadarIconHTML(discCount),
        iconSize: [size, size],
        iconAnchor: [size / 2, size / 2]
      });

      window.L.marker(coords, { icon: radarIcon }).addTo(map.value)
        .bindPopup(`
          <div class="radar-popup">
            <h4>${country.name}</h4>
            <p><strong>📍 Code:</strong> ${country.code}</p>
            <p><strong>💿 Disques:</strong> ${discCount}</p>
          </div>
        `);
    });
  };

  // Mettre à jour les données de la carte
  const updateMapData = (countries) => {
    if (!map.value) return;
    
    // Supprimer tous les marqueurs existants
    map.value.eachLayer((layer) => {
      if (layer instanceof window.L.Marker) {
        map.value.removeLayer(layer);
      }
    });
    
    // Ajouter les nouveaux marqueurs
    addMarkersToMap(countries);
    forceMapResize();
  };

  // Détruire la carte
  const destroyMap = () => {
    if (resizeObserver.value) {
      resizeObserver.value.disconnect();
      resizeObserver.value = null;
    }
    
    if (map.value) {
      map.value.remove();
      map.value = null;
      mapInitialized.value = false;
      console.log('🗺️ Carte détruite');
    }
  };

  // Redimensionner la carte
  const resizeMap = () => {
    if (map.value) {
      console.log('🗺️ Redimensionnement manuel de la carte');
      forceMapResize();
    }
  };

  return {
    map,
    mapInitialized,
    mapLoading,
    mapError,
    initMap,
    destroyMap,
    resizeMap
  };
}