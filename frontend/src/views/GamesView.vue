<script setup>
import GameModal from '@/components/GameModal.vue';
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useApi } from '@/composables/useApi';

const { apiFetch } = useApi();
const games = ref([]);
const platforms = ref([]);
const gameGenres = ref([]);
const isLoading = ref(false);

const isConfirmModalOpen = ref(false);
const gameToDelete = ref(null);
const selectedRowId = ref(null);
const isModalOpen = ref(false);
const currentGame = ref(null);
const apiError = ref(null);
const isMobileView = ref(window.innerWidth < 768);

const updateMobileView = () => {
  isMobileView.value = window.innerWidth < 768;
};

const searchQuery = ref('');
const filterPlatform = ref('');
const filterGenre = ref('');
const filterYear = ref('');
const showFilters = ref(true);

const sortKey = ref('title');
const sortOrder = ref('asc');

const currentPage = ref(1);
const itemsPerPage = ref(15);
const pageInputValue = ref('');

const fetchGames = async () => {
  isLoading.value = true;
  try {
    const data = await apiFetch('/games');
    games.value = Array.isArray(data) ? data : [];
  } catch (error) {
    console.error('Erreur chargement jeux:', error);
    games.value = [];
  } finally {
    isLoading.value = false;
  }
};

const fetchLookups = async () => {
  try {
    const [platformData, genreData] = await Promise.all([
      apiFetch('/platforms'),
      apiFetch('/game-genres'),
    ]);
    platforms.value = Array.isArray(platformData) ? platformData : [];
    gameGenres.value = Array.isArray(genreData) ? genreData : [];
  } catch (error) {
    console.error('Erreur chargement listes de référence jeux:', error);
  }
};

const yearOptions = computed(() => {
  const years = new Set();
  for (const g of games.value) {
    if (g.release_year) years.add(g.release_year);
  }
  return Array.from(years).sort((a, b) => b - a);
});

const activeFiltersCount = computed(() => {
  let count = 0;
  if (filterPlatform.value) count += 1;
  if (filterGenre.value) count += 1;
  if (filterYear.value) count += 1;
  return count;
});

const resetFilters = () => {
  searchQuery.value = '';
  filterPlatform.value = '';
  filterGenre.value = '';
  filterYear.value = '';
};

const filteredGames = computed(() => {
  const query = searchQuery.value.trim().toLowerCase();
  return games.value.filter((g) => {
    if (query) {
      const haystack = [g.title, g.platform_name, g.genre_name, g.publisher_name]
        .filter(Boolean)
        .join(' ')
        .toLowerCase();
      if (!haystack.includes(query)) return false;
    }
    if (filterPlatform.value && String(g.platform_id) !== String(filterPlatform.value)) return false;
    if (filterGenre.value && String(g.genre_id) !== String(filterGenre.value)) return false;
    if (filterYear.value && String(g.release_year) !== String(filterYear.value)) return false;
    return true;
  });
});

const sortBy = (key) => {
  if (sortKey.value === key) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc';
  } else {
    sortKey.value = key;
    sortOrder.value = 'asc';
  }
};

const getSortIcon = (key) => {
  if (sortKey.value !== key) return '<span style="font-weight:700">👆 👇</span>';
  return sortOrder.value === 'asc' ? '👆' : '👇';
};

const sortedGames = computed(() => {
  const order = sortOrder.value === 'asc' ? 1 : -1;
  return [...filteredGames.value].sort((a, b) => {
    let aVal = a[sortKey.value];
    let bVal = b[sortKey.value];
    if (sortKey.value === 'release_year') {
      aVal = Number(aVal) || 0;
      bVal = Number(bVal) || 0;
      return (aVal - bVal) * order;
    }
    aVal = String(aVal || '').toLowerCase();
    bVal = String(bVal || '').toLowerCase();
    if (aVal < bVal) return -1 * order;
    if (aVal > bVal) return 1 * order;
    return 0;
  });
});

