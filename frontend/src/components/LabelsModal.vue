<!-- frontend/src/components/LabelsModal.vue -->
<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="isOpen" class="modal-overlay" @mousedown="onOverlayMouseDown" @click="onOverlayClick">
        <div class="modal-card" :class="{ 'is-saving': saving }">
          <div class="modal-header">
            <div class="header-content">
              <h2>{{ isEditing ? '✏️ Modifier le label' : '➕ Ajouter un label' }}</h2>
              <p class="modal-subtitle">
                {{ isEditing ? 'Modifiez les informations du label musical' : 'Ajoutez un nouveau label à votre collection' }}
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

          <form @submit.prevent="handleSave" class="label-form">
            <div class="form-section">
              <h3 class="section-title">🏷️ Informations du label</h3>
              <div class="form-grid">
                <label class="form-field">
                  <span class="required">Nom du label</span>
                  <div class="input-with-suggestion">
                    <input
                      id="name"
                      ref="nameInput"
                      v-model.trim="formData.name"
                      type="text"
                      required
                      placeholder="Ex: Columbia Records, Universal Music..."
                      :class="{ 'has-suggestion': prefillName && !formData.name, 'input-error': !isNameValid && formData.name }"
                      :disabled="saving"
                      autofocus
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
                    v-model.trim="formData.description"
                    rows="3"
                    placeholder="Histoire du label, artistes phares, spécialités..."
                    maxlength="500"
                    :disabled="saving"
                  ></textarea>
                  <div class="char-counter">
                    {{ formData.description?.length || 0 }}/500 caractères
                  </div>
                  <button
                    v-if="isEditing"
                    type="button"
                    @click="suggestDescription"
                    class="suggest-description-button"
                    :disabled="saving || suggestingDescription"
                  >
                    {{ suggestingDescription ? '⏳ Recherche…' : '🔄 Suggérer automatiquement (Discogs)' }}
                  </button>
                  <div v-if="suggestDescriptionError" class="form-error">{{ suggestDescriptionError }}</div>
                </label>

                <div class="form-grid-2cols">
                  <label class="form-field">
                    <span>Pays d'origine</span>
                    <select
                      id="country_id"
                      v-model="formData.country_id"
                      :disabled="saving"
                    >
                      <option value="">Sélectionnez un pays</option>
                      <option v-for="country in countries" :key="country.id" :value="country.id">
                        {{ country.name }}
                      </option>
                    </select>
                  </label>
                  <label class="form-field">
                    <span>Année de fondation</span>
                    <input
                      id="founding_year"
                      v-model.number="formData.founding_year"
                      type="number"
                      min="1900"
                      :max="new Date().getFullYear()"
                      placeholder="Ex: 1980"
                      :disabled="saving"
                    />
                  </label>
                </div>

                <label class="form-field form-field-full">
                  <span>Site web <span class="optional-text">(optionnel)</span></span>
                  <input
                    id="website"
                    v-model.trim="formData.website"
                    type="url"
                    placeholder="https://www.example.com"
                    :disabled="saving"
                  />
                </label>
              </div>
            </div>

            <div v-if="isEditing && labelStats" class="form-section stats-section">
              <h3 class="section-title">📊 Statistiques</h3>
              <div class="stats-grid">
                <div class="stat-card">
                  <div class="stat-icon">💿</div>
                  <div class="stat-content">
                    <div class="stat-label">Disques dans la collection</div>
                    <div class="stat-value">{{ labelStats.discs_count || 0 }}</div>
                  </div>
                </div>
                <div class="stat-card">
                  <div class="stat-icon">🎤</div>
                  <div class="stat-content">
                    <div class="stat-label">Artistes associés</div>
                    <div class="stat-value">{{ labelStats.artists_count || 0 }}</div>
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
          <p>Vous avez commencé à modifier le label. Si vous fermez maintenant, ces modifications seront perdues.</p>
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
              <h2>Label déjà existant</h2>
              <p class="header-subtitle">
                Un label avec ce nom existe déjà dans votre collection
              </p>
            </div>
          </div>
          <button class="icon-action-btn" @click="closeDuplicateModal" aria-label="Fermer">
            <span>&times;</span>
          </button>
        </div>
        <div class="info-section duplicate-warning">
          <h4>🏷️ Label en conflit</h4>
          <p>Le nom "<strong>{{ duplicateLabelName }}</strong>" est déjà utilisé par un label existant dans votre collection.</p>
          <p class="help-text">Pour créer ce label, vous devez choisir un nom différent.</p>
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
import { ref, computed, watch, onMounted, nextTick } from 'vue';
import { useApi } from '@/composables/useApi';

const emit = defineEmits(['close', 'label-saved']);
const api = useApi();
const props = defineProps({
  isOpen: Boolean,
  labelData: Object,
  prefillName: String,
  apiError: String
});

