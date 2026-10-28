# Releasing

Releases are built by [`.github/workflows/release.yml`](../.github/workflows/release.yml) whenever a semver tag is pushed.

## Cutting a release

1. Make sure `main` is green and the milestone's issues are closed.
2. Run the benchmark suite against a large repo (see [PLAN.md](PLAN.md#performance-targets)).
3. Tag and push:

   ```sh
   git tag -a v0.1.0 -m "v0.1.0"
   git push origin v0.1.0
   ```

   A tag with a pre-release suffix, such as `v0.2.0-rc.1`, is published as a GitHub pre-release.

The workflow then does the following:

1. Stamps the tag's version into the binary (`internal/version.Version`) and into `wails.json`.
2. Builds the Windows NSIS installer. It is unsigned, so SmartScreen will warn.
3. Builds a universal macOS `.dmg`. When the Apple secrets are configured, it also signs, notarizes and staples it.
4. Creates the GitHub Release, with notes generated from merged PRs.

## Dry runs

PRs that change `release.yml` run the build and packaging steps without signing or publishing. The resulting artifacts are attached to the workflow run.

## macOS signing secrets

Set these in **Settings → Secrets and variables → Actions**. If `APPLE_CERTIFICATE_P12` or `APPLE_ID` is missing, the macOS build is published unsigned.

| Secret | Value |
|---|---|
| `APPLE_CERTIFICATE_P12` | Your "Developer ID Application" certificate and private key, exported as `.p12` and base64-encoded (`base64 -i cert.p12 \| pbcopy`) |
| `APPLE_CERTIFICATE_PASSWORD` | The password used when exporting the `.p12` |
| `APPLE_SIGNING_IDENTITY` | For example `Developer ID Application: Your Name (TEAMID)` |
| `APPLE_ID` | The Apple ID email used for notarization |
| `APPLE_APP_PASSWORD` | An [app-specific password](https://support.apple.com/en-us/102654) for that Apple ID |
| `APPLE_TEAM_ID` | Your 10-character Team ID |
