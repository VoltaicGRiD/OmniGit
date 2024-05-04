# OmniGit
![License](https://img.shields.io/badge/license-MIT-green.svg)
![0.1](https://img.shields.io/badge/version-0.1%20Beta-green)

![GitHub issues](https://img.shields.io/github/issues/VoltaicGRiD/omnigit)
![GitHub pull requests](https://img.shields.io/github/issues-pr/VoltaicGRiD/omnigit)

![I Stand With Ukraine](https://img.shields.io/badge/-I_Stand_With_Ukraine-gray.svg?logo=data:image/png%2bxml;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAIAAAD8GO2jAAAACXBIWXMAAAGKAAABigEzlzBYAAAANUlEQVRIiWNkjNjBQEvARFPTRy0YtWDUglELRooFjP+u0taCoR9EoxaMWjBqwagFoxZQAwAAM/cDI/lLSCIAAAAASUVORK5CYII=)


## Overview
OmniGit is a comprehensive Git submodule management tool designed to streamline the handling of multiple submodules within multiple Git repositories. Built to be utilized within the terminal for fast operation, quick modifications, and ease-of-access.

## Features
- Display current time and submodule status updates in real time.
- Manage submodule paths, branches, and repository details interactively.
- Execute Git commands within the context of each submodule's specific directory path.
- Graphical interface built with `tview` and `tcell` for a more intuitive user experience.

## To-Do
- [ ] Allow for updating all submodules with a single input window
  - [ ] Branches must exist / be created for all submodules selected
- [X] ~~Implement [Lazygit](https://github.com/jesseduffield/lazygit) style push / pull~~
  - [X] ~~Implement `lazygit` as a built-in overlay app? (could be useful)~~
- [ ] Multi-platform support
  - [ ] Linux
    - [ ] Arch
    - [ ] Debian
    - [ ] Ubuntu
  - [ ] MacOS
- [ ] Apt-distribution
  - [ ] Choco
  - [ ] Winget
  - [ ] Apt-get

## Prerequisites
Before you begin, ensure you have the following installed:
- Go (version 1.15 or higher)
- Git
- [Lazygit](https://github.com/jesseduffield/lazygit)

## Commands 
While focused on the tree, the following keys are useful to know:
| Key | Output |
| ----- | ------------------ |
| J / K | Navigate up & down |
| F2    | Re-enter primary folder |
| Enter | Select Submodule |
| q     | Quit Omnigit |
| a     | Update all highlighted submodules |
| u     | Update currently selected submodule |
| l     | Open lazygit to current repo |

## Installation
Clone the repository to your local machine:
```bash
git clone https://github.com/your-repository/omnigit.git
cd omnigit
```

Build the application with:
```bash
go build -o omnigit
```

## Usage
Run the application using:

```bash
./omnigit
```

Navigate through the interface using the keyboard. Use the 'a' key to apply updates to all submodules, 'u' to update the highlighted repo's submodules, or an individule submodule, 'RET' to highlight a submodule for the 'a' command, and 'q' to quit the application.

## Contributing
Contributions to OmniGit are welcome! Please feel free to fork the repository, make changes, and submit a pull request.

## License
Distributed under the MIT License. See LICENSE file for more information.
