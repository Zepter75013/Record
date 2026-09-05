<script setup>
// ======================================================================
// records-manager/frontend/src/views/ArtistsView.vue
// ======================================================================
import ArtistsModal from '@/components/ArtistsModal.vue';
import ArtistsBulkBiographyModal from '@/components/ArtistsBulkBiographyModal.vue';
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { useApi } from '@/composables/useApi';

const { apiFetch } = useApi();

// États principaux
const artists = ref([]);
const isConfirmModalOpen = ref(false);
const artistToDelete = ref(null);
const selectedRowId = ref(null);

// API URL
const API_URL = '/artists';

// Modale
const isModalOpen = ref(false);
const currentArtist = ref(null);
const apiError = ref(null);
const isBulkModalOpen = ref(false);

// **DÉTECTION RESPONSIVE**
const isMobileView = ref(window.innerWidth < 768);

const updateMobileView = () => {
  isMobileView.value = window.innerWidth < 768;
};

// Tri
const sortKey = ref('name');
const sortOrder = ref(1);

// ======================================================================
// FONCTIONS
// ======================================================================

// Charger les artistes
const fetchArtists = async () => {
  console.log('🔄 [ArtistsView] fetchArtists appelé');
  try {
    const data = await apiFetch(API_URL);
    console.log('✅ [ArtistsView] Artistes récupérés:', data);
    artists.value = Array.isArray(data) ? data : [];
  } catch (error) {
    console.error('❌ [ArtistsView] Erreur récupération:', error);
    artists.value = [];
  }
};

// Tri
const sortBy = (key) => {
  console.log('🔍 [ArtistsView] Tri par:', key);
  if (sortKey.value === key) {
    sortOrder.value *= -1;
  } else {
    sortKey.value = key;
    sortOrder.value = 1;
  }
};

const getSortIcon = (key) => {
  if (sortKey.value !== key) {
    return '<span style="font-weight:700">A-Z</span>';
  }
  return sortOrder.value === 1 ? '▲' : '▼';
};

// Artistes triés
const sortedArtists = computed(() => {
  const dataToSort = Array.isArray(artists.value) ? artists.value : [];
  return [...dataToSort].sort((a, b) => {
    const aValue = String(a[sortKey.value] || '').toLowerCase();
    const bValue = String(b[sortKey.value] || '').toLowerCase();
    if (aValue < bValue) return -1 * sortOrder.value;
    if (aValue > bValue) return 1 * sortOrder.value;
    return 0;
  });
});

// Ouvrir modale
const openModal = (artist = null) => {
  console.log('📝 [ArtistsView] openModal appelé avec:', artist);
  apiError.value = null;
  currentArtist.value = artist ? { ...artist } : null;
  isModalOpen.value = true;
};

// Fermer modale
const closeModal = () => {
  console.log('📝 [ArtistsView] closeModal appelé');
  apiError.value = null;
  isModalOpen.value = false;
  currentArtist.value = null;
};

// Sauvegarder
const handleSave = (savedArtist) => {
  console.log('🎯 [ArtistsView] handleSave appelé avec:', savedArtist);
  const existingIndex = artists.value.findIndex((a) => a.id === savedArtist.id);
  if (existingIndex !== -1) {
    artists.value[existingIndex] = savedArtist;
    if (selectedRowId.value === savedArtist.id) {
      selectedRowId.value = savedArtist.id;
    }
  } else {
    if (!Array.isArray(artists.value)) {
      artists.value = [];
    }
    artists.value.push(savedArtist);
  }
};

// Préparer suppression
const prepareDelete = (artist) => {
  console.log('🗑️ [ArtistsView] Préparation suppression:', artist);
  artistToDelete.value = artist;
  isConfirmModalOpen.value = true;
};

// Exécuter suppression
const executeDelete = async () => {
  if (!artistToDelete.value) return;
  const id = artistToDelete.value.id;
  isConfirmModalOpen.value = false;

  try {
    await apiFetch(`${API_URL}/${id}`, { method: 'DELETE' });
    artists.value = artists.value.filter((a) => a.id !== id);
    if (selectedRowId.value === id) {
      selectedRowId.value = null;
    }
    artistToDelete.value = null;
  } catch (error) {
    console.error('❌ [ArtistsView] Erreur suppression:', error);
    alert(`Échec suppression: ${error.message}`);
  }
};

// Sélection ligne
const handleRowClick = (artist) => {
  selectedRowId.value = selectedRowId.value === artist.id ? null : artist.id;
};

const handleRowDoubleClick = (artist) => {
  selectedRowId.value = artist.id;
  openModal(artist);
};

