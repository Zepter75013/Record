<script setup>
import FormatsModal from '@/components/FormatsModal.vue';
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { useApi } from '@/composables/useApi';

const { apiFetch } = useApi();
const formats = ref([]);
const isConfirmModalOpen = ref(false);
const formatToDelete = ref(null);
const selectedRowId = ref(null);
const API_URL = '/formats';
const isModalOpen = ref(false);
const currentFormat = ref(null);
const apiError = ref(null);
const isMobileView = ref(window.innerWidth < 768);

const updateMobileView = () => {
  isMobileView.value = window.innerWidth < 768;
};

const sortKey = ref('name');
const sortOrder = ref(1);

const fetchFormats = async () => {
  try {
    const data = await apiFetch(API_URL);
    formats.value = Array.isArray(data) ? data : [];
  } catch (error) {
    console.error('Erreur:', error);
    formats.value = [];
  }
};

const sortBy = (key) => {
  if (sortKey.value === key) {
    sortOrder.value *= -1;
  } else {
    sortKey.value = key;
    sortOrder.value = 1;
  }
};

const getSortIcon = (key) => {
  if (sortKey.value !== key) return '<span style="font-weight:700">A-Z</span>';
  return sortOrder.value === 1 ? '▲' : '▼';
};

const sortedFormats = computed(() => {
  const dataToSort = Array.isArray(formats.value) ? formats.value : [];
  return [...dataToSort].sort((a, b) => {
    const aValue = String(a[sortKey.value] || '').toLowerCase();
    const bValue = String(b[sortKey.value] || '').toLowerCase();
    if (aValue < bValue) return -1 * sortOrder.value;
    if (aValue > bValue) return 1 * sortOrder.value;
    return 0;
  });
});

const openModal = (format = null) => {
  apiError.value = null;
  currentFormat.value = format ? { ...format } : null;
  isModalOpen.value = true;
};

const closeModal = () => {
  apiError.value = null;
  isModalOpen.value = false;
  currentFormat.value = null;
};

const handleSave = (savedFormat) => {
  const existingIndex = formats.value.findIndex((f) => f.id === savedFormat.id);
  if (existingIndex !== -1) {
    formats.value[existingIndex] = savedFormat;
  } else {
    formats.value.push(savedFormat);
  }
};

const prepareDelete = (format) => {
  formatToDelete.value = format;
  isConfirmModalOpen.value = true;
};

const executeDelete = async () => {
  if (!formatToDelete.value) return;
  const id = formatToDelete.value.id;
  isConfirmModalOpen.value = false;
  try {
    await apiFetch(`${API_URL}/${id}`, { method: 'DELETE' });
    formats.value = formats.value.filter((f) => f.id !== id);
    if (selectedRowId.value === id) selectedRowId.value = null;
    formatToDelete.value = null;
  } catch (error) {
    alert(`Échec: ${error.message}`);
  }
};

const handleRowClick = (format) => {
  selectedRowId.value = selectedRowId.value === format.id ? null : format.id;
};

const handleRowDoubleClick = (format) => {
  selectedRowId.value = format.id;
  openModal(format);
};

onMounted(() => {
  fetchFormats();
  window.addEventListener('resize', updateMobileView);
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateMobileView);
});
</script>

