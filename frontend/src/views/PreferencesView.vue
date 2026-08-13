<script setup>
import { onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useThemeStore } from '@/stores/theme'
import { useStreamingPreferencesStore } from '@/stores/streamingPreferences'
import { useDisplayPreferencesStore } from '@/stores/displayPreferences'
import { formatDate } from '@/utils/format'
import RestoreBackupModal from '@/components/RestoreBackupModal.vue'
import {
  fetchBackups,
  createBackup,
  downloadBackup,
  restoreBackup,
  restoreUploadedBackup,
  deleteBackup,
  fetchBackupSettings,
  updateBackupSettings,
  pickBackupDirectory,
} from '@/services/backups'

const themeStore = useThemeStore()
const { theme } = storeToRefs(themeStore)

const THEME_OPTIONS = [
  { value: 'dark', label: 'Sombre', emoji: '🌙', description: 'Fond sombre, comme aujourd’hui.' },
  { value: 'light', label: 'Clair', emoji: '☀️', description: 'Fond clair pour un usage de jour.' },
  { value: 'system', label: 'Système', emoji: '🖥️', description: 'Suit le réglage de ton appareil.' },
]

function selectTheme(value) {
  themeStore.setTheme(value)
}

const displayPrefs = useDisplayPreferencesStore()
const { currencyCode, currencyPosition, numberFormatStyle, dateFormatStyle } = storeToRefs(displayPrefs)

const CURRENCY_OPTIONS = [
  { value: 'EUR', label: 'Euro — €' },
  { value: 'USD', label: 'Dollar US — $' },
  { value: 'GBP', label: 'Livre Sterling — £' },
  { value: 'CHF', label: 'Franc Suisse — CHF' },
]

const CURRENCY_POSITION_OPTIONS = [
  { value: 'after', label: 'Après le montant', example: '12,50 €' },
  { value: 'before', label: 'Avant le montant', example: '€12,50' },
]

const NUMBER_FORMAT_OPTIONS = [
  { value: 'fr', label: 'Français', example: '1 234,56' },
  { value: 'us', label: 'Anglais US', example: '1,234.56' },
  { value: 'ch', label: 'Suisse', example: "1'234.56" },
]

const DATE_FORMAT_OPTIONS = [
  { value: 'dmy', label: 'Jour/Mois/Année', example: '11/08/2026' },
  { value: 'mdy', label: 'Mois/Jour/Année', example: '08/11/2026' },
  { value: 'iso', label: 'Année-Mois-Jour', example: '2026-08-11' },
]

const streamingPrefs = useStreamingPreferencesStore()
const { preferredPlatform } = storeToRefs(streamingPrefs)

const PLATFORM_OPTIONS = [
  { value: 'apple_music', label: 'Apple Music', emoji: '🎵', description: 'Bouton « Écouter » pointant vers Apple Music.' },
  { value: 'spotify', label: 'Spotify', emoji: '🎧', description: 'Bouton « Écouter » pointant vers Spotify.' },
  { value: 'deezer', label: 'Deezer', emoji: '🎶', description: 'Bouton « Écouter » pointant vers Deezer.' },
  { value: 'youtube', label: 'YouTube', emoji: '📺', description: 'Bouton « Écouter » pointant vers YouTube.' },
]

function selectPlatform(value) {
  streamingPrefs.setPreferredPlatform(preferredPlatform.value === value ? null : value)
}

const backups = ref([])
const isLoadingBackups = ref(false)
const isCreatingBackup = ref(false)
const isRestoring = ref(false)
const backupError = ref('')
const backupSuccess = ref('')
const restoreTarget = ref(null)
const isRestoreModalOpen = ref(false)
const uploadedFile = ref(null)
const fileInput = ref(null)

const backupDirectory = ref('')
const backupDirectoryInput = ref('')
const isSavingDirectory = ref(false)
const isPickingDirectory = ref(false)

async function loadBackupSettings() {
  try {
    const settings = await fetchBackupSettings()
    backupDirectory.value = settings.directory
    backupDirectoryInput.value = settings.directory
  } catch (err) {
    backupError.value = err instanceof Error ? err.message : 'Impossible de charger le dossier de sauvegarde.'
  }
}

