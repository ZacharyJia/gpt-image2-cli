# GPT Image 2 CLI（Go 版）

这是一个用 Go 实现的 GPT Image 2 命令行工具，同时附带 Agent Skill 表面，支持 Codex / Claude Code / OpenClaw 等支持 Skill 的 Agent 运行时。

项目参考了 [`wuyoscar/gpt_image_2_skill`](https://github.com/wuyoscar/gpt_image_2_skill) 的 Python 版本，复用了它的提示词图库（references），而 CLI 本身使用 Go 编写，**零外部依赖**。

## 功能

- 文生图（`/v1/images/generations`）
- 参考图编辑，包括多参考图编辑（`/v1/images/edits`）
- 基于 Alpha 通道的局部重绘（`-i` + `-m`）
- 支持全部主要 OpenAI 图像参数：size、quality、background、moderation、format、compression、user
- 尺寸简写：`1k`、`2k`、`4k`、`portrait`、`landscape`、`square`、`wide`、`tall`
- 按 env、`.env`、`~/.env` 顺序读取 `OPENAI_API_KEY` / `API_KEY` 和可选的 `BASE_URL`（进程 env 优先级最高，不会被覆盖）
- 自动生成输出文件名；若存在 `./fig/` 目录则写入其中
- 提供 Agent Skill launcher（`skills/gpt-image/scripts/generate.py`）

## 安装

```bash
# 从源码构建（Go 1.22+）
git clone git@github.com:ZacharyJia/gpt-image2-cli.git
cd gpt-image2-cli
go build -o gpt-image ./cmd/gpt-image

# 可选：安装到 PATH
go install ./cmd/gpt-image
```

## 快速使用

```bash
# 基础生成
gpt-image -p "a cat astronaut on the moon"

# 指定输出、竖版 2K、高质量
gpt-image -p "Chinese tea poster" -f poster.png --size portrait --quality high

# 编辑已有图片
gpt-image -p "colorize this manga page" -i page.jpg -f colored.png

# 多参考图编辑
gpt-image -p "77 × KFC collab poster" -i cat.png -i kfc_logo.png -f collab.png

# 使用 Alpha 蒙版局部重绘
gpt-image -p "replace sky with aurora" -i photo.jpg -m sky_mask.png -f aurora.png

# 一次生成 4 张，不透明背景，webp 格式
gpt-image -p "isometric chair, minimalist" -n 4 --background opaque --format webp
```

## 配置

CLI 按以下顺序读取凭证（进程 env 优先级最高，文件不会覆盖它）：

1. `OPENAI_API_KEY` 环境变量
2. `API_KEY` 环境变量（兜底）
3. `./.env`
4. `~/.env`

如需使用自定义或测试端点，设置 `BASE_URL`：

```bash
export BASE_URL="https://your-test-endpoint.example.com/v1"
export OPENAI_API_KEY="sk-..."
```

## 参数

| 标志 | 取值 | 默认值 | 说明 |
|---|---|---|---|
| `-p, --prompt` | 字符串 | **必需** | 提示词 / 编辑指令 |
| `-f, --file` | 路径 | 自动生成 | 输出路径 |
| `-i, --image` | 可重复路径 | — | 参考图；存在时切换到 edits 端点 |
| `-m, --mask` | PNG 路径 | — | Alpha 蒙版；需要 `-i` |
| `--model` | 字符串 | `gpt-image-2` | 图像模型 |
| `--size` | 简写 / 字面量 | `1024x1024` | 画布尺寸 |
| `--quality` | `auto` / `low` / `medium` / `high` | `high` | 质量 / 成本调节 |
| `-n, --n` | 整数 | `1` | 生成数量 |
| `--background` | `auto` / `opaque` | — | 生成背景 |
| `--moderation` | `auto` / `low` | `low` | 内容审核 |
| `--input-fidelity` | `low` / `high` | — | 编辑参数；对 `gpt-image-2` 本地丢弃 |
| `--format` | `png` / `jpeg` / `webp` | `png` | 输出编码 |
| `--compression` | `0-100` | — | JPEG/WebP 压缩 |
| `--user` | 字符串 | — | 最终用户标识 |

退出代码：`0` 成功 · `1` API/拒绝错误 · `2` 参数错误或缺失密钥。

## Agent Skill 安装

`skills/gpt-image/` 文件夹可安装到任何支持 Skill 的 Agent 运行时。

### Codex

```bash
$skill-installer
Install this skill from GitHub:
https://github.com/ZacharyJia/gpt-image2-cli/tree/main/skills/gpt-image
```

或手动：

```bash
mkdir -p "${CODEX_HOME:-$HOME/.codex}/skills"
test -e "${CODEX_HOME:-$HOME/.codex}/skills/gpt-image" && echo "already exists" && exit 1
cp -R skills/gpt-image "${CODEX_HOME:-$HOME/.codex}/skills/"
```

### OpenClaw / Claude Code / Hermes Agent

```bash
export AGENT_SKILLS_DIR="/path/to/your/agent/skills"
mkdir -p "$AGENT_SKILLS_DIR"
cp -R skills/gpt-image "$AGENT_SKILLS_DIR/"
```

## 许可证

MIT
