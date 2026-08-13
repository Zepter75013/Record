<template>
  <Teleport to="body">
    <div v-if="isOpen" class="scanner-overlay" @click.self="handleClose">
      <div class="scanner-container">
        <div class="scanner-header">
          <h3>📷 Scanner le code-barres</h3>
          <button class="close-button" @click="handleClose" aria-label="Fermer">
            <span>&times;</span>
          </button>
        </div>

        <div class="scanner-body">
          <!-- Zone vidéo pour la caméra -->
          <div class="video-container">
            <div ref="scannerContainer" id="scanner-container"></div>
            
            <!-- Guide de cadrage -->
            <div class="scan-guide">
              <div class="guide-frame">
                <div class="corner top-left"></div>
                <div class="corner top-right"></div>
                <div class="corner bottom-left"></div>
                <div class="corner bottom-right"></div>
              </div>
            </div>
          </div>

          <!-- Message d'état -->
          <div v-if="statusMessage" class="status-message" :class="statusType">
            {{ statusMessage }}
          </div>

          <!-- Code-barres détecté -->
          <div v-if="detectedBarcode" class="detected-barcode">
            <div class="barcode-icon">✅</div>
            <div class="barcode-value">{{ detectedBarcode }}</div>
          </div>

          <!-- Instructions -->
          <div class="instructions">
            <p><strong>💡 Conseils pour un bon scan :</strong></p>
            <ul class="tips-list">
              <li>✓ Placez le code-barres dans le cadre orange</li>
              <li>✓ Maintenez une distance de 10-15 cm</li>
              <li>✓ Assurez-vous d'avoir un bon éclairage</li>
              <li>✓ Tenez l'appareil bien stable</li>
              <li>✓ Le code doit être net et lisible</li>
            </ul>
            <p class="instruction-detail">Le scan se fera automatiquement dès détection</p>
          </div>
        </div>

        <div class="scanner-actions">
          <button @click="handleClose" class="btn secondary">
            Annuler
          </button>
          <button 
            v-if="detectedBarcode" 
            @click="confirmBarcode" 
            class="btn primary"
          >
            ✓ Utiliser ce code-barres
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, watch, onBeforeUnmount, nextTick } from 'vue';
import Quagga from '@ericblade/quagga2';

const emit = defineEmits(['close', 'barcode-detected']);

const props = defineProps({
  isOpen: Boolean
});

const scannerContainer = ref(null);
const detectedBarcode = ref('');
const statusMessage = ref('');
const statusType = ref('info');
const isScanning = ref(false);

const handleClose = () => {
  stopScanning();
  emit('close');
};

const confirmBarcode = () => {
  if (detectedBarcode.value) {
    emit('barcode-detected', detectedBarcode.value);
    handleClose();
  }
};

const startScanning = async () => {
  if (isScanning.value) return;

  try {
    statusMessage.value = 'Démarrage de la caméra...';
    statusType.value = 'info';

    await nextTick();

    Quagga.init({
      inputStream: {
        name: "Live",
        type: "LiveStream",
        target: scannerContainer.value,
        constraints: {
          width: { ideal: 1280 },
          height: { ideal: 720 },
          facingMode: "environment" // Caméra arrière sur mobile
        },
      },
      locator: {
        patchSize: "medium",
        halfSample: true
      },
      numOfWorkers: 4,
      frequency: 10,
      decoder: {
        readers: [
          "ean_reader",
          "ean_8_reader",
          "code_128_reader",
          "code_39_reader",
          "upc_reader",
          "upc_e_reader"
        ],
        multiple: false
      },
      locate: true
    }, (err) => {
      if (err) {
        console.error('Erreur initialisation Quagga:', err);
        statusMessage.value = `Erreur: ${err.message}`;
        statusType.value = 'error';
        return;
      }

      console.log("✅ Quagga initialisé avec succès");
      statusMessage.value = 'Caméra prête, scannez le code-barres...';
      statusType.value = 'success';
      isScanning.value = true;

      Quagga.start();
    });

    // Événement de détection
    Quagga.onDetected((result) => {
      const code = result.codeResult.code;
      console.log("✅ Code-barres détecté:", code);
      
      detectedBarcode.value = code;
      statusMessage.value = `Code-barres détecté: ${code}`;
      statusType.value = 'success';
      
      // Arrêter le scan après détection
      stopScanning();
    });

  } catch (error) {
    console.error('Erreur démarrage scanner:', error);
    statusMessage.value = `Erreur: ${error.message}`;
    statusType.value = 'error';
    isScanning.value = false;
  }
};

