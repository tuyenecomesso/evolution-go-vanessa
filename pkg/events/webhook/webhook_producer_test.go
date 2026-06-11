package webhook_producer

import "testing"

func TestResolveWebhookURLPrefersInstanceWebhook(t *testing.T) {
	got := resolveWebhookURL("https://global.example/webhook", "https://instance.example/webhook")

	if got != "https://instance.example/webhook" {
		t.Fatalf("expected instance webhook, got %q", got)
	}
}

func TestResolveWebhookURLFallsBackToGlobalWebhook(t *testing.T) {
	got := resolveWebhookURL("https://global.example/webhook", "")

	if got != "https://global.example/webhook" {
		t.Fatalf("expected global webhook fallback, got %q", got)
	}
}

func TestResolveWebhookURLTrimsWhitespace(t *testing.T) {
	got := resolveWebhookURL(" https://global.example/webhook ", " https://instance.example/webhook ")

	if got != "https://instance.example/webhook" {
		t.Fatalf("expected trimmed instance webhook, got %q", got)
	}
}
