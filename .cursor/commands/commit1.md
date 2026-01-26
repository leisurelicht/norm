# Commit All Current Changes (Structured & Intent-Driven)

This command commits all current changes in the repository using **well-structured, Conventional Commits–compliant messages**, with **automatic splitting into multiple commits when multiple independent changes are detected**.

---

## 1. Repository Inspection (Mandatory)

When this command is run, **always begin with a full inspection of the repository state**:

1. Check the working tree status:
```bash
git status
```
- Identify modified, added, deleted, and untracked files.

2. Inspect actual code changes:
```bash
git diff
```
- If any files are already staged, also run:
```bash
git diff --cached
```

---

## 2. Change Analysis & Classification

Based on the diffs, infer:

- **What** was changed:
  - feature, bug fix, refactor, performance, tests, documentation, tooling, CI, etc.
- **Why** the change was made:
  - bug prevention, correctness, performance, API clarity, maintainability, test coverage, etc.
- **Where** the change occurred:
  - which files, directories, and logical modules are involved.

Group files by **logical concern**, not by commit convenience.

---

## 3. Mandatory Multi-Commit Splitting Rule

If the working tree contains **more than one independent logical change**, you **MUST NOT** combine them into a single commit.

### 3.1 Detection Criteria

Treat changes as independent if **any** of the following apply:

- Different **commit types** (e.g. `feat` + `fix`, `refactor` + `docs`)
- Different **scopes or modules** (e.g. `auth/` vs `billing/`)
- Different **intent**:
  - user-facing behavior vs internal cleanup
  - functional change vs formatting / docs / tooling
- Changes that could be **reverted independently** without breaking each other

If yes → **split into multiple commits**.

---

## 4. Commit Splitting Procedure

For **each logical group of changes**, perform the following steps independently:

1. **Isolate the changes**
```bash
git add <paths>
```
or
```bash
git add -p
```

2. **Verify staging**
```bash
git diff --cached
```
Ensure that only **one coherent concern** is staged.

3. **Generate a commit message** using the structure below.

4. **Commit** using a full multi-line message.

5. Repeat until all changes are committed.

---

## 5. Commit Message Format (Required)

Every commit **must** follow **Conventional Commits**.

### 5.1 Title Line

- Single line, **≤ 72 characters**
- Focus on **why**, not just what
- Format:
```
<type>(<scope>): <short reason-focused summary>
```
- Types:
```
feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
```
- Use **imperative mood**
- Breaking change: add `!` after type

---

### 5.2 Body (Wrapped at ~72 chars)

```
Why:
- Explain the underlying reason or problem.

What:
- Summarize the main technical changes.

Files:
- List key files or directories touched.
```

---

## 6. Commit Creation

1. Stage relevant files only:
```bash
git add ...
```

2. Verify staged diff:
```bash
git diff --cached
```

3. Commit using an exact multi-line message:
```bash
git commit -m "$(cat <<'EOF'
<type>(<scope>): <summary>

Why:
- ...

What:
- ...

Files:
- ...
EOF
)"
```

---

## 7. Commit Ordering Rules

1. `chore` / `build` / `ci`
2. `refactor`
3. `feat`
4. `fix`
5. `test`
6. `docs`

---

## 8. Final Verification & Output

1. Ensure a clean working tree:
```bash
git status
```

2. Print a confirmation for **each commit**:
- Commit hash
- Title line
- Files changed
- Insertions / deletions

---

## 9. Safety Constraints

- Do **not** create empty commits
- Do **not** modify git configuration
- Do **not** push to any remote
- Each commit must be logically coherent and independently revertible
