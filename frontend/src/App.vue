<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { ElMessage, ElDialog, ElSlider } from "element-plus";
import { api, events, isWails } from "./services/wails";
import type { ScanResult, LogEntry, CompletedJob } from "./types";

// Settings
const showSettings = ref(false);
const settings = ref({
  concurrency: 10,
  autoOpenOutput: true,
});

// Load settings from localStorage
function loadSettings() {
  const saved = localStorage.getItem('compressor-settings');
  if (saved) {
    try {
      const parsed = JSON.parse(saved);
      settings.value = { ...settings.value, ...parsed };
    } catch (e) {
      console.error('Failed to load settings:', e);
    }
  }
}

// Save settings to localStorage
function saveSettings() {
  localStorage.setItem('compressor-settings', JSON.stringify(settings.value));
  ElMessage.success('设置已保存');
  showSettings.value = false;
}

// State
const inputPaths = ref<string[]>([]);
const outputDir = ref("");
const outputMode = ref<"AUTO" | "MANUAL">("AUTO");
const preset = ref<"HIGH_QUALITY" | "BALANCED" | "HIGH_COMPRESSION">("HIGH_QUALITY");

// Task state
const status = ref<string>("idle"); // idle / running / paused / cancelled / completed
const total = ref<number>(0);
const done = ref<number>(0);
const failed = ref<number>(0);

const isRunning = computed(() => status.value === "running");
const isPaused = computed(() => status.value === "paused");
const isCompleted = computed(() => status.value === "completed");

// Progress
const progressPercent = computed(() => {
  if (total.value === 0) return 0;
  return Math.round((done.value + failed.value) / total.value * 100);
});

// Logs
const logs = ref<LogEntry[]>([]);

// Completed jobs for table
const completedJobs = ref<CompletedJob[]>([]);

// Preview state
const showPreview = ref(false);
const previewIndex = ref(0);
const previewSource = ref("");
const previewTarget = ref("");
const previewFileName = ref("");
const previewSourceSize = ref(0);
const previewTargetSize = ref(0);
const previewRatio = ref("");

// Image viewer state
const showImageViewer = ref(false);
const viewerImageSrc = ref("");
const viewerImageTitle = ref("");

// Open image viewer
function openImageViewer(src: string, title: string) {
  viewerImageSrc.value = src;
  viewerImageTitle.value = title;
  showImageViewer.value = true;
}

// Scan result
const scanResult = ref<ScanResult | null>(null);

// Concurrency (from settings)
const concurrency = computed(() => settings.value.concurrency);

// Estimated savings based on preset
const estimatedSavings = computed(() => {
  if (!scanResult.value || scanResult.value.totalBytes === 0) {
    return { bytes: 0, percent: 0 };
  }

  let ratio = 0.50;
  switch (preset.value) {
    case 'HIGH_QUALITY':
      ratio = 0.45;
      break;
    case 'BALANCED':
      ratio = 0.55;
      break;
    case 'HIGH_COMPRESSION':
      ratio = 0.65;
      break;
  }

  const totalBytes = scanResult.value.totalBytes;
  return {
    bytes: Math.round(totalBytes * ratio),
    percent: Math.round(ratio * 100)
  };
});

// Format functions
function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

// Actions
async function selectInput() {
  scanResult.value = null;

  const result = await api.selectInputDir();
  if (result.code === 0 && result.path) {
    inputPaths.value = [result.path];

    if (outputMode.value === "AUTO") {
      outputDir.value = result.path + "-compressed";
    }

    await scanInputs();
  }
}

async function selectOutput() {
  const result = await api.selectOutputDir();
  if (result.code === 0 && result.path) {
    outputDir.value = result.path;
    outputMode.value = "MANUAL";
  }
}

async function scanInputs() {
  if (inputPaths.value.length === 0) return;
  scanResult.value = null;
  const result = await api.scanInputPaths(inputPaths.value);
  if (result.code === 0) {
    scanResult.value = result;
  }
}

function setPreset(p: "HIGH_QUALITY" | "BALANCED" | "HIGH_COMPRESSION") {
  preset.value = p;
}

