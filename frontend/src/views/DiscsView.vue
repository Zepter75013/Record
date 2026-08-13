<script setup>
// records-manager/frontend/src/views/DiscsView.vue
import DiscsModal from '@/components/DiscsModal/DiscsModal.vue';
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue';
import { useApi } from '@/composables/useApi';
import { formatCurrency } from '@/utils/format';
const { apiFetch } = useApi();
// ✅ ÉTATS PRINCIPAUX
const discs = ref([]);
const apiError = ref(null);
const isConfirmModalOpen = ref(false);
const discToDelete = ref(null);
const selectedRowId = ref(null);
const API_URL = '/discs';
// ✅ ÉTATS FILTRES & RECHERCHE
const searchQuery = ref('');
const filters = ref({
genre_id: '',
format_id: '',
country_id: '',
label_id: ''
});
const showFilters = ref(true);
const isClientSideFiltering = ref(false);
// ✅ DEBOUNCE AMÉLIORÉ
const searchDebounceTimer = ref(null);
const DEBOUNCE_DELAY = 300;
// ✅ LISTES FILTRES (initialisation avec tableaux vides)
const genres = ref([]);
const formats = ref([]);
const countries = ref([]);
const labels = ref([]);
const filterOptionsLoaded = ref(false);
// ✅ ÉTATS CHARGEMENT
const isLoading = ref(false);
const isInitialLoad = ref(true);
const isSaving = ref(false);
// ✅ PAGINATION
const currentPage = ref(1);
const itemsPerPage = ref(15);
const totalItems = ref(0);
const pageInputValue = ref('');
// ✅ MODALE
const isModalOpen = ref(false);
const currentDisc = ref(null);
// ✅ LOG LEVEL
const LOG_LEVEL = import.meta.env.VITE_LOG_LEVEL || 'info';
// ✅ LOGGER STRUCTURÉ
const logger = {
debug: (...args) => LOG_LEVEL === 'debug' && console.debug('[DiscsView]', ...args),
info: (...args) => ['debug', 'info'].includes(LOG_LEVEL) && console.info('[DiscsView]', ...args),
warn: (...args) => console.warn('[DiscsView]', ...args),
error: (...args) => console.error('[DiscsView]', ...args)
};
// 🔧 AJOUT : référence pour le champ de recherche
const searchInputRef = ref(null);
// 🔧 AJOUT : référence pour le conteneur de table (sticky columns)
const tableContainer = ref(null);
// ✅ Vue mobile
const isMobileView = ref(window.innerWidth < 768);
const updateMobileView = () => {
isMobileView.value = window.innerWidth < 768;
};
// 🆕 GESTION DES COLONNES (masquage/redimensionnement)
const showColumnMenu = ref(false);
const columnMenuRef = ref(null);
// Configuration par défaut des colonnes
const defaultColumns = [
{ key: 'cover', label: 'JACKET', labelShort: 'JACKET', labelLong: 'JACKET', visible: true, width: 100, minWidth: 100, resizable: false, sticky: true, stickyLeft: 0 },
{ key: 'title', label: 'TITRE', labelShort: 'TITRE', labelLong: 'TITRE', visible: true, width: 220, minWidth: 180, resizable: true, sticky: false },
{ key: 'artist', label: 'ARTISTE', labelShort: 'ARTISTE', labelLong: 'ARTISTE', visible: true, width: 200, minWidth: 160, resizable: true, sticky: false },
{ key: 'quantity', label: 'QTÉ', labelShort: 'QTÉ', labelLong: 'NB DISC', visible: true, width: 120, minWidth: 120, resizable: true, sticky: false },
{ key: 'genre', label: 'GENRE', labelShort: 'GENRE', labelLong: 'GENRE', visible: true, width: 130, minWidth: 120, resizable: true, sticky: false },
{ key: 'year', label: 'ANNÉE', labelShort: 'ANNÉE', labelLong: 'ANNÉE', visible: true, width: 130, minWidth: 130, resizable: true, sticky: false },
{ key: 'country', label: 'PAYS', labelShort: 'PAYS', labelLong: 'PAYS', visible: true, width: 120, minWidth: 110, resizable: true, sticky: false },
{ key: 'format', label: 'FORMAT', labelShort: 'FORMAT', labelLong: 'FORMAT', visible: true, width: 130, minWidth: 120, resizable: true, sticky: false },
{ key: 'label', label: 'LABEL', labelShort: 'LABEL', labelLong: 'LABEL', visible: true, width: 170, minWidth: 140, resizable: true, sticky: false },
{ key: 'barcode', label: 'BARCODE', labelShort: 'BARCODE', labelLong: 'CODE-BARRES', visible: true, width: 150, minWidth: 140, resizable: true, sticky: false },
{ key: 'price', label: 'PRIX', labelShort: 'PRIX', labelLong: 'PRIX', visible: true, width: 100, minWidth: 90, resizable: true, sticky: false },
{ key: 'actions', label: 'ACTIONS', labelShort: 'ACTIONS', labelLong: 'ACTIONS', visible: true, width: 150, minWidth: 150, resizable: false, sticky: true, stickyRight: 0 }
];
// Charger la configuration depuis localStorage
const loadColumnsConfig = () => {
try {
const saved = localStorage.getItem('discsview-columns');
if (saved) {
const parsed = JSON.parse(saved);
return defaultColumns.map(col => {
const savedCol = parsed.find(c => c.key === col.key);
return savedCol ? { ...col, ...savedCol } : col;
});
}
} catch (err) {
logger.warn('Erreur chargement config colonnes:', err);
}
return defaultColumns;
};
const columns = ref(loadColumnsConfig());
// Sauvegarder la configuration
const saveColumnsConfig = () => {
try {
localStorage.setItem('discsview-columns', JSON.stringify(columns.value));
logger.debug('Configuration colonnes sauvegardée');
} catch (err) {
logger.warn('Erreur sauvegarde config colonnes:', err);
}
};
// Computed pour filtrer les colonnes visibles
const visibleColumns = computed(() => columns.value.filter(col => col.visible));
// 🆕 GESTION ADAPTATIVE DES LABELS (sidebar réduite/étendue)
const windowWidth = ref(window.innerWidth);
const SIDEBAR_COLLAPSED_BREAKPOINT = 1200; // Largeur en dessous de laquelle on considère que la sidebar est réduite

// Computed pour déterminer si on utilise les labels courts ou longs
const useShortLabels = computed(() => {
	return windowWidth.value < SIDEBAR_COLLAPSED_BREAKPOINT;
});

// Fonction pour mettre à jour les labels des colonnes selon l'espace disponible
const updateColumnLabels = () => {
	const shouldUseShort = useShortLabels.value;
	columns.value = columns.value.map(col => ({
		...col,
		label: shouldUseShort ? col.labelShort : col.labelLong
	}));
	logger.debug('Labels mis à jour:', shouldUseShort ? 'courts' : 'longs');
};

// Fonction pour gérer le redimensionnement de la fenêtre
const handleWindowResize = () => {
	windowWidth.value = window.innerWidth;
	updateMobileView();
	updateColumnLabels();
};
// Toggle visibilité d'une colonne
const toggleColumn = (key) => {
const col = columns.value.find(c => c.key === key);
if (col && col.key !== 'actions') { // La colonne actions reste toujours visible
col.visible = !col.visible;
saveColumnsConfig();
}
};
// 🔒 Fonctions pour colonnes figées (sticky)
const canToggleSticky = (key) => {
  if (key === 'actions') return true;
  const col = columns.value.find(c => c.key === key);
  if (!col || !col.visible) return false;
  const visibleCols = columns.value.filter(c => c.visible && c.key !== 'actions');
  const currentIdx = visibleCols.findIndex(c => c.key === key);
  if (currentIdx === -1) return false;
  if (col.sticky) {
    for (let i = currentIdx + 1; i < visibleCols.length; i++) {
      if (visibleCols[i].sticky) return false;
    }
    return true;
  } else {
    if (currentIdx === 0) return true;
    for (let i = 0; i < currentIdx; i++) {
      if (!visibleCols[i].sticky) return false;
    }
    return true;
  }
};

const toggleSticky = (key) => {
  if (!canToggleSticky(key)) {
    const col = columns.value.find(c => c.key === key);
    if (col?.sticky) {
      alert(`⚠️ Impossible de déverrouiller\n\nDéverrouillez d'abord les colonnes à droite.`);
    } else {
      alert(`⚠️ Impossible de verrouiller\n\nVerrouillez d'abord TOUTES les colonnes à gauche.`);
    }
    return;
  }
  const col = columns.value.find(c => c.key === key);
  if (col) {
    col.sticky = !col.sticky;
    updateStickyPositions();
    saveColumnsConfig();
  }
};

// 📏 Auto-ajuster la largeur d'une colonne (double-clic Excel-like)
const autoFitColumn = (columnKey) => {
  const col = columns.value.find(c => c.key === columnKey);
  if (!col) return;
  
  const measureEl = document.createElement('span');
  measureEl.style.cssText = 'visibility:hidden;position:absolute;white-space:nowrap;font-size:14px';
  document.body.appendChild(measureEl);
  
  let maxWidth = 0;
  
  // Mesurer en-tête
  measureEl.textContent = col.label;
  maxWidth = Math.max(maxWidth, measureEl.offsetWidth + 60);
  
  // Mesurer contenu
  discs.value.forEach(disc => {
    let text = '';
    switch(columnKey) {
      case 'title': text = disc.title || ''; break;
      case 'artist': text = disc.artistname || ''; break;
      case 'quantity': text = formatQuantity(disc.quantity); break;
      case 'genre': text = disc.genrename || ''; break;
      case 'year': text = disc.year || ''; break;
      case 'country': text = disc.countryname || ''; break;
      case 'format': text = disc.formatname || ''; break;
      case 'label': text = disc.labelname || ''; break;
      case 'barcode': text = disc.barcode || ''; break;
      case 'price': text = formatPrice(disc.price); break;
    }
    measureEl.textContent = text;
    maxWidth = Math.max(maxWidth, measureEl.offsetWidth + 16);
  });
  
  document.body.removeChild(measureEl);
  
  const minWidth = col.minWidth || 100;
  col.width = Math.max(minWidth, Math.min(maxWidth, 600));
  
  if (col.sticky) updateStickyPositions();
  saveColumnsConfig();
  logger.info(`Auto-fit "${col.label}": ${col.width}px`);
};
const updateStickyPositions = () => {
  let leftOffset = 0;
  let rightOffset = 0;
  
  // ✅ CORRECTION ICI : toutes les colonnes sticky (sauf actions) à gauche
  const leftStickyCols = columns.value.filter(
    c => c.sticky && c.visible && c.key !== 'actions'
  );
  
  logger.debug('updateStickyPositions: Colonnes sticky à gauche:', leftStickyCols.map(c => c.key));
  
  for (const col of leftStickyCols) {
    col.stickyLeft = leftOffset;
    leftOffset += col.width;
    delete col.stickyRight; // s'assurer qu'une colonne n'est pas des 2 côtés
    logger.debug(`  → ${col.key}: left=${col.stickyLeft}px, width=${col.width}px`);
  }

  // Actions à droite
  const actionsCol = columns.value.find(c => c.key === 'actions');
  if (actionsCol?.sticky && actionsCol.visible) {
    actionsCol.stickyRight = 0;
    logger.debug('updateStickyPositions: Actions sticky à droite');
  }

  nextTick(() => {
    initializeStickyStyles();
  });
};
// 🆕 CACHE pour les éléments sticky (évite querySelectorAll à chaque scroll)
const stickyElementsCache = ref({
left: [],  // { element, offset }
right: []  // { element }
});

// 🆕 FONCTION - Construire le cache des éléments sticky
const buildStickyCache = () => {
if (!tableContainer.value) return;

const leftElements = [];
const rightElements = [];

// 🎨 Couleurs thème-aware pour les colonnes figées (lit les custom properties
// au lieu de hex fixes, pour suivre le thème actif clair/sombre)
const rootStyles = getComputedStyle(document.documentElement);
const stickyBaseBg = rootStyles.getPropertyValue('--bg-elevated').trim() || 'var(--bg-elevated)';
const stickyTintRgb = rootStyles.getPropertyValue('--tint-rgb').trim() || '255, 255, 255';

columns.value.forEach(col => {
if (!col.sticky || !col.visible) return;

const selector = `th.${col.key}-column, td.${col.key}-column`;
const cells = tableContainer.value.querySelectorAll(selector);

cells.forEach(cell => {
// Appliquer les styles de base
cell.style.position = 'relative';
cell.style.zIndex = col.key === 'cover' || col.key === 'actions' ? '10' : '5';

// Background selon le type (opaque, thème-aware)
if (cell.tagName === 'TH') {
cell.style.background = stickyBaseBg;
} else {
const row = cell.closest('tr');
const tintAlpha = row?.classList.contains('even-row') ? 0.04 : 0.02;
cell.style.background = `linear-gradient(rgba(${stickyTintRgb}, ${tintAlpha}), rgba(${stickyTintRgb}, ${tintAlpha})), ${stickyBaseBg}`;
}

if (col.stickyLeft !== undefined) {
leftElements.push({ element: cell, offset: col.stickyLeft });
} else if (col.stickyRight !== undefined) {
rightElements.push({ element: cell });
}
});
});

stickyElementsCache.value = { left: leftElements, right: rightElements };
logger.debug(`buildStickyCache: ${leftElements.length} left, ${rightElements.length} right`);
};

// 🆕 FONCTION OPTIMISÉE - Appliquer transforms (utilise le cache, zéro allocation)
const applyStickyStyles = () => {
const container = tableContainer.value;
if (!container) return;

const scrollLeft = container.scrollLeft;
const maxScroll = container.scrollWidth - container.clientWidth;
const scrollRight = maxScroll - scrollLeft;

// ⚡ Utiliser les tableaux directement sans déstructuration
const leftArr = stickyElementsCache.value.left;
const rightArr = stickyElementsCache.value.right;

// Colonnes gauche - string template optimisé
const leftTransform = `translate3d(${scrollLeft}px,0,0)`;
for (let i = 0, len = leftArr.length; i < len; i++) {
leftArr[i].element.style.transform = leftTransform;
}

// Colonnes droite
const rightTransform = `translate3d(${-scrollRight}px,0,0)`;
for (let i = 0, len = rightArr.length; i < len; i++) {
rightArr[i].element.style.transform = rightTransform;
}
};

