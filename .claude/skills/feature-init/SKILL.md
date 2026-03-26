---
name: feature-init
description: >
  Initialize the creation of a new feature for the spm CLI through iterative questioning.
  Asks structured questions in French to build a complete, coherent feature specification,
  then implements it. Triggers when: user wants to add a new feature, command, flag, or
  behavior to spm. Also triggers on: "nouvelle feature", "new feature", "ajouter une commande",
  "add a command", "feature-init".
---

# Feature Init

Iterative questioning process to specify a new spm feature before implementation.

**Language**: Questions and exchanges in French. All code, comments, commits, and docs in English.

**Approach**: Ask 2-4 questions at a time. Evaluate answers. Ask more if needed. Never dump all questions at once.

## Phase 1 — Discovery

Before asking anything, read these files for current project state:
- `cmd/root.go` — existing commands and flags
- `internal/resolver/resolver.go` — command mapping per PM
- `AGENTS.md` — project conventions

Then ask about (adapt to context, prioritize what's unclear):

### Functional
- What does the feature do? What problem does it solve?
- Concrete usage example (full command the user would type)

### CLI Interface
- New Cobra command, new flag on existing command, or internal-only change?
- Command/flag name, aliases
- Positional args vs flags
- Expected output (text, JSON, interactive?)

### Package Manager Compatibility
- Same behavior for npm/yarn/pnpm/bun or differences?
- Depends on PM-specific capabilities?

## Phase 2 — Deep Dive

Based on Phase 1 answers, probe these areas:

### Terminal UX
- Visual feedback during execution (spinner, progress, colors?)
- Non-interactive behavior (piped output, CI)
- Integration with Charmbracelet styling (lipgloss, huh, bubbletea)

### Error Handling
- What happens when the underlying PM fails?
- What happens when no PM is detected?
- Expected error messages (clear, actionable)

### Integration with Existing Features
- Interaction with `--dry-run` (REQUIRED if the feature has side effects)
- Interaction with `--vibes`, `--notify`
- Impact on resolver (`internal/resolver`)?
- Impact on detector (`internal/detector`)?

### Edge Cases
- Behavior without a lock file
- Behavior from a subdirectory
- Unknown args/flags → pass-through to PM?

## Phase 3 — Validation

Before moving to implementation, verify ALL these criteria:

### Completeness Checklist
- Behavior described for each supported PM (npm, yarn, pnpm, bun)
- Clear CLI interface (command/flag name, args, output)
- At least one concrete command example
- Error handling defined for main failure cases
- `--dry-run` support addressed (required, or justified why N/A)
- Impact on existing files/packages identified

### Coherence Check
- No contradictions between answers
- Described behavior is feasible with current architecture
- Scope is clear — no deferred structural decisions

**If any criterion is unmet** → ask targeted questions about the identified gaps. Explain clearly why the information is needed.

**If all criteria are met** → proceed to Phase 4.

## Phase 4 — Implementation

Summarize the validated plan in 5-10 lines, then implement following project conventions:

1. Go code in appropriate packages (`cmd/`, `internal/`)
2. Tests in corresponding `_test.go` files
3. `--dry-run` support if applicable
4. Update `README.md` (features list + usage examples)
5. Update `CHANGELOG.md` under `[Unreleased]` > Added
6. Run `just fmt` and `just test` before finalizing