<template>
  <div class="settings-formats-view">
    <header class="page-header">
      <div class="header-wrapper">
        <div class="header-top-row">
          <div class="title-section">
            <button @click="$router.push('/dashboard')" class="back-button ghost-btn">
              <span class="icon">🏠</span>
            </button>
            <span class="title-icon">📀</span>
            <h1>Formats</h1>
          </div>
          <div class="toolbar">
            <button @click="openModal()" class="primary-btn add-button">
              <span class="icon">➕</span>
              Ajouter un Format
            </button>
          </div>
        </div>
        <p class="subtitle">Gérez les formats de disques de votre collection</p>
        <p class="user-hint">
          <span>💡</span>
          <strong>Astuce :</strong> Clic pour sélectionner, double-clic pour modifier
        </p>
      </div>
    </header>

    <div class="content-panel panel">
      <div v-if="!isMobileView" class="table-responsive">
        <table class="data-table">
          <thead>
            <tr>
              <th class="id-column sortable" @click="sortBy('id')">
                <div class="sortable-content">
                  <span class="header-text">ID</span>
                  <span class="sort-icon" v-html="getSortIcon('id')"></span>
                </div>
              </th>
              <th class="name-column sortable" @click="sortBy('name')">
                <div class="sortable-content">
                  <span class="header-text">NOM</span>
                  <span class="sort-icon" v-html="getSortIcon('name')"></span>
                </div>
              </th>
              <th class="actions-column">
                <div class="sortable-content">
                  <span class="header-text">ACTIONS</span>
                </div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr 
              v-for="(format, index) in sortedFormats" 
              :key="format.id"
              :class="{ 'odd-row': index % 2 === 0, 'even-row': index % 2 !== 0, 'selected-row': selectedRowId === format.id }"
              @click="handleRowClick(format)"
              @dblclick="handleRowDoubleClick(format)"
            >
              <td class="id-column">{{ format.id }}</td>
              <td class="name-column">{{ format.name || '-' }}</td>
              <td class="actions-column">
                <div class="action-buttons-container">
                  <button @click.stop="openModal(format)" class="icon-action-btn edit-button">
                    <span class="icon">✏️</span>
                  </button>
                  <button @click.stop="prepareDelete(format)" class="icon-action-btn icon-action-btn-danger delete-button">
                    <span class="icon">🗑️</span>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-else class="cards-container">
        <div 
          v-for="format in sortedFormats" 
          :key="format.id"
          class="card-item"
          :class="{ 'selected-row': selectedRowId === format.id }"
        >
          <div class="card-main" @click="handleRowClick(format)" @dblclick="handleRowDoubleClick(format)">
            <div class="card-content">
              <div class="card-id">
                <span class="card-label">ID:</span>
                <span class="card-value">{{ format.id }}</span>
              </div>
              <div class="card-name">
                <span class="card-label">Nom:</span>
                <span class="card-value">{{ format.name || '-' }}</span>
              </div>
            </div>
          </div>
          <div class="card-actions-bottom">
            <button @click.stop="openModal(format)" class="action-button-mobile edit-button">
              <span class="icon">✏️</span>
              <span>Modifier</span>
            </button>
            <button @click.stop="prepareDelete(format)" class="action-button-mobile delete-button">
              <span class="icon">🗑️</span>
              <span>Supprimer</span>
            </button>
          </div>
        </div>
      </div>

      <div v-if="!formats || formats.length === 0" class="empty-table-message-standalone">
        Aucun format défini.
      </div>
      <div v-else class="table-summary-wrapper-standalone">
        <div class="summary-box">Nombre : {{ formats.length }}</div>
      </div>
    </div>

    <FormatsModal :is-open="isModalOpen" :format-data="currentFormat" :api-error="apiError" @close="closeModal" @format-saved="handleSave" />

    <Teleport to="body">
      <div v-if="isConfirmModalOpen" class="modal-overlay" @click.self="isConfirmModalOpen = false">
        <div class="modal-card">
          <div class="modal-header">
            <div class="header-title-container">
              <span class="title-icon">🗑️</span>
              <h2>Confirmer la Suppression</h2>
            </div>
            <button class="icon-action-btn" @click="isConfirmModalOpen = false"><span>&times;</span></button>
          </div>
          <div v-if="formatToDelete" class="delete-info">
            <h4>Format à supprimer : {{ formatToDelete.name }}</h4>
          </div>
          <div class="modal-actions">
            <button @click="isConfirmModalOpen = false" class="ghost-btn">Annuler</button>
            <button @click="executeDelete" class="danger-btn">Supprimer</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.settings-formats-view {
  padding: 20px;
  min-height: 100vh;
  box-sizing: border-box;
}

.title-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-button {
  display: none;
  padding: 0;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  align-items: center;
  justify-content: center;
}

@media (max-width: 767px) {
  .back-button {
    display: flex;
  }
}

