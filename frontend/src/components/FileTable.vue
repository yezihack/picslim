<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { api } from '../services/wails';
import type { FileJob, FileJobPage, FileJobStatus } from '../types';

// Props
const props = defineProps<{
  taskId: string;
  completed: boolean;
}>();

// Emits
const emit = defineEmits<{
  selectFile: [jobId: string];
}>();

// State
const files = ref<FileJob[]>([]);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(10);
const loading = ref(false);
const filterStatus = ref<FileJobStatus | ''>('');
const searchText = ref('');

// Computed
const totalPages = computed(() => Math.ceil(total.value / pageSize.value));

const filteredFiles = computed(() => {
  let result = files.value;

  if (filterStatus.value) {
    result = result.filter(f => f.status === filterStatus.value);
  }

  if (searchText.value) {
    const search = searchText.value.toLowerCase();
    result = result.filter(f =>
      f.sourcePath.toLowerCase().includes(search)
    );
  }

  return result;
});

const statusOptions = [
  { label: '全部', value: '' },
  { label: '成功', value: 'SUCCESS' as FileJobStatus },
  { label: '失败', value: 'FAILED' as FileJobStatus },
  { label: '无收益', value: 'SKIPPED_NO_GAIN' as FileJobStatus },
];

// Format functions
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

function formatPercent(before: number, after: number): string {
  if (before === 0) return '-';
  const percent = ((before - after) / before) * 100;
  if (percent < 0) return '+0%';
  return `-${percent.toFixed(1)}%`;
}

function getStatusText(status: FileJobStatus): string {
  switch (status) {
    case 'SUCCESS': return '成功';
    case 'FAILED': return '失败';
    case 'SKIPPED_NO_GAIN': return '无收益';
    case 'PENDING': return '等待中';
    case 'RUNNING': return '处理中';
    default: return status;
  }
}

function getStatusClass(status: FileJobStatus): string {
  switch (status) {
    case 'SUCCESS': return 'status-success';
    case 'FAILED': return 'status-failed';
    case 'SKIPPED_NO_GAIN': return 'status-skipped';
    default: return '';
  }
}

// Load files
async function loadFiles() {
  if (!props.taskId) return;

  loading.value = true;

  try {
    const result: FileJobPage = await api.listTaskFiles(props.taskId, currentPage.value, pageSize.value);
    if (result.code === 0) {
      files.value = result.items || [];
      total.value = result.total;
    }
  } catch (e) {
    console.error('Failed to load files:', e);
  } finally {
    loading.value = false;
  }
}

// Pagination
function goToPage(page: number) {
  if (page < 1 || page > totalPages.value) return;
  currentPage.value = page;
  loadFiles();
}

function prevPage() {
  goToPage(currentPage.value - 1);
}

function nextPage() {
  goToPage(currentPage.value + 1);
}

// Select file
function selectFile(job: FileJob) {
  emit('selectFile', job.jobId);
}

// Watch
watch(() => props.completed, (newVal) => {
  if (newVal && props.taskId) {
    loadFiles();
  }
});

watch([currentPage, pageSize], () => {
  if (props.taskId) {
    loadFiles();
  }
});

// Mount
onMounted(() => {
  if (props.completed && props.taskId) {
    loadFiles();
  }
});
</script>

<template>
  <section class="card table-section">
    <div class="table-header">
      <h2>文件明细</h2>
      <div class="table-controls">
        <input
          v-model="searchText"
          type="text"
          placeholder="搜索文件名..."
          class="search-input"
        />
        <select v-model="filterStatus" class="filter-select">
          <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>
    </div>

    <!-- Loading -->
    <div class="table-loading" v-if="loading">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>

    <!-- Empty -->
    <div class="table-empty" v-else-if="filteredFiles.length === 0">
      <p>暂无数据</p>
    </div>

    <!-- Table -->
    <div class="table-container" v-else>
      <table>
        <thead>
          <tr>
            <th class="col-name">文件名</th>
            <th class="col-format">格式</th>
            <th class="col-size">压缩前</th>
            <th class="col-size">压缩后</th>
            <th class="col-ratio">比例</th>
            <th class="col-status">状态</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="file in filteredFiles"
            :key="file.jobId"
            @click="selectFile(file)"
            class="file-row"
          >
            <td class="col-name" :title="file.sourcePath">
              {{ file.sourcePath.split(/[/\\]/).pop() }}
            </td>
            <td class="col-format">{{ file.format.toUpperCase() }}</td>
            <td class="col-size">{{ formatBytes(file.bytesBefore) }}</td>
            <td class="col-size">
              {{ file.bytesAfter ? formatBytes(file.bytesAfter) : '-' }}
            </td>
            <td class="col-ratio">
              {{ file.bytesAfter ? formatPercent(file.bytesBefore, file.bytesAfter) : '-' }}
            </td>
            <td class="col-status" :class="getStatusClass(file.status)">
              {{ getStatusText(file.status) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div class="pagination" v-if="totalPages > 1">
      <button
        class="btn btn-ghost"
        @click="prevPage"
        :disabled="currentPage === 1"
      >
        上一页
      </button>
      <span class="page-info">
        第 {{ currentPage }} / {{ totalPages }} 页，共 {{ total }} 条
      </span>
      <button
        class="btn btn-ghost"
        @click="nextPage"
        :disabled="currentPage === totalPages"
      >
        下一页
      </button>
    </div>
  </section>
</template>

<style scoped>
.table-section {
  margin-top: 20px;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}

.table-header h2 {
  margin: 0;
}

.table-controls {
  display: flex;
  gap: 12px;
}

.search-input {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  width: 200px;
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  background: white;
}

.table-loading,
.table-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #666;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #e0e0e0;
  border-top-color: var(--primary, #409eff);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.table-container {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #eee;
}

th {
  font-weight: 600;
  color: #666;
  font-size: 13px;
}

.file-row {
  cursor: pointer;
  transition: background 0.2s;
}

.file-row:hover {
  background: #f5f7fa;
}

.col-name {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-format {
  width: 80px;
}

.col-size {
  width: 100px;
}

.col-ratio {
  width: 80px;
}

.col-status {
  width: 100px;
}

.status-success {
  color: var(--success, #67c23a);
  font-weight: 500;
}

.status-failed {
  color: var(--danger, #f56c6c);
  font-weight: 500;
}

.status-skipped {
  color: var(--warning, #e6a23c);
  font-weight: 500;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #eee;
}

.page-info {
  font-size: 14px;
  color: #666;
}

@media (max-width: 768px) {
  .table-controls {
    width: 100%;
  }

  .search-input {
    flex: 1;
  }

  .col-name {
    max-width: 150px;
  }
}
</style>