package compressor

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yezihack/PicSlim/internal/dto"

	"github.com/disintegration/imaging"
	"go.uber.org/zap"
)

// PresetConfig 预设配置
type PresetConfig struct {
	JPEGQuality int
	WebPQuality int
	PNGLevel    int
}

// 预设配置映射
var presetConfigs = map[dto.Preset]PresetConfig{
	dto.PresetHighQuality: {
		JPEGQuality: 90,
		WebPQuality: 88,
		PNGLevel:    4,
	},
	dto.PresetBalanced: {
		JPEGQuality: 82,
		WebPQuality: 80,
		PNGLevel:    6,
	},
	dto.PresetHighCompression: {
		JPEGQuality: 75,
		WebPQuality: 72,
		PNGLevel:    8,
	},
}

// Compressor 图片压缩器
type Compressor struct {
	logger       *zap.Logger
	preset       dto.Preset
	presetConfig PresetConfig
	outputDir    string
	keepStruct   bool
	maxWidth     int
	maxHeight    int
}

// New 创建新的压缩器
func New(logger *zap.Logger, preset dto.Preset, outputDir string, keepStruct bool, maxWidth, maxHeight int) *Compressor {
	config, ok := presetConfigs[preset]
	if !ok {
		config = presetConfigs[dto.PresetHighQuality]
	}

	return &Compressor{
		logger:       logger,
		preset:       preset,
		presetConfig: config,
		outputDir:    outputDir,
		keepStruct:   keepStruct,
		maxWidth:     maxWidth,
		maxHeight:    maxHeight,
	}
}

// Process 处理单个文件
func (c *Compressor) Process(ctx context.Context, job *dto.FileJob) (*dto.FileJob, error) {
	startTime := time.Now()

	// 更新状态
	job.Status = dto.FileJobStatusRunning
	job.Attempt++

	// 读取源文件
	srcFile, err := os.Open(job.SourcePath)
	if err != nil {
		job.Status = dto.FileJobStatusFailed
		job.ErrorCode = "E_READ_SOURCE"
		job.ErrorMessage = fmt.Sprintf("cannot open source file: %v", err)
		return job, err
	}
	defer srcFile.Close()

	// 获取文件大小
	srcInfo, err := srcFile.Stat()
	if err != nil {
		job.Status = dto.FileJobStatusFailed
		job.ErrorCode = "E_STAT_SOURCE"
		job.ErrorMessage = fmt.Sprintf("cannot stat source file: %v", err)
		return job, err
	}
	job.BytesBefore = srcInfo.Size()

	// 解码图片
	var img image.Image
	switch job.Format {
	case "jpg", "jpeg":
		img, err = jpeg.Decode(srcFile)
	case "png":
		img, err = png.Decode(srcFile)
	case "webp":
		img, err = imaging.Decode(srcFile)
	default:
		err = fmt.Errorf("unsupported format: %s", job.Format)
	}

	if err != nil {
		job.Status = dto.FileJobStatusFailed
		job.ErrorCode = "E_DECODE"
		job.ErrorMessage = fmt.Sprintf("cannot decode image: %v", err)
		return job, err
	}

	// 尺寸调整
	if c.maxWidth > 0 || c.maxHeight > 0 {
		img = c.resizeImage(img)
	}

	// 确定输出路径
	targetPath, err := c.getTargetPath(job.SourcePath)
	if err != nil {
		job.Status = dto.FileJobStatusFailed
		job.ErrorCode = "E_PATH_MAPPING"
		job.ErrorMessage = fmt.Sprintf("cannot determine target path: %v", err)
		return job, err
	}
	job.TargetPath = targetPath

	// 创建输出目录
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		job.Status = dto.FileJobStatusFailed
		job.ErrorCode = "E_MKDIR"
		job.ErrorMessage = fmt.Sprintf("cannot create output directory: %v", err)
		return job, err
	}

	// 编码并保存
	if err := c.encodeAndSave(img, targetPath, job.Format); err != nil {
		job.Status = dto.FileJobStatusFailed
		job.ErrorCode = "E_ENCODE"
		job.ErrorMessage = fmt.Sprintf("cannot encode image: %v", err)
		return job, err
	}

	// 检查压缩效果
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		job.Status = dto.FileJobStatusFailed
		job.ErrorCode = "E_STAT_TARGET"
		job.ErrorMessage = fmt.Sprintf("cannot stat target file: %v", err)
		return job, err
	}
	job.BytesAfter = targetInfo.Size()

	// 收益守护：如果压缩后更大，保留原文件
	if job.BytesAfter >= job.BytesBefore {
		c.logger.Info("no compression gain, keeping original",
			zap.String("source", job.SourcePath),
			zap.Int64("before", job.BytesBefore),
			zap.Int64("after", job.BytesAfter))

		// 用原文件替换压缩后的文件
		if err := c.copyOriginalFile(job.SourcePath, targetPath); err != nil {
			c.logger.Warn("failed to copy original file", zap.Error(err))
		} else {
			targetInfo, _ = os.Stat(targetPath)
			job.BytesAfter = targetInfo.Size()
		}
		job.Status = dto.FileJobStatusSkippedNoGain
	} else {
		job.Status = dto.FileJobStatusSuccess
	}

	job.DurationMs = time.Since(startTime).Milliseconds()

	c.logger.Info("file compressed",
		zap.String("source", job.SourcePath),
		zap.String("target", targetPath),
		zap.Int64("before", job.BytesBefore),
		zap.Int64("after", job.BytesAfter),
		zap.String("status", string(job.Status)))

	return job, nil
}