async function startCompression() {
  if (inputPaths.value.length === 0) {
    ElMessage.warning("请先选择输入目录");
    return;
  }

  console.log("Starting compression...", {
    inputDir: inputPaths.value[0],
    outputDir: outputDir.value,
    preset: preset.value,
    concurrency: concurrency.value
  });

  // Reset state
  logs.value = [];
  done.value = 0;
  failed.value = 0;
  total.value = 0;

  try {
    console.log("Calling api.startCompress...");
    const result = await api.startCompress(
      inputPaths.value[0],
      outputDir.value,
      preset.value,
      concurrency.value
    );

    console.log("Result:", result);

    if (result.code !== 0) {
      ElMessage.error(`启动失败: ${result.message || "未知错误"}`);
      return;
    }

    status.value = "running";
    ElMessage.success("压缩任务已开始");
  } catch (err: any) {
    console.error("startCompression error:", err);
    ElMessage.error(`操作失败: ${err?.message || "未知错误"}`);
  }
}

async function openOutput() {
  await api.openOutputDir();
}

async function exportReport() {
  await api.exportReport("csv");
}

// Preview functions
async function openPreview(index: number) {
  const result = await api.getPreview(index);
  if (result.code !== 0) {
    ElMessage.error(result.message || "无法加载预览");
    return;
  }
  previewIndex.value = index;
  previewSource.value = result.sourceBase64;
  previewTarget.value = result.targetBase64;
  previewFileName.value = result.fileName;
  previewSourceSize.value = result.sourceSize;
  previewTargetSize.value = result.targetSize;
  previewRatio.value = result.ratio;
  showPreview.value = true;
}

async function prevPreview() {
  if (previewIndex.value > 0) {
    await openPreview(previewIndex.value - 1);
  }
}

async function nextPreview() {
  if (previewIndex.value < completedJobs.value.length - 1) {
    await openPreview(previewIndex.value + 1);
  }
}

// Random preview - pick a random completed image
async function randomPreview() {
  const doneJobs = completedJobs.value.filter(job => job.status === 'done');
  if (doneJobs.length === 0) {
    ElMessage.warning('没有成功的压缩图片可预览');
    return;
  }
  const randomJob = doneJobs[Math.floor(Math.random() * doneJobs.length)];
  await openPreview(randomJob.index - 1);
}

// Load completed jobs
async function loadCompletedJobs() {
  const result = await api.getCompletedJobs();
  if (result.code === 0) {
    completedJobs.value = result.jobs || [];
  }
}

// Event handlers
function handleInit(data: any) {
  total.value = data.total;
  done.value = 0;
  failed.value = 0;
  logs.value = [];
  status.value = "running";
}

function handleLog(entry: LogEntry) {
  logs.value.unshift(entry);
  if (logs.value.length > 100) {
    logs.value = logs.value.slice(0, 100);
  }
}

function handleProgress(data: any) {
  total.value = data.total;
  done.value = data.done;
  failed.value = data.failed;
}

function handleCompleted(data: any) {
  status.value = "completed";
  done.value = data.done;
  failed.value = data.failed;
  ElMessage.success(`压缩完成！成功 ${data.done} 张，失败 ${data.failed} 张`);
  // Load completed jobs for table and auto show first preview
  loadCompletedJobs().then(() => {
    // Auto show first successful preview
    const firstSuccessJob = completedJobs.value.find(job => job.status === 'done');
    if (firstSuccessJob) {
      openPreview(firstSuccessJob.index - 1);
    }
  });
  // Auto open output directory if enabled
  if (settings.value.autoOpenOutput) {
    openOutput();
  }
}

function handleCancelled() {
  status.value = "idle";
}

function handlePaused() {
  status.value = "paused";
}

function handleResumed() {
  status.value = "running";
}

// Lifecycle
onMounted(() => {
  loadSettings();
  if (isWails()) {
    events.onInit(handleInit);
    events.onLog(handleLog);
    events.onProgress(handleProgress);
    events.onCompleted(handleCompleted);
    events.onCancelled(handleCancelled);
    events.onPaused(handlePaused);
    events.onResumed(handleResumed);
  }
});

onUnmounted(() => {
  events.offAll();
});
</script>