const { get } = useApi();

// Refs
const nameInput = ref(null);
const saving = ref(false);
const countries = ref([]);
const labelStats = ref(null);
const localApiError = ref(null);
const isUnsavedChangesModalOpen = ref(false);
const isDuplicateModalOpen = ref(false);  // ✅ Ajouté
const duplicateLabelName = ref('');      // ✅ Ajouté
const overlayMouseDownTarget = ref(null); // Pour éviter fermeture lors de sélection texte
const suggestingDescription = ref(false);
const suggestDescriptionError = ref(null);

const formData = ref({
  id: null,
  name: '',
  description: '',
  country_id: '',
  founding_year: '',
  website: '',
  created_at: '',
  updated_at: ''
});

// Computed
const isEditing = computed(() => !!props.labelData?.id);
const isNameValid = computed(() => formData.value.name.trim().length >= 2);
const isFormValid = computed(() => {
  return formData.value.name?.trim() && isNameValid.value;
});

// Méthodes
const loadCountries = async () => {
  try {
    const countriesRes = await get('/countries');
    countries.value = countriesRes || [];
    console.log('🌍 [LabelsModal] Pays chargés:', countries.value.length);
  } catch (error) {
    console.error('❌ [LabelsModal] Erreur chargement pays:', error);
  }
};

