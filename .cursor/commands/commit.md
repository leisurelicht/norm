# AI Git Automator: Atomic Commits

**Goal:** Auto-split independent logic into **Conventional Commits** with mandatory emojis.

## 1. Inspect & Group (Mandatory)

* **Ops:** `git status`, `git diff`, `git diff --cached`.
* **Logic:** Cluster files by **concern** (type/scope/intent).
* **Split Criteria:** Separate by **Type** (`feat`/`fix`), **Scope**, or **Independent Revertibility**.

## 2. Execution Loop

*Repeat until worktree clean:*

1. **Stage:** `git add <paths>` or `git add -p` (1 concern only).
2. **Verify:** `git diff --cached` (Confirm atomicity).
3. **Commit:**

```bash
git commit -m "<emoji> <type>(<scope>): <summary>

Why:
- [Reason/Problem]
What:
- [Technical changes]
Files:
- [Key files, omit trivial]"

```

## 3. Standards & Constraints

* **Title:** ≤72 chars, imperative. **Format:** `<emoji> <type>(<scope>): <summary>`.
* **Emoji Map:** 🎸feat, 🐛fix, ✏️docs, 💄style, 💡refactor, ⚡️perf, 💍test, 🎡ci, 🤖chore.
* **Breaking:** Add `!` after type (e.g., `feat!:`).
* **Order:** `chore/ci` > `refactor` > `feat` > `fix` > `test` > `docs`.
* **Safety:** No empty commits; No config/push; Print Hash + Stats after each