<template>
  <div class="app">
    <header class="topbar">
      <h1>智能图片压缩器</h1>
      <div class="topbar-actions">
        <button class="btn btn-secondary" @click="showSettings = true">设置</button>
      </div>
    </header>

    <main class="content">
      <!-- Left Panel: Input/Output Config -->
      <section class="card left-panel">
        <h2>输入与输出</h2>

        <div class="input-group">
          <label>输入目录 / 文件</label>
          <div class="row">
            <input :value="inputPaths[0] || '请选择输入目录'" readonly />
            <button class="btn btn-secondary" @click="selectInput">选择输入</button>
          </div>
          <p class="hint" v-if="scanResult">
            已扫描 {{ scanResult.totalFiles }} 张图片，共 {{ formatBytes(scanResult.totalBytes) }}
            <span v-if="scanResult.filteredFiles.length > 0">
              ，过滤 {{ scanResult.filteredFiles.length }} 个不支持的文件
            </span>
          </p>
        </div>

        <div class="input-group">
          <label>输出目录（默认自动创建，可手动修改）</label>
          <div class="row">
            <input :value="outputDir || '将自动创建'" readonly />
            <button class="btn btn-secondary" @click="selectOutput">选择输出</button>
          </div>
        </div>

        <div class="input-group">
          <label>压缩预设</label>
          <div class="preset-row">
            <button
              class="btn"
              :class="preset === 'HIGH_QUALITY' ? 'btn-primary' : 'btn-secondary'"
              @click="setPreset('HIGH_QUALITY')"
            >
              高画质（默认）
            </button>
            <button
              class="btn"
              :class="preset === 'BALANCED' ? 'btn-primary' : 'btn-secondary'"
              @click="setPreset('BALANCED')"
            >
              均衡
            </button>
            <button
              class="btn"
              :class="preset === 'HIGH_COMPRESSION' ? 'btn-primary' : 'btn-secondary'"
              @click="setPreset('HIGH_COMPRESSION')"
            >
              高压缩
            </button>
          </div>
        </div>

        
        <!-- Action Bar -->
        <div class="start-wrap action-bar">
          <div class="action-status">
            <p class="action-status-main">
              {{ scanResult ? `已选择 ${scanResult.totalFiles} 张图片` : "请选择输入目录" }}
            </p>
            <p class="action-status-sub" v-if="scanResult && scanResult.totalFiles > 0">
              预计节省 {{ formatBytes(estimatedSavings.bytes) }}（约 -{{ estimatedSavings.percent }}%）
            </p>
          </div>
          <button class="btn btn-hero" @click="startCompression">
            开始压缩
          </button>
        </div>
      </section>

      <!-- Right Panel: Progress -->
      <section class="card right-panel">
        <div class="panel-head">
          <h2>任务进度</h2>
          <div class="status-chips">
            <span class="chip" :class="{
              'chip-running': isRunning,
              'chip-paused': isPaused,
              'chip-completed': isCompleted
            }">
              {{ status === 'idle' ? '待开始' : status === 'running' ? '进行中' : status === 'paused' ? '已暂停' : status === 'completed' ? '已完成' : status }}
            </span>
          </div>
        </div>

        <!-- Progress bar -->
        <div class="progress-wrap">
          <div class="progress-label">
            <span>总进度 {{ progressPercent }}%</span>
            <span>{{ done + failed }} / {{ total }}</span>
          </div>
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: `${progressPercent}%` }"></div>
          </div>
        </div>

        <!-- Stats -->
        <div class="stats-grid">
          <div class="stat-box">
            <p>总文件</p>
            <strong>{{ total }}</strong>
          </div>
          <div class="stat-box">
            <p>成功</p>
            <strong>{{ done }}</strong>
          </div>
          <div class="stat-box">
            <p>失败</p>
            <strong>{{ failed }}</strong>
          </div>
        </div>

        <!-- Completed actions -->
        <div v-if="isCompleted" class="completed-actions">
          <button class="btn btn-secondary" @click="openOutput">打开输出目录</button>
          <button class="btn btn-secondary" @click="exportReport">导出报告</button>
        </div>

        <!-- Log list -->
        <div class="log-section" v-if="logs.length > 0">
          <h3>处理日志</h3>
          <div class="log-list">
            <div v-for="log in logs" :key="log.index" class="log-entry" :class="log.status">
              <span class="log-index">#{{ log.index }}</span>
              <span class="log-name">{{ log.fileName }}</span>
              <span class="log-size" v-if="log.status === 'done'">
                {{ formatBytes(log.oldSize) }} → {{ formatBytes(log.newSize) }}
              </span>
              <span class="log-ratio" v-if="log.status === 'done'">{{ log.ratio }}</span>
              <span class="log-error" v-if="log.status === 'error'">{{ log.message }}</span>
              <span class="log-time">{{ log.time }}</span>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- Preview Section -->
    <section class="card preview-section" v-if="isCompleted && completedJobs.length > 0">
      <div class="preview-header">
        <h2>压缩前后预览对比</h2>
        <div class="preview-nav">
          <button class="btn btn-secondary" @click="randomPreview">随机抽查</button>
          <button class="btn btn-secondary" :disabled="previewIndex === 0" @click="prevPreview">上一张</button>
          <span class="preview-counter">{{ previewIndex + 1 }} / {{ completedJobs.length }}</span>
          <button class="btn btn-secondary" :disabled="previewIndex >= completedJobs.length - 1" @click="nextPreview">下一张</button>
        </div>
      </div>
      <div class="preview-grid" v-if="previewSource">
        <figure>
          <figcaption>压缩前（{{ formatBytes(previewSourceSize) }}）</figcaption>
          <img :src="previewSource" alt="压缩前" @click="openImageViewer(previewSource, `${previewFileName} - 压缩前`)" class="clickable-img" />
        </figure>
        <figure>
          <figcaption>压缩后（{{ formatBytes(previewTargetSize) }}）- {{ previewRatio }}</figcaption>
          <img :src="previewTarget" alt="压缩后" @click="openImageViewer(previewTarget, `${previewFileName} - 压缩后`)" class="clickable-img" />
        </figure>
      </div>
      <div class="preview-placeholder" v-else>
        <p>点击下方表格中的图片查看预览</p>
      </div>
    </section>

    <!-- File Detail Table -->
    <section class="card table-section" v-if="isCompleted && completedJobs.length > 0">
      <h2>文件明细</h2>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>#</th>
              <th>文件名</th>
              <th>压缩前</th>
              <th>压缩后</th>
              <th>比例</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="job in completedJobs" :key="job.index" :class="job.status">
              <td>{{ job.index }}</td>
              <td class="file-name">{{ job.fileName }}</td>
              <td>{{ formatBytes(job.oldSize) }}</td>
              <td>{{ job.status === 'done' ? formatBytes(job.newSize) : '-' }}</td>
              <td :class="job.status === 'done' ? 'ratio-good' : ''">{{ job.status === 'done' ? job.ratio : '-' }}</td>
              <td>
                <span class="status-badge" :class="job.status">{{ job.status === 'done' ? '成功' : '失败' }}</span>
              </td>
              <td>
                <button class="btn btn-small" @click="openPreview(job.index - 1)" v-if="job.status === 'done'">预览</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- Settings Dialog -->
    <ElDialog
      v-model="showSettings"
      title="设置"
      width="360px"
      :close-on-click-modal="false"
      class="settings-dialog"
    >
      <div class="settings-content">
        <div class="setting-item">
          <label>并发数量</label>
          <div class="setting-value">{{ settings.concurrency }}</div>
          <ElSlider
            v-model="settings.concurrency"
            :min="1"
            :max="32"
            :marks="{ 1: '1', 16: '16', 32: '32' }"
          />
          <p class="setting-hint">同时处理的图片数量，建议根据 CPU 核心数调整</p>
        </div>

        <div class="setting-item setting-item-left">
          <div class="setting-row">
            <label>完成后自动打开输出目录</label>
            <label class="switch">
              <input type="checkbox" v-model="settings.autoOpenOutput" />
              <span class="slider"></span>
            </label>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <button class="btn btn-secondary" @click="showSettings = false">取消</button>
          <button class="btn btn-primary" @click="saveSettings">保存</button>
        </div>
      </template>
    </ElDialog>

    <!-- Image Viewer Dialog -->
    <ElDialog
      v-model="showImageViewer"
      :title="viewerImageTitle"
      fullscreen
      :close-on-click-modal="true"
      class="image-viewer-dialog"
    >
      <div class="image-viewer-content">
        <img :src="viewerImageSrc" alt="预览图片" class="viewer-image" />
      </div>
    </ElDialog>
  </div>