// ✅ Pagination côté client — même principe que DiscsView.vue (15 par page)
const totalItems = computed(() => sortedGames.value.length);
const totalPages = computed(() => Math.ceil(totalItems.value / itemsPerPage.value) || 1);

const visiblePages = computed(() => {
  const pages = [];
  const maxVisible = 5;
  let start = Math.max(1, currentPage.value - Math.floor(maxVisible / 2));
  let end = Math.min(totalPages.value, start + maxVisible - 1);
  if (end - start + 1 < maxVisible) {
    start = Math.max(1, end - maxVisible + 1);
  }
  for (let i = start; i <= end; i += 1) pages.push(i);
  return pages;
});

const paginatedGames = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage.value;
  return sortedGames.value.slice(startIndex, startIndex + itemsPerPage.value);
});

const goToPage = (page) => {
  if (page >= 1 && page <= totalPages.value) currentPage.value = page;
};
const goToFirstPage = () => goToPage(1);
const goToLastPage = () => goToPage(totalPages.value);
const goToInputPage = () => {
  const page = parseInt(pageInputValue.value, 10);
  if (!isNaN(page) && page >= 1 && page <= totalPages.value) {
    goToPage(page);
  }
  pageInputValue.value = '';
};
const handlePageInputEnter = (event) => {
  if (event.key === 'Enter') goToInputPage();
};

watch([searchQuery, filterPlatform, filterGenre, filterYear], () => {
  currentPage.value = 1;
});

const getImageUrl = (coverPath) => {
  if (!coverPath) return '';
  if (coverPath.startsWith('http')) return coverPath;
  const SERVER_BASE_URL = import.meta.env.VITE_SERVER_BASE_URL || '';
  if (coverPath.startsWith('/')) return `${SERVER_BASE_URL}${coverPath}`;
  return `${SERVER_BASE_URL}/uploads/${coverPath}`;
};

