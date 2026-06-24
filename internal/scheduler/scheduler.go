package scheduler

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yezihack/PicSlim/internal/dto"

	"go.uber.org/zap"
)

// Processor 文件处理器接口
type Processor interface {
	Process(ctx context.Context, job *dto.FileJob) (*dto.FileJob, error)
}

// Scheduler 任务调度器
type Scheduler struct {
	mu           sync.RWMutex
	logger       *zap.Logger
	processor    Processor
	workerCount  int

	// 任务状态
	taskID       string
	state        dto.TaskState
	jobs         []*dto.FileJob
	jobQueue     chan int // job index
	pendingJobs  []int    // 等待处理的 job indices

	// 进度追踪
	totalFiles    int32
	doneFiles     int32
	successFiles  int32
	failedFiles   int32
	currentFile   string
	startTime     time.Time

	// 控制
	ctx        context.Context
	cancel     context.CancelFunc
	pauseCh    chan struct{}
	resumeCh   chan struct{}
	doneCh     chan struct{}

	// 事件回调
	onProgress   func(snapshot *dto.ProgressSnapshot)
	onStateChange func(taskID string, oldState, newState dto.TaskState)
	onFileDone   func(job *dto.FileJob)
	onError      func(taskID string, err error)
}

// New 创建新的调度器
func New(logger *zap.Logger, processor Processor) *Scheduler {
	workerCount := runtime.NumCPU() - 1
	if workerCount < 1 {
		workerCount = 1
	}

	return &Scheduler{
		logger:      logger,
		processor:   processor,
		workerCount: workerCount,
		state:       dto.TaskStatePending,
		pauseCh:     make(chan struct{}),
		resumeCh:    make(chan struct{}),
		doneCh:      make(chan struct{}),
	}
}

// CreateTask 创建任务
func (s *Scheduler) CreateTask(taskID string, jobs []*dto.FileJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != dto.TaskStatePending && s.state != dto.TaskStateCompleted &&
	   s.state != dto.TaskStateFailed && s.state != dto.TaskStateCancelled {
		return fmt.Errorf("cannot create task in state: %s", s.state)
	}

	s.taskID = taskID
	s.jobs = jobs
	s.pendingJobs = make([]int, len(jobs))
	for i := range jobs {
		s.pendingJobs[i] = i
	}

	atomic.StoreInt32(&s.totalFiles, int32(len(jobs)))
	atomic.StoreInt32(&s.doneFiles, 0)
	atomic.StoreInt32(&s.successFiles, 0)
	atomic.StoreInt32(&s.failedFiles, 0)
	s.state = dto.TaskStatePending
	s.startTime = time.Time{}

	s.logger.Info("task created",
		zap.String("taskId", taskID),
		zap.Int("totalFiles", len(jobs)))

	return nil
}

// Start 启动任务
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != dto.TaskStatePending && s.state != dto.TaskStatePaused {
		return fmt.Errorf("cannot start task in state: %s", s.state)
	}

	oldState := s.state
	s.state = dto.TaskStateRunning
	s.startTime = time.Now()

	// 创建上下文
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// 初始化任务队列
	s.jobQueue = make(chan int, len(s.pendingJobs))

	// 将待处理的任务放入队列
	for _, idx := range s.pendingJobs {
		s.jobQueue <- idx
	}
	s.pendingJobs = s.pendingJobs[:0] // 清空待处理队列

	// 启动 worker
	for i := 0; i < s.workerCount; i++ {
		go s.worker(i)
	}

	// 启动进度监控
	go s.monitorProgress()

	s.notifyStateChange(oldState, s.state)
	s.logger.Info("task started", zap.String("taskId", s.taskID))

	return nil
}

// Pause 暂停任务
func (s *Scheduler) Pause() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != dto.TaskStateRunning {
		return fmt.Errorf("cannot pause task in state: %s", s.state)
	}

	oldState := s.state
	s.state = dto.TaskStatePaused

	// 关闭任务队列，停止分发新任务
	close(s.jobQueue)

	// 收集未处理的任务
	for idx := range s.jobQueue {
		s.pendingJobs = append(s.pendingJobs, idx)
	}

	s.notifyStateChange(oldState, s.state)
	s.logger.Info("task paused", zap.String("taskId", s.taskID))

	return nil
}

// Resume 继续任务
func (s *Scheduler) Resume() error {
	return s.Start()
}

// Cancel 取消任务
func (s *Scheduler) Cancel() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != dto.TaskStateRunning && s.state != dto.TaskStatePaused {
		return fmt.Errorf("cannot cancel task in state: %s", s.state)
	}

	oldState := s.state
	s.state = dto.TaskStateCancelled

	if s.cancel != nil {
		s.cancel()
	}

	s.notifyStateChange(oldState, s.state)
	s.logger.Info("task cancelled", zap.String("taskId", s.taskID))

	return nil
}

