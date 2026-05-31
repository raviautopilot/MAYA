package web

import (
	"fmt"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

// TestTaskWorkflow outlines the E2E workflow for creating, viewing,
// editing, deleting, and transitioning tasks on the MyKanban dashboard.
func (s *WebTestSuite) TestTaskWorkflow() {
	// 1. Navigate to Frontend and Login
	s.NavigateTo(s.frontendURL)
	s.ClickElement("[e2e-test-id='login-google-btn']", "Google Login Button")
	s.WaitTillElementFound("[data-testid='mock-login-user-btn']", 10*time.Second)
	s.ClickElement("[data-testid='mock-login-user-btn']", "Test Developer profile selection")
	s.WaitTillElementFound("[e2e-test-id='header']", 10*time.Second)

	// 2. Create parent Project, Board, and Resource via backend API (fast & robust)
	parentProjectName := fmt.Sprintf("Task Parent Proj %d", time.Now().Unix())
	projID, err := s.CreateProjectViaAPI(parentProjectName, "Parent project for Task E2E test", "personal")
	s.Require().NoError(err)

	parentBoardName := fmt.Sprintf("Task Parent Board %d", time.Now().Unix())
	_, err = s.CreateBoardViaAPI(projID, parentBoardName, []string{"To Do", "In Progress", "Done"}, []string{"Bug", "Feature", "Chore"})
	s.Require().NoError(err)

	resourceName := fmt.Sprintf("Assignee %d", time.Now().Unix())
	_, err = s.CreateResourceViaAPI(resourceName, "Global", []string{})
	s.Require().NoError(err)

	// 3. Go to Tasks Page
	s.NavigateTo(s.frontendURL + "/tasks")
	s.WaitTillElementFound("[e2e-test-id='tasks-page']", 10*time.Second)
	s.TakeScreenshot("task_crud_01_page")

	// 5. Go to Tasks Page
	s.NavigateTo(s.frontendURL + "/tasks")
	s.WaitTillElementFound("[e2e-test-id='tasks-page']", 10*time.Second)
	s.TakeScreenshot("task_crud_01_page")

	// 6. Click New Task button
	s.ClickElement("[e2e-test-id='task-create-btn']", "New Task button")
	s.WaitTillElementFound("[e2e-test-id='task-form']", 10*time.Second)

	taskTitle := fmt.Sprintf("E2E Task %d", time.Now().Unix())
	taskTitleInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-title-input']")
	s.Require().NoError(err)
	err = taskTitleInput.SendKeys(taskTitle)
	s.Require().NoError(err)

	taskDescInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-description-input']")
	s.Require().NoError(err)
	err = taskDescInput.SendKeys("A comprehensive task created by E2E test suite")
	s.Require().NoError(err)

	// Wait for Board option to load in dropdown
	var boardOption selenium.WebElement
	end := time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		boardOption, err = s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//select[@e2e-test-id='task-board-select']/option[contains(text(), '%s')]", parentBoardName))
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().NoError(err, "Parent board option was not found in dropdown")

	taskBoardSelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-board-select']")
	s.Require().NoError(err)
	err = taskBoardSelect.Click()
	s.Require().NoError(err)
	err = boardOption.Click()
	s.Require().NoError(err)

	// Wait and Select Swimlane
	swimlaneSelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-swimlane-select']")
	s.Require().NoError(err)
	err = swimlaneSelect.Click()
	s.Require().NoError(err)
	swimlaneOption, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-swimlane-select'] option[value='To Do']")
	s.Require().NoError(err)
	err = swimlaneOption.Click()
	s.Require().NoError(err)

	// Wait and Select Task Type
	typeSelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-type-select']")
	s.Require().NoError(err)
	err = typeSelect.Click()
	s.Require().NoError(err)
	typeOption, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-type-select'] option[value='Feature']")
	s.Require().NoError(err)
	err = typeOption.Click()
	s.Require().NoError(err)

	// Wait and Select Priority
	prioritySelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-priority-select']")
	s.Require().NoError(err)
	err = prioritySelect.Click()
	s.Require().NoError(err)
	priorityOption, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-priority-select'] option[value='High']")
	s.Require().NoError(err)
	err = priorityOption.Click()
	s.Require().NoError(err)

	// Wait and Select Assignee
	var assigneeOption selenium.WebElement
	end = time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		assigneeOption, err = s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//select[@e2e-test-id='task-assignee-select']/option[contains(text(), '%s')]", resourceName))
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().NoError(err, "Assignee resource option was not found in dropdown")

	assigneeSelect, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-assignee-select']")
	s.Require().NoError(err)
	err = assigneeSelect.Click()
	s.Require().NoError(err)
	err = assigneeOption.Click()
	s.Require().NoError(err)

	// Estimation
	taskEstInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-estimation-input']")
	s.Require().NoError(err)
	err = taskEstInput.SendKeys("60")
	s.Require().NoError(err)

	// Cost
	taskCostInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-cost-input']")
	s.Require().NoError(err)
	err = taskCostInput.SendKeys("150.00")
	s.Require().NoError(err)

	// Add a Reminder
	s.ClickElement("[e2e-test-id='task-add-reminder-btn']", "Add Reminder button")
	s.WaitTillElementFound("[e2e-test-id='task-reminder-time-0']", 5*time.Second)

	reminderTimeInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-reminder-time-0']")
	s.Require().NoError(err)
	_, err = s.WD.ExecuteScript(`
		var input = arguments[0];
		var val = arguments[1];
		var setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
		setter.call(input, val);
		input.dispatchEvent(new Event('input', { bubbles: true }));
		input.dispatchEvent(new Event('change', { bubbles: true }));
	`, []interface{}{reminderTimeInput, "2026-06-15T09:00"})
	s.Require().NoError(err)

	reminderNoteInput, err := s.WD.FindElement(selenium.ByCSSSelector, "[e2e-test-id='task-reminder-note-0']")
	s.Require().NoError(err)
	err = reminderNoteInput.SendKeys("Automated reminder note")
	s.Require().NoError(err)

	s.TakeScreenshot("task_crud_02_form_filled")

	// 7. Click Submit
	s.ClickElement("[e2e-test-id='task-submit-btn']", "Task Submit button")

	// 8. Assert Task is in default Kanban view under "To Do"
	s.CurrentTest.LogStep("Verify Kanban Task", "INFO", fmt.Sprintf("Waiting for task card '%s' in To Do lane", taskTitle))
	taskCardFound := false
	end = time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		elem, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//div[contains(@e2e-test-id, 'kanban-lane-to-do')]//h4[contains(text(), '%s')]", taskTitle))
		if err == nil {
			if displayed, _ := elem.IsDisplayed(); displayed {
				taskCardFound = true
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().True(taskCardFound, "Task card was not found in the To Do swimlane")
	s.TakeScreenshot("task_crud_03_kanban_todo")

	// 9. Transition Task: Click move to "In Progress"
	s.CurrentTest.LogStep("Transition Task", "INFO", "Moving task to In Progress lane")

	// Find task card and parse task ID for cleanup tracking
	taskCard, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//h4[contains(text(), '%s')]/ancestor::div[contains(@e2e-test-id, 'task-card-')]", taskTitle))
	if err == nil {
		if attr, err := taskCard.GetAttribute("e2e-test-id"); err == nil {
			id := strings.TrimPrefix(attr, "task-card-")
			s.CreatedTaskIDs = append(s.CreatedTaskIDs, id)
		}
	}

	moveBtn, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//h4[contains(text(), '%s')]/ancestor::div[contains(@e2e-test-id, 'task-card-')]//button[contains(@e2e-test-id, 'task-move-') and contains(@e2e-test-id, '-in-progress')]", taskTitle))
	s.Require().NoError(err)
	err = moveBtn.Click()
	s.Require().NoError(err)

	// Wait for task card to appear in "In Progress" swimlane
	taskTransitioned := false
	end = time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		elem, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//div[contains(@e2e-test-id, 'kanban-lane-in-progress')]//h4[contains(text(), '%s')]", taskTitle))
		if err == nil {
			if displayed, _ := elem.IsDisplayed(); displayed {
				taskTransitioned = true
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().True(taskTransitioned, "Task card did not transition to the In Progress swimlane")
	s.TakeScreenshot("task_crud_04_kanban_in_progress")

	// 10. Toggle to Table view and verify
	s.ClickElement("[e2e-test-id='tasks-view-table']", "Switch to Table View button")

	s.CurrentTest.LogStep("Verify Table Row", "INFO", fmt.Sprintf("Waiting for task '%s' in table", taskTitle))
	rowFound := false
	end = time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		elem, err := s.WD.FindElement(selenium.ByXPATH, fmt.Sprintf("//*[contains(text(), '%s')]", taskTitle))
		if err == nil {
			if displayed, _ := elem.IsDisplayed(); displayed {
				rowFound = true
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.Require().True(rowFound, "Task row was not found in the table after view toggle")
	s.TakeScreenshot("task_crud_05_in_table")

	s.CurrentTest.LogStep("Task Workflow Complete", "PASSED", "Successfully completed and verified the Task creation, transition, and table view workflow")
}
