<!-- frontend/src/components/FormatsModal.vue -->
<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="isOpen" class="modal-overlay" @mousedown="onOverlayMouseDown" @click="onOverlayClick">
        <div class="modal-card" :class="{ 'is-saving': saving }">
          <div class="modal-header">
            <div class="header-content">
              <h2>{{ isEditing ? '✏️ Modifier le format' : '➕ Ajouter un format' }}</h2>
              <p class="modal-subtitle">
                {{ isEditing ? 'Modifiez les informations du format de disque' : 'Ajoutez un nouveau format à votre collection' }}
              </p>
            </div>
            <button
              @click="handleClose"
              class="icon-action-btn"
              :disabled="saving"
              aria-label="Fermer"
            >
              ✕
            </button>
          </div>

          <div v-if="apiError || localApiError" class="form-error modal-error-banner">
            <span class="error-icon">⚠️</span>
            <span>{{ apiError || localApiError }}</span>
          </div>

          <form @submit.prevent="handleSave" class="format-form">
            <div class="form-section">
              <h3 class="section-title">📀 Informations du format</h3>
              <div class="form-grid">
                <label class="form-field">
                  <span class="required">Nom du format</span>
                  <div class="input-with-suggestion">
                    <input
                      id="name"
                      v-model="formData.name"
                      type="text"
                      required
                      placeholder="Ex: LP, EP, Single, CD, Cassette..."
                      :class="{ 'has-suggestion': prefillName && !formData.name, 'input-error': !isNameValid && formData.name }"
                      :disabled="saving"
                    />
                    <div v-if="prefillName && !formData.name" class="suggestion-badge">
                      <span class="suggestion-text">Suggestion : {{ prefillName }}</span>
                      <button type="button" @click="applyPrefillName" class="suggestion-button">
                        Utiliser
                      </button>
                    </div>
                  </div>
                  <div v-if="!isNameValid && formData.name" class="form-error">
                    Le nom doit contenir au moins 2 caractères
                  </div>
                </label>

                <label class="form-field form-field-full">
                  <span>Description <span class="optional-text">(optionnel)</span></span>
                  <textarea
                    id="description"
                    v-model="formData.description"
                    rows="3"
                    placeholder="Description du format, caractéristiques, capacité, notes..."
                    maxlength="500"
                    :disabled="saving"
                  ></textarea>
                  <div class="char-counter">
                    {{ formData.description?.length || 0 }}/500 caractères
                  </div>
                </label>
              </div>
            </div>

            <div v-if="isEditing && formatStats" class="form-section stats-section">
              <h3 class="section-title">📊 Statistiques</h3>
              <div class="stats-grid">
                <div class="stat-card">
                  <div class="stat-icon">💿</div>
                  <div class="stat-content">
                    <div class="stat-label">Disques dans la collection</div>
                    <div class="stat-value">{{ formatStats.discs_count || 0 }}</div>
                  </div>
                </div>
                <div class="stat-card">
                  <div class="stat-icon">📅</div>
                  <div class="stat-content">
                    <div class="stat-label">Ajouté le</div>
                    <div class="stat-date">{{ formatDate(formData.created_at) }}</div>
                  </div>
                </div>
                <div v-if="formData.updated_at && formData.updated_at !== formData.created_at" class="stat-card">
                  <div class="stat-icon">✏️</div>
                  <div class="stat-content">
                    <div class="stat-label">Modifié le</div>
                    <div class="stat-date">{{ formatDate(formData.updated_at) }}</div>
                  </div>
                </div>
              </div>
            </div>

            <div class="modal-actions">
              <button type="button" @click="handleClose" class="ghost-btn" :disabled="saving">
                Annuler
              </button>
              <button type="submit" class="primary-btn" :disabled="!isFormValid || saving">
                <span v-if="saving" class="spinner"></span>
                <span v-else>{{ isEditing ? 'Mettre à jour' : 'Enregistrer' }}</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </Transition>

    <!-- Modal de confirmation pour modifications non sauvegardées -->
    <div v-if="isUnsavedChangesModalOpen" class="modal-overlay" @click.self="cancelClose">
      <div class="modal-card">
        <div class="modal-header">
          <div class="header-title-container">
            <span class="title-icon">⚠️</span>
            <div>
              <h2>Modifications non enregistrées</h2>
              <p class="header-subtitle">
                Les modifications que vous avez effectuées seront perdues si vous continuez
              </p>
            </div>
          </div>
          <button class="icon-action-btn" @click="cancelClose" aria-label="Fermer">
            <span>&times;</span>
          </button>
        </div>
        <div class="info-section">
          <h4>📝 Que souhaitez-vous faire ?</h4>
          <p>Vous avez commencé à modifier le format. Si vous fermez maintenant, ces modifications seront perdues.</p>
        </div>
        <div class="modal-actions">
          <button @click="confirmCloseWithoutSaving" class="ghost-btn">
            ❌ Fermer sans enregistrer
          </button>
          <button @click="cancelClose" class="primary-btn">
            ↩️ Continuer l'édition
          </button>
        </div>
      </div>
    </div>

    <!-- Modal de confirmation de doublon -->
    <div v-if="isDuplicateModalOpen" class="modal-overlay" @click.self="closeDuplicateModal">
      <div class="modal-card">
        <div class="modal-header">
          <div class="header-title-container">
            <span class="title-icon">⚠️</span>
            <div>
              <h2>Format déjà existant</h2>
              <p class="header-subtitle">
                Un format avec ce nom existe déjà dans votre collection
              </p>
            </div>
          </div>
          <button class="icon-action-btn" @click="closeDuplicateModal" aria-label="Fermer">
            <span>&times;</span>
          </button>
        </div>
        <div class="info-section duplicate-warning">
          <h4>📀 Format en conflit</h4>
          <p>Le nom "<strong>{{ duplicateFormatName }}</strong>" est déjà utilisé par un format existant dans votre collection.</p>
          <p class="help-text">Pour créer ce format, vous devez choisir un nom différent.</p>
        </div>
        <div class="modal-actions">
          <button @click="closeDuplicateModal" class="primary-btn">
            ✏️ Modifier le nom
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import { useApi } from '@/composables/useApi';

