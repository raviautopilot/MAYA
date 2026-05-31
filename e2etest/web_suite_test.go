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

func (s *WebTestSuite) TestCreateProject() {
	// 1. Navigate to Frontend
	s.NavigateTo(s.frontendURL)
	s.TakeScreenshot("project_create_01_landing")

	// 2. Click Google Login -> Trigger Mock OAuth flow
	s.ClickElement("[e2e-test-id='login-google-btn']", "Google Login Button")
	
	// Wait for redirect to mock consent portal
	s.WaitTillElementFound("[data-testid='mock-login-user-btn']", 10*time.Second)
	s.TakeScreenshot("project_create_02_mock_consent")

	// 3. Authorize Mock Profile -> Callback -> Dashboard redirect
	s.ClickElement("[data-testid='mock-login-user-btn']", "Test Developer profile selection")
	
	// Wait to be redirected back to frontend dashboard
	s.WaitTillElementFound("[e2e-test-id='header']", 10*time.Second)
	s.TakeScreenshot("project_create_03_dashboard")

	// 4. Click Quick Action button for New Project
	s.ClickElement("[e2e-test-id='dashboard-quick-new-project']", "New Project Quick Action button")

	// 5. Wait for Projects page to load
	s.WaitTillElementFound("[e2e-test-id='projects-page']", 10*time.Second)
	s.TakeScreenshot("project_create_04_projects_page")

	// 6. Click New Project button to open modal
	s.ClickElement("[e2e-test-id='project-create-btn']", "New Project button")
	s.WaitTillElementFound("[e2e-test-id='project-form']", 10*time.Second)

	// 7. Fill the name and description
	projectName := fmt.Sprintf("E2E Project %d", time.Now().Unix())
	
	nameInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-name-input']")
	s.Require().NoError(err)
	err = nameInput.Clear()
	s.Require().NoError(err)
	err = nameInput.SendKeys(projectName)
	s.Require().NoError(err)

	descInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-description-input']")
	s.Require().NoError(err)
	err = descInput.Clear()
	s.Require().NoError(err)
	err = descInput.SendKeys("A professional workspace created by Go E2E tests")
	s.Require().NoError(err)

	// 8. Select type as professional
	typeSelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-type-select']")
	s.Require().NoError(err)
	err = typeSelect.Click()
	s.Require().NoError(err)

	typeOption, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-type-select'] option[value='professional']")
	s.Require().NoError(err)
	err = typeOption.Click()
	s.Require().NoError(err)

	s.TakeScreenshot("project_create_05_form_filled")

	// 9. Submit the form
	s.ClickElement("[e2e-test-id='project-submit-btn']", "Project Submit button")

	// 10. Verify project appears in the table
	s.CurrentTest.LogStep("Verify Table Row", "INFO", fmt.Sprintf("Waiting for project '%s' to appear in the table", projectName))
	projectFound := false
	end := time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		elem, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//span[contains(text(), '%s')]", projectName))
		if err == nil {
			if displayed, _ := elem.IsDisplayed(); displayed {
				projectFound = true
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().True(projectFound, "New project row was not found in the table after submission")

	s.TakeScreenshot("project_create_06_project_in_table")
	s.CurrentTest.LogStep("Project Creation Complete", "PASSED", fmt.Sprintf("Successfully verified creation of project '%s' in the projects list table", projectName))
}

func TestRunWebSuite(t *testing.T) {
	suite.Run(t, new(WebTestSuite))
}
