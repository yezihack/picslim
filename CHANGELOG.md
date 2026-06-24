# 更新日志

本项目的所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [0.4.1] - 2026-03-27

### 修复

- 移除前端模拟数据模式，确保应用只在 Wails 环境下运行真实功能
- 修复 `isCompleted` computed 属性被错误赋值的问题
- 更新 wails.json 配置使用 pnpm 替代 npm

### 功能验证

- ✅ 后端 Go 代码编译成功
- ✅ 前端 Vue 构建成功
- ✅ Wails 应用打包成功
- ✅ 选择输入目录、扫描图片功能可用
- ✅ 选择输出目录功能可用
- ✅ 创建压缩任务、启动/暂停/继续/取消功能可用
- ✅ 实时进度更新通过事件推送

## [0.4.0] - 2026-03-27

### 新增 - Phase 3: 预览与体验完善

#### 前端组件

- **PreviewSection** (`frontend/src/components/PreviewSection.vue`)：预览对比组件
  - 压缩前后图片并排展示
  - 显示文件大小和节省比例
  - 上一张/下一张导航功能
  - 预览计数显示（当前/总数）
  - 加载中/错误/空态展示
  - 文件信息展示（文件名、格式、耗时）

- **FileTable** (`frontend/src/components/FileTable.vue`)：文件明细表组件
  - 分页展示文件列表
  - 按状态过滤（成功/失败/无收益）
  - 文件名搜索功能
  - 状态颜色标记
  - 点击行触发预览

- **ProgressPanel** (`frontend/src/components/ProgressPanel.vue`)：进度面板组件
  - 任务状态标签（进行中/已暂停/已完成）
  - 并发数、平均速度、预计完成时间展示
  - 总进度条和当前文件进度条
  - 已运行时间、剩余时间展示
  - 完成后显示统计卡片

#### App.vue 重构

- 模块化组件结构，分离关注点
- 左侧面板：输入输出配置、预设选择、任务控制
- 右侧面板：实时进度展示
- 预览区：压缩前后对比、导航控制
- 文件明细：分页、搜索、过滤

#### 体验细节

- 任务状态实时更新
- 空态提示（等待选择输入、暂无预览）
- 错误提示友好
- 加载动画
- 响应式布局适配

### 验收标准

- ✅ 预览切换流畅，无卡顿
- ✅ 大图/异常图有兜底提示
- ✅ Demo 视觉元素保留

## [0.3.0] - 2026-03-27

### 新增 - Phase 2: 结果、统计与报告

#### 后端模块

- **report** (`internal/report`)：统计汇总与报告导出
  - 聚合统计：总文件数、成功数、失败数、字节统计、节省比例
  - CSV 报告导出：支持 UTF-8 BOM，Excel 兼容
  - 详细报告：包含序号、文件名、路径、格式、字节大小、节省比例、状态、重试次数、错误信息、耗时
  - 摘要报告：总体统计指标
  - 文件过滤：按状态、格式、大小、文件名搜索过滤
  - 分页支持：支持大数据集分页读取
  - 格式分组统计：按图片格式分组统计压缩效果

- **preview** (`internal/preview`)：预览对比能力
  - 预览对生成：加载源图片与目标图片，转换为 Base64
  - 尺寸自适应：自动缩放到最大 1920x1080，避免大图卡顿
  - 缩略图生成：生成 300x300 缩略图用于快速预览
  - 缓存机制：内存缓存已加载的图片，提升响应速度
  - 导航辅助：支持上一张/下一张/指定索引预览
  - 批量预览：支持批量获取预览对

#### App 模块增强

- 新增 `ListFailedFiles`：获取失败文件列表
- 新增 `GetPreviewPair`：根据 JobID 获取预览对
- 新增 `GetFirstPreview`：获取第一个预览
- 新增 `GetNextPreview`：获取下一个预览
- 新增 `GetPreviousPreview`：获取上一个预览
- 新增 `GetPreviewIndex`：获取当前预览索引
- 新增 `GetPreviewTotal`：获取预览总数
- 新增 `GetFormatStats`：获取格式分组统计
- 增强 `ExportReport`：使用 reporter 模块导出完整 CSV 报告