// 🆕 FONCTION - Initialiser les styles sticky (appelée une seule fois ou quand les données changent)
const initializeStickyStyles = () => {
if (!tableContainer.value) {
logger.warn('initializeStickyStyles: tableContainer non disponible');
return;
}

logger.debug('initializeStickyStyles: Construction du cache et initialisation');

// Construire le cache des éléments
buildStickyCache();

// Appliquer les transforms immédiatement
applyStickyStyles();

logger.info(`initializeStickyStyles: Cache construit avec ${stickyElementsCache.value.left.length + stickyElementsCache.value.right.length} éléments`);
};
const getStickyStyle = (column) => {
if (!column) return {};
// Retourner au minimum la largeur de la colonne
return {
width: column.width ? `${column.width}px` : 'auto'
};
};
// Réinitialiser les colonnes
const resetColumns = () => {
columns.value = JSON.parse(JSON.stringify(defaultColumns));
saveColumnsConfig();
};
// État pour le redimensionnement
const resizingColumn = ref(null);
const resizeStartX = ref(0);
const resizeStartWidth = ref(0);
// Démarrer le redimensionnement
const startResize = (event, column) => {
if (!column.resizable) return;
event.preventDefault();
event.stopPropagation();
resizingColumn.value = column;
resizeStartX.value = event.clientX;
resizeStartWidth.value = column.width;
document.addEventListener('mousemove', handleResize);
document.addEventListener('mouseup', stopResize);
document.body.style.cursor = 'col-resize';
document.body.style.userSelect = 'none';
};
// Gérer le redimensionnement
const handleResize = (event) => {
if (!resizingColumn.value) return;
const delta = event.clientX - resizeStartX.value;
const newWidth = Math.max(50, resizeStartWidth.value + delta);
resizingColumn.value.width = newWidth;
if (resizingColumn.value.sticky) {
// Recalculer les positions sticky et les appliquer
nextTick(() => {
updateStickyPositions();
});
}
};
// Arrêter le redimensionnement
const stopResize = () => {
if (resizingColumn.value) {
saveColumnsConfig();
resizingColumn.value = null;
}
document.removeEventListener('mousemove', handleResize);
document.removeEventListener('mouseup', stopResize);
document.body.style.cursor = '';
document.body.style.userSelect = '';
};
// Fermer le menu des colonnes quand on clique ailleurs
const handleClickOutside = (event) => {
if (columnMenuRef.value && !columnMenuRef.value.contains(event.target)) {
showColumnMenu.value = false;
}
};
// ✅ LIFECYCLE HOOKS
onMounted(() => {
window.addEventListener('resize', handleWindowResize);
document.addEventListener('click', handleClickOutside);
fetchDiscs();
loadFilterOptions();

// ⚡ Fonction de scroll ULTRA-optimisée - appel direct sans RAF
// Le code est assez optimisé pour être appelé à chaque événement scroll
const handleScroll = () => {
applyStickyStyles();
};

// Appliquer les styles sticky après montage du DOM
nextTick(() => {
if (tableContainer.value) {
updateStickyPositions();
// 🆕 Ajouter écouteur scroll optimisé avec passive: true
tableContainer.value.addEventListener('scroll', handleScroll, { passive: true });
} else {
// Fallback sécurisé (au cas où)
setTimeout(() => {
if (tableContainer.value) {
updateStickyPositions();
tableContainer.value.addEventListener('scroll', handleScroll, { passive: true });
}
}, 100);
}
});
});
onBeforeUnmount(() => {
window.removeEventListener('resize', handleWindowResize);
document.removeEventListener('click', handleClickOutside);
stopResize(); // Nettoyer au cas où
});
// ✅ COMPUTED: Filtres actifs
const activeFilters = computed(() => {
const list = [];
if (searchQuery.value) list.push({
type: 'search',
label: `🔍 ${searchQuery.value}`,
value: searchQuery.value
});
if (filters.value.genre_id) {
const g = genres.value.find(x => x.id == filters.value.genre_id);
if (g) list.push({ type: 'genre_id', label: `🎵 ${g.name}`, value: g.id });
}
if (filters.value.format_id) {
const f = formats.value.find(x => x.id == filters.value.format_id);
if (f) list.push({ type: 'format_id', label: `📏 ${f.name}`, value: f.id });
}
if (filters.value.country_id) {
const c = countries.value.find(x => x.id == filters.value.country_id);
if (c) list.push({ type: 'country_id', label: `🌍 ${c.name}`, value: c.id });
}
if (filters.value.label_id) {
const l = labels.value.find(x => x.id == filters.value.label_id);
if (l) list.push({ type: 'label_id', label: `🏷️ ${l.name}`, value: l.id });
}
return list;
});
// ✅ COMPUTED: Pagination
const totalPages = computed(() => Math.ceil(totalItems.value / itemsPerPage.value));
// ✅ COMPUTED: Pages visibles
const visiblePages = computed(() => {
const pages = [];
const maxVisible = 5;
let start = Math.max(1, currentPage.value - Math.floor(maxVisible / 2));
let end = Math.min(totalPages.value, start + maxVisible - 1);
if (end - start + 1 < maxVisible) {
start = Math.max(1, end - maxVisible + 1);
}
for (let i = start; i <= end; i++) {
pages.push(i);
}
return pages;
});
// 🆕 WATCH - Réappliquer les styles sticky quand les données changent
watch(() => discs.value.length, () => {
logger.debug('Watch discs.length: Données changées, réapplication styles sticky');
nextTick(() => {
setTimeout(() => {
if (tableContainer.value) {
initializeStickyStyles();
}
}, 50); // Petit délai pour s'assurer que le DOM est bien rendu
});
}, { flush: 'post' });

// 🆕 WATCH - Réappliquer les styles sticky quand on change de page
watch(() => currentPage.value, () => {
logger.debug('Watch currentPage: Page changée, reconstruction du cache sticky');
nextTick(() => {
setTimeout(() => {
if (tableContainer.value) {
initializeStickyStyles();
}
}, 50);
});
}, { flush: 'post' });

// 🆕 WATCH - Réappliquer les styles sticky quand on change les colonnes
watch(() => columns.value.map(c => ({ key: c.key, sticky: c.sticky, visible: c.visible })), () => {
logger.debug('Watch columns: Configuration colonnes changée');
nextTick(() => {
if (tableContainer.value) {
initializeStickyStyles();
}
});
}, { deep: true });

// ✅ HELPER: URL des images avec cache-buster
const getImageUrl = (coverPath, cacheBuster = null) => {
if (!coverPath) return '/images/default-cover.png';
if (coverPath.startsWith('http')) return coverPath;
// ✅ CORRECTION: URL relative - Nginx fait le proxy
const SERVER_BASE_URL = import.meta.env.VITE_SERVER_BASE_URL || '';
let url;
if (coverPath.startsWith('/')) {
url = `${SERVER_BASE_URL}${coverPath}`;
} else {
url = `${SERVER_BASE_URL}/uploads/${coverPath}`;
}
// Ajouter cache-buster pour forcer le rechargement après modification
if (cacheBuster) {
const timestamp = typeof cacheBuster === 'string' ? new Date(cacheBuster).getTime() : cacheBuster;
url += (url.includes('?') ? '&' : '?') + `t=${timestamp}`;
}
return url;
};
// ✅ HELPER: Gestion erreur image
const handleImageError = (event, disc) => {
event.target.style.display = 'none';
const fallback = event.target.nextElementSibling;
if (fallback) {
fallback.style.display = 'flex';
if (disc?.title) {
const shortTitle = disc.title.length > 10 ? disc.title.substring(0, 10) + '...' : disc.title;
fallback.innerHTML = `
<span class="disc-icon">💿</span>
<span class="cover-fallback-text">${shortTitle}</span>
`;
}
}
};
// ✅ HELPER: Affichage des valeurs
const displayValue = (value) => value || '-';
// Formatage du prix, piloté par les préférences de format d'affichage
const formatPrice = (price) => {
if (price == null || price === '' || isNaN(parseFloat(price))) return '–';
const num = typeof price === 'string' ? parseFloat(price) : price;
return formatCurrency(num);
};
// ✅ NOUVEAU: Calcul du total global (prix × quantité)
// ✅ CORRIGÉ: Total global = somme des prix des albums (prix complet tel qu'affiché, sans × quantity)
const totalValue = computed(() => {
return discs.value.reduce((sum, disc) => {
const price = parseFloat(disc.price) || 0;
return sum + price; // ← quantity ignorée
}, 0);
});
// ✅ NOUVEAU: Calcul du nombre total de disques (incluant quantity)
const totalDiscsCount = computed(() => {
return discs.value.reduce((sum, disc) => {
const qty = parseInt(disc.quantity) || 1;
return sum + qty;
}, 0);
});
// ✅ NOUVEAU: Formatage pour afficher la quantité
const formatQuantity = (quantity) => {
if (quantity == null || quantity === '' || isNaN(parseInt(quantity))) return '1';
return parseInt(quantity).toString();
};
// ✅ FONCTION: Charger les options de filtrage (avec cache)
const loadFilterOptions = async () => {
if (filterOptionsLoaded.value) {
logger.debug('Options de filtrage déjà chargées, utilisation du cache');
return;
}
try {
// Charger toutes les options en parallèle
const promises = [
apiFetch('/genres').then(data => data || []).catch(() => []),
apiFetch('/formats').then(data => data || []).catch(() => []),
apiFetch('/countries').then(data => data || []).catch(() => []),
apiFetch('/labels').then(data => data || []).catch(() => [])
];
const [g, f, c, l] = await Promise.all(promises);
genres.value = g;
formats.value = f;
countries.value = c;
labels.value = l;
filterOptionsLoaded.value = true;
logger.info('Options de filtrage chargées avec succès', {
genres: g.length,
formats: f.length,
countries: c.length,
labels: l.length
});
} catch (err) {
logger.error('Erreur chargement filtres:', err);
// Initialiser avec des tableaux vides en cas d'erreur
genres.value = [];
formats.value = [];
countries.value = [];
labels.value = [];
filterOptionsLoaded.value = true;
apiError.value = 'Erreur lors du chargement des options de filtrage';
}
};
// ✅ FONCTION: Filtrage côté client
const applyClientSideFilters = (data) => {
if (!data || !Array.isArray(data)) return [];
const filtered = data.filter(disc => {
// Filtre par format
if (filters.value.format_id && disc.format_id != filters.value.format_id) {
return false;
}
// Filtre par genre
if (filters.value.genre_id && disc.genre_id != filters.value.genre_id) {
return false;
}
// Filtre par pays
if (filters.value.country_id && disc.country_id != filters.value.country_id) {
return false;
}
// Filtre par label
if (filters.value.label_id && disc.label_id != filters.value.label_id) {
return false;
}
// Filtre par recherche
if (searchQuery.value) {
const query = searchQuery.value.toLowerCase();
const searchableFields = [
disc.title,
disc.artist_name,
disc.genre_name,
disc.format_name,
disc.country_name,
disc.label_name,
disc.barcode
].filter(Boolean).join(' ').toLowerCase();
if (!searchableFields.includes(query)) {
return false;
}
}
return true;
});
logger.debug('Filtrage côté client appliqué', {
total: data.length,
filtrés: filtered.length,
filtres: filters.value,
recherche: searchQuery.value
});
return filtered;
};
// ✅ FONCTION: Charger les disques — ✅ CORRIGÉE : restauration du focus
const fetchDiscs = async () => {
// ✅ Sauvegarder l'état du focus AVANT fetch
const wasFocused = document.activeElement === searchInputRef.value;
isLoading.value = true;
apiError.value = null;
try {
const hasActiveFilters = searchQuery.value ||
filters.value.genre_id ||
filters.value.format_id ||
filters.value.country_id ||
filters.value.label_id;
let data;
if (hasActiveFilters) {
logger.debug('Filtres actifs détectés, filtrage côté client');
const allData = await apiFetch(API_URL);
logger.debug('Données brutes récupérées:', allData?.length || 0, 'disques');
data = applyClientSideFilters(allData || []);
isClientSideFiltering.value = true;
} else {
logger.debug('Aucun filtre, requête simple');
data = await apiFetch(API_URL);
isClientSideFiltering.value = false;
}
discs.value = Array.isArray(data) ? data : [];
totalItems.value = discs.value.length;
logger.info('Disques chargés avec succès', {
total: discs.value.length,
filtrageCôtéClient: isClientSideFiltering.value
});
} catch (error) {
logger.error('Erreur récupération disques:', error);
apiError.value = `Erreur lors du chargement des disques: ${error.message}`;
discs.value = [];
} finally {
isLoading.value = false;
isInitialLoad.value = false;
// ✅ RESTAURER LE FOCUS APRÈS MISE À JOUR
if (wasFocused && searchInputRef.value) {
nextTick(() => {
searchInputRef.value?.focus();
});
}
}
};
// ✅ FONCTION: Mise à jour locale après sauvegarde
const updateLocalDiscs = async (savedDisc) => {
const index = discs.value.findIndex(d => d.id === savedDisc.id);
// ✅ Forcer un nouveau timestamp si updated_at n'est pas présent
// Cela invalide le cache du navigateur pour l'image
if (!savedDisc.updated_at) {
savedDisc.updated_at = new Date().toISOString();
}
// ✅ Vérifier si des filtres sont actifs
const hasActiveFilters = searchQuery.value ||
filters.value.genre_id ||
filters.value.format_id ||
filters.value.country_id ||
filters.value.label_id;
if (index !== -1) {
// Mise à jour d'un disque existant
discs.value.splice(index, 1, savedDisc);
logger.debug('Disque mis à jour localement:', savedDisc.id);
// ✅ Si filtres actifs, vérifier si le disque modifié correspond toujours aux filtres
if (hasActiveFilters) {
const matchesFilters = checkDiscMatchesFilters(savedDisc);
if (!matchesFilters) {
// Le disque ne correspond plus aux filtres, le retirer de la liste
discs.value = discs.value.filter(d => d.id !== savedDisc.id);
totalItems.value = discs.value.length;
logger.debug('Disque retiré de la liste (ne correspond plus aux filtres):', savedDisc.id);
}
}
// Force le recalcul des computed properties
discs.value = [...discs.value];
} else {
// Ajout d'un nouveau disque
// ✅ CORRECTION: Si des filtres sont actifs, recharger la liste depuis le serveur
// pour éviter d'afficher un disque qui ne correspond pas aux filtres
if (hasActiveFilters) {
logger.debug('Filtres actifs détectés, rechargement de la liste depuis le serveur');
await fetchDiscs();
} else {
// Pas de filtres, ajouter directement
discs.value.unshift(savedDisc);
totalItems.value = discs.value.length;
logger.debug('Nouveau disque ajouté localement:', savedDisc.id);
// Force le recalcul des computed properties
discs.value = [...discs.value];
}
}
};
// ✅ NOUVELLE FONCTION: Vérifier si un disque correspond aux filtres actifs
const checkDiscMatchesFilters = (disc) => {
// Filtre par recherche
if (searchQuery.value) {
const query = searchQuery.value.toLowerCase();
const searchableFields = [
disc.title,
disc.artist_name,
disc.genre_name,
disc.format_name,
disc.country_name,
disc.label_name,
disc.barcode
].filter(Boolean).join(' ').toLowerCase();
if (!searchableFields.includes(query)) {
return false;
}
}
// Filtre par genre
if (filters.value.genre_id && disc.genre_id != filters.value.genre_id) {
return false;
}
// Filtre par format
if (filters.value.format_id && disc.format_id != filters.value.format_id) {
return false;
}
// Filtre par pays
if (filters.value.country_id && disc.country_id != filters.value.country_id) {
return false;
}
// Filtre par label
if (filters.value.label_id && disc.label_id != filters.value.label_id) {
return false;
}
return true;
};
// ✅ TRI CÔTÉ CLIENT
const sortKey = ref('title');
const sortOrder = ref('asc');
const sortBy = (key) => {
if (sortKey.value === key) {
sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc';
} else {
sortKey.value = key;
sortOrder.value = 'asc';
}
logger.debug('Tri appliqué', { clé: key, ordre: sortOrder.value });
};
const getSortIcon = (key) => {
if (sortKey.value !== key) return '<span style="font-weight:700">👆 👇</span>';
return sortOrder.value === 'asc' ? '👆' : '👇';
};
// ✅ COMPUTED: Tri et pagination
const sortedAndPaginatedDiscs = computed(() => {
if (discs.value.length === 0) return [];
// 🔄 Tri
let result = [...discs.value];
const key = sortKey.value;
const order = sortOrder.value === 'asc' ? 1 : -1;
result.sort((a, b) => {
let aVal = a[key] ?? '';
let bVal = b[key] ?? '';
// Gestion des champs relationnels
if (key === 'artist_name') aVal = a.artist_name ?? '';
if (key === 'genre_name') aVal = a.genre_name ?? '';
if (key === 'format_name') aVal = a.format_name ?? '';
if (key === 'country_name') aVal = a.country_name ?? '';
if (key === 'label_name') aVal = a.label_name ?? '';
// Tri numérique pour l'année
if (key === 'release_year') {
aVal = a.release_year ?? 0;
bVal = b.release_year ?? 0;
return (aVal - bVal) * order;
}
// ✅ TRI NUMÉRIQUE POUR QUANTITY
if (key === 'quantity') {
aVal = parseInt(a.quantity) || 1;
bVal = parseInt(b.quantity) || 1;
return (aVal - bVal) * order;
}
// ✅ TRI NUMÉRIQUE POUR LE PRIX
if (key === 'price') {
aVal = parseFloat(a.price) || 0;
bVal = parseFloat(b.price) || 0;
return (aVal - bVal) * order;
}
// Tri textuel pour les autres champs
aVal = String(aVal).toLowerCase();
bVal = String(bVal).toLowerCase();
return aVal < bVal ? -order : aVal > bVal ? order : 0;
});
// 📄 Pagination côté client
const startIndex = (currentPage.value - 1) * itemsPerPage.value;
const endIndex = startIndex + itemsPerPage.value;
const paginated = result.slice(startIndex, endIndex);
return paginated;
});
// ✅ FONCTION: Ouvrir modale
const openModal = (disc = null) => {
logger.info('Ouverture de la modale:', disc ? 'Édition' : 'Création', disc?.id);
apiError.value = null;
currentDisc.value = disc ? { ...disc } : null;
isModalOpen.value = true;
};
// ✅ FONCTION: Fermer modale
const closeModal = () => {
logger.debug('Fermeture de la modale');
apiError.value = null;
isModalOpen.value = false;
currentDisc.value = null;
isSaving.value = false; // Réinitialiser l'état de sauvegarde
};
// ✅ FONCTION: Édition depuis code-barres
const handleEditExistingDisc = (discId) => {
logger.info('Ouverture du disque existant:', discId);
if (!discId) {
logger.error('ID du disque non défini');
apiError.value = 'Erreur: ID du disque non défini';
return;
}
const existingDisc = discs.value.find(disc => disc.id === discId);
if (existingDisc) {
currentDisc.value = { ...existingDisc };
isModalOpen.value = true;
logger.debug('Disque existant chargé:', existingDisc.title, 'ID:', existingDisc.id);
} else {
logger.error('Disque non trouvé avec ID:', discId);
apiError.value = `Disque non trouvé (ID: ${discId})`;
}
};
// ✅ NOUVELLE FONCTION: Vérifier ou créer artiste
const ensureArtistExists = async (artistName) => {
if (!artistName || artistName.trim() === '') {
return null;
}
try {
logger.debug('Recherche de l\'artiste:', artistName);
// Vérifier si l'artiste existe déjà
const existingArtists = await apiFetch(`/artists?name=${encodeURIComponent(artistName)}`);
if (existingArtists && existingArtists.length > 0) {
logger.debug('Artiste trouvé:', existingArtists[0].id);
return existingArtists[0].id;
}
// Créer l'artiste
logger.info('Création d\'un nouvel artiste:', artistName);
const newArtist = await apiFetch('/artists', {
method: 'POST',
body: JSON.stringify({ name: artistName.trim() }),
headers: { 'Content-Type': 'application/json' }
});
logger.info('Artiste créé avec succès:', newArtist.id);
return newArtist.id;
} catch (error) {
logger.error('Erreur lors de la création de l\'artiste:', error);
// Si l'artiste existe déjà mais n'a pas été trouvé dans la recherche
if (error.message.includes('existe déjà') || error.message.includes('Duplicate entry')) {
// Essayer de récupérer l'artiste existant
try {
const artists = await apiFetch('/artists');
const existingArtist = artists.find(a =>
a.name.toLowerCase() === artistName.toLowerCase()
);
if (existingArtist) {
return existingArtist.id;
}
} catch (searchError) {
logger.error('Erreur recherche artiste:', searchError);
}
}
throw new Error(`Impossible de créer ou trouver l\'artiste "${artistName}": ${error.message}`);
}
};
// ✅ FONCTION: Gérer la sauvegarde avec création d'artiste
// 📍 DANS DiscsView.vue - Fonction handleSaveWithValidation
// Cherchez cette section et remplacez-la ENTIÈREMENT :

