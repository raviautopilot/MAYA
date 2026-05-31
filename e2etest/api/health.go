package api

import (
	"encoding/json"
	"net/http"
)

func (s *APITestSuite) TestAPIHealthEndpoint() {
	s.CurrentTest.LogStep("Prep Request", "INFO", "Preparing API request for health check endpoint")
	
	req, err := http.NewRequest("GET", s.backendURL+"/health", nil)
	s.Require().NoError(err)

	resp, bodyBytes, err := s.LogAndDoRequest(req, nil)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	s.CurrentTest.LogStep("Decode Response", "INFO", "Decoding health check response body")
	var body map[string]interface{}
	err = json.Unmarshal(bodyBytes, &body)
	s.Require().NoError(err, "Failed to decode health check response JSON")

	// Assertions
	s.CurrentTest.LogStep("Assert Schema", "INFO", "Validating health response schemas and keys")
	status, ok := body["status"].(string)
	s.True(ok, "Expected status as string")
	s.Equal("UP", status)

	env, ok := body["environment"].(string)
	s.True(ok, "Expected environment as string")
	s.NotEmpty(env)

	memory, ok := body["memory"].(map[string]interface{})
	s.True(ok, "Expected memory stats block")
	s.Contains(memory, "alloc_mb")

	deps, ok := body["dependencies"].(map[string]interface{})
	s.True(ok, "Expected dependencies block")
	s.Contains(deps, "google_oauth")
	
	s.CurrentTest.LogStep("Assertions Completed", "PASSED", "All health check fields validated successfully")
}
