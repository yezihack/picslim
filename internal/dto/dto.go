package dto

// TaskState 任务状态定义
type TaskState string

const (
	TaskStatePending   TaskState = "PENDING"
	TaskStateRunning   TaskState = "RUNNING"
	TaskStatePaused    TaskState = "PAUSED"
	TaskStateCompleted TaskState = "COMPLETED"
	TaskStateFailed    TaskState = "FAILED"
	TaskStateCancelled TaskState = "CANCELLED"
)

// FileJobStatus 单文件状态
type FileJobStatus string

const (
	FileJobStatusPending         FileJobStatus = "PENDING"
	FileJobStatusRunning         FileJobStatus = "RUNNING"
	FileJobStatusSuccess         FileJobStatus = "SUCCESS"
	FileJobStatusFailed          FileJobStatus = "FAILED"
	FileJobStatusSkippedNoGain   FileJobStatus = "SKIPPED_NO_GAIN"
)

// OutputMode 输出模式
type OutputMode string

const (
	OutputModeAuto   OutputMode = "AUTO"
	OutputModeManual OutputMode = "MANUAL"
)

// Preset 压缩预设
type Preset string

const (
	PresetHighQuality    Preset = "HIGH_QUALITY"
	PresetBalanced       Preset = "BALANCED"
	PresetHighCompression Preset = "HIGH_COMPRESSION"
)

// TaskConfig 任务配置
type TaskConfig struct {
	InputPaths           []string `json:"inputPaths"`
	OutputMode           string   `json:"outputMode"`
	OutputDir            string   `json:"outputDir"`
	KeepStruct           bool     `json:"keepStructure"`
	NameConflictPolicy   string   `json:"nameConflictPolicy"`
	Preset               string   `json:"preset"`
	QualityMin           int      `json:"qualityMin"`
	MaxWidth             int      `json:"maxWidth"`
	MaxHeight            int      `json:"maxHeight"`
	RetryTimes           int      `json:"retryTimes"`
	RetryBackoffSeconds  []int    `json:"retryBackoffSeconds"`
}

// Task 任务实体
type Task struct {
	TaskID    string    `json:"taskId"`
	State     TaskState `json:"state"`
	Config    TaskConfig `json:"config"`
	CreatedAt string    `json:"createdAt"`
	StartedAt string    `json:"startedAt"`
	FinishedAt string   `json:"finishedAt"`
}

// FileJob 单文件任务
type FileJob struct {
	JobID      string        `json:"jobId"`
	TaskID     string        `json:"taskId"`
	SourcePath string        `json:"sourcePath"`
	TargetPath string        `json:"targetPath"`
	Format     string        `json:"format"`
	Attempt    int           `json:"attempt"`
	MaxAttempts int          `json:"maxAttempts"`
	Status     FileJobStatus `json:"status"`
	ErrorCode  string        `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
	BytesBefore  int64       `json:"bytesBefore"`
	BytesAfter   int64       `json:"bytesAfter"`
	DurationMs   int64       `json:"durationMs"`
}

// ProgressSnapshot 进度快照
type ProgressSnapshot struct {
	TaskID                     string  `json:"taskId"`
	State                      string  `json:"state"`
	TotalFiles                 int     `json:"totalFiles"`
	DoneFiles                  int     `json:"doneFiles"`
	SuccessFiles               int     `json:"successFiles"`
	FailedFiles                int     `json:"failedFiles"`
	CurrentFile                string  `json:"currentFile"`
	CurrentFileProgressPercent float64 `json:"currentFileProgressPercent"`
	ElapsedSeconds             int     `json:"elapsedSeconds"`
	RemainingSeconds           int     `json:"remainingSeconds"`
	ETASeconds                 int     `json:"etaSeconds"`
	ThroughputFilesPerSec      float64 `json:"throughputFilesPerSec"`
}

// ResultSummary 结果统计
type ResultSummary struct {
	TotalFiles   int     `json:"totalFiles"`
	SuccessFiles int     `json:"successFiles"`
	FailedFiles  int     `json:"failedFiles"`
	BytesBefore  int64   `json:"bytesBefore"`
	BytesAfter   int64   `json:"bytesAfter"`
	BytesSaved   int64   `json:"bytesSaved"`
	SavedPercent float64 `json:"savedPercent"`
}

// BasicResult 基础返回结果
type BasicResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SelectPathsResult 选择路径结果
type SelectPathsResult struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Paths   []string `json:"paths"`
}

// SelectDirResult 选择目录结果
type SelectDirResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

// CreateTaskResult 创建任务结果
type CreateTaskResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TaskID  string `json:"taskId"`
}

// FileJobPage 文件任务分页结果
type FileJobPage struct {
	Code       int        `json:"code"`
	Message    string     `json:"message"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	Items      []*FileJob `json:"items"`
}

// ExportResult 导出结果
type ExportResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

// PreviewPairResult 预览对结果
type PreviewPairResult struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	JobID        string `json:"jobId"`
	SourcePath   string `json:"sourcePath"`
	TargetPath   string `json:"targetPath"`
	SourceSize   int64  `json:"sourceSize"`
	TargetSize   int64  `json:"targetSize"`
	SourceBase64 string `json:"sourceBase64"`
	TargetBase64 string `json:"targetBase64"`
}

// ScanResult 扫描结果
type ScanResult struct {
	Code            int          `json:"code"`
	Message         string       `json:"message"`
	TotalFiles      int          `json:"totalFiles"`
	TotalBytes      int64        `json:"totalBytes"`
	SupportedFiles  []FileInfo   `json:"supportedFiles"`
	FilteredFiles   []FilterInfo `json:"filteredFiles"`
}

// FileInfo 文件信息
type FileInfo struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Format    string `json:"format"`
	Size      int64  `json:"size"`
	ModTime   string `json:"modTime"`
}

// FilterInfo 过滤文件信息
type FilterInfo struct {
	Path      string `json:"path"`
	Reason    string `json:"reason"`
}

// TaskStatusResult 任务状态结果
type TaskStatusResult struct {
	Total  int    `json:"total"`
	Done   int    `json:"done"`
	Failed int    `json:"failed"`
	Status string `json:"status"`
}