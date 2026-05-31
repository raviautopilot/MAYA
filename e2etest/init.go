package e2etest

import (
	"fmt"
	"os"
	"path/filepath"
)

// GlobalConfig holds test fixtures and options loaded from e2e-test.json
var GlobalConfig *Config
var initialized bool

// InitTestEnvironment locates and loads e2e-test.json, initializes the report dashboard,
// and redirects ChromeDriver logs to deep-debug.log.
func InitTestEnvironment() error {
	if initialized {
		return nil
	}

	// Locate e2e-test.json by walking up the directory tree
	configPath := "e2e-test.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if _, err := os.Stat("../e2e-test.json"); err == nil {
			configPath = "../e2e-test.json"
		} else if _, err := os.Stat("../../e2e-test.json"); err == nil {
			configPath = "../../e2e-test.json"
		}
	}

	var err error
	GlobalConfig, err = LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config %s: %w", configPath, err)
	}

	report := GetReport()

	// Ensure E2E_RUN_DIR is set to a consistent root-relative path
	runDir := os.Getenv("E2E_RUN_DIR")
	if runDir == "" {
		baseDir := "reports"
		// If running in a subpackage (web/ or api/), configPath was resolved to parent "../e2e-test.json"
		if _, err := os.Stat("e2e-test.json"); os.IsNotExist(err) {
			baseDir = "../reports"
		}
		runDir = filepath.Join(baseDir, "run_"+report.StartTime.Format("2006-01-02_15-04-05"))
		os.Setenv("E2E_RUN_DIR", runDir)
	}

	if err := os.MkdirAll(runDir, 0755); err != nil {
		return fmt.Errorf("failed to create run dir %s: %w", runDir, err)
	}

	// Open execution.log
	execLogFile, err := os.OpenFile(filepath.Join(runDir, "execution.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		ExecutionLogWriter = execLogFile
	}

	// Open deep-debug.log
	deepDebugFile, err := os.OpenFile(filepath.Join(runDir, "deep-debug.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		DeepDebugWriter = deepDebugFile
	}

	initialized = true
	return nil
}

// WriteReport finalizes the active test run statistics and outputs the interactive HTML dashboard.
func WriteReport() {
	if !initialized || GlobalConfig == nil {
		return
	}
	report := GetReport()
	report.Finalize()

	runDir := report.GetRunDirectory()
	reportFilePath := filepath.Join(runDir, GlobalConfig.ReportPath)
	if err := report.GenerateHTML(reportFilePath); err != nil {
		println("Error generating E2E HTML test report: ", err.Error())
	} else {
		println("Interactive E2E test report generated successfully: ", reportFilePath)
	}
}
