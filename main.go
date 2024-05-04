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

	// "github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	updateForm *tview.Form
	app *tview.Application
	pages *tview.Pages
	log *tview.TextView
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

						if _, err := os.Stat(submodulePath); err == nil {
							newNode := tview.NewTreeNode(dir.Name())
							menuTree.GetRoot().AddChild(newNode)
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
						} else {
							newNode := tview.NewTreeNode(dir.Name())
							menuTree.GetRoot().AddChild(newNode)
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

func WriteLog(text string) {
	w := log.BatchWriter()
	defer w.Close()
	fmt.Fprintln(w, text)
}

func WriteErr(text string) {
	err := fmt.Sprintf("[red]%v[white]", text)
	w := log.BatchWriter()
	defer w.Close()
	fmt.Fprintln(w, err)
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

	menu := newGrid().SetRows(0, 4).SetColumns(0)
	menuTree := tview.NewTreeView()
	menuTreeRoot := tview.NewTreeNode("Current Dir")
	menuTree.SetRoot(menuTreeRoot).SetCurrentNode(menuTreeRoot)
	menu.AddItem(menuTree, 0, 0, 1, 1, 0, 0, false)
	menuHelp := newTextView(fmt.Sprintf("J / K : Up / Down\nu : Update Submodule(s)\nq : Quit OmniGit\na : Update all highlighted"))
	menu.AddItem(menuHelp, 1, 0, 1, 1, 0, 0, false)
	main := newTextView("Main content")

	header := newTextView("")
	dirInput := newInputField("Parent repo directory: ").SetText(cwd)
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
	updateModal := NewModal(updateForm, 60, 20)

	pages = tview.NewPages().AddPage("main", grid, true, true).AddPage("update", updateModal, true, false)

	app = tview.NewApplication()
	nodes := []*tview.TreeNode{}
	currentDir := dirInput.GetText()

	dirInput.SetDoneFunc(func(key tcell.Key) {
		app.SetFocus(menuTree)
	})

	menuTree.SetInputCapture(func(key *tcell.EventKey) *tcell.EventKey {
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
