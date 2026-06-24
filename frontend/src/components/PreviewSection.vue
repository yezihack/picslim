<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import { api } from '../services/wails';
import type { PreviewPairResult, FileJob, ResultSummary } from '../types';
import ImageCompareSlider from './ImageCompareSlider.vue';

// Props
const props = defineProps<{
  taskId: string;
  files: FileJob[];
  summary: ResultSummary;
  completed: boolean;
}>();

// Emits
const emit = defineEmits<{
  openOutput: [];
  exportReport: [];
  selectFile: [jobId: string];
}>();

// State
const currentIndex = ref(0);
const total = ref(0);
const loading = ref(false);
const error = ref('');
const previewPair = ref<PreviewPairResult | null>(null);

// Fullscreen viewer
const showViewer = ref(false);
const viewerImageSrc = ref('');
const viewerImageType = ref('');

// Computed
const displayIndex = computed(() => currentIndex.value + 1);
const hasFiles = computed(() => props.files.length > 0);
const successFiles = computed(() => props.files.filter(f => f.status === 'SUCCESS' || f.status === 'SKIPPED_NO_GAIN'));
const currentFile = computed(() => {
  if (previewPair.value) {
    return props.files.find(f => f.jobId === previewPair.value?.jobId);
  }
  return null;
});

// Format functions
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

// Load preview
async function loadPreview() {
  if (!props.taskId || !props.completed) return;

  loading.value = true;
  error.value = '';

  try {
    const result = await api.getFirstPreview(props.taskId);
    if (result.code === 0) {
      previewPair.value = result;
      // Get total and index
      total.value = await api.getPreviewTotal();
      currentIndex.value = await api.getPreviewIndex();
    } else {
      error.value = result.message;
    }
  } catch (e: any) {
    error.value = e.message || '加载预览失败';
  } finally {
    loading.value = false;
  }
}

// Navigation
async function goNext() {
  if (!props.taskId || loading.value) return;

  loading.value = true;
  error.value = '';

  try {
    const result = await api.getNextPreview(props.taskId);
    if (result.code === 0) {
      previewPair.value = result;
      currentIndex.value = await api.getPreviewIndex();
    } else {
      error.value = result.message;
    }
  } catch (e: any) {
    error.value = e.message || '加载失败';
  } finally {
    loading.value = false;
  }
}

async function goPrevious() {
  if (!props.taskId || loading.value) return;

  loading.value = true;
  error.value = '';

  try {
    const result = await api.getPreviousPreview(props.taskId);
    if (result.code === 0) {
      previewPair.value = result;
      currentIndex.value = await api.getPreviewIndex();
    } else {
      error.value = result.message;
    }
  } catch (e: any) {
    error.value = e.message || '加载失败';
  } finally {
    loading.value = false;
  }
}

// Watch for task completion
watch(() => props.completed, (newVal) => {
  if (newVal && props.taskId) {
    loadPreview();
  }
});

// Mount
onMounted(() => {
  if (props.completed && props.taskId) {
    loadPreview();
  }
});

// Expose methods
defineExpose({
  loadPreview,
  goNext,
  goPrevious,
});

// Viewer handlers
function viewBefore(src: string) {
  viewerImageSrc.value = src;
  viewerImageType.value = '压缩前';
  showViewer.value = true;
}

function viewAfter(src: string) {
  viewerImageSrc.value = src;
  viewerImageType.value = '压缩后';
  showViewer.value = true;
}

// Close viewer
function closeViewer() {
  showViewer.value = false;
}
</script>