const stopScanning = () => {
  if (isScanning.value) {
    Quagga.stop();
    isScanning.value = false;
  }
  detectedBarcode.value = '';
  statusMessage.value = '';
};

// Watch pour démarrer/arrêter le scan quand le modal s'ouvre/ferme
watch(() => props.isOpen, async (isOpen) => {
  if (isOpen) {
    await nextTick();
    await new Promise(resolve => setTimeout(resolve, 300)); // Délai pour DOM
    startScanning();
  } else {
    stopScanning();
  }
});

onBeforeUnmount(() => {
  stopScanning();
});
</script>

<style scoped>
.scanner-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.9);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2000;
  padding: 20px;
  animation: fadeIn 0.3s ease;
}

.scanner-container {
  background: white;
  border-radius: 16px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.scanner-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  background: #65496E;
  color: white;
  border-bottom: 3px solid #ff8c00;
}

.scanner-header h3 {
  margin: 0;
  font-size: 1.3em;
  font-weight: 600;
}

.close-button {
  background: none;
  border: none;
  font-size: 2em;
  cursor: pointer;
  color: white;
  padding: 0;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s;
}

.close-button:hover {
  background: rgba(255, 255, 255, 0.2);
}

.scanner-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px;
  overflow-y: auto;
}

.video-container {
  position: relative;
  width: 100%;
  max-width: 500px;
  aspect-ratio: 4/3;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 20px;
}

#scanner-container {
  width: 100%;
  height: 100%;
}

/* Quagga génère ses propres éléments vidéo et canvas */
#scanner-container video,
#scanner-container canvas {
  width: 100% !important;
  height: 100% !important;
  object-fit: cover;
}

.scan-guide {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
}

.guide-frame {
  position: relative;
  width: 80%;
  height: 40%;
  border: 2px solid rgba(255, 140, 0, 0.5);
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.1);
}

.corner {
  position: absolute;
  width: 30px;
  height: 30px;
  border: 3px solid #ff8c00;
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

.status-message {
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 16px;
  font-weight: 500;
  text-align: center;
  width: 100%;
  max-width: 500px;
}

.status-message.info {
  background: #e7f3ff;
  border: 1px solid #b3d9ff;
  color: #0066cc;
}

.status-message.success {
  background: #d4edda;
  border: 1px solid #c3e6cb;
  color: #155724;
}

.status-message.error {
  background: #f8d7da;
  border: 1px solid #f5c6cb;
  color: #721c24;
}

.detected-barcode {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 24px;
  background: #d4edda;
  border: 2px solid #28a745;
  border-radius: 12px;
  margin-bottom: 16px;
  width: 100%;
  max-width: 500px;
}

.barcode-icon {
  font-size: 2em;
}

.barcode-value {
  font-size: 1.2em;
  font-weight: 600;
  color: #155724;
  font-family: 'Courier New', monospace;
}

.instructions {
  text-align: center;
  color: #6c757d;
  margin-top: 8px;
  max-width: 500px;
}

.instructions p {
  margin: 8px 0;
}

.instructions strong {
  color: #495057;
}

.tips-list {
  text-align: left;
  list-style: none;
  padding: 0;
  margin: 12px 0;
  background: #f8f9fa;
  border-radius: 8px;
  padding: 16px 20px;
}

.tips-list li {
  padding: 6px 0;
  font-size: 0.9em;
  color: #495057;
}

.instruction-detail {
  font-size: 0.9em;
  color: #adb5bd;
}

.scanner-actions {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 20px;
  border-top: 1px solid #e9ecef;
}

.btn {
  flex: 1;
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 1em;
}

.btn.primary {
  background: #ff8c00;
  color: white;
}

.btn.primary:hover {
  background: #cc7000;
}

.btn.secondary {
  background: #6c757d;
  color: white;
}

.btn.secondary:hover {
  background: #5a6268;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* Responsive */
@media (max-width: 768px) {
  .scanner-overlay {
    padding: 0;
  }
  
  .scanner-container {
    width: 100%;
    max-width: 100%;
    height: 100vh;
    max-height: 100vh;
    border-radius: 0;
  }

  .video-container {
    max-width: 100%;
  }
}
</style>
