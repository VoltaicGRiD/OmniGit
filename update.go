package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func RequestUpdateAll() {
	branches := []string{}
	subms := []Submodule{}

	for _, node := range selected {
		if p, ok := node.GetReference().(Submodule); ok {
			newBranches := GetBranches(p)
			for _, branch := range newBranches {
				branches = append(branches, fmt.Sprintf("%v : %v", p.path, branch))
			}
			subms = append(subms, p)
		}
	}

	for index := range branches {
		branches[index] = fmt.Sprintf(" %v ", strings.TrimSpace(branches[index]))
	}

	var master Submodule
	var action int

	updateForm.Clear(true).
		AddTextView("Updating", "ALL SELECTED SUBMODULES", 0, 1, false, false).
		AddDropDown("Checkout branch ", branches, 0, func(option string, optionIndex int) {
			if strings.Contains(option, "NEW BRANCH") {
				if updateForm.GetFormItemIndex("New branch") == -1 {
					updateForm.AddInputField("New branch", "", 0, nil, func(text string) {
						master.newCustomBranch = text
					})
				}
			}

			if strings.Contains(option, " * ") {
				master.newBranch = ""
			} else {
				master.newBranch = option
			}
		}).
		AddDropDown("Action", []string{" Update submodules ", " Fetch and pull ", " Close dialog "}, 0, func(option string, optionIndex int) {
			// 0 - update submodules
			// 1 - fetch and pull
			// 2 - close dialog
		        action = optionIndex
	        }).
		AddButton("Submit action", func() {
			if action == 0 {
				for _, subm := range subms {
					subm.newCustomBranch = master.newCustomBranch
					subm.newBranch = master.newBranch
					UpdateSubmodule(subm)
				}

				action := updateForm.GetFormItemByLabel("Action")
				if p, ok := action.(*tview.DropDown); ok {
					p.SetCurrentOption(1)
				}
			}

			if action == 1 {
				for _, subm := range subms {
					path := subm.path
					parent := subm.parent
					PullSubmodule(path, parent)
				
}

				action := updateForm.GetFormItemByLabel("Action")
				if p, ok := action.(*tview.DropDown); ok {
					p.SetCurrentOption(2)
				}
			}

			if action == 2 {
				pages.HidePage("update")
				app.SetFocus(menuTree)
			}
		}).
		SetBackgroundColor(tcell.ColorDarkGray).
		SetBorder(true)

	updateForm.SetInputCapture(func(key *tcell.EventKey) *tcell.EventKey {
		if key.Key() == tcell.KeyEscape || key.Key() == tcell.KeyEsc || key.Key() == tcell.KeyESC {
			pages.HidePage("update")
			app.SetFocus(menuTree)

			return nil
		}

		return key
	})
}

func RequestUpdate(subm Submodule) {
	branches := GetBranches(subm)

	for index := range branches {
		branches[index] = fmt.Sprintf(" %v ", strings.TrimSpace(branches[index]))
	}

	updateForm.Clear(true).
		AddTextView("Repository", subm.repo, 0, 1, false, false).
		AddTextView("Submodule", subm.path, 0, 1, false, false).
		AddDropDown("Checkout branch ", branches, 0, func(option string, optionIndex int) {
			if strings.Contains(option, "NEW BRANCH") {
				if updateForm.GetFormItemIndex("New branch") == -1 {
					updateForm.AddInputField("New branch", "", 0, nil, func(text string) {
						subm.newCustomBranch = text
				})
					updateForm.AddCheckbox("Create remote branch", true, func(checked bool) {
						subm.createRemote = checked
					})
				}
			}

			if strings.Contains(option, " * ") {
				subm.newBranch = ""
			} else {
				subm.newBranch = option
			}
		}).
		AddInputField("Change name", subm.path, 0, nil, func(text string) {
			subm.newPath = fmt.Sprintf("%v\\%v", subm.parent, text)
		}).
		AddButton("Submit", func() {
			UpdateSubmodule(subm)
			pages.HidePage("update")
		}).
		SetBackgroundColor(tcell.ColorDarkGray).
		SetBorder(true)

	updateForm.SetInputCapture(func(key *tcell.EventKey) *tcell.EventKey {
		if key.Key() == tcell.KeyEscape || key.Key() == tcell.KeyEsc || key.Key() == tcell.KeyESC {
			pages.HidePage("update")
			app.SetFocus(menuTree)

			return nil
		}

		return key
	})
}
