// gpt-image is a Go CLI for OpenAI GPT Image 2 generations and edits.
//
// Usage:
//
//	# Generate
//	gpt-image -p "a cat astronaut on the moon"
//
//	# Edit
//	gpt-image -p "colorize this manga page" -i page.jpg -f colored.png
//
//	# Inpaint
//	gpt-image -p "replace sky with aurora" -i photo.jpg -m mask.png -f aurora.png
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZacharyJia/gpt-image2-cli/internal/api"
	"github.com/ZacharyJia/gpt-image2-cli/internal/env"
	"github.com/ZacharyJia/gpt-image2-cli/internal/output"
)

const (
	defaultModel      = "gpt-image-2"
	defaultSize       = "1024x1024"
	defaultModeration = "low"
)

var (
	prompt                string
	outFile               string
	images                stringList
	mask                  string
	model                 string
	size                  string
	quality               string
	n                     int
	background            string
	moderation            string
	inputFidelity         string
	outputFormat          string
	outputCompression     int
	outputCompressionSet  bool
	user                  string
)

// stringList implements flag.Value for repeatable -i flags.
type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ", ") }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func init() {
	flag.StringVar(&prompt, "p", "", "Text prompt / edit instruction (required)")
	flag.StringVar(&prompt, "prompt", "", "Text prompt / edit instruction (required)")
	flag.StringVar(&outFile, "f", "", "Output path (auto-generated if omitted)")
	flag.StringVar(&outFile, "file", "", "Output path (auto-generated if omitted)")
	flag.Var(&images, "i", "Reference image path (repeatable; switches to edits endpoint)")
	flag.Var(&images, "image", "Reference image path (repeatable; switches to edits endpoint)")
	flag.StringVar(&mask, "m", "", "Alpha-channel PNG mask (requires -i)")
	flag.StringVar(&mask, "mask", "", "Alpha-channel PNG mask (requires -i)")
	flag.StringVar(&model, "model", defaultModel, "Model ID")
	flag.StringVar(&size, "size", defaultSize, "Image size: 1k,2k,4k,portrait,landscape,square,wide,tall, or literal WxH")
	flag.StringVar(&quality, "quality", "high", "Rendering quality: auto, low, medium, high")
	flag.IntVar(&n, "n", 1, "Number of images")
	flag.StringVar(&background, "background", "", "Background: auto, opaque")
	flag.StringVar(&moderation, "moderation", defaultModeration, "Moderation: auto, low")
	flag.StringVar(&inputFidelity, "input-fidelity", "", "Edits only: low, high (dropped for gpt-image-2)")
	flag.StringVar(&outputFormat, "format", "", "Output encoding: png, jpeg, webp")
	flag.IntVar(&outputCompression, "compression", 0, "JPEG/WebP compression 0-100")
	flag.StringVar(&user, "user", "", "Optional end-user identifier")
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: gpt-image -p PROMPT [options]\n\n")
	fmt.Fprintf(os.Stderr, "Generate or edit images with OpenAI GPT Image 2.\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExit codes: 0 success, 1 API/refusal error, 2 bad arguments/missing key.\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "compression" {
			outputCompressionSet = true
		}
	})

	if prompt == "" {
		fmt.Fprintln(os.Stderr, "error: -p/--prompt is required")
		flag.Usage()
		os.Exit(2)
	}
	if mask != "" && len(images) == 0 {
		fmt.Fprintln(os.Stderr, "error: --mask requires --image")
		os.Exit(2)
	}

	credentials, err := env.Resolve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
	if credentials.APIKey == "" {
		fmt.Fprintln(os.Stderr, "error: API key not set. Add api_key to ~/.gpt-image2-cli/config.json, or set OPENAI_API_KEY (or API_KEY) in env / .env / ~/.env.")
		os.Exit(2)
	}

	client, err := api.New(credentials.APIKey, credentials.BaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	ext := outputFormat
	if ext == "" {
		ext = "png"
	}

	var outPath string
	if outFile != "" {
		outPath = filepath.Clean(outFile)
	} else {
		outPath = output.DefaultPath(prompt, ext)
	}

	req := api.Request{
		Model:        model,
		Prompt:       prompt,
		Size:         size,
		Quality:      quality,
		N:            n,
		Background:   background,
		Moderation:   moderation,
		OutputFormat: outputFormat,
		User:         user,
	}
	if outputCompressionSet {
		v := outputCompression
		req.OutputCompression = &v
	}

	ctx := context.Background()
	var resp *api.Response
	if len(images) > 0 {
		req.Images = images
		req.Mask = mask
		req.InputFidelity = inputFidelity
		resp, err = client.Edit(ctx, req)
	} else {
		resp, err = client.Generate(ctx, req)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(resp.Data) == 0 {
		fmt.Fprintln(os.Stderr, "error: no image data in response")
		os.Exit(1)
	}

	written := 0
	for idx, img := range resp.Data {
		raw, err := api.FetchImage(img, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: item %d: %v\n", idx, err)
			os.Exit(1)
		}

		target := outPath
		if n > 1 {
			stem := strings.TrimSuffix(outPath, filepath.Ext(outPath))
			target = fmt.Sprintf("%s_%d%s", stem, idx, filepath.Ext(outPath))
		}
		if err := output.Write(target, raw); err != nil {
			fmt.Fprintf(os.Stderr, "error: write %s: %v\n", target, err)
			os.Exit(1)
		}
		fmt.Println(target)
		written++
	}
	_ = written
}
