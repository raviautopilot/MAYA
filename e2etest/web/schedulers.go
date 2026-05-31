package web

import (
	"fmt"
	"time"

	"github.com/tebeka/selenium"
)

// TestSchedulerWorkflow outlines the E2E workflow for creating, viewing,
// and verifying scheduler cron jobs on the MyKanban dashboard.
func (s *WebTestSuite) TestSchedulerWorkflow() {
	// 1. Navigate to Frontend and Login
	s.NavigateTo(s.frontendURL)
	s.ClickElement("[e2e-test-id='login-google-btn']", "Google Login Button")
	s.WaitTillElementFound("[data-testid='mock-login-user-btn']", 10*time.Second)
	s.ClickElement("[data-testid='mock-login-user-btn']", "Test Developer profile selection")
	s.WaitTillElementFound("[e2e-test-id='header']", 10*time.Second)

	// 2. Go to Schedulers Page
	s.NavigateTo(s.frontendURL + "/schedulers")
	s.WaitTillElementFound("[e2e-test-id='schedulers-page']", 10*time.Second)
	s.TakeScreenshot("scheduler_crud_01_page")

	// 3. Click New Scheduler button
	s.ClickElement("[e2e-test-id='scheduler-create-btn']", "New Scheduler button")
	s.WaitTillElementFound("[e2e-test-id='scheduler-form']", 10*time.Second)

	schedName := fmt.Sprintf("E2E Scheduler %d", time.Now().Unix())
	schedNameInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='scheduler-name-input']")
	s.Require().NoError(err)
	err = schedNameInput.SendKeys(schedName)
	s.Require().NoError(err)

	// Select Type (cron)
	typeSelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='scheduler-type-select']")
	s.Require().NoError(err)
	err = typeSelect.Click()
	s.Require().NoError(err)
	typeOption, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='scheduler-type-select'] option[value='cron']")
	s.Require().NoError(err)
	err = typeOption.Click()
	s.Require().NoError(err)

	// Enter Cron Expression
	cronInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='scheduler-cron-input']")
	s.Require().NoError(err)
	err = cronInput.SendKeys("*/5 * * * *")
	s.Require().NoError(err)

	// Select Linked Task Template if available (optional)
	taskSelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='scheduler-linked-task-select']")
	s.Require().NoError(err)
	err = taskSelect.Click()
	s.Require().NoError(err)
	taskOption, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='scheduler-linked-task-select'] option:nth-child(2)")
	if err == nil {
		err = taskOption.Click()
		s.Require().NoError(err)
	}

	s.TakeScreenshot("scheduler_crud_02_form_filled")

	// 4. Submit Scheduler
	s.ClickElement("[e2e-test-id='scheduler-submit-btn']", "Scheduler Submit button")

	// 5. Assert Scheduler Table
	s.CurrentTest.LogStep("Verify Scheduler Table", "INFO", fmt.Sprintf("Waiting for scheduler '%s' in table", schedName))
	schedFound := false
	end := time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		elem, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), '%s')]", schedName))
		if err == nil {
			if displayed, _ := elem.IsDisplayed(); displayed {
				schedFound = true
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().True(schedFound, "Scheduler was not found in the table after submission")
	s.TakeScreenshot("scheduler_crud_03_in_table")

	s.CurrentTest.LogStep("Scheduler Workflow Complete", "PASSED", "Successfully completed and verified the Scheduler creation and validation workflow")
}
