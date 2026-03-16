# Project: spm (Smart Package Manager)

Go CLI that auto-detects npm/yarn/pnpm and proxies commands.

## Development

```bash
just setup               # install git hooks and dev tools
just test                # run tests
just fmt                 # format all Go files
just build               # build binary
```

## Conventions

- **README (OBLIGATOIRE)**: When adding or changing flags, commands, or user-facing behavior, you **must** update `README.md` (features list + usage examples). The CI reviewer will flag this.
- **Tests**: Add or update tests for any new functionality in the corresponding `_test.go` files.
- **Dry-run (OBLIGATOIRE)**: Toute nouvelle fonctionnalité qui exécute des commandes ou produit des effets de bord doit respecter le flag `--dry-run`. Vérifier que `runner.Run` (ou tout nouveau chemin d'exécution) court-circuite correctement quand `dryRun` est `true`. Les tests doivent couvrir le cas dry-run.
- **Formatting**: A pre-commit hook runs `goimports` on staged Go files. Run `just setup` after cloning.
- **CI**: GitHub Actions runs `go test ./... -v -race`, format checks, and cross-platform builds on every PR.

## Changelog (OBLIGATOIRE)

**Toute modification de code DOIT être accompagnée d'une mise à jour du CHANGELOG.**

Avant de soumettre ou pousser des changements :
1. Ouvrir `CHANGELOG.md`
2. Ajouter une entrée sous `## [Unreleased]` dans la catégorie appropriée :
   - **Added** : nouvelles fonctionnalités
   - **Changed** : modifications de fonctionnalités existantes
   - **Deprecated** : fonctionnalités bientôt supprimées
   - **Removed** : fonctionnalités supprimées
   - **Fixed** : corrections de bugs
   - **Security** : corrections de vulnérabilités
3. Le CI **bloquera le merge** si aucune entrée n'est présente dans `[Unreleased]`
4. Ne jamais modifier les sections versionnées (ex: `[0.1.5]`), uniquement `[Unreleased]`

## Release

Les releases sont déclenchées **manuellement** via GitHub Actions :
1. Aller dans **Actions > Release > Run workflow**
2. Choisir le type de bump : `patch`, `minor`, ou `major`
3. Le workflow calcule la version, met à jour le changelog, tag, et publie via GoReleaser
