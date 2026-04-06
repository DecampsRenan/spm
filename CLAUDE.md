# spm — Agent instructions

## Changelog updates (REQUIRED)

After implementing any feature, change, fix, deprecation, removal, or
security-related change in this repo, you **must** update `CHANGELOG.md`:

1. Add an entry under the `## [Unreleased]` section.
2. Use the correct category heading: `Added`, `Changed`, `Deprecated`,
   `Removed`, `Fixed`, or `Security` (Keep a Changelog format).
3. Write entries from the user's perspective — describe the observable
   behavior change, not the implementation detail.
4. Include the changelog edit in the same commit as the code change.

This rule applies to **every** code change, including refactors that have
user-visible effects, new tests for new behavior, and bug fixes. Skip the
changelog only for purely internal cleanups with no user-facing impact
(e.g. renaming an unexported variable, fixing a typo in a comment).
