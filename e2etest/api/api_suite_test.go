package api

import (
	"e2etest"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	e2etest.BaseAPISuite
	backendURL string
}

func (s *APITestSuite) SetupSuite() {
	s.BaseAPISuite.SetupSuite()
	s.backendURL = e2etest.GlobalConfig.BackendURL
}

func TestMain(m *testing.M) {
	if err := e2etest.InitTestEnvironment(); err != nil {
		println("Error initializing E2E test environment: ", err.Error())
		os.Exit(1)
	}

	e2etest.LogExecution("================================================================================")
	e2etest.LogExecution("API SUITE RUN STARTED | Package: api")
	e2etest.LogExecution("================================================================================")

	exitCode := m.Run()

	e2etest.WriteReport()
	os.Exit(exitCode)
}

func TestRunAPISuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