const handleSaveWithValidation = async (formData) => {
  apiError.value = null;
  isSaving.value = true;
  const isUpdate = !!formData.id;
  
  logger.info('Sauvegarde en cours', {
    mode: isUpdate ? 'MODIFICATION' : 'CRÉATION',
    id: formData.id,
    titre: formData.title,
    code_barres: formData.barcode,
    artiste: formData.artist_name
  });
  try {
    if (!isUpdate && formData.barcode?.trim()) {
      const existingDisc = discs.value.find(disc => disc.barcode === formData.barcode);
      if (existingDisc) {
        apiError.value = `Le code-barres "${formData.barcode}" est déjà utilisé par le disque "${existingDisc.title}" (ID: ${existingDisc.id}).`;
        logger.warn('Conflit détecté:', apiError.value);
        isSaving.value = false;
        return;
      }
    }
    let artistId = formData.artist_id;
    if (formData.artist_name && formData.artist_name.trim() && !artistId) {
      try {
        artistId = await ensureArtistExists(formData.artist_name);
        logger.debug('ID artiste obtenu:', artistId);
      } catch (artistError) {
        apiError.value = artistError.message;
        isSaving.value = false;
        return;
      }
    }
    
    const url = isUpdate ? `${API_URL}/${formData.id}` : API_URL;
    
    const body = {
      title: formData.title?.trim() || '',
      artist_id: artistId || null,
      artist_name: formData.artist_name?.trim() || '',
      genre_id: formData.genre_id || null,
      format_id: formData.format_id || null,
      country_id: formData.country_id || null,
      label_id: formData.label_id || null,
      release_year: formData.release_year ? parseInt(formData.release_year) : null,
      barcode: formData.barcode?.trim() || null,
      price: formData.price !== undefined && formData.price !== '' ? parseFloat(formData.price) : null,
      quantity: formData.quantity !== undefined && formData.quantity !== '' ? parseInt(formData.quantity) : 1,
      notes: formData.notes?.trim() || null,
      isrc: formData.isrc?.trim() || null
    };
    
    if (!isUpdate) {
      body.cover_image = formData.cover_image || null;
    } else {
      if (formData.cover_image === '') {
        body.cover_image = '';
      } else if (formData.cover_image) {
        body.cover_image = formData.cover_image;
      }
    }
    
    Object.keys(body).forEach(key => {
      if (key !== 'cover_image' && key !== 'price' && (body[key] === '' || body[key] === null)) {
        body[key] = null;
      }
    });
    
    logger.debug('Données envoyées au backend:', body);

    const result = await apiFetch(url, {
      method: isUpdate ? 'PUT' : 'POST',
      body: JSON.stringify(body),
      headers: {
        'Content-Type': 'application/json'
      }
    });

    await updateLocalDiscs(result);
    closeModal();
    await nextTick();
    logger.info(`${isUpdate ? 'Modification' : 'Création'} réussie:`, result.id);
  } catch (error) {
    logger.error('Erreur sauvegarde:', error);
    handleSaveError(error, formData.id);
  } finally {
    isSaving.value = false;
  }
};


// ✅ FONCTION: Gestion des erreurs de sauvegarde
const handleSaveError = (error, discId = null) => {
const isUpdate = !!discId;
if (error.message.includes('existe déjà') || error.message.includes('existe deja') || error.message.includes('Conflict')) {
const existingId = error.message.match(/ID:?\s*(\d+)/)?.[1] || 'inconnu';
apiError.value = `Erreur : Un disque avec ce code-barres existe déjà (ID: ${existingId}). ${isUpdate ? 'Vous ne pouvez pas modifier le code-barres pour un code déjà utilisé par un autre disque.' : 'Ce code-barres est déjà utilisé.'}`;
}
else if (error.message.includes('1452') || error.message.includes('foreign key constraint')) {
apiError.value = 'Erreur de clé étrangère. Vérifiez que l\'artiste, le genre, le format, le pays ou le label existe bien dans la base de données.';
}
else if (error.message.includes('400') || error.message.includes('Bad Request')) {
apiError.value = 'Données invalides envoyées au serveur. Vérifiez que tous les champs sont corrects.';
}
else if (error.message.includes('404') || error.message.includes('Not Found')) {
apiError.value = isUpdate ? 'Disque non trouvé. Il a peut-être été supprimé entre-temps.' : 'Erreur lors de la création.';
}
else if (error.message.includes('500') || error.message.includes('Internal Server Error')) {
apiError.value = 'Erreur interne du serveur. Veuillez réessayer.';
}
else if (error.message.includes('Cannot add or update a child row')) {
apiError.value = 'Erreur de référence : Une des références (artiste, genre, format, pays ou label) n\'existe pas.';
}
else {
apiError.value = `${isUpdate ? 'Modification' : 'Création'} échouée: ${error.message}`;
}
};
// ✅ FONCTION: Préparation suppression
const prepareDelete = (disc) => {
discToDelete.value = disc;
isConfirmModalOpen.value = true;
logger.debug('Préparation suppression:', disc.id, disc.title);
};
// ✅ FONCTION: Exécuter suppression
const executeDelete = async () => {
if (!discToDelete.value) return;
const id = discToDelete.value.id;
const discTitle = discToDelete.value.title;
isConfirmModalOpen.value = false;
try {
await apiFetch(`${API_URL}/${id}`, { method: 'DELETE' });
// Mise à jour locale
const index = discs.value.findIndex(d => d.id === id);
if (index !== -1) {
discs.value.splice(index, 1);
totalItems.value = discs.value.length;
}
if (selectedRowId.value === id) selectedRowId.value = null;
discToDelete.value = null;
logger.info('Suppression réussie:', id, discTitle);
} catch (error) {
logger.error('Erreur suppression:', error);
apiError.value = `Échec suppression: ${error.message}`;
}
};
// ✅ FONCTION: Réinitialiser filtres
const resetFilters = () => {
searchQuery.value = '';
filters.value = { genre_id: '', format_id: '', country_id: '', label_id: '' };
currentPage.value = 1;
isClientSideFiltering.value = false;
fetchDiscs();
logger.debug('Filtres réinitialisés');
};
// ✅ FONCTION: Supprimer filtre spécifique
const removeFilter = (type) => {
if (type === 'search') {
searchQuery.value = '';
} else {
filters.value[type] = '';
}
currentPage.value = 1;
fetchDiscs();
logger.debug('Filtre supprimé:', type);
};
// ✅ FONCTION: Basculer affichage filtres
const toggleFilters = () => {
showFilters.value = !showFilters.value;
logger.debug('Affichage filtres:', showFilters.value ? 'visibles' : 'masqués');
};
// ✅ FONCTION: Navigation pagination
const goToPage = (page) => {
if (page >= 1 && page <= totalPages.value) {
currentPage.value = page;
logger.debug('Navigation page:', page);
}
};
// 🆕 FONCTION: Aller à la première page
const goToFirstPage = () => {
if (currentPage.value !== 1) {
goToPage(1);
}
};
// 🆕 FONCTION: Aller à la dernière page
const goToLastPage = () => {
if (currentPage.value !== totalPages.value) {
goToPage(totalPages.value);
}
};
// 🆕 FONCTION: Aller directement à une page saisie
const goToInputPage = () => {
const page = parseInt(pageInputValue.value);
if (!isNaN(page) && page >= 1 && page <= totalPages.value) {
goToPage(page);
pageInputValue.value = '';
} else {
// Réinitialiser si invalide
pageInputValue.value = '';
}
};
// 🆕 FONCTION: Gérer la touche Entrée dans le champ de saisie
const handlePageInputEnter = (event) => {
if (event.key === 'Enter') {
goToInputPage();
}
};
// ✅ FONCTION: Gestion sélection ligne
const handleRowClick = (disc) => {
selectedRowId.value = selectedRowId.value === disc.id ? null : disc.id;
logger.debug('Sélection ligne:', disc.id);
};
const handleRowDoubleClick = (disc) => {
selectedRowId.value = disc.id;
openModal(disc);
};
// ✅ FONCTION: Tooltip notes
const showNotesTooltip = (disc, event) => {
if (!disc.notes) return;
const existingTooltips = document.querySelectorAll('.notes-tooltip');
existingTooltips.forEach(tooltip => tooltip.remove());
const buttonRect = event.target.closest('button').getBoundingClientRect();
const tooltip = document.createElement('div');
tooltip.className = 'notes-tooltip';
tooltip.setAttribute('role', 'tooltip');
tooltip.setAttribute('aria-live', 'polite');
const tooltipTop = buttonRect.top + window.scrollY;
const tooltipLeft = buttonRect.left + window.scrollX;
tooltip.style.position = 'absolute';
tooltip.style.top = `${tooltipTop}px`;
tooltip.style.left = `${tooltipLeft}px`;
tooltip.style.transform = 'translateX(calc(-100% - 10px))';
tooltip.innerHTML = `
<div class="tooltip-header">
<span class="tooltip-icon" aria-hidden="true">📝</span>
<strong>${disc.title}</strong>
</div>
<div class="tooltip-content">${disc.notes}</div>
`;
document.body.appendChild(tooltip);
};
const hideNotesTooltip = () => {
const tooltips = document.querySelectorAll('.notes-tooltip');
tooltips.forEach(tooltip => tooltip.remove());
};
// ✅ FONCTION: Aperçu pochette
const showCoverPreview = (disc, event) => {
if (!disc.cover_url) return;
const existingPreviews = document.querySelectorAll('.cover-preview');
existingPreviews.forEach(preview => preview.remove());
const coverRect = event.target.getBoundingClientRect();
const preview = document.createElement('div');
preview.className = 'cover-preview';
preview.setAttribute('role', 'tooltip');
preview.setAttribute('aria-label', `Aperçu pochette: ${disc.title}`);
// Dimensions de la tooltip
const tooltipWidth = 324; // 300px image + 12px padding × 2
const tooltipHeight = 470; // estimation (300px image + infos + stats)
// Détection mode mobile
const isMobile = window.innerWidth < 768;
let previewLeft, previewTop;
if (isMobile) {
// Mode mobile : centrer la tooltip à l'écran
previewLeft = (window.innerWidth - tooltipWidth) / 2 + window.scrollX;
previewTop = Math.max(50, (window.innerHeight - tooltipHeight) / 2) + window.scrollY;
} else {
// Mode desktop : à côté de la pochette
// Position par défaut : à droite de la pochette
previewLeft = coverRect.right + window.scrollX + 10;
previewTop = coverRect.top + window.scrollY;
// Vérifier si la tooltip dépasse à droite
if (previewLeft + tooltipWidth > window.innerWidth + window.scrollX) {
// Afficher à gauche de la pochette
previewLeft = coverRect.left + window.scrollX - tooltipWidth - 10;
}
// Si encore à gauche de l'écran, forcer à droite avec marge
if (previewLeft < window.scrollX + 10) {
previewLeft = window.scrollX + 10;
}
// Vérifier si la tooltip dépasse en bas
const maxTop = window.innerHeight + window.scrollY - tooltipHeight - 20;
if (previewTop > maxTop) {
previewTop = Math.max(window.scrollY + 10, maxTop);
}
// Vérifier si la tooltip dépasse en haut
if (previewTop < window.scrollY + 10) {
previewTop = window.scrollY + 10;
}
}
preview.style.position = 'absolute';
preview.style.top = `${previewTop}px`;
preview.style.left = `${previewLeft}px`;
preview.style.zIndex = '10000';
// Ajouter overlay et bouton fermer en mode mobile
const closeButton = isMobile ? `
<button class="cover-preview-close" onclick="this.closest('.cover-preview').remove(); document.querySelector('.cover-preview-overlay')?.remove();" aria-label="Fermer">×</button>
` : '';
preview.innerHTML = `
${closeButton}
<div class="cover-preview-content">
<div class="cover-preview-image-container">
<img
src="${getImageUrl(disc.cover_url, disc.updated_at)}"
alt="Pochette: ${disc.title}"
class="cover-preview-image"
onerror="this.style.display='none'"
/>
</div>
<div class="cover-preview-stats">
<div class="preview-stat-item">
<span class="preview-label">Nb disc :</span>
<span class="preview-value">${formatQuantity(disc.quantity)}</span>
</div>
<div class="preview-stat-item">
<span class="preview-label">Genre :</span>
<span class="preview-value">${displayValue(disc.genre_name)}</span>
</div>
<div class="preview-stat-item">
<span class="preview-label">Année :</span>
<span class="preview-value">${displayValue(disc.release_year)}</span>
</div>
<div class="preview-stat-item">
<span class="preview-label">Pays :</span>
<span class="preview-value">${displayValue(disc.country_name)}</span>
</div>
</div>
<div class="cover-preview-info">
<strong>${disc.title}</strong>
${disc.artist_name ? `<div>${disc.artist_name}</div>` : ''}
</div>
</div>
`;
// Ajouter overlay en mode mobile
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
const previews = document.querySelectorAll('.cover-preview');
previews.forEach(preview => preview.remove());
};
// ✅ FONCTION: Export données
const exportData = () => {
if (discs.value.length === 0) {
apiError.value = 'Aucune donnée à exporter';
return;
}
const dataStr = JSON.stringify(discs.value, null, 2);
const dataBlob = new Blob([dataStr], { type: 'application/json' });
const url = URL.createObjectURL(dataBlob);
const link = document.createElement('a');
link.href = url;
link.download = `discs-export-${new Date().toISOString().split('T')[0]}.json`;
link.click();
URL.revokeObjectURL(url);
logger.info('Export réalisé:', discs.value.length, 'disques');
};
// ✅ FONCTION: Création artiste (pour la modale)
const handleCreateArtist = async (artistName) => {
logger.info('Création artiste demandée depuis la modale:', artistName);
try {
const artistId = await ensureArtistExists(artistName);
logger.info('Artiste créé avec succès, ID:', artistId);
return artistId;
} catch (error) {
logger.error('Erreur création artiste:', error);
apiError.value = error.message;
return null;
}
};
// ✅ WATCH: Recherche et filtres avec debounce
watch([searchQuery, filters], () => {
if (searchDebounceTimer.value) clearTimeout(searchDebounceTimer.value);
searchDebounceTimer.value = setTimeout(() => {
currentPage.value = 1;
fetchDiscs();
}, DEBOUNCE_DELAY);
}, { deep: true });
// 🆕 WATCH: Colonnes pour réappliquer les styles sticky
watch(columns, () => {
nextTick(() => {
updateStickyPositions();
});
}, { deep: true });
</script>
<template>
<div class="discs-view">
<!-- ✅ HEADER -->
<header class="page-header">
<div class="header-wrapper">
<div class="header-top-row">
<div class="title-section">
<button @click="$router.push('/dashboard')" class="back-button ghost-btn">
<span class="icon">🏠</span>
</button>
<span class="title-icon" aria-hidden="true">💿</span>
<h1>Gestion des Disques</h1>
</div>
<div class="toolbar">
<button
@click="toggleFilters"
class="filter-toggle-button"
:class="{ 'filters-hidden': !showFilters }"
:aria-expanded="showFilters"
aria-label="Basculer l'affichage des filtres"
:disabled="isLoading"
>
<span class="icon" aria-hidden="true">{{ showFilters ? '👁️' : '👁️‍🗨️' }}</span>
{{ showFilters ? 'Masquer' : 'Afficher' }} Filtres
</button>
<!-- 🆕 BOUTON GESTION COLONNES -->
<div class="column-menu-wrapper" ref="columnMenuRef" style="position: relative;">
<button
@click.stop="showColumnMenu = !showColumnMenu"
class="filter-toggle-button"
aria-label="Gérer les colonnes visibles"
:disabled="isLoading"
title="Afficher/masquer les colonnes"
>
<span class="icon" aria-hidden="true">⚙️</span>
Colonnes
</button>
<!-- Menu déroulant colonnes -->
<div v-if="showColumnMenu" class="column-menu">
<div class="column-menu-header">
<h3>Colonnes visibles</h3>
<button
class="btn-reset"
@click="resetColumns"
title="Réinitialiser"
aria-label="Réinitialiser les colonnes"
>
🔄
</button>
</div>
<div class="column-menu-list">
<div
v-for="col in columns"
:key="col.key"
class="column-menu-item"
>
<div class="column-menu-item-controls">
<span class="column-label">{{ col.label }}</span>
<div class="column-actions">
<button
                      @click.stop="canToggleSticky(col.key) && toggleSticky(col.key)"
                      class="sticky-toggle-btn"
                      :class="{ active: col.sticky, disabled: !canToggleSticky(col.key) }"
                      :disabled="!canToggleSticky(col.key)"
                      :title="col.sticky ? 'Déverrouiller' : 'Verrouiller'"
                    >
                      {{ col.sticky ? '🔓' : '🔒' }}
                    </button>
