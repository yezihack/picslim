# 智能图片压缩器 (Desktop Image Compressor)

一款基于 Wails 构建的跨平台桌面图片压缩工具，支持 Windows、Linux 和 macOS，兼容 AMD64 和 ARM64 架构。

## 界面预览

![20260624200119](https://cdn.jsdelivr.net/gh/yezihack/assets/b/20260624200119.png)

## 功能特性

- **多格式支持**：JPEG、PNG、WebP
- **智能压缩**：自动选择最佳压缩策略，收益守护（压缩后更大时保留原文件）
- **预设模式**：高画质、均衡、高压缩三种预设
- **批量处理**：支持递归扫描目录，批量压缩图片
- **实时进度**：任务进度、ETA 计算、实时状态更新
- **预览对比**：压缩前后图片并排展示，文件大小对比
- **任务控制**：开始、暂停、继续、取消任务
- **报告导出**：CSV 报告导出，Excel 兼容
- **统计汇总**：总文件数、成功数、失败数、节省比例等统计指标

## 系统要求

### Windows

- Windows 10/11
- WebView2 运行时（Windows 11 已内置）

### Linux

- GTK 3
- WebKit2GTK
- libsoup 2.4

安装依赖：

```bash
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev libappindicator3-dev libnspr4-dev libnss3-dev libasound2-dev
```

### macOS

- macOS 10.15+
- Xcode Command Line Tools

## 下载安装

从 [Releases](https://github.com/yezihack/PicSlim/releases) 页面下载对应平台的安装包：

| 平台 | 架构 | 下载文件 |
|------|------|----------|
| Windows | x64 | `desktop-imagecompressor-windows-amd64.zip` |
| Windows | ARM64 | `desktop-imagecompressor-windows-arm64.zip` |
| Linux | x64 | `desktop-imagecompressor-linux-amd64.tar.gz` |
| Linux | ARM64 | `desktop-imagecompressor-linux-arm64.tar.gz` |
| macOS | Intel | `desktop-imagecompressor-darwin-amd64.zip` |
| macOS | Apple Silicon | `desktop-imagecompressor-darwin-arm64.zip` |

## 开发

### 前置要求

- Go 1.22+
- Node.js 20+
- pnpm 9+
- Wails CLI

安装 Wails CLI：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 本地开发

```bash
# 克隆项目
git clone https://github.com/yezihack/PicSlim.git
cd desktop-imagecompressor

# 安装前端依赖
cd frontend
pnpm install

# 开发模式运行
wails dev
```

### 构建

```bash
# 构建当前平台
wails build

# 构建指定平台
wails build -platform windows/amd64
wails build -platform windows/arm64
wails build -platform linux/amd64
wails build -platform linux/arm64
wails build -platform darwin/amd64
wails build -platform darwin/arm64
```

## 项目结构

```
desktop-imagecompressor/
├── frontend/                 # Vue3 前端
│   ├── src/
│   │   ├── components/       # Vue 组件
│   │   ├── services/         # Wails API 封装
│   │   ├── types/            # TypeScript 类型定义
│   │   └── App.vue           # 主应用组件
│   └── package.json
├── internal/                 # Go 后端模块
│   ├── app/                  # Wails 应用绑定
│   ├── compressor/           # 图片压缩核心
│   ├── config/               # 配置管理
│   ├── dto/                  # 数据传输对象
│   ├── events/               # 事件推送
│   ├── httpapi/              # HTTP API 服务器
│   ├── logx/                 # 日志模块
│   ├── preview/              # 预览对比
│   ├── report/               # 报告导出
│   ├── retry/                # 重试机制
│   ├── scanner/              # 文件扫描
│   ├── scheduler/            # 任务调度
│   └── task/                 # 任务管理
├── main.go                   # 应用入口
├── wails.json                # Wails 配置
└── go.mod                    # Go 模块定义
```

## 技术栈

- **后端**：Go 1.22, Wails v2, imaging (图片处理), ants (goroutine pool), zap (日志)
- **前端**：Vue 3, TypeScript, Element Plus, Vite
- **框架**：Wails (Go + WebView)

## 配置说明

应用支持以下配置项（环境变量）：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `LOG_LEVEL` | `info` | 日志级别 |
| `HTTP_PORT` | `8080` | HTTP API 端口 |
| `ENABLE_LOCAL_HTTP` | `false` | 是否启用 HTTP API |

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
