package vision

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Vision struct{}

func New() *Vision {
	return &Vision{}
}

type ImageData struct {
	URL       string
	Base64    string
	MimeType  string
	ImageURL string
}

func (v *Vision) EncodeImage(path string) (ImageData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ImageData{}, err
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".jpg" {
		ext = ".jpeg"
	}

	mimeType := "image/" + strings.TrimPrefix(ext, ".")
	if mimeType == "image/" {
		mimeType = "image/jpeg"
	}

	return ImageData{
		Base64:   base64.StdEncoding.EncodeToString(data),
		MimeType: mimeType,
	}, nil
}

func (v *Vision) EncodeImageURL(imageURL string) (ImageData, error) {
	if _, err := url.Parse(imageURL); err != nil {
		return ImageData{}, fmt.Errorf("invalid URL: %w", err)
	}

	return ImageData{
		URL:       imageURL,
		ImageURL: imageURL,
		MimeType:  "image/jpeg",
	}, nil
}

func (v *Vision) EncodeFromURL(httpURL string) (ImageData, error) {
	resp, err := http.Get(httpURL)
	if err != nil {
		return ImageData{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ImageData{}, err
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	return ImageData{
		Base64:   base64.StdEncoding.EncodeToString(data),
		MimeType: contentType,
	}, nil
}

func (v *Vision) DetectMimeType(data []byte) string {
	if len(data) < 4 {
		return "application/octet-stream"
	}

	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "image/gif"
	}
	if data[0] == 0x42 && data[1] == 0x4D {
		return "image/bmp"
	}
	if data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		return "image/webp"
	}

	return "application/octet-stream"
}

type Message struct {
	Type     string   `json:"type"`
	ImageURL string   `json:"image_url,omitempty"`
	ImageURLData string  `json:"image_url_data,omitempty"`
}

func (v *Vision) BuildVisionMessage(msg string, images []string) ([]byte, error) {
	if len(images) == 0 {
		return []byte(msg), nil
	}

	var parts []string
	parts = append(parts, msg)

	for _, img := range images {
		imgData, err := v.EncodeImage(img)
		if err != nil {
			if strings.HasPrefix(img, "http") {
				imgData, err = v.EncodeFromURL(img)
			} else {
				continue
			}
		}
		if err != nil {
			continue
		}

		Message := Message{
			Type:  "image_url",
			ImageURLData: imgData.Base64,
		}
		msgBytes, _ := json.Marshal(Message)
		parts = append(parts, string(msgBytes))
	}

	return []byte(fmt.Sprintf("%s\n%s", parts[0], strings.Join(parts[1:], "\n"))), nil
}