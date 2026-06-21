package directives

import "strings"

type RunMode int

const (
	ModeInterpret	RunMode	= iota
	ModeCompile
)

type Config struct {
	Mode		RunMode
	NonStrict	bool
	WarnAsError	bool
}

func Extract(source string) Config {
	cfg := Config{Mode: ModeInterpret}
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "><>") {
			break
		}
		rest := strings.TrimSpace(trimmed[3:])
		if !strings.HasPrefix(rest, "!") {
			continue
		}
		parts := strings.SplitN(rest[1:], " ", 2)
		cmd := parts[0]
		switch cmd {
		case "interpret":
			cfg.Mode = ModeInterpret
		case "compile":
			cfg.Mode = ModeCompile
		case "non-strict":
			cfg.NonStrict = true
		case "warn-as-error":
			cfg.WarnAsError = true
		}
	}
	return cfg
}
