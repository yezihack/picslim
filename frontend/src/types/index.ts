// Task states
export type TaskState = 'PENDING' | 'RUNNING' | 'PAUSED' | 'COMPLETED' | 'FAILED' | 'CANCELLED';

// File job status
export type FileJobStatus = 'PENDING' | 'RUNNING' | 'SUCCESS' | 'FAILED' | 'SKIPPED_NO_GAIN';

// Preset types
export type Preset = 'HIGH_QUALITY' | 'BALANCED' | 'HIGH_COMPRESSION';

// Output mode
export type OutputMode = 'AUTO' | 'MANUAL';

// Task configuration
export interface TaskConfig {
  inputPaths: string[];
  outputMode: string;
  outputDir: string;
  keepStructure: boolean;
  nameConflictPolicy: string;
  preset: string;
  qualityMin: number;
  maxWidth: number;
  maxHeight: number;
  retryTimes: number;
  retryBackoffSeconds: number[];
}

// File job
export interface FileJob {
  jobId: string;
  taskId: string;
  sourcePath: string;
  targetPath: string;
  format: string;
  attempt: number;
  maxAttempts: number;
  status: FileJobStatus;
  errorCode: string;
  errorMessage: string;
  bytesBefore: number;
  bytesAfter: number;
  durationMs: number;
}

// Progress snapshot
export interface ProgressSnapshot {
  taskId: string;
  state: string;
  totalFiles: number;
  doneFiles: number;
  successFiles: number;
  failedFiles: number;
  currentFile: string;
  currentFileProgressPercent: number;
  elapsedSeconds: number;
  remainingSeconds: number;
  etaSeconds: number;
  throughputFilesPerSec: number;
}

// Result summary
export interface ResultSummary {
  totalFiles: number;
  successFiles: number;
  failedFiles: number;
  bytesBefore: number;
  bytesAfter: number;
  bytesSaved: number;
  savedPercent: number;
}

// Basic result
export interface BasicResult {
  code: number;
  message: string;
}

// Select paths result
export interface SelectPathsResult {
  code: number;
  message: string;
  paths: string[];
}

// Select dir result
export interface SelectDirResult {
  code: number;
  message: string;
  path: string;
}

// Create task result
export interface CreateTaskResult {
  code: number;
  message: string;
  taskId: string;
}

// File job page
export interface FileJobPage {
  code: number;
  message: string;
  total: number;
  page: number;
  pageSize: number;
  items: FileJob[];
}

// Scan result
export interface ScanResult {
  code: number;
  message: string;
  totalFiles: number;
  totalBytes: number;
  supportedFiles: FileInfo[];
  filteredFiles: FilterInfo[];
}

// File info
export interface FileInfo {
  path: string;
  name: string;
  format: string;
  size: number;
  modTime: string;
}

// Filter info
export interface FilterInfo {
  path: string;
  reason: string;
}

// Export result
export interface ExportResult {
  code: number;
  message: string;
  path: string;
}

// Progress event
export interface ProgressEvent {
  taskId: string;
  state: string;
  totalFiles: number;
  doneFiles: number;
  successFiles: number;
  failedFiles: number;
  currentFile: string;
  currentFileProgressPercent: number;
  elapsedSeconds: number;
  remainingSeconds: number;
  etaSeconds: number;
  throughputFilesPerSec: number;
}

// State change event
export interface StateChangeEvent {
  taskId: string;
  oldState: string;
  newState: string;
}

// File done event
export interface FileDoneEvent {
  taskId: string;
  jobId: string;
  sourcePath: string;
  targetPath: string;
  status: string;
  bytesBefore: number;
  bytesAfter: number;
  durationMs: number;
}

// Preview pair result
export interface PreviewPairResult {
  code: number;
  message: string;
  jobId: string;
  sourcePath: string;
  targetPath: string;
  sourceSize: number;
  targetSize: number;
  sourceBase64: string;
  targetBase64: string;
}

// Format stats
export interface FormatStats {
  totalFiles: number;
  successFiles: number;
  failedFiles: number;
  bytesBefore: number;
  bytesAfter: number;
  bytesSaved: number;
  savedPercent: number;
}

// Task init event
export interface TaskInitEvent {
  total: number;
}

// Log entry
export interface LogEntry {
  index: number;
  fileName: string;
  oldSize: number;
  newSize: number;
  ratio: string;
  status: 'done' | 'error';
  message: string;
  time: string;
}

// Task progress event
export interface TaskProgressEvent {
  total: number;
  done: number;
  failed: number;
}

// Task status result
export interface TaskStatusResult {
  total: number;
  done: number;
  failed: number;
  status: string;
}