// GetSnapshot 获取进度快照
func (s *Scheduler) GetSnapshot() *dto.ProgressSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	elapsed := 0
	if !s.startTime.IsZero() {
		elapsed = int(time.Since(s.startTime).Seconds())
	}

	done := int(atomic.LoadInt32(&s.doneFiles))
	total := int(atomic.LoadInt32(&s.totalFiles))
	success := int(atomic.LoadInt32(&s.successFiles))
	failed := int(atomic.LoadInt32(&s.failedFiles))

	var eta int
	if done > 0 && elapsed > 0 {
		avgPerFile := float64(elapsed) / float64(done)
		remaining := total - done
		eta = int(avgPerFile * float64(remaining))
	}

	var throughput float64
	if elapsed > 0 {
		throughput = float64(done) / float64(elapsed)
	}

	return &dto.ProgressSnapshot{
		TaskID:                     s.taskID,
		State:                      string(s.state),
		TotalFiles:                 total,
		DoneFiles:                  done,
		SuccessFiles:               success,
		FailedFiles:                failed,
		CurrentFile:                s.currentFile,
		CurrentFileProgressPercent: 0,
		ElapsedSeconds:             elapsed,
		RemainingSeconds:           eta,
		ETASeconds:                 eta,
		ThroughputFilesPerSec:      throughput,
	}
}

// GetState 获取任务状态
func (s *Scheduler) GetState() dto.TaskState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// SetCallbacks 设置事件回调
func (s *Scheduler) SetCallbacks(
	onProgress func(*dto.ProgressSnapshot),
	onStateChange func(string, dto.TaskState, dto.TaskState),
	onFileDone func(*dto.FileJob),
	onError func(string, error),
) {
	s.onProgress = onProgress
	s.onStateChange = onStateChange
	s.onFileDone = onFileDone
	s.onError = onError
}

// worker 工作协程
func (s *Scheduler) worker(id int) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case idx, ok := <-s.jobQueue:
			if !ok {
				return
			}

			job := s.jobs[idx]

			// 更新当前处理的文件
			s.mu.Lock()
			s.currentFile = job.SourcePath
			s.mu.Unlock()

			// 处理文件
			result, err := s.processor.Process(s.ctx, job)
			if err != nil {
				s.logger.Error("process file failed",
					zap.String("taskId", s.taskID),
					zap.String("file", job.SourcePath),
					zap.Error(err))

				if s.onError != nil {
					s.onError(s.taskID, err)
				}
			}

			// 更新结果
			if result != nil {
				s.jobs[idx] = result
				if result.Status == dto.FileJobStatusSuccess || result.Status == dto.FileJobStatusSkippedNoGain {
					atomic.AddInt32(&s.successFiles, 1)
				} else if result.Status == dto.FileJobStatusFailed {
					atomic.AddInt32(&s.failedFiles, 1)
				}
			}

			atomic.AddInt32(&s.doneFiles, 1)

			if s.onFileDone != nil {
				s.onFileDone(result)
			}
		}
	}
}

// monitorProgress 监控进度
func (s *Scheduler) monitorProgress() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if s.onProgress != nil {
				snapshot := s.GetSnapshot()
				s.onProgress(snapshot)

				// 检查是否完成
				if snapshot.DoneFiles >= snapshot.TotalFiles && snapshot.TotalFiles > 0 {
					s.mu.Lock()
					oldState := s.state
					s.state = dto.TaskStateCompleted
					s.mu.Unlock()

					s.notifyStateChange(oldState, s.state)
					s.logger.Info("task completed",
						zap.String("taskId", s.taskID),
						zap.Int("success", int(atomic.LoadInt32(&s.successFiles))),
						zap.Int("failed", int(atomic.LoadInt32(&s.failedFiles))))
					return
				}
			}
		}
	}
}

// notifyStateChange 通知状态变化
func (s *Scheduler) notifyStateChange(oldState, newState dto.TaskState) {
	if s.onStateChange != nil {
		s.onStateChange(s.taskID, oldState, newState)
	}
}

// GetResultSummary 获取结果摘要
func (s *Scheduler) GetResultSummary() *dto.ResultSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bytesBefore, bytesAfter int64
	var successCount, failedCount int

	for _, job := range s.jobs {
		bytesBefore += job.BytesBefore
		if job.Status == dto.FileJobStatusSuccess || job.Status == dto.FileJobStatusSkippedNoGain {
			bytesAfter += job.BytesAfter
			successCount++
		} else if job.Status == dto.FileJobStatusFailed {
			failedCount++
		}
	}

	bytesSaved := bytesBefore - bytesAfter
	var savedPercent float64
	if bytesBefore > 0 {
		savedPercent = float64(bytesSaved) / float64(bytesBefore) * 100
	}

	return &dto.ResultSummary{
		TotalFiles:   len(s.jobs),
		SuccessFiles: successCount,
		FailedFiles:  failedCount,
		BytesBefore:  bytesBefore,
		BytesAfter:   bytesAfter,
		BytesSaved:   bytesSaved,
		SavedPercent: savedPercent,
	}
}

// GetFileJobs 获取文件任务列表
func (s *Scheduler) GetFileJobs(page, pageSize int) ([]*dto.FileJob, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.jobs)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		return []*dto.FileJob{}, total
	}
	if end > total {
		end = total
	}

	result := make([]*dto.FileJob, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.jobs[i])
	}

	return result, total
}