package api

import (
	"encoding/json"
	"net/http"
)

func (s *APITestSuite) TestAPISwaggerEndpoint() {
	s.CurrentTest.LogStep("Prep HTML Request", "INFO", "Querying Swagger UI HTML page")
	reqHTML, err := http.NewRequest("GET", s.backendURL+"/swagger/index.html", nil)
	s.Require().NoError(err)
	
	respHTML, _, err := s.LogAndDoRequest(reqHTML, nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, respHTML.StatusCode)

	s.CurrentTest.LogStep("Prep JSON Request", "INFO", "Querying Swagger JSON API document spec")
	reqJSON, err := http.NewRequest("GET", s.backendURL+"/swagger/doc.json", nil)
	s.Require().NoError(err)

	respJSON, jsonBytes, err := s.LogAndDoRequest(reqJSON, nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, respJSON.StatusCode)

	s.CurrentTest.LogStep("Decode Swagger Spec", "INFO", "Decoding Swagger specification document")
	var swagDoc map[string]interface{}
	err = json.Unmarshal(jsonBytes, &swagDoc)
	s.Require().NoError(err, "Failed to decode swagger JSON docs")

	s.CurrentTest.LogStep("Assert Swagger Spec", "INFO", "Validating schema attributes")
	s.Equal("2.0", swagDoc["swagger"])
	
	info, ok := swagDoc["info"].(map[string]interface{})
	s.True(ok, "Expected swagger info block")
	s.Contains(info, "title")

	s.CurrentTest.LogStep("Assertions Completed", "PASSED", "Swagger UI and spec documents successfully verified")
}
