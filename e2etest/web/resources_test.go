package web

import (
	"fmt"
	"time"

	"github.com/tebeka/selenium"
)

// TestResourceWorkflow outlines the E2E workflow for creating, viewing,
// and verifying resources on the MyKanban dashboard.
func (s *WebTestSuite) TestResourceWorkflow() {
	// 1. Navigate to Frontend and Login
	s.NavigateTo(s.frontendURL)
	s.ClickElement("[e2e-test-id='login-google-btn']", "Google Login Button")
	s.WaitTillElementFound("[data-testid='mock-login-user-btn']", 10*time.Second)
	s.ClickElement("[data-testid='mock-login-user-btn']", "Test Developer profile selection")
	s.WaitTillElementFound("[e2e-test-id='header']", 10*time.Second)

	// 2. Go to Resources Page
	s.NavigateTo(s.frontendURL + "/resources")
	s.WaitTillElementFound("[e2e-test-id='resources-page']", 10*time.Second)
	s.TakeScreenshot("resource_crud_01_page")

	// 3. Click New Resource button
	s.ClickElement("[e2e-test-id='resource-create-btn']", "New Resource button")
	s.WaitTillElementFound("[e2e-test-id='resource-form']", 10*time.Second)

	resName := fmt.Sprintf("E2E Resource %d", time.Now().Unix())
	resNameInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='resource-name-input']")
	s.Require().NoError(err)
	err = resNameInput.SendKeys(resName)
	s.Require().NoError(err)

	// Select Type (Global)
	typeSelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='resource-type-select']")
	s.Require().NoError(err)
	err = typeSelect.Click()
	s.Require().NoError(err)
	typeOption, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='resource-type-select'] option[value='Global']")
	s.Require().NoError(err)
	err = typeOption.Click()
	s.Require().NoError(err)

	// Fill Linked Items
	linkedItemsInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='resource-linked-items-input']")
	s.Require().NoError(err)
	err = linkedItemsInput.SendKeys("item-uuid-1, item-uuid-2")
	s.Require().NoError(err)

	s.TakeScreenshot("resource_crud_02_form_filled")

	// 4. Submit Resource
	s.ClickElement("[e2e-test-id='resource-submit-btn']", "Resource Submit button")

	// 5. Assert Resource Table
	s.CurrentTest.LogStep("Verify Resource Table", "INFO", fmt.Sprintf("Waiting for resource '%s' in table", resName))
	resFound := false
	end := time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		elem, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), '%s')]", resName))
		if err == nil {
			if displayed, _ := elem.IsDisplayed(); displayed {
				resFound = true
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().True(resFound, "Resource was not found in the table after submission")
	s.TakeScreenshot("resource_crud_03_in_table")

	s.CurrentTest.LogStep("Resource Workflow Complete", "PASSED", "Successfully completed and verified the Resource creation and validation workflow")
}
