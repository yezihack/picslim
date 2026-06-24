package task

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yezihack/picslim/internal/compressor"
	"github.com/yezihack/picslim/internal/dto"
	"github.com/yezihack/picslim/internal/events"
	"github.com/yezihack/picslim/internal/scanner"

	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

// PresetQuality 预设对应的质量参数
var PresetQuality = map[string]int{
	"HIGH_QUALITY":     85,
	"BALANCED":         70,
	"HIGH_COMPRESSION": 50,
}

// ImageJob 单张图片的任务单元
type ImageJob struct {
	Index      int
	InputPath  string
	OutputPath string
}

// CompletedJob 已完成的任务信息
type CompletedJob struct {
	Index      int    `json:"index"`
	FileName   string `json:"fileName"`
	SourcePath string `json:"sourcePath"`
	TargetPath string `json:"targetPath"`
	OldSize    int64  `json:"oldSize"`
	NewSize    int64  `json:"newSize"`
	Status     string `json:"status"` // "done" / "error" / "skipped"
	Message    string `json:"message"`
	Ratio      string `json:"ratio"`
}

// Manager 任务管理器
type Manager struct {
	ctx      context.Context
	logger   *zap.Logger
	emitter  *events.Emitter
	comp     *compressor.Compressor
	scanner  *scanner.Scanner

	// 任务状态
	total    int32
	done     int32
	failed   int32
	status   string // running / paused / cancelled / completed

	// 完成的任务列表
	completedJobs []*CompletedJob

	// 控制
	pool     *ants.PoolWithFunc
	wg       sync.WaitGroup
	cancelCh chan struct{}
	pauseMu  sync.Mutex
	paused   bool
	mu       sync.Mutex
}

// NewManager 创建新的任务管理器
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		logger:        logger,
		scanner:       scanner.New(),
		status:        "idle",
		completedJobs: make([]*CompletedJob, 0),
	}
}

// StartCompress 开始压缩
func (m *Manager) StartCompress(ctx context.Context, inputDir, outputDir, preset string, concurrency int) error {
	m.ctx = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1-1. 扫描目录，找出所有图片
	images, err := m.scanImages(inputDir)
	if err != nil {
		return fmt.Errorf("扫描目录失败: %w", err)
	}

	if len(images) == 0 {
		return fmt.Errorf("没有找到图片")
	}

	// 1-2. 初始化任务状态
	m.total = int32(len(images))
	m.done = 0
	m.failed = 0
	m.status = "running"
	m.cancelCh = make(chan struct{})
	m.paused = false
	m.completedJobs = make([]*CompletedJob, 0)

	// 创建压缩器
	quality := PresetQuality[preset]
	if quality == 0 {
		quality = 85 // 默认高质量
	}
	m.comp = compressor.New(m.logger, dto.Preset(preset), outputDir, true, 0, 0)

	// 1-3. 推送初始化事件给前端
	if m.emitter != nil {
		m.emitter.EmitInit(len(images))
	}

	// 1-4. 创建协程池
	m.pool, err = ants.NewPoolWithFunc(concurrency, func(i interface{}) {
		job := i.(*ImageJob)
		m.processImage(job)
	})
	if err != nil {
		return fmt.Errorf("创建协程池失败: %w", err)
	}

	// 1-5. 在独立 goroutine 里逐个提交任务
	go func() {
		defer m.pool.Release()

		for i, imgPath := range images {
			// 检查是否已取消
			if m.isCancelled() {
				return
			}

			// 构造输出路径：保持子目录结构
			outPath := m.buildOutputPath(imgPath, inputDir, outputDir)

			job := &ImageJob{
				Index:      i + 1,
				InputPath:  imgPath,
				OutputPath: outPath,
			}

			m.wg.Add(1)
			// 提交到协程池
			if err := m.pool.Invoke(job); err != nil {
				m.wg.Done()
				m.logger.Error("提交任务失败", zap.Error(err))
			}
		}

		// 等待所有任务完成
		m.wg.Wait()
	}()

	return nil
}

