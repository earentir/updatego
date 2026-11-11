package local

import (
	"updatego/config"
	"updatego/utils"
)

// RunTest runs install/update preflight tests (permissions only, no downloads).
func RunTest() error {
	goFullPath := config.GlobalConfig.GoFullPath
	rows := make([]checkRow, 0, 8)

	if err := utils.RefuseWindows(); err != nil {
		rows = append(rows, checkRow{test: "Operating system is supported", kind: statusFail, detail: err.Error()})
		printCheckTable(rows)
		printCheckSummary(rows, "tests")
		return err
	}
	rows = append(rows, checkRow{test: "Operating system is supported", kind: statusOK, detail: "yes"})

	rows = append(rows, checkRow{
		test:   "Configuration load",
		kind:   statusOK,
		detail: "extract root " + config.GlobalConfig.GoExtractPathRoot,
	})

	for _, c := range utils.InspectInstallPermissions(
		config.GlobalConfig.GoExtractPathRoot,
		goFullPath,
		config.GlobalConfig.TempDir,
	) {
		test := c.Test
		if test == "" {
			test = c.Name
		}
		addCheck(&rows, test, c.Pass, c.Detail)
	}

	return finishReport(rows, "tests")
}
