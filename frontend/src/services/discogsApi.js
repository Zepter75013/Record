// frontend/src/services/discogsApi.js
import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

const getAuthToken = () => localStorage.getItem('token') || localStorage.getItem('auth_token');

const axiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
});

axiosInstance.interceptors.request.use(
  (config) => {
    const token = getAuthToken();
    if (token) config.headers.Authorization = `Bearer ${token}`;
    return config;
  },
  (error) => Promise.reject(error)
);

export const discogsApi = {
  async checkBarcodeExists(barcode) {
    if (!barcode || !barcode.trim()) {
      throw new Error('Le code-barres est requis');
    }
    try {
      const response = await axiosInstance.get(
        `/check-barcode/${encodeURIComponent(barcode.trim())}`
      );
      return response.data.exists ? response.data.disc : null;
    } catch (error) {
      if (error.response?.status === 404) {
        return null;
      }
      throw error;
    }
  },

  async searchByBarcode(barcode) {
    if (!barcode || !barcode.trim()) {
      throw new Error('Le code-barres est requis');
    }
    try {
      const response = await axiosInstance.post('/search-discogs', {
        barcode: barcode.trim()
      });
      const data = response.data;
      if (data.id && data.title) return data;
      if (data.results && Array.isArray(data.results)) {
        return data.results.length === 1 ? data.results[0] : data;
      }
      if (Array.isArray(data)) return data.length === 1 ? data[0] : { results: data };
      return data;
    } catch (error) {
      console.error('Erreur recherche code-barres:', error);
      let errorMessage = 'Erreur lors de la recherche par code-barres';
      if (error.response?.status === 401) {
        errorMessage = 'Non autorisé. Veuillez vous reconnecter.';
      } else if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }
      throw new Error(errorMessage);
    }
  },

  async searchByTitleArtist(title, artist, country = '', year = '') {
    if ((!title || !title.trim()) && (!artist || !artist.trim())) {
      throw new Error('Veuillez fournir au moins un titre ou un artiste');
    }
    try {
      const requestData = {};
      if (title && title.trim()) requestData.title = title.trim();
      if (artist && artist.trim()) requestData.artist = artist.trim();
      if (country && country.trim()) requestData.country = country.trim();
      if (year && year.trim()) requestData.year = year.trim();

      const response = await axiosInstance.post('/search-discogs', requestData);
      const data = response.data;

      if (Array.isArray(data)) return data;
      if (data.results && Array.isArray(data.results)) return data.results;
      if (data.id && data.title) return [data];
      return [];
    } catch (error) {
      console.error('Erreur recherche manuelle:', error);
      let errorMessage = 'Erreur lors de la recherche manuelle';
      if (error.response?.status === 401) {
        errorMessage = 'Non autorisé. Veuillez vous reconnecter.';
      } else if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }
      throw new Error(errorMessage);
    }
  },

  // 🔧 CORRIGÉ : conversion stricte en number
  async getRelease(releaseId) {
    const id = Number(releaseId);
    if (!Number.isInteger(id) || id <= 0 || !isFinite(id)) {
      throw new Error(`ID Discogs invalide : ${releaseId}`);
    }

    try {
      const response = await axiosInstance.post('/select-discogs-result', {
        release_id: id
      });
      return response.data;
    } catch (error) {
      console.error(`❌ getRelease(${id}) échoué:`, error);
      throw new Error(
        error.response?.data?.error ||
        `Échec récupération release ${id}`
      );
    }
  },

  async uploadCoverImage(file, discId = null, onProgress = null) {
    if (!file || !(file instanceof File)) {
      throw new Error('Fichier invalide');
    }
    const validTypes = ['image/jpeg', 'image/png', 'image/webp', 'image/gif'];
    if (!validTypes.includes(file.type)) {
      throw new Error('Type de fichier non supporté. Utilisez JPEG, PNG, WebP ou GIF.');
    }
    const maxSize = 10 * 1024 * 1024;
    if (file.size > maxSize) {
      throw new Error('Le fichier est trop volumineux (max 10MB)');
    }

    const formData = new FormData();
    formData.append('cover', file);
    if (discId) formData.append('disc_id', discId.toString());

    try {
      const response = await axiosInstance.post(
        '/upload-cover',
        formData,
        {
          headers: { 'Content-Type': 'multipart/form-data' },
          onUploadProgress: (progressEvent) => {
            if (onProgress && typeof onProgress === 'function') {
              const percentCompleted = Math.round(
                (progressEvent.loaded * 100) / (progressEvent.total || 1)
              );
              onProgress(percentCompleted);
            }
          },
          timeout: 60000
        }
      );
      return response.data;
    } catch (error) {
      console.error('Erreur upload cover:', error);
      throw new Error(
        error.response?.data?.error ||
        'Erreur lors de l\'upload de l\'image'
      );
    }
  },

  async getReleaseImages(releaseId) {
    if (!releaseId || isNaN(releaseId)) {
      throw new Error('ID de release invalide');
    }
    try {
      const release = await this.getRelease(releaseId);
      return release.images || [];
    } catch (error) {
      console.error('Erreur récupération images:', error);
      return [];
    }
  },

  async advancedSearch(params = {}) {
    return this.searchByTitleArtist(params.title, params.artist, params.country, params.year);
  }
};

export default discogsApi;