#### 前端增强

- **类型定义**：新增 `PreviewPairResult`、`FormatStats` 类型
- **API 服务**：新增所有 Phase 2 API 方法
  - `getPreviewPair`、`getFirstPreview`、`getNextPreview`、`getPreviousPreview`
  - `getPreviewIndex`、`getPreviewTotal`
  - `listFailedFiles`、`getFormatStats`

### 验收标准

- ✅ 统计值与文件明细可对账
- ✅ CSV 报告字段完整且可在 Excel 打开
- ✅ 输出目录可一键打开
- ✅ 预览对接口可用

## [0.2.0] - 2026-03-27

### 新增 - Phase 1: 任务核心能力

#### 后端模块

- **scanner** (`internal/scanner`)：输入路径扫描、格式过滤、文件总数统计
  - 支持递归目录扫描
  - 支持格式过滤：jpg、jpeg、png、webp
  - 返回文件信息：路径、文件名、格式、大小、修改时间

- **scheduler** (`internal/scheduler`)：任务状态机、worker pool、任务控制
  - 任务状态：PENDING、RUNNING、PAUSED、COMPLETED、FAILED、CANCELLED
  - Worker pool 支持可配置并发数（默认：CPU 核心数 - 1）
  - 任务控制：开始、暂停、继续、取消
  - 进度监控与 ETA 计算

- **compressor** (`internal/compressor`)：按格式压缩图片
  - JPEG 压缩，可配置质量参数
  - PNG 优化
  - WebP 压缩
  - 预设支持：高画质、均衡、高压缩
  - 尺寸调整支持（最大宽高限制）
  - 收益守护：压缩后更大时保留原文件

- **retry** (`internal/retry`)：自动重试与指数退避
  - 默认重试 3 次，退避时间 1s、2s、4s
  - 可重试错误码：E_DECODE、E_ENCODE、E_WRITE_OUTPUT、E_TEMP_FILE

- **events** (`internal/events`)：通过 Wails 运行时推送事件
  - 事件类型：task:progress、task:state_changed、task:file_done、task:error、task:completed

- **dto** (`internal/dto`)：完整的数据传输对象定义
  - TaskConfig、Task、FileJob、ProgressSnapshot、ResultSummary
  - 所有 API 调用的请求/响应类型

#### 前端集成

- **类型定义** (`frontend/src/types`)：所有 DTO 的 TypeScript 接口
- **Wails 服务** (`frontend/src/services/wails.ts`)：API 封装，支持开发模式 Mock
- **App.vue**：更新为使用真实后端 API
  - 输入/输出路径选择
  - 预设选择
  - 任务控制按钮（开始、暂停、继续、取消）
  - 实时进度显示
  - 事件订阅获取进度更新

#### 依赖项

- `github.com/disintegration/imaging v1.6.2`：图片处理
- `github.com/google/uuid`：任务 ID 生成
- `go.uber.org/zap`：结构化日志
- `github.com/gin-gonic/gin`：HTTP 服务器（可选调试接口）
- `github.com/wailsapp/wails/v2`：桌面应用框架

### 修改

- 更新 `main.go` 嵌入 `frontend/dist` 而非 `frontend/public`
- 修复 `App.go` 中 emitter 初始化

## [0.1.0] - 2026-03-27

### 新增 - Phase 0: 工程脚手架

- Wails 项目初始化，使用 Vue3 + Element Plus 前端
- 后端模块基础目录结构
- 初始 DTO 定义
- 日志模块 (`internal/logx`)
- 配置模块 (`internal/config`)
- HTTP API 服务器脚手架 (`internal/httpapi`)
- Demo UI 布局对齐设计规范