const emit = defineEmits(['close', 'format-saved']);
const api = useApi()

const props = defineProps({
  isOpen: Boolean,
  formatData: Object,
  prefillName: String,
  apiError: String
});

const { get } = useApi();

// États
const saving = ref(false);
const formatStats = ref(null);
const localApiError = ref(null);
const isUnsavedChangesModalOpen = ref(false);
const isDuplicateModalOpen = ref(false);
const duplicateFormatName = ref('');
const overlayMouseDownTarget = ref(null); // Pour éviter fermeture lors de sélection texte

const formData = ref({
  id: null,
  name: '',
  description: '',
  created_at: '',
  updated_at: ''
});

// Computed
const isEditing = computed(() => !!props.formatData?.id);
const isNameValid = computed(() => formData.value.name.trim().length >= 2);
const isFormValid = computed(() => {
  return formData.value.name?.trim() && isNameValid.value;
});

// Méthodes
const loadFormatStats = async () => {
  if (!isEditing.value) return;
  
  try {
    const stats = await get(`/formats/${formData.value.id}/stats`);
    formatStats.value = stats;
  } catch (error) {
    console.error('Erreur chargement statistiques:', error);
    formatStats.value = null;
  }
};

const formatDate = (dateString) => {
  if (!dateString) return 'Non disponible';
  try {
    return new Date(dateString).toLocaleDateString('fr-FR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  } catch {
    return dateString;
  }
};

const applyPrefillName = () => {
  if (props.prefillName) {
    formData.value.name = props.prefillName;
  }
};

const handleSave = async () => {
  console.log('💿 [FormatsModal] Début handleSave, mode édition:', isEditing.value);
  console.log('💾 [FormatsModal] Sauvegarde des données:', formData.value);
  
  saving.value = true;
  localApiError.value = null;
  
  try {
    let savedFormat;
    
    // Préparer les données pour l'API
    const dataToSend = {
      name: formData.value.name.trim(),
      description: formData.value.description?.trim() || null
    };
    
    console.log('📤 [FormatsModal] Données à envoyer:', dataToSend);
    
    if (isEditing.value) {
      console.log('🔄 [FormatsModal] Mise à jour format ID:', props.formatData.id);
      const response = await api.put(`/formats/${props.formatData.id}`, dataToSend);
      console.log('🔍 [FormatsModal] Structure response PUT:', response);
      savedFormat = response.data || response;
    } else {
      console.log('➕ [FormatsModal] Création nouveau format');
      const response = await api.post('/formats', dataToSend);
      console.log('🔍 [FormatsModal] Structure response POST:', response);
      savedFormat = response.data || response;
      console.log('✅ [FormatsModal] Format créé avec ID:', savedFormat.id);
    }
    
    console.log('💿 [FormatsModal] Format sauvegardé:', savedFormat);
    
    // Émettre l'événement avec le format complet
    emit('format-saved', savedFormat);
    
    // Fermer la modale
    emit('close');
    
  } catch (error) {
    console.error('❌ [FormatsModal] Erreur lors de la sauvegarde:', error);
    console.error('❌ [FormatsModal] Stack:', error.stack);
    console.error('❌ [FormatsModal] Message:', error.message);
    console.error('❌ [FormatsModal] Response data:', error.response?.data);
    console.error('❌ [FormatsModal] Response status:', error.response?.status);
    
    // Vérifier si c'est une erreur de doublon (409 Conflict)
    const isDuplicate = error.response?.status === 409 || 
                        error.message?.includes('existe déjà') ||
                        error.message?.includes('Conflict');
    
    if (isDuplicate) {
      // Ouvrir la modal de confirmation de doublon
      duplicateFormatName.value = formData.value.name.trim();
      isDuplicateModalOpen.value = true;
    } else {
      // Pour les autres erreurs, afficher le message d'erreur
      localApiError.value = error.message || error.response?.data?.message || 'Une erreur est survenue lors de la sauvegarde';
    }
    
  } finally {
    saving.value = false;
    console.log('💿 [FormatsModal] Fin handleSave, saving:', saving.value);
  }
};

// Gestion fermeture overlay (évite fermeture lors de sélection texte)
const onOverlayMouseDown = (e) => {
  overlayMouseDownTarget.value = e.target;
};

const onOverlayClick = (e) => {
  // Ne fermer que si mousedown ET click sont sur l'overlay lui-même
  if (e.target === e.currentTarget && overlayMouseDownTarget.value === e.currentTarget) {
    handleClose();
  }
  overlayMouseDownTarget.value = null;
};

const handleClose = () => {
  console.log('💿 [FormatsModal] handleClose appelé, saving:', saving.value);
  
  if (saving.value) return;
  
  // Vérifiez si le formulaire a des modifications non sauvegardées
  const originalName = props.formatData?.name || '';
  const originalDescription = props.formatData?.description || '';
  
  const hasUnsavedChanges = (
    formData.value.name !== originalName ||
    formData.value.description !== originalDescription
  );
  
  console.log('💿 [FormatsModal] hasUnsavedChanges:', hasUnsavedChanges);
  
  if (hasUnsavedChanges) {
    isUnsavedChangesModalOpen.value = true;
    return;
  }
  
  resetForm();
  emit('close');
};

const confirmCloseWithoutSaving = () => {
  isUnsavedChangesModalOpen.value = false;
  resetForm();
  emit('close');
};

const cancelClose = () => {
  isUnsavedChangesModalOpen.value = false;
};

const closeDuplicateModal = () => {
  isDuplicateModalOpen.value = false;
  duplicateFormatName.value = '';
  // Le formulaire reste ouvert pour que l'utilisateur puisse modifier le nom
};

const resetForm = () => {
  console.log('💿 [FormatsModal] Réinitialisation du formulaire');
  formData.value = {
    id: null,
    name: '',
    description: '',
    created_at: '',
    updated_at: ''
  };
  formatStats.value = null;
  localApiError.value = null;
  isDuplicateModalOpen.value = false;
  duplicateFormatName.value = '';
};

const initializeForm = () => {
  console.log('💿 [FormatsModal] Initialisation du formulaire');
  console.log('💿 [FormatsModal] prefillName:', props.prefillName);
  console.log('💿 [FormatsModal] formatData:', props.formatData);
  
  if (props.formatData && props.formatData.id) {
    // Mode édition - formatData a priorité
    console.log('📝 [FormatsModal] Mode édition activé');
    formData.value = { 
      ...props.formatData,
      created_at: props.formatData.created_at || '',
      updated_at: props.formatData.updated_at || ''
    };
    loadFormatStats();
    console.log('📝 [FormatsModal] Données formulaire édition:', formData.value);
  } else if (props.prefillName) {
    // Mode création avec pré-remplissage
    console.log('🎯 [FormatsModal] Mode création avec pré-remplissage');
    formData.value = {
      id: null,
      name: props.prefillName,
      description: '',
      created_at: '',
      updated_at: ''
    };
    console.log('🎯 [FormatsModal] Données formulaire pré-remplies:', formData.value);
  } else if (props.formatData && props.formatData.name) {
    // Mode création avec données de pré-remplissage (sans ID)
    console.log('🎯 [FormatsModal] Mode création avec données formatData');
    formData.value = {
      id: null,
      name: props.formatData.name,
      description: props.formatData.description || '',
      created_at: '',
      updated_at: ''
    };
    console.log('🎯 [FormatsModal] Données formulaire avec formatData:', formData.value);
  } else {
    // Mode création normal
    console.log('➕ [FormatsModal] Mode création normal');
    resetForm();
  }
  
  localApiError.value = null;
  console.log('💿 [FormatsModal] Formulaire initialisé, formData:', formData.value);
};

// Watchers
watch(() => props.isOpen, (isOpen) => {
  console.log('🔄 [FormatsModal] Watcher isOpen, nouvelle valeur:', isOpen);
  if (isOpen) {
    console.log('🔄 [FormatsModal] Modale ouverte, initialisation...');
    initializeForm();
  } else {
    console.log('🔄 [FormatsModal] Modale fermée');
  }
});

watch(() => props.prefillName, (newPrefillName) => {
  console.log('🔄 [FormatsModal] Watcher prefillName, nouvelle valeur:', newPrefillName);
  if (props.isOpen && newPrefillName && !formData.value.id) {
    console.log('🔄 [FormatsModal] prefillName changé, mise à jour du nom:', newPrefillName);
    formData.value.name = newPrefillName;
  }
});

watch(() => props.formatData, (newFormatData) => {
  console.log('🔄 [FormatsModal] Watcher formatData, nouvelle valeur:', newFormatData);
  if (props.isOpen) {
    console.log('🔄 [FormatsModal] formatData changé, réinitialisation...');
    initializeForm();
  }
}, { deep: true });
</script>

<style scoped>
.modal-card.is-saving {
  pointer-events: none;
  opacity: 0.9;
}

.header-content {
  flex: 1;
}

.modal-header h2 {
  margin: 0 0 6px;
  font-size: 1.5em;
}

.modal-subtitle {
  margin: 0;
  color: var(--text-soft);
  font-size: 0.9em;
}

.modal-error-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.28);
  border-radius: 14px;
  margin-bottom: 20px;
}

