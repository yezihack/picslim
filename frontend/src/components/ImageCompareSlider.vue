<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';

const props = defineProps<{
  beforeSrc: string;
  afterSrc: string;
  beforeLabel?: string;
  afterLabel?: string;
}>();

const emit = defineEmits<{
  viewBefore: [src: string];
  viewAfter: [src: string];
}>();

const containerRef = ref<HTMLElement | null>(null);
const position = ref(50);
const isDragging = ref(false);

const clipPath = computed(() => `inset(0 ${100 - position.value}% 0 0)`);

function getPosition(clientX: number): number {
  if (!containerRef.value) return 50;
  const rect = containerRef.value.getBoundingClientRect();
  const x = clientX - rect.left;
  return Math.max(0, Math.min(100, (x / rect.width) * 100));
}

function startDrag(e: MouseEvent | TouchEvent) {
  e.preventDefault();
  isDragging.value = true;
  updatePosition(e);
}

function updatePosition(e: MouseEvent | TouchEvent) {
  if (!isDragging.value) return;
  const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX;
  position.value = getPosition(clientX);
}

function stopDrag() {
  isDragging.value = false;
}

onMounted(() => {
  document.addEventListener('mousemove', updatePosition);
  document.addEventListener('mouseup', stopDrag);
  document.addEventListener('touchmove', updatePosition, { passive: false });
  document.addEventListener('touchend', stopDrag);
});

onUnmounted(() => {
  document.removeEventListener('mousemove', updatePosition);
  document.removeEventListener('mouseup', stopDrag);
  document.removeEventListener('touchmove', updatePosition);
  document.removeEventListener('touchend', stopDrag);
});
</script>

<template>
  <div class="compare-wrapper">
    <!-- Labels -->
    <div class="compare-labels">
      <span class="label label-before">
        <span class="label-text">{{ beforeLabel || '压缩前' }}</span>
      </span>
      <span class="label label-after">
        <span class="label-text">{{ afterLabel || '压缩后' }}</span>
      </span>
    </div>

    <!-- Slider Container -->
    <div
      ref="containerRef"
      class="compare-container"
      :class="{ dragging: isDragging }"
      @mousedown="startDrag"
      @touchstart.prevent="startDrag"
    >
      <!-- After (bottom layer) -->
      <img
        :src="afterSrc"
        alt="Compressed"
        class="compare-image compare-after clickable"
        @click.stop="emit('viewAfter', afterSrc)"
      />

      <!-- Before (top layer, clipped) -->
      <img
        :src="beforeSrc"
        alt="Original"
        class="compare-image compare-before clickable"
        :style="{ clipPath }"
        @click.stop="emit('viewBefore', beforeSrc)"
      />

      <!-- Slider handle -->
      <div class="slider-line" :style="{ left: `${position}%` }">
        <div class="slider-knob">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
            <path d="M8 5L3 12L8 19M16 5L21 12L16 19" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.compare-wrapper {
  width: 100%;
}

.compare-labels {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  padding: 0 4px;
}

.label {
  font-size: 13px;
  font-weight: 500;
}

.label-before {
  color: var(--primary, #409eff);
}

.label-after {
  color: var(--success, #67c23a);
}

.compare-container {
  position: relative;
  width: 100%;
  aspect-ratio: 4 / 3;
  min-height: 300px;
  max-height: 600px;
  border-radius: 12px;
  overflow: hidden;
  cursor: ew-resize;
  background: #f5f5f5;
  user-select: none;
}

.compare-container.dragging {
  cursor: grabbing;
}

.compare-image {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  pointer-events: none;
}

.compare-image.clickable {
  cursor: zoom-in;
  pointer-events: auto;
}

.compare-container:not(.dragging) .compare-image.clickable:hover {
  opacity: 0.9;
}

.compare-after {
  z-index: 1;
}

.compare-before {
  z-index: 2;
}

.slider-line {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 3px;
  background: white;
  z-index: 10;
  transform: translateX(-50%);
  box-shadow: 0 0 8px rgba(0, 0, 0, 0.3);
}

.slider-knob {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 44px;
  height: 44px;
  background: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.25);
  color: #666;
}

.slider-knob svg {
  margin: 0 -2px;
}

.compare-container:hover .slider-knob {
  background: var(--primary, #409eff);
  color: white;
}
</style>
