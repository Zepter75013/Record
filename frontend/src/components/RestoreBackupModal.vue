<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  backup: {
    type: Object,
    default: null,
  },
  isRestoring: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:modelValue', 'confirm'])

const CONFIRM_WORD = 'RESTAURER'
const confirmText = ref('')

watch(
  () => props.modelValue,
  (isOpen) => {
    if (!isOpen) confirmText.value = ''
  }
)

function closeModal() {
  if (props.isRestoring) return
  emit('update:modelValue', false)
}

function confirmRestore() {
  if (!props.backup || props.isRestoring || confirmText.value.trim() !== CONFIRM_WORD) return
  emit('confirm', props.backup)
}
</script>

<template>
  <div v-if="modelValue && backup" class="modal-overlay" @click.self="closeModal">
    <section
      class="modal-card restore-modal-card"
      role="dialog"
      aria-modal="true"
      aria-labelledby="restore-backup-title"
      aria-describedby="restore-backup-description"
    >
      <div class="modal-header">
        <div class="restore-modal-heading">
          <span class="restore-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none">
              <path
                d="M4 4v6h6M20 20v-6h-6M4.5 15a8 8 0 0 0 13.9 3.2M19.5 9A8 8 0 0 0 5.6 5.8"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </span>

          <div>
            <p class="eyebrow">Restauration</p>
            <h2 id="restore-backup-title">Restaurer cette sauvegarde ?</h2>
          </div>
        </div>
      </div>

      <p id="restore-backup-description" class="restore-modal-text">
        Toutes les données actuelles (disques, artistes, genres, formats, pays, labels…) seront
        <strong>remplacées</strong> par le contenu de la sauvegarde
        <strong>{{ backup.name }}</strong>
        . Cette action est irréversible.
      </p>

      <p class="restore-modal-note">
        Par sécurité, une sauvegarde de l'état actuel sera créée automatiquement juste
        avant la restauration — tu pourras y revenir en cas d'erreur.
      </p>

      <label class="confirm-field">
        <span>
          Tape <strong>{{ CONFIRM_WORD }}</strong> pour confirmer
        </span>
        <input
          v-model="confirmText"
          type="text"
          autocomplete="off"
          :placeholder="CONFIRM_WORD"
          :disabled="isRestoring"
        />
      </label>

      <div class="modal-actions">
        <button class="ghost-btn" type="button" :disabled="isRestoring" @click="closeModal">
          Annuler
        </button>

        <button
          class="danger-btn"
          type="button"
          :disabled="isRestoring || confirmText.trim() !== CONFIRM_WORD"
          @click="confirmRestore"
        >
          {{ isRestoring ? 'Restauration...' : 'Restaurer' }}
        </button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.restore-modal-card {
  animation: modal-pop-in 180ms ease;
  transform-origin: center;
}

.modal-header {
  margin-bottom: 0.4rem;
}

.restore-modal-heading {
  display: flex;
  align-items: flex-start;
  gap: 0.9rem;
}

.restore-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 999px;
  background: rgba(239, 68, 68, 0.12);
  color: var(--negative-text);
  flex-shrink: 0;
}

.restore-icon svg {
  width: 1.25rem;
  height: 1.25rem;
}

.eyebrow {
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: 0.74rem;
}

.modal-header h2 {
  margin-top: 0.35rem;
  font-size: 1.45rem;
  line-height: 1.15;
  color: var(--text);
}

.restore-modal-text {
  margin: 1rem 0 0.75rem;
  line-height: 1.6;
  color: var(--text-soft);
}

.restore-modal-text strong {
  color: var(--text);
}

.restore-modal-note {
  margin: 0 0 1rem;
  padding: 0.85rem 0.95rem;
  border-radius: 14px;
  background: rgba(var(--tint-rgb), 0.04);
  border: 1px solid var(--line-soft);
  color: var(--text-dim);
  line-height: 1.5;
  font-size: 0.92rem;
}

.confirm-field {
  display: grid;
  gap: 0.4rem;
  margin-bottom: 1.5rem;
  font-size: 0.88rem;
  color: var(--text-soft);
}

.confirm-field strong {
  color: var(--text);
}

.confirm-field input {
  border-radius: 12px;
  border: 1px solid var(--line);
  background: rgba(var(--tint-rgb), 0.04);
  color: var(--text);
  padding: 0.7rem 0.85rem;
  font-size: 0.95rem;
}

.confirm-field input:focus {
  outline: 2px solid rgba(220, 38, 38, 0.4);
  outline-offset: 1px;
}

@media (max-width: 640px) {
  .modal-actions {
    flex-direction: column-reverse;
  }

  .modal-actions .ghost-btn,
  .modal-actions .danger-btn {
    width: 100%;
  }
}

@keyframes modal-pop-in {
  from {
    opacity: 0;
    transform: translateY(8px) scale(0.985);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@media (prefers-reduced-motion: reduce) {
  .restore-modal-card {
    animation: none;
  }
}
</style>
