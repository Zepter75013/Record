<!-- frontend/src/components/BarcodeScanner.vue -->
<template>
  <Teleport to="body">
    <div v-if="isOpen" class="scanner-modal-overlay" @click.self="handleClose">
      <div class="scanner-modal">
        <!-- En-tête -->
        <div class="scanner-header">
          <h3>📷 Scanner un code-barres</h3>
          <button class="close-btn" @click="handleClose">✕</button>
        </div>
        
        <!-- Zone vidéo -->
        <div class="scanner-body">
          <div class="video-container" ref="scannerContainer">
            <!-- Quagga injecte la vidéo ici -->
            
            <!-- Overlay de cadrage -->
            <div class="scanner-overlay">
              <div class="scanner-frame">
                <div class="corner top-left"></div>
                <div class="corner top-right"></div>
                <div class="corner bottom-left"></div>
                <div class="corner bottom-right"></div>
              </div>
              <p class="scanner-instruction">
                Placez le code-barres dans le cadre
              </p>
            </div>
          </div>
          
          <!-- État du scan -->
          <div class="scan-status">
            <div v-if="initializing" class="status-initializing">
              <div class="spinner"></div>
              <p>⏳ Initialisation de la caméra...</p>
            </div>
            
            <div v-else-if="scanning" class="status-scanning">
              <div class="scanner-line"></div>
              <p>🔍 Recherche de code-barres...</p>
            </div>
            
            <div v-else-if="detectedCode" class="status-detected">
              <p>✅ Code détecté: <strong>{{ detectedCode }}</strong></p>
            </div>
            
            <div v-else-if="error" class="status-error">
              <p>⚠️ {{ error }}</p>
            </div>
            
            <div v-else class="status-waiting">
              <p>Prêt à scanner</p>
            </div>
          </div>
          
          <!-- Sélection caméra -->
          <div v-if="cameras.length > 1" class="camera-selector">
            <label for="camera-select">Caméra:</label>
            <select
              id="camera-select"
              v-model="selectedCamera"
              @change="switchCamera"
              class="camera-select"
            >
              <option
                v-for="camera in cameras"
                :key="camera.deviceId"
                :value="camera.deviceId"
              >
                {{ camera.label || `Caméra ${cameras.indexOf(camera) + 1}` }}
              </option>
            </select>
          </div>
        </div>
        
        <!-- Actions -->
        <div class="scanner-footer">
          <button @click="handleClose" class="btn-secondary">
            Annuler
          </button>
          <button
            v-if="detectedCode"
            @click="handleUseCode"
            class="btn-primary"
          >
            Utiliser ce code
          </button>
          <button
            v-if="detectedCode"
            @click="rescan"
            class="btn-secondary"
          >
            🔄 Rescanner
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, onUnmounted, watch, nextTick } from 'vue'
import Quagga from '@ericblade/quagga2'

const props = defineProps({
  isOpen: {
    type: Boolean,
    required: true
  }
})

const emit = defineEmits(['close', 'barcodeDetected'])

// Refs
const scannerContainer = ref(null)
const scanning = ref(false)
const initializing = ref(false)
const detectedCode = ref(null)
const error = ref(null)
const cameras = ref([])
const selectedCamera = ref(null)
const quaggaInitialized = ref(false)

// Configuration Quagga
const getQuaggaConfig = (deviceId = null) => ({
  inputStream: {
    name: 'Live',
    type: 'LiveStream',
    target: scannerContainer.value,
    constraints: {
      width: { min: 640, ideal: 1280, max: 1920 },
      height: { min: 480, ideal: 720, max: 1080 },
      facingMode: 'environment',
      ...(deviceId && { deviceId: { exact: deviceId } })
    }
  },
  locator: {
    patchSize: 'medium',
    halfSample: true
  },
  numOfWorkers: navigator.hardwareConcurrency || 4,
  frequency: 10,
  decoder: {
    readers: [
      'ean_reader',        // EAN-13 (codes-barres vinyles EU)
      'ean_8_reader',      // EAN-8
      'upc_reader',        // UPC-A (codes-barres vinyles US)
      'upc_e_reader',      // UPC-E
      'code_128_reader'    // Code 128 (parfois utilisé)
    ]
  },
  locate: true
})

