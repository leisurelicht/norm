# AI Git Automator: Atomic Commits (Optimized)

**Goal:** Auto-split independent changes into **well-structured Conventional Commits** with mandatory emojis.

## 1. Analysis & Grouping (Mandatory)

* **Ops:** Run `git status` and `git diff`.
* **Logic:** Cluster files by **logical concern** (type, scope, or intent).
* **Constraint:** **MUST SPLIT** commits if changes have different types (e.g., `feat` vs `fix`), different scopes, or can be reverted independently. **Do NOT proceed** without full understanding.

## 2. Execution Loop

*Repeat until worktree clean:*

1. **Stage:** `git add <paths>` (Select **ONE** coherent concern only).
2. **Commit:** Execute the following command immediately (Verify atomicity internally):
```bash
git commit -m "<emoji> <type>(<scope>): <summary>

Why:
- [Briefly explain the underlying reason/problem]
What:
- [Summary of technical changes]
Files:
- [Key files/dirs, omit trivial]"

```
## 3. Final Output (Once Finished)

After all commits are done:

Print a summary table or list: [Hash] | [Emoji] [Title] | [Files Changed].

## 4. Standards & Rules

* **Title:** ≤ 72 chars, imperative mood, focus on **WHY**.
* **Emoji Map:** 🎸feat, 🐛fix, ✏️docs, 💄style, 💡refactor, ⚡️perf, 💍test, 🎡ci, 🤖chore.
* **Format:** `<emoji> <type>(<scope>): <summary>` (e.g., `🎸 feat(auth): add google oauth`).
* **Breaking Change:** Add `!` after type (e.g., `feat!(api): ...`).
* **Order:** `chore/ci` > `refactor` > `feat` > `fix` > `test` > `docs`.
* **Safety:** No empty commits; **No config changes**; **No `git push`**.