<label class="toggle-switch" @click.stop>
<input
type="checkbox"
:checked="col.visible"
@change="toggleColumn(col.key)"
:disabled="col.key === 'actions'"
/>
<span class="toggle-slider"></span>
</label>
</div>
</div>
</div>
</div>
</div>
</div>
<button
@click="exportData"
class="export-button"
:disabled="discs.length === 0 || isLoading"
aria-label="Exporter les données"
>
<span class="icon" aria-hidden="true">📤</span> Exporter
</button>
<button
@click="openModal()"
class="primary-btn add-button"
:disabled="isLoading"
aria-label="Ajouter un nouveau disque"
>
<span class="icon" aria-hidden="true">➕</span> Ajouter un Disque
</button>
</div>
</div>
<p class="subtitle">Gérez votre collection de disques vinyles.</p>
<p class="user-hint">
<span aria-hidden="true">💡</span>
<strong>Astuce :</strong> Clic pour sélectionner, double-clic pour modifier un disque
</p>
</div>
</header>
<!-- ✅ MESSAGE FILTRAGE CLIENT -->
<div
v-if="isClientSideFiltering && activeFilters.length > 0"
class="info-message"
role="status"
aria-live="polite"
>
<span class="info-icon" aria-hidden="true">ℹ️</span>
<span>Filtrage appliqué localement ({{ sortedAndPaginatedDiscs.length }} résultat(s))</span>
</div>
<!-- ✅ BARRE FILTRES -->
<div v-if="showFilters" class="filter-bar">
<div class="search-group">
<label for="search">Rechercher :</label>
<div class="search-wrapper">
<!-- 🔧 MODIF ICI : ajout de ref="searchInputRef" et key fixe -->
<input
ref="searchInputRef"
key="search-input"
id="search"
v-model="searchQuery"
type="text"
placeholder="Titre, artiste, genre, label, code-barres..."
class="search-input"
aria-label="Rechercher des disques"
:disabled="isLoading"
/>
<span class="search-icon" aria-hidden="true">🔍</span>
</div>
</div>
<div class="filters-grid">
<div class="filter-group">
<label for="filter-genre">Genre :</label>
<select
id="filter-genre"
v-model="filters.genre_id"
class="filter-select"
:disabled="isLoading"
aria-label="Filtrer par genre"
>
<option value="">Tous</option>
<option v-for="g in genres" :key="g.id" :value="g.id">{{ g.name }}</option>
</select>
</div>
<div class="filter-group">
<label for="filter-format">Format :</label>
<select
id="filter-format"
v-model="filters.format_id"
class="filter-select"
:disabled="isLoading"
aria-label="Filtrer par format"
>
<option value="">Tous</option>
<option v-for="f in formats" :key="f.id" :value="f.id">{{ f.name }}</option>
</select>
</div>
<div class="filter-group">
<label for="filter-country">Pays :</label>
<select
id="filter-country"
v-model="filters.country_id"
class="filter-select"
:disabled="isLoading"
aria-label="Filtrer par pays"
>
<option value="">Tous</option>
<option v-for="c in countries" :key="c.id" :value="c.id">{{ c.name }}</option>
</select>
</div>
<div class="filter-group">
<label for="filter-label">Label :</label>
<select
id="filter-label"
v-model="filters.label_id"
class="filter-select"
:disabled="isLoading"
aria-label="Filtrer par label"
>
<option value="">Tous</option>
<option v-for="l in labels" :key="l.id" :value="l.id">{{ l.name }}</option>
</select>
</div>
</div>
<button
@click="resetFilters"
class="reset-button"
:disabled="activeFilters.length === 0 || isLoading"
aria-label="Réinitialiser tous les filtres"
>
<span aria-hidden="true">🗑️</span> Réinitialiser ({{ activeFilters.length }})
</button>
</div>
<!-- ✅ BADGES FILTRES ACTIFS -->
<div
v-if="activeFilters.length > 0"
class="active-filters-badges"
role="region"
aria-label="Filtres actifs"
>
<span class="filters-title">Filtres actifs :</span>
<span
v-for="(filter, index) in activeFilters"
:key="index"
class="filter-badge"
@click="removeFilter(filter.type)"
:aria-label="`Supprimer le filtre ${filter.label}`"
role="button"
tabindex="0"
@keydown.enter="removeFilter(filter.type)"
@keydown.space="removeFilter(filter.type)"
>
{{ filter.label }}
<span class="remove-icon" aria-hidden="true">×</span>
</span>
</div>
<!-- ✅ CHARGEMENT -->
<div
v-if="isLoading"
class="loading-indicator"
role="status"
aria-live="polite"
aria-busy="true"
>
<div class="spinner" aria-hidden="true"></div>
<span>Chargement des disques...</span>
</div>
<!-- ✅ ERREUR -->
<div
v-if="apiError && !isLoading"
class="error-message"
role="alert"
aria-live="assertive"
>
<span class="error-icon" aria-hidden="true">⚠️</span>
<span class="error-text">{{ apiError }}</span>
<button
@click="apiError = null"
class="error-close"
aria-label="Fermer le message d'erreur"
>
×
</button>
</div>
<!-- ✅ TABLEAU ou CARDS -->
<div class="content-panel panel" v-if="!isInitialLoad">
<!-- ✅ TABLEAU : Desktop / tablette -->
<div v-if="!isMobileView" class="table-responsive" ref="tableContainer">
<table class="data-table" aria-label="Liste des disques">
<thead>
<tr>
<!-- Colonne JACKET -->
<th
v-if="visibleColumns.find(c => c.key === 'cover')"
class="cover-column resizable-header"
			title="Pochette du disque"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'cover').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'cover').stickyLeft !== undefined
}"
scope="col"
:style="{
width: columns.find(c => c.key === 'cover').width + 'px',
...getStickyStyle(visibleColumns.find(c => c.key === 'cover'))
}"
>
<div class="sortable-content header-content">
<span class="header-text">
<span v-if="visibleColumns.find(c => c.key === 'cover').sticky" class="pin-icon" title="Colonne figée">📌</span>
{{ columns.find(c => c.key === 'cover').label }}
</span>
</div>
</th>
<!-- Colonne TITRE -->
<th
v-if="visibleColumns.find(c => c.key === 'title')"
class="title-column sortable resizable-header"
			title="Titre du disque"
@click="sortBy('title')"
:class="{
'active-sort': sortKey === 'title',
'sticky-column': visibleColumns.find(c => c.key === 'title').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'title').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'title').stickyRight !== undefined
}"
scope="col"
role="button"
:aria-sort="sortKey === 'title' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'"
tabindex="0"
@keydown.enter="sortBy('title')"
@keydown.space="sortBy('title')"
:style="{
width: columns.find(c => c.key === 'title').width + 'px',
...getStickyStyle(visibleColumns.find(c => c.key === 'title'))
}"
style="position: relative"
>
<div class="sortable-content header-content">
<span class="header-text">
<span v-if="visibleColumns.find(c => c.key === 'title').sticky" class="pin-icon" title="Colonne figée">📌</span>
{{ columns.find(c => c.key === 'title').label }}
</span>
<span class="sort-icon" v-html="getSortIcon('title')" aria-hidden="true"></span>
</div>
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'title'))" @dblclick.stop="autoFitColumn('title')" title="Double-cliquez pour ajuster automatiquement"></div>
</th>
<!-- Colonne ARTISTE -->
<th
v-if="visibleColumns.find(c => c.key === 'artist')"
class="artist-column sortable resizable-header"
			title="Nom de l'artiste"
@click="sortBy('artist_name')"
:class="{ 'active-sort': sortKey === 'artist_name' }"
scope="col"
role="button"
:aria-sort="sortKey === 'artist_name' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'"
tabindex="0"
@keydown.enter="sortBy('artist_name')"
@keydown.space="sortBy('artist_name')"
:style="{
width: columns.find(c => c.key === 'artist').width + 'px',
...getStickyStyle(visibleColumns.find(c => c.key === 'artist'))
}"
style="position: relative"
>
<div class="sortable-content header-content">
<span class="header-text">
<span v-if="visibleColumns.find(c => c.key === 'artist').sticky" class="pin-icon" title="Colonne figée">📌</span>
{{ columns.find(c => c.key === 'artist').label }}
</span>
<span class="sort-icon" v-html="getSortIcon('artist_name')" aria-hidden="true"></span>
</div>
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'artist'))" @dblclick.stop="autoFitColumn('artist')" title="Double-cliquez pour ajuster automatiquement"></div>
</th>
<!-- Colonne NB DISC -->
<th
v-if="visibleColumns.find(c => c.key === 'quantity')"
class="quantity-column sortable resizable-header"
			title="Quantité de disques"
@click="sortBy('quantity')"
:class="{
'active-sort': sortKey === 'quantity',
'sticky-column': visibleColumns.find(c => c.key === 'quantity').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'quantity').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'quantity').stickyRight !== undefined
}"
scope="col"
role="button"
:aria-sort="sortKey === 'quantity' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'"
tabindex="0"
@keydown.enter="sortBy('quantity')"
@keydown.space="sortBy('quantity')"
:style="{
width: columns.find(c => c.key === 'quantity').width + 'px',
...getStickyStyle(visibleColumns.find(c => c.key === 'quantity'))
}"
style="position: relative"
>
<div class="sortable-content header-content">
<span class="header-text">
<span v-if="visibleColumns.find(c => c.key === 'quantity').sticky" class="pin-icon" title="Colonne figée">📌</span>
{{ columns.find(c => c.key === 'quantity').label }}
</span>
<span class="sort-icon" v-html="getSortIcon('quantity')" aria-hidden="true"></span>
</div>
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'quantity'))" @dblclick.stop="autoFitColumn('quantity')" title="Double-cliquez pour ajuster automatiquement"></div>
</th>
<!-- Colonne GENRE -->
<th
v-if="visibleColumns.find(c => c.key === 'genre')"
class="genre-column sortable resizable-header"
			title="Genre musical"
