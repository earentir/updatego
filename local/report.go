package local

// ReportExit is returned after the report was printed; it triggers exit 1 with no extra message.
type ReportExit struct{}

func (ReportExit) Error() string { return "" }

func addCheck(rows *[]checkRow, test string, pass bool, detail string) {
	kind := statusOK
	if !pass {
		kind = statusFail
	}
	*rows = append(*rows, checkRow{test: test, kind: kind, detail: detail})
}

func addCheckWarn(rows *[]checkRow, test, detail string) {
	*rows = append(*rows, checkRow{test: test, kind: statusWarn, detail: detail})
}

func addCheckInfo(rows *[]checkRow, test, detail string) {
	*rows = append(*rows, checkRow{test: test, kind: statusOK, detail: detail})
}

func finishReport(rows []checkRow, unit string) error {
	printCheckTable(rows)
	printCheckSummary(rows, unit)
	for _, r := range rows {
		if r.kind == statusFail {
			return ReportExit{}
		}
	}
	return nil
}
