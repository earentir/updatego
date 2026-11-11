package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const goIntrospectTimeout = 15 * time.Second

// offlineGoEnvKeys are unset so a child go process cannot inherit settings
// that trigger toolchain or module downloads.
var offlineGoEnvKeys = []string{
	"GOTOOLCHAIN", "GOROOT", "GOPROXY", "GOSUMDB", "GOFLAGS", "GOENV",
}

// GoTreeEnv returns an environment for running goroot/bin/go without network use.
// GOTOOLCHAIN=local disables toolchain downloads; GOPROXY/GOSUMDB=off block fetches.
func GoTreeEnv(goroot string) []string {
	env := os.Environ()
	for _, key := range offlineGoEnvKeys {
		env = withoutEnvKey(env, key)
	}
	env = append(env,
		"GOTOOLCHAIN=local",
		"GOPROXY=off",
		"GOSUMDB=off",
	)
	if goroot != "" {
		env = append(env, "GOROOT="+goroot)
	}
	return env
}

func withoutEnvKey(env []string, key string) []string {
	out := make([]string, 0, len(env))
	for _, e := range env {
		if strings.HasPrefix(e, key+"=") || e == key {
			continue
		}
		out = append(out, e)
	}
	return out
}

// RunGoTree runs goroot/bin/go with args using GOTOOLCHAIN=local and a timeout.
func RunGoTree(goroot string, args ...string) ([]byte, error) {
	if goroot == "" {
		return nil, fmt.Errorf("empty Go install root")
	}
	goBin := filepath.Join(goroot, "bin", "go")
	ctx, cancel := context.WithTimeout(context.Background(), goIntrospectTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, goBin, args...)
	cmd.Env = GoTreeEnv(goroot)
	cmd.Dir = goroot
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return out, fmt.Errorf("timed out after %s", goIntrospectTimeout)
		}
		return out, err
	}
	return out, nil
}
