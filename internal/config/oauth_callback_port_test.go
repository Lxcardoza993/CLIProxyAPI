package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func unmarshalOAuthCallbackPort(t *testing.T, in string) (OAuthCallbackPortConfig, error) {
	t.Helper()
	var cfg struct {
		OAuthCallbackPort OAuthCallbackPortConfig `yaml:"oauth-callback-port"`
	}
	err := yaml.Unmarshal([]byte(in), &cfg)
	return cfg.OAuthCallbackPort, err
}

func TestOAuthCallbackPortScalarForm(t *testing.T) {
	got, err := unmarshalOAuthCallbackPort(t, "oauth-callback-port: 51129\n")
	if err != nil {
		t.Fatalf("scalar form failed: %v", err)
	}
	if got.Global != 51129 {
		t.Fatalf("Global = %d, want 51129", got.Global)
	}
	if got.Antigravity != 0 || got.Anthropic != 0 || got.Codex != 0 || got.XAI != 0 {
		t.Fatalf("per-channel values should stay zero, got %+v", got)
	}
}

func TestOAuthCallbackPortMapForm(t *testing.T) {
	got, err := unmarshalOAuthCallbackPort(t, "oauth-callback-port:\n  antigravity: 51129\n  anthropic: 54546\n  codex: 1456\n  xai: 56122\n")
	if err != nil {
		t.Fatalf("map form failed: %v", err)
	}
	if got.Global != 0 || got.Antigravity != 51129 || got.Anthropic != 54546 || got.Codex != 1456 || got.XAI != 56122 {
		t.Fatalf("unexpected values: %+v", got)
	}
}

func TestOAuthCallbackPortRejectsInvalid(t *testing.T) {
	cases := []string{
		"oauth-callback-port: not-a-port\n",
		"oauth-callback-port:\n  - 51129\n",
	}
	for _, in := range cases {
		if _, err := unmarshalOAuthCallbackPort(t, in); err == nil {
			t.Fatalf("expected error for input %q", strings.TrimSpace(in))
		}
	}
}

func TestOAuthCallbackPortZeroIsUnset(t *testing.T) {
	// Zero and negative values are treated as unset for backward compatibility
	// with the previous int-typed field.
	for _, in := range []string{"oauth-callback-port: 0\n", "oauth-callback-port: -1\n"} {
		got, err := unmarshalOAuthCallbackPort(t, in)
		if err != nil {
			t.Fatalf("input %q should parse: %v", strings.TrimSpace(in), err)
		}
		if got.Global <= 0 && got.Antigravity == 0 && got.Anthropic == 0 && got.Codex == 0 && got.XAI == 0 {
			continue
		}
		t.Fatalf("input %q should resolve to unset, got %+v", strings.TrimSpace(in), got)
	}
	if got := (OAuthCallbackPortConfig{Global: 0}).Resolve("anthropic", 54545); got != 54545 {
		t.Fatalf("global=0 should fall through to provider default: got %d", got)
	}
}

func TestOAuthCallbackPortResolve(t *testing.T) {
	cfg := OAuthCallbackPortConfig{Global: 5000, Anthropic: 54546}
	if got := cfg.Resolve("anthropic", 54545); got != 54546 {
		t.Fatalf("per-channel should win: got %d", got)
	}
	if got := cfg.Resolve("codex", 1455); got != 5000 {
		t.Fatalf("global scalar should beat provider default: got %d", got)
	}
	if got := (OAuthCallbackPortConfig{}).Resolve("codex", 1455); got != 1455 {
		t.Fatalf("empty config should use provider default: got %d", got)
	}
	if got := cfg.Resolve("unknown-channel", 0); got != 5000 {
		t.Fatalf("unknown channel should still honor global scalar: got %d", got)
	}
}

func TestOAuthCallbackPortRoundTrip(t *testing.T) {
	// Scalar-only config must round-trip as the legacy scalar.
	scalar := OAuthCallbackPortConfig{Global: 51129}
	out, err := yaml.Marshal(scalar)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "51129" {
		t.Fatalf("scalar round-trip = %q", string(out))
	}
	back, err := unmarshalOAuthCallbackPort(t, "oauth-callback-port: "+string(out))
	if err != nil {
		t.Fatalf("re-unmarshal failed: %v", err)
	}
	if back.Global != 51129 || back.Antigravity != 0 {
		t.Fatalf("scalar round-trip mismatch: %+v", back)
	}

	// Map config must round-trip preserving per-channel values.
	mapped := OAuthCallbackPortConfig{Antigravity: 51129, Anthropic: 54546, Codex: 1456, XAI: 56122}
	out, err = yaml.Marshal(mapped)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var backMap OAuthCallbackPortConfig
	if err := yaml.Unmarshal(out, &backMap); err != nil {
		t.Fatalf("re-unmarshal failed: %v", err)
	}
	if backMap != mapped {
		t.Fatalf("map round-trip mismatch: %+v vs %+v", backMap, mapped)
	}
}