@click="sortBy('genre_name')"
:class="{ 'active-sort': sortKey === 'genre_name' }"
scope="col"
role="button"
:aria-sort="sortKey === 'genre_name' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'"
tabindex="0"
@keydown.enter="sortBy('genre_name')"
@keydown.space="sortBy('genre_name')"
:style="{
width: columns.find(c => c.key === 'genre').width + 'px',
...getStickyStyle(visibleColumns.find(c => c.key === 'genre'))
}"
style="position: relative"
>
<div class="sortable-content header-content">
<span class="header-text">
<span v-if="visibleColumns.find(c => c.key === 'genre').sticky" class="pin-icon" title="Colonne figée">📌</span>
{{ columns.find(c => c.key === 'genre').label }}
</span>
<span class="sort-icon" v-html="getSortIcon('genre_name')" aria-hidden="true"></span>
</div>
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'genre'))" @dblclick.stop="autoFitColumn('genre')" title="Double-cliquez pour ajuster automatiquement"></div>
</th>
<!-- Colonne ANNÉE -->
<th
v-if="visibleColumns.find(c => c.key === 'year')"
class="year-column sortable resizable-header"
			title="Année de sortie"
@click="sortBy('release_year')"
:class="{ 'active-sort': sortKey === 'release_year' }"
scope="col"
role="button"
:aria-sort="sortKey === 'release_year' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'"
tabindex="0"
@keydown.enter="sortBy('release_year')"
@keydown.space="sortBy('release_year')"
:style="{
width: columns.find(c => c.key === 'year').width + 'px',
...getStickyStyle(visibleColumns.find(c => c.key === 'year'))
}"
style="position: relative"
>
<div class="sortable-content header-content">
<span class="header-text">
<span v-if="visibleColumns.find(c => c.key === 'year').sticky" class="pin-icon" title="Colonne figée">📌</span>
{{ columns.find(c => c.key === 'year').label }}
</span>
<span class="sort-icon" v-html="getSortIcon('release_year')" aria-hidden="true"></span>
</div>
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'year'))" @dblclick.stop="autoFitColumn('year')" title="Double-cliquez pour ajuster automatiquement"></div>
</th>
<!-- Colonne PAYS -->
<th
v-if="visibleColumns.find(c => c.key === 'country')"
class="country-column sortable resizable-header"
			title="Pays d'origine"
@click="sortBy('country_name')"
:class="{ 'active-sort': sortKey === 'country_name' }"
scope="col"
role="button"
:aria-sort="sortKey === 'country_name' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'"
tabindex="0"
@keydown.enter="sortBy('country_name')"
@keydown.space="sortBy('country_name')"
:style="{
width: columns.find(c => c.key === 'country').width + 'px',
...getStickyStyle(visibleColumns.find(c => c.key === 'country'))
}"
style="position: relative"
>
<div class="sortable-content header-content">
<span class="header-text">
<span v-if="visibleColumns.find(c => c.key === 'country').sticky" class="pin-icon" title="Colonne figée">📌</span>
{{ columns.find(c => c.key === 'country').label }}
</span>
<span class="sort-icon" v-html="getSortIcon('country_name')" aria-hidden="true"></span>
</div>
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'country'))" @dblclick.stop="autoFitColumn('country')" title="Double-cliquez pour ajuster automatiquement"></div>
</th>
<!-- Colonne FORMAT -->
<th
v-if="visibleColumns.find(c => c.key === 'format')"
class="format-column sortable resizable-header"
			title="Format du support"
@click="sortBy('format_name')"
:class="{ 'active-sort': sortKey === 'format_name' }"
scope="col"
role="button"
:aria-sort="sortKey === 'format_name' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'"
tabindex="0"
@keydown.enter="sortBy('format_name')"
@keydown.space="sortBy('format_name')"
:style="{
width: columns.find(c => c.key === 'format').width + 'px',
...getStickyStyle(visibleColumns.find(c => c.key === 'format'))
}"
style="position: relative"
>
<div class="sortable-content header-content">
<span class="header-text">
<span v-if="visibleColumns.find(c => c.key === 'format').sticky" class="pin-icon" title="Colonne figée">📌</span>
{{ columns.find(c => c.key === 'format').label }}
</span>
<span class="sort-icon" v-html="getSortIcon('format_name')" aria-hidden="true"></span>
</div>
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'format'))" @dblclick.stop="autoFitColumn('format')" title="Double-cliquez pour ajuster automatiquement"></div>
</th>
<!-- Colonne LABEL -->
<th
v-if="visibleColumns.find(c => c.key === 'label')"
class="label-column sortable resizable-header"
			title="Label / Maison de disques"
@click="sortBy('label_name')"
:class="{ 'active-sort': sortKey === 'label_name' }"
scope="col"
role="button"
:aria-sort="sortKey === 'label_name' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'"
tabindex="0"
@keydown.enter="sortBy('label_name')"
@keydown.space="sortBy('label_name')"
:style="{
width: columns.find(c => c.key === 'label').width + 'px',
...getStickyStyle(visibleColumns.find(c => c.key === 'label'))
}"
style="position: relative"
>
<div class="sortable-content header-content">
<span class="header-text">
<span v-if="visibleColumns.find(c => c.key === 'label').sticky" class="pin-icon" title="Colonne figée">📌</span>
{{ columns.find(c => c.key === 'label').label }}
</span>
<span class="sort-icon" v-html="getSortIcon('label_name')" aria-hidden="true"></span>
</div>
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'label'))" @dblclick.stop="autoFitColumn('label')" title="Double-cliquez pour ajuster automatiquement"></div>
</th>
<!-- Colonne CODE-BARRES -->
<th
v-if="visibleColumns.find(c => c.key === 'barcode')"
class="barcode-column sortable resizable-header"
			title="Code-barres"
@click="sortBy('barcode')"
:class="{
'active-sort': sortKey === 'barcode',
'sticky-column': visibleColumns.find(c => c.key === 'barcode').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'barcode').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'barcode').stickyRight !== undefined
}"
scope="col"
role="button"
:aria-sort="sortKey === 'barcode' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'"
tabindex="0"
@keydown.enter="sortBy('barcode')"
@keydown.space="sortBy('barcode')"
:style="{
width: columns.find(c => c.key === 'barcode').width + 'px',
...getStickyStyle(visibleColumns.find(c => c.key === 'barcode'))
}"
style="position: relative"
>
<div class="sortable-content header-content">
<span class="header-text">
<span v-if="visibleColumns.find(c => c.key === 'barcode').sticky" class="pin-icon" title="Colonne figée">📌</span>
{{ columns.find(c => c.key === 'barcode').label }}
</span>
<span class="sort-icon" v-html="getSortIcon('barcode')" aria-hidden="true"></span>
</div>
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'barcode'))" @dblclick.stop="autoFitColumn('barcode')" title="Double-cliquez pour ajuster automatiquement"></div>
</th>
<!-- Colonne PRIX -->
<th
v-if="visibleColumns.find(c => c.key === 'price')"
class="price-column sortable resizable-header"
			title="Prix"
@click="sortBy('price')"
:class="{
'active-sort': sortKey === 'price',
'sticky-column': visibleColumns.find(c => c.key === 'price').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'price').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'price').stickyRight !== undefined
}"
scope="col"
role="button"
:aria-sort="sortKey === 'price' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'"
tabindex="0"
@keydown.enter="sortBy('price')"
@keydown.space="sortBy('price')"
:style="{
width: columns.find(c => c.key === 'price').width + 'px',
...getStickyStyle(visibleColumns.find(c => c.key === 'price'))
}"
style="position: relative"
>
<div class="sortable-content header-content">
<span class="header-text">
<span v-if="visibleColumns.find(c => c.key === 'price').sticky" class="pin-icon" title="Colonne figée">📌</span>
{{ columns.find(c => c.key === 'price').label }}
</span>
<span class="sort-icon" v-html="getSortIcon('price')" aria-hidden="true"></span>
</div>
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'price'))" @dblclick.stop="autoFitColumn('price')" title="Double-cliquez pour ajuster automatiquement"></div>
</th>
<!-- Colonne ACTIONS (toujours visible) -->
<th
v-if="visibleColumns.find(c => c.key === 'actions')"
class="actions-column resizable-header"
			title="Actions disponibles"
