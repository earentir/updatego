package local

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

type statusKind int

const (
	statusInfo statusKind = iota
	statusOK
	statusWarn
	statusFail
)

type statusRow struct {
	label  string
	kind   statusKind
	detail string
}

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func paint(enabled bool, code, s string) string {
	if !enabled || code == "" {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}

func statusMark(kind statusKind, enabled bool) string {
	switch kind {
	case statusOK:
		return paint(enabled, "32", "ok")
	case statusWarn:
		return paint(enabled, "33", "warn")
	case statusFail:
		return paint(enabled, "31", "fail")
	default:
		return paint(enabled, "90", "-")
	}
}

func printStatusTable(rows []statusRow) {
	enabled := colorEnabled()
	labelWidth := 0
	for _, r := range rows {
		if n := utf8.RuneCountInString(r.label); n > labelWidth {
			labelWidth = n
		}
	}
	// "fail"/"warn"/"ok"/"-" ; pad on visible width only (ANSI ignored).
	markWidth := 4

	for _, r := range rows {
		mark := statusMark(r.kind, enabled)
		pad := strings.Repeat(" ", labelWidth-utf8.RuneCountInString(r.label))
		visibleMark := visibleStatusMark(r.kind)
		markPad := strings.Repeat(" ", markWidth-utf8.RuneCountInString(visibleMark))
		fmt.Printf("%s%s  %s%s  %s\n", r.label, pad, mark, markPad, paintDetail(enabled, r.kind, r.detail))
	}
}

func visibleStatusMark(kind statusKind) string {
	switch kind {
	case statusOK:
		return "ok"
	case statusWarn:
		return "warn"
	case statusFail:
		return "fail"
	default:
		return "-"
	}
}

func paintDetail(enabled bool, kind statusKind, detail string) string {
	switch kind {
	case statusOK:
		return paint(enabled, "32", detail)
	case statusWarn:
		return paint(enabled, "33", detail)
	case statusFail:
		return paint(enabled, "31", detail)
	default:
		return detail
	}
}