// Lister les caméras disponibles
const listCameras = async () => {
  try {
    console.log('📹 Listage des caméras...')
    
    if (!navigator.mediaDevices || !navigator.mediaDevices.enumerateDevices) {
      console.warn('⚠️ API mediaDevices non disponible')
      error.value = 'Votre navigateur ne supporte pas l\'accès à la caméra.'
      return false
    }
    
    // Demander la permission d'abord
    try {
      const tempStream = await navigator.mediaDevices.getUserMedia({ video: true })
      tempStream.getTracks().forEach(track => track.stop())
    } catch (permErr) {
      console.log('⚠️ Pas d\'accès caméra préalable:', permErr.name, permErr.message)
      
      if (permErr.name === 'NotFoundError') {
        error.value = 'Aucune caméra détectée sur cet appareil.'
        return false
      }
      
      if (permErr.name === 'NotAllowedError') {
        error.value = 'Accès à la caméra refusé. Veuillez autoriser l\'accès dans les paramètres du navigateur.'
        return false
      }
    }
    
    const devices = await navigator.mediaDevices.enumerateDevices()
    cameras.value = devices.filter(device => device.kind === 'videoinput')
    
    console.log(`📷 ${cameras.value.length} caméra(s) trouvée(s)`)
    
    if (cameras.value.length === 0) {
      error.value = 'Aucune caméra détectée sur cet appareil.'
      return false
    }
    
    if (!selectedCamera.value) {
      const backCamera = cameras.value.find(camera =>
        camera.label.toLowerCase().includes('back') ||
        camera.label.toLowerCase().includes('rear') ||
        camera.label.toLowerCase().includes('arrière')
      )
      selectedCamera.value = backCamera?.deviceId || cameras.value[0].deviceId
    }
    
    return true
  } catch (err) {
    console.error('❌ Erreur listage caméras:', err)
    error.value = 'Erreur lors de la détection des caméras.'
    return false
  }
}

// Initialiser Quagga
const initQuagga = async (deviceId = null) => {
  try {
    initializing.value = true
    error.value = null
    
    // Arrêter l'instance précédente si existe
    await stopQuagga()
    
    await nextTick()
    
    if (!scannerContainer.value) {
      error.value = 'Conteneur vidéo non disponible.'
      return
    }
    
    const config = getQuaggaConfig(deviceId)
    
    console.log('🎥 Initialisation Quagga...')
    
    await new Promise((resolve, reject) => {
      Quagga.init(config, (err) => {
        if (err) {
          console.error('❌ Erreur init Quagga:', err)
          reject(err)
        } else {
          resolve()
        }
      })
    })
    
    // Configurer le callback de détection
    Quagga.onDetected(onBarcodeDetected)
    
    // Démarrer le scan
    Quagga.start()
    quaggaInitialized.value = true
    scanning.value = true
    initializing.value = false
    
    console.log('✅ Quagga démarré')
    
  } catch (err) {
    console.error('❌ Erreur initialisation scanner:', err)
    initializing.value = false
    
    if (err.name === 'NotAllowedError') {
      error.value = 'Accès à la caméra refusé.'
    } else if (err.name === 'NotFoundError') {
      error.value = 'Aucune caméra trouvée.'
    } else if (err.name === 'NotReadableError') {
      error.value = 'La caméra est utilisée par une autre application.'
    } else {
      error.value = 'Impossible d\'initialiser le scanner.'
    }
  }
}

// Callback de détection
const onBarcodeDetected = (result) => {
  const code = result.codeResult.code
  
  // Vérifier la fiabilité (éviter les faux positifs)
  const errors = result.codeResult.decodedCodes
    .filter(x => x.error !== undefined)
    .map(x => x.error)
  
  const avgError = errors.reduce((a, b) => a + b, 0) / errors.length
  
  console.log(`📊 Code détecté: ${code}, erreur moyenne: ${avgError.toFixed(4)}`)
  
  // Accepter seulement si l'erreur est faible
  if (avgError < 0.1) {
    console.log(`✅ Code validé: ${code}`)
    detectedCode.value = code
    scanning.value = false
    
    // Arrêter le scan mais garder la vidéo
    Quagga.offDetected(onBarcodeDetected)
    Quagga.pause()
  }
}

// Arrêter Quagga
const stopQuagga = async () => {
  if (quaggaInitialized.value) {
    try {
      Quagga.offDetected(onBarcodeDetected)
      Quagga.stop()
      quaggaInitialized.value = false
      console.log('🛑 Quagga arrêté')
    } catch (err) {
      console.warn('Erreur arrêt Quagga:', err)
    }
  }
  scanning.value = false
}

// Changer de caméra
const switchCamera = async () => {
  if (selectedCamera.value) {
    await initQuagga(selectedCamera.value)
  }
}

// Rescanner
const rescan = () => {
  detectedCode.value = null
  Quagga.onDetected(onBarcodeDetected)
  Quagga.start()
  scanning.value = true
}

// Utiliser le code détecté
const handleUseCode = () => {
  emit('barcodeDetected', detectedCode.value)
  handleClose()
}

// Fermer le modal
const handleClose = async () => {
  await stopQuagga()
  detectedCode.value = null
  error.value = null
  initializing.value = false
  emit('close')
}

