package validation

import "testing"

func TestWorkspaceIDMatchesCellContract(t *testing.T) {
	for _, value := range []string{"default", "gale-7c35", "a1"} {
		if !WorkspaceID(value) {
			t.Fatalf("WorkspaceID(%q)=false", value)
		}
	}
	for _, value := range []string{"a", "-gale", "gale-", "Gale", "gale_space"} {
		if WorkspaceID(value) {
			t.Fatalf("WorkspaceID(%q)=true", value)
		}
	}
}
