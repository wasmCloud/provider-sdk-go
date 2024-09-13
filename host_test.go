package provider

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestRedactedStringLogging(t *testing.T) {
	var buf bytes.Buffer
	secret := "its-a-secret"
	redactedString := RedactedString(secret)

	jsonSlog := slog.New(slog.NewJSONHandler(&buf, nil))
	jsonSlog.Info("jsonSlog", "redactedString", redactedString)

	if !strings.Contains(buf.String(), "\"redactedString\":\"redacted(string)\"") || strings.Contains(buf.String(), secret) {
		t.Error("json slog handler output should not have contained the secret string")
	}

	buf.Reset()

	textSlog := slog.New(slog.NewTextHandler(&buf, nil))
	textSlog.Info("textSlog", "redactedString", redactedString)

	if !strings.Contains(buf.String(), "redactedString=redacted(string)") || strings.Contains(buf.String(), secret) {
		t.Error("text slog handler should not have contained the secret string")
	}
}