// Lifecycle
watch(() => props.isOpen, async (newValue) => {
  if (newValue) {
    error.value = null
    detectedCode.value = null
    
    await nextTick()
    
    const camerasAvailable = await listCameras()
    
    if (camerasAvailable) {
      await initQuagga(selectedCamera.value)
    }
  } else {
    await stopQuagga()
  }
})

onUnmounted(async () => {
  await stopQuagga()
})
</script>

<style scoped>
.scanner-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 20px;
}

.scanner-modal {
  background: white;
  border-radius: 16px;
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.scanner-header {
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.scanner-header h3 {
  margin: 0;
  font-size: 1.3em;
  color: #2c3e50;
}

.close-btn {
  width: 36px;
  height: 36px;
  border: none;
  background: #f5f5f5;
  border-radius: 50%;
  cursor: pointer;
  font-size: 1.5em;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #e74c3c;
  color: white;
  transform: rotate(90deg);
}

.scanner-body {
  flex: 1;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
}

.video-container {
  position: relative;
  width: 100%;
  aspect-ratio: 4/3;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
}

/* Style pour la vidéo injectée par Quagga */
.video-container :deep(video) {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-container :deep(canvas) {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.scanner-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  pointer-events: none;
  z-index: 10;
}

.scanner-frame {
  position: relative;
  width: 80%;
  max-width: 300px;
  aspect-ratio: 3/1;
  border: 2px solid rgba(255, 255, 255, 0.5);
  border-radius: 8px;
}

.corner {
  position: absolute;
  width: 30px;
  height: 30px;
  border: 3px solid #2ecc71;
}

.corner.top-left {
  top: -3px;
  left: -3px;
  border-right: none;
  border-bottom: none;
  border-radius: 8px 0 0 0;
}

.corner.top-right {
  top: -3px;
  right: -3px;
  border-left: none;
  border-bottom: none;
  border-radius: 0 8px 0 0;
}

.corner.bottom-left {
  bottom: -3px;
  left: -3px;
  border-right: none;
  border-top: none;
  border-radius: 0 0 0 8px;
}

.corner.bottom-right {
  bottom: -3px;
  right: -3px;
  border-left: none;
  border-top: none;
  border-radius: 0 0 8px 0;
}

.scanner-instruction {
  margin-top: 20px;
  color: white;
  background: rgba(0, 0, 0, 0.7);
  padding: 10px 20px;
  border-radius: 20px;
  font-size: 0.9em;
}

.scan-status {
  padding: 16px;
  border-radius: 8px;
  text-align: center;
}

.status-initializing {
  background: #fff3e0;
  border: 1px solid #ffcc80;
  color: #e65100;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid #ffcc80;
  border-top-color: #e65100;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.status-scanning {
  background: #e7f3ff;
  border: 1px solid #b3d9ff;
  color: #0066cc;
  position: relative;
  overflow: hidden;
}

.scanner-line {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: #3498db;
  animation: scan 2s linear infinite;
}

@keyframes scan {
  0% { transform: translateY(0); }
  100% { transform: translateY(80px); }
}

.status-detected {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  color: #15803d;
}

.status-detected strong {
  font-family: monospace;
  font-size: 1.1em;
}

.status-error {
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #dc2626;
}

.status-waiting {
  background: #f8f9fa;
  border: 1px solid #dee2e6;
  color: #6c757d;
}

.camera-selector {
  display: flex;
  align-items: center;
  gap: 12px;
}

.camera-selector label {
  font-weight: 500;
  color: #495057;
}

.camera-select {
  flex: 1;
  padding: 8px 32px 8px 12px;
  border: 2px solid #ddd;
  border-radius: 8px;
  font-size: 0.9em;
  cursor: pointer;
  -webkit-appearance: none;
  -moz-appearance: none;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23495057' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  background-size: 14px;
}

.camera-select:focus {
  outline: none;
  border-color: #3498db;
}

.scanner-footer {
  padding: 20px;
  border-top: 1px solid #e0e0e0;
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.btn-secondary,
.btn-primary {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  font-size: 0.95em;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary {
  background: #ecf0f1;
  color: #34495e;
}

.btn-secondary:hover {
  background: #bdc3c7;
}

.btn-primary {
  background: #2ecc71;
  color: white;
}

.btn-primary:hover {
  background: #27ae60;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(46, 204, 113, 0.3);
}

/* Responsive */
@media (max-width: 768px) {
  .scanner-modal-overlay {
    padding: 10px;
  }
  
  .scanner-modal {
    max-height: 95vh;
  }
  
  .scanner-header {
    padding: 16px;
  }
  
  .scanner-body {
    padding: 16px;
  }
  
  .scanner-footer {
    padding: 16px;
    flex-direction: column;
  }
  
  .btn-secondary,
  .btn-primary {
    width: 100%;
  }
}
</style>
