package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yezihack/PicSlim/internal/dto"
)

// SupportedFormats 支持的图片格式
var SupportedFormats = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// Scanner 文件扫描器
type Scanner struct {
	supportedExts map[string]bool
}

// New 创建新的扫描器
func New() *Scanner {
	return &Scanner{
		supportedExts: SupportedFormats,
	}
}

// ScanPaths 扫描输入路径
func (s *Scanner) ScanPaths(inputPaths []string) (*dto.ScanResult, error) {
	result := &dto.ScanResult{
		Code:           0,
		Message:        "success",
		SupportedFiles: make([]dto.FileInfo, 0),
		FilteredFiles:  make([]dto.FilterInfo, 0),
	}

	for _, path := range inputPaths {
		info, err := os.Stat(path)
		if err != nil {
			result.FilteredFiles = append(result.FilteredFiles, dto.FilterInfo{
				Path:   path,
				Reason: fmt.Sprintf("cannot access: %v", err),
			})
			continue
		}

		if info.IsDir() {
			s.scanDir(path, result)
		} else {
			s.scanFile(path, result)
		}
	}

	result.TotalFiles = len(result.SupportedFiles)
	// Calculate total bytes
	var totalBytes int64
	for _, f := range result.SupportedFiles {
		totalBytes += f.Size
	}
	result.TotalBytes = totalBytes
	return result, nil
}

// scanDir 递归扫描目录
func (s *Scanner) scanDir(dirPath string, result *dto.ScanResult) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.FilteredFiles = append(result.FilteredFiles, dto.FilterInfo{
				Path:   path,
				Reason: fmt.Sprintf("walk error: %v", err),
			})
			return nil
		}

		if info.IsDir() {
			return nil
		}

		s.scanFile(path, result)
		return nil
	})

	if err != nil {
		result.FilteredFiles = append(result.FilteredFiles, dto.FilterInfo{
			Path:   dirPath,
			Reason: fmt.Sprintf("directory scan error: %v", err),
		})
	}
}

// scanFile 扫描单个文件
func (s *Scanner) scanFile(filePath string, result *dto.ScanResult) {
	ext := strings.ToLower(filepath.Ext(filePath))

	if !s.supportedExts[ext] {
		result.FilteredFiles = append(result.FilteredFiles, dto.FilterInfo{
			Path:   filePath,
			Reason: fmt.Sprintf("unsupported format: %s", ext),
		})
		return
	}

	info, err := os.Stat(filePath)
	if err != nil {
		result.FilteredFiles = append(result.FilteredFiles, dto.FilterInfo{
			Path:   filePath,
			Reason: fmt.Sprintf("cannot stat file: %v", err),
		})
		return
	}

	format := s.getFormat(ext)
	result.SupportedFiles = append(result.SupportedFiles, dto.FileInfo{
		Path:    filePath,
		Name:    filepath.Base(filePath),
		Format:  format,
		Size:    info.Size(),
		ModTime: info.ModTime().Format(time.RFC3339),
	})
}

// getFormat 根据扩展名获取格式
func (s *Scanner) getFormat(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "jpg"
	case ".png":
		return "png"
	case ".webp":
		return "webp"
	default:
		return strings.TrimPrefix(ext, ".")
	}
}

// GetFileCount 获取支持的文件数量
func (s *Scanner) GetFileCount(inputPaths []string) (int, error) {
	result, err := s.ScanPaths(inputPaths)
	if err != nil {
		return 0, err
	}
	return result.TotalFiles, nil
}

// EstimateSaving 估算压缩节省空间
func (s *Scanner) EstimateSaving(files []dto.FileInfo, preset dto.Preset) (int64, float64) {
	var totalBytes int64
	for _, f := range files {
		totalBytes += f.Size
	}

	// 根据预设估算节省比例
	var estimateRatio float64
	switch preset {
	case dto.PresetHighQuality:
		estimateRatio = 0.45
	case dto.PresetBalanced:
		estimateRatio = 0.55
	case dto.PresetHighCompression:
		estimateRatio = 0.65
	default:
		estimateRatio = 0.50
	}

	saved := int64(float64(totalBytes) * estimateRatio)
	return saved, estimateRatio * 100
}