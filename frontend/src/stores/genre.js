// src/stores/genre.js
import { defineStore } from 'pinia';
import axios from 'axios';
import { useAuthStore } from './auth';

// ✅ CORRECTION: URL relative pour passer par le proxy/Nginx
const API_URL = (import.meta.env.VITE_API_BASE_URL || '/api') + '/genres';

export const useGenreStore = defineStore('genre', {
  state: () => ({
    genres: [], // Tableau pour stocker les objets Genre ({id: 1, name: 'Pop'})
    loading: false,
    error: null,
  }),

  actions: {
    // ------------------------------------------
    // A. Récupérer tous les genres (READ)
    // ------------------------------------------
    async fetchGenres() {
      this.loading = true;
      this.error = null;
      try {
        const authStore = useAuthStore();
        
        const response = await axios.get(API_URL, {
          headers: {
            // Utilise le JWT stocké dans le store d'authentification
            'Authorization': `Bearer ${authStore.token}`, 
          },
        });
        this.genres = response.data.data; // Mettre à jour le state avec les données reçues
      } catch (err) {
        this.error = "Erreur lors de la récupération des genres.";
        console.error(err);
      } finally {
        this.loading = false;
      }
    },

    // ------------------------------------------
    // B. Créer un nouveau genre (CREATE)
    // ------------------------------------------
    async createGenre(name) {
      this.loading = true;
      this.error = null;
      try {
        const authStore = useAuthStore();
        
        const response = await axios.post(API_URL, { name }, {
          headers: {
            'Authorization': `Bearer ${authStore.token}`,
          },
        });
        
        // Ajoute le nouveau genre à la liste existante
        this.genres.push(response.data.data); 
        return true; // Succès
      } catch (err) {
        this.error = "Erreur lors de la création du genre.";
        console.error(err);
        return false; // Échec
      } finally {
        this.loading = false;
      }
    },

    // ------------------------------------------
    // C. Supprimer un genre (DELETE)
    // ------------------------------------------
    async deleteGenre(id) {
      this.loading = true;
      this.error = null;
      try {
        const authStore = useAuthStore();
        
        await axios.delete(`${API_URL}/${id}`, {
          headers: {
            'Authorization': `Bearer ${authStore.token}`,
          },
        });
        
        // Filtre et retire le genre supprimé du state
        this.genres = this.genres.filter(genre => genre.id !== id);
        return true; // Succès
      } catch (err) {
        this.error = "Erreur lors de la suppression du genre.";
        console.error(err);
        return false; // Échec
      } finally {
        this.loading = false;
      }
    },
    
    // NOTE: L'action UPDATE n'est pas incluse pour rester simple. Elle serait similaire à CREATE/DELETE.
  }
});