.header-wrapper { padding: 0; margin: 0; }
.header-top-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 5px; flex-wrap: wrap; gap: 16px; }
.title-icon { font-size: 2em; }
.page-header h1 { color: var(--text); font-size: 2em; margin: 0; font-weight: bold; }
.subtitle { color: var(--text-soft); margin: 0 0 10px 0; font-size: 1.1em; }
.user-hint {
  color: var(--text-soft);
  margin: 0 0 20px 0;
  background: rgba(var(--tint-rgb), 0.04);
  padding: 8px 12px;
  border-radius: 10px;
  border-left: 4px solid var(--accent);
}
.toolbar { margin: 8px 0 0 0; display: flex; gap: 10px; }
.add-button .icon { margin-right: 8px; }
.content-panel { padding: 20px; }
.table-responsive { overflow-x: auto; margin-bottom: 20px; border: 1px solid var(--line); border-radius: 12px; }
.data-table { width: 100%; border-collapse: separate; border-spacing: 0; font-size: 0.9em; border-radius: 12px; overflow: hidden; }
.data-table thead { background-color: var(--table-header-color); }
.data-table thead th { padding: 12px 15px; background: var(--table-header-color); color: white; text-align: center; height: 50px; position: relative; }
.data-table thead th::after { content: ''; position: absolute; right: 0; top: 0; height: 100%; width: 1px; background-color: rgba(255, 255, 255, 0.2); }
.data-table thead th:last-child::after { display: none; }
.data-table tbody tr { transition: all 0.2s ease; cursor: pointer; height: 40px; }
.data-table tbody tr:hover { background: rgba(var(--tint-rgb), 0.06); }
.data-table tbody tr.selected-row { background: rgba(59, 130, 246, 0.12) !important; border-left: 4px solid var(--accent) !important; }
.data-table td { padding: 6px 15px; border-bottom: 1px solid var(--line-soft); text-align: left; font-size: 0.85em; height: 40px; color: var(--text); }
.data-table .id-column { width: 15%; text-align: center; font-weight: 600; color: var(--text-soft); }
.data-table .name-column { width: 70%; }
.data-table .actions-column { width: 15%; text-align: center; }
.action-buttons-container { display: flex; gap: 8px; justify-content: center; }
.edit-button .icon { color: var(--accent-soft); }
.delete-button .icon { color: var(--negative-text); }
.sortable { cursor: pointer; }
.sortable:hover { background: rgba(255, 255, 255, 0.1); }
.sortable-content { display: flex; justify-content: space-between; align-items: center; width: 100%; }
.header-text { flex-grow: 1; text-align: center; }
.sort-icon { margin-left: 8px; }
.empty-table-message-standalone { text-align: center; padding: 30px; color: var(--text-dim); font-style: italic; border: 2px dashed var(--line); border-radius: 12px; }
.summary-box { text-align: right; font-weight: bold; padding: 12px 15px; border: 1px solid var(--line); border-radius: 12px; color: var(--text-soft); }
@media (max-width: 767px) {
  .cards-container { display: flex; flex-direction: column; gap: 16px; }
  .card-item { background: var(--bg-elevated); border-radius: 14px; box-shadow: var(--shadow); border-left: 4px solid var(--accent); overflow: hidden; }
  .card-item.selected-row { border-left-color: var(--accent-soft); background: rgba(59, 130, 246, 0.08); }
  .card-main { padding: 14px; cursor: pointer; }
  .card-content { display: flex; flex-direction: column; gap: 10px; }
  .card-id, .card-name { display: flex; gap: 8px; }
  .card-label { font-weight: 700; color: var(--text-soft); min-width: 60px; }
  .card-value { color: var(--text); }
  .card-actions-bottom { display: flex; gap: 8px; padding: 10px 14px; background: rgba(var(--tint-rgb), 0.04); border-top: 1px solid var(--line-soft); }
  .action-button-mobile { flex: 1; display: flex; align-items: center; justify-content: center; gap: 6px; padding: 10px 16px; border: none; border-radius: 10px; font-weight: 600; cursor: pointer; min-height: 44px; }
  .action-button-mobile.edit-button { background: var(--accent); color: white; }
  .action-button-mobile.delete-button { background: var(--negative-text); color: #2a0a0a; }
}
@media (max-width: 768px) {
  .settings-formats-view { padding: 10px; }
  .header-top-row { flex-direction: column; align-items: flex-start; }
  .add-button { width: 100%; justify-content: center; }
}

.header-title-container {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.delete-info {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.28);
  border-radius: 14px;
  padding: 16px;
  margin-bottom: 20px;
  color: var(--text);
}
</style>
