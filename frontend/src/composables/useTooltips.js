// composables/useTooltips.js
import { ref } from 'vue';

export function useTooltips() {
  // Tooltip pour les camemberts
  const tooltip = ref({
    visible: false,
    x: 0,
    y: 0,
    name: '',
    count: 0,
    percentage: '0'
  });

  const showTooltip = (data) => {
    const { event, name, count, percentage } = data;
    tooltip.value = {
      visible: true,
      x: event.clientX,
      y: event.clientY,
      name,
      count,
      percentage
    };
  };

  const hideTooltip = () => {
    tooltip.value.visible = false;
  };

  const updateTooltipPosition = (event) => {
    if (tooltip.value.visible) {
      tooltip.value.x = event.clientX;
      tooltip.value.y = event.clientY;
    }
  };

  return {
    tooltip,
    showTooltip,
    hideTooltip,
    updateTooltipPosition
  };
}
