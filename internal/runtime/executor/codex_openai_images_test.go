package executor

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestCodexBuildImagesResponsesRequest_DoesNotForceToolChoice(t *testing.T) {
	body := codexBuildImagesResponsesRequest("draw an apple", nil, []byte(`{"type":"image_generation","model":"gpt-image-2","size":"1024x1024"}`))

	if gjson.GetBytes(body, "tool_choice").Exists() {
		t.Fatalf("tool_choice should not be set for OpenAI image requests: %s", string(body))
	}

	tools := gjson.GetBytes(body, "tools")
	if !tools.IsArray() || len(tools.Array()) != 1 {
		t.Fatalf("expected one image_generation tool, got %s", tools.Raw)
	}
	if got := tools.Array()[0].Get("type").String(); got != "image_generation" {
		t.Fatalf("tool type = %q, want image_generation", got)
	}
	if got := gjson.GetBytes(body, "input.0.content.0.text").String(); got != "draw an apple" {
		t.Fatalf("prompt = %q, want draw an apple", got)
	}
}
