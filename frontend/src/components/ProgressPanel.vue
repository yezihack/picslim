<script setup lang="ts">
import { computed } from 'vue';
import type { ProgressSnapshot } from '../types';

// Props
const props = defineProps<{
  progress: ProgressSnapshot;
  concurrency: number;
  preset: string;
}>();

// Computed
const overallPercent = computed(() => {
  if (props.progress.totalFiles === 0) return 0;
  return (props.progress.doneFiles / props.progress.totalFiles) * 100;
});

const currentPercent = computed(() => props.progress.currentFileProgressPercent);

const speedText = computed(() => {
  const throughput = props.progress.throughputFilesPerSec;
  if (throughput === 0) return '0 张/秒';
  return `${throughput.toFixed(1)} 张/秒`;
});

const presetText = computed(() => {
  switch (props.preset) {
    case 'HIGH_QUALITY': return '高画质';
    case 'BALANCED': return '均衡';
    case 'HIGH_COMPRESSION': return '高压缩';
    default: return props.preset;
  }
});

const isRunning = computed(() => props.progress.state === 'RUNNING');
const isPaused = computed(() => props.progress.state === 'PAUSED');
const isCompleted = computed(() => props.progress.state === 'COMPLETED');

// Methods
function formatClock(seconds: number): string {
  const safe = Math.max(0, Math.floor(seconds));
  const hh = String(Math.floor(safe / 3600)).padStart(2, '0');
  const mm = String(Math.floor((safe % 3600) / 60)).padStart(2, '0');
  const ss = String(safe % 60).padStart(2, '0');
  return `${hh}:${mm}:${ss}`;
}

function getStateText(state: string): string {
  switch (state) {
    case 'RUNNING': return '进行中';
    case 'PAUSED': return '已暂停';
    case 'COMPLETED': return '已完成';
    case 'PENDING': return '待开始';
    case 'FAILED': return '失败';
    case 'CANCELLED': return '已取消';
    default: return state;
  }
}

function getStateClass(state: string): string {
  switch (state) {
    case 'RUNNING': return 'chip-running';
    case 'PAUSED': return 'chip-paused';
    case 'COMPLETED': return 'chip-completed';
    case 'FAILED': return 'chip-failed';
    default: return '';
  }
}
</script>

<template>
  <div class="panel-head">
    <h2>任务进度</h2>
    <div class="status-chips">
      <span class="chip" :class="getStateClass(progress.state)">
        {{ getStateText(progress.state) }}
      </span>
      <span class="chip">自动重试 3 次</span>
    </div>
  </div>

  <p class="muted">
    当前文件：
    <span v-if="progress.currentFile">{{ progress.currentFile }}</span>
    <span v-else>-</span>
  </p>

  <div class="mini-kpis">
    <div class="mini-kpi">
      <span>并发任务</span>
      <strong>{{ concurrency }}</strong>
    </div>
    <div class="mini-kpi">
      <span>平均速度</span>
      <strong>{{ speedText }}</strong>
    </div>
    <div class="mini-kpi">
      <span>预计完成</span>
      <strong>{{ formatClock(progress.etaSeconds) }}</strong>
    </div>
  </div>

  <div class="progress-wrap">
    <div class="progress-label">
      <span>总进度 {{ overallPercent.toFixed(0) }}%</span>
      <span>{{ progress.doneFiles }} / {{ progress.totalFiles }}</span>
    </div>
    <div class="progress-bar">
      <div class="progress-fill" :style="{ width: `${overallPercent}%` }"></div>
    </div>
  </div>

  <div class="progress-wrap" v-if="isRunning">
    <div class="progress-label">
      <span>当前文件进度 {{ currentPercent.toFixed(0) }}%</span>
      <span>已运行 {{ formatClock(progress.elapsedSeconds) }}</span>
    </div>
    <div class="progress-bar">
      <div class="progress-fill current" :style="{ width: `${currentPercent}%` }"></div>
    </div>
    <div class="progress-meta">
      <span>预计剩余：{{ formatClock(progress.remainingSeconds) }}</span>
      <span>目标质量：{{ presetText }}</span>
    </div>
  </div>

  <!-- Stats when completed -->
  <template v-if="isCompleted">
    <h3>压缩结果统计</h3>
    <div class="stats-grid">
      <div class="stat-box">
        <p>总文件</p>
        <strong>{{ progress.totalFiles }}</strong>
      </div>
      <div class="stat-box">
        <p>成功</p>
        <strong>{{ progress.successFiles }}</strong>
      </div>
      <div class="stat-box">
        <p>失败</p>
        <strong>{{ progress.failedFiles }}</strong>
      </div>
      <div class="stat-box">
        <p>耗时</p>
        <strong>{{ formatClock(progress.elapsedSeconds) }}</strong>
      </div>
    </div>
  </template>
</template>

<style scoped>
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

.chip-failed {
  background: var(--danger-light, #fef0f0);
  color: var(--danger, #f56c6c);
}

.muted {
  color: #666;
  font-size: 14px;
  margin: 12px 0;
}

.mini-kpis {
  display: flex;
  gap: 16px;
  margin: 16px 0;
}

.mini-kpi {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.mini-kpi span {
  font-size: 12px;
  color: #666;
}

.mini-kpi strong {
  font-size: 16px;
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
  height: 8px;
  background: #e0e0e0;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--primary, #409eff);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-fill.current {
  background: var(--success, #67c23a);
}

.progress-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #666;
  margin-top: 8px;
}

h3 {
  font-size: 14px;
  margin: 20px 0 12px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
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

@media (max-width: 600px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>