// ✅ Aperçu de la jaquette au survol (desktop) / au tap (mobile) — mêmes
// classes CSS globales que la liste des disques (DiscsView.vue, dernier
// bloc <style> non scopé) : le DOM est créé hors de l'arbre Vue via
// document.createElement, donc un style scoped à ce composant ne
// s'appliquerait pas — on réutilise volontairement les classes déjà
// globales plutôt que d'en dupliquer une variante ici.
const showCoverPreview = (game, event) => {
  if (!game.cover_url) return;
  document.querySelectorAll('.cover-preview').forEach((preview) => preview.remove());
  document.querySelectorAll('.cover-preview-overlay').forEach((overlay) => overlay.remove());

  const coverRect = event.currentTarget.getBoundingClientRect();
  const preview = document.createElement('div');
  preview.className = 'cover-preview';
  preview.setAttribute('role', 'tooltip');
  preview.setAttribute('aria-label', `Aperçu jaquette: ${game.title}`);

  const tooltipWidth = 324;
  const tooltipHeight = 470;
  const isMobile = window.innerWidth < 768;
  let previewLeft, previewTop;

  if (isMobile) {
    previewLeft = (window.innerWidth - tooltipWidth) / 2 + window.scrollX;
    previewTop = Math.max(50, (window.innerHeight - tooltipHeight) / 2) + window.scrollY;
  } else {
    previewLeft = coverRect.right + window.scrollX + 10;
    previewTop = coverRect.top + window.scrollY;
    if (previewLeft + tooltipWidth > window.innerWidth + window.scrollX) {
      previewLeft = coverRect.left + window.scrollX - tooltipWidth - 10;
    }
    if (previewLeft < window.scrollX + 10) {
      previewLeft = window.scrollX + 10;
    }
    const maxTop = window.innerHeight + window.scrollY - tooltipHeight - 20;
    if (previewTop > maxTop) {
      previewTop = Math.max(window.scrollY + 10, maxTop);
    }
    if (previewTop < window.scrollY + 10) {
      previewTop = window.scrollY + 10;
    }
  }

  preview.style.position = 'absolute';
  preview.style.top = `${previewTop}px`;
  preview.style.left = `${previewLeft}px`;
  preview.style.zIndex = '10000';

  const closeButton = isMobile
    ? `<button class="cover-preview-close" onclick="this.closest('.cover-preview').remove(); document.querySelector('.cover-preview-overlay')?.remove();" aria-label="Fermer">×</button>`
    : '';

  preview.innerHTML = `
    ${closeButton}
    <div class="cover-preview-content">
      <div class="cover-preview-image-container">
        <img
          src="${getImageUrl(game.cover_url)}"
          alt="Jaquette: ${game.title}"
          class="cover-preview-image"
          onerror="this.style.display='none'"
        />
      </div>
      <div class="cover-preview-stats">
        <div class="preview-stat-item">
          <span class="preview-label">Qté :</span>
          <span class="preview-value">${game.quantity ?? '-'}</span>
        </div>
        <div class="preview-stat-item">
          <span class="preview-label">Genre :</span>
          <span class="preview-value">${game.genre_name || '-'}</span>
        </div>
        <div class="preview-stat-item">
          <span class="preview-label">Année :</span>
          <span class="preview-value">${game.release_year || '-'}</span>
        </div>
        <div class="preview-stat-item">
          <span class="preview-label">Éditeur :</span>
          <span class="preview-value">${game.publisher_name || '-'}</span>
        </div>
      </div>
      <div class="cover-preview-info">
        <strong>${game.title}</strong>
        ${game.platform_name ? `<div>${game.platform_name}</div>` : ''}
      </div>
    </div>
  `;

  if (isMobile) {
    const overlay = document.createElement('div');
    overlay.className = 'cover-preview-overlay';
    overlay.onclick = () => {
      preview.remove();
      overlay.remove();
    };
    document.body.appendChild(overlay);
  }

  document.body.appendChild(preview);
};

const hideCoverPreview = () => {
  document.querySelectorAll('.cover-preview').forEach((preview) => preview.remove());
};

const openModal = (game = null) => {
  apiError.value = null;
  currentGame.value = game ? { ...game } : null;
  isModalOpen.value = true;
};

const closeModal = () => {
  apiError.value = null;
  isModalOpen.value = false;
  currentGame.value = null;
};

const handleSave = (savedGame) => {
  const existingIndex = games.value.findIndex((g) => g.id === savedGame.id);
  if (existingIndex !== -1) {
    games.value[existingIndex] = savedGame;
  } else {
    games.value.push(savedGame);
  }
  closeModal();
};

const prepareDelete = (game) => {
  gameToDelete.value = game;
  isConfirmModalOpen.value = true;
};

const executeDelete = async () => {
  if (!gameToDelete.value) return;
  const id = gameToDelete.value.id;
  isConfirmModalOpen.value = false;
  try {
    await apiFetch(`/games/${id}`, { method: 'DELETE' });
    games.value = games.value.filter((g) => g.id !== id);
    if (selectedRowId.value === id) selectedRowId.value = null;
    gameToDelete.value = null;
  } catch (error) {
    alert(`Échec: ${error.message}`);
  }
};

const handleRowClick = (game) => {
  selectedRowId.value = selectedRowId.value === game.id ? null : game.id;
};

const handleRowDoubleClick = (game) => {
  selectedRowId.value = game.id;
  openModal(game);
};

