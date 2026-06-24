package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/yezihack/PicSlim/internal/dto"
	"github.com/yezihack/PicSlim/internal/events"
	"github.com/yezihack/PicSlim/internal/preview"
	"github.com/yezihack/PicSlim/internal/report"
	"github.com/yezihack/PicSlim/internal/scanner"
	"github.com/yezihack/PicSlim/internal/task"

	"github.com/disintegration/imaging"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/zap"
)

// App 应用主结构
type App struct {
	ctx      context.Context
	logger   *zap.Logger
	mu       sync.RWMutex

	// 模块
	scanner  *scanner.Scanner
	manager  *task.Manager
	emitter  *events.Emitter
	reporter *report.Reporter
	previewer *preview.Previewer

	// 任务状态
	outputDir string

	// 扫描结果
	scanResult *dto.ScanResult

	// 预览导航
	previewNav *preview.NavigationHelper
}

// New 创建新的应用实例
func New(logger *zap.Logger) *App {
	return &App{
		logger: logger,
	}
}

// Startup 应用启动
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.emitter = events.NewEmitter(ctx, a.logger)

	// 创建扫描器
	a.scanner = scanner.New()

	// 创建任务管理器
	a.manager = task.NewManager(a.logger)
	a.manager.SetEmitter(a.emitter)

	// 创建报告生成器
	a.reporter = report.New(a.logger)

	// 创建预览管理器
	a.previewer = preview.New(a.logger, "")

	a.logger.Info("application started")
}

// SelectInputPaths 选择输入路径
func (a *App) SelectInputPaths() dto.SelectPathsResult {
	files, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择图片文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "图片文件 (*.jpg, *.jpeg, *.png, *.webp)", Pattern: "*.jpg;*.jpeg;*.png;*.webp"},
			{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
		},
	})

	if err != nil {
		a.logger.Error("failed to open file dialog", zap.Error(err))
		return dto.SelectPathsResult{
			Code:    -1,
			Message: err.Error(),
		}
	}

	if len(files) == 0 {
		return dto.SelectPathsResult{
			Code:    0,
			Message: "cancelled",
		}
	}

	return dto.SelectPathsResult{
		Code:    0,
		Message: "success",
		Paths:   files,
	}
}

// SelectInputDir 选择输入目录
func (a *App) SelectInputDir() dto.SelectDirResult {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择输入目录",
	})

	if err != nil {
		a.logger.Error("failed to open directory dialog", zap.Error(err))
		return dto.SelectDirResult{
			Code:    -1,
			Message: err.Error(),
		}
	}

	if dir == "" {
		return dto.SelectDirResult{
			Code:    0,
			Message: "cancelled",
		}
	}

	return dto.SelectDirResult{
		Code:    0,
		Message: "success",
		Path:    dir,
	}
}

// SelectOutputDir 选择输出目录
func (a *App) SelectOutputDir() dto.SelectDirResult {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择输出目录",
	})

	if err != nil {
		a.logger.Error("failed to open directory dialog", zap.Error(err))
		return dto.SelectDirResult{
			Code:    -1,
			Message: err.Error(),
		}
	}

	if dir == "" {
		return dto.SelectDirResult{
			Code:    0,
			Message: "cancelled",
		}
	}

	return dto.SelectDirResult{
		Code:    0,
		Message: "success",
		Path:    dir,
	}
}

// ScanInputPaths 扫描输入路径
func (a *App) ScanInputPaths(paths []string) dto.ScanResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	result, err := a.scanner.ScanPaths(paths)
	if err != nil {
		return dto.ScanResult{
			Code:    -1,
			Message: err.Error(),
		}
	}

	a.scanResult = result

	a.logger.Info("scan completed",
		zap.Int("totalFiles", result.TotalFiles),
		zap.Int("filtered", len(result.FilteredFiles)))

	return *result
}

// StartCompress 开始压缩（新接口）
func (a *App) StartCompress(inputDir, outputDir, preset string, concurrency int) dto.BasicResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 确定输出目录
	if outputDir == "" {
		// 自动创建输出目录
		if info, err := os.Stat(inputDir); err == nil && info.IsDir() {
			outputDir = inputDir + "-compressed"
		} else {
			outputDir = filepath.Dir(inputDir) + "-compressed"
		}

		// 检查目录是否存在，若存在则追加时间戳
		if _, err := os.Stat(outputDir); err == nil {
			timestamp := time.Now().Format("0150405")
			outputDir = outputDir + "-" + timestamp
		}
	}

	a.outputDir = outputDir

	err := a.manager.StartCompress(a.ctx, inputDir, outputDir, preset, concurrency)
	if err != nil {
		return dto.BasicResult{Code: -1, Message: err.Error()}
	}

	return dto.BasicResult{Code: 0, Message: "started"}
}

// PauseTask 暂停任务
func (a *App) PauseTask() dto.BasicResult {
	if err := a.manager.PauseTask(); err != nil {
		return dto.BasicResult{Code: -1, Message: err.Error()}
	}
	return dto.BasicResult{Code: 0, Message: "paused"}
}

// ResumeTask 继续任务
func (a *App) ResumeTask() dto.BasicResult {
	if err := a.manager.ResumeTask(); err != nil {
		return dto.BasicResult{Code: -1, Message: err.Error()}
	}
	return dto.BasicResult{Code: 0, Message: "resumed"}
}