.error-icon {
  font-size: 1.1em;
}

.format-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.form-section {
  background: rgba(var(--tint-rgb), 0.03);
  padding: 20px;
  border-radius: 16px;
  border: 1px solid var(--line-soft);
}

.section-title {
  margin: 0 0 18px 0;
  font-size: 1.1em;
  color: var(--text);
  padding-bottom: 10px;
  border-bottom: 1px solid var(--line-soft);
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 18px;
}

.form-field-full {
  grid-column: 1 / -1;
}

.required::after {
  content: ' *';
  color: var(--negative-text);
}

.optional-text {
  color: var(--text-dim);
  font-weight: normal;
  font-size: 0.9em;
}

.input-with-suggestion {
  position: relative;
}

.suggestion-badge {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 4px;
  padding: 8px 12px;
  background: rgba(242, 168, 120, 0.14);
  border: 1px solid rgba(242, 168, 120, 0.32);
  border-radius: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.9em;
  z-index: 5;
}

.suggestion-text {
  color: var(--accent-sand);
  font-weight: 500;
}

.suggestion-button {
  padding: 4px 12px;
  background: var(--accent-sand);
  color: #2a1608;
  border: none;
  border-radius: 8px;
  font-size: 0.85em;
  font-weight: 600;
  cursor: pointer;
}

