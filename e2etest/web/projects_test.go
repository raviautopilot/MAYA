package web

import (
	"fmt"
	"time"

	"github.com/tebeka/selenium"
)

// TestProjectWorkflow outlines the E2E workflow for creating, viewing,
// editing, and soft-deleting projects on the MyKanban dashboard.
func (s *WebTestSuite) TestProjectWorkflow() {
	// 1. Navigate to Frontend and Login
	s.NavigateTo(s.frontendURL)
	s.ClickElement("[e2e-test-id='login-google-btn']", "Google Login Button")
	s.WaitTillElementFound("[data-testid='mock-login-user-btn']", 10*time.Second)
	s.ClickElement("[data-testid='mock-login-user-btn']", "Test Developer profile selection")
	s.WaitTillElementFound("[e2e-test-id='header']", 10*time.Second)

	// 2. Go to Projects Page
	s.NavigateTo(s.frontendURL + "/projects")
	s.WaitTillElementFound("[e2e-test-id='projects-page']", 10*time.Second)
	s.TakeScreenshot("project_crud_01_page")

	// 3. Click New Project button
	s.ClickElement("[e2e-test-id='project-create-btn']", "New Project button")
	s.WaitTillElementFound("[e2e-test-id='project-form']", 10*time.Second)

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
	err = descInput.SendKeys("Initial E2E test project description")
	s.Require().NoError(err)

	typeSelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-type-select']")
	s.Require().NoError(err)
	err = typeSelect.Click()
	s.Require().NoError(err)
	typeOption, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-type-select'] option[value='personal']")
	s.Require().NoError(err)
	err = typeOption.Click()
	s.Require().NoError(err)

	s.TakeScreenshot("project_crud_02_form_filled")

	// 4. Click Submit
	s.ClickElement("[e2e-test-id='project-submit-btn']", "Project Submit button")

	// 5. Verify in Project Table
	s.CurrentTest.LogStep("Verify Table Row", "INFO", fmt.Sprintf("Waiting for project '%s' to appear in the table", projectName))
	projectFound := false
	end := time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		elem, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), '%s')]", projectName))
		if err == nil {
			if displayed, _ := elem.IsDisplayed(); displayed {
				projectFound = true
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().True(projectFound, "New project row was not found in the table after submission")
	s.TakeScreenshot("project_crud_03_in_table")

	// 6. Click Edit Project
	editBtn, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), '%s')]/ancestor::tr//button[contains(@e2e-test-id, 'project-edit-btn-')]", projectName))
	s.Require().NoError(err)
	err = editBtn.Click()
	s.Require().NoError(err)

	s.WaitTillElementFound("[e2e-test-id='project-form']", 10*time.Second)

	editedProjectName := projectName + " (Edited)"
	nameInput, err = s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-name-input']")
	s.Require().NoError(err)
	err = nameInput.Clear()
	s.Require().NoError(err)
	err = nameInput.SendKeys(editedProjectName)
	s.Require().NoError(err)

	descInput, err = s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-description-input']")
	s.Require().NoError(err)
	err = descInput.Clear()
	s.Require().NoError(err)
	err = descInput.SendKeys("Updated project description via E2E Edit workflow")
	s.Require().NoError(err)

	typeSelect, err = s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-type-select']")
	s.Require().NoError(err)
	err = typeSelect.Click()
	s.Require().NoError(err)
	typeOption, err = s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-type-select'] option[value='professional']")
	s.Require().NoError(err)
	err = typeOption.Click()
	s.Require().NoError(err)

	// Update estimated hours
	estHoursInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-estimated-hours-input']")
	s.Require().NoError(err)
	err = estHoursInput.Clear()
	s.Require().NoError(err)
	err = estHoursInput.SendKeys("80")
	s.Require().NoError(err)

	s.TakeScreenshot("project_crud_04_edit_form_filled")

	// Submit edit
	s.ClickElement("[e2e-test-id='project-submit-btn']", "Project Update Submit button")

	// Verify updated project in table
	s.CurrentTest.LogStep("Verify Edit", "INFO", fmt.Sprintf("Waiting for project '%s' to appear in the table", editedProjectName))
	editFound := false
	end = time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		elem, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), '%s')]", editedProjectName))
		if err == nil {
			if displayed, _ := elem.IsDisplayed(); displayed {
				editFound = true
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().True(editFound, "Edited project row was not found in the table")
	s.TakeScreenshot("project_crud_05_edited_in_table")

	// 7. Click Delete Project
	deleteBtn, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), '%s')]/ancestor::tr//button[contains(@e2e-test-id, 'project-delete-btn-')]", editedProjectName))
	s.Require().NoError(err)
	err = deleteBtn.Click()
	s.Require().NoError(err)

	s.WaitTillElementFound("[e2e-test-id='project-delete-dialog-confirm-btn']", 10*time.Second)
	s.TakeScreenshot("project_crud_06_delete_dialog")

	// 8. Confirm deletion in Dialog
	s.ClickElement("[e2e-test-id='project-delete-dialog-confirm-btn']", "Delete confirmation button")

	// Verify deleted
	s.CurrentTest.LogStep("Verify Deletion", "INFO", fmt.Sprintf("Waiting for project '%s' to be removed from the table", editedProjectName))
	deleted := false
	end = time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		_, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), '%s')]", editedProjectName))
		if err != nil {
			deleted = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().True(deleted, "Project was not deleted from the table")
	s.TakeScreenshot("project_crud_07_deleted_successfully")
	s.CurrentTest.LogStep("Project CRUD Complete", "PASSED", "Successfully executed and verified complete Project CRUD workflow")
}
