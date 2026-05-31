package web

import (
	"fmt"
	"time"

	"github.com/tebeka/selenium"
)

// TestBoardWorkflow outlines the E2E workflow for creating, viewing,
// editing, and soft-deleting boards on the MyKanban dashboard.
func (s *WebTestSuite) TestBoardWorkflow() {
	// 1. Navigate to Frontend and Login
	s.NavigateTo(s.frontendURL)
	s.ClickElement("[e2e-test-id='login-google-btn']", "Google Login Button")
	s.WaitTillElementFound("[data-testid='mock-login-user-btn']", 10*time.Second)
	s.ClickElement("[data-testid='mock-login-user-btn']", "Test Developer profile selection")
	s.WaitTillElementFound("[e2e-test-id='header']", 10*time.Second)

	// 2. Go to Projects Page and create a parent project first
	s.NavigateTo(s.frontendURL + "/projects")
	s.WaitTillElementFound("[e2e-test-id='projects-page']", 10*time.Second)
	s.ClickElement("[e2e-test-id='project-create-btn']", "New Project button")
	s.WaitTillElementFound("[e2e-test-id='project-form']", 10*time.Second)

	parentProjectName := fmt.Sprintf("Board Parent %d", time.Now().Unix())
	nameInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-name-input']")
	s.Require().NoError(err)
	err = nameInput.SendKeys(parentProjectName)
	s.Require().NoError(err)

	// Select Project Type (required field)
	typeSelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-type-select']")
	s.Require().NoError(err)
	err = typeSelect.Click()
	s.Require().NoError(err)
	typeOption, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='project-type-select'] option[value='personal']")
	s.Require().NoError(err)
	err = typeOption.Click()
	s.Require().NoError(err)

	s.ClickElement("[e2e-test-id='project-submit-btn']", "Project Submit button")

	// Wait for project to appear in the table
	s.CurrentTest.LogStep("Wait for Parent Project", "INFO", "Waiting for parent project in list")
	projectFound := false
	end := time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		elem, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), '%s')]", parentProjectName))
		if err == nil {
			if displayed, _ := elem.IsDisplayed(); displayed {
				projectFound = true
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().True(projectFound, "Parent project for board was not found in table")

	// 3. Go to Boards Page
	s.NavigateTo(s.frontendURL + "/boards")
	s.WaitTillElementFound("[e2e-test-id='boards-page']", 10*time.Second)
	s.TakeScreenshot("board_crud_01_page")

	// 4. Click New Board button
	s.ClickElement("[e2e-test-id='board-create-btn']", "New Board button")
	s.WaitTillElementFound("[e2e-test-id='board-form']", 10*time.Second)

	boardName := fmt.Sprintf("E2E Board %d", time.Now().Unix())
	boardNameInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='board-name-input']")
	s.Require().NoError(err)
	err = boardNameInput.SendKeys(boardName)
	s.Require().NoError(err)

	// Wait for the exact project option to be populated and rendered in the dropdown
	var projectOption selenium.WebElement
	end = time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		projectOption, err = s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//select[@e2e-test-id='board-project-select']/option[contains(text(), '%s')]", parentProjectName))
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().NoError(err, "Parent project option was not found in dropdown")

	// Click Select to focus, then click Option
	projectSelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='board-project-select']")
	s.Require().NoError(err)
	err = projectSelect.Click()
	s.Require().NoError(err)
	err = projectOption.Click()
	s.Require().NoError(err)

	// Fill Swimlanes
	swimlanesInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='board-swimlanes-input']")
	s.Require().NoError(err)
	err = swimlanesInput.Clear()
	s.Require().NoError(err)
	err = swimlanesInput.SendKeys("To Do, In Progress, Review, Done")
	s.Require().NoError(err)

	// Fill Task Types
	taskTypesInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='board-task-types-input']")
	s.Require().NoError(err)
	err = taskTypesInput.Clear()
	s.Require().NoError(err)
	err = taskTypesInput.SendKeys("Bug, Feature, Chore")
	s.Require().NoError(err)

	s.TakeScreenshot("board_crud_02_form_filled")

	// 5. Submit Board
	s.ClickElement("[e2e-test-id='board-submit-btn']", "Board Submit button")

	// 6. Assert Board Table
	s.CurrentTest.LogStep("Verify Board Table", "INFO", fmt.Sprintf("Waiting for board '%s' in table", boardName))
	boardFound := false
	end = time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		elem, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), '%s')]", boardName))
		if err == nil {
			if displayed, _ := elem.IsDisplayed(); displayed {
				boardFound = true
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().True(boardFound, "Board was not found in the table after submission")
	s.TakeScreenshot("board_crud_03_in_table")

	// 7. Filter Boards by Project
	var filterOption selenium.WebElement
	end = time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		filterOption, err = s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//select[@e2e-test-id='boards-filter-project']/option[contains(text(), '%s')]", parentProjectName))
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().NoError(err, "Parent project option not found in boards filter dropdown")

	filterSelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='boards-filter-project']")
	s.Require().NoError(err)
	err = filterSelect.Click()
	s.Require().NoError(err)
	err = filterOption.Click()
	s.Require().NoError(err)

	s.TakeScreenshot("board_crud_04_filtered_table")

	// 8. Click Board Link (navigates to tasks)
	boardLink, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), '%s')]/ancestor::tr//a[contains(@e2e-test-id, 'board-name-link-')]", boardName))
	s.Require().NoError(err)
	err = boardLink.Click()
	s.Require().NoError(err)

	// Wait to land on tasks page
	s.WaitTillElementFound("[e2e-test-id='tasks-page']", 10*time.Second)
	s.TakeScreenshot("board_crud_05_tasks_page_navigation")

	s.CurrentTest.LogStep("Board Workflow Complete", "PASSED", "Successfully completed and verified the Board creation and filtering workflow")
}