</template>

<style>
@import './styles/app.css';

.hint {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.panel-head h2 {
  margin: 0;
}

.status-chips {
  display: flex;
  gap: 8px;
}

.chip {
  padding: 4px 12px;
  background: var(--bg-secondary, #f0f0f0);
  border-radius: 12px;
  font-size: 12px;
}

.chip-running {
  background: var(--primary-light, #ecf5ff);
  color: var(--primary, #409eff);
}

.chip-paused {
  background: var(--warning-light, #fdf6ec);
  color: var(--warning, #e6a23c);
}

.chip-completed {
  background: var(--success-light, #f0f9eb);
  color: var(--success, #67c23a);
}

.progress-wrap {
  margin: 16px 0;
}

.progress-label {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  margin-bottom: 8px;
}

.progress-bar {
  height: 12px;
  background: #e0e0e0;
  border-radius: 6px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #409eff, #67c23a);
  border-radius: 6px;
  transition: width 0.3s ease;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin: 16px 0;
}

.stat-box {
  text-align: center;
  padding: 12px;
  background: var(--bg-secondary, #f5f7fa);
  border-radius: 8px;
}

.stat-box p {
  font-size: 12px;
  color: #666;
  margin: 0 0 4px;
}

.stat-box strong {
  font-size: 18px;
}

.completed-actions {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}

.log-section {
  margin-top: 20px;
  border-top: 1px solid #eee;
  padding-top: 16px;
}

.log-section h3 {
  font-size: 14px;
  margin: 0 0 12px;
}

.log-list {
  max-height: 300px;
  overflow-y: auto;
}

.log-entry {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
  font-size: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.log-entry.error {
  color: #f56c6c;
}

.log-index {
  color: #999;
  min-width: 40px;
}

.log-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.log-size {
  color: #666;
}

.log-ratio {
  color: #67c23a;
  font-weight: 500;
}

.log-error {
  color: #f56c6c;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.log-time {
  color: #999;
  min-width: 60px;
  text-align: right;
}

/* Preview Section */
.preview-section {
  margin-top: 16px;
  width: 100%;
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

.preview-nav {
  display: flex;
  align-items: center;
  gap: 12px;
}

.preview-counter {
  font-size: 13px;
  color: #666;
  min-width: 60px;
  text-align: center;
}

.preview-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-top: 16px;
}

.preview-grid figure {
  margin: 0;
  text-align: center;
}

.preview-grid figcaption {
  font-size: 13px;
  color: #666;
  margin-bottom: 8px;
}

.preview-grid img {
  max-width: 100%;
  max-height: 500px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  object-fit: contain;
}

.clickable-img {
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.clickable-img:hover {
  transform: scale(1.02);
  box-shadow: 0 4px 16px rgba(0,0,0,0.15);
}

/* Image Viewer Dialog */
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

.image-viewer-dialog .image-viewer-content {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  background: #fff;
  padding: 40px;
  margin: 0;
  overflow: auto;
}

.image-viewer-dialog .viewer-image {
  width: auto;
  height: auto;
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: 0;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.preview-placeholder {
  text-align: center;
  padding: 40px;
  color: #999;
}

/* Table Section */
.table-section {
  margin-top: 16px;
  width: 100%;
}

.table-section h2 {
  margin: 0 0 16px;
}

.table-wrap {
  overflow-x: auto;
}

.table-section table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.table-section th,
.table-section td {
  padding: 10px 12px;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.table-section th {
  background: #f5f7fa;
  font-weight: 500;
  color: #333;
  white-space: nowrap;
}

.table-section tr:hover td {
  background: #fafafa;
}

.table-section tr.error td {
  color: #f56c6c;
}

.table-section .file-name {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.table-section .ratio-good {
  color: #67c23a;
  font-weight: 500;
}

.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.status-badge.done {
  background: #f0f9eb;
  color: #67c23a;
}

.status-badge.error {
  background: #fef0f0;
  color: #f56c6c;
}

.btn-small {
  padding: 4px 12px;
  font-size: 12px;
}

/* Settings Dialog Styles */
.settings-dialog .settings-content {
  padding: 16px 20px;
}

.settings-dialog .setting-item {
  margin-bottom: 28px;
}

.settings-dialog .setting-item label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 12px;
  color: #333;
  text-align: center;
}

.settings-dialog .setting-value {
  font-size: 32px;
  font-weight: 600;
  color: #409eff;
  margin-bottom: 8px;
  text-align: center;
}

.settings-dialog .setting-item .el-slider {
  margin: 0;
}

.settings-dialog .setting-hint {
  font-size: 12px;
  color: #999;
  margin-top: 12px;
  text-align: center;
}

/* Left-aligned setting item */
.settings-dialog .setting-item-left {
  margin-bottom: 0;
}

.settings-dialog .setting-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.settings-dialog .setting-row label {
  margin-bottom: 0;
  text-align: left;
  font-weight: normal;
}

/* Switch styles */
.switch {
  position: relative;
  display: inline-block;
  width: 48px;
  height: 24px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.switch .slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #ccc;
  transition: 0.3s;
  border-radius: 24px;
}

.switch .slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

.switch input:checked + .slider {
  background-color: #409eff;
}

.switch input:checked + .slider:before {
  transform: translateX(24px);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>