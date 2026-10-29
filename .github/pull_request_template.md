## Summary

<!-- What changes and why. -->

Closes #

## Checks

- [ ] Branch is named `<issue#>-short-slug` and commits follow [Conventional Commits](https://www.conventionalcommits.org/)
- [ ] `go test ./...` and `golangci-lint run` pass
- [ ] `npm run check`, `npm run lint` and `npm test` pass (in `frontend`)
- [ ] Wails bindings in `frontend/wailsjs` are committed exactly as generated, if the Go API changed
- [ ] UI changes are keyboard-accessible and work in light and dark themes (see [ACCESSIBILITY.md](../ACCESSIBILITY.md))
