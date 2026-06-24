package events

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/zap"
)

// Event types
const (
	EventTaskInit      = "task:init"
	EventTaskLog       = "task:log"
	EventTaskProgress  = "task:progress"
	EventTaskCompleted = "task:completed"
	EventTaskCancelled = "task:cancelled"
	EventTaskPaused    = "task:paused"
	EventTaskResumed   = "task:resumed"
)

// TaskInitEvent 任务初始化事件
type TaskInitEvent struct {
	Total int `json:"total"` // 图片总数
}

// LogEntry 单条日志（每张图处理完推一条）
type LogEntry struct {
	Index    int    `json:"index"`    // 第几张
	FileName string `json:"fileName"` // 文件名
	OldSize  int64  `json:"oldSize"`  // 原始大小
	NewSize  int64  `json:"newSize"`  // 压缩后大小
	Ratio    string `json:"ratio"`    // 压缩率 "32.4%"
	Status   string `json:"status"`   // "done" / "error"
	Message  string `json:"message"`  // 错误时填写
	Time     string `json:"time"`     // "14:32:05"
}

// TaskProgressEvent 进度事件
type TaskProgressEvent struct {
	Total  int `json:"total"`  // 图片总数
	Done   int `json:"done"`   // 已完成
	Failed int `json:"failed"` // 失败数
}

// Emitter 事件发射器
type Emitter struct {
	ctx    context.Context
	logger *zap.Logger
}

// NewEmitter 创建新的事件发射器
func NewEmitter(ctx context.Context, logger *zap.Logger) *Emitter {
	return &Emitter{
		ctx:    ctx,
		logger: logger,
	}
}

// EmitInit 发送任务初始化事件
func (e *Emitter) EmitInit(total int) {
	event := TaskInitEvent{Total: total}
	runtime.EventsEmit(e.ctx, EventTaskInit, event)

	e.logger.Info("task init event emitted", zap.Int("total", total))
}

// EmitLog 发送单条日志事件
func (e *Emitter) EmitLog(entry *LogEntry) {
	runtime.EventsEmit(e.ctx, EventTaskLog, entry)

	e.logger.Debug("log event emitted",
		zap.Int("index", entry.Index),
		zap.String("fileName", entry.FileName),
		zap.String("status", entry.Status))
}

// EmitProgress 发送进度事件
func (e *Emitter) EmitProgress(total, done, failed int) {
	event := TaskProgressEvent{
		Total:  total,
		Done:   done,
		Failed: failed,
	}
	runtime.EventsEmit(e.ctx, EventTaskProgress, event)
}

// EmitCompleted 发送任务完成事件
func (e *Emitter) EmitCompleted(total, done, failed int) {
	event := TaskProgressEvent{
		Total:  total,
		Done:   done,
		Failed: failed,
	}
	runtime.EventsEmit(e.ctx, EventTaskCompleted, event)

	e.logger.Info("task completed event emitted",
		zap.Int("done", done),
		zap.Int("failed", failed))
}

// EmitCancelled 发送任务取消事件
func (e *Emitter) EmitCancelled() {
	runtime.EventsEmit(e.ctx, EventTaskCancelled, nil)
	e.logger.Info("task cancelled event emitted")
}

// EmitPaused 发送任务暂停事件
func (e *Emitter) EmitPaused() {
	runtime.EventsEmit(e.ctx, EventTaskPaused, nil)
	e.logger.Info("task paused event emitted")
}

// EmitResumed 发送任务继续事件
func (e *Emitter) EmitResumed() {
	runtime.EventsEmit(e.ctx, EventTaskResumed, nil)
	e.logger.Info("task resumed event emitted")
}

// FormatRatio 格式化压缩率
func FormatRatio(oldSize, newSize int64) string {
	if oldSize == 0 {
		return "0%"
	}
	ratio := float64(oldSize-newSize) / float64(oldSize) * 100
	return formatPercent(ratio)
}

// FormatTime 格式化时间
func FormatTime() string {
	return time.Now().Format("15:04:05")
}

func formatPercent(value float64) string {
	if value >= 0 {
		return formatFloat(value) + "%"
	}
	return "-" + formatFloat(-value) + "%"
}

func formatFloat(value float64) string {
	s := fmt.Sprintf("%.1f", value)
	return trimZeroes(s)
}

func trimZeroes(s string) string {
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}