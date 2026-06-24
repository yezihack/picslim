package preview

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yezihack/PicSlim/internal/dto"

	"github.com/disintegration/imaging"
	"go.uber.org/zap"
)

// PreviewSize 预览图尺寸限制
const (
	MaxPreviewWidth  = 1920
	MaxPreviewHeight = 1080
	ThumbnailSize    = 300
)

// Previewer 预览管理器
type Previewer struct {
	logger      *zap.Logger
	cacheDir    string
	cacheMu     sync.RWMutex
	cache       map[string]string // jobID -> base64 cache
	enableCache bool
}

// New 创建新的预览管理器
func New(logger *zap.Logger, cacheDir string) *Previewer {
	if cacheDir == "" {
		cacheDir = os.TempDir()
	}

	p := &Previewer{
		logger:      logger,
		cacheDir:    cacheDir,
		cache:       make(map[string]string),
		enableCache: true,
	}

	// 创建缓存目录
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		logger.Warn("failed to create cache directory", zap.Error(err))
		p.enableCache = false
	}

	return p
}

// PreviewPair 预览对
type PreviewPair struct {
	JobID        string `json:"jobId"`
	SourcePath   string `json:"sourcePath"`
	TargetPath   string `json:"targetPath"`
	SourceSize   int64  `json:"sourceSize"`
	TargetSize   int64  `json:"targetSize"`
	SourceBase64 string `json:"sourceBase64"`
	TargetBase64 string `json:"targetBase64"`
	SourceFormat string `json:"sourceFormat"`
	TargetFormat string `json:"targetFormat"`
	HasError     bool   `json:"hasError"`
	ErrorMessage string `json:"errorMessage"`
}

// GetPreviewPair 获取预览对
func (p *Previewer) GetPreviewPair(ctx context.Context, job *dto.FileJob) (*PreviewPair, error) {
	pair := &PreviewPair{
		JobID:      job.JobID,
		SourcePath: job.SourcePath,
		TargetPath: job.TargetPath,
		SourceSize: job.BytesBefore,
		TargetSize: job.BytesAfter,
	}

	// 检查是否有目标文件
	if job.TargetPath == "" || job.Status == dto.FileJobStatusFailed {
		pair.HasError = true
		pair.ErrorMessage = "该文件压缩失败，无法预览"
		return pair, nil
	}

	// 加载源图片
	sourceBase64, sourceFormat, err := p.loadImageAsBase64(job.SourcePath, MaxPreviewWidth, MaxPreviewHeight)
	if err != nil {
		p.logger.Warn("failed to load source image",
			zap.String("path", job.SourcePath),
			zap.Error(err))
		pair.HasError = true
		pair.ErrorMessage = fmt.Sprintf("无法加载源图片: %v", err)
		return pair, nil
	}
	pair.SourceBase64 = sourceBase64
	pair.SourceFormat = sourceFormat

	// 加载目标图片
	targetBase64, targetFormat, err := p.loadImageAsBase64(job.TargetPath, MaxPreviewWidth, MaxPreviewHeight)
	if err != nil {
		p.logger.Warn("failed to load target image",
			zap.String("path", job.TargetPath),
			zap.Error(err))
		// 目标加载失败时仍返回源图片
		pair.HasError = true
		pair.ErrorMessage = fmt.Sprintf("无法加载压缩后图片: %v", err)
		return pair, nil
	}
	pair.TargetBase64 = targetBase64
	pair.TargetFormat = targetFormat

	return pair, nil
}

// GetPreviewPairByJobID 根据任务ID和JobID获取预览对
func (p *Previewer) GetPreviewPairByJobID(ctx context.Context, jobs []*dto.FileJob, jobID string) (*PreviewPair, error) {
	for _, job := range jobs {
		if job.JobID == jobID {
			return p.GetPreviewPair(ctx, job)
		}
	}
	return nil, fmt.Errorf("job not found: %s", jobID)
}

// loadImageAsBase64 加载图片并转换为 Base64
func (p *Previewer) loadImageAsBase64(path string, maxWidth, maxHeight int) (string, string, error) {
	// 检查缓存
	if p.enableCache {
		p.cacheMu.RLock()
		if cached, ok := p.cache[path]; ok {
			p.cacheMu.RUnlock()
			return cached, "", nil
		}
		p.cacheMu.RUnlock()
	}

	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		return "", "", fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	// 解码图片
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
		// 尝试通用解码
		file.Seek(0, 0)
		img, err = imaging.Decode(file)
	}

	if err != nil {
		return "", "", fmt.Errorf("cannot decode image: %w", err)
	}

	// 记录原始格式
	format := strings.TrimPrefix(ext, ".")
	if format == "jpeg" {
		format = "jpg"
	}

	// 缩放图片
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	if imgWidth > maxWidth || imgHeight > maxHeight {
		img = imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)
		p.logger.Debug("resized image for preview",
			zap.String("path", path),
			zap.Int("origWidth", imgWidth),
			zap.Int("origHeight", imgHeight),
			zap.Int("newWidth", img.Bounds().Dx()),
			zap.Int("newHeight", img.Bounds().Dy()))
	}

	// 编码为 JPEG (用于预览，统一格式)
	var buf strings.Builder
	buf.WriteString("data:image/jpeg;base64,")

	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	if err := jpeg.Encode(encoder, img, &jpeg.Options{Quality: 85}); err != nil {
		return "", "", fmt.Errorf("cannot encode image: %w", err)
	}
	encoder.Close()

	result := buf.String()

	// 缓存结果
	if p.enableCache {
		p.cacheMu.Lock()
		p.cache[path] = result
		p.cacheMu.Unlock()
	}

	return result, format, nil
}

