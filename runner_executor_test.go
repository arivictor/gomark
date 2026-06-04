package gomark

import (
	"context"
	"strings"
	"testing"
)

func TestUsesOnlyAllowedImports(t *testing.T) {
	ok := usesOnlyAllowedImports("package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"ok\")}")
	if !ok {
		t.Fatalf("expected fmt import to be allowed")
	}

	badExternal := usesOnlyAllowedImports("package main\nimport \"github.com/pkg/errors\"\nfunc main(){}")
	if badExternal {
		t.Fatalf("expected external import to be rejected")
	}

	badNetwork := usesOnlyAllowedImports("package main\nimport \"net/http\"\nfunc main(){}")
	if badNetwork {
		t.Fatalf("expected network import to be rejected")
	}

	badExec := usesOnlyAllowedImports("package main\nimport \"os/exec\"\nfunc main(){}")
	if badExec {
		t.Fatalf("expected os/exec import to be rejected")
	}
}

func TestGoExecutorRunSuccess(t *testing.T) {
	r := GoExecutor{}
	result := r.Run(context.Background(), "package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"hello\")}")
	if !result.OK {
		t.Fatalf("expected success, got error: %#v", result)
	}
	if !strings.Contains(result.Output, "hello") {
		t.Fatalf("expected output to contain hello, got: %q", result.Output)
	}
}
