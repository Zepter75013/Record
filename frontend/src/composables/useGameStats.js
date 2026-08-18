// composables/useGameStats.js — mirroir de useStats.js pour la collection de jeux
import { ref } from 'vue';
import { useApi } from '@/composables/useApi';

export function useGameStats() {
  const stats = ref({
    games: 0, platforms: 0, genres: 0, publishers: 0
  });

  const platformDistribution = ref([]);
  const genreDistribution = ref([]);
  const publisherDistribution = ref([]);
  const recentGames = ref([]);

  const isLoading = ref(false);
  const error = ref(null);

  const { get } = useApi();

  const normalizeCoverUrl = (url) => {
    if (!url) return null;
    if (url.startsWith('http://') || url.startsWith('https://')) return url;
    const baseUrl = import.meta.env.VITE_SERVER_BASE_URL || '';
    return baseUrl + (url.startsWith('/') ? url : `/${url}`);
  };

  const fetchGameStats = async () => {
    isLoading.value = true;
    error.value = null;

    try {
      const games = await get('/games');
      const list = Array.isArray(games) ? games : [];

      const platformIds = new Set();
      const genreIds = new Set();
      const publisherIds = new Set();
      list.forEach((g) => {
        if (g.platform_id) platformIds.add(g.platform_id);
        if (g.genre_id) genreIds.add(g.genre_id);
        if (g.publisher_id) publisherIds.add(g.publisher_id);
      });

      stats.value = {
        games: list.length,
        platforms: platformIds.size,
        genres: genreIds.size,
        publishers: publisherIds.size
      };

      return stats.value;
    } catch (err) {
      console.error('❌ Erreur stats jeux:', err);
      error.value = err.message || 'Impossible de charger les statistiques des jeux';
      throw err;
    } finally {
      isLoading.value = false;
    }
  };

  const fetchPlatformDistribution = async () => {
    try {
      const games = await get('/games');
      if (!Array.isArray(games)) {
        platformDistribution.value = [];
        return;
      }

      const counts = {};
      games.forEach((g) => {
        const platform = g.platform_name || 'Non classé';
        counts[platform] = (counts[platform] || 0) + 1;
      });

      platformDistribution.value = Object.entries(counts)
        .map(([name, count]) => ({ name, count }))
        .sort((a, b) => b.count - a.count);
    } catch (err) {
      console.error('❌ Erreur répartition plateformes:', err);
      platformDistribution.value = [];
    }
  };

  const fetchGameGenreDistribution = async () => {
    try {
      const games = await get('/games');
      if (!Array.isArray(games)) {
        genreDistribution.value = [];
        return;
      }

      const counts = {};
      games.forEach((g) => {
        const genre = g.genre_name || 'Non classé';
        counts[genre] = (counts[genre] || 0) + 1;
      });

      genreDistribution.value = Object.entries(counts)
        .map(([name, count]) => ({ name, count }))
        .sort((a, b) => b.count - a.count);
    } catch (err) {
      console.error('❌ Erreur répartition genres jeux:', err);
      genreDistribution.value = [];
    }
  };

  const fetchPublisherDistribution = async () => {
    try {
      const games = await get('/games');
      if (!Array.isArray(games)) {
        publisherDistribution.value = [];
        return;
      }

      const counts = {};
      games.forEach((g) => {
        const publisher = g.publisher_name || 'Non classé';
        counts[publisher] = (counts[publisher] || 0) + 1;
      });

      publisherDistribution.value = Object.entries(counts)
        .map(([name, count]) => ({ name, count }))
        .sort((a, b) => b.count - a.count)
        .slice(0, 8);
    } catch (err) {
      console.error('❌ Erreur répartition éditeurs:', err);
      publisherDistribution.value = [];
    }
  };

  const fetchRecentGames = async () => {
    try {
      const games = await get('/games');

      recentGames.value = Array.isArray(games)
        ? games
            .sort((a, b) => (b.id || 0) - (a.id || 0))
            .slice(0, 5)
            .map((g) => ({
              ...g,
              cover_url: normalizeCoverUrl(g.cover_url)
            }))
        : [];
    } catch (err) {
      console.error('❌ Erreur jeux récents:', err);
      recentGames.value = [];
    }
  };

  return {
    stats,
    platformDistribution,
    genreDistribution,
    publisherDistribution,
    recentGames,
    isLoading,
    error,
    fetchGameStats,
    fetchPlatformDistribution,
    fetchGameGenreDistribution,
    fetchPublisherDistribution,
    fetchRecentGames,
    normalizeCoverUrl
  };
}
