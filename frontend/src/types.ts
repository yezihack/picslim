// API result types
export interface BasicResult {
  code: number;
  message: string;
}

export interface SelectPathsResult {
  code: number;
  message: string;
  paths: string[];
}

export interface SelectDirResult {
  code: number;
  message: string;
  path: string;
}

export interface FileInfo {
  path: string;
  name: string;
  format: string;
  size: number;
  modTime: string;
}

export interface FilterInfo {
  path: string;
  reason: string;
}

export interface ScanResult {
  code: number;
  message: string;
  totalFiles: number;
  totalBytes: number;
  supportedFiles: FileInfo[];
  filteredFiles: FilterInfo[];
}

export interface ExportResult {
  code: number;
  message: string;
  path: string;
}

export interface TaskStatusResult {
  total: number;
  done: number;
  failed: number;
  status: string;
}

// Log entry from events
export interface LogEntry {
  index: number;
  fileName: string;
  oldSize: number;
  newSize: number;
  ratio: string;
  status: string; // "done" | "error"
  message: string;
  time: string;
}

// Completed job from API
export interface CompletedJob {
  index: number;
  fileName: string;
  sourcePath: string;
  targetPath: string;
  oldSize: number;
  newSize: number;
  status: string;
  message: string;
  ratio: string;
}

export interface CompletedJobResult {
  code: number;
  message: string;
  jobs: CompletedJob[];
}

export interface PreviewResult {
  code: number;
  message: string;
  index: number;
  fileName: string;
  sourceBase64: string;
  targetBase64: string;
  sourceSize: number;
  targetSize: number;
  ratio: string;
}