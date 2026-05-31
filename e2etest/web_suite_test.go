package e2etest

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tebeka/selenium"
)

type WebTestSuite struct {
	BaseWebSuite
	frontendURL string
}

func (s *WebTestSuite) SetupSuite() {
	s.BaseWebSuite.SetupSuite()
	s.frontendURL = GlobalConfig.FrontendURL
}

func (s *WebTestSuite) TestFrontendWorkflow() {
	// 1. Navigate to Frontend
	s.NavigateTo(s.frontendURL)
	s.TakeScreenshot("01_landing_page")

	// 2. Assert Logged Out state
	s.CurrentTest.LogStep("Verify Landing Page State", "INFO", "Validating presence of login card and elements")
	loginCard, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='login-card']")
	s.Require().NoError(err, "Could not find login card. Is frontend running?")
	
	displayed, err := loginCard.IsDisplayed()
	s.Require().NoError(err)
	s.True(displayed, "Login card should be visible")

	_, err = s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='login-google-btn']")
	s.Require().NoError(err, "Google sign-in button missing")

	// Check health status badge (optional, skipped if missing)
	badge, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='health-status-badge']")
	if err == nil {
		text, _ := badge.Text()
		s.CurrentTest.LogStep("Check Health Status Badge", "PASSED", fmt.Sprintf("Health status badge reports: %s", text))
	} else {
		s.CurrentTest.LogStep("Check Health Status Badge", "INFO", "No health status badge found on landing page")
	}

	// 3. Click Google Login -> Trigger Mock OAuth flow
	s.ClickElement("[e2e-test-id='login-google-btn']", "Google Login Button")
	
	// Wait for redirect to mock consent portal
	s.WaitTillElementFound("[data-testid='mock-login-user-btn']", 10*time.Second)
	s.TakeScreenshot("02_mock_consent_portal")

	// 4. Authorize Mock Profile -> Callback -> Dashboard redirect
	s.ClickElement("[data-testid='mock-login-user-btn']", "Test Developer profile selection button")
	
	// Wait to be redirected back to frontend dashboard
	s.WaitTillElementFound("[e2e-test-id='header']", 10*time.Second)
	s.TakeScreenshot("03_dashboard_page")

	// 5. Assert Logged-in Dashboard state
	s.CurrentTest.LogStep("Verify Dashboard Components", "INFO", "Validating presence of cards on dashboard")
	
	projectCard, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='dashboard-projects-card']")
	s.Require().NoError(err, "Could not find projects card on dashboard")
	visible, _ := projectCard.IsDisplayed()
	s.True(visible, "Projects card should be visible")

	tasksCard, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='dashboard-tasks-card']")
	s.Require().NoError(err, "Could not find tasks card on dashboard")
	visible, _ = tasksCard.IsDisplayed()
	s.True(visible, "Tasks card should be visible")

	dashUserInfo, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='header-user-info']")
	if err == nil {
		text, _ := dashUserInfo.Text()
		s.CurrentTest.LogStep("Check Dashboard User Info", "PASSED", fmt.Sprintf("Dashboard reports user: %s", text))
	}

	// 6. Logout assertion
	s.CurrentTest.LogStep("Wait for toast to clear", "INFO", "Waiting for toast notification to clear before clicking logout")
	time.Sleep(5 * time.Second)
	s.ClickElement("[e2e-test-id='header-logout-btn']", "Logout Button")

	// Wait to return to logged out screen
	s.WaitTillElementFound("[e2e-test-id='login-card']", 10*time.Second)
	s.TakeScreenshot("04_returned_to_landing")

	s.CurrentTest.LogStep("Workflow Complete", "PASSED", "Successfully traversed login, mock authorization, dashboard verification, and logout workflow")
}

func TestRunWebSuite(t *testing.T) {
	suite.Run(t, new(WebTestSuite))
}
