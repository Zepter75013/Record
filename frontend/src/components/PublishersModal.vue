<!-- frontend/src/components/PublishersModal.vue -->
<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="isOpen" class="modal-overlay" @mousedown="onOverlayMouseDown" @click="onOverlayClick">
        <div class="modal-card" :class="{ 'is-saving': saving }">
          <div class="modal-header">
            <div class="header-content">
              <h2>{{ isEditing ? '✏️ Modifier l\'éditeur' : '➕ Ajouter un éditeur' }}</h2>
              <p class="modal-subtitle">
                {{ isEditing ? 'Modifiez les informations de l\'éditeur' : 'Ajoutez un nouvel éditeur à votre base' }}
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

          <form @submit.prevent="handleSave" class="publisher-form">
            <div class="form-section">
              <h3 class="section-title">🏢 Informations de l'éditeur</h3>
              <div class="form-grid">
                <label class="form-field">
                  <span class="required">Nom de l'éditeur</span>
                  <div class="input-with-suggestion">
                    <input
                      id="name"
                      ref="nameInput"
                      v-model.trim="formData.name"
                      type="text"
                      required
                      placeholder="Ex: Nintendo, Sony, Electronic Arts..."
                      :class="{ 'has-suggestion': prefillName && !formData.name, 'input-error': !isNameValid && formData.name }"
                      :disabled="saving"
                      @blur="validateName"
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
                    placeholder="Pays d'origine, notes..."
                    maxlength="500"
                    :disabled="saving"
                  ></textarea>
                  <div class="char-counter">
                    {{ formData.description?.length || 0 }}/500 caractères
                  </div>
                </label>
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
          <p>Vous avez commencé à modifier l'éditeur. Si vous fermez maintenant, ces modifications seront perdues.</p>
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
              <h2>Éditeur déjà existant</h2>
              <p class="header-subtitle">
                Un éditeur avec ce nom existe déjà dans votre collection
              </p>
            </div>
          </div>
          <button class="icon-action-btn" @click="closeDuplicateModal" aria-label="Fermer">
            <span>&times;</span>
          </button>
        </div>
        <div class="info-section duplicate-warning">
          <h4>🎮 Éditeur en conflit</h4>
          <p>Le nom "<strong>{{ duplicatePublisherName }}</strong>" est déjà utilisé par un éditeur existant dans votre collection.</p>
          <p class="help-text">Pour créer cet éditeur, vous devez choisir un nom différent.</p>
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
import { ref, computed, watch, nextTick } from 'vue';
import { useApi } from '@/composables/useApi';

const emit = defineEmits(['close', 'publisher-saved']);
const props = defineProps({
  isOpen: Boolean,
  publisherData: Object,
  prefillName: String,
  apiError: String
});

// États
const nameInput = ref(null);
const saving = ref(false);
const localApiError = ref(null);
const isUnsavedChangesModalOpen = ref(false);
const isDuplicateModalOpen = ref(false);
const duplicatePublisherName = ref('');
const overlayMouseDownTarget = ref(null); // Pour éviter fermeture lors de sélection texte

const formData = ref({
  id: null,
  name: '',
  description: '',
  created_at: '',
  updated_at: ''
});

// Computed
const isEditing = computed(() => !!props.publisherData?.id);
const isNameValid = computed(() => formData.value.name.trim().length >= 2);
const isFormValid = computed(() => {
  return formData.value.name?.trim() && isNameValid.value;
});

// Validation
const validateName = () => {
  if (formData.value.name && !isNameValid.value) {
    // Message déjà affiché via `!isNameValid && formData.name`
  }
};

const applyPrefillName = () => {
  if (props.prefillName) {
    formData.value.name = props.prefillName;
  }
};

const handleSave = async () => {
  if (!isFormValid.value) return;

  saving.value = true;
  localApiError.value = null;

  try {
    const dataToSend = {
      name: formData.value.name.trim(),
      description: formData.value.description?.trim() || null
    };

    let savedPublisher;

    if (isEditing.value) {
      const response = await useApi().put(`/publishers/${props.publisherData.id}`, dataToSend);
      savedPublisher = response.data || response;
    } else {
      const response = await useApi().post('/publishers', dataToSend);
      savedPublisher = response.data || response;
    }

    emit('publisher-saved', savedPublisher);
    emit('close');

  } catch (error) {
    console.error('❌ [PublishersModal] Erreur lors de la sauvegarde:', error);

    const isDuplicate = error.response?.status === 409 ||
                        error.message?.includes('existe déjà') ||
                        error.message?.includes('Conflict') ||
                        error.response?.data?.message?.includes('existe déjà');

    if (isDuplicate) {
      duplicatePublisherName.value = formData.value.name.trim();
      isDuplicateModalOpen.value = true;
    } else {
      localApiError.value = error.message || error.response?.data?.message || 'Une erreur est survenue lors de la sauvegarde';
    }

  } finally {
    saving.value = false;
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
  if (saving.value) return;

  const originalName = props.publisherData?.name || '';
  const originalDescription = props.publisherData?.description || '';

  const hasUnsavedChanges = (
    formData.value.name !== originalName ||
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
  duplicatePublisherName.value = '';
};

const resetForm = () => {
  formData.value = {
    id: null,
    name: '',
    description: '',
    created_at: '',
    updated_at: ''
  };
  localApiError.value = null;
  isDuplicateModalOpen.value = false;
  duplicatePublisherName.value = '';
};

const initializeForm = () => {
  if (props.publisherData?.id) {
    formData.value = {
      ...props.publisherData,
      created_at: props.publisherData.created_at || '',
      updated_at: props.publisherData.updated_at || ''
    };
  } else if (props.prefillName) {
    formData.value = {
      id: null,
      name: props.prefillName,
      description: ''
    };
  } else if (props.publisherData?.name) {
    formData.value = {
      id: null,
      name: props.publisherData.name,
      description: props.publisherData.description || ''
    };
  } else {
    resetForm();
  }

  localApiError.value = null;

  // Focus automatique sur le champ name en mode création
  nextTick(() => {
    if (nameInput.value && !isEditing.value) {
      nameInput.value.focus();
    }
  });
};

// Watchers
watch(() => props.isOpen, (isOpen) => {
  if (isOpen) {
    initializeForm();
  }
});

watch(() => props.prefillName, (newPrefillName) => {
  if (props.isOpen && newPrefillName && !formData.value.id) {
    formData.value.name = newPrefillName;
  }
});

watch(() => props.publisherData, () => {
  if (props.isOpen) {
    initializeForm();
  }
}, { deep: true });
</script>

<style scoped>

/* Teleporté à body : sans ce z-index dédié, hérite du .modal-overlay
   partagé (z-index: 50, App.vue), invisible derrière la modale
   d'ajout/édition de jeu (z-index: 9999) quand ouvert via son bouton "+". */
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

.publisher-form {
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
