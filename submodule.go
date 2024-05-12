package main

import (
	"fmt"
	"strings"
)


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

func (subm Submodule) FullPath() string {
	fullpath := fmt.Sprintf("%v\\%v", subm.parent, subm.path)
	return fullpath 
}

func GetBranches(subm Submodule) []string {
	branches := []string{}

	fullpath := fmt.Sprintf("%v\\%v", subm.parent, subm.path)
	out, _ := RunExecCommand("git", []string{"-C", fullpath, "branch", "--remote", "-l"})
	values := strings.Split(string(out), "\n")
	branches = append(branches, values[:len(values) - 1]...)

	out, _ = RunExecCommand("git", []string{"-C", fullpath, "branch", "-l"})
	values = strings.Split(string(out), "\n")
	branches = append(branches, values[:len(values) - 1]...)

	branches = append(branches, "NEW BRANCH")

	return branches
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

func PullSubmodule(path string, parent string) {
	_, err := RunExecCommand("git", []string{"-C", parent, "submodule", "update", "--init", "--remote", path})

	if len(err) > 0 {
		print := fmt.Sprintf("Syncronized submodule [green::b]'%v'[-:-:-] changes", path)
		WriteLog(print)
	}
}

func UpdatePath(subm Submodule) {
	fixed := strings.TrimSpace(subm.newPath)

	_, err := RunExecCommand("git", []string{"-C", subm.parent, "mv", subm.path, fixed})

	if len(err) > 0 {
		print := fmt.Sprintf("Submodule [green::b]'%v'[-:-:-] renamed / relocated to [green::b]'%v'[-:-:-]", subm.path, fixed)
		WriteLog(print)
	}
}

func UpdateBranch(subm Submodule) {
	updateToBranch := strings.TrimSpace(subm.newBranch)
	updateToBranch = strings.Split(updateToBranch, " -> ")[0]

	if updateToBranch == "NEW BRANCH" {
		updateToBranch = subm.newCustomBranch

		fullpath := fmt.Sprintf("%v\\%v", subm.parent, subm.path)
		_, err := RunExecCommand("git", []string{"-C", fullpath, "switch", "-c", updateToBranch})
		if len(err) > 0 {
			print := fmt.Sprintf("Created and switched to branch [green::b]'%v'[-:-:-]", updateToBranch)
			WriteLog(print)
		} 

		if subm.createRemote {
			_, err := RunExecCommand("git", []string{"push", "-u", "origin", updateToBranch})
			if len(err) > 0 {
				print := fmt.Sprintf("Created remote [green::b]'origin/%v'[-:-:-] branch", updateToBranch)
				WriteLog(print)
			}
		}
	}

	_, err := RunExecCommand("git", []string{"-C", subm.parent, "submodule", "set-branch", "--branch", updateToBranch, subm.path})
	if len(err) > 0 {
		print := fmt.Sprintf("Submodule [green::b]'%v'[-:-:-] updated to branch [green::b]'%v'[-:-:-]", subm.path, updateToBranch)
		WriteLog(print)
		PullSubmodule(subm.path, subm.parent)
	} 
}
