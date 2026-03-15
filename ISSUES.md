# Planned GitHub Issues

## High value, low effort

### Issue 1: Add `spm remove <package>` command

**Labels:** `enhancement`, `good first issue`

**Description:**

Currently, there is no way to remove a package using `spm`. Each package manager uses a different command for this:

| spm (proposed) | npm | yarn | pnpm |
|---|---|---|---|
| `spm remove foo` | `npm uninstall foo` | `yarn remove foo` | `pnpm remove foo` |

**Expected behavior:**

```sh
spm remove react
# → translates to the correct uninstall command for the detected package manager
```

This is a core operation that should be part of the default command set.

---

### Issue 2: Add `spm run` with interactive script selection

**Labels:** `enhancement`, `ux`

**Description:**

When running `spm run` without any argument, spm should display an interactive menu listing all available scripts from `package.json`, allowing the user to pick one.

This avoids having to manually open `package.json` just to check what scripts are available.

**Expected behavior:**

```sh
$ spm run
? Select a script to run:
  ❯ dev
    build
    test
    lint
    format
```

The selected script should then be executed using the detected package manager.

---

### Issue 3: Add `spm update [package]` command

**Labels:** `enhancement`

**Description:**

Add support for updating dependencies. Each package manager has a slightly different command:

| spm (proposed) | npm | yarn | pnpm |
|---|---|---|---|
| `spm update` | `npm update` | `yarn upgrade` | `pnpm update` |
| `spm update foo` | `npm update foo` | `yarn upgrade foo` | `pnpm update foo` |

**Expected behavior:**

```sh
# Update all dependencies
spm update

# Update a specific package
spm update react
```

---

## Existing issues to implement

### Issue 4: Add `spm doctor` command

**Labels:** `enhancement`

**Ref:** #9

**Description:**

Add a `spm doctor` command that performs a health check on the current project's package management setup. This is especially useful during onboarding or when debugging environment issues.

**Checks to perform:**

- [ ] Detected package manager is installed on the system
- [ ] Package manager version is compatible / up to date
- [ ] Lock file is consistent with `package.json` (no drift)
- [ ] `node_modules` exists and is not corrupted
- [ ] Node.js version satisfies `engines` field (if present)

**Expected output:**

```sh
$ spm doctor
✔ Package manager: yarn (v4.1.0)
✔ Lock file: yarn.lock is consistent
✔ node_modules: present
✔ Node.js: v20.11.0 (satisfies >=18)

All checks passed!
```

---

### Issue 5: Add `spm why <package>` command

**Labels:** `enhancement`

**Ref:** #2

**Description:**

Add a `spm why <package>` command that explains why a given package is installed — i.e., which dependency pulled it in.

| spm (proposed) | npm | yarn | pnpm |
|---|---|---|---|
| `spm why foo` | `npm explain foo` | `yarn why foo` | `pnpm why foo` |

**Expected behavior:**

```sh
$ spm why tslib
# → runs the appropriate "why" command for the detected package manager
```

This is a high-value debugging tool when investigating dependency trees.

---

## Comfort features

### Issue 6: Add `spm clean` command

**Labels:** `enhancement`

**Description:**

Add a `spm clean` command that removes `node_modules` and optionally the lock file. This is a frequent operation that most developers do manually with `rm -rf node_modules`.

**Expected behavior:**

```sh
# Remove node_modules only
spm clean

# Remove node_modules AND the lock file
spm clean --lock
```

A confirmation prompt should be shown before deleting, especially when `--lock` is used, since regenerating a lock file can introduce version changes.

---

### Issue 7: Add `spm outdated` command

**Labels:** `enhancement`

**Description:**

Add a `spm outdated` command to list packages that have newer versions available.

| spm (proposed) | npm | yarn | pnpm |
|---|---|---|---|
| `spm outdated` | `npm outdated` | `yarn outdated` | `pnpm outdated` |

**Expected behavior:**

```sh
$ spm outdated
Package    Current  Wanted  Latest
react      18.2.0   18.2.0  19.0.0
vite       5.0.0    5.0.12  6.1.0
```

The output format should match whatever the underlying package manager returns.

---

## UX / Polish

### Issue 8: Show detected package manager at startup (verbose mode)

**Labels:** `enhancement`, `ux`

**Description:**

Add an optional verbose output that shows which package manager was detected and why. This helps users confirm that spm is using the right tool, and is useful for debugging detection issues.

**Expected behavior:**

```sh
$ spm install --verbose
→ using yarn (detected via yarn.lock)
yarn install v4.1.0
...
```

This could also be enabled via an environment variable like `SPM_VERBOSE=1`.

---

### Issue 9: Add shell completions (`spm completion bash/zsh/fish`)

**Labels:** `enhancement`, `ux`

**Description:**

Add a `spm completion` command that generates shell completion scripts for bash, zsh, and fish. This is a significant quality-of-life improvement.

**Completions should include:**

- All spm subcommands (`install`, `add`, `remove`, `run`, etc.)
- Scripts defined in `package.json` (for `spm run <TAB>`)
- Package names from `node_modules` (for `spm remove <TAB>`, `spm update <TAB>`)

**Expected behavior:**

```sh
# Generate completions for your shell
spm completion bash >> ~/.bashrc
spm completion zsh >> ~/.zshrc
spm completion fish > ~/.config/fish/completions/spm.fish
```

Cobra (the CLI framework used by spm) has [built-in support for shell completions](https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md), which should make this easier to implement.
