# GPT Image 2 CLI (Go)

A Go implementation of the GPT Image 2 command-line tool, plus an Agentic Skill surface for Codex / Claude Code / OpenClaw and other skill-aware runtimes.

This project mirrors the Python reference implementation in [`wuyoscar/gpt_image_2_skill`](https://github.com/wuyoscar/gpt_image_2_skill) and reuses its prompt-gallery references, while the CLI itself is written in Go with **zero external dependencies**.

## Features

- Text-to-image generation (`/v1/images/generations`)
- Reference-image editing, including multi-reference edits (`/v1/images/edits`)
- Alpha-channel inpainting (`-i` + `-m`)
- All major OpenAI image parameters: size, quality, background, moderation, format, compression, user
- Size shortcuts: `1k`, `2k`, `4k`, `portrait`, `landscape`, `square`, `wide`, `tall`
- Reads `OPENAI_API_KEY` (or `API_KEY`) and optional `BASE_URL` from env, `.env`, `~/.env` (process env wins)
- Auto-named output files, written to `./fig/` when present
- Skill launcher (`skills/gpt-image/scripts/generate.py`) for agent runtimes

## Install

### From Release (recommended)

Download the latest prebuilt binary for your platform from the [Releases](https://github.com/ZacharyJia/gpt-image2-cli/releases/latest) page, or install via the command line:

```bash
# macOS Apple Silicon
curl -L -o gpt-image "https://github.com/ZacharyJia/gpt-image2-cli/releases/latest/download/gpt-image-darwin-arm64"
chmod +x gpt-image
sudo mv gpt-image /usr/local/bin/

# macOS Intel
curl -L -o gpt-image "https://github.com/ZacharyJia/gpt-image2-cli/releases/latest/download/gpt-image-darwin-amd64"
chmod +x gpt-image
sudo mv gpt-image /usr/local/bin/

# Linux AMD64
curl -L -o gpt-image "https://github.com/ZacharyJia/gpt-image2-cli/releases/latest/download/gpt-image-linux-amd64"
chmod +x gpt-image
sudo mv gpt-image /usr/local/bin/

# Windows AMD64 (PowerShell)
curl -L -o gpt-image.exe "https://github.com/ZacharyJia/gpt-image2-cli/releases/latest/download/gpt-image-windows-amd64.exe"
```

### Build from source

```bash
# Build from source (Go 1.26+)
git clone git@github.com:ZacharyJia/gpt-image2-cli.git
cd gpt-image2-cli
go build -o gpt-image ./cmd/gpt-image

# Optional: install to PATH
go install ./cmd/gpt-image
```

## Quick start

```bash
# Basic generation
gpt-image -p "a cat astronaut on the moon"

# Named output, portrait 2K, high quality
gpt-image -p "Chinese tea poster" -f poster.png --size portrait --quality high

# Edit existing image
gpt-image -p "colorize this manga page" -i page.jpg -f colored.png

# Multi-reference edit
gpt-image -p "77 × KFC collab poster" -i cat.png -i kfc_logo.png -f collab.png

# Inpaint with alpha mask
gpt-image -p "replace sky with aurora" -i photo.jpg -m sky_mask.png -f aurora.png

# Grid of 4, opaque background, webp
gpt-image -p "isometric chair, minimalist" -n 4 --background opaque --format webp
```

## Configuration

The CLI reads credentials in this order (process env wins, files never override it):

1. `OPENAI_API_KEY` environment variable
2. `API_KEY` environment variable (fallback)
3. `./.env`
4. `~/.env`

For custom or test endpoints, set `BASE_URL`:

```bash
export BASE_URL="https://your-test-endpoint.example.com/v1"
export OPENAI_API_KEY="sk-..."
```

## Flags

| Flag | Values | Default | Description |
|---|---|---|---|
| `-p, --prompt` | string | **required** | Prompt / edit instruction |
| `-f, --file` | path | auto | Output path |
| `-i, --image` | repeatable path | — | Reference image(s); switches to edits endpoint |
| `-m, --mask` | PNG path | — | Alpha mask; requires `-i` |
| `--model` | string | `gpt-image-2` | Image model |
| `--size` | shortcuts / literal | `1024x1024` | Canvas size |
| `--quality` | `auto` / `low` / `medium` / `high` | `high` | Quality / cost dial |
| `-n, --n` | int | `1` | Number of images |
| `--background` | `auto` / `opaque` | — | Generation background |
| `--moderation` | `auto` / `low` | `low` | Generation moderation |
| `--input-fidelity` | `low` / `high` | — | Dropped locally for `gpt-image-2` |
| `--format` | `png` / `jpeg` / `webp` | `png` | Output encoding |
| `--compression` | `0-100` | — | JPEG/WebP compression |
| `--user` | string | — | End-user identifier |

Exit codes: `0` success, `1` API/refusal error, `2` bad arguments/missing key.

## Agent Skill installation

The `skills/gpt-image/` folder can be installed into any skill-aware agent runtime.

### Codex

```bash
$skill-installer
Install this skill from GitHub:
https://github.com/ZacharyJia/gpt-image2-cli/tree/main/skills/gpt-image
```

Or manually:

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

## License

MIT
