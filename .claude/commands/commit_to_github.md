Your task is to commit all of today's work to Git and push it to the remote origin. Follow these steps exactly.

## Step 1 — Gather state

Run these three commands in parallel:
- `git status` — see what is staged, modified, or untracked
- `git diff HEAD` — see the full diff of all changes
- `git log --oneline --format="%h %ad %s" --date=short -10` — review recent commit messages for style

## Step 2 — Identify what to stage

Stage files that represent real work done today. Use specific file paths — never `git add .` or `git add -A`.

Skip:
- `.env` or any file containing secrets or credentials
- Binary build artifacts (`bin/`, `dist/`, `*.exe`)
- `node_modules/`, vendor directories
- Generated files that should not be tracked
- `.claude/session_memory.db` (already gitignored)

If there is nothing meaningful to commit, tell the user and stop.

## Step 3 — Write a commit message

- First line: short imperative summary (≤ 72 chars) — what was built or fixed
- If multiple distinct things changed, add a blank line then a brief bullet list
- Match the tone and style of recent commits in this repo
- End with the attribution line:
  `Authored by Mapleton Design`

Use a HEREDOC to pass the message:
```bash
git commit -m "$(cat <<'EOF'
<message here>

Authored by Mapleton Design
EOF
)"
```

## Step 4 — Push to origin

```bash
git push origin $(git branch --show-current)
```

## Step 5 — Confirm

Tell the user:
- The commit hash and message
- Which branch was pushed
- A one-line summary of what was included
