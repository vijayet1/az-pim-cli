package automation

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Runner interface {
	Run(context.Context, string, []string, string, ...string) ([]byte, error)
}
type ExecRunner struct{}
type CommandError struct {
	HTTPStatus int
	ExitCode   int
	Message    string
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("command failed (exit=%d HTTP=%d): %s", e.ExitCode, e.HTTPStatus, e.Message)
}

var httpStatusPattern = regexp.MustCompile(`\(HTTP ([0-9]{3})\)`)

func (ExecRunner) Run(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
	var c *exec.Cmd
	switch name {
	case "git":
		c = exec.CommandContext(ctx, "git", args...)
	case "gh":
		c = exec.CommandContext(ctx, "gh", args...)
	default:
		return nil, fmt.Errorf("unsupported executable: only git and gh are allowed")
	}
	c.Dir = dir
	c.Env = append(os.Environ(), env...)
	var stderr bytes.Buffer
	c.Stderr = &stderr
	out, err := c.Output()
	if err == nil {
		return out, nil
	}
	message := stderr.String()
	for _, entry := range c.Env {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || value == "" {
			continue
		}
		if key == "GH_TOKEN" || key == "GITHUB_TOKEN" {
			message = strings.ReplaceAll(message, value, "[REDACTED]")
			message = strings.ReplaceAll(message, base64.StdEncoding.EncodeToString([]byte("x-access-token:"+value)), "[REDACTED]")
		}
		if strings.HasPrefix(key, "GIT_CONFIG_VALUE_") {
			message = strings.ReplaceAll(message, value, "[REDACTED]")
		}
	}
	status := 0
	if m := httpStatusPattern.FindStringSubmatch(message); m != nil {
		status, _ = strconv.Atoi(m[1])
	}
	code := -1
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		code = exit.ExitCode()
	}
	if len(message) > 2000 {
		message = message[:2000]
	}
	return nil, &CommandError{status, code, message}
}
func is404(err error) bool { var e *CommandError; return errors.As(err, &e) && e.HTTPStatus == 404 }
func exitCode(err error, code int) bool {
	var e *CommandError
	return errors.As(err, &e) && e.ExitCode == code
}

type commands struct {
	ctx context.Context
	r   Runner
	dir string
}

func (c commands) git(args ...string) (string, error) {
	b, e := c.r.Run(c.ctx, c.dir, nil, "git", append([]string{"--no-pager", "-c", "core.hooksPath=/dev/null"}, args...)...)
	return strings.TrimSpace(string(b)), e
}
func (c commands) gh(args ...string) ([]byte, error) {
	return c.r.Run(c.ctx, c.dir, nil, "gh", args...)
}
func (c commands) api(path string) ([]byte, error) {
	return c.gh("api", "--hostname", "github.com", path)
}
func forkAuth() []string {
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return nil
	}
	return []string{"GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0=http.https://github.com/" + Fork + ".git.extraheader", "GIT_CONFIG_VALUE_0=AUTHORIZATION: basic " + base64.StdEncoding.EncodeToString([]byte("x-access-token:"+token))}
}
func (c commands) push(args ...string) error {
	_, e := c.r.Run(c.ctx, c.dir, forkAuth(), "git", append([]string{"-c", "core.hooksPath=/dev/null", "push"}, args...)...)
	return e
}
