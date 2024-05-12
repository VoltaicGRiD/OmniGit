package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	updateForm *tview.Form
	app *tview.Application
	pages *tview.Pages
	log *tview.TextView
	menuTree *tview.TreeView
	selected []*tview.TreeNode
)

type CurrentTime struct {
	time   string
	rwlock sync.RWMutex
}

type Submodule struct {
	path string
	parent string
	repo string
	branch string

	newPath string
	newBranch string
	newCustomBranch string
	createRemote bool
}

func (ct *CurrentTime) Set(value string) {
	ct.rwlock.Lock()
	defer ct.rwlock.Unlock()
	ct.time = value
}

func (ct *CurrentTime) Get() string {
	ct.rwlock.RLock()
	defer ct.rwlock.RUnlock()
	return ct.time
}

func GetBranches(subm Submodule) []string {
	branches := []string{}

	fullpath := fmt.Sprintf("%v\\%v", subm.parent, subm.path)
	cmd := exec.Command("git", []string{"-C", fullpath, "branch", "--remote", "-l"}...)
	stderr, err := cmd.StderrPipe()
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		WriteErr(err.Error())
	}

	if err := cmd.Start(); err != nil {
		slurp, _ := io.ReadAll(stderr)
		WriteErr(fmt.Sprintf("%s", slurp))
		WriteErr(err.Error())
	}

	slurp, _ := io.ReadAll(stdout)
	values := strings.Split(string(slurp), "\n")
	branches = append(branches, values[:len(values) - 1]...)

	cmd = exec.Command("git", []string{"-C", fullpath, "branch", "-l"}...)
	stderr, err = cmd.StderrPipe()
	stdout, err = cmd.StdoutPipe()

	if err != nil {
		WriteErr(err.Error())
	}

	if err := cmd.Start(); err != nil {
		slurp, _ := io.ReadAll(stderr)
		WriteErr(fmt.Sprintf("%s", slurp))
		WriteErr(err.Error())
	}

	slurp, _ = io.ReadAll(stdout)
	values = strings.Split(string(slurp), "\n")
	branches = append(branches, values[:len(values) - 1]...)

	branches = append(branches, "NEW BRANCH")

	return branches
}

func UpdateBranch(subm Submodule) {
	updateToBranch := strings.TrimSpace(subm.newBranch)
	updateToBranch = strings.Split(updateToBranch, " -> ")[0]

	if updateToBranch == "NEW BRANCH" {
		updateToBranch = subm.newCustomBranch

		fullPath := fmt.Sprintf("%v\\%v", subm.parent, subm.path)
		cmd := exec.Command("git", []string{"-C", fullPath, "switch", "-c", updateToBranch}...)

		stderr, err := cmd.StderrPipe()

		if err != nil {
			WriteErr(err.Error())
		}

		if err := cmd.Start(); err != nil {
			slurp, _ := io.ReadAll(stderr)
			WriteErr(fmt.Sprintf("%s", slurp))
			WriteErr(err.Error())
		} else {
			out := fmt.Sprintf("Created and switched to branch [green::b]'%v'[-:-:-]", updateToBranch)
			WriteLog(out)
		}

		if subm.createRemote {
			cmd := exec.Command("git", []string{"push", "-u", "origin", updateToBranch}...)

			stderr, err := cmd.StderrPipe()

			if err != nil {
				WriteErr(err.Error())
			}

			if err := cmd.Start(); err != nil {
				slurp, _ := io.ReadAll(stderr)
				WriteErr(fmt.Sprintf("%s", slurp))
				WriteErr(err.Error())
			} else {
				out := fmt.Sprintf("Created remote [green::b]'origin/%v'[-:-:-] branch", updateToBranch)
				WriteLog(out)
			}
		}
	}


	cmd := exec.Command("git", []string{"-C", subm.parent, "submodule", "set-branch", "--branch", updateToBranch, subm.path}...)
	stderr, err := cmd.StderrPipe()

	if err != nil {
		WriteErr(err.Error())
	}

	if err := cmd.Start(); err != nil {
		slurp, _ := io.ReadAll(stderr)
		WriteErr(fmt.Sprintf("%s", slurp))
		WriteErr(err.Error())
	} else {
		out := fmt.Sprintf("Submodule [green::b]'%v'[-:-:-] updated to branch [green::b]'%v'[-:-:-]", subm.path, updateToBranch)
		WriteLog(out)
		PullSubmodule(subm.path, subm.parent)
	}
}

func PullSubmodule(path string, parent string) {
	cmd := exec.Command("git", []string{"-C", parent, "submodule", "update", "--init", "--remote", path}...)
	stderr, err := cmd.StderrPipe()

	if err != nil {
		WriteErr(err.Error())
	}

	if err := cmd.Start(); err != nil {
		slurp, _ := io.ReadAll(stderr)
		WriteErr(fmt.Sprintf("%s", slurp))
		WriteErr(err.Error())
	} else {
		out := fmt.Sprintf("Syncronized submodule [green::b]'%v'[-:-:-] changes", path)
		WriteLog(out)
	}
}

