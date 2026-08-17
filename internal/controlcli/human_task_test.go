package controlcli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHumanTaskDecideGeneratesRequiredIdempotencyKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/w/team/human-tasks/task-1/decision" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		key := r.Header.Get("Idempotency-Key")
		if !strings.HasPrefix(key, "imprun-human-task-") || len(key) != len("imprun-human-task-")+32 {
			t.Fatalf("idempotency key was not generated")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		value, ok := body["value"].(map[string]any)
		if body["outcome"] != "submit" || !ok || value["completed"] != true {
			t.Fatalf("request body = %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"task":{"id":"task-1"},"replayed":false}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	exit := RunImprun(
		[]string{"--api-url", server.URL, "--workspace", "team", "human-task", "decide", "task-1", "--outcome", "submit", "--value-file", "-"},
		strings.NewReader(`{"completed":true}`),
		&stdout,
		&stderr,
	)
	if exit != ExitOK {
		t.Fatalf("exit=%d stderr=%s", exit, stderr.String())
	}
	if strings.Contains(stdout.String(), "completed") || strings.Contains(stderr.String(), "completed") {
		t.Fatal("decision value leaked to command output")
	}
}

func TestHumanTaskCancelUsesCallerIdempotencyKeyWithoutDecisionValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Idempotency-Key"); got != "decision-1" {
			t.Fatalf("idempotency key = %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["outcome"] != "cancel" || len(body) != 1 {
			t.Fatalf("request body = %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"task":{"id":"task-2"},"replayed":false}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	exit := RunImprun(
		[]string{"--api-url", server.URL, "--workspace", "team", "human-task", "decide", "task-2", "--outcome", "cancel", "--idempotency-key", "decision-1"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if exit != ExitOK {
		t.Fatalf("exit=%d stderr=%s", exit, stderr.String())
	}
}

func TestHumanTaskDecideHelpDoesNotRequireConfiguration(t *testing.T) {
	t.Setenv("IMPRUN_CONFIG", t.TempDir())
	var stdout, stderr bytes.Buffer
	exit := RunImprun(
		[]string{"human-task", "decide", "--help"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if exit != ExitOK || !strings.Contains(stdout.String(), "always sends an Idempotency-Key") || stderr.Len() != 0 {
		t.Fatalf("exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
}