const loadLabelStats = async () => {
  if (!isEditing.value) return;
  try {
    const stats = await get(`/labels/${formData.value.id}/stats`);
    labelStats.value = stats;
    console.log('📊 [LabelsModal] Statistiques chargées:', stats);
  } catch (error) {
    console.error('❌ [LabelsModal] Erreur chargement statistiques:', error);
    labelStats.value = null;
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

const suggestDescription = async () => {
  if (!formData.value.id) return;
  suggestingDescription.value = true;
  suggestDescriptionError.value = null;
  try {
    const suggestion = await api.post(`/labels/${formData.value.id}/description/suggest`);
    const data = suggestion.data || suggestion;
    formData.value.description = data.description;
  } catch (error) {
    console.error('Erreur suggestion description:', error);
    suggestDescriptionError.value = error.message || 'Impossible de trouver une description pour ce label';
  } finally {
    suggestingDescription.value = false;
  }
};

const handleSave = async () => {
  console.log('💿 [LabelsModal] Début handleSave, mode édition:', isEditing.value);
  console.log('💾 [LabelsModal] Sauvegarde des données:', formData.value);
  saving.value = true;
  localApiError.value = null;

  try {
    // Préparer les données
    const dataToSend = {
      name: formData.value.name.trim(),
      description: formData.value.description?.trim() || null,
      country_id: formData.value.country_id ? parseInt(formData.value.country_id) : null,
      founding_year: formData.value.founding_year ? parseInt(formData.value.founding_year) : null,
      website: formData.value.website?.trim() || null
    };
    console.log('📤 [LabelsModal] Données à envoyer:', dataToSend);

    let savedLabel;

    if (isEditing.value) {
      console.log('🔄 [LabelsModal] Mise à jour label ID:', props.labelData.id);
      const response = await api.put(`/labels/${props.labelData.id}`, dataToSend);
      console.log('🔍 [LabelsModal] Structure response PUT:', response);
      savedLabel = response.data || response;
    } else {
      console.log('➕ [LabelsModal] Création nouveau label');
      const response = await api.post('/labels', dataToSend);
      console.log('🔍 [LabelsModal] Structure response POST:', response);
      savedLabel = response.data || response;
      console.log('✅ [LabelsModal] Label créé avec ID:', savedLabel.id);
    }

    console.log('💿 [LabelsModal] Label sauvegardé:', savedLabel);
    emit('label-saved', savedLabel);
    resetForm();
    emit('close');

  } catch (error) {
    console.error('❌ [LabelsModal] Erreur lors de la sauvegarde:', error);
    console.error('❌ [LabelsModal] Stack:', error.stack);
    console.error('❌ [LabelsModal] Message:', error.message);
    console.error('❌ [LabelsModal] Response data:', error.response?.data);
    console.error('❌ [LabelsModal] Response status:', error.response?.status);

    const isDuplicate = error.response?.status === 409 ||
                        error.message?.includes('existe déjà') ||
                        error.message?.includes('Conflict') ||
                        error.response?.data?.message?.includes('existe déjà');

    if (isDuplicate) {
      duplicateLabelName.value = formData.value.name.trim();
      isDuplicateModalOpen.value = true;
      console.log('⚠️ [LabelsModal] Doublon détecté — modal ouverte');
    } else {
      localApiError.value = error.message || error.response?.data?.message || 'Une erreur est survenue lors de la sauvegarde';
    }

  } finally {
    saving.value = false;
    console.log('💿 [LabelsModal] Fin handleSave, saving:', saving.value);
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
  console.log('💿 [LabelsModal] handleClose appelé, saving:', saving.value);
  if (saving.value) return;

  const originalName = props.labelData?.name || '';
  const originalDescription = props.labelData?.description || '';
  const originalWebsite = props.labelData?.website || '';
  const originalCountry = props.labelData?.country_id || '';
  const originalYear = props.labelData?.founding_year || '';

  const hasUnsavedChanges = (
    formData.value.name !== originalName ||
    formData.value.description !== originalDescription ||
    formData.value.website !== originalWebsite ||
    formData.value.country_id !== originalCountry ||
    formData.value.founding_year !== originalYear
  );

  console.log('💿 [LabelsModal] hasUnsavedChanges:', hasUnsavedChanges);
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
  duplicateLabelName.value = '';
  // Formulaire reste ouvert — comme dans FormatsModal
};

const resetForm = () => {
  console.log('💿 [LabelsModal] Réinitialisation du formulaire');
  formData.value = {
    id: null,
    name: '',
    description: '',
    country_id: '',
    founding_year: '',
    website: '',
    created_at: '',
    updated_at: ''
  };
  labelStats.value = null;
  localApiError.value = null;
  isDuplicateModalOpen.value = false;
  duplicateLabelName.value = '';
};

const initializeForm = async () => {
  console.log('💿 [LabelsModal] Initialisation du formulaire');
  console.log('💿 [LabelsModal] prefillName:', props.prefillName);
  console.log('💿 [LabelsModal] labelData:', props.labelData);

  if (props.labelData?.id) {
    console.log('📝 [LabelsModal] Mode édition activé');
    formData.value = {
      ...props.labelData,
      created_at: props.labelData.created_at || '',
      updated_at: props.labelData.updated_at || ''
    };
    await loadLabelStats();
    console.log('📝 [LabelsModal] Données formulaire édition:', formData.value);
  } else if (props.prefillName) {
    console.log('🎯 [LabelsModal] Mode création avec pré-remplissage');
    formData.value = {
      id: null,
      name: props.prefillName,
      description: '',
      country_id: '',
      founding_year: '',
      website: '',
      created_at: '',
      updated_at: ''
    };
    console.log('🎯 [LabelsModal] Données formulaire pré-remplies:', formData.value);
  } else if (props.labelData && props.labelData.name) {
    console.log('🎯 [LabelsModal] Mode création avec données labelData');
    formData.value = {
      id: null,
      name: props.labelData.name,
      description: props.labelData.description || '',
      country_id: props.labelData.country_id || '',
      founding_year: props.labelData.founding_year || '',
      website: props.labelData.website || '',
      created_at: '',
      updated_at: ''
    };
    console.log('🎯 [LabelsModal] Données formulaire avec labelData:', formData.value);
  } else {
    console.log('➕ [LabelsModal] Mode création normal');
    resetForm();
  }

  localApiError.value = null;

  // Focus sur le champ name après initialisation
  await nextTick();
  if (nameInput.value && !isEditing.value) {
    nameInput.value.focus();
  }
};

// Watchers
watch(() => props.isOpen, (isOpen) => {
  console.log('🔄 [LabelsModal] Watcher isOpen →', isOpen);
  if (isOpen) {
    initializeForm();
  }
});

watch(() => props.prefillName, (newPrefillName) => {
  console.log('🔄 [LabelsModal] Watcher prefillName →', newPrefillName);
  if (props.isOpen && newPrefillName && !formData.value.id) {
    formData.value.name = newPrefillName;
  }
});

watch(() => props.labelData, (newLabelData) => {
  console.log('🔄 [LabelsModal] Watcher labelData →', newLabelData);
  if (props.isOpen) {
    initializeForm();
  }
}, { deep: true });

// Lifecycle
onMounted(() => {
  loadCountries();
});
</script>

<style scoped>

/* Teleporté à body : sans ce z-index dédié, hérite du .modal-overlay
   partagé (z-index: 50, App.vue), invisible derrière la modale
   d'ajout/édition de disque (.discs-modal-overlay, z-index: 9999) quand
   ouvert via son bouton "+". */
.modal-overlay {
  z-index: 10000;
}
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

.label-form {
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

.form-grid-2cols {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 18px;
}

@media (max-width: 768px) {
  .form-grid-2cols { grid-template-columns: 1fr; }
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

.suggest-description-button {
  margin-top: 8px;
  padding: 8px 14px;
  background: rgba(var(--tint-rgb), 0.05);
  color: var(--accent-soft);
  border: 1px solid var(--line);
  border-radius: 10px;
  font-size: 0.85em;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.suggest-description-button:hover:not(:disabled) {
  background: rgba(var(--tint-rgb), 0.09);
  color: var(--accent);
}

.suggest-description-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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