// GenerateThumbnail 生成缩略图
func (p *Previewer) GenerateThumbnail(sourcePath, cacheKey string) (string, error) {
	// 打开源文件
	file, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	// 解码图片
	var img image.Image
	ext := strings.ToLower(filepath.Ext(sourcePath))

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
		return "", fmt.Errorf("cannot decode image: %w", err)
	}

	// 生成缩略图
	thumbnail := imaging.Thumbnail(img, ThumbnailSize, ThumbnailSize, imaging.Lanczos)

	// 保存到缓存目录
	thumbnailPath := filepath.Join(p.cacheDir, cacheKey+".jpg")
	outFile, err := os.Create(thumbnailPath)
	if err != nil {
		return "", fmt.Errorf("cannot create thumbnail file: %w", err)
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, thumbnail, &jpeg.Options{Quality: 80}); err != nil {
		return "", fmt.Errorf("cannot encode thumbnail: %w", err)
	}

	return thumbnailPath, nil
}

// ClearCache 清除缓存
func (p *Previewer) ClearCache() {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	p.cache = make(map[string]string)

	// 清除缓存目录
	if p.enableCache && p.cacheDir != "" {
		files, err := os.ReadDir(p.cacheDir)
		if err != nil {
			p.logger.Warn("failed to read cache directory", zap.Error(err))
			return
		}

		for _, file := range files {
			if !file.IsDir() {
				path := filepath.Join(p.cacheDir, file.Name())
				if err := os.Remove(path); err != nil {
					p.logger.Warn("failed to remove cache file",
						zap.String("path", path),
						zap.Error(err))
				}
			}
		}
	}

	p.logger.Info("preview cache cleared")
}

// GetCacheStats 获取缓存统计
func (p *Previewer) GetCacheStats() map[string]interface{} {
	p.cacheMu.RLock()
	defer p.cacheMu.RUnlock()

	return map[string]interface{}{
		"cachedItems":  len(p.cache),
		"enableCache":  p.enableCache,
		"cacheDir":     p.cacheDir,
	}
}

// BatchGetPreviewPairs 批量获取预览对
func (p *Previewer) BatchGetPreviewPairs(ctx context.Context, jobs []*dto.FileJob, limit int) ([]*PreviewPair, error) {
	if limit <= 0 || limit > len(jobs) {
		limit = len(jobs)
	}

	results := make([]*PreviewPair, 0, limit)

	for i := 0; i < limit; i++ {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
			pair, err := p.GetPreviewPair(ctx, jobs[i])
			if err != nil {
				p.logger.Warn("failed to get preview pair",
					zap.String("jobId", jobs[i].JobID),
					zap.Error(err))
				continue
			}
			results = append(results, pair)
		}
	}

	return results, nil
}

// NavigationHelper 预览导航辅助
type NavigationHelper struct {
	jobs     []*dto.FileJob
	current  int
	previewer *Previewer
}

// NewNavigationHelper 创建导航辅助
func NewNavigationHelper(jobs []*dto.FileJob, previewer *Previewer) *NavigationHelper {
	return &NavigationHelper{
		jobs:      jobs,
		current:   0,
		previewer: previewer,
	}
}

// GetCurrent 获取当前预览
func (n *NavigationHelper) GetCurrent(ctx context.Context) (*PreviewPair, error) {
	if len(n.jobs) == 0 {
		return nil, fmt.Errorf("no jobs available")
	}
	return n.previewer.GetPreviewPair(ctx, n.jobs[n.current])
}

// GetNext 获取下一个预览
func (n *NavigationHelper) GetNext(ctx context.Context) (*PreviewPair, error) {
	if len(n.jobs) == 0 {
		return nil, fmt.Errorf("no jobs available")
	}

	n.current++
	if n.current >= len(n.jobs) {
		n.current = 0 // 循环
	}

	return n.previewer.GetPreviewPair(ctx, n.jobs[n.current])
}

// GetPrevious 获取上一个预览
func (n *NavigationHelper) GetPrevious(ctx context.Context) (*PreviewPair, error) {
	if len(n.jobs) == 0 {
		return nil, fmt.Errorf("no jobs available")
	}

	n.current--
	if n.current < 0 {
		n.current = len(n.jobs) - 1 // 循环
	}

	return n.previewer.GetPreviewPair(ctx, n.jobs[n.current])
}

// GetAtIndex 获取指定索引的预览
func (n *NavigationHelper) GetAtIndex(ctx context.Context, index int) (*PreviewPair, error) {
	if index < 0 || index >= len(n.jobs) {
		return nil, fmt.Errorf("index out of range: %d", index)
	}
	n.current = index
	return n.previewer.GetPreviewPair(ctx, n.jobs[n.current])
}

// GetCurrentIndex 获取当前索引
func (n *NavigationHelper) GetCurrentIndex() int {
	return n.current
}

// GetTotal 获取总数
func (n *NavigationHelper) GetTotal() int {
	return len(n.jobs)
}