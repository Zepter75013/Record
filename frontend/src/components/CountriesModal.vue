<!-- frontend/src/components/CountriesModal.vue - VERSION HARMONISÉE COMPLÈTE -->
<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="isOpen" class="modal-overlay" @mousedown="onOverlayMouseDown" @click="onOverlayClick">
        <div class="modal-card" :class="{ 'is-saving': saving || deleting }">
          <div class="modal-header">
            <div class="header-content">
              <h2>{{ isEditing ? '✏️ Modifier le pays' : '➕ Ajouter un pays' }}</h2>
              <p class="modal-subtitle">
                {{ isEditing ? 'Modifiez les informations du pays' : 'Ajoutez un nouveau pays à votre collection' }}
              </p>
            </div>
            <button
              @click="handleClose"
              class="icon-action-btn"
              :disabled="saving || deleting"
              aria-label="Fermer"
            >
              ✕
            </button>
          </div>

          <div v-if="apiError || localApiError" class="form-error modal-error-banner">
            <span class="error-icon">⚠️</span>
            <span>{{ apiError || localApiError }}</span>
          </div>

          <form @submit.prevent="handleSave" class="country-form">
            <div class="form-section">
              <h3 class="section-title">🌍 Informations du pays</h3>
              <div class="form-grid">
                <label class="form-field">
                  <span class="required">Nom du pays</span>
                  <div class="input-with-suggestion">
                    <input
                      id="name"
                      ref="nameInput"
                      v-model.trim="formData.name"
                      type="text"
                      required
                      placeholder="Ex: France, États-Unis, Japon..."
                      :class="{ 'has-suggestion': prefillName && !formData.name, 'input-error': fieldErrors.name && formData.name }"
                      :disabled="saving || deleting"
                      @blur="validateField('name')"
                      autofocus
                    />
                    <div v-if="prefillName && !formData.name" class="suggestion-badge">
                      <span class="suggestion-text">Suggestion : {{ prefillName }}</span>
                      <button type="button" @click="applyPrefillName" class="suggestion-button">
                        Utiliser
                      </button>
                    </div>
                  </div>
                  <div v-if="fieldErrors.name" class="form-error">
                    {{ fieldErrors.name }}
                  </div>
                  <div class="input-hint">
                    Nom complet du pays
                  </div>
                </label>

                <label class="form-field">
                  <span class="required">Code du pays</span>
                  <input
                    id="code"
                    v-model="formData.code"
                    type="text"
                    required
                    placeholder="Ex: FR, US, JP..."
                    maxlength="3"
                    :class="{ 'input-error': fieldErrors.code && formData.code }"
                    :disabled="saving || deleting"
                    @blur="validateField('code')"
                    @input="formData.code = formData.code.toUpperCase().trim()"
                    style="text-transform: uppercase;"
                  />
                  <div v-if="fieldErrors.code" class="form-error">
                    {{ fieldErrors.code }}
                  </div>
                  <div class="input-hint">
                    Code ISO à 2 ou 3 lettres
                  </div>
                </label>

                <label class="form-field form-field-full">
                  <span>Description <span class="optional-text">(optionnel)</span></span>
                  <textarea
                    id="description"
                    v-model.trim="formData.description"
                    rows="3"
                    placeholder="Informations sur le pays, culture musicale, histoire..."
                    maxlength="1000"
                    :disabled="saving || deleting"
                  ></textarea>
                  <div class="char-counter">
                    {{ formData.description?.length || 0 }}/1000 caractères
                  </div>
                </label>
              </div>
            </div>

            <div v-if="isEditing && countryStats" class="form-section stats-section">
              <h3 class="section-title">📊 Statistiques</h3>
              <div class="stats-grid">
                <div class="stat-card">
                  <div class="stat-icon">👤</div>
                  <div class="stat-content">
                    <div class="stat-label">Artistes associés</div>
                    <div class="stat-value">{{ countryStats.artists_count || 0 }}</div>
                  </div>
                </div>
                <div class="stat-card">
                  <div class="stat-icon">💿</div>
                  <div class="stat-content">
                    <div class="stat-label">Disques dans la collection</div>
                    <div class="stat-value">{{ countryStats.discs_count || 0 }}</div>
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

            <div class="modal-actions form-actions">
              <button
                v-if="isEditing"
                type="button"
                @click="handleDelete"
                class="danger-btn actions-left-btn"
                :disabled="saving || deleting"
              >
                <span v-if="deleting" class="spinner"></span>
                <span v-else>Supprimer</span>
              </button>
              <div class="actions-right">
                <button type="button" @click="handleClose" class="ghost-btn" :disabled="saving || deleting">
                  Annuler
                </button>
                <button type="submit" class="primary-btn" :disabled="!isFormValid || saving || deleting">
                  <span v-if="saving" class="spinner"></span>
                  <span v-else>{{ isEditing ? 'Mettre à jour' : 'Enregistrer' }}</span>
                </button>
              </div>
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
          <p>Vous avez commencé à modifier le pays. Si vous fermez maintenant, ces modifications seront perdues.</p>
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
              <h2>Pays déjà existant</h2>
              <p class="header-subtitle">
                Un pays avec ce nom ou ce code existe déjà dans votre collection
              </p>
            </div>
          </div>
          <button class="icon-action-btn" @click="closeDuplicateModal" aria-label="Fermer">
            <span>&times;</span>
          </button>
        </div>
        <div class="info-section duplicate-warning">
          <h4>🌍 Pays en conflit</h4>
          <p>Le nom "<strong>{{ duplicateCountryName }}</strong>" ou le code "<strong>{{ duplicateCountryCode }}</strong>" est déjà utilisé par un pays existant.</p>
          <p class="help-text">Pour créer ce pays, vous devez choisir un nom ou un code différent.</p>
        </div>
        <div class="modal-actions">
          <button @click="closeDuplicateModal" class="primary-btn">
            ✏️ Modifier les informations
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue';
import { useApi } from '@/composables/useApi';

