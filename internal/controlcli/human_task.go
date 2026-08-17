package controlcli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

func (r *runner) humanTask(args []string) error {
	if len(args) == 0 {
		return usageError{"human-task requires decide"}
	}
	if args[0] != "decide" {
		return usageError{fmt.Sprintf("unknown human-task command %q", args[0])}
	}
	if len(args) < 2 {
		return usageError{"usage: imprun human-task decide <task-id> --outcome submit|cancel [flags]"}
	}
	taskID := strings.TrimSpace(args[1])
	if taskID == "" {
		return usageError{"human-task ID must not be empty"}
	}

	fs := r.flags("human-task decide")
	outcome := fs.String("outcome", "", "submit or cancel")
	value := fs.String("value", "null", "JSON decision value for submit")
	valueFile := fs.String("value-file", "", "JSON decision value file, or - for stdin")
	idempotencyKey := fs.String("idempotency-key", "", "principal-scoped idempotency key; generated when omitted")
	if err := fs.Parse(args[2:]); err != nil {
		return usageError{err.Error()}
	}
	if fs.NArg() != 0 || (*outcome != "submit" && *outcome != "cancel") {
		return usageError{"usage: imprun human-task decide <task-id> --outcome submit|cancel [flags]"}
	}

	body := map[string]any{"outcome": *outcome}
	if *outcome == "submit" {
		decisionValue, err := r.readJSON(*value, *valueFile)
		if err != nil {
			return err
		}
		body["value"] = decisionValue
	} else if *valueFile != "" || *value != "null" {
		return usageError{"--value and --value-file are only valid with --outcome submit"}
	}

	key := strings.TrimSpace(*idempotencyKey)
	if key == "" {
		generated, err := newHumanTaskIdempotencyKey()
		if err != nil {
			return err
		}
		key = generated
	}
	if len(key) > 200 {
		return usageError{"--idempotency-key must not exceed 200 characters"}
	}

	return r.jsonWithHeaders(
		http.MethodPost,
		r.client.WorkspacePath("human-tasks", taskID, "decision"),
		body,
		map[string]string{"Idempotency-Key": key},
	)
}

func newHumanTaskIdempotencyKey() (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate HumanTask idempotency key: %w", err)
	}
	return "imprun-human-task-" + hex.EncodeToString(random), nil
}
