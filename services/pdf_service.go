package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"net/url"
	"os/exec"
	"time"

	"lite-collector/utils"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

// PDFService converts markdown reports into PDFs via headless Chromium.
// It is nil-safe: when chromium is unavailable the constructor returns nil,
// and Render on a nil receiver surfaces ErrPDFNotAvailable so callers can
// return a clean 503 to the frontend.
type PDFService struct {
	execPath string
}

// NewPDFService looks up a chromium binary and returns a service if found,
// or nil when no executable is available.
func NewPDFService() *PDFService {
	for _, c := range []string{
		"chromium",
		"chromium-browser",
		"google-chrome",
		"google-chrome-stable",
	} {
		if path, err := exec.LookPath(c); err == nil {
			return &PDFService{execPath: path}
		}
	}
	return nil
}

// Render converts markdown body + title to a PDF byte slice.
func (s *PDFService) Render(title, mdText string) ([]byte, error) {
	if s == nil {
		return nil, utils.ErrPDFNotAvailable
	}

	htmlBody := markdownToHTML(mdText)
	document := buildHTMLDocument(title, htmlBody)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(s.execPath),
			chromedp.NoSandbox,
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("disable-dev-shm-usage", true),
		)...,
	)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	timed, cancelTimed := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimed()

	var buf []byte
	dataURL := "data:text/html;charset=utf-8," + url.PathEscape(document)
	err := chromedp.Run(timed,
		chromedp.Navigate(dataURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			data, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				Do(ctx)
			if err != nil {
				return err
			}
			buf = data
			return nil
		}),
	)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: chromium timed out", utils.ErrPDFGenerateFail)
		}
		return nil, fmt.Errorf("%w: %v", utils.ErrPDFGenerateFail, err)
	}
	return buf, nil
}

func markdownToHTML(md string) string {
	p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs)
	return string(markdown.ToHTML([]byte(md), p, nil))
}

func buildHTMLDocument(title, body string) string {
	var b bytes.Buffer
	b.WriteString(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8">`)
	b.WriteString(`<title>`)
	b.WriteString(html.EscapeString(title))
	b.WriteString(`</title>`)
	b.WriteString(`<style>
		@page { size: A4; margin: 20mm 18mm; }
		body { font-family: "Noto Sans CJK SC", "Noto Sans SC", "PingFang SC",
			"Microsoft YaHei", "WenQuanYi Micro Hei", sans-serif;
			line-height: 1.65; color: #1f2328; font-size: 14px; }
		h1, h2, h3, h4 { margin-top: 1.4em; }
		h1 { font-size: 22px; border-bottom: 1px solid #d0d7de; padding-bottom: 4px; }
		h2 { font-size: 18px; }
		h3 { font-size: 16px; }
		p { margin: 0.6em 0; }
		code { font-family: "SFMono-Regular", Consolas, monospace;
			background: #f6f8fa; padding: 1px 4px; border-radius: 3px; }
		pre { background: #f6f8fa; padding: 12px; border-radius: 6px; overflow: auto; }
		ul, ol { padding-left: 24px; }
		blockquote { border-left: 3px solid #d0d7de; margin: 0; padding-left: 12px; color: #57606a; }
		table { border-collapse: collapse; width: 100%; margin: 0.8em 0; }
		th, td { border: 1px solid #d0d7de; padding: 6px 10px; text-align: left; }
		th { background: #f6f8fa; }
		.report-header { margin-bottom: 18px; }
		.report-title { font-size: 24px; font-weight: 700; }
	</style></head><body>`)
	b.WriteString(`<div class="report-header"><div class="report-title">`)
	b.WriteString(html.EscapeString(title))
	b.WriteString(`</div></div>`)
	b.WriteString(body)
	b.WriteString(`</body></html>`)
	return b.String()
}