// CancelTask 取消任务
func (a *App) CancelTask() dto.BasicResult {
	if err := a.manager.CancelTask(); err != nil {
		return dto.BasicResult{Code: -1, Message: err.Error()}
	}
	return dto.BasicResult{Code: 0, Message: "cancelled"}
}

// GetTaskStatus 获取任务状态
func (a *App) GetTaskStatus() dto.TaskStatusResult {
	total, done, failed, status := a.manager.GetStatus()
	return dto.TaskStatusResult{
		Total:  total,
		Done:   done,
		Failed: failed,
		Status: status,
	}
}

// OpenOutputDir 打开输出目录
func (a *App) OpenOutputDir() dto.BasicResult {
	a.mu.RLock()
	outputDir := a.outputDir
	a.mu.RUnlock()

	if outputDir == "" {
		return dto.BasicResult{Code: -1, Message: "no output directory"}
	}

	// 检查目录是否存在
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return dto.BasicResult{Code: -1, Message: "output directory does not exist"}
	}

	// 使用系统命令打开目录
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", outputDir)
	case "darwin":
		cmd = exec.Command("open", outputDir)
	default: // linux
		cmd = exec.Command("xdg-open", outputDir)
	}

	if err := cmd.Start(); err != nil {
		return dto.BasicResult{Code: -1, Message: err.Error()}
	}

	return dto.BasicResult{Code: 0, Message: "opened"}
}

// ExportReport 导出报告
func (a *App) ExportReport(format string) dto.ExportResult {
	// 选择保存路径
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存报告",
		DefaultFilename: "compression_report.csv",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV 文件 (*.csv)", Pattern: "*.csv"},
		},
	})

	if err != nil {
		return dto.ExportResult{Code: -1, Message: err.Error()}
	}

	if filePath == "" {
		return dto.ExportResult{Code: 0, Message: "cancelled"}
	}

	// TODO: 实现报告导出
	if err := a.reporter.ExportCSV(nil, filePath); err != nil {
		return dto.ExportResult{Code: -1, Message: err.Error()}
	}

	return dto.ExportResult{
		Code:    0,
		Message: "success",
		Path:    filePath,
	}
}

// GetOutputDir 获取输出目录
func (a *App) GetOutputDir() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.outputDir
}

// CompletedJobResult 已完成任务结果
type CompletedJobResult struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Jobs    []*task.CompletedJob `json:"jobs"`
}

// GetCompletedJobs 获取已完成的任务列表
func (a *App) GetCompletedJobs() CompletedJobResult {
	jobs := a.manager.GetCompletedJobs()
	return CompletedJobResult{
		Code:    0,
		Message: "success",
		Jobs:    jobs,
	}
}

// PreviewResult 预览结果
type PreviewResult struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	Index        int    `json:"index"`
	FileName     string `json:"fileName"`
	SourceBase64 string `json:"sourceBase64"`
	TargetBase64 string `json:"targetBase64"`
	SourceSize   int64  `json:"sourceSize"`
	TargetSize   int64  `json:"targetSize"`
	Ratio        string `json:"ratio"`
}

// GetPreview 获取预览图片
func (a *App) GetPreview(index int) PreviewResult {
	job := a.manager.GetPreviewInfo(index)
	if job == nil {
		return PreviewResult{
			Code:    -1,
			Message: "任务不存在",
		}
	}

	result := PreviewResult{
		Code:       0,
		Index:      job.Index,
		FileName:   job.FileName,
		SourceSize: job.OldSize,
		TargetSize: job.NewSize,
		Ratio:      job.Ratio,
	}

	// 加载源图片
	sourceBase64, err := a.loadImageAsBase64(job.SourcePath)
	if err != nil {
		result.Message = fmt.Sprintf("无法加载源图片: %v", err)
		result.Code = -1
		return result
	}
	result.SourceBase64 = sourceBase64

	// 加载目标图片（如果存在）
	if job.TargetPath != "" && job.Status == "done" {
		targetBase64, err := a.loadImageAsBase64(job.TargetPath)
		if err != nil {
			result.Message = fmt.Sprintf("无法加载压缩后图片: %v", err)
		} else {
			result.TargetBase64 = targetBase64
		}
	}

	return result
}

// loadImageAsBase64 加载图片并转换为 Base64
func (a *App) loadImageAsBase64(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	var img image.Image
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	case ".webp":
		img, err = imaging.Decode(file)
	default:
		file.Seek(0, 0)
		img, err = imaging.Decode(file)
	}

	if err != nil {
		return "", fmt.Errorf("无法解码图片: %w", err)
	}

	// 缩放图片以适应预览
	bounds := img.Bounds()
	maxWidth := 800
	maxHeight := 600
	if bounds.Dx() > maxWidth || bounds.Dy() > maxHeight {
		img = imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)
	}

	// 编码为 JPEG Base64
	var buf strings.Builder
	buf.WriteString("data:image/jpeg;base64,")

	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	if err := jpeg.Encode(encoder, img, &jpeg.Options{Quality: 85}); err != nil {
		return "", fmt.Errorf("无法编码图片: %w", err)
	}
	encoder.Close()

	return buf.String(), nil
}