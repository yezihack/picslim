package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yezihack/PicSlim/internal/dto"

	"go.uber.org/zap"
)

// Reporter 报告生成器
type Reporter struct {
	logger *zap.Logger
}

// New 创建新的报告生成器
func New(logger *zap.Logger) *Reporter {
	return &Reporter{logger: logger}
}

// AggregateStats 聚合统计结果
func (r *Reporter) AggregateStats(jobs []*dto.FileJob) *dto.ResultSummary {
	summary := &dto.ResultSummary{
		TotalFiles: len(jobs),
	}

	for _, job := range jobs {
		summary.BytesBefore += job.BytesBefore

		switch job.Status {
		case dto.FileJobStatusSuccess:
			summary.SuccessFiles++
			summary.BytesAfter += job.BytesAfter
		case dto.FileJobStatusSkippedNoGain:
			summary.SuccessFiles++
			summary.BytesAfter += job.BytesAfter
		case dto.FileJobStatusFailed:
			summary.FailedFiles++
		}
	}

	summary.BytesSaved = summary.BytesBefore - summary.BytesAfter
	if summary.BytesBefore > 0 {
		summary.SavedPercent = float64(summary.BytesSaved) / float64(summary.BytesBefore) * 100
	}

	return summary
}

// FormatStats 格式化统计结果用于显示
type FormattedStats struct {
	TotalFiles     int
	SuccessFiles   int
	FailedFiles    int
	BytesBefore    string
	BytesAfter     string
	BytesSaved     string
	SavedPercent   string
	AvgFileSize    string
	AvgCompression string
	AvgDuration    string
}

// FormatSummary 格式化结果摘要
func (r *Reporter) FormatSummary(summary *dto.ResultSummary, jobs []*dto.FileJob) *FormattedStats {
	stats := &FormattedStats{
		TotalFiles:   summary.TotalFiles,
		SuccessFiles: summary.SuccessFiles,
		FailedFiles:  summary.FailedFiles,
		BytesBefore:  formatBytes(summary.BytesBefore),
		BytesAfter:   formatBytes(summary.BytesAfter),
		BytesSaved:   formatBytes(summary.BytesSaved),
		SavedPercent: fmt.Sprintf("%.1f%%", summary.SavedPercent),
	}

	// 计算平均值
	if len(jobs) > 0 {
		var totalDuration int64
		var totalCompression float64
		validCount := 0

		for _, job := range jobs {
			if job.Status == dto.FileJobStatusSuccess || job.Status == dto.FileJobStatusSkippedNoGain {
				totalDuration += job.DurationMs
				if job.BytesBefore > 0 {
					totalCompression += float64(job.BytesBefore-job.BytesAfter) / float64(job.BytesBefore) * 100
				}
				validCount++
			}
		}

		if validCount > 0 {
			stats.AvgFileSize = formatBytes(summary.BytesBefore / int64(len(jobs)))
			stats.AvgCompression = fmt.Sprintf("%.1f%%", totalCompression/float64(validCount))
			stats.AvgDuration = formatDuration(totalDuration / int64(validCount))
		}
	}

	return stats
}

// ExportCSV 导出 CSV 报告
func (r *Reporter) ExportCSV(jobs []*dto.FileJob, outputPath string) error {
	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create CSV file: %w", err)
	}
	defer file.Close()

	// 写入 BOM 以支持 Excel 正确识别 UTF-8
	if _, err := file.WriteString("\xEF\xBB\xBF"); err != nil {
		return fmt.Errorf("cannot write BOM: %w", err)
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := []string{
		"序号",
		"文件名",
		"源路径",
		"目标路径",
		"格式",
		"压缩前(字节)",
		"压缩前(可读)",
		"压缩后(字节)",
		"压缩后(可读)",
		"节省(字节)",
		"节省比例",
		"状态",
		"重试次数",
		"错误信息",
		"耗时(毫秒)",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("cannot write headers: %w", err)
	}

	// 按状态排序：成功在前，失败在后
	sortedJobs := make([]*dto.FileJob, len(jobs))
	copy(sortedJobs, jobs)
	sort.Slice(sortedJobs, func(i, j int) bool {
		// 成功排在前面
		if sortedJobs[i].Status != sortedJobs[j].Status {
			return sortedJobs[i].Status == dto.FileJobStatusSuccess ||
				(sortedJobs[i].Status == dto.FileJobStatusSkippedNoGain &&
					sortedJobs[j].Status == dto.FileJobStatusFailed)
		}
		return sortedJobs[i].SourcePath < sortedJobs[j].SourcePath
	})

	// 写入数据行
	for i, job := range sortedJobs {
		saved := job.BytesBefore - job.BytesAfter
		var savedPercent string
		if job.BytesBefore > 0 {
			savedPercent = fmt.Sprintf("%.2f%%", float64(saved)/float64(job.BytesBefore)*100)
		}

		statusText := "成功"
		switch job.Status {
		case dto.FileJobStatusSkippedNoGain:
			statusText = "成功(无收益)"
		case dto.FileJobStatusFailed:
			statusText = "失败"
		}

		row := []string{
			fmt.Sprintf("%d", i+1),
			filepath.Base(job.SourcePath),
			job.SourcePath,
			job.TargetPath,
			strings.ToUpper(job.Format),
			fmt.Sprintf("%d", job.BytesBefore),
			formatBytes(job.BytesBefore),
			fmt.Sprintf("%d", job.BytesAfter),
			formatBytes(job.BytesAfter),
			fmt.Sprintf("%d", saved),
			savedPercent,
			statusText,
			fmt.Sprintf("%d", job.Attempt),
			job.ErrorMessage,
			fmt.Sprintf("%d", job.DurationMs),
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("cannot write row: %w", err)
		}
	}

	r.logger.Info("CSV report exported",
		zap.String("path", outputPath),
		zap.Int("rows", len(jobs)))

	return nil
}