.form-field input.has-suggestion {
  border-color: var(--accent-sand);
}

.form-field input.input-error {
  border-color: var(--negative-text);
}

.form-field input.input-error:focus {
  border-color: var(--negative-text);
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.15);
}

.char-counter {
  text-align: right;
  font-size: 0.82em;
  color: var(--text-dim);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 14px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px;
  background: var(--bg-elevated);
  border-radius: 14px;
  border: 1px solid var(--line-soft);
}

.stat-icon {
  font-size: 1.6em;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--accent), var(--accent-blue));
  color: white;
  border-radius: 50%;
  flex-shrink: 0;
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-label {
  font-size: 0.85em;
  color: var(--text-dim);
  margin-bottom: 4px;
}

.stat-value {
  font-size: 1.5em;
  font-weight: 700;
  color: var(--text);
}

.stat-date {
  font-size: 1em;
  font-weight: 600;
  color: var(--text);
}

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.35);
  border-top: 2px solid white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: all 0.3s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-from .modal-card,
.modal-fade-leave-to .modal-card {
  transform: scale(0.95) translateY(20px);
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.header-title-container {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.title-icon {
  font-size: 1.8em;
}

.header-subtitle {
  margin: 4px 0 0;
  color: var(--text-soft);
  font-size: 0.88em;
}

.info-section {
  background: rgba(242, 168, 120, 0.1);
  border: 1px solid rgba(242, 168, 120, 0.28);
  border-radius: 14px;
  padding: 18px;
  margin-bottom: 20px;
}

.info-section h4 {
  margin: 0 0 10px 0;
  color: var(--accent-sand);
  font-size: 1.05em;
}

.info-section p {
  margin: 0;
  color: var(--text-soft);
  line-height: 1.5;
}

.info-section.duplicate-warning {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.28);
}

.duplicate-warning h4 {
  color: var(--negative-text);
}

.duplicate-warning p {
  margin-bottom: 10px;
}

.duplicate-warning p.help-text {
  font-style: italic;
  font-size: 0.92em;
  margin-top: 12px;
  margin-bottom: 0;
}

@media (max-width: 480px) {
  .modal-header {
    flex-direction: column;
    gap: 12px;
  }
}
</style>
