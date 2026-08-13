// composables/useCounter.js
import { ref } from 'vue';

export function useCounter(duration = 1200) {
  const current = ref(0);
  let startTime = null;
  let animationFrame = null;

  const animate = (timestamp, target, startValue) => {
    if (!startTime) startTime = timestamp;
    const progress = Math.min((timestamp - startTime) / duration, 1);
    current.value = Math.floor(startValue + (target - startValue) * progress);
    
    if (progress < 1) {
      animationFrame = requestAnimationFrame((ts) => animate(ts, target, startValue));
    } else {
      current.value = target;
    }
  };

  const start = (newTarget) => {
    if (animationFrame) {
      cancelAnimationFrame(animationFrame);
    }
    
    startTime = null;
    const startValue = current.value;
    animationFrame = requestAnimationFrame((timestamp) => animate(timestamp, newTarget, startValue));
  };

  return {
    current,
    start
  };
}