func UpdatePath(subm Submodule) {
	fixed := strings.TrimSpace(subm.newPath)

	cmd := exec.Command("git", []string{"-C", subm.parent, "mv", subm.path, fixed}...)
	stderr, err := cmd.StderrPipe()

	if err != nil {
		WriteErr(err.Error())
	}

	if err := cmd.Start(); err != nil {
		slurp, _ := io.ReadAll(stderr)
		WriteErr(fmt.Sprintf("%s", slurp))
		WriteErr(err.Error())
	} else {
		out := fmt.Sprintf("Submodule [green::b]'%v'[-:-:-] renamed / relocated to [green::b]'%v'[-:-:-]", subm.path, fixed)
		WriteLog(out)
	}
}

func (subm Submodule) GetRelativePath() string {
	return ""
}

func UpdateSubmodule(subm Submodule) {
	if subm.newBranch != "" && subm.newBranch != subm.branch {
		UpdateBranch(subm)
		subm.branch = subm.newBranch
		subm.newBranch = ""
	}

	if subm.newPath != "" && subm.newPath != subm.path {
		UpdatePath(subm)
		subm.path = subm.newPath
		subm.newPath = ""
	}
}

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

func UpdateNodes(root string, menuTree *tview.TreeView) (nodes []*tview.TreeNode) {
	nodes = []*tview.TreeNode{}

	err := filepath.WalkDir(root, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else {

			if dir.IsDir() {
				// Check if the path of this file / directory is the same as our parent
				if path != root {
					trimmedRoot := strings.TrimRight(root, "\\")
					trimmedPath := strings.TrimRight(path, "\\")

					// Check how deep this path goes
					rootDepth := strings.Count(trimmedRoot, "\\")
					pathDepth := strings.Count(trimmedPath, "\\")

					if pathDepth - rootDepth == 1 {
						// Check if dir has submodules
						submodulePath := fmt.Sprintf("%v\\.gitmodules", trimmedPath)

						newNode := tview.NewTreeNode(dir.Name())
						newNode.SetReference(trimmedPath)
						menuTree.GetRoot().AddChild(newNode)

						if _, err := os.Stat(submodulePath); err == nil {
							file, err := os.Open(submodulePath)
							if err != nil {
							} else {
								defer file.Close()
								scanner := bufio.NewScanner(file)

								submodules := []Submodule{}

								var (
									mpath string
									branch string
								)

								mpath = ""
								branch = ""

								more := true

								for more {
									if strings.HasPrefix(strings.TrimSpace(scanner.Text()), "path = ") {
										if mpath != "" {
											split := strings.Split(path, "\\")
											repo := split[len(split) - 1]
											s := Submodule{mpath, path, repo, branch, "", "", "", true}
											submodules = append(submodules, s)
										}

										value := strings.Split(scanner.Text(), " = ")[1]
										mpath = value
									}

									if strings.HasPrefix(strings.TrimSpace(scanner.Text()), "branch = ") {
										branch = strings.Split(scanner.Text(), " = ")[1]
									}

									more = scanner.Scan()
								}

								if mpath != "" {
									split := strings.Split(path, "\\")
									repo := split[len(split) - 1]
									s := Submodule{mpath, path, repo, branch, "", "", "", true}
									submodules = append(submodules, s)
								}

								if err := scanner.Err(); err != nil {
									WriteErr(err.Error())
								}

								for _, element := range submodules {
									newNode.SetColor(tcell.ColorYellow)
									newNode.SetReference(path)
									newChild := tview.NewTreeNode(element.path).SetColor(tcell.ColorGreen)
									newChild.SetReference(element)
									newChild.SetSelectedFunc(func() {
										if newChild.GetColor() == tcell.ColorRed {
											newChild.SetColor(tcell.ColorGreen)
										} else {
											newChild.SetColor(tcell.ColorRed)
										}
									})
									selected = append(selected, newChild)
									newNode.AddChild(newChild)
								}
								newNode.SetReference(submodules)
							}
						}
					}
				}
			}

			return nil
		}
	})

	if err != nil {
		WriteErr(err.Error())
	}

	return nodes
}

// Writes a message to the logger
// This message is not formatted before being output
func WriteLog(text string) *tview.TextView {
	w := log.BatchWriter()
	defer w.Close()
	fmt.Fprintln(w, text)

	return log
}

// Writes an error to the logger
// Message is written-out with red text
func WriteErr(text string) *tview.TextView {
	err := fmt.Sprintf("[red]%v[-]", text)
	w := log.BatchWriter()
	defer w.Close()
	fmt.Fprintln(w, err)

	return log
}

