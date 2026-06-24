package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/yezihack/PicSlim/internal/dto"

	"go.uber.org/zap"
)

// DefaultBackoffTimes 默认退避时间（秒）
var DefaultBackoffTimes = []int{1, 2, 4}

// RetryableErrorCodes 可重试的错误码
var RetryableErrorCodes = map[string]bool{
	"E_DECODE":        true,
	"E_ENCODE":        true,
	"E_WRITE_OUTPUT":  true,
	"E_TEMP_FILE":     true,
}

// Retry 重试器
type Retry struct {
	logger       *zap.Logger
	maxAttempts  int
	backoffTimes []int
}

// New 创建新的重试器
func New(logger *zap.Logger, maxAttempts int, backoffTimes []int) *Retry {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if len(backoffTimes) == 0 {
		backoffTimes = DefaultBackoffTimes
	}

	return &Retry{
		logger:       logger,
		maxAttempts:  maxAttempts,
		backoffTimes: backoffTimes,
	}
}

// Processor 处理器接口
type Processor interface {
	Process(ctx context.Context, job *dto.FileJob) (*dto.FileJob, error)
}

// RetryableProcessor 可重试的处理器
type RetryableProcessor struct {
	processor Processor
	retry     *Retry
}

// NewRetryableProcessor 创建可重试的处理器
func NewRetryableProcessor(processor Processor, retry *Retry) *RetryableProcessor {
	return &RetryableProcessor{
		processor: processor,
		retry:     retry,
	}
}

// Process 处理文件，支持自动重试
func (rp *RetryableProcessor) Process(ctx context.Context, job *dto.FileJob) (*dto.FileJob, error) {
	for {
		result, err := rp.processor.Process(ctx, job)

		// 成功或无需重试
		if err == nil || !rp.shouldRetry(job) {
			return result, err
		}

		// 检查是否达到最大重试次数
		if job.Attempt >= rp.retry.maxAttempts {
			rp.retry.logger.Warn("max retry attempts reached",
				zap.String("jobId", job.JobID),
				zap.String("source", job.SourcePath),
				zap.Int("attempts", job.Attempt),
				zap.String("errorCode", job.ErrorCode))
			return result, fmt.Errorf("max retry attempts reached: %s", job.ErrorMessage)
		}

		// 等待退避时间
		backoffSec := rp.getBackoffTime(job.Attempt)
		rp.retry.logger.Info("retrying after backoff",
			zap.String("jobId", job.JobID),
			zap.String("source", job.SourcePath),
			zap.Int("attempt", job.Attempt),
			zap.Int("nextAttempt", job.Attempt+1),
			zap.Int("backoffSec", backoffSec))

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(time.Duration(backoffSec) * time.Second):
			// 继续重试
		}
	}
}

// shouldRetry 判断是否应该重试
func (rp *RetryableProcessor) shouldRetry(job *dto.FileJob) bool {
	if job.Status == dto.FileJobStatusSuccess || job.Status == dto.FileJobStatusSkippedNoGain {
		return false
	}

	if job.ErrorCode == "" {
		return false
	}

	// 检查错误码是否可重试
	return RetryableErrorCodes[job.ErrorCode]
}

// getBackoffTime 获取退避时间
func (rp *RetryableProcessor) getBackoffTime(attempt int) int {
	idx := attempt - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(rp.retry.backoffTimes) {
		idx = len(rp.retry.backoffTimes) - 1
	}
	return rp.retry.backoffTimes[idx]
}

// IsRetryableError 判断是否为可重试错误
func IsRetryableError(errorCode string) bool {
	return RetryableErrorCodes[errorCode]
}