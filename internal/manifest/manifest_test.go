package manifest

import "testing"

func TestParseReadsPortableAppKey(t *testing.T) {
	summary, err := Parse([]byte(`{"app":"hello_world","actions":{"run":{}}}`))
	if err != nil {
		t.Fatalf("Parse() error=%v", err)
	}
	if summary.App != "hello_world" {
		t.Fatalf("Parse() app=%q", summary.App)
	}
}

func TestParseRejectsMissingOrUnsafeAppKey(t *testing.T) {
	for _, data := range []string{`{}`, `{"app":"../secret"}`, `{"app":"한글"}`} {
		if _, err := Parse([]byte(data)); err == nil {
			t.Fatalf("Parse(%s) unexpectedly succeeded", data)
		}
	}
}
