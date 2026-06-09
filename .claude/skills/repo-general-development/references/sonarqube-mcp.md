# SonarQube MCP Server

When a SonarQube MCP server is available, follow these guidelines during a task.

## Basic workflow

- **At the start of a task:** if a `toggle_automatic_analysis` tool exists, **disable** automatic
  analysis.
- **At the very end of a task,** after you've finished generating or modifying code files: call
  `analyze_file_list` (if it exists) to analyze the files you created or changed.
- **Then re-enable** automatic analysis with `toggle_automatic_analysis` (if it exists).

## Project keys

- When the user mentions a project key, call `search_my_sonarqube_projects` first to find the
  exact key. Don't guess keys.

## Language detection

- Detect the language from the code's syntax. If unclear, ask or make an educated guess.

## Branch / PR context

- Many operations support branch-specific analysis — if the user is on a feature branch, pass the
  branch parameter.

## Issues and violations

- After fixing issues, **don't** verify them via `search_sonar_issues_in_projects` — the server
  won't reflect the updates yet.

## Troubleshooting

- **Auth:** SonarQube requires **USER** tokens (not project tokens). A
  `SonarQube answered with Not authorized` error usually means the wrong token type.
- **Project not found:** use `search_my_sonarqube_projects`; verify key spelling/format.
- **Analysis quality:** specify the language correctly, provide full file content (snippet
  analysis is weaker), and remember snippet analysis doesn't replace a full project scan.