// resizeImage 调整图片尺寸
func (c *Compressor) resizeImage(img image.Image) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	maxW := c.maxWidth
	maxH := c.maxHeight

	if maxW <= 0 {
		maxW = srcWidth
	}
	if maxH <= 0 {
		maxH = srcHeight
	}

	if srcWidth <= maxW && srcHeight <= maxH {
		return img
	}

	return imaging.Fit(img, maxW, maxH, imaging.Lanczos)
}

// getTargetPath 获取目标文件路径
func (c *Compressor) getTargetPath(sourcePath string) (string, error) {
	// 获取文件名和扩展名
	fileName := filepath.Base(sourcePath)
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, ext)

	// 处理重名
	targetName := baseName + "_compressed" + ext

	// 构建目标路径
	if c.keepStruct {
		// 保持原目录结构
		relPath, err := filepath.Rel(filepath.Dir(filepath.Dir(sourcePath)), sourcePath)
		if err != nil {
			return "", err
		}
		targetPath := filepath.Join(c.outputDir, filepath.Dir(relPath), targetName)
		return targetPath, nil
	}

	return filepath.Join(c.outputDir, targetName), nil
}

// encodeAndSave 编码并保存图片
func (c *Compressor) encodeAndSave(img image.Image, targetPath, format string) error {
	// 创建临时文件
	tmpPath := targetPath + ".tmp"

	switch format {
	case "jpg", "jpeg":
		out, err := os.Create(tmpPath)
		if err != nil {
			return err
		}

		options := &jpeg.Options{Quality: c.presetConfig.JPEGQuality}
		err = jpeg.Encode(out, img, options)
		out.Close() // 显式关闭，不用 defer
		if err != nil {
			os.Remove(tmpPath)
			return err
		}

	case "png":
		out, err := os.Create(tmpPath)
		if err != nil {
			return err
		}

		encoder := png.Encoder{CompressionLevel: png.DefaultCompression}
		err = encoder.Encode(out, img)
		out.Close() // 显式关闭，不用 defer
		if err != nil {
			os.Remove(tmpPath)
			return err
		}

	case "webp":
		if err := imaging.Save(img, tmpPath, imaging.JPEGQuality(c.presetConfig.WebPQuality)); err != nil {
			os.Remove(tmpPath)
			return err
		}

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// 原子替换（此时文件已关闭）
	if err := os.Rename(tmpPath, targetPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

// copyOriginalFile 复制原文件
func (c *Compressor) copyOriginalFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = dstFile.ReadFrom(srcFile)
	dstFile.Close() // 显式关闭
	return err
}

// ProcessToFile 处理单个文件（简化接口）
func (c *Compressor) ProcessToFile(src, dst string) error {
	// 读取源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot open source file: %w", err)
	}

	// 解码图片
	var img image.Image
	ext := strings.ToLower(filepath.Ext(src))
	format := strings.TrimPrefix(ext, ".")

	switch format {
	case "jpg", "jpeg":
		img, err = jpeg.Decode(srcFile)
	case "png":
		img, err = png.Decode(srcFile)
	case "webp":
		img, err = imaging.Decode(srcFile)
	default:
		srcFile.Close()
		return fmt.Errorf("unsupported format: %s", format)
	}

	// 关闭源文件（解码完成后立即关闭）
	srcFile.Close()

	if err != nil {
		return fmt.Errorf("cannot decode image: %w", err)
	}

	// 尺寸调整
	if c.maxWidth > 0 || c.maxHeight > 0 {
		img = c.resizeImage(img)
	}

	// 创建输出目录
	targetDir := filepath.Dir(dst)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	// 编码并保存
	if err := c.encodeAndSave(img, dst, format); err != nil {
		return fmt.Errorf("cannot encode image: %w", err)
	}

	return nil
}