const exportData = () => {
  const blob = new Blob([JSON.stringify(games.value, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = `jeux-export-${new Date().toISOString().slice(0, 10)}.json`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
};

onMounted(() => {
  fetchGames();
  fetchLookups();
  window.addEventListener('resize', updateMobileView);
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateMobileView);
});
</script>

<template>
  <div class="games-view">
    <header class="page-header">
      <div class="header-wrapper">
        <div class="header-top-row">
          <div class="title-section">
            <button @click="$router.push('/dashboard')" class="back-button ghost-btn">
              <span class="icon">🏠</span>
            </button>
            <span class="title-icon">🎮</span>
            <h1>Jeux</h1>
          </div>
          <div class="toolbar">
            <button @click="showFilters = !showFilters" class="ghost-btn filter-toggle-button" :disabled="isLoading">
              <span class="icon" aria-hidden="true">👁️</span> {{ showFilters ? 'Masquer Filtres' : 'Afficher Filtres' }}
            </button>
            <button @click="exportData" class="ghost-btn" :disabled="isLoading || games.length === 0">
              <span class="icon">📤</span> Exporter
            </button>
            <button @click="openModal()" class="primary-btn add-button">
              <span class="icon">➕</span>
              Ajouter un Jeu
            </button>
          </div>
        </div>
        <p class="subtitle">Gérez votre collection de jeux vidéo (PS5, Xbox, Switch, Switch 2).</p>
        <p class="user-hint">
          <span>💡</span>
          <strong>Astuce :</strong> Clic pour sélectionner, double-clic pour modifier un jeu
        </p>
      </div>
    </header>

    <div v-if="showFilters" class="panel filters-panel">
      <div class="filters-grid">
        <label class="form-field">
          <span>Rechercher</span>
          <input v-model="searchQuery" type="text" placeholder="Titre, plateforme, genre, éditeur..." />
        </label>
        <label class="form-field">
          <span>Plateforme</span>
          <select v-model="filterPlatform">
            <option value="">Toutes</option>
            <option v-for="p in platforms" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </label>
        <label class="form-field">
          <span>Genre</span>
          <select v-model="filterGenre">
            <option value="">Tous</option>
            <option v-for="g in gameGenres" :key="g.id" :value="g.id">{{ g.name }}</option>
          </select>
        </label>
        <label class="form-field">
          <span>Année</span>
          <select v-model="filterYear">
            <option value="">Toutes</option>
            <option v-for="y in yearOptions" :key="y" :value="y">{{ y }}</option>
          </select>
        </label>
      </div>
      <button v-if="activeFiltersCount > 0 || searchQuery" @click="resetFilters" class="ghost-btn reset-filters-button">
        Réinitialiser ({{ activeFiltersCount }})
      </button>
    </div>

    <div class="content-panel panel">
      <div v-if="!isMobileView" class="table-responsive">
        <table class="data-table">
          <thead>
            <tr>
              <th class="cover-column">JAQUETTE</th>
              <th class="title-column sortable" @click="sortBy('title')">
                <div class="sortable-content">
                  <span class="header-text">TITRE</span>
                  <span class="sort-icon" v-html="getSortIcon('title')"></span>
                </div>
              </th>
              <th class="platform-column sortable" @click="sortBy('platform_name')">
                <div class="sortable-content">
                  <span class="header-text">PLATEFORME</span>
                  <span class="sort-icon" v-html="getSortIcon('platform_name')"></span>
                </div>
              </th>
              <th class="genre-column sortable" @click="sortBy('genre_name')">
                <div class="sortable-content">
                  <span class="header-text">GENRE</span>
                  <span class="sort-icon" v-html="getSortIcon('genre_name')"></span>
                </div>
              </th>
              <th class="publisher-column sortable" @click="sortBy('publisher_name')">
                <div class="sortable-content">
                  <span class="header-text">ÉDITEUR</span>
                  <span class="sort-icon" v-html="getSortIcon('publisher_name')"></span>
                </div>
              </th>
              <th class="year-column sortable" @click="sortBy('release_year')">
                <div class="sortable-content">
                  <span class="header-text">ANNÉE</span>
                  <span class="sort-icon" v-html="getSortIcon('release_year')"></span>
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
              v-for="(game, index) in paginatedGames"
              :key="game.id"
              :class="{ 'odd-row': index % 2 === 0, 'even-row': index % 2 !== 0, 'selected-row': selectedRowId === game.id }"
              @click="handleRowClick(game)"
              @dblclick="handleRowDoubleClick(game)"
            >
              <td class="cover-column">
                <div
                  class="cover-thumbnail-container"
                  @mouseenter="showCoverPreview(game, $event)"
                  @mouseleave="hideCoverPreview"
                  :aria-label="`Jaquette de ${game.title}`"
                >
                  <img v-if="game.cover_url" :src="getImageUrl(game.cover_url)" :alt="`Jaquette: ${game.title}`" class="cover-thumbnail" />
                  <div v-else class="cover-fallback"><span class="disc-icon">🎮</span></div>
                </div>
              </td>
              <td class="title-column">{{ game.title || '-' }}</td>
              <td class="platform-column">{{ game.platform_name || '-' }}</td>
              <td class="genre-column">{{ game.genre_name || '-' }}</td>
              <td class="publisher-column">{{ game.publisher_name || '-' }}</td>
              <td class="year-column">{{ game.release_year || '-' }}</td>
              <td class="actions-column">
                <div class="action-buttons-container">
                  <button @click.stop="openModal(game)" class="icon-action-btn edit-button">
                    <span class="icon">✏️</span>
                  </button>
                  <button @click.stop="prepareDelete(game)" class="icon-action-btn icon-action-btn-danger delete-button">
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
          v-for="game in paginatedGames"
          :key="game.id"
          class="card-item"
          :class="{ 'selected-row': selectedRowId === game.id }"
        >
          <div class="card-main">
            <div class="card-cover" @click.stop="showCoverPreview(game, $event)">
              <img v-if="game.cover_url" :src="getImageUrl(game.cover_url)" :alt="`Jaquette: ${game.title}`" class="cover-thumbnail" />
              <div v-else class="cover-fallback"><span class="disc-icon">🎮</span></div>
            </div>
            <div class="card-content" @click="handleRowClick(game)" @dblclick="handleRowDoubleClick(game)">
              <div class="card-title">{{ game.title || '-' }}</div>
              <div class="card-platform">{{ game.platform_name || '-' }}</div>
              <div class="card-details">
                <span>{{ game.release_year || '-' }}</span>
              </div>
            </div>
          </div>
          <div class="card-actions-bottom">
            <button @click.stop="openModal(game)" class="action-button-mobile edit-button">
              <span class="icon">✏️</span>
              <span>Modifier</span>
            </button>
            <button @click.stop="prepareDelete(game)" class="action-button-mobile delete-button">
              <span class="icon">🗑️</span>
              <span>Supprimer</span>
            </button>
          </div>
        </div>
      </div>

      <div v-if="!isLoading && sortedGames.length === 0" class="empty-table-message-standalone">
        Aucun jeu dans la collection.
      </div>
      <div v-else-if="!isLoading" class="table-summary-wrapper-standalone">
        <div class="summary-box">
          {{ totalItems }} jeu{{ totalItems > 1 ? 'x' : '' }}
        </div>
      </div>

      <div
        v-if="totalPages > 1 && paginatedGames.length > 0"
        class="pagination-bottom"
        role="navigation"
        aria-label="Pagination"
      >
        <div class="pagination-controls">
          <button
            @click="goToFirstPage"
            :disabled="currentPage <= 1 || isLoading"
            class="pagination-button pagination-first"
            aria-label="Première page"
            title="Première page"
          >
            ⏮️
          </button>
          <button
            @click="goToPage(currentPage - 1)"
            :disabled="currentPage <= 1 || isLoading"
            class="pagination-button"
            aria-label="Page précédente"
          >
            ‹ Précédent
          </button>
          <span class="pagination-numbers">
            <button
              v-for="page in visiblePages"
              :key="page"
              @click="goToPage(page)"
              :class="{ active: currentPage === page }"
              class="page-button"
              :aria-label="`Page ${page}`"
              :aria-current="currentPage === page ? 'page' : null"
              :disabled="isLoading"
            >
              {{ page }}
            </button>
          </span>
          <button
            @click="goToPage(currentPage + 1)"
            :disabled="currentPage >= totalPages || isLoading"
            class="pagination-button"
            aria-label="Page suivante"
          >
            Suivant ›
          </button>
          <button
            @click="goToLastPage"
            :disabled="currentPage >= totalPages || isLoading"
            class="pagination-button pagination-last"
            aria-label="Dernière page"
            title="Dernière page"
          >
            ⏭️
          </button>
        </div>
        <div class="pagination-goto">
          <label for="goto-page-games" class="goto-label">Aller à la page :</label>
          <input
            id="goto-page-games"
            v-model="pageInputValue"
            type="number"
            min="1"
            :max="totalPages"
            class="goto-input"
            placeholder="#"
            @keydown.enter="handlePageInputEnter"
            :disabled="isLoading"
            aria-label="Numéro de page"
          />
          <button
            @click="goToInputPage"
            :disabled="isLoading || !pageInputValue"
            class="goto-button"
            aria-label="Aller à la page"
          >
            Aller
          </button>
          <span class="total-pages-info">sur {{ totalPages }}</span>
        </div>
      </div>
    </div>

    <GameModal
      :is-open="isModalOpen"
      :game-data="currentGame"
      :api-error="apiError"
      @close="closeModal"
      @game-saved="handleSave"
      @lookups-changed="fetchLookups"
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
          <div v-if="gameToDelete" class="delete-info">
            <h4>Jeu à supprimer : {{ gameToDelete.title }}</h4>
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
.games-view {
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

.filters-panel { padding: 16px 20px; margin-bottom: 16px; }
.filters-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 0.9rem; }
@media (max-width: 900px) { .filters-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (max-width: 560px) { .filters-grid { grid-template-columns: 1fr; } }
.reset-filters-button { margin-top: 12px; }

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
.data-table .cover-column { width: 60px; text-align: center; }
.data-table .actions-column { width: 90px; text-align: center; }
.data-table .year-column { text-align: right; }
.action-buttons-container { display: flex; gap: 8px; justify-content: center; }
.edit-button .icon { color: var(--accent-soft); }
.delete-button .icon { color: var(--negative-text); }
.sortable { cursor: pointer; }
.sortable:hover { background: rgba(255, 255, 255, 0.1); }
.sortable-content { display: flex; justify-content: space-between; align-items: center; width: 100%; }
.header-text { flex-grow: 1; text-align: center; }
.sort-icon { margin-left: 8px; }

.cover-thumbnail-container {
  width: 42px;
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
  position: relative;
  cursor: pointer;
  border-radius: 6px;
  overflow: hidden;
}
.cover-thumbnail { width: 100%; height: 100%; object-fit: cover; border-radius: 6px; display: block; margin: 0 auto; }
.cover-fallback { width: 100%; height: 100%; border-radius: 6px; background: rgba(var(--tint-rgb), 0.08); display: flex; align-items: center; justify-content: center; margin: 0 auto; }
.disc-icon { font-size: 1.2em; }

.empty-table-message-standalone { text-align: center; padding: 30px; color: var(--text-dim); font-style: italic; border: 2px dashed var(--line); border-radius: 12px; }
.summary-box { text-align: right; font-weight: bold; padding: 12px 15px; border: 1px solid var(--line); border-radius: 12px; color: var(--text-soft); }

@media (max-width: 767px) {
  .cards-container { display: flex; flex-direction: column; gap: 16px; }
  .card-item { background: var(--bg-elevated); border-radius: 14px; box-shadow: var(--shadow); border-left: 4px solid var(--accent); overflow: hidden; }
  .card-item.selected-row { border-left-color: var(--accent-soft); background: rgba(59, 130, 246, 0.08); }
  .card-main { display: flex; gap: 12px; padding: 14px; }
  .card-cover { flex-shrink: 0; cursor: pointer; }
  .card-cover .cover-thumbnail, .card-cover .cover-fallback { width: 56px; height: 56px; }
  .card-content { flex: 1; min-width: 0; cursor: pointer; }
  .card-title { font-weight: 700; color: var(--text); }
  .card-platform { color: var(--text-soft); font-size: 0.9em; margin-top: 2px; }
  .card-details { color: var(--text-dim); font-size: 0.85em; margin-top: 4px; }
  .card-actions-bottom { display: flex; gap: 8px; padding: 10px 14px; background: rgba(var(--tint-rgb), 0.04); border-top: 1px solid var(--line-soft); }
  .action-button-mobile { flex: 1; min-width: 0; display: flex; align-items: center; justify-content: center; gap: 6px; padding: 10px 8px; border: none; border-radius: 10px; font-weight: 600; cursor: pointer; min-height: 44px; }
  .action-button-mobile.edit-button { background: var(--accent); color: white; }
  .action-button-mobile.delete-button { background: var(--negative-text); color: #2a0a0a; }
}
@media (max-width: 768px) {
  .games-view { padding: 10px; }
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

/* ✅ PAGINATION — même style que DiscsView.vue (var(--accent) à la place de
   var(--color-secondary), qui n'existe que localement dans DiscsView.vue) */
.pagination-bottom {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 16px 0;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 16px;
}
.pagination-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.pagination-button {
  padding: 8px 16px;
  border: 1px solid var(--line);
  background: var(--bg-elevated);
  color: var(--text);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-weight: 600;
  min-height: 36px;
}
.pagination-button:hover:not(:disabled) {
  background: rgba(var(--tint-rgb), 0.08);
  border-color: var(--accent);
  transform: translateY(-1px);
}
.pagination-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.pagination-numbers {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}
.page-button {
  padding: 8px 12px;
  border: 1px solid var(--line);
  background: var(--bg-elevated);
  color: var(--text);
  border-radius: 6px;
  cursor: pointer;
  min-width: 40px;
  transition: all 0.2s ease;
  font-weight: 600;
}
.page-button:hover:not(:disabled) {
  background: rgba(var(--tint-rgb), 0.08);
  border-color: var(--accent);
}
.page-button.active {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
}
.page-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.pagination-first,
.pagination-last {
  min-width: 40px;
  font-size: 1.2em;
  padding: 6px 12px;
}
.pagination-goto {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px;
  background: rgba(var(--tint-rgb), 0.04);
  border-radius: 8px;
  border: 1px solid var(--line);
}
.goto-label {
  font-size: 0.9em;
  color: var(--text-soft);
  font-weight: 600;
  white-space: nowrap;
}
.goto-input {
  width: 60px;
  padding: 6px 8px;
  border: 1px solid var(--line);
  background: var(--bg-elevated);
  color: var(--text);
  border-radius: 4px;
  text-align: center;
  font-size: 0.95em;
  font-weight: 600;
  transition: all 0.2s;
}
.goto-input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 2px rgba(96, 165, 250, 0.15);
}
.goto-input:disabled {
  background: rgba(var(--tint-rgb), 0.02);
  cursor: not-allowed;
}
.goto-input::-webkit-inner-spin-button,
.goto-input::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
.goto-input[type=number] {
  -moz-appearance: textfield;
}
.goto-button {
  padding: 6px 16px;
  background: var(--accent-soft);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.9em;
  transition: all 0.2s;
  white-space: nowrap;
}
.goto-button:hover:not(:disabled) {
  filter: brightness(1.1);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(var(--tint-rgb), 0.15);
}
.goto-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.total-pages-info {
  font-size: 0.9em;
  color: var(--text-dim);
  font-weight: 600;
  white-space: nowrap;
}
</style>
