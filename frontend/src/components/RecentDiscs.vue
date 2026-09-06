<!-- components/RecentDiscs.vue - VERSION RESPONSIVE MOBILE OPTIMISÉE -->
<script setup>
import { formatDate as formatDateShared } from '@/utils/format';

const props = defineProps({
  discs: {
    type: Array,
    default: () => []
  }
});

const emit = defineEmits(['image-error', 'edit-disc']);

const formatDate = (dateStr) => {
  if (!dateStr) return '—';
  return formatDateShared(dateStr, { withTime: true }) || '—';
};

// Format court pour mobile
const formatDateShort = (dateStr) => {
  if (!dateStr) return '—';
  return formatDateShared(dateStr, { short: true }) || '—';
};

const handleImageError = (e, disc) => {
  emit('image-error', { event: e, disc });
};
</script>

<template>
  <div class="widget recent-activity-widget">
    <h3 class="widget-title">🎧 Derniers ajouts</h3>
    <div v-if="discs.length === 0" class="placeholder-text">
      Aucun disque récent
    </div>
    <div v-else class="recent-activity-list">
      <div
        v-for="(disc, index) in discs"
        :key="disc.id"
        class="activity-item"
        :class="{ 'even': index % 2 === 1 }"
        title="Double-cliquer pour modifier"
        @dblclick="emit('edit-disc', disc)"
      >
        <div class="cover-container">
          <img
            v-if="disc.cover_url"
            :src="disc.cover_url"
            :alt="`Couverture: ${disc.title || 'Sans titre'}`"
            class="cover"
            loading="lazy"
            @error="(e) => handleImageError(e, disc)"
          />
          <div 
            class="cover-fallback" 
            :style="{ display: disc.cover_url ? 'none' : 'flex' }"
          >
            <span class="disc-icon">💿</span>
          </div>
        </div>
        <div class="activity-content">
          <div class="artist-title">
            <span class="artist">{{ disc.artist_name || 'Artiste inconnu' }}</span>
            <span class="separator"> • </span>
            <span class="title">{{ disc.title || 'Sans titre' }}</span>
          </div>
          <div class="metadata">
            <span class="date date-full">
              📅 {{ formatDate(disc.created_at || disc.updated_at) }}
            </span>
            <span class="date date-short">
              📅 {{ formatDateShort(disc.created_at || disc.updated_at) }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.widget {
  background: var(--panel-bg);
  padding: 16px;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
  border: 1px solid var(--line-soft);
  transition: all 0.3s ease;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
}

.widget:hover {
  transform: translateY(-2px);
}

.widget-title {
  color: var(--text);
  margin: 0 0 12px 0;
  font-weight: 700;
  font-size: 0.95em;
  border-bottom: 1px solid var(--line-soft);
  padding-bottom: 8px;
}

.placeholder-text {
  color: var(--text-dim);
  font-style: italic;
  text-align: center;
  margin: 30px 0;
  font-size: 0.95em;
}

.recent-activity-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px;
  border-radius: 10px;
  background: rgba(var(--tint-rgb), 0.03);
  border: 1px solid var(--line-soft);
  transition: all 0.3s ease;
  position: relative;
  cursor: pointer;
  overflow: hidden;
}

.activity-item::before {
  content: '';
  position: absolute;
  top: 0; left: 0; bottom: 0;
  width: 4px;
  background: linear-gradient(to bottom, var(--accent), var(--accent-blue));
  opacity: 0.7;
}

.activity-item.even {
  background: rgba(var(--tint-rgb), 0.015);
}

.activity-item:hover {
  transform: translateX(4px);
  border-color: var(--accent-soft);
}

.cover-container {
  flex-shrink: 0;
  width: 45px;
  height: 45px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 3px 8px rgba(0,0,0,0.12);
  background: rgba(var(--tint-rgb), 0.06);
}

.cover {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: transform 0.3s ease;
}

.activity-item:hover .cover {
  transform: scale(1.05);
}

.cover-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
}

.disc-icon {
  font-size: 1.4em;
  color: var(--accent);
}

.activity-content {
  flex-grow: 1;
  min-width: 0;
}

.artist-title {
  font-size: 0.85em;
  margin-bottom: 3px;
  line-height: 1.3;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.artist {
  font-weight: 600;
  color: var(--text);
}

.title {
  font-style: italic;
  color: var(--text-soft);
  font-weight: 500;
}

.separator {
  color: var(--text-dim);
  margin: 0 4px;
}

.metadata {
  display: flex;
  justify-content: space-between;
  font-size: 0.78em;
  color: var(--text-dim);
}

.date {
  display: flex;
  align-items: center;
  gap: 4px;
}

.date-short {
  display: none;
}

/* ——— RESPONSIVE ——— */
@media (max-width: 992px) {
  .widget {
    padding: 14px;
  }
  
  .recent-activity-list {
    max-height: 300px;
    overflow-y: auto;
  }
}

@media (max-width: 767px) {
  .widget {
    padding: 12px;
  }
  
  .widget-title {
    font-size: 0.9em;
    margin-bottom: 10px;
    padding-bottom: 6px;
  }
  
  .recent-activity-list {
    gap: 6px;
    max-height: none;
  }
  
  .activity-item {
    padding: 8px;
    gap: 10px;
  }
  
  .cover-container {
    width: 40px;
    height: 40px;
  }
  
  .disc-icon {
    font-size: 1.2em;
  }
  
  .artist-title {
    font-size: 0.8em;
  }
  
  .date-full {
    display: none;
  }
  
  .date-short {
    display: flex;
  }
}

@media (max-width: 480px) {
  .widget {
    padding: 10px;
  }
  
  .activity-item {
    padding: 6px 8px;
    gap: 8px;
  }
  
  .activity-item::before {
    width: 3px;
  }
  
  .cover-container {
    width: 36px;
    height: 36px;
  }
  
  .disc-icon {
    font-size: 1em;
  }
  
  .artist-title {
    font-size: 0.75em;
    margin-bottom: 2px;
  }
  
  .separator {
    margin: 0 2px;
  }
  
  .metadata {
    font-size: 0.7em;
  }
}

@media (max-width: 360px) {
  .cover-container {
    width: 32px;
    height: 32px;
  }
  
  .activity-item {
    gap: 6px;
    padding: 5px 6px;
  }
  
  .artist-title {
    font-size: 0.7em;
  }
}
</style>
