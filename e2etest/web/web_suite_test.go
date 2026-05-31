package web

import (
	"e2etest"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type WebTestSuite struct {
	e2etest.BaseWebSuite
	frontendURL string
}

func (s *WebTestSuite) SetupSuite() {
	s.BaseWebSuite.SetupSuite()
	s.frontendURL = e2etest.GlobalConfig.FrontendURL
}

func TestMain(m *testing.M) {
	if err := e2etest.InitTestEnvironment(); err != nil {
		println("Error initializing E2E test environment: ", err.Error())
		os.Exit(1)
	}

	e2etest.LogExecution("================================================================================")
	e2etest.LogExecution("WEB SUITE RUN STARTED | Package: web")
	e2etest.LogExecution("================================================================================")

	exitCode := m.Run()

	e2etest.WriteReport()
	os.Exit(exitCode)
}

func TestRunWebSuite(t *testing.T) {
	suite.Run(t, new(WebTestSuite))
}