scope="col"
:style="{ width: columns.find(c => c.key === 'actions').width + 'px' }"
>
<div class="sortable-content action-header-content header-content">
<span class="header-text">
<span v-if="visibleColumns.find(c => c.key === 'actions').sticky" class="pin-icon" title="Colonne figée">📌</span>
{{ columns.find(c => c.key === 'actions').label }}
</span>
</div>
</th>
</tr>
</thead>
<tbody>
<tr
v-for="(disc, index) in sortedAndPaginatedDiscs"
:key="disc.id"
:class="{
'odd-row': index % 2 === 0,
'even-row': index % 2 !== 0,
'clickable-row': true,
'selected-row': selectedRowId === disc.id
}"
@click="handleRowClick(disc)"
@dblclick="handleRowDoubleClick(disc)"
:aria-selected="selectedRowId === disc.id"
:aria-rowindex="index + 1"
role="row"
tabindex="0"
@keydown.enter="handleRowClick(disc)"
@keydown.space="handleRowClick(disc)"
@keydown.delete="prepareDelete(disc)"
>
<td
v-if="visibleColumns.find(c => c.key === 'cover')"
class="cover-column"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'cover').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'cover').stickyLeft !== undefined
}"
:style="getStickyStyle(visibleColumns.find(c => c.key === 'cover'))"
role="cell">
<div
class="cover-thumbnail-container"
@mouseenter="showCoverPreview(disc, $event)"
@mouseleave="hideCoverPreview"
:aria-label="`Pochette de ${disc.title}`"
>
<img
v-if="disc.cover_url"
:src="getImageUrl(disc.cover_url, disc.updated_at)"
:alt="`Pochette: ${disc.title}`"
class="cover-thumbnail"
loading="lazy"
@error="handleImageError($event, disc)"
/>
<div
class="cover-fallback"
:style="{ display: disc.cover_url ? 'none' : 'flex' }"
:aria-hidden="disc.cover_url ? 'true' : 'false'"
>
<span class="disc-icon" aria-hidden="true">💿</span>
</div>
</div>
</td>
<td
v-if="visibleColumns.find(c => c.key === 'title')"
class="title-column"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'title').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'title').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'title').stickyRight !== undefined
}"
:style="{ ...getStickyStyle(visibleColumns.find(c => c.key === 'title')), position: 'relative' }"
role="cell"
data-label="Titre">
{{ displayValue(disc.title) }}
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'title'))" @dblclick.stop="autoFitColumn('title')" title="Double-cliquez pour ajuster automatiquement"></div>
</td>
<td
v-if="visibleColumns.find(c => c.key === 'artist')"
class="artist-column"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'artist').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'artist').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'artist').stickyRight !== undefined
}"
:style="{ ...getStickyStyle(visibleColumns.find(c => c.key === 'artist')), position: 'relative' }"
role="cell"
data-label="Artiste">
{{ displayValue(disc.artist_name) }}
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'artist'))" @dblclick.stop="autoFitColumn('artist')" title="Double-cliquez pour ajuster automatiquement"></div>
</td>
<td
v-if="visibleColumns.find(c => c.key === 'quantity')"
class="quantity-column"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'quantity').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'quantity').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'quantity').stickyRight !== undefined
}"
:style="{ ...getStickyStyle(visibleColumns.find(c => c.key === 'quantity')), position: 'relative' }"
role="cell"
data-label="Nb disc">
{{ formatQuantity(disc.quantity) }}
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'quantity'))" @dblclick.stop="autoFitColumn('quantity')" title="Double-cliquez pour ajuster automatiquement"></div>
</td>
<td
v-if="visibleColumns.find(c => c.key === 'genre')"
class="genre-column"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'genre').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'genre').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'genre').stickyRight !== undefined
}"
:style="{ ...getStickyStyle(visibleColumns.find(c => c.key === 'genre')), position: 'relative' }"
role="cell"
data-label="Genre">
{{ displayValue(disc.genre_name) }}
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'genre'))" @dblclick.stop="autoFitColumn('genre')" title="Double-cliquez pour ajuster automatiquement"></div>
</td>
<td
v-if="visibleColumns.find(c => c.key === 'year')"
class="year-column"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'year').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'year').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'year').stickyRight !== undefined
}"
:style="{ ...getStickyStyle(visibleColumns.find(c => c.key === 'year')), position: 'relative' }"
role="cell"
data-label="Année">
{{ displayValue(disc.release_year) }}
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'year'))" @dblclick.stop="autoFitColumn('year')" title="Double-cliquez pour ajuster automatiquement"></div>
</td>
<td
v-if="visibleColumns.find(c => c.key === 'country')"
class="country-column"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'country').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'country').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'country').stickyRight !== undefined
}"
:style="{ ...getStickyStyle(visibleColumns.find(c => c.key === 'country')), position: 'relative' }"
role="cell"
data-label="Pays">
{{ displayValue(disc.country_name) }}
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'country'))" @dblclick.stop="autoFitColumn('country')" title="Double-cliquez pour ajuster automatiquement"></div>
</td>
<td
v-if="visibleColumns.find(c => c.key === 'format')"
class="format-column"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'format').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'format').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'format').stickyRight !== undefined
}"
:style="{ ...getStickyStyle(visibleColumns.find(c => c.key === 'format')), position: 'relative' }"
role="cell"
data-label="Format">
{{ displayValue(disc.format_name) }}
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'format'))" @dblclick.stop="autoFitColumn('format')" title="Double-cliquez pour ajuster automatiquement"></div>
</td>
<td
v-if="visibleColumns.find(c => c.key === 'label')"
class="label-column"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'label').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'label').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'label').stickyRight !== undefined
}"
:style="{ ...getStickyStyle(visibleColumns.find(c => c.key === 'label')), position: 'relative' }"
role="cell"
data-label="Label">
{{ displayValue(disc.label_name) }}
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'label'))" @dblclick.stop="autoFitColumn('label')" title="Double-cliquez pour ajuster automatiquement"></div>
</td>
<td
v-if="visibleColumns.find(c => c.key === 'barcode')"
class="barcode-column"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'barcode').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'barcode').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'barcode').stickyRight !== undefined
}"
:style="{ ...getStickyStyle(visibleColumns.find(c => c.key === 'barcode')), position: 'relative' }"
role="cell"
data-label="Code-barres">
{{ displayValue(disc.barcode) }}
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'barcode'))" @dblclick.stop="autoFitColumn('barcode')" title="Double-cliquez pour ajuster automatiquement"></div>
</td>
<td
v-if="visibleColumns.find(c => c.key === 'price')"
class="price-column"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'price').sticky,
'sticky-left': visibleColumns.find(c => c.key === 'price').stickyLeft !== undefined,
'sticky-right': visibleColumns.find(c => c.key === 'price').stickyRight !== undefined
}"
:style="{ ...getStickyStyle(visibleColumns.find(c => c.key === 'price')), position: 'relative' }"
role="cell"
data-label="Prix">
{{ formatPrice(disc.price) }}
<div class="column-separator" @mousedown="startResize($event, columns.find(c => c.key === 'price'))" @dblclick.stop="autoFitColumn('price')" title="Double-cliquez pour ajuster automatiquement"></div>
</td>
<td
v-if="visibleColumns.find(c => c.key === 'actions')"
class="actions-column"
:class="{
'sticky-column': visibleColumns.find(c => c.key === 'actions').sticky,
'sticky-right': visibleColumns.find(c => c.key === 'actions').stickyRight !== undefined
}"
:style="getStickyStyle(visibleColumns.find(c => c.key === 'actions'))"
role="cell">
<div class="action-buttons-container">
<button
@click.stop="openModal(disc)"
class="icon-action-btn edit-button"
:aria-label="`Modifier le disque ${disc.title}`"
:disabled="isLoading || isSaving"
>
<span class="icon" aria-hidden="true">✏️</span>
</button>
<button
@click.stop="prepareDelete(disc)"
class="icon-action-btn icon-action-btn-danger delete-button"
:aria-label="`Supprimer le disque ${disc.title}`"
:disabled="isLoading"
>
<span class="icon" aria-hidden="true">🗑️</span>
</button>
<button
v-if="disc.notes"
@mouseenter="showNotesTooltip(disc, $event)"
@mouseleave="hideNotesTooltip"
class="icon-action-btn info-button"
:aria-label="`Afficher les notes pour ${disc.title}`"
>
<span class="icon" aria-hidden="true">ℹ️</span>
</button>
</div>
</td>
</tr>
</tbody>
</table>
</div>
<!-- ✅ CARDS : Mobile uniquement -->
<div v-else class="cards-container">
<div
v-for="(disc, index) in sortedAndPaginatedDiscs"
:key="disc.id"
class="card-item"
:class="{ 'selected-row': selectedRowId === disc.id }"
>
<!-- Contenu principal de la carte -->
<div class="card-main">
<!-- Pochette cliquable -->
<div
class="card-cover"
@click.stop="showCoverPreview(disc, $event)"
>
<img
v-if="disc.cover_url"
:src="getImageUrl(disc.cover_url, disc.updated_at)"
:alt="`Pochette: ${disc.title}`"
class="cover-thumbnail"
@error="handleImageError($event, disc)"
/>
<div v-else class="cover-fallback">
<span class="disc-icon">💿</span>
</div>
</div>
<!-- Infos principales cliquables -->
<div class="card-content" @click="handleRowClick(disc)" @dblclick="handleRowDoubleClick(disc)">
<div class="card-title">{{ displayValue(disc.title) }}</div>
<div class="card-artist">{{ displayValue(disc.artist_name) }}</div>
<div class="card-details">
<span v-if="disc.release_year">{{ disc.release_year }} • </span>
<span>{{ formatPrice(disc.price) }}</span>
</div>
</div>
</div>
<!-- Actions en dessous -->
<div class="card-actions-bottom">
<button
@click.stop="openModal(disc)"
class="action-button-mobile edit-button"
:disabled="isLoading || isSaving"
>
<span class="icon" aria-hidden="true">✏️</span>
<span>Modifier</span>
</button>
<button
@click.stop="prepareDelete(disc)"
class="action-button-mobile delete-button"
:disabled="isLoading"
>
<span class="icon" aria-hidden="true">🗑️</span>
<span>Supprimer</span>
</button>
</div>
</div>
</div>
</div>
<!-- ✅ MESSAGE TABLEAU VIDE -->
<div
v-if="sortedAndPaginatedDiscs.length === 0 && !isLoading"
class="empty-table-message-standalone"
role="status"
aria-live="polite"
>
<template v-if="discs.length === 0">
Aucun disque dans la collection.
<button
@click="openModal()"
class="text-button"
:disabled="isLoading"
>
Ajoutez le premier disque !
</button>
</template>
<template v-else>
Aucun disque ne correspond aux critères. Essayez d'élargir la recherche ou de réinitialiser les filtres.
</template>
</div>
<!-- ✅ SUMMARY BOX — CORRIGÉ -->
<!-- ✅ SUMMARY BOX AVEC TOTAL GLOBAL -->
<div v-else class="table-summary-wrapper-standalone">
<div class="summary-box">
<span class="summary-count">
{{ sortedAndPaginatedDiscs.length }} support{{ sortedAndPaginatedDiscs.length > 1 ? 's' : '' }}
sur
<strong>{{ totalItems }}</strong> support{{ totalItems > 1 ? 's' : '' }} au total
</span>
</div>
<!-- ✅ NOUVEAU: Total global en € -->
<div class="total-value-box">
<span class="total-label">💰 Valeur totale de la collection :</span>
<span class="total-amount">{{ formatPrice(totalValue) }}</span>
</div>
</div>
<!-- ✅ PAGINATION -->
<div
v-if="totalPages > 1 && sortedAndPaginatedDiscs.length > 0"
class="pagination-bottom"
role="navigation"
aria-label="Pagination"
>
<div class="pagination-controls">
<!-- Bouton Première page -->
<button
@click="goToFirstPage"
:disabled="currentPage <= 1 || isLoading"
class="pagination-button pagination-first"
aria-label="Première page"
title="Première page"
>
⏮️
</button>
<!-- Bouton Précédent -->
<button
@click="goToPage(currentPage - 1)"
:disabled="currentPage <= 1 || isLoading"
class="pagination-button"
aria-label="Page précédente"
>
‹ Précédent
</button>
<!-- Numéros de pages -->
<span class="pagination-numbers">
<button
v-for="page in visiblePages"
:key="page"
@click="goToPage(page)"
:class="{ 'active': currentPage === page }"
class="page-button"
:aria-label="`Page ${page}`"
:aria-current="currentPage === page ? 'page' : null"
:disabled="isLoading"
>
{{ page }}
</button>
</span>
<!-- Bouton Suivant -->
<button
@click="goToPage(currentPage + 1)"
:disabled="currentPage >= totalPages || isLoading"
class="pagination-button"
aria-label="Page suivante"
>
Suivant ›
</button>
<!-- Bouton Dernière page -->
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
<!-- Zone saisie directe numéro de page -->
<div class="pagination-goto">
<label for="goto-page" class="goto-label">Aller à la page :</label>
<input
id="goto-page"
v-model="pageInputValue"
type="number"
min="1"
:max="totalPages"
class="goto-input"
placeholder="#"
@keydown.enter="goToInputPage"
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
<!-- ✅ MODALE PRINCIPALE -->
<DiscsModal
:is-open="isModalOpen"
:disc-data="currentDisc"
:api-error="apiError"
:is-saving="isSaving"
@close="closeModal"
@save="handleSaveWithValidation"
@create-artist="handleCreateArtist"
@edit-existing-disc="handleEditExistingDisc"
/>
<!-- ✅ MODALE CONFIRMATION SUPPRESSION (Design unifié) -->
<Teleport to="body">
<div
v-if="isConfirmModalOpen"
class="modal-overlay"
@click.self="isConfirmModalOpen = false"
role="dialog"
aria-labelledby="confirm-delete-title"
aria-modal="true"
>
<div class="modal-card">
<div class="modal-header">
<div class="header-title-container">
<span class="title-icon">🗑️</span>
<div class="header-text">
<h2 id="confirm-delete-title">Confirmer la Suppression</h2>
<p class="header-subtitle">
Cette action est irréversible
</p>
</div>
</div>
<button class="icon-action-btn" @click="isConfirmModalOpen = false" aria-label="Fermer">
<span>&times;</span>
</button>
</div>
<div v-if="discToDelete" class="delete-disc-info">
<div class="info-section">
<h4>📋 Disque à supprimer</h4>
<div class="disc-details">
<div class="detail-row">
<strong>Titre:</strong> {{ discToDelete.title }}
</div>
<div class="detail-row">
<strong>Artiste:</strong> {{ discToDelete.artist_name || '-' }}
</div>
<div v-if="discToDelete.barcode" class="detail-row">
<strong>Code-barres:</strong> {{ discToDelete.barcode }}
</div>
<div class="detail-row">
<strong>ID:</strong> {{ discToDelete.id }}
</div>
</div>
</div>
</div>
<div class="modal-actions">
<button
@click="isConfirmModalOpen = false"
class="ghost-btn"
:disabled="isLoading"
>
❌ Annuler
</button>
<button
@click="executeDelete"
class="danger-btn"
:disabled="isLoading"
>
🗑️ Supprimer Définitivement
</button>
</div>
</div>
</div>
</Teleport>
</div>
</template>
<style scoped>
/* ——— VARIABLES CSS ——— */
.discs-view {
/* ——— Ponts vers les tokens partagés (App.vue) : ne plus jamais coder les
   couleurs en dur ici, tout doit passer par ces variables locales ——— */
--action-edit-color: var(--accent-soft);
--header-separator-color: var(--accent-soft);
--header-bg-color: var(--table-header-color);
--header-separator-vertical: rgba(var(--tint-rgb), 0.3);
--font-size-base: 16px;
--font-size-large: 1.1em;
--font-size-xlarge: 1.2em;
--table-font-size: 0.9em;
--table-data-font-size: 0.85em;
--table-border-color: var(--line);
--table-border-dark: var(--line);
--table-border-light: var(--line-soft);
--row-hover-color: rgba(var(--tint-rgb), 0.06);
--row-selected-color: rgba(59, 130, 246, 0.12);
--row-selected-accent: var(--accent);
--color-primary: var(--accent);
--color-secondary: var(--accent-soft);
--color-danger: var(--negative-text);
--color-success: var(--positive-text);
--color-warning: var(--accent-sand);
--color-info: var(--accent-blue);
font-size: var(--font-size-base);
padding: 20px;
font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
background: transparent;
min-height: 100vh;
}
/* ✅ MESSAGE INFORMATIF */
.info-message {
display: flex;
align-items: center;
gap: 8px;
padding: 10px 16px;
background: rgba(59, 130, 246, 0.1);
border: 1px solid rgba(59, 130, 246, 0.24);
border-radius: 6px;
margin-bottom: 16px;
color: var(--accent-soft);
font-size: 0.9em;
}
.info-icon {
font-size: 1.1em;
}
/* ✅ HEADER */
.header-wrapper {
padding: 0;
margin: 0;
}

.header-top-row {
display: flex;
justify-content: space-between;
align-items: center;
margin-bottom: 5px;
flex-wrap: wrap;
gap: 16px;
}

/* Title section avec bouton retour */
.title-section {
display: flex;
align-items: center;
gap: 12px;
}

/* Bouton retour Dashboard (responsive uniquement) */
.back-button {
display: none; /* Masqué par défaut en desktop */
background: linear-gradient(135deg, #4b3d7a 0%, #d87d3a 100%);
border: 3px solid rgba(255, 255, 255, 0.3);
border-radius: 50%;
color: white;
cursor: pointer;
padding: 0;
font-size: 1.3em;
transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
box-shadow: 
  0 4px 12px rgba(75, 61, 122, 0.4),
  0 2px 6px rgba(0, 0, 0, 0.15),
  inset 0 -2px 0 rgba(0, 0, 0, 0.15),
  inset 0 2px 6px rgba(255, 255, 255, 0.3);
align-items: center;
justify-content: center;
width: 38px;
height: 38px;
position: relative;
overflow: visible;
}

@media (max-width: 767px) {
.back-button {
  display: flex; /* Visible en mobile */
}
}

.back-button::before {
content: '';
position: absolute;
top: 6px;
left: 6px;
right: 6px;
height: 40%;
background: radial-gradient(ellipse at center top, rgba(255, 255, 255, 0.4), transparent);
border-radius: 50% 50% 50% 50% / 60% 60% 40% 40%;
pointer-events: none;
}

.back-button:hover {
background: linear-gradient(135deg, #5a4d8f 0%, #e89248 100%);
transform: scale(1.08) translateY(-2px);
box-shadow: 
  0 6px 16px rgba(75, 61, 122, 0.5),
  0 3px 10px rgba(216, 125, 58, 0.3),
  inset 0 -2px 0 rgba(0, 0, 0, 0.15),
  inset 0 2px 8px rgba(255, 255, 255, 0.4);
border-color: rgba(255, 255, 255, 0.5);
}

.title-icon {
color: var(--text);
font-size: 2em;
}

.page-header h1 {
color: var(--text);
font-size: 2em;
margin: 0;
font-weight: bold;
}
.page-header .subtitle {
color: var(--text-soft);
margin: 0 0 10px 0;
font-size: var(--font-size-large);
}
.user-hint {
color: var(--text-soft);
margin: 0 0 20px 0;
font-size: var(--font-size-base);
background: rgba(var(--tint-rgb), 0.04);
padding: 8px 12px;
border-radius: 6px;
border-left: 4px solid var(--accent);
}
.user-hint strong {
color: var(--text);
}
.toolbar {
margin: 8px 0 0 0;
text-align: right;
display: flex;
gap: 10px;
align-items: center;
flex-wrap: wrap;
}
/* ✅ BOUTONS TOOLBAR */
.filter-toggle-button,
.export-button {
display: inline-flex;
align-items: center;
padding: 10px 16px;
border: none;
border-radius: 8px;
cursor: pointer;
font-weight: 600;
font-size: var(--font-size-large);
transition: all 0.3s ease;
min-height: 44px;
}
/* .add-button utilise désormais .primary-btn (pilule/dégradé partagés) : on ne
   garde ici que la mise en page de l'icône, pas de couleur/forme locale. */
.add-button {
display: inline-flex;
align-items: center;
}
.filter-toggle-button {
background: var(--color-primary);
color: white;
}
.filter-toggle-button:hover:not(:disabled) {
filter: brightness(1.12);
transform: translateY(-2px);
}
.filter-toggle-button.filters-hidden {
background: var(--text-dim);
}
.filter-toggle-button.filters-hidden:hover:not(:disabled) {
filter: brightness(1.15);
}
.export-button {
background: var(--color-success);
color: white;
}
.export-button:hover:not(:disabled) {
filter: brightness(1.1);
transform: translateY(-2px);
}
.export-button:disabled {
opacity: 0.6;
cursor: not-allowed;
}
.filter-toggle-button .icon,
.export-button .icon,
.add-button .icon {
margin-right: 8px;
font-size: 1.1em;
}
/* ✅ BARRE FILTRES */
.filter-bar {
background: var(--bg-elevated);
border: 1px solid var(--line);
padding: 16px;
border-radius: 10px;
box-shadow: var(--shadow);
margin-bottom: 12px;
display: flex;
flex-wrap: wrap;
gap: 16px;
align-items: flex-end;
}
.search-group {
flex: 1;
min-width: 200px;
}
.search-group label,
.filter-group label {
display: block;
margin-bottom: 6px;
font-weight: 600;
color: var(--text-soft);
font-size: 0.9em;
}
.search-wrapper {
position: relative;
}
.search-input {
width: 100%;
padding: 10px 36px 10px 14px;
border: 1px solid var(--line);
border-radius: 6px;
background: rgba(var(--tint-rgb), 0.04);
color: var(--text);
font-size: var(--table-data-font-size);
transition: border-color 0.2s;
min-height: 44px;
}
.search-input:focus {
border-color: var(--color-secondary);
outline: none;
box-shadow: 0 0 0 2px rgba(96, 165, 250, 0.2);
}
.search-input:disabled {
background-color: rgba(var(--tint-rgb), 0.02);
cursor: not-allowed;
}
.search-icon {
position: absolute;
right: 12px;
top: 50%;
transform: translateY(-50%);
color: var(--text-dim);
font-size: 14px;
}
.filters-grid {
display: grid;
grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
gap: 16px;
flex: 3;
}
.filter-select {
width: 100%;
padding: 10px 32px 10px 12px;
border: 1px solid var(--line);
border-radius: 6px;
background: var(--bg-elevated);
color: var(--text);
font-size: var(--table-data-font-size);
cursor: pointer;
min-height: 44px;
-webkit-appearance: none;
-moz-appearance: none;
appearance: none;
background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23888' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
background-repeat: no-repeat;
background-position: right 10px center;
background-size: 16px;
}
.filter-select:focus {
border-color: var(--color-secondary);
outline: none;
box-shadow: 0 0 0 2px rgba(96, 165, 250, 0.2);
}
.filter-select:disabled {
background-color: rgba(var(--tint-rgb), 0.02);
cursor: not-allowed;
opacity: 0.7;
}
.reset-button {
padding: 10px 16px;
background: var(--text-dim);
color: white;
border: none;
border-radius: 6px;
cursor: pointer;
font-weight: 600;
white-space: nowrap;
display: flex;
align-items: center;
gap: 6px;
min-height: 44px;
transition: all 0.3s ease;
}
.reset-button:hover:not(:disabled) {
filter: brightness(1.15);
transform: translateY(-2px);
}
.reset-button:disabled {
opacity: 0.6;
cursor: not-allowed;
}
/* ✅ BADGES FILTRES ACTIFS */
.active-filters-badges {
display: flex;
flex-wrap: wrap;
gap: 8px;
margin-bottom: 16px;
padding: 12px;
background: rgba(var(--tint-rgb), 0.04);
border-radius: 8px;
border-left: 4px solid var(--color-secondary);
}
.filters-title {
font-weight: 700;
color: var(--text);
margin-right: 12px;
align-self: center;
}
.filter-badge {
display: inline-flex;
align-items: center;
gap: 6px;
background: rgba(var(--tint-rgb), 0.08);
color: var(--text);
padding: 6px 12px;
border-radius: 20px;
font-size: 0.85em;
font-weight: 600;
cursor: pointer;
transition: all 0.2s ease;
border: 1px solid transparent;
}
.filter-badge:hover {
background: rgba(var(--tint-rgb), 0.14);
border-color: var(--accent);
}
.filter-badge:focus {
outline: 2px solid var(--color-secondary);
outline-offset: 2px;
}
.remove-icon {
display: inline-flex;
align-items: center;
justify-content: center;
width: 18px;
height: 18px;
border-radius: 50%;
background: rgba(var(--tint-rgb), 0.22);
color: var(--text);
font-size: 12px;
font-weight: bold;
transition: background 0.2s;
}
.filter-badge:hover .remove-icon {
background: rgba(var(--tint-rgb), 0.32);
}
/* ✅ CHARGEMENT */
.loading-indicator {
display: flex;
align-items: center;
justify-content: center;
gap: 12px;
padding: 30px;
background: var(--bg-elevated);
border: 1px solid var(--line);
border-radius: 10px;
margin-bottom: 16px;
font-size: var(--font-size-large);
color: var(--text-dim);
}
.spinner {
width: 24px;
height: 24px;
border: 3px solid var(--line-soft);
border-top: 3px solid var(--color-secondary);
border-radius: 50%;
animation: spin 1s linear infinite;
}
@keyframes spin {
0% { transform: rotate(0deg); }
100% { transform: rotate(360deg); }
}
/* ✅ ERREUR */
.error-message {
display: flex;
align-items: center;
gap: 12px;
padding: 16px;
background: rgba(239, 68, 68, 0.12);
border: 1px solid rgba(239, 68, 68, 0.28);
border-radius: 8px;
margin-bottom: 16px;
color: var(--negative-text);
font-weight: 600;
}
.error-icon {
font-size: 1.2em;
flex-shrink: 0;
}
.error-text {
flex: 1;
}
.error-close {
margin-left: auto;
background: none;
border: none;
font-size: 1.5em;
cursor: pointer;
color: var(--negative-text);
padding: 0;
width: 24px;
height: 24px;
display: flex;
align-items: center;
justify-content: center;
border-radius: 4px;
transition: background 0.2s;
}
.error-close:hover {
background: rgba(239, 68, 68, 0.14);
}
/* ✅ PAGINATION */
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
border-color: var(--color-secondary);
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
border-color: var(--color-secondary);
}
.page-button.active {
background: var(--color-secondary);
color: white;
border-color: var(--color-secondary);
}
.page-button:disabled {
opacity: 0.6;
cursor: not-allowed;
}
/* 🆕 BOUTONS PREMIÈRE/DERNIÈRE PAGE */
.pagination-first,
.pagination-last {
min-width: 40px;
font-size: 1.2em;
padding: 6px 12px;
}
/* 🆕 ZONE SAISIE DIRECTE PAGE */
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
border-color: var(--color-secondary);
box-shadow: 0 0 0 2px rgba(96, 165, 250, 0.15);
}
.goto-input:disabled {
background: rgba(var(--tint-rgb), 0.02);
cursor: not-allowed;
}
/* Masquer les flèches spinner du input number */
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
/* ✅ TABLEAU — le fond/ombre/bordure viennent désormais de la classe .panel
   partagée (ajoutée sur cet élément dans le template) ; on ne garde ici que
   l'agencement interne. */
