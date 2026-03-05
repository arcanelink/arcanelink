package handler

import (
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/arcane/arcanelink/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

type LinkPreview struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	SiteName    string `json:"site_name"`
}

// GetLinkPreview fetches metadata for a given URL
func (h *APIHandler) GetLinkPreview(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Missing url parameter")
		return
	}

	// Validate URL
	if !isValidURL(url) {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid URL")
		return
	}

	preview, err := fetchLinkPreview(url)
	if err != nil {
		logger.Warn("Failed to fetch link preview, returning basic info", zap.Error(err), zap.String("url", url))
		// Return basic preview with just URL
		preview = &LinkPreview{
			URL:   url,
			Title: url,
		}
	}

	respondJSON(w, http.StatusOK, preview)
}

func isValidURL(url string) bool {
	pattern := `^https?://[^\s]+$`
	matched, _ := regexp.MatchString(pattern, url)
	return matched
}

func fetchLinkPreview(url string) (*LinkPreview, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ArcaneLink/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	// Limit response body size to 1MB
	limitedReader := io.LimitReader(resp.Body, 1024*1024)
	doc, err := html.Parse(limitedReader)
	if err != nil {
		return nil, err
	}

	preview := &LinkPreview{
		URL: url,
	}

	extractMetadata(doc, preview)

	return preview, nil
}

func extractMetadata(n *html.Node, preview *LinkPreview) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "title":
			if n.FirstChild != nil {
				preview.Title = n.FirstChild.Data
			}
		case "meta":
			var property, content, name string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "property":
					property = attr.Val
				case "content":
					content = attr.Val
				case "name":
					name = attr.Val
				}
			}

			// Open Graph tags
			switch property {
			case "og:title":
				if preview.Title == "" {
					preview.Title = content
				}
			case "og:description":
				preview.Description = content
			case "og:image":
				preview.Image = content
			case "og:site_name":
				preview.SiteName = content
			}

			// Twitter Card tags
			switch name {
			case "twitter:title":
				if preview.Title == "" {
					preview.Title = content
				}
			case "twitter:description":
				if preview.Description == "" {
					preview.Description = content
				}
			case "twitter:image":
				if preview.Image == "" {
					preview.Image = content
				}
			case "description":
				if preview.Description == "" {
					preview.Description = content
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractMetadata(c, preview)
	}
}
