package store

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// passPrefix is the folder of the password store that holds supawho's entries.
const passPrefix = "supawho/"

// ErrPassNotInstalled is returned when the pass backend is selected but the
// `pass` executable is not on PATH.
var ErrPassNotInstalled = errors.New("pass is not installed")

// passBackend stores secrets with pass (https://www.passwordstore.org/):
// one GPG-encrypted entry per key under supawho/. It needs no D-Bus or
// desktop session, so it works on headless machines, and gpg-agent caches
// the GPG passphrase.
//
// Secrets are only ever written to pass's stdin and read from its stdout:
// they never appear in arguments, temp files or error messages.
type passBackend struct{}

// NewPass returns a Store backed by pass.
func NewPass() *Store {
	return newStore(passBackend{})
}

func (passBackend) name() string { return BackendPass }

func (passBackend) get(key string) (string, error) {
	out, err := runPass("", "show", key)
	if err != nil {
		return "", err
	}
	line, _, _ := strings.Cut(out, "\n")
	return strings.TrimSuffix(line, "\r"), nil
}

func (passBackend) set(key, value string) error {
	_, err := runPass(value, "insert", "--multiline", "--force", key)
	return err
}

func (passBackend) delete(key string) error {
	_, err := runPass("", "rm", "--force", key)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}

// runPass runs `pass <verb> [flags] supawho/<key>` with stdin as input and
// returns stdout. The key is always the last argument.
func runPass(stdin, verb string, args ...string) (string, error) {
	key := args[len(args)-1]
	if err := validPassKey(key); err != nil {
		return "", err
	}
	args[len(args)-1] = passPrefix + key

	cmd := exec.Command("pass", append([]string{verb}, args...)...)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if errors.Is(err, exec.ErrNotFound) {
		return "", ErrPassNotInstalled
	}
	if err != nil {
		// Only stderr is reported: stdout may hold a decrypted secret.
		msg := firstLine(stderr.String())
		if strings.Contains(msg, "is not in the password store") {
			return "", ErrNotFound
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("pass %s: %s", verb, msg)
	}
	return stdout.String(), nil
}

// validPassKey rejects keys that could escape the supawho/ folder or be
// read by pass as a flag.
func validPassKey(key string) error {
	if key == "" || strings.ContainsAny(key, "/\\\x00") || strings.HasPrefix(key, ".") || strings.HasPrefix(key, "-") {
		return fmt.Errorf("invalid account name %q for the pass backend", key)
	}
	return nil
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	return strings.TrimSpace(line)
}
