package main

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/playwright-community/playwright-go"
)

type InputType uint

const (
	FileType InputType = iota
	UrlType
	TermType
)

func getType(input string) InputType {
	if files, err := filepath.Glob(input); err == nil {
		for _, file := range files {
			if _, err := os.Stat(file); err == nil {
				return FileType
			}
		}
	}
	if u, err := url.ParseRequestURI(input); err == nil && u.Scheme != "" {
		return UrlType
	}
	return TermType
}

var defaultBrowser playwright.Browser

func startBrowser() error {
	runOptions := &playwright.RunOptions{
		SkipInstallBrowsers: true,
	}
	if err := playwright.Install(runOptions); err != nil {
		return err
	}
	pw, err := playwright.Run(runOptions)
	if err != nil {
		return err
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Channel:  playwright.String("chrome"),
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return err
	}
	defaultBrowser = browser
	return nil
}

func scrape(url string) (string, error) {
	if defaultBrowser == nil || !defaultBrowser.IsConnected() {
		if err := startBrowser(); err != nil {
			return "", err
		}
	}
	page, err := defaultBrowser.NewPage()
	if err != nil {
		return "", err
	}
	res, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		return "", err
	}
	if err := res.Finished(); err != nil {
		return "", err
	}
	defer page.Close()
	b, err := res.Body()
	if err != nil {
		return "", err
	}
	return htmltomarkdown.ConvertString(string(b))
}

func readFilePattern(pattern string) (string, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	content := make([]string, len(files))
	for i, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		content[i] = string(b)
	}
	return strings.Join(content, "\n"), nil
}

func readContext(paths []string) (string, error) {
	var contexts strings.Builder
	for _, path := range paths {
		switch getType(path) {
		case UrlType:
			content, err := scrape(path)
			if err != nil {
				return "", err
			}
			contexts.WriteString(content + "\n")
		case FileType:
			content, err := readFilePattern(path)
			if err != nil {
				return "", err
			}
			contexts.WriteString(content + "\n")
		default:
			return strings.Join(paths, " "), nil
		}
	}
	return contexts.String(), nil
}