.content-panel {
padding: 20px;
}
.table-responsive {
width: 100%;
max-width: 100%;
overflow-x: auto;
margin-bottom: 20px;
/* ✅ AJOUTÉ — bordure comme dans FormatsView.vue */
border: 1px solid var(--table-border-color);
border-radius: 8px;
background: transparent;
/* ⚡ Optimisations pour le scroll fluide */
-webkit-overflow-scrolling: touch;
scroll-behavior: auto; /* Pas de smooth scroll pour éviter les delays */
	/* 🆕 AMÉLIORATION RESPONSIVE pour la sidebar */
	position: relative;
	display: block;
	box-sizing: border-box;
}
.data-table {
width: 100%;
border-collapse: separate;
border-spacing: 0;
margin-bottom: 20px;
font-size: var(--table-font-size);
/* ✅ CORRIGÉ — suppression de la bordure double */
border: none;
border-radius: 8px;
overflow: hidden;
min-width: auto;        /* ✅ MODIFICATION 1/3 */
table-layout: fixed;    /* ✅ MODIFICATION 2/3 */
background: transparent;
}
.data-table thead {
background-color: var(--header-bg-color);
}
.data-table tbody {
background: transparent;
}
.data-table thead th {
font-weight: bold;
padding: 12px 15px;
border: none;
border-bottom: 3px solid var(--header-separator-color);
background: var(--header-bg-color);
color: white;
text-align: center;
font-size: var(--table-font-size);
height: 50px;
white-space: nowrap;
}
.data-table thead th::after {
content: '';
position: absolute;
right: 0;
top: 0;
height: 100%;
width: 1px;
background-color: var(--header-separator-vertical);
}
.data-table thead th:last-child::after {
display: none;
}
.data-table thead th .header-text {
font-weight: bold;
font-size: var(--table-font-size);
}
.data-table tbody tr {
transition: all 0.3s ease;
cursor: pointer;
height: 40px;
background: transparent;
}
.data-table tbody tr.odd-row {
background: var(--odd-row-color);
}
.data-table tbody tr.even-row {
background: var(--even-row-color);
}
.data-table tbody tr:hover {
background: var(--row-hover-color);
transform: translateY(-1px);
box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
.data-table tbody tr.selected-row {
background: var(--row-selected-color) !important;
border-left: 4px solid var(--row-selected-accent) !important;
position: relative;
}
.data-table tbody tr.selected-row td {
font-weight: 600;
color: var(--text);
}
.data-table td {
padding: 6px 15px;
border-bottom: 1px solid var(--table-border-color);
border-right: 1px solid var(--table-border-light);
text-align: left;
font-size: var(--table-data-font-size);
display: flex;
align-items: center;
height: 40px;
white-space: nowrap;
overflow: hidden;
text-overflow: ellipsis;
background: transparent;
}
.data-table td:last-child {
border-right: none;
text-align: center;
justify-content: center;
}
/* ✅ POCHETTE */
.data-table .cover-column {
min-width: 80px;
text-align: center;
padding: 4px;
flex-shrink: 0;
}
.cover-thumbnail-container {
width: 40px;
height: 40px;
display: flex;
align-items: center;
justify-content: center;
margin: 0 auto;
position: relative;
cursor: pointer;
border-radius: 4px;
overflow: hidden;
/* Effet d'agrandissement supprimé */
}
.cover-thumbnail {
width: 100%;
height: 100%;
object-fit: cover;
border-radius: 4px;
box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}
.cover-fallback {
width: 100%;
height: 100%;
background: rgba(var(--tint-rgb), 0.08);
border-radius: 4px;
display: flex;
align-items: center;
justify-content: center;
font-size: 20px;
flex-direction: column;
gap: 2px;
}
.cover-fallback-text {
font-size: 8px;
color: var(--text-dim);
text-align: center;
max-width: 100%;
overflow: hidden;
text-overflow: ellipsis;
white-space: nowrap;
}
/* LARGEURS DES COLONNES */
.data-table .cover-column { width: 60px; flex: 0 0 60px; }
.data-table .title-column { width: 19%; flex: 0 0 19%; }
.data-table .artist-column { width: 15%; flex: 0 0 15%; }
.data-table .format-column { width: 12%; flex: 0 0 12%; }
.data-table .label-column { width: 15%; flex: 0 0 15%; }
.data-table .barcode-column { width: 15%; flex: 0 0 15%; min-width: 110px; text-align: center; font-variant-numeric: tabular-nums; }
.data-table .price-column { width: 9%; flex: 0 0 9%; text-align: right; font-variant-numeric: tabular-nums; }
.data-table .actions-column { width: 10%; flex: 0 0 10%; text-align: center; }
/* ✅ ACTIONS — les boutons desktop utilisent les classes partagées
   .icon-action-btn / .icon-action-btn-danger (App.vue) pour la forme, le fond
   et le hover ; on ne garde ici que la couleur d'icône par variante. */
.action-buttons-container {
display: flex;
gap: 4px;
justify-content: center;
align-items: center;
width: 100%;
height: 100%;
}
.icon-action-btn .icon {
font-size: 18px;
line-height: 1;
display: flex;
align-items: center;
justify-content: center;
}
.edit-button .icon {
color: var(--action-edit-color);
}
.delete-button .icon {
color: var(--negative-text);
}
/* ✅ TRI */
.sortable {
cursor: pointer;
transition: background 0.2s;
}
.sortable:hover {
background: rgba(var(--tint-rgb), 0.12);
}
.sortable:focus {
outline: 2px solid var(--color-secondary);
outline-offset: -2px;
}
.sortable-content {
display: flex;
justify-content: space-between;
align-items: center;
width: 100%;
padding: 0 8px;
}
.header-text {
text-align: left;
overflow: hidden;
text-overflow: ellipsis;
white-space: nowrap;
flex: 1;
}
.sort-icon {
flex-shrink: 0;
white-space: nowrap;
font-size: 0.9em;
margin-left: 8px;
}
/* Centrer les colonnes non triables */
.cover-column .sortable-content,
.actions-column .sortable-content {
justify-content: center;
}
.cover-column .header-text,
.actions-column .header-text {
text-align: center;
}
/* ✅ MESSAGE TABLEAU VIDE */
.empty-table-message-standalone {
text-align: center;
padding: 30px;
color: var(--text-dim);
font-style: italic;
font-size: var(--font-size-large);
border: 2px dashed var(--table-border-color);
border-radius: 8px;
margin: 20px 0;
background: rgba(var(--tint-rgb), 0.02);
}
.text-button {
background: none;
border: none;
color: var(--color-secondary);
text-decoration: underline;
cursor: pointer;
font-weight: 600;
padding: 4px 8px;
border-radius: 4px;
transition: all 0.2s;
}
.text-button:hover:not(:disabled) {
color: var(--accent);
background: rgba(var(--tint-rgb), 0.08);
}
.text-button:disabled {
opacity: 0.5;
cursor: not-allowed;
}
/* ✅ RÉSUMÉ */
.table-summary-wrapper-standalone {
display: flex;
flex-direction: column;
gap: 8px;
padding: 16px 0;
}
.summary-box {
display: flex;
justify-content: center;
align-items: center;
padding: 12px 20px;
background: rgba(var(--tint-rgb), 0.05);
border-radius: 8px;
border: 1px solid var(--line);
}
.summary-count {
color: var(--text-soft);
font-size: 0.95em;
}
.summary-count strong {
color: var(--accent-soft);
font-weight: 600;
}
.total-value-box {
display: flex;
align-items: center;
justify-content: flex-end;
gap: 12px;
padding: 16px 20px;
background: linear-gradient(135deg, var(--accent), var(--accent-blue));
border-radius: 8px;
margin-top: 12px;
box-shadow: var(--shadow);
}
.total-label {
color: rgba(255, 255, 255, 0.9);
font-size: 1em;
font-weight: 500;
}
.total-amount {
color: var(--accent-sand);
font-size: 1.3em;
font-weight: 700;
text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}
/* ✅ RESPONSIVE PRINCIPAL */
@media (max-width: 1024px) {
.filters-grid {
grid-template-columns: repeat(2, 1fr);
}
.toolbar {
flex-direction: column;
align-items: stretch;
}
.filter-toggle-button,
.export-button,
.add-button {
width: 100%;
justify-content: center;
}
}
/* ✅ TABLEAU RESPONSIVE — CARTES MOBILE */
@media (max-width: 768px) {
.discs-view {
padding: 10px;
}
.header-top-row {
flex-direction: column;
align-items: flex-start;
}
.toolbar {
width: 100%;
margin-top: 16px;
}
.filter-bar {
flex-direction: column;
align-items: stretch;
}
.search-group {
width: 100%;
}
.filters-grid {
grid-template-columns: 1fr;
}
.pagination-bottom {
flex-direction: column;
align-items: stretch;
text-align: center;
}
.pagination-controls {
justify-content: center;
}
.action-buttons-container {
gap: 6px;
}
.icon-action-btn {
width: 36px;
height: 36px;
font-size: 16px;
}
.data-table {
font-size: 0.85em;
min-width: auto !important;
width: 100% !important;
}
/* → Table → cartes */
/* ... (supprimé pour concision — déjà remplacé par le mode cartes ci-dessus) */
}
/* ✅ CARTES MOBILE – responsive < 768px */
@media (max-width: 767px) {
.cards-container {
display: flex;
flex-direction: column;
gap: 16px;
padding: 0;
}
.card-item {
display: flex;
flex-direction: column;
background: var(--bg-elevated);
border-radius: 12px;
box-shadow: var(--shadow);
border-left: 4px solid var(--accent);
transition: all 0.2s;
overflow: hidden;
}
.card-item.selected-row {
border-left-color: var(--accent-soft);
background: rgba(59, 130, 246, 0.08);
}
.card-main {
display: flex;
gap: 12px;
padding: 14px;
}
.card-cover {
flex-shrink: 0;
width: 80px;
height: 80px;
border-radius: 8px;
overflow: hidden;
cursor: pointer;
transition: transform 0.2s;
}
.card-cover:active {
transform: scale(0.95);
}
.cover-thumbnail,
.cover-fallback {
width: 100%;
height: 100%;
object-fit: cover;
display: flex;
align-items: center;
justify-content: center;
font-size: 28px;
}
.card-content {
flex: 1;
min-width: 0;
display: flex;
flex-direction: column;
gap: 6px;
cursor: pointer;
}
.card-title {
font-weight: 700;
font-size: 1em;
margin: 0;
color: var(--text);
line-height: 1.3;
word-wrap: break-word;
overflow-wrap: break-word;
}
.card-artist {
font-size: 0.9em;
color: var(--text-soft);
margin: 0;
font-weight: 500;
line-height: 1.3;
word-wrap: break-word;
overflow-wrap: break-word;
}
.card-details {
font-size: 0.85em;
color: var(--text-dim);
display: flex;
gap: 4px;
align-items: center;
margin-top: auto;
}
.card-actions-bottom {
display: flex;
gap: 8px;
padding: 10px 14px;
background: rgba(var(--tint-rgb), 0.04);
border-top: 1px solid var(--line-soft);
}
.action-button-mobile {
flex: 1;
display: flex;
align-items: center;
justify-content: center;
gap: 6px;
padding: 10px 16px;
border: none;
border-radius: 8px;
font-size: 0.9em;
font-weight: 600;
cursor: pointer;
transition: all 0.2s;
}
.action-button-mobile .icon {
font-size: 18px;
}
.action-button-mobile.edit-button {
background: var(--accent);
color: white;
}
.action-button-mobile.edit-button:active {
filter: brightness(0.9);
}
.action-button-mobile.delete-button {
background: var(--negative-text);
color: white;
}
.action-button-mobile.delete-button:active {
filter: brightness(0.9);
}
.action-button-mobile:disabled {
opacity: 0.5;
cursor: not-allowed;
}
}
/* ✅ ULTRA-MOBILE */
@media (max-width: 480px) {
.page-header h1 {
font-size: 1.5em;
}
.modal-actions {
flex-direction: column;
}
.save-button,
.cancel-button {
width: 100%;
}
}
/* ✅ TOUCH FRIENDLY */
@media (hover: none) and (pointer: coarse) {
.icon-action-btn,
.pagination-button,
.page-button,
.filter-badge,
tr {
min-height: 48px;
min-width: 48px;
}
}
</style>
<style scoped>
/* ✅ ÉLÉMENTS RENDUS PAR LE TEMPLATE VUE — scopés (déplacés depuis l'ancien
   bloc <style> global qui les laissait fuiter sur le reste de l'appli) */
.column-separator {
position: absolute;
top: 0;
bottom: 0;
right: -4px;
width: 8px;
cursor: col-resize;
z-index: 10;
background: transparent;
pointer-events: auto;
}
.column-separator::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 2px;
  height: 20px;
  background: rgba(var(--tint-rgb), 0.3);
  opacity: 0;
  transition: opacity 0.2s;
  pointer-events: none;
}
.column-separator:hover::after {
  opacity: 1;
}
.sr-only {
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
button:focus-visible,
input:focus-visible,
select:focus-visible,
.filter-badge:focus-visible {
outline: 3px solid var(--accent);
outline-offset: 2px;
}
@media print {
.toolbar,
.filter-bar,
.active-filters-badges,
.pagination-bottom,
.actions-column {
display: none !important;
}
.data-table {
border: 1px solid #000;
box-shadow: none;
}
.discs-view {
background: white;
padding: 0;
}
.content-panel {
box-shadow: none;
padding: 0;
}
}
/* ✅ MODALE SUPPRESSION (Teleport) — utilise .modal-overlay/.modal-card/.modal-header/.modal-actions globaux (App.vue) */
.delete-disc-info {
background: rgba(239, 68, 68, 0.12);
border: 1px solid rgba(239, 68, 68, 0.28);
border-radius: 14px;
padding: 20px;
margin-bottom: 24px;
}
.delete-disc-info .info-section h4 {
margin: 0 0 12px 0;
color: var(--text);
}
.delete-disc-info .disc-details {
display: flex;
flex-direction: column;
gap: 12px;
}
.delete-disc-info .detail-row {
display: flex;
justify-content: space-between;
padding: 8px 0;
border-bottom: 1px solid var(--line-soft);
color: var(--text-soft);
}
.delete-disc-info .detail-row:last-child {
border-bottom: none;
}
/* 🆕 MENU COLONNES */
.column-menu-wrapper {
position: relative;
}
.column-menu {
position: absolute;
top: 100%;
right: 0;
margin-top: 8px;
background: var(--modal-bg);
border: 1px solid var(--line-soft);
border-radius: 12px;
box-shadow: var(--shadow);
min-width: 220px;
z-index: 10000;
}
.column-menu-header {
display: flex;
justify-content: space-between;
align-items: center;
padding: 12px 16px;
background: linear-gradient(135deg, var(--accent), var(--accent-blue));
border-radius: 12px 12px 0 0;
}
.column-menu-header h3 {
margin: 0;
color: white;
font-size: 1em;
font-weight: 600;
}
.btn-reset {
background: rgba(255, 255, 255, 0.2);
border: none;
color: white;
width: 28px;
height: 28px;
border-radius: 50%;
cursor: pointer;
font-size: 1em;
display: flex;
align-items: center;
justify-content: center;
transition: all 0.2s;
}
.btn-reset:hover {
background: rgba(255, 255, 255, 0.3);
transform: scale(1.1);
}
.column-menu-list {
max-height: 400px;
overflow-y: auto;
padding: 8px 0;
}
.column-menu-item {
display: flex;
align-items: center;
justify-content: space-between;
gap: 10px;
padding: 10px 16px;
cursor: pointer;
transition: background 0.2s;
margin: 0;
}
.column-menu-item:hover:not(.disabled) {
background: rgba(var(--tint-rgb), 0.06);
}
.column-menu-item.disabled {
opacity: 0.5;
cursor: not-allowed;
}
.column-label {
flex: 1;
color: var(--text);
font-size: 0.95em;
font-weight: 500;
text-align: left;
}
/* 🆕 TOGGLE SWITCH */
.toggle-switch {
position: relative;
display: inline-block;
width: 44px;
height: 24px;
margin: 0;
cursor: pointer;
}
.toggle-switch input {
opacity: 0;
width: 0;
height: 0;
}
.toggle-slider {
position: absolute;
cursor: pointer;
top: 0;
left: 0;
right: 0;
bottom: 0;
background-color: var(--line);
transition: 0.3s;
border-radius: 24px;
}
.toggle-slider:before {
position: absolute;
content: "";
height: 18px;
width: 18px;
left: 3px;
bottom: 3px;
background-color: white;
transition: 0.3s;
border-radius: 50%;
box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}
.toggle-switch input:checked + .toggle-slider {
background: linear-gradient(135deg, var(--accent), var(--accent-blue));
}
.toggle-switch input:checked + .toggle-slider:before {
transform: translateX(20px);
}
.toggle-switch input:disabled + .toggle-slider {
background-color: var(--line);
cursor: not-allowed;
opacity: 0.6;
}
.toggle-switch input:disabled + .toggle-slider:before {
background-color: var(--bg-elevated);
}
.toggle-switch input:disabled + .toggle-slider {
cursor: not-allowed;
}
.disabled .toggle-switch {
cursor: not-allowed;
}
/* 🆕 RESPONSIVE PAGINATION */
@media (max-width: 767px) {
.pagination-bottom {
flex-direction: column;
gap: 12px;
}
.pagination-controls {
width: 100%;
justify-content: center;
}
.pagination-button {
padding: 6px 12px;
font-size: 0.9em;
min-height: 32px;
}
.pagination-first,
.pagination-last {
font-size: 1em;
padding: 6px 10px;
min-width: 36px;
}
.page-button {
padding: 6px 10px;
min-width: 36px;
font-size: 0.9em;
}
.pagination-goto {
width: 100%;
justify-content: center;
padding: 8px 12px;
}
.goto-label {
font-size: 0.85em;
}
.goto-input {
width: 50px;
padding: 5px 6px;
font-size: 0.9em;
}
.goto-button {
padding: 5px 12px;
font-size: 0.85em;
}
.total-pages-info {
font-size: 0.85em;
}
}
/* 🆕 RESPONSIVE POUR LES NOUVELLES COLONNES */
.quantity-column {
text-align: center;
font-weight: 600;
color: var(--text);
}
.genre-column,
.year-column,
.country-column {
text-align: left;
color: var(--text);
}
/* 🔒 COLONNES FIGÉES (STICKY) - VERSION TRANSFORM OPTIMISÉE
   Le fond de ces cellules est fixé dynamiquement en JS (buildStickyCache(),
   via getComputedStyle sur --bg-elevated/--tint-rgb) pour suivre le thème
   actif — pas de background-color ici, ça écraserait ce style inline. */
.data-table td.sticky-column,
.data-table th.sticky-column {
position: relative !important;
z-index: 3;
/* ⚡ OPTIMISATIONS GPU - Forcer la composition sur un layer séparé */
will-change: transform;
backface-visibility: hidden;
-webkit-backface-visibility: hidden;
transform: translateZ(0); /* Force GPU layer dès le départ */
contain: layout style; /* Isolation du reflow */
}
thead .sticky-column {
z-index: 10 !important;
}
.sticky-left {
box-shadow: 2px 0 4px rgba(0, 0, 0, 0.3);
}
.sticky-right {
box-shadow: -2px 0 4px rgba(0, 0, 0, 0.3);
}
.sticky-toggle-btn {
background: transparent;
border: 1px solid var(--line);
color: var(--text-dim);
padding: 4px 8px;
border-radius: 4px;
cursor: pointer;
font-size: 14px;
transition: all 0.2s ease;
margin-right: 8px;
min-width: 32px;
}
.sticky-toggle-btn:hover {
background: rgba(var(--tint-rgb), 0.06);
border-color: var(--line);
color: var(--text);
}
.sticky-toggle-btn.active {
background: rgba(59, 130, 246, 0.16);
border-color: var(--accent);
color: var(--accent);
}
.column-menu-item-controls {
display: flex;
align-items: center;
justify-content: space-between;
width: 100%;
padding: 8px 12px;
}
.column-actions {
display: flex;
align-items: center;
gap: 8px;
}
.column-menu-item {
cursor: default;
}
/* 📌 ICÔNE ÉPINGLE POUR COLONNES FIGÉES */
.pin-icon {
display: inline-block;
margin-right: 6px;
font-size: 0.9em;
opacity: 0.9;
filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.3));
animation: pinPulse 2s ease-in-out infinite;
}
@keyframes pinPulse {
0%, 100% { transform: scale(1); }
50% { transform: scale(1.1); }
}

