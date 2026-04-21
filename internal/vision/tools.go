package vision

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hermes-v2/aigo/internal/tools"
)

type ImageEncodeTool struct {
	vision *Vision
}

func NewImageEncodeTool() *ImageEncodeTool {
	return &ImageEncodeTool{vision: New()}
}

func (t *ImageEncodeTool) Name() string   { return "vision_encode" }
func (t *ImageEncodeTool) Description() string {
	return "Encode image to base64 for vision models"
}

func (t *ImageEncodeTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *ImageEncodeTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "vision_encode",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Image file path or URL"},
				},
				"required": []string{"path"},
			},
		},
	}
}

func (t *ImageEncodeTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	var imgData ImageData
	var err error

	if strings.HasPrefix(path, "http") {
		imgData, err = t.vision.EncodeFromURL(path)
	} else {
		imgData, err = t.vision.EncodeImage(path)
	}
	if err != nil {
		return "", err
	}

	data, _ := json.Marshal(map[string]string{
		"base64":  imgData.Base64,
		"type":   imgData.MimeType,
		"status": "encoded",
	})
	return string(data), nil
}

type ImageDetectTool struct {
	vision *Vision
}

func NewImageDetectTool() *ImageDetectTool {
	return &ImageDetectTool{vision: New()}
}

func (t *ImageDetectTool) Name() string   { return "vision_detect_type" }
func (t *ImageDetectTool) Description() string {
	return "Detect image MIME type from file or bytes"
}

func (t *ImageDetectTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *ImageDetectTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "vision_detect_type",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Image file path"},
				},
				"required": []string{"path"},
			},
		},
	}
}

func (t *ImageDetectTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	mimeType := t.vision.DetectMimeType(data)
	return fmt.Sprintf("MIME: %s", mimeType), nil
}

func RegisterVisionTools(reg *tools.Registry) {
	encodeTool := NewImageEncodeTool()
	reg.Register(encodeTool)

	detectTool := NewImageDetectTool()
	reg.Register(detectTool)
}