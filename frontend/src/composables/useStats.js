// composables/useStats.js
import { ref } from 'vue';
import { useApi } from '@/composables/useApi';

export function useStats() {
  const stats = ref({
    vinyls: 0, artists: 0, genres: 0, formats: 0, countries: 0, labels: 0
  });
  
  const genreDistribution = ref([]);
  const formatDistribution = ref([]);
  const artistDistribution = ref([]);
  const recentDiscs = ref([]);
  const countries = ref([]);
  
  const isLoading = ref(false);
  const error = ref(null);
  
  const { get } = useApi();

  // ✅ CORRECTION: URL relative pour passer par Nginx en production
  const normalizeCoverUrl = (url) => {
    if (!url) return null;
    if (url.startsWith('http://') || url.startsWith('https://')) return url;
    const baseUrl = import.meta.env.VITE_SERVER_BASE_URL || '';
    return baseUrl + (url.startsWith('/') ? url : `/${url}`);
  };

  const fetchStats = async () => {
    isLoading.value = true;
    error.value = null;
    
    try {
      const data = await get('/stats');
      
      stats.value = {
        vinyls: data.vinyls || 0,
        artists: data.artists || 0,
        genres: data.genres || 0,
        formats: data.formats || 0,
        countries: data.countries || 0,
        labels: data.labels || 0
      };
      
      return stats.value;
    } catch (err) {
      console.error('❌ Erreur stats:', err);
      error.value = err.message || 'Impossible de charger les statistiques';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const fetchGenreDistribution = async () => {
    try {
      const discs = await get('/discs');
      
      if (!Array.isArray(discs)) {
        genreDistribution.value = [];
        return;
      }
      
      const genreCounts = {};
      discs.forEach(disc => {
        const genre = disc.genre_name || 'Non classé';
        genreCounts[genre] = (genreCounts[genre] || 0) + 1;
      });
      
      genreDistribution.value = Object.entries(genreCounts)
        .map(([name, count]) => ({ name, count }))
        .sort((a, b) => b.count - a.count);
        
    } catch (err) {
      console.error('❌ Erreur genre distribution:', err);
      genreDistribution.value = [];
    }
  };

  const fetchFormatDistribution = async () => {
    try {
      const discs = await get('/discs');
      
      if (!Array.isArray(discs)) {
        formatDistribution.value = [];
        return;
      }
      
      const formatCounts = {};
      discs.forEach(disc => {
        const format = disc.format_name || 'Non spécifié';
        formatCounts[format] = (formatCounts[format] || 0) + 1;
      });
      
      formatDistribution.value = Object.entries(formatCounts)
        .map(([name, count]) => ({ name, count }))
        .sort((a, b) => b.count - a.count)
        .slice(0, 8);
        
    } catch (err) {
      console.error('❌ Erreur format distribution:', err);
      formatDistribution.value = [];
    }
  };

  const fetchArtistDistribution = async () => {
    try {
      const discs = await get('/discs');
      
      if (!Array.isArray(discs)) {
        artistDistribution.value = [];
        return;
      }
      
      const artistCounts = {};
      discs.forEach(disc => {
        const artist = disc.artist_name || 'Artiste inconnu';
        artistCounts[artist] = (artistCounts[artist] || 0) + 1;
      });
      
      artistDistribution.value = Object.entries(artistCounts)
        .map(([name, count]) => ({ name, count }))
        .sort((a, b) => b.count - a.count);
        
    } catch (err) {
      console.error('❌ Erreur artist distribution:', err);
      artistDistribution.value = [];
    }
  };

  const fetchRecentDiscs = async () => {
    try {
      const discs = await get('/discs');
      
      recentDiscs.value = Array.isArray(discs)
        ? discs
            .sort((a, b) => (b.id || 0) - (a.id || 0))
            .slice(0, 5)
            .map(disc => ({
              ...disc,
              cover_url: normalizeCoverUrl(disc.cover_url)
            }))
        : [];
        
    } catch (err) {
      console.error('❌ Erreur discs:', err);
      recentDiscs.value = [];
    }
  };

  const fetchCountries = async () => {
    try {
      const [countriesData, discsData] = await Promise.all([
        get('/countries'),
        get('/discs')
      ]);
      
      if (!Array.isArray(countriesData) || !Array.isArray(discsData)) {
        countries.value = [];
        return;
      }
      
      const discCountByCountry = {};
      discsData.forEach(disc => {
        const cid = Number(disc.country_id);
        if (cid && !isNaN(cid)) {
          discCountByCountry[cid] = (discCountByCountry[cid] || 0) + 1;
        }
      });
      
      countries.value = countriesData.map(country => ({
        ...country,
        disc_count: discCountByCountry[Number(country.id)] || 0
      }));
      
    } catch (err) {
      console.error('❌ Erreur countries:', err);
      countries.value = [];
    }
  };

  return {
    stats,
    genreDistribution,
    formatDistribution,
    artistDistribution,
    recentDiscs,
    countries,
    isLoading,
    error,
    fetchStats,
    fetchGenreDistribution,
    fetchFormatDistribution,
    fetchArtistDistribution,
    fetchRecentDiscs,
    fetchCountries,
    normalizeCoverUrl
  };
}