// processImage 处理单张图片
func (m *Manager) processImage(job *ImageJob) {
	defer m.wg.Done()

	// 2-1. 暂停检查
	m.waitIfPaused()

	// 2-2. 再次检查取消
	if m.isCancelled() {
		return
	}

	startTime := time.Now()
	oldSize := m.getFileSize(job.InputPath)

	// 2-3. 确保输出子目录存在
	os.MkdirAll(filepath.Dir(job.OutputPath), 0755)

	// 2-4. 执行压缩
	err := m.comp.ProcessToFile(job.InputPath, job.OutputPath)

	// 2-5. 加锁更新任务计数，推送事件
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := &events.LogEntry{
		Index:    job.Index,
		FileName: filepath.Base(job.InputPath),
		Time:     events.FormatTime(),
	}

	completedJob := &CompletedJob{
		Index:      job.Index,
		FileName:   filepath.Base(job.InputPath),
		SourcePath: job.InputPath,
		TargetPath: job.OutputPath,
		OldSize:    oldSize,
	}

	if err != nil {
		atomic.AddInt32(&m.failed, 1)
		entry.Status = "error"
		entry.Message = err.Error()
		completedJob.Status = "error"
		completedJob.Message = err.Error()
		m.logger.Error("压缩失败",
			zap.String("file", job.InputPath),
			zap.Error(err))
	} else {
		atomic.AddInt32(&m.done, 1)
		newSize := m.getFileSize(job.OutputPath)
		entry.OldSize = oldSize
		entry.NewSize = newSize
		entry.Ratio = events.FormatRatio(oldSize, newSize)
		entry.Status = "done"
		completedJob.Status = "done"
		completedJob.NewSize = newSize
		completedJob.Ratio = entry.Ratio
	}

	// 保存完成的任务
	m.completedJobs = append(m.completedJobs, completedJob)

	// 推送单条日志
	if m.emitter != nil {
		m.emitter.EmitLog(entry)
	}

	// 推送整体进度
	if m.emitter != nil {
		m.emitter.EmitProgress(int(m.total), int(m.done), int(m.failed))
	}

	m.logger.Debug("图片处理完成",
		zap.Int("index", job.Index),
		zap.String("file", job.InputPath),
		zap.Duration("duration", time.Since(startTime)))

	// 2-6. 全部完成时推送完成事件
	if m.done+m.failed == m.total {
		m.status = "completed"
		if m.emitter != nil {
			m.emitter.EmitCompleted(int(m.total), int(m.done), int(m.failed))
		}
	}
}

// PauseTask 暂停任务
func (m *Manager) PauseTask() error {
	m.pauseMu.Lock()
	m.paused = true
	m.pauseMu.Unlock()

	m.mu.Lock()
	m.status = "paused"
	m.mu.Unlock()

	if m.emitter != nil {
		m.emitter.EmitPaused()
	}
	return nil
}

// ResumeTask 继续任务
func (m *Manager) ResumeTask() error {
	m.pauseMu.Lock()
	m.paused = false
	m.pauseMu.Unlock()

	m.mu.Lock()
	m.status = "running"
	m.mu.Unlock()

	if m.emitter != nil {
		m.emitter.EmitResumed()
	}
	return nil
}

// CancelTask 取消任务
func (m *Manager) CancelTask() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancelCh != nil {
		close(m.cancelCh)
	}

	// 先解除暂停，避免 worker 永久阻塞
	m.pauseMu.Lock()
	m.paused = false
	m.pauseMu.Unlock()

	if m.pool != nil {
		m.pool.Release()
	}

	m.status = "cancelled"
	if m.emitter != nil {
		m.emitter.EmitCancelled()
	}
	return nil
}

// SetEmitter 设置事件发射器
func (m *Manager) SetEmitter(emitter *events.Emitter) {
	m.emitter = emitter
}

// GetStatus 获取任务状态
func (m *Manager) GetStatus() (total, done, failed int, status string) {
	return int(m.total), int(m.done), int(m.failed), m.status
}

// GetCompletedJobs 获取已完成的任务列表
func (m *Manager) GetCompletedJobs() []*CompletedJob {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.completedJobs
}

// GetPreviewInfo 获取预览信息
func (m *Manager) GetPreviewInfo(index int) *CompletedJob {
	m.mu.Lock()
	defer m.mu.Unlock()
	if index < 0 || index >= len(m.completedJobs) {
		return nil
	}
	return m.completedJobs[index]
}

// waitIfPaused 暂停检查
func (m *Manager) waitIfPaused() {
	for {
		m.pauseMu.Lock()
		p := m.paused
		m.pauseMu.Unlock()
		if !p {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// isCancelled 检查是否已取消
func (m *Manager) isCancelled() bool {
	if m.cancelCh == nil {
		return false
	}
	select {
	case <-m.cancelCh:
		return true
	default:
		return false
	}
}

// scanImages 扫描图片
func (m *Manager) scanImages(dir string) ([]string, error) {
	var images []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" {
			images = append(images, path)
		}
		return nil
	})

	return images, err
}

// buildOutputPath 构建输出路径
func (m *Manager) buildOutputPath(imgPath, inputDir, outputDir string) string {
	rel, err := filepath.Rel(inputDir, imgPath)
	if err != nil {
		return filepath.Join(outputDir, filepath.Base(imgPath))
	}
	return filepath.Join(outputDir, rel)
}

// getFileSize 获取文件大小
func (m *Manager) getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}