const emit = defineEmits(['close', 'country-saved', 'delete']);
const api = useApi();
const props = defineProps({
  isOpen: Boolean,
  countryData: Object,
  prefillName: String,
  apiError: String
});

const { get } = useApi();

// États
const nameInput = ref(null);
const saving = ref(false);
const deleting = ref(false);
const countryStats = ref(null);
const localApiError = ref(null);
const fieldErrors = ref({ name: '', code: '' });
const isUnsavedChangesModalOpen = ref(false);
const isDuplicateModalOpen = ref(false);
const duplicateCountryName = ref('');
const duplicateCountryCode = ref('');
const overlayMouseDownTarget = ref(null); // Pour éviter fermeture lors de sélection texte

const formData = ref({
  id: null,
  name: '',
  code: '',
  description: '',
  created_at: '',
  updated_at: ''
});

// Computed
const isEditing = computed(() => !!props.countryData?.id);
const isFormValid = computed(() => {
  return formData.value.name?.trim() &&
         formData.value.code?.trim() &&
         !fieldErrors.value.name &&
         !fieldErrors.value.code;
});

// Méthodes
const loadCountryStats = async () => {
  if (!isEditing.value) return;
  try {
    const stats = await get(`/countries/${formData.value.id}/stats`);
    countryStats.value = stats;
    console.log('🌍 [CountriesModal] Statistiques chargées:', stats);
  } catch (error) {
    console.error('🌍 [CountriesModal] Erreur chargement statistiques:', error);
    countryStats.value = null;
  }
};

const validateField = (fieldName) => {
  const value = formData.value[fieldName]?.trim();
  switch (fieldName) {
    case 'name':
      if (!value) {
        fieldErrors.value.name = 'Le nom du pays est obligatoire';
      } else if (value.length < 2) {
        fieldErrors.value.name = 'Le nom doit contenir au moins 2 caractères';
      } else {
        fieldErrors.value.name = '';
      }
      break;
    case 'code':
      if (!value) {
        fieldErrors.value.code = 'Le code du pays est obligatoire';
      } else if (value.length < 2 || value.length > 3) {
        fieldErrors.value.code = 'Le code doit contenir 2 ou 3 caractères';
      } else if (!/^[A-Z]{2,3}$/.test(value)) {
        fieldErrors.value.code = 'Le code doit contenir uniquement des lettres';
      } else {
        fieldErrors.value.code = '';
      }
      break;
  }
};

const validateForm = () => {
  validateField('name');
  validateField('code');
  return !fieldErrors.value.name && !fieldErrors.value.code;
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
    fieldErrors.value.name = '';
  }
};

const handleSave = async () => {
  console.log('🌍 [CountriesModal] Début handleSave, mode édition:', isEditing.value);
  console.log('💾 [CountriesModal] Sauvegarde des données:', formData.value);
  if (!validateForm()) return;

  saving.value = true;
  localApiError.value = null;

  try {
    const dataToSend = {
      name: formData.value.name.trim(),
      code: formData.value.code.trim().toUpperCase(),
      description: formData.value.description?.trim() || null
    };
    console.log('📤 [CountriesModal] Données à envoyer:', dataToSend);

    let savedCountry;

    if (isEditing.value) {
      console.log('🔄 [CountriesModal] Mise à jour pays ID:', props.countryData.id);
      const response = await api.put(`/countries/${props.countryData.id}`, dataToSend);
      console.log('🔍 [CountriesModal] Structure response PUT:', response);
      savedCountry = response.data || response;
    } else {
      console.log('➕ [CountriesModal] Création nouveau pays');
      const response = await api.post('/countries', dataToSend);
      console.log('🔍 [CountriesModal] Structure response POST:', response);
      savedCountry = response.data || response;
      console.log('✅ [CountriesModal] Pays créé avec ID:', savedCountry.id);
    }

    console.log('🌍 [CountriesModal] Pays sauvegardé:', savedCountry);
    emit('country-saved', savedCountry);
    emit('close');

  } catch (error) {
    console.error('❌ [CountriesModal] Erreur lors de la sauvegarde:', error);
    console.error('❌ [CountriesModal] Stack:', error.stack);
    console.error('❌ [CountriesModal] Message:', error.message);
    console.error('❌ [CountriesModal] Response data:', error.response?.data);
    console.error('❌ [CountriesModal] Response status:', error.response?.status);

    const isDuplicate = error.response?.status === 409 ||
                        error.message?.includes('existe déjà') ||
                        error.message?.includes('Conflict') ||
                        error.response?.data?.message?.includes('existe déjà');

    if (isDuplicate) {
      duplicateCountryName.value = formData.value.name.trim();
      duplicateCountryCode.value = formData.value.code.trim();
      isDuplicateModalOpen.value = true;
      console.log('⚠️ [CountriesModal] Doublon détecté — modal ouverte');
    } else {
      localApiError.value = error.message || error.response?.data?.message || 'Une erreur est survenue lors de la sauvegarde';
    }

  } finally {
    saving.value = false;
    console.log('🌍 [CountriesModal] Fin handleSave, saving:', saving.value);
  }
};


