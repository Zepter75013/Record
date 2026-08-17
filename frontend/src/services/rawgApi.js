// frontend/src/services/rawgApi.js
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

export const rawgApi = {
  async searchByTitle(title) {
    if (!title || !title.trim()) {
      throw new Error('Le titre est requis');
    }
    try {
      const response = await axiosInstance.post('/search-rawg', { title: title.trim() });
      // Le backend renvoie toujours l'objet GamePreview complet
      // ({found, title, cover_url, results: [...]}).
      return response.data;
    } catch (error) {
      console.error('Erreur recherche RAWG:', error);
      let errorMessage = 'Erreur lors de la recherche du jeu';
      if (error.response?.status === 401) {
        errorMessage = 'Non autorisé. Veuillez vous reconnecter.';
      } else if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }
      throw new Error(errorMessage);
    }
  },

  async getGameDetails(rawgId) {
    const id = Number(rawgId);
    if (!Number.isInteger(id) || id <= 0 || !isFinite(id)) {
      throw new Error(`ID RAWG invalide : ${rawgId}`);
    }
    try {
      const response = await axiosInstance.post('/select-rawg-result', { rawg_id: id });
      return response.data;
    } catch (error) {
      console.error(`Erreur récupération jeu RAWG ${id}:`, error);
      throw new Error(
        error.response?.data?.error || `Échec de la récupération du jeu ${id}`
      );
    }
  },
};

export default rawgApi;
