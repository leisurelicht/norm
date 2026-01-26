**Goal:** Commit all changes using **well-structured, Conventional Commits** with **auto-splitting into multiple commits** for independent logic.

## 1. Always First Inspection & Analysis (Mandatory)

* **Action:** Run `git status`, `git diff`, and `git diff --cached`.
* **Analyze:** Identify **What** (type), **Why** (reason), and **Where** (scope).
* **Group:** Cluster files by **logical concern**, not convenience.
* **Do NOT proceed** if you don't fully understand every change.

## 2. Splitting Rules (Atomic Commits)

**MUST SPLIT** if changes differ in:

* **Type:** e.g., `feat` vs `fix`, `refactor` vs `docs`.
* **Scope:** e.g., `auth/` vs `billing/`.
* **Intent:** Behavior change vs internal cleanup.
* **Revertibility:** If one can be reverted without breaking the other.

## 3. Execution Loop

*Repeat until worktree clean:*

1. **Stage:** `git add <paths>` or `git add -p` (one coherent concern).
2. **Verify:** `git diff --cached` (Confirm atomicity).
3. **Commit:** Multi-line commit using the format below.

## 4. Message Format (Strict)

### 4.1 Title Line (≤ 72 chars)

* **Format:** `<type>(<scope>): <imperative summary of WHY>`
* **Types:** 🎸 feat,🐛 fix, ✏️ docs, 💄 style, 💡 refactor, ⚡️ perf, 💍 test, 🎡 ci, 🤖 chore, don't lose emoji.
* **Breaking:** Add `!` after type (e.g., `feat!: ...`).

### 4.2 Body (Wrap ~72 chars)

```text
Why:
- Underlying reason or problem.
What:
- Summary of technical changes.
Files:
- List key files/dirs (omit trivial).

```

## 5. Implementation Command

```bash
git commit -m "<type>(<scope>): <summary>

Why:
- ...

What:
- ...

Files:
- ..."

```

## 6. Ordering & Safety

* **Order:** `chore/ci` > `refactor` > `feat` > `fix` > `test` > `docs`.
* **Safety:** No empty commits; No config changes; No `git push`.
* **Verify:** Run `git status` after each loop; Print Hash + Title + Stats.