const handleBulkArtistUpdated = ({ id, biography }) => {
  const existingIndex = artists.value.findIndex((a) => a.id === id);
  if (existingIndex !== -1) {
    artists.value[existingIndex] = { ...artists.value[existingIndex], biography };
  }
};

// Lifecycle
onMounted(() => {
  fetchArtists();
  window.addEventListener('resize', updateMobileView);
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateMobileView);
});
</script>

<template>
  <div class="settings-artists-view">
    <!-- HEADER -->
    <header class="page-header">
      <div class="header-wrapper">
        <div class="header-top-row">
          <div class="title-section">
            <button @click="$router.push('/dashboard')" class="back-button ghost-btn">
              <span class="icon">🏠</span>
            </button>
            <span class="title-icon">🎤</span>
            <h1>Artistes</h1>
          </div>
          <div class="toolbar">
            <button @click="isBulkModalOpen = true" class="ghost-btn bulk-update-button">
              <span class="icon">🔄</span>
              Mettre à jour les biographies
            </button>
            <button @click="openModal()" class="primary-btn add-button">
              <span class="icon">➕</span>
              Ajouter un Artiste
            </button>
          </div>
        </div>
        <p class="subtitle">Gérez les artistes de votre collection de disques.</p>
        <p class="user-hint">
          <span>💡</span>
          <strong>Astuce :</strong> Clic pour sélectionner, double-clic pour modifier un artiste
        </p>
      </div>
    </header>

    <!-- CONTENU -->
    <div class="content-panel panel">
      <!-- TABLEAU Desktop/Tablette -->
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
              <th class="biography-column sortable" @click="sortBy('biography')">
                <div class="sortable-content">
                  <span class="header-text">BIOGRAPHIE</span>
                  <span class="sort-icon" v-html="getSortIcon('biography')"></span>
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
              v-for="(artist, index) in sortedArtists"
              :key="artist.id"
              :class="{
                'odd-row': index % 2 === 0,
                'even-row': index % 2 !== 0,
                'selected-row': selectedRowId === artist.id
              }"
              @click="handleRowClick(artist)"
              @dblclick="handleRowDoubleClick(artist)"
            >
              <td class="id-column">{{ artist.id }}</td>
              <td class="name-column">{{ artist.name || '-' }}</td>
              <td class="country-column">{{ artist.countryname || '-' }}</td>
              <td class="biography-column" :title="artist.biography">{{ artist.biography || '-' }}</td>
              <td class="actions-column">
                <div class="action-buttons-container">
                  <button @click.stop="openModal(artist)" class="icon-action-btn edit-button">
                    <span class="icon">✏️</span>
                  </button>
                  <button @click.stop="prepareDelete(artist)" class="icon-action-btn icon-action-btn-danger delete-button">
                    <span class="icon">🗑️</span>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- CARTES Mobile -->
      <div v-else class="cards-container">
        <div
          v-for="artist in sortedArtists"
          :key="artist.id"
          class="card-item"
          :class="{ 'selected-row': selectedRowId === artist.id }"
        >
          <div class="card-main" @click="handleRowClick(artist)" @dblclick="handleRowDoubleClick(artist)">
            <div class="card-content">
              <div class="card-id">
                <span class="card-label">ID:</span>
                <span class="card-value">{{ artist.id }}</span>
              </div>
              <div class="card-name">
                <span class="card-label">Nom:</span>
                <span class="card-value">{{ artist.name || '-' }}</span>
              </div>
              <div class="card-country">
                <span class="card-label">Pays:</span>
                <span class="card-value">{{ artist.countryname || '-' }}</span>
              </div>
              <div class="card-biography">
                <span class="card-label">Biographie:</span>
                <span class="card-value">{{ artist.biography || '-' }}</span>
              </div>
            </div>
          </div>
          <div class="card-actions-bottom">
            <button @click.stop="openModal(artist)" class="action-button-mobile edit-button">
              <span class="icon">✏️</span>
              <span>Modifier</span>
            </button>
            <button @click.stop="prepareDelete(artist)" class="action-button-mobile delete-button">
              <span class="icon">🗑️</span>
              <span>Supprimer</span>
            </button>
          </div>
        </div>
      </div>

      <!-- MESSAGE VIDE -->
      <div v-if="!artists || artists.length === 0" class="empty-table-message-standalone">
        Aucun artiste défini. Utilisez le bouton "Ajouter un Artiste" pour commencer.
      </div>

      <!-- RÉSUMÉ -->
      <div v-else class="table-summary-wrapper-standalone">
        <div class="summary-box">
          Nombre d'enregistrements : {{ artists.length }}
        </div>
      </div>
    </div>

    <!-- MODALE PRINCIPALE -->
    <ArtistsModal
      :is-open="isModalOpen"
      :artist-data="currentArtist"
      :api-error="apiError"
      @close="closeModal"
      @artist-saved="handleSave"
    />

    <ArtistsBulkBiographyModal
      :is-open="isBulkModalOpen"
      :artists="sortedArtists"
      @close="isBulkModalOpen = false"
      @artist-updated="handleBulkArtistUpdated"
    />

    <!-- MODALE CONFIRMATION SUPPRESSION -->
    <Teleport to="body">
      <div v-if="isConfirmModalOpen" class="modal-overlay" @click.self="isConfirmModalOpen = false">
        <div class="modal-card">
          <div class="modal-header">
            <div class="header-title-container">
              <span class="title-icon">🗑️</span>
              <div>
                <h2>Confirmer la Suppression</h2>
                <p class="header-subtitle">Cette action est irréversible</p>
              </div>
            </div>
            <button class="icon-action-btn" @click="isConfirmModalOpen = false" aria-label="Fermer">
              <span>&times;</span>
            </button>
          </div>
          <div v-if="artistToDelete" class="delete-info">
            <h4>📋 Artiste à supprimer</h4>
            <div class="artist-details">
              <div class="detail-row">
                <strong>Nom :</strong> {{ artistToDelete.name }}
              </div>
              <div class="detail-row">
                <strong>Pays :</strong> {{ artistToDelete.countryname || '—' }}
              </div>
              <div class="detail-row">
                <strong>ID :</strong> {{ artistToDelete.id }}
              </div>
            </div>
          </div>
          <div class="modal-actions">
            <button @click="isConfirmModalOpen = false" class="ghost-btn">Annuler</button>
            <button @click="executeDelete" class="danger-btn">🗑️ Supprimer Définitivement</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.settings-artists-view {
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
.data-table { width: 100%; border-collapse: separate; border-spacing: 0; font-size: 0.9em; border-radius: 12px; overflow: hidden; min-width: 600px; }
.data-table thead { background-color: var(--table-header-color); }
.data-table thead th { padding: 12px 15px; background: var(--table-header-color); color: white; text-align: center; height: 50px; position: relative; white-space: nowrap; }
.data-table thead th::after { content: ''; position: absolute; right: 0; top: 0; height: 100%; width: 1px; background-color: rgba(255, 255, 255, 0.2); }
.data-table thead th:last-child::after { display: none; }
.data-table tbody tr { transition: all 0.2s ease; cursor: pointer; height: 40px; }
.data-table tbody tr:hover { background: rgba(var(--tint-rgb), 0.06); }
.data-table tbody tr.selected-row { background: rgba(59, 130, 246, 0.12) !important; border-left: 4px solid var(--accent) !important; }
.data-table td { padding: 6px 15px; border-bottom: 1px solid var(--line-soft); text-align: left; font-size: 0.85em; height: 40px; color: var(--text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.data-table .id-column { width: 8%; text-align: center; font-weight: 600; color: var(--text-soft); }
.data-table .name-column { width: 20%; }
.data-table .country-column { width: 17%; }
.data-table .biography-column { width: 40%; }
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
  .card-id, .card-name, .card-country, .card-biography { display: flex; gap: 8px; }
  .card-label { font-weight: 700; color: var(--text-soft); min-width: 60px; }
  .card-value { color: var(--text); }
  .card-actions-bottom { display: flex; gap: 8px; padding: 10px 14px; background: rgba(var(--tint-rgb), 0.04); border-top: 1px solid var(--line-soft); }
  .action-button-mobile { flex: 1; display: flex; align-items: center; justify-content: center; gap: 6px; padding: 10px 16px; border: none; border-radius: 10px; font-weight: 600; cursor: pointer; min-height: 44px; }
  .action-button-mobile.edit-button { background: var(--accent); color: white; }
  .action-button-mobile.delete-button { background: var(--negative-text); color: #2a0a0a; }
}
@media (max-width: 768px) {
  .settings-artists-view { padding: 10px; }
  .header-top-row { flex-direction: column; align-items: flex-start; }
  .toolbar { width: 100%; margin-top: 12px; }
  .add-button { width: 100%; justify-content: center; }
}

.header-title-container {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.header-subtitle {
  margin: 4px 0 0;
  color: var(--text-soft);
  font-size: 0.88em;
}

.delete-info {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.28);
  border-radius: 14px;
  padding: 16px;
  margin-bottom: 20px;
  color: var(--text);
}

.delete-info h4 {
  margin: 0 0 12px 0;
  color: var(--negative-text);
}

.artist-details {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid var(--line-soft);
}

.detail-row:last-child {
  border-bottom: none;
}
</style>
