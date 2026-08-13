// frontend/src/composables/useApi.js - VERSION COMPLETE AVEC UPLOAD
import { useAuthStore } from '@/stores/auth';
import { useRouter } from 'vue-router';

export function useApi() {
  const authStore = useAuthStore();
  const router = useRouter();

  // ✅ CORRECTION: URL relative pour passer par Nginx en production
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

  // Helper pour construire l'URL absolue
  const buildUrl = (endpoint) => {
    if (endpoint.startsWith('http')) {
      return endpoint;
    }
    
    const cleanEndpoint = endpoint.replace(/^\/+/, '');
    let fullUrl = `${API_BASE_URL}/${cleanEndpoint}`;
    fullUrl = fullUrl.replace(/([^:]\/)\/+/g, '$1');
    
    return fullUrl;
  };

  const apiFetch = async (endpoint, options = {}) => {
    const token = authStore.token;
    const fullUrl = buildUrl(endpoint);

    console.log('API Call:', fullUrl);
    console.log('Token utilisé:', token ? `${token.substring(0, 20)}...` : 'AUCUN TOKEN');

    const defaultOptions = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    };

    if (token) {
      defaultOptions.headers['Authorization'] = `Bearer ${token}`;
    }

    try {
      const fetchOptions = {
        ...defaultOptions,
        ...options,
        headers: {
          ...defaultOptions.headers,
          ...options.headers,
        }
      };

      const response = await fetch(fullUrl, fetchOptions);

      console.log('API Response:', response.status, response.statusText, response.ok);

      if (response.status === 401) {
        console.error('Token expiré ou invalide (401), déconnexion');
        authStore.logout();
        if (router) router.push('/login');
        throw new Error('Session expirée. Veuillez vous reconnecter.');
      }

      if (!response.ok) {
        let errorMessage = `Erreur ${response.status}: ${response.statusText}`;
        
        try {
          const errorText = await response.text();
          if (errorText) {
            errorMessage = `Erreur ${response.status}: ${errorText}`;
          }
        } catch (e) {
          // Ignorer
        }
        
        console.error('API Error:', errorMessage);
        throw new Error(errorMessage);
      }

      if (response.status === 204) {
        console.log('API: Réponse 204 (No Content)');
        return null;
      }

      const contentType = response.headers.get('content-type');
      
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text();
        console.log('API: Réponse non-JSON reçue');
        return text || null;
      }

      const text = await response.text();
      
      if (!text || text.trim() === '' || text === 'null') {
        console.warn('API: Réponse vide ou null reçue');
        return null;
      }

      try {
        const data = JSON.parse(text);
        console.log('API Data received:', data);
        return data;
      } catch (error) {
        console.error('Erreur parsing JSON:', error, 'Contenu:', text);
        throw new Error('Réponse invalide du serveur (JSON mal formé)');
      }
    } catch (error) {
      if (error.name === 'TypeError' && error.message.includes('fetch')) {
        const networkError = 'Erreur réseau : impossible de joindre le serveur. Vérifiez votre connexion et que le serveur est lancé sur ' + API_BASE_URL;
        console.error(networkError);
        throw new Error(networkError);
      }
      
      throw error;
    }
  };

  const get = (endpoint, options = {}) => 
    apiFetch(endpoint, { method: 'GET', ...options });

  const post = (endpoint, data = null, options = {}) => {
    const fetchOptions = { method: 'POST', ...options };
    if (data !== null) {
      fetchOptions.body = JSON.stringify(data);
    }
    return apiFetch(endpoint, fetchOptions);
  };

  const put = (endpoint, data = null, options = {}) => {
    const fetchOptions = { method: 'PUT', ...options };
    if (data !== null) {
      fetchOptions.body = JSON.stringify(data);
    }
    return apiFetch(endpoint, fetchOptions);
  };

  const del = (endpoint, options = {}) => 
    apiFetch(endpoint, { method: 'DELETE', ...options });

  // ✅ Fonction upload pour gérer FormData
  const upload = async (endpoint, formData, options = {}) => {
    const token = authStore.token;
    const fullUrl = buildUrl(endpoint);

    console.log('API Upload:', fullUrl);
    console.log('Token utilisé:', token ? `${token.substring(0, 20)}...` : 'AUCUN TOKEN');

    const fetchOptions = {
      method: 'POST',
      headers: {
        // ⚠️ NE PAS définir Content-Type pour FormData
        // Le navigateur le fera automatiquement avec le boundary
        ...(token && { 'Authorization': `Bearer ${token}` }),
        ...options.headers,
      },
      body: formData,
      ...options,
    };

    try {
      const response = await fetch(fullUrl, fetchOptions);

      console.log('API Upload Response:', response.status, response.statusText, response.ok);

      // Gestion du 401
      if (response.status === 401) {
        console.error('Token expiré ou invalide (401), déconnexion');
        authStore.logout();
        if (router) router.push('/login');
        throw new Error('Session expirée. Veuillez vous reconnecter.');
      }

      if (!response.ok) {
        let errorMessage = `Erreur ${response.status}: ${response.statusText}`;
        try {
          const errorText = await response.text();
          if (errorText) {
            errorMessage = `Erreur ${response.status}: ${errorText}`;
          }
        } catch (e) {
          // Ignorer
        }
        console.error('API Upload Error:', errorMessage);
        throw new Error(errorMessage);
      }

      // Parse JSON response
      const data = await response.json();
      console.log('API Upload Data received:', data);
      return data;
    } catch (error) {
      if (error.name === 'TypeError' && error.message.includes('fetch')) {
        const networkError = 'Erreur réseau : impossible de joindre le serveur';
        console.error(networkError);
        throw new Error(networkError);
      }
      throw error;
    }
  };

  return { 
    apiFetch,
    get,
    post,
    put,
    del,
    upload,
    API_BASE_URL
  };
}