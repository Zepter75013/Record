<script setup>
import LabelsModal from '@/components/LabelsModal.vue';
import LabelsBulkDescriptionModal from '@/components/LabelsBulkDescriptionModal.vue';
import LabelsBulkInfoModal from '@/components/LabelsBulkInfoModal.vue';
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { useApi } from '@/composables/useApi';

const { apiFetch } = useApi();
const labels = ref([]);
const isConfirmModalOpen = ref(false);
const labelToDelete = ref(null);
const selectedRowId = ref(null);
const API_URL = '/labels';
const isModalOpen = ref(false);
const currentLabel = ref(null);
const apiError = ref(null);
const isBulkModalOpen = ref(false);
const isBulkInfoModalOpen = ref(false);
const isMobileView = ref(window.innerWidth < 768);

const updateMobileView = () => {
  isMobileView.value = window.innerWidth < 768;
};

const sortKey = ref('name');
const sortOrder = ref(1);

const fetchLabels = async () => {
  try {
    const data = await apiFetch(API_URL);
    labels.value = Array.isArray(data) ? data : [];
  } catch (error) {
    console.error('Erreur:', error);
    labels.value = [];
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

const sortedLabels = computed(() => {
  const dataToSort = Array.isArray(labels.value) ? labels.value : [];
  return [...dataToSort].sort((a, b) => {
    const aValue = String(a[sortKey.value] || '').toLowerCase();
    const bValue = String(b[sortKey.value] || '').toLowerCase();
    if (aValue < bValue) return -1 * sortOrder.value;
    if (aValue > bValue) return 1 * sortOrder.value;
    return 0;
  });
});

const openModal = (label = null) => {
  apiError.value = null;
  currentLabel.value = label ? { ...label } : null;
  isModalOpen.value = true;
};

const closeModal = () => {
  apiError.value = null;
  isModalOpen.value = false;
  currentLabel.value = null;
};

const handleSave = (savedLabel) => {
  const existingIndex = labels.value.findIndex((l) => l.id === savedLabel.id);
  if (existingIndex !== -1) {
    labels.value[existingIndex] = savedLabel;
  } else {
    labels.value.push(savedLabel);
  }
};

const prepareDelete = (label) => {
  labelToDelete.value = label;
  isConfirmModalOpen.value = true;
};

const executeDelete = async () => {
  if (!labelToDelete.value) return;
  const id = labelToDelete.value.id;
  isConfirmModalOpen.value = false;
  try {
    await apiFetch(`${API_URL}/${id}`, { method: 'DELETE' });
    labels.value = labels.value.filter((l) => l.id !== id);
    if (selectedRowId.value === id) selectedRowId.value = null;
    labelToDelete.value = null;
  } catch (error) {
    alert(`Échec: ${error.message}`);
  }
};

const handleRowClick = (label) => {
  selectedRowId.value = selectedRowId.value === label.id ? null : label.id;
};

const handleRowDoubleClick = (label) => {
  selectedRowId.value = label.id;
  openModal(label);
};

const handleBulkLabelUpdated = ({ id, description }) => {
  const existingIndex = labels.value.findIndex((l) => l.id === id);
  if (existingIndex !== -1) {
    labels.value[existingIndex] = { ...labels.value[existingIndex], description };
  }
};

const handleBulkInfoUpdated = ({ id, country_id, countryname, founding_year, website }) => {
  const existingIndex = labels.value.findIndex((l) => l.id === id);
  if (existingIndex !== -1) {
    labels.value[existingIndex] = { ...labels.value[existingIndex], country_id, countryname, founding_year, website };
  }
};

onMounted(() => {
  fetchLabels();
  window.addEventListener('resize', updateMobileView);
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateMobileView);
});
</script>

<template>
  <div class="settings-labels-view">
    <header class="page-header">
      <div class="header-wrapper">
        <div class="header-top-row">
          <div class="title-section">
            <button @click="$router.push('/dashboard')" class="back-button ghost-btn">
              <span class="icon">🏠</span>
            </button>
            <span class="title-icon">🏷️</span>
            <h1>Labels</h1>
          </div>
          <div class="toolbar">
            <button @click="isBulkModalOpen = true" class="ghost-btn bulk-update-button">
              <span class="icon">🔄</span>
              Mettre à jour les descriptions
            </button>
            <button @click="isBulkInfoModalOpen = true" class="ghost-btn bulk-update-button">
              <span class="icon">🔄</span>
              Mettre à jour pays / année / site web
            </button>
            <button @click="openModal()" class="primary-btn add-button">
              <span class="icon">➕</span>
              Ajouter un Label
            </button>
          </div>
        </div>
        <p class="subtitle">Gérez les labels (maisons de disques) de votre collection.</p>
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
              <th class="country-column sortable" @click="sortBy('countryname')">
                <div class="sortable-content">
                  <span class="header-text">PAYS</span>
                  <span class="sort-icon" v-html="getSortIcon('countryname')"></span>
                </div>
              </th>
              <th class="year-column sortable" @click="sortBy('founding_year')">
                <div class="sortable-content">
                  <span class="header-text">ANNÉE</span>
                  <span class="sort-icon" v-html="getSortIcon('founding_year')"></span>
                </div>
              </th>
              <th class="website-column sortable" @click="sortBy('website')">
                <div class="sortable-content">
                  <span class="header-text">SITE WEB</span>
                  <span class="sort-icon" v-html="getSortIcon('website')"></span>
                </div>
              </th>
              <th class="description-column sortable" @click="sortBy('description')">
                <div class="sortable-content">
                  <span class="header-text">DESCRIPTION</span>
                  <span class="sort-icon" v-html="getSortIcon('description')"></span>
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
              v-for="(label, index) in sortedLabels" 
              :key="label.id"
              :class="{ 'odd-row': index % 2 === 0, 'even-row': index % 2 !== 0, 'selected-row': selectedRowId === label.id }"
              @click="handleRowClick(label)"
              @dblclick="handleRowDoubleClick(label)"
            >
              <td class="id-column">{{ label.id }}</td>
              <td class="name-column">{{ label.name || '-' }}</td>
              <td class="country-column">{{ label.countryname || '-' }}</td>
              <td class="year-column">{{ label.founding_year || '-' }}</td>
              <td class="website-column">
                <a v-if="label.website" :href="label.website" target="_blank" rel="noopener noreferrer" @click.stop class="website-link" :title="label.website">🔗 Lien</a>
                <span v-else>-</span>
              </td>
              <td class="description-column" :title="label.description">
                <span class="clamp-text">{{ label.description || '-' }}</span>
              </td>
              <td class="actions-column">
                <div class="action-buttons-container">
                  <button @click.stop="openModal(label)" class="icon-action-btn edit-button">
                    <span class="icon">✏️</span>
                  </button>
                  <button @click.stop="prepareDelete(label)" class="icon-action-btn icon-action-btn-danger delete-button">
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
          v-for="label in sortedLabels" 
          :key="label.id"
          class="card-item"
          :class="{ 'selected-row': selectedRowId === label.id }"
        >
          <div class="card-main" @click="handleRowClick(label)" @dblclick="handleRowDoubleClick(label)">
            <div class="card-content">
              <div class="card-id">
                <span class="card-label">ID:</span>
                <span class="card-value">{{ label.id }}</span>
              </div>
              <div class="card-name">
                <span class="card-label">Nom:</span>
                <span class="card-value">{{ label.name || '-' }}</span>
              </div>
              <div class="card-country">
                <span class="card-label">Pays:</span>
                <span class="card-value">{{ label.countryname || '-' }}</span>
              </div>
              <div class="card-year">
                <span class="card-label">Année:</span>
                <span class="card-value">{{ label.founding_year || '-' }}</span>
              </div>
              <div class="card-website">
                <span class="card-label">Site web:</span>
                <a v-if="label.website" :href="label.website" target="_blank" rel="noopener noreferrer" @click.stop class="card-value website-link">{{ label.website }}</a>
                <span v-else class="card-value">-</span>
              </div>
              <div class="card-description">
                <span class="card-label">Description:</span>
                <span class="card-value clamp-text">{{ label.description || '-' }}</span>
              </div>
            </div>
          </div>
          <div class="card-actions-bottom">
            <button @click.stop="openModal(label)" class="action-button-mobile edit-button">
              <span class="icon">✏️</span>
              <span>Modifier</span>
            </button>
            <button @click.stop="prepareDelete(label)" class="action-button-mobile delete-button">
              <span class="icon">🗑️</span>
              <span>Supprimer</span>
            </button>
          </div>
        </div>
      </div>

      <div v-if="!labels || labels.length === 0" class="empty-table-message-standalone">
        Aucun label défini.
      </div>
      <div v-else class="table-summary-wrapper-standalone">
        <div class="summary-box">Nombre : {{ labels.length }}</div>
      </div>
    </div>

    <LabelsModal :is-open="isModalOpen" :label-data="currentLabel" :api-error="apiError" @close="closeModal" @label-saved="handleSave" />

    <LabelsBulkDescriptionModal
      :is-open="isBulkModalOpen"
      :labels="sortedLabels"
      @close="isBulkModalOpen = false"
      @label-updated="handleBulkLabelUpdated"
    />

    <LabelsBulkInfoModal
      :is-open="isBulkInfoModalOpen"
      :labels="sortedLabels"
      @close="isBulkInfoModalOpen = false"
      @label-updated="handleBulkInfoUpdated"
    />

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
          <div v-if="labelToDelete" class="delete-info">
            <h4>Label à supprimer : {{ labelToDelete.name }}</h4>
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
.settings-labels-view {
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
.toolbar { margin: 8px 0 0 0; display: flex; gap: 10px; flex-wrap: wrap; }
.add-button .icon { margin-right: 8px; }
.bulk-update-button .icon { margin-right: 8px; }
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
.data-table .id-column { width: 6%; text-align: center; font-weight: 600; color: var(--text-soft); }
.data-table .name-column { width: 18%; }
.data-table .country-column { width: 12%; }
.data-table .year-column { width: 8%; text-align: center; }
.data-table .website-column { width: 10%; text-align: center; }
.data-table .description-column {
  width: 31%;
  vertical-align: top;
  padding-top: 10px;
  padding-bottom: 10px;
}
.website-link { color: var(--accent); text-decoration: none; white-space: nowrap; }
.website-link:hover { text-decoration: underline; }
.clamp-text {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
}
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
  .card-id, .card-name, .card-country, .card-year, .card-website { display: flex; gap: 8px; }
  .card-label { font-weight: 700; color: var(--text-soft); min-width: 90px; flex-shrink: 0; }
  .card-value { color: var(--text); }
  .card-website .card-value { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  /* Contrairement aux champs courts ci-dessus (ID/Nom), la description est
     un texte long : sur la même ligne que son label, elle déborde et se
     chevauche avec lui. Empilée, avec troncature à 5 lignes — le texte
     complet reste consultable via double-clic sur la carte (ouvre la
     fiche de modification). */
  .card-description { display: flex; flex-direction: column; gap: 4px; }
  .card-description .clamp-text {
    display: -webkit-box;
    -webkit-line-clamp: 5;
    -webkit-box-orient: vertical;
    overflow: hidden;
    line-height: 1.4;
  }
  .card-actions-bottom { display: flex; gap: 8px; padding: 10px 14px; background: rgba(var(--tint-rgb), 0.04); border-top: 1px solid var(--line-soft); }
  .action-button-mobile { flex: 1; display: flex; align-items: center; justify-content: center; gap: 6px; padding: 10px 16px; border: none; border-radius: 10px; font-weight: 600; cursor: pointer; min-height: 44px; }
  .action-button-mobile.edit-button { background: var(--accent); color: white; }
  .action-button-mobile.delete-button { background: var(--negative-text); color: #2a0a0a; }
}
@media (max-width: 768px) {
  .settings-labels-view { padding: 10px; }
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
