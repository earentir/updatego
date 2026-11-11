package local

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

type statusKind int

const (
	statusOK statusKind = iota
	statusWarn
	statusFail
)

type checkRow struct {
	test   string
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

func printCheckTable(rows []checkRow) {
	enabled := colorEnabled()
	testWidth := utf8.RuneCountInString("Test")
	statusHeader := "Status"
	statusWidth := utf8.RuneCountInString(statusHeader)
	for _, r := range rows {
		if n := utf8.RuneCountInString(r.test); n > testWidth {
			testWidth = n
		}
		if n := utf8.RuneCountInString(visibleStatusMark(r.kind)); n > statusWidth {
			statusWidth = n
		}
	}

	fmt.Printf("%s%s  %s%s  %s\n",
		"Test", strings.Repeat(" ", testWidth-utf8.RuneCountInString("Test")),
		statusHeader, strings.Repeat(" ", statusWidth-utf8.RuneCountInString(statusHeader)),
		"Details",
	)

	for _, r := range rows {
		mark := statusMark(r.kind, enabled)
		testPad := strings.Repeat(" ", testWidth-utf8.RuneCountInString(r.test))
		visibleMark := visibleStatusMark(r.kind)
		statusPad := strings.Repeat(" ", statusWidth-utf8.RuneCountInString(visibleMark))
		fmt.Printf("%s%s  %s%s  %s\n", r.test, testPad, mark, statusPad, paintDetail(enabled, r.kind, r.detail))
	}
}

func printCheckSummary(rows []checkRow, unit string) {
	var pass, warn, fail int
	for _, r := range rows {
		switch r.kind {
		case statusOK:
			pass++
		case statusWarn:
			warn++
		case statusFail:
			fail++
		}
	}
	enabled := colorEnabled()
	total := pass + warn + fail
	fmt.Println()
	if !enabled {
		line := fmt.Sprintf("%d %s: %d passed", total, unit, pass)
		if warn > 0 {
			line += fmt.Sprintf(", %d warned", warn)
		}
		if fail > 0 {
			line += fmt.Sprintf(", %d failed", fail)
		}
		fmt.Println(line)
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d %s: ", total, unit)
	b.WriteString(paint(enabled, "32", fmt.Sprintf("%d passed", pass)))
	if warn > 0 {
		b.WriteString(", ")
		b.WriteString(paint(enabled, "33", fmt.Sprintf("%d warned", warn)))
	}
	if fail > 0 {
		b.WriteString(", ")
		b.WriteString(paint(enabled, "31", fmt.Sprintf("%d failed", fail)))
	}
	fmt.Println(b.String())
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