<template>
  <section class="card preview-section">
    <div class="preview-header">
      <h2>压缩前后预览对比</h2>
      <div class="preview-actions">
        <span class="preview-counter" v-if="total > 0">
          {{ displayIndex }} / {{ total }}
        </span>
        <button class="btn btn-ghost" @click="goPrevious" :disabled="loading || total <= 1">
          上一张
        </button>
        <button class="btn btn-ghost" @click="goNext" :disabled="loading || total <= 1">
          下一张
        </button>
        <button class="btn btn-primary-action" @click="emit('openOutput')">
          打开输出目录
        </button>
        <button class="btn btn-secondary" @click="emit('exportReport')">
          导出 CSV 报告
        </button>
      </div>
    </div>

    <!-- Loading state -->
    <div class="preview-loading" v-if="loading">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <!-- Error state -->
    <div class="preview-error" v-else-if="error">
      <div class="error-icon">⚠️</div>
      <p>{{ error }}</p>
    </div>

    <!-- Empty state -->
    <div class="preview-empty" v-else-if="!completed || !previewPair">
      <div class="empty-icon">📷</div>
      <p v-if="!completed">任务完成后可预览压缩效果</p>
      <p v-else>暂无可预览的文件</p>
    </div>

    <!-- Preview content -->
    <div class="preview-slider" v-else>
      <ImageCompareSlider
        v-if="previewPair.sourceBase64 && previewPair.targetBase64"
        :before-src="previewPair.sourceBase64"
        :after-src="previewPair.targetBase64"
        :before-label="`压缩前 (${formatBytes(previewPair.sourceSize)})`"
        :after-label="`压缩后 (${formatBytes(previewPair.targetSize)})`"
        @view-before="viewBefore"
        @view-after="viewAfter"
      />
      <div v-else class="image-placeholder">
        <span>无法加载图片</span>
      </div>

      <!-- Compression stats -->
      <div class="compression-stats" v-if="previewPair.sourceSize > previewPair.targetSize">
        <span class="stat-item">
          <span class="stat-label">节省</span>
          <span class="stat-value saved">
            -{{ ((previewPair.sourceSize - previewPair.targetSize) / previewPair.sourceSize * 100).toFixed(1) }}%
          </span>
          <span class="stat-detail">
            ({{ formatBytes(previewPair.sourceSize - previewPair.targetSize) }})
          </span>
        </span>
      </div>
    </div>

    <!-- File info -->
    <div class="preview-file-info" v-if="previewPair && currentFile">
      <span class="file-name">{{ currentFile.sourcePath.split(/[/\\]/).pop() }}</span>
      <span class="file-format">{{ currentFile.format.toUpperCase() }}</span>
      <span class="file-duration">{{ currentFile.durationMs }}ms</span>
    </div>

    <!-- Fullscreen Image Viewer -->
    <ElDialog
      v-model="showViewer"
      :title="viewerImageType"
      fullscreen
      class="image-viewer-dialog"
      :show-close="true"
      @close="closeViewer"
    >
      <template #default>
        <div class="viewer-container">
          <img :src="viewerImageSrc" alt="Preview" class="viewer-image" />
        </div>
      </template>
    </ElDialog>
  </section>
</template>

<style scoped>
.preview-section {
  margin-top: 20px;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.preview-header h2 {
  margin: 0;
}

.preview-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.preview-counter {
  padding: 4px 12px;
  background: var(--bg-secondary, #f0f0f0);
  border-radius: 4px;
  font-size: 14px;
}

.preview-loading,
.preview-error,
.preview-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #666;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #e0e0e0;
  border-top-color: var(--primary, #409eff);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-icon,
.empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.preview-slider {
  margin-top: 20px;
}

.compression-stats {
  display: flex;
  justify-content: center;
  gap: 24px;
  margin-top: 16px;
  padding: 12px 20px;
  background: var(--bg-secondary, #f5f7fa);
  border-radius: 8px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.stat-label {
  color: #666;
}

.stat-value.saved {
  color: var(--success, #67c23a);
  font-weight: 600;
  font-size: 16px;
}

.stat-detail {
  color: #999;
}

.image-placeholder {
  aspect-ratio: 4/3;
  background: #f5f5f5;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 14px;
}

.preview-file-info {
  display: flex;
  gap: 16px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #eee;
  font-size: 14px;
  color: #666;
}

.file-name {
  color: var(--text-primary, #333);
}

/* Fullscreen Image Viewer */
:deep(.image-viewer-dialog) {
  display: flex !important;
  flex-direction: column;
}

:deep(.image-viewer-dialog .el-dialog) {
  display: flex;
  flex-direction: column;
  max-width: 100vw !important;
  max-height: 100vh !important;
}

:deep(.image-viewer-dialog .el-dialog__header) {
  padding: 12px 20px;
  border-bottom: 1px solid #f0f0f0;
  background: #fafafa;
}

:deep(.image-viewer-dialog .el-dialog__body) {
  flex: 1;
  display: flex;
  padding: 0 !important;
  margin: 0;
  overflow: auto;
  background: #fff;
  align-items: center;
  justify-content: center;
}

.viewer-container {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background: #fff;
  padding: 40px;
  margin: 0;
  overflow: auto;
}

.viewer-image {
  width: auto;
  height: auto;
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: 0;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}
</style>