// Clears the logger and returns it for writting
func ClearLog() *tview.TextView {
	log.Clear()
	return log
}

func NewModal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

func main() {
	c := time.Tick(1 * time.Second)
	currentTime := &CurrentTime{}
	selected = []*tview.TreeNode{}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	newTextView := func(text string) *tview.TextView {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}

	newInputField := func(label string) *tview.InputField {
		return tview.NewInputField().
			SetLabel(label).
			SetFieldWidth(100)
	}

	newGrid := func() *tview.Grid {
		return tview.NewGrid()
	}

	menu := newGrid().SetRows(0).SetColumns(0)
	menuTree = tview.NewTreeView()
	menuTreeRoot := tview.NewTreeNode("Current Dir")
	menuTree.SetRoot(menuTreeRoot).SetCurrentNode(menuTreeRoot)
	menu.AddItem(menuTree, 0, 0, 1, 1, 0, 0, false)
	controlTable := tview.NewTable()
	controlTable.SetCellSimple(0, 0, "Key")
	controlTable.SetCellSimple(0, 1, " Function")
	controlTable.SetCellSimple(2, 0, "  j  ")
	controlTable.SetCellSimple(2, 1, " Navigate down")
	controlTable.SetCellSimple(3, 0, "  k  ")
	controlTable.SetCellSimple(3, 1, " Navigate up")
	controlTable.SetCellSimple(4, 0, "  u  ")
	controlTable.SetCellSimple(4, 1, " Update submodules or all submodules under highlighted parent individually")
	controlTable.SetCellSimple(5, 0, "  a  ")
	controlTable.SetCellSimple(5, 1, " Update all green-colored submodules individually")
	controlTable.SetCellSimple(6, 0, "  l  ")
	controlTable.SetCellSimple(6, 1, " Open lazygit to highlighted repo")
	controlTable.SetCellSimple(7, 0, "  q  ")
	controlTable.SetCellSimple(7, 1, " Quit OmniGit")
	controlTable.SetCellSimple(9, 0, "  o  ")
	controlTable.SetCellSimple(9, 1, " Omni-update all green-colored submodules simulatneously")
	controlTable.SetCellSimple(10, 0, " RET ")
	controlTable.SetCellSimple(10, 1, " Mark / unmark submodule for omni-update")
	controlTable.SetSeparator(tview.Borders.Vertical)
	main := newGrid().SetRows(0).SetColumns(0)
	main.AddItem(controlTable, 0, 0, 1, 1, 0, 0, false)

	header := newTextView("")
	dirInput := newInputField("Parent repo directory: ").SetText(cwd)
	dirOutput := newTextView("")
	log = newTextView("").SetTextAlign(tview.AlignLeft).SetDynamicColors(true).SetRegions(true)

	grid := tview.NewGrid().
		SetRows(2, 1, 0, 10).
		SetColumns(50, 0).
		SetBorders(true).
		AddItem(header, 0, 0, 1, 3, 0, 0, false).
		AddItem(dirInput, 1, 0, 1, 3, 0, 0, true).
		AddItem(log, 3, 0, 1, 3, 0, 0, false)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	grid.AddItem(menu, 0, 0, 2, 0, 0, 0, false).
		AddItem(main, 1, 0, 1, 3, 0, 0, false)

	// Layout for screens wider than 100 cells.
	grid.AddItem(menu, 2, 0, 1, 1, 0, 100, false).
		AddItem(main, 2, 1, 1, 2, 0, 100, false)

	updateForm = tview.NewForm()
	updateForm.AddCheckbox("test", false, func(changed bool) {WriteLog(fmt.Sprintf("%t", changed))})
	updateModal := NewModal(updateForm, 80, 20)

	pages = tview.NewPages().AddPage("main", grid, true, true).AddPage("update", updateModal, true, false)

	app = tview.NewApplication()
	nodes := []*tview.TreeNode{}
	currentDir := dirInput.GetText()
	dirOutput.SetText(currentDir)

	dirInput.SetDoneFunc(func(key tcell.Key) {
		grid.RemoveItem(dirInput)
		grid.AddItem(dirOutput, 1, 0, 1, 3, 0, 0, false)
		currentDir := dirInput.GetText()
		dirOutput.SetText(currentDir)
		app.SetFocus(menuTree)
	})

	menuTree.SetInputCapture(func(key *tcell.EventKey) *tcell.EventKey {
		// Allows the user to re-enter the root directory to look for submodules
		if key.Key() == tcell.KeyF2 {
			grid.RemoveItem(dirOutput)
			grid.AddItem(dirInput, 1, 0, 1, 3, 0, 0, false)
			app.SetFocus(dirInput)
		}

		// Updates a submodule or multiple submodules, if the parent node is selected
		if key.Rune() == 'u' {
			c := menuTree.GetCurrentNode()
			var submodule Submodule
			if p, ok := c.GetReference().(Submodule); ok {
				submodule = p

				RequestUpdate(submodule)
				pages.ShowPage("update")
			} else if p, ok := c.GetReference().([]Submodule); ok {
				var submodules []Submodule = p
				for _, element := range submodules {
					RequestUpdate(element)
					pages.ShowPage("update")
				}
			}

			return nil
		}

		// Updates all nodes in the range with an updateForm for each and every one
		if key.Rune() == 'a' {
			r := menuTree.GetRoot()

			var submodules []Submodule

			r.Walk(func(node *tview.TreeNode, parent *tview.TreeNode) bool {
				if node.GetColor() == tcell.ColorGreen {
					submodules = append(submodules, node.GetReference().(Submodule))
				}

				return true
			})

			for _, element := range submodules {
				RequestUpdate(element)
				pages.ShowPage("update")
			}
		}

		// Opens LazyGit to the selected repository if one exists,
		// otherwise, prompts the user to create a git repo
		if key.Rune() == 'l' {
			c := menuTree.GetCurrentNode()
			var path string
			if p, ok := c.GetReference().(Submodule); ok {
				path = fmt.Sprintf("%v\\%v", p.parent, p.path)

			} else if p, ok := c.GetReference().([]Submodule); ok {
				path = p[0].parent
			} else if p, ok := c.GetReference().(string); ok {
				path = p
			}

			cmd := exec.Command("lazygit", []string{"-p", path}...)
			stderr, err := cmd.StderrPipe()

			exists := true

			app.Suspend(func() {
				if err != nil {
					WriteErr(err.Error())
					return
				}

				if err := cmd.Start(); err != nil {
					WriteErr(err.Error())
					return
				}

				slurp, _ := io.ReadAll(stderr)
				slurpout := fmt.Sprintf("%s", slurp)
				WriteErr(slurpout)

				if strings.Contains(slurpout, "is not a valid git repository") {
					exists = false
					return
				}

				if err := cmd.Wait(); err != nil {
					WriteErr(err.Error())
					return
				}

				return
			})

			if exists == false {
				create := false

				createRepoForm := tview.NewForm()
				createRepoModal := NewModal(createRepoForm, 40, 8)
				pages.AddPage("create", createRepoModal, true, true)

				createRepoForm.
					Clear(true).
					AddTextView("Not a repo", "create one?", 0, 1, true, true).
					AddButton("Yes", func() {
						create = true
						pages.RemovePage("create")
						app.SetFocus(menuTree)
					}).
					AddButton("No", func() {
						create = false
						pages.RemovePage("create")
						app.SetFocus(menuTree)
					}).
					SetBackgroundColor(tcell.ColorDarkGray).
					SetBorder(true)

				app.SetFocus(createRepoForm)

				if create == true {
					cmd := exec.Command("git", []string{"-C", path, "init"}...)

					if err := cmd.Start(); err != nil {
						WriteErr(err.Error())
					}

					if err := cmd.Wait(); err != nil {
						WriteErr(err.Error())
					}

					WriteLog(fmt.Sprintf("Git repository initialized for [green::b]%v[-:-:-]", path))
				}

			}

		}

		// Updates all nodes that are highlighted green, with a single update form
		// If a branch doesn't exist for any submodule in particular, it will be created
		if key.Rune() == 'o' {
			RequestUpdateAll()

			pages.ShowPage("update")
			app.SetFocus(updateForm)

			return nil
		}

		// Closes the application
		// TODO: Add a confirmation dialog
		if key.Rune() == 'q' {
			app.Stop()

			return nil
		}

		return key
	})

	go func() {
		for {
			go getTime(currentTime)

			headerText := fmt.Sprintf("OmniGit\n%v", currentTime.Get())

			select {
			case <-c:
				app.QueueUpdateDraw(func() {
					header.SetText(headerText)

					trimmedDir := strings.TrimSpace(dirInput.GetText())
					if trimmedDir != currentDir {
						menuTree.GetCurrentNode().ClearChildren()
						stat, err := os.Stat(trimmedDir)
						if err != nil {
							if !os.IsNotExist(err) {
								WriteErr("Error accessing directory: " + err.Error())
							}
						} else {
							if stat.IsDir() {
								nodes = []*tview.TreeNode{}
								nodes = UpdateNodes(trimmedDir, menuTree)
								fmt.Println(len(nodes))
							} else {
								WriteErr("Not a directory: " + trimmedDir)
							}
						}

						currentDir = trimmedDir
					}

					log.ScrollToEnd()
				})
			}
		}
	}()

	if err := app.SetRoot(pages, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}

func getTime(currentTime *CurrentTime) {
	now := time.Now()
	currentTime.Set(fmt.Sprintf("%d-%d-%d %d:%d:%d\n",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second()))
}
