# OmniGit
![License](https://img.shields.io/badge/license-MIT-green.svg)
![0.1 Beta](https://img.shields.io/github/v/release/VoltaicGRiD/omnigit)
![GitHub issues](https://img.shields.io/github/issues/VoltaicGRiD/omnigit)
![GitHub pull requests](https://img.shields.io/github/issues-pr/VoltaicGRiD/omnigit)
![I Stand With Ukraine](https://img.shields.io/badge/I%20stand%20with%20Ukraine-007FFF?style=flat&logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAABhElEQVR42mNkoBAw4pF/4///nEd02s7OxsgauHTpUjMCJBUcDP7//x8DkrF3797BYPfu3TB27NghTCYTNn369GFQUDD8//8fR3n27BlMjU3zrVu3BorFxUWdA3JpaTlIJgaPHz8eYAlHRkaGIMl///6Fq4MHD8a8vDxKTEzE4MGDB8PKlSvj/PnzBfz58wcGxsbGAr///j3IoE2bNkXx9ttvB8bjx4/DgAEDAn5/fx/29vYOCYODg7F///4Fi7Bp0yY4OzvDiMViGH///j0YPT09DHZ2doY5cuQIQsaPHx8Mjc3VxkUFFRoAUQymVwYkNOnT2OQl5cXQiQS0Z4xY4YBf3/+/GHg6emJMaOjo3F+fr5N8P79+wUCAwMxtmzZEm6//faFpEmTJghOnjzJ9O3bF4ZRo0YJzLly5cKF4OzZs0NkcnKywWJiYmJk/fr1Bfr8+fPhwIEDsXHjxgVr167FyJEjBQI/Pz8Ynp6eDoqKisISGRmJwcHBwfj48SPCzMwMcXFxMXj8+DHcuHEDpqenY+LEiREOHjwIHj48CH379oUDBw4ESUpKilGpUiUkJSUJgoKCAo6Pj0NERAT27NkTM2bMEIiLi8P58+dBbm5uGDRoEDZt2oR///4F5uPjo3j48CEyMjLCF198AYWFhYHg4GDMnj0bqqqqMD09XWCQmJgIQUFBAAAgAElEQVQYvXr1CqqqqsLmzZuRnJwcQkJCQpCbm4urq6sGDx4cqFKlSiQnJ8do0KAB9u7dCyEhIQGhUChITEyEi4uL+/fvY+PGjQgMDERFRQUiIyODU6dOYezYsQgLC0NSUhK6d++OsLAwBAUFYcWKFbF9+3ZERkaGwZEjR4ZOnTrF2LFjkZaWJhQKBY4ePQpFRUUYGRmJs2fP4u7du7F//37s2bNHnDhxAn5+fkRGRoZBQUEYuXNixIgRiIqKEps3b0ZkZCRu3LiwEBYFBQWJhoYGgoKC4OrVq7Fw4UKkpKQIuVzO0KFDsXTpUly7dg0TJ05EYWGhwCAxMTHCxcVFeHh4oKSkBMuWLYOjo6Phy5cvyMjIwIwZM5CQkIDg4GDMmzePmJgY/Pz8sHnzZly6dAnHjh3D3bt3cfnyZbS0tEBoaCgGDBiA+fPnIzw8HJ6enqJ6hw8fRnBwMJ48eYJFixbhwoUL0tLQIH/+/MGRo9mzZw9ydnbGv//+K4nM7du3o6KiIvqGBgYG2LdvHyZMmICjR49i+vTpwJkzZyIjIwM///wD27dvx9ixY+Ht7Y2MjAwYGBggsVgsgwYNQk5ODk6ePIkTJ06IL+/u1RUvU6dOVKkqVgCAwMdTDoAAAMA8GoVEzeFNeQAAAAASUVORK5CYII=)


## Overview
OmniGit is a comprehensive Git submodule management tool designed to streamline the handling of multiple submodules within a Git repository. This application provides a graphical user interface to facilitate easy tracking and updating of submodule branches and paths.

## Features
- Display current time and submodule status updates in real time.
- Manage submodule paths, branches, and repository details interactively.
- Execute Git commands within the context of each submodule's specific directory path.
- Graphical interface built with `tview` and `tcell` for a more intuitive user experience.

## Prerequisites
Before you begin, ensure you have the following installed:
- Go (version 1.15 or higher)
- Git

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

Navigate through the interface using the keyboard. Use the 'a' key to apply updates to all submodules, 'u' to update the highlighted repo's submodules, or an individule submodule, and 'q' to quit the application.

## Contributing
Contributions to OmniGit are welcome! Please feel free to fork the repository, make changes, and submit a pull request.

## License
Distributed under the MIT License. See LICENSE file for more information.