/* ⚠️ LARGEURS MINIMALES FORCÉES (qualifiées à .data-table pour ne plus
   fuiter sur les autres tableaux de l'appli) */
.data-table th,
.data-table td {
  min-width: 100px !important;
}

.data-table th.cover-column,
.data-table td.cover-column {
  min-width: 80px !important;
}

.data-table th.title-column,
.data-table td.title-column,
.data-table th.artist-column,
.data-table td.artist-column,
.data-table th.barcode-column,
.data-table td.barcode-column {
  min-width: 150px !important;
}

.data-table th.actions-column,
.data-table td.actions-column {
  min-width: 140px !important;
}
</style>

<style>
/* ✅ STYLES GLOBAUX — DOM créé hors du template Vue (document.createElement
   + document.body.appendChild), donc doit rester non scopé. Couleurs
   tokenisées pour suivre le thème actif malgré tout. */
/* ✅ TOOLTIP NOTES */
.notes-tooltip {
position: absolute;
background: var(--modal-bg) !important;
border: 2px solid var(--accent) !important;
border-radius: 8px;
padding: 12px 16px;
width: 350px;
max-width: 90vw;
max-height: 60vh;
box-shadow: var(--shadow) !important;
z-index: 10001;
font-size: 0.9em;
line-height: 1.5;
overflow-y: auto;
color: var(--text) !important;
pointer-events: none;
animation: fadeIn 0.2s ease;
}
.notes-tooltip .tooltip-header {
display: flex;
align-items: center;
gap: 8px;
margin-bottom: 10px;
padding: 8px;
padding-bottom: 10px;
border-bottom: 2px solid var(--accent);
color: var(--accent) !important;
font-weight: 600;
font-size: 1em;
background: rgba(0, 0, 0, 0);
border-radius: 4px;
}
.notes-tooltip .tooltip-icon {
font-size: 16px;
}
.notes-tooltip .tooltip-content {
color: var(--text) !important;
white-space: pre-wrap;
word-wrap: break-word;
font-size: 0.95em;
line-height: 1.6;
background: rgba(var(--tint-rgb), 0.06) !important;
padding: 12px;
border-radius: 6px;
border-left: 3px solid var(--accent);
margin-top: 8px;
}
/* ✅ APERÇU POCHETTE */
.cover-preview {
position: absolute;
background: var(--modal-bg);
border-radius: 12px;
box-shadow: var(--shadow);
padding: 12px;
z-index: 10000;
animation: fadeInScale 0.2s ease-out;
border: 2px solid var(--accent);
max-width: 400px;
}
/* Overlay mobile */
.cover-preview-overlay {
position: fixed;
top: 0;
left: 0;
right: 0;
bottom: 0;
background: var(--modal-overlay);
z-index: 9999;
animation: fadeIn 0.2s ease;
}
/* Bouton de fermeture mobile */
.cover-preview-close {
position: absolute;
top: 8px;
right: 8px;
width: 32px;
height: 32px;
border: none;
background: rgba(0, 0, 0, 0.6);
color: white;
border-radius: 50%;
font-size: 24px;
line-height: 1;
cursor: pointer;
z-index: 1;
display: flex;
align-items: center;
justify-content: center;
padding: 0;
transition: all 0.2s;
}
.cover-preview-close:active {
background: rgba(0, 0, 0, 0.8);
transform: scale(0.95);
}
/* Responsive mobile pour la tooltip */
@media (max-width: 767px) {
.cover-preview {
position: fixed !important;
max-width: 90vw;
max-height: 85vh;
overflow-y: auto;
}
.cover-preview-image-container {
width: 100%;
height: auto;
max-width: 280px;
margin: 0 auto;
}
.cover-preview-image {
width: 100%;
height: auto;
aspect-ratio: 1;
}
}
.cover-preview-content {
display: flex;
flex-direction: column;
gap: 10px;
}
.cover-preview-image-container {
position: relative;
width: 300px;
height: 300px;
}
.cover-preview-image {
width: 100%;
height: 100%;
object-fit: cover;
border-radius: 8px;
border: 1px solid var(--line);
}
.cover-preview-stats {
display: grid;
grid-template-columns: 1fr 1fr;
gap: 8px;
padding: 10px 12px;
background: linear-gradient(135deg, var(--accent), var(--accent-blue));
border-radius: 8px;
}
.preview-stat-item {
display: flex;
align-items: center;
gap: 6px;
min-width: 0;
}
.preview-label {
color: rgba(255, 255, 255, 0.85);
font-size: 0.8em;
font-weight: 700;
white-space: nowrap;
flex-shrink: 0;
}
.preview-value {
color: white;
font-size: 0.85em;
font-weight: 600;
overflow: hidden;
text-overflow: ellipsis;
white-space: nowrap;
flex: 1;
min-width: 0;
}
.preview-price {
color: white;
}
.cover-preview-info {
text-align: center;
font-size: 0.9em;
color: var(--text);
max-width: 300px;
}
.cover-preview-info strong {
color: var(--accent);
display: block;
margin-bottom: 4px;
font-size: 1em;
}
.cover-preview-info div {
color: var(--text-soft);
font-size: 0.85em;
}
/* ✅ ANIMATIONS */
@keyframes fadeInScale {
from {
opacity: 0;
transform: scale(0.8) translateX(-10px);
}
to {
opacity: 1;
transform: scale(1) translateX(0);
}
}
@keyframes fadeIn {
from { opacity: 0; }
to { opacity: 1; }
}
</style>