// ExportSummaryCSV 导出摘要 CSV
func (r *Reporter) ExportSummaryCSV(summary *dto.ResultSummary, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create summary CSV: %w", err)
	}
	defer file.Close()

	// 写入 BOM
	if _, err := file.WriteString("\xEF\xBB\xBF"); err != nil {
		return fmt.Errorf("cannot write BOM: %w", err)
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入摘要信息
	rows := [][]string{
		{"指标", "值"},
		{"总文件数", fmt.Sprintf("%d", summary.TotalFiles)},
		{"成功数", fmt.Sprintf("%d", summary.SuccessFiles)},
		{"失败数", fmt.Sprintf("%d", summary.FailedFiles)},
		{"压缩前总大小", formatBytes(summary.BytesBefore)},
		{"压缩后总大小", formatBytes(summary.BytesAfter)},
		{"节省空间", formatBytes(summary.BytesSaved)},
		{"节省比例", fmt.Sprintf("%.2f%%", summary.SavedPercent)},
		{"导出时间", time.Now().Format("2006-01-02 15:04:05")},
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("cannot write row: %w", err)
		}
	}

	return nil
}

// FilterJobs 过滤任务
type JobFilter struct {
	Status   dto.FileJobStatus // 按状态过滤
	Format   string            // 按格式过滤
	MinSize  int64             // 最小文件大小
	MaxSize  int64             // 最大文件大小
	Search   string            // 搜索文件名
	Page     int               // 页码
	PageSize int               // 每页数量
}

// FilterAndPaginate 过滤并分页任务
func (r *Reporter) FilterAndPaginate(jobs []*dto.FileJob, filter JobFilter) []*dto.FileJob {
	// 过滤
	var filtered []*dto.FileJob
	for _, job := range jobs {
		// 状态过滤
		if filter.Status != "" && job.Status != filter.Status {
			continue
		}

		// 格式过滤
		if filter.Format != "" && job.Format != filter.Format {
			continue
		}

		// 大小过滤
		if filter.MinSize > 0 && job.BytesBefore < filter.MinSize {
			continue
		}
		if filter.MaxSize > 0 && job.BytesBefore > filter.MaxSize {
			continue
		}

		// 搜索过滤
		if filter.Search != "" {
			fileName := filepath.Base(job.SourcePath)
			if !containsIgnoreCase(fileName, filter.Search) {
				continue
			}
		}

		filtered = append(filtered, job)
	}

	// 分页
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}

	start := (filter.Page - 1) * filter.PageSize
	end := start + filter.PageSize

	if start >= len(filtered) {
		return []*dto.FileJob{}
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end]
}

// GetFailedJobs 获取失败的任务
func (r *Reporter) GetFailedJobs(jobs []*dto.FileJob) []*dto.FileJob {
	var failed []*dto.FileJob
	for _, job := range jobs {
		if job.Status == dto.FileJobStatusFailed {
			failed = append(failed, job)
		}
	}
	return failed
}

// GetSuccessJobs 获取成功的任务
func (r *Reporter) GetSuccessJobs(jobs []*dto.FileJob) []*dto.FileJob {
	var success []*dto.FileJob
	for _, job := range jobs {
		if job.Status == dto.FileJobStatusSuccess || job.Status == dto.FileJobStatusSkippedNoGain {
			success = append(success, job)
		}
	}
	return success
}

// GroupByFormat 按格式分组统计
func (r *Reporter) GroupByFormat(jobs []*dto.FileJob) map[string]*dto.ResultSummary {
	result := make(map[string]*dto.ResultSummary)

	for _, job := range jobs {
		format := job.Format
		if _, exists := result[format]; !exists {
			result[format] = &dto.ResultSummary{}
		}

		result[format].TotalFiles++
		result[format].BytesBefore += job.BytesBefore

		switch job.Status {
		case dto.FileJobStatusSuccess:
			result[format].SuccessFiles++
			result[format].BytesAfter += job.BytesAfter
		case dto.FileJobStatusSkippedNoGain:
			result[format].SuccessFiles++
			result[format].BytesAfter += job.BytesAfter
		case dto.FileJobStatusFailed:
			result[format].FailedFiles++
		}
	}

	// 计算节省比例
	for _, summary := range result {
		summary.BytesSaved = summary.BytesBefore - summary.BytesAfter
		if summary.BytesBefore > 0 {
			summary.SavedPercent = float64(summary.BytesSaved) / float64(summary.BytesBefore) * 100
		}
	}

	return result
}

// 辅助函数

func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	sizes := []string{"B", "KB", "MB", "GB", "TB"}

	i := 0
	fb := float64(bytes)
	for fb >= unit && i < len(sizes)-1 {
		fb /= unit
		i++
	}

	return fmt.Sprintf("%.2f %s", fb, sizes[i])
}

func formatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%d ms", ms)
	}
	seconds := float64(ms) / 1000
	if seconds < 60 {
		return fmt.Sprintf("%.1f s", seconds)
	}
	minutes := int(seconds / 60)
	remainingSeconds := int(seconds) % 60
	return fmt.Sprintf("%d m %d s", minutes, remainingSeconds)
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}