const handleDelete = async () => {
  if (!confirm(`Êtes-vous sûr de vouloir supprimer le pays "${formData.value.name}" ?`)) return;

  deleting.value = true;
  localApiError.value = null;
  try {
    emit('delete', formData.value.id);
  } catch (error) {
    localApiError.value = error.message || 'Erreur lors de la suppression du pays';
    deleting.value = false;
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
  if (saving.value || deleting.value) return;

  const originalName = props.countryData?.name || '';
  const originalCode = props.countryData?.code || '';
  const originalDescription = props.countryData?.description || '';

  const hasUnsavedChanges = (
    formData.value.name !== originalName ||
    formData.value.code !== originalCode ||
    formData.value.description !== originalDescription
  );

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
  duplicateCountryName.value = '';
  duplicateCountryCode.value = '';
  // Réinitialise les erreurs pour permettre un nouvel essai
  fieldErrors.value.name = '';
  fieldErrors.value.code = '';
};

const resetForm = () => {
  formData.value = {
    id: null,
    name: '',
    code: '',
    description: '',
    created_at: '',
    updated_at: ''
  };
  fieldErrors.value = { name: '', code: '' };
  countryStats.value = null;
  localApiError.value = null;
  isDuplicateModalOpen.value = false;
  duplicateCountryName.value = '';
  duplicateCountryCode.value = '';
};

const initializeForm = async () => {
  if (props.countryData?.id) {
    formData.value = {
      ...props.countryData,
      created_at: props.countryData.created_at || '',
      updated_at: props.countryData.updated_at || ''
    };
    await loadCountryStats();
  } else if (props.prefillName) {
    formData.value = {
      id: null,
      name: props.prefillName,
      code: '',
      description: ''
    };
    fieldErrors.value.name = '';
  } else if (props.countryData?.name) {
    formData.value = {
      id: null,
      name: props.countryData.name,
      code: props.countryData.code || '',
      description: props.countryData.description || ''
    };
    fieldErrors.value.name = '';
    fieldErrors.value.code = '';
  } else {
    resetForm();
  }

  localApiError.value = null;

  // Focus sur le champ name après initialisation (en création)
  await nextTick();
  if (nameInput.value && !isEditing.value) {
    nameInput.value.focus();
  }
};

// Watchers
watch(() => props.isOpen, (isOpen) => {
  if (isOpen) initializeForm();
});

watch(() => props.prefillName, (newPrefillName) => {
  if (props.isOpen && newPrefillName && !formData.value.id) {
    formData.value.name = newPrefillName;
    fieldErrors.value.name = '';
  }
});

watch(() => props.countryData, (newCountryData) => {
  if (props.isOpen) initializeForm();
}, { deep: true });

watch(() => formData.value.name, () => {
  if (formData.value.name && fieldErrors.value.name) validateField('name');
});

watch(() => formData.value.code, () => {
  if (formData.value.code) {
    formData.value.code = formData.value.code.toUpperCase().trim();
    if (fieldErrors.value.code) validateField('code');
  }
});

// Lifecycle
onMounted(() => {
  // Rien à charger ici spécifiquement — countries déjà géré via props ou API externe
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

.country-form {
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
  grid-template-columns: 1fr 1fr;
  gap: 18px;
}

@media (max-width: 680px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
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

.input-hint {
  color: var(--text-dim);
  font-size: 0.85em;
  line-height: 1.4;
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

.form-actions {
  justify-content: space-between;
}

.actions-right {
  display: flex;
  gap: 12px;
}

@media (max-width: 680px) {
  .form-actions {
    flex-direction: column-reverse;
  }
  .actions-left-btn,
  .actions-right {
    width: 100%;
  }
  .actions-right {
    flex-direction: column-reverse;
  }
  .actions-right .primary-btn,
  .actions-right .ghost-btn {
    width: 100%;
  }
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