async function handleSaveDirectory() {
  if (isSavingDirectory.value) return

  isSavingDirectory.value = true
  backupError.value = ''
  backupSuccess.value = ''

  try {
    const settings = await updateBackupSettings(backupDirectoryInput.value)
    backupDirectory.value = settings.directory
    backupDirectoryInput.value = settings.directory
    backupSuccess.value = 'Dossier de sauvegarde mis à jour.'
    await loadBackups()
  } catch (err) {
    backupError.value = err instanceof Error ? err.message : 'Impossible de changer le dossier de sauvegarde.'
  } finally {
    isSavingDirectory.value = false
  }
}

async function handlePickDirectory() {
  if (isPickingDirectory.value) return

  isPickingDirectory.value = true
  backupError.value = ''
  backupSuccess.value = ''

  try {
    const picked = await pickBackupDirectory()
    if (picked) {
      backupDirectoryInput.value = picked
      await handleSaveDirectory()
    }
  } catch (err) {
    backupError.value = err instanceof Error ? err.message : 'Impossible d’ouvrir le sélecteur de dossier.'
  } finally {
    isPickingDirectory.value = false
  }
}

function formatBackupDate(value) {
  return formatDate(value, { withTime: true })
}

function formatFileSize(bytes) {
  const value = Number(bytes || 0)
  if (value < 1024) return `${value} o`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} Ko`
  return `${(value / (1024 * 1024)).toFixed(1)} Mo`
}

async function loadBackups() {
  isLoadingBackups.value = true
  backupError.value = ''

  try {
    backups.value = await fetchBackups()
  } catch (err) {
    backupError.value = err instanceof Error ? err.message : 'Impossible de charger les sauvegardes.'
  } finally {
    isLoadingBackups.value = false
  }
}

async function handleCreateBackup() {
  if (isCreatingBackup.value) return

  isCreatingBackup.value = true
  backupError.value = ''
  backupSuccess.value = ''

  try {
    await createBackup()
    backupSuccess.value = 'Sauvegarde créée avec succès.'
    await loadBackups()
  } catch (err) {
    backupError.value = err instanceof Error ? err.message : 'Impossible de créer la sauvegarde.'
  } finally {
    isCreatingBackup.value = false
  }
}

async function handleDownload(backup) {
  backupError.value = ''

  try {
    await downloadBackup(backup.name)
  } catch (err) {
    backupError.value = err instanceof Error ? err.message : 'Impossible de télécharger la sauvegarde.'
  }
}

function requestRestore(backup) {
  restoreTarget.value = backup
  isRestoreModalOpen.value = true
}

function triggerFileInput() {
  fileInput.value?.click()
}

function handleFileChange(event) {
  const file = event.target.files?.[0]
  if (!file) return

  uploadedFile.value = file
  restoreTarget.value = { name: file.name, isUpload: true }
  isRestoreModalOpen.value = true
}

function closeRestoreModal() {
  if (isRestoring.value) return

  isRestoreModalOpen.value = false
  restoreTarget.value = null
  uploadedFile.value = null
  if (fileInput.value) fileInput.value.value = ''
}

async function confirmRestore(backup) {
  isRestoring.value = true
  backupError.value = ''
  backupSuccess.value = ''

  try {
    if (backup.isUpload && uploadedFile.value) {
      await restoreUploadedBackup(uploadedFile.value)
    } else {
      await restoreBackup(backup.name)
    }

    backupSuccess.value = 'Restauration effectuée avec succès.'
    isRestoreModalOpen.value = false
    restoreTarget.value = null
    uploadedFile.value = null
    if (fileInput.value) fileInput.value.value = ''
    await loadBackups()
  } catch (err) {
    backupError.value = err instanceof Error ? err.message : 'Impossible de restaurer cette sauvegarde.'
  } finally {
    isRestoring.value = false
  }
}

async function handleDelete(backup) {
  backupError.value = ''
  backupSuccess.value = ''

  try {
    await deleteBackup(backup.name)
    await loadBackups()
  } catch (err) {
    backupError.value = err instanceof Error ? err.message : 'Impossible de supprimer la sauvegarde.'
  }
}

onMounted(() => {
  loadBackupSettings()
  loadBackups()
})
</script>

<template>
  <div class="settings-preferences-view">
    <header class="page-header">
      <div class="header-wrapper">
        <div class="header-top-row">
          <div class="title-section">
            <button @click="$router.push('/dashboard')" class="back-button ghost-btn">
              <span class="icon">🏠</span>
            </button>
            <span class="title-icon">⚙️</span>
            <h1>Préférences</h1>
          </div>
        </div>
        <p class="subtitle">Ajuste l'apparence de l'application et gère tes sauvegardes.</p>
      </div>
    </header>

    <div class="preferences-view">
      <section class="panel preferences-card">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Apparence</p>
            <h2>Thème de l'application</h2>
          </div>
        </div>

        <div class="theme-options">
          <button
            v-for="option in THEME_OPTIONS"
            :key="option.value"
            type="button"
            class="theme-option"
            :class="{ 'is-active': theme === option.value }"
            @click="selectTheme(option.value)"
          >
            <span class="theme-option-emoji" aria-hidden="true">{{ option.emoji }}</span>
            <span class="theme-option-label">{{ option.label }}</span>
            <span class="theme-option-description">{{ option.description }}</span>
          </button>
        </div>
      </section>

      <section class="panel preferences-card">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Personnalisation</p>
            <h2>Format d'affichage</h2>
          </div>
        </div>

        <div class="preferences-field-grid">
          <label class="form-field">
            <span>Devise</span>
            <select :value="currencyCode" @change="displayPrefs.setCurrencyCode($event.target.value)">
              <option v-for="option in CURRENCY_OPTIONS" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>

          <label class="form-field">
            <span>Placement de la devise</span>
            <select :value="currencyPosition" @change="displayPrefs.setCurrencyPosition($event.target.value)">
              <option v-for="option in CURRENCY_POSITION_OPTIONS" :key="option.value" :value="option.value">
                {{ option.label }} ({{ option.example }})
              </option>
            </select>
          </label>

          <label class="form-field">
            <span>Format des nombres</span>
            <select :value="numberFormatStyle" @change="displayPrefs.setNumberFormatStyle($event.target.value)">
              <option v-for="option in NUMBER_FORMAT_OPTIONS" :key="option.value" :value="option.value">
                {{ option.label }} ({{ option.example }})
              </option>
            </select>
          </label>

          <label class="form-field">
            <span>Format de date</span>
            <select :value="dateFormatStyle" @change="displayPrefs.setDateFormatStyle($event.target.value)">
              <option v-for="option in DATE_FORMAT_OPTIONS" :key="option.value" :value="option.value">
                {{ option.label }} ({{ option.example }})
              </option>
            </select>
          </label>
        </div>
      </section>

      <section class="panel preferences-card">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Écoute</p>
            <h2>Plateforme de streaming</h2>
          </div>
        </div>

        <p class="preferences-hint">
          Choisis ta plateforme : chaque disque affichera un seul bouton « Écouter », pointant
          vers cette plateforme (ou une recherche sur celle-ci si le disque n'a pas de lien direct).
        </p>

        <div class="theme-options">
          <button
            v-for="option in PLATFORM_OPTIONS"
            :key="option.value"
            type="button"
            class="theme-option"
            :class="{ 'is-active': preferredPlatform === option.value }"
            @click="selectPlatform(option.value)"
          >
            <span class="theme-option-emoji" aria-hidden="true">{{ option.emoji }}</span>
            <span class="theme-option-label">{{ option.label }}</span>
            <span class="theme-option-description">{{ option.description }}</span>
          </button>
        </div>
      </section>

      <section class="panel preferences-card">
        <div class="panel-header">
          <div>
            <p class="eyebrow">Données</p>
            <h2>Sauvegarde et restauration</h2>
          </div>
          <button class="primary-btn" type="button" :disabled="isCreatingBackup" @click="handleCreateBackup">
            {{ isCreatingBackup ? 'Création...' : '+ Créer une sauvegarde' }}
          </button>
        </div>

        <p v-if="backupError" class="form-error">{{ backupError }}</p>
        <p v-if="backupSuccess" class="form-success">{{ backupSuccess }}</p>

        <div class="backup-directory-field">
          <label>
            <span>Dossier de stockage des sauvegardes</span>
            <div class="backup-directory-input-row">
              <input
                v-model="backupDirectoryInput"
                type="text"
                placeholder="/Users/toi/Sauvegardes/DisquesManager"
                :disabled="isSavingDirectory || isPickingDirectory"
              />
              <button
                class="ghost-btn"
                type="button"
                :disabled="isPickingDirectory || isSavingDirectory"
                @click="handlePickDirectory"
              >
                {{ isPickingDirectory ? 'Ouverture du Finder...' : '📁 Parcourir...' }}
              </button>
              <button
                class="ghost-btn"
                type="button"
                :disabled="isSavingDirectory || isPickingDirectory || backupDirectoryInput.trim() === backupDirectory"
                @click="handleSaveDirectory"
              >
                {{ isSavingDirectory ? 'Enregistrement...' : 'Enregistrer' }}
              </button>
            </div>
          </label>
          <p class="backup-directory-hint">
            Chemin absolu, sur la machine qui fait tourner le serveur. « Parcourir » ouvre
            le sélecteur Finder directement sur cette machine (macOS uniquement). Changer ce
            dossier ne déplace pas les sauvegardes déjà existantes.
          </p>
        </div>

        <p v-if="isLoadingBackups" class="backups-empty">Chargement des sauvegardes...</p>

        <p v-else-if="!backups.length" class="backups-empty">
          Aucune sauvegarde pour l'instant — clique sur « Créer une sauvegarde » pour en générer une.
        </p>

        <div v-else class="backups-list">
          <article v-for="backup in backups" :key="backup.name" class="backup-row">
            <div class="backup-row-main">
              <strong>{{ formatBackupDate(backup.created_at) }}</strong>
              <span class="backup-row-meta">{{ backup.name }} · {{ formatFileSize(backup.size_bytes) }}</span>
            </div>

            <div class="backup-row-actions">
              <button class="ghost-btn" type="button" @click="handleDownload(backup)">Télécharger</button>
              <button class="ghost-btn" type="button" @click="requestRestore(backup)">Restaurer</button>
              <button class="ghost-btn ghost-btn-danger" type="button" @click="handleDelete(backup)">
                Supprimer
              </button>
            </div>
          </article>
        </div>

        <div class="backup-upload">
          <p class="backup-upload-label">Restaurer depuis un fichier de sauvegarde (.sql)</p>
          <div class="backup-upload-control">
            <button class="ghost-btn" type="button" @click="triggerFileInput">Choisir un fichier</button>
            <span class="backup-upload-filename">{{ uploadedFile?.name || 'Aucun fichier choisi' }}</span>
            <input
              ref="fileInput"
              class="backup-upload-input"
              type="file"
              accept=".sql"
              @change="handleFileChange"
            />
          </div>
        </div>
      </section>
    </div>

    <RestoreBackupModal
      :model-value="isRestoreModalOpen"
      :backup="restoreTarget"
      :is-restoring="isRestoring"
      @update:model-value="closeRestoreModal"
      @confirm="confirmRestore"
    />
  </div>
</template>

<style scoped>
.settings-preferences-view {
  padding: 20px;
  min-height: 100vh;
  box-sizing: border-box;
}

.header-wrapper { padding: 0; margin: 0; }
.header-top-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 5px; flex-wrap: wrap; gap: 16px; }
.title-section { display: flex; align-items: center; gap: 12px; }
.back-button { display: none; padding: 0; width: 38px; height: 38px; border-radius: 50%; align-items: center; justify-content: center; }

@media (max-width: 767px) {
  .back-button { display: flex; }
  .settings-preferences-view { padding: 10px; }
}

.title-icon { font-size: 2em; }
.page-header h1 { color: var(--text); font-size: 2em; margin: 0; font-weight: bold; }
.subtitle { color: var(--text-soft); margin: 0 0 20px 0; font-size: 1.1em; }

.preferences-view {
  display: grid;
  gap: 0.9rem;
}

.preferences-card {
  padding: 1.1rem;
}

.preferences-field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.9rem;
  margin-top: 0.9rem;
}

@media (max-width: 640px) {
  .preferences-field-grid {
    grid-template-columns: 1fr;
  }
}

.theme-options {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
  margin-top: 0.9rem;
}

.theme-option {
  display: grid;
  justify-items: center;
  gap: 0.4rem;
  padding: 1.1rem 0.9rem;
  border-radius: 16px;
  border: 1px solid var(--line-soft);
  background: rgba(var(--tint-rgb), 0.03);
  color: var(--text-soft);
  cursor: pointer;
  text-align: center;
  transition: background 140ms ease, border-color 140ms ease, transform 140ms ease, color 140ms ease;
}

.theme-option:hover {
  transform: translateY(-1px);
  background: rgba(var(--tint-rgb), 0.06);
}

.theme-option.is-active {
  border-color: rgba(59, 130, 246, 0.5);
  background: rgba(59, 130, 246, 0.14);
  color: var(--text);
}

.theme-option-emoji {
  font-size: 1.6rem;
  line-height: 1;
}

.theme-option-label {
  font-weight: 700;
  font-size: 0.95rem;
  color: var(--text);
}

.theme-option-description {
  font-size: 0.78rem;
  color: var(--text-dim);
}

@media (max-width: 640px) {
  .theme-options {
    grid-template-columns: 1fr;
  }
}

.ghost-btn-danger {
  color: var(--negative-text);
}

.preferences-hint {
  margin: 0 0 0.2rem;
  color: var(--text-dim);
  font-size: 0.84rem;
  line-height: 1.5;
}

.backup-directory-field {
  margin-top: 0.9rem;
  padding: 0.9rem 1rem;
  border-radius: 14px;
  background: rgba(var(--tint-rgb), 0.03);
  border: 1px solid var(--line-soft);
}

.backup-directory-field label {
  display: grid;
  gap: 0.45rem;
  font-size: 0.85rem;
  color: var(--text-soft);
}

.backup-directory-input-row {
  display: flex;
  gap: 0.6rem;
}

.backup-directory-input-row input {
  flex: 1;
  min-width: 0;
  border-radius: 12px;
  border: 1px solid var(--line);
  background: rgba(var(--tint-rgb), 0.04);
  color: var(--text);
  padding: 0.6rem 0.8rem;
  font-size: 0.9rem;
  font-family: inherit;
}

.backup-directory-input-row .ghost-btn {
  flex-shrink: 0;
}

.backup-directory-hint {
  margin: 0.5rem 0 0;
  color: var(--text-dim);
  font-size: 0.76rem;
  line-height: 1.4;
}

@media (max-width: 640px) {
  .backup-directory-input-row {
    flex-direction: column;
  }
}

.backups-empty {
  margin: 0.9rem 0 0;
  color: var(--text-dim);
  font-size: 0.88rem;
}

.backups-list {
  display: grid;
  gap: 0.6rem;
  margin-top: 0.9rem;
}

.backup-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.85rem 1rem;
  border-radius: 14px;
  background: rgba(var(--tint-rgb), 0.03);
  border: 1px solid var(--line-soft);
  flex-wrap: wrap;
}

.backup-row-main {
  display: grid;
  gap: 0.2rem;
  min-width: 0;
}

.backup-row-main strong {
  color: var(--text);
  font-size: 0.92rem;
}

.backup-row-meta {
  color: var(--text-dim);
  font-size: 0.78rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.backup-row-actions {
  display: flex;
  gap: 0.5rem;
  flex-shrink: 0;
}

.backup-row-actions .ghost-btn {
  padding: 0.5rem 0.85rem;
  font-size: 0.82rem;
}

.backup-upload {
  margin-top: 1.2rem;
  padding-top: 1rem;
  border-top: 1px solid var(--line-soft);
}

.backup-upload-label {
  margin: 0 0 0.5rem;
  color: var(--text-soft);
  font-size: 0.88rem;
}

.backup-upload-control {
  display: flex;
  align-items: center;
  gap: 0.7rem;
}

.backup-upload-filename {
  color: var(--text-dim);
  font-size: 0.86rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Le vrai <input type="file"> reste dans le DOM (accessible, ciblable par
   .click()) mais invisible — c'est le bouton .ghost-btn juste à côté qui
   déclenche le sélecteur natif. */
.backup-upload-input {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@media (max-width: 640px) {
  .backup-row {
    flex-direction: column;
    align-items: stretch;
  }

  .backup-row-actions {
    flex-wrap: wrap;
  }
}
</style>
