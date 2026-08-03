---
title: imprun command-line client
description: Install and use the thin client for an existing Windforce Cell.
---

`imprun` is the supported command-line client for Imprun Cloud and connected Windforce Cells. Installing it does not install or start a server, worker, database, or queue.

The client exposes high-level app, release, and Run workflows together with the complete low-level Control Plane API. Interactive terminals receive readable labels and tables. Redirected output remains compact JSON unless `--pretty`, `--json`, `--jq`, or `--template` selects another stable automation format. These output flags may appear before or after the command, matching common `gh` usage. Hosted Device Authorization and direct Cell credentials are supported.

## Install

Tagged releases contain `imprun` archives for Windows, macOS, and Linux on amd64 and arm64. Windows releases also publish direct portable `.exe` assets for package managers. These are operator-host binaries and are unrelated to the architecture of a Kubernetes Cell image.

The standalone installers select the current operating system and architecture, resolve the latest stable release unless a version is pinned, verify the downloaded asset's SHA-256, and replace the installed executable only after its version smoke test succeeds. They never install a server, worker, service, or credential.

Windows PowerShell installs to `%LOCALAPPDATA%\Programs\Imprun\bin` and adds that directory to the user `PATH` when necessary:

```powershell
irm https://github.com/imprun/cli/releases/latest/download/install.ps1 | iex
```

Linux and macOS install to `${XDG_BIN_HOME:-$HOME/.local/bin}` without modifying shell profiles:

```shell
curl -fsSL https://github.com/imprun/cli/releases/latest/download/install.sh | sh
```

Both installers automatically verify the keyless Sigstore bundle when `cosign` is available. Without `cosign`, they still require a unique matching SHA-256 entry and print that signer verification was skipped. For a fail-closed signed installation, inspect the script first and require `cosign` verification:

```powershell
irm https://github.com/imprun/cli/releases/latest/download/install.ps1 -OutFile install.ps1
Get-Content .\install.ps1
.\install.ps1 -Version 0.3.1 -RequireSignature
```

```shell
curl -fsSLO https://github.com/imprun/cli/releases/latest/download/install.sh
less install.sh
sh install.sh --version 0.3.1 --require-signature
```

Use `-InstallDir` and `-NoModifyPath` on Windows, or `--install-dir` on Linux/macOS, for an explicit location. The equivalent environment variables are `IMPRUN_VERSION`, `IMPRUN_INSTALL_DIR`, `IMPRUN_NO_MODIFY_PATH=1`, and `IMPRUN_REQUIRE_SIGNATURE=1`.

After the stable package is accepted into the Windows Package Manager community repository, install or upgrade it with:

```powershell
winget install --id Imprun.CLI --exact
winget upgrade --id Imprun.CLI --exact
```

For a manual installation, download the matching release asset together with `checksums.txt` and `checksums.txt.sigstore.json`. Verify the signer and checksum before placing the executable on `PATH`:

```shell
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity "https://github.com/imprun/cli/.github/workflows/release.yml@refs/tags/<VERSION>" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  checksums.txt
sha256sum --check --ignore-missing checksums.txt
```

Then extract the selected macOS/Linux archive or stage the Windows `.exe`, run `--version`, and move it into an executable directory on `PATH`.

## Upgrade

Rerun the standalone installer to upgrade. It stages and verifies the new executable before replacing the installed file, while keeping configuration and credentials outside the installation directory:

```powershell
irm https://github.com/imprun/cli/releases/latest/download/install.ps1 | iex
imprun --version
imprun context show
```

```shell
curl -fsSL https://github.com/imprun/cli/releases/latest/download/install.sh | sh
imprun --version
imprun context show
```

The executable contains no context or credential state. Contexts remain in the operating-system user configuration directory under `imprun/config.json`, or at `IMPRUN_CONFIG` when explicitly selected. Credentials remain in Windows Credential Manager, macOS Keychain, or the Linux Secret Service. Do not copy a credential into an upgrade directory or back it up as plaintext.

After upgrading, `imprun context show` proves that the selected target survived. `imprun auth status` additionally probes the selected workspace when it is reachable.

The repository build produces only `imprun`:

```shell
make build
```

On Windows the output is `.tmp/bin/imprun.exe`.

## Migrate from `wf`

The retired `wf` executable is not installed as an alias. When the default `imprun/config.json` does not exist, `imprun` reads `wf/config.json`; the next context mutation writes the new Imprun configuration. When a matching credential exists only under the operating-system credential service named `wf`, the first authenticated read copies it into the `imprun` service without exposing it as plaintext.

An imported hosted credential keeps its original OAuth client identifier so refresh and revocation continue to work during the transition. New Device Authorization logins use the public client `imprun-cli`. After `imprun auth status` succeeds, remove the old `wf` executable from `PATH`; the old release asset itself remains immutable.

## Configure a context

A context contains non-secret connection metadata: API URL, workspace, optional audit actor, account label, and authentication type. Interactive credentials are stored in Windows Credential Manager, macOS Keychain, or the Linux Secret Service. `IMPRUN_TOKEN` supplies a bearer token to one process and takes precedence without being written to the configuration file or credential store.

For a hosted target, login discovers a secretless OAuth 2.0 Device Authorization client from the selected target, opens the system browser, and validates the resulting access against the selected workspace:

```powershell
imprun context set hosted `
  --api-url https://cell.example.test `
  --workspace team `
  --use

imprun auth login
imprun auth status
imprun app list --summary
```

Use `imprun auth login --no-browser` on a remote shell to print the verification URL instead of trying to open it. `--web` explicitly selects the same hosted flow. `--account <label>` keeps credentials for another hosted account under a distinct local label.

The target publishes only non-secret discovery metadata at `/.well-known/imprun-cli.json`:

```json
{
  "schema_version": 1,
  "authentication": {
    "type": "oauth2-device",
    "issuer": "https://identity.example.test",
    "client_id": "imprun-cli",
    "audience": "windforce-api",
    "scopes": ["openid", "profile", "email", "offline_access"]
  }
}
```

The issuer's OpenID configuration supplies the Device Authorization and token endpoints. The public client has no client secret. `imprun` stores the short-lived access token and refresh token only in the operating-system credential store and refreshes before expiry. Discovery and OAuth endpoints require HTTPS except for loopback development; cross-origin discovery redirects and OAuth POST redirects are rejected.

For a direct self-hosted Cell, read a named workspace credential from standard input:

```powershell
imprun context set local `
  --api-url http://127.0.0.1:18091 `
  --workspace default `
  --actor developer@example.test `
  --use

Get-Content .\workspace-token.txt | imprun auth login --with-token --account operator
imprun auth status
imprun app list --summary
```

Use `imprun context list`, `imprun context show`, and `imprun context use <name>` to inspect or select contexts. `imprun auth switch <account>` selects another credential already stored for the same host and verifies it before changing the context. `IMPRUN_CONFIG` selects an explicit configuration file. The migration behavior for the retired `wf` configuration applies only when this override is absent.

Context creation rejects malformed API URLs, non-loopback HTTP, embedded URL credentials, query strings, fragments, control characters, and non-portable token-environment names before writing configuration. An authenticated context cannot be retargeted to another origin until it is logged out, preventing an unreachable credential from being orphaned. Use `imprun context delete <name> --yes` to remove non-secret context metadata. An authenticated context must be logged out first, and another context must be selected before deleting the current context when alternatives exist.

Both login modes validate the credential against the selected workspace before storing it. If the system credential store is unavailable, login fails without writing a plaintext fallback. If hosted Device Authorization succeeds but the workspace probe or local commit fails, `imprun` revokes the unused refresh token before returning the error.

For a hosted account, `imprun auth logout` first revokes the CLI refresh token at the provider and removes local state only after revocation succeeds. If the provider is unavailable, the credential remains available for a retry; use `imprun auth logout --local-only` only when intentionally abandoning remote revocation. Imported `wf-cli` credentials are supported by discovering the revocation endpoint from their stored issuer. Direct Cell credentials have no generic remote revocation contract and are removed locally. Logging out clears every local context that references the same host and account so no context falsely reports a deleted credential.

CLI token revocation does not end the central Identity browser session. Use the product's explicit central logout flow when that separate session must end.

Inspect or change the workspace without creating another login:

```shell
imprun workspace show
imprun workspace list
imprun workspace view team
imprun workspace use team
```

`workspace use` probes the target with the current credential before updating the context. A failed authorization therefore leaves the previous workspace unchanged. The global list and view endpoints require an instance-admin credential on a direct Cell or equivalent authorization from a hosted product.

For automation, prefer the one-process environment override:

```powershell
$env:IMPRUN_TOKEN = "<WORKSPACE_TOKEN>"
imprun app list --summary
Remove-Item Env:IMPRUN_TOKEN
```

Existing automation can still name a token environment variable:

```powershell
$env:WORKSPACE_TOKEN = "<WORKSPACE_TOKEN>"
imprun context set hosted `
  --api-url https://cell.example.test `
  --workspace team `
  --token-env WORKSPACE_TOKEN `
  --use
```

The configuration stores only `WORKSPACE_TOKEN`, never its value. This is a compatibility workflow; new interactive use should use the system credential store.

Global flags override the selected context:

```shell
imprun --context staging app list --summary
imprun --api-url https://cell.example.test --workspace team app list
```

The primary process overrides are:

| Variable | Meaning |
| --- | --- |
| `IMPRUN_CONFIG` | Explicit non-secret configuration path |
| `IMPRUN_CONTEXT` | Selected context |
| `IMPRUN_API_URL` | Control Plane API base URL |
| `IMPRUN_WORKSPACE` | Selected workspace |
| `IMPRUN_ACTOR` | Direct-connection audit actor |
| `IMPRUN_TOKEN` | One-process bearer credential; never persisted |

## Publish a release

Run the primary workflow from the app directory or any child directory:

```shell
imprun app publish .
imprun app publish . --message "Ship invoice validation"
```

`imprun app publish` finds the nearest `windforce.json` and Git worktree, resolves the remote, branch, repository subpath, and full `HEAD` commit, finds or registers a matching source, synchronizes with an exact-commit precondition, and publishes the immutable bundle. If the remote branch does not resolve to that commit, publication fails before activation.

The worktree must be clean by default. `--allow-dirty` explicitly publishes only committed `HEAD`; uncommitted files, including a changed manifest, are ignored. The result identifies the app, exact commit, source, release, bundle digest, workspace, and context. Private repositories use a server-side credential reference:

```shell
imprun app publish . --creds-ref github-app-installation
imprun app publish . --source-id 12
```

For automation, select the immutable release identifier from the publication result and pass it to later commands instead of asking for the newest release:

```shell
imprun app publish . --json release_id
imprun release view example <RELEASE_ID>
```

A successful Cell response must contain `release_id`. `imprun` rejects an older or incompatible response that omits it rather than performing a race-prone release history lookup.

The low-level Register, Sync, and Publish sequence remains available for advanced operations:

```powershell
$env:GIT_ACCESS_TOKEN = "<TOKEN>"
imprun source probe `
  --repo-url https://git.example.test/team/app.git `
  --branch main `
  --auth-method pat `
  --access-token-env GIT_ACCESS_TOKEN

imprun source register `
  --name example `
  --repo-url https://git.example.test/team/app.git `
  --branch main `
  --subpath apps/example `
  --auth-method pat `
  --access-token-env GIT_ACCESS_TOKEN

imprun source list
imprun source sync 12 --expected-commit <FULL_COMMIT>
imprun source publish 12 --expected-commit <FULL_COMMIT> --message "Publish validated revision"
```

`publish` calls the existing release-publication endpoint. The legacy spelling `source deploy` remains accepted during migration. Workers never receive repository credentials and never contact Git.

## Inspect and activate releases

```shell
imprun release list example
imprun release view example <RELEASE_ID>
imprun release activate example <RELEASE_ID> \
  --reason "Restore the last known good release" \
  --yes
```

`activate` and its explicit `rollback` alias validate the immutable bundle in the Cell before changing the active release. A reason and `--yes` are required so automation cannot mutate the active release accidentally.

## Inspect apps and schemas

```shell
imprun app list --summary
imprun app show example
imprun app history example
imprun action show example health
imprun action schema example health
imprun app openapi example
imprun openapi
```

On a terminal, commands render readable labels or tables. Redirected output is compact JSON. Output flags work in the familiar command-local position:

```shell
imprun app show example --pretty
imprun app publish . --json app,commit,bundle_digest
imprun run show <RUN_ID> --json app,state --jq '.state'
imprun run show <RUN_ID> --template '{{.app}} {{.state}}'
```

`--json` accepts comma-separated top-level fields or `*`. Unknown fields fail with the available field names instead of silently returning empty data. `--jq` uses jq syntax, and `--template` uses a Go template. Standard output is reserved for results; progress and diagnostics use standard error.

## Run and inspect work

```shell
imprun run create example health --input '{"ping":true}'
imprun run wait example parse --input-file input.json --timeout 30s
imprun run show <RUN_ID>
imprun run watch <RUN_ID> --result
imprun run result <RUN_ID>
imprun run cancel <RUN_ID> --reason "operator request"

imprun job list --app example --status running
imprun job show <JOB_ID>
imprun job logs <JOB_ID> --tail-bytes 65536
```

`--input-file -` reads JSON from standard input. Job logs are written as the raw response so they can be piped.

`imprun run watch` prints state changes to standard error, polls no faster than 100 ms, and prints the terminal Run or successful result to standard output. Its default timeout is ten minutes.

## Shell completion and help

Help, version, and completion do not require a configured context or login:

```shell
imprun app publish --help
imprun help release
imprun completion bash
imprun completion zsh
imprun completion fish
imprun completion powershell
```

Load the generated script using the normal mechanism for the selected shell.

## Call the API directly

`imprun api` is the authenticated escape hatch for Control Plane operations that do not yet have a high-level command:

```shell
imprun api apps
imprun api git_sources
imprun api git_sources/12/sync --field expected_commit=<FULL_COMMIT>
imprun api /healthz
imprun api provisioning/apply --method POST --input request.json
```

A relative endpoint is resolved below `/api/w/<workspace>/`. A path beginning with `/` is resolved on the selected context host. Absolute URLs, scheme-relative URLs, fragments, and parent traversal are rejected so the selected credential cannot be redirected to another host. `--field` converts booleans, null, numbers, arrays, and objects; `--raw-field` always sends a string.

## Provisioning

```shell
imprun provisioning export --format yaml --output windforce.yaml
imprun provisioning apply --file windforce.yaml --dry-run
imprun provisioning apply --file windforce.yaml
```

Exported secret values remain redacted. Environment-specific secret resources must use the provisioning `valueFrom` contract.

## Exit codes

| Code | Meaning |
| ---: | --- |
| `0` | Command completed successfully |
| `1` | The requested operation completed unsuccessfully |
| `2` | Invalid command or arguments |
| `3` | Invalid local context or configuration |
| `4` | Authentication is required or invalid |
| `5` | The authenticated principal is not authorized |
| `10` | Local I/O or HTTP transport failure |
| `20` | Control Plane returned a 4xx response |
| `21` | Control Plane returned a 5xx response |

JSON API errors are written to standard error. This preserves the automation contract while keeping authentication, authorization, client, and server failures distinct.

## Troubleshooting

- Exit `4`: run `imprun auth status`, then `imprun auth login`. On a remote shell use `imprun auth login --no-browser`.
- Exit `5`: the credential is valid but lacks access to the selected hosted tenant or workspace. Check `imprun context show` and `imprun workspace show`; do not replace the credential with a Cloud management token.
- `Git worktree has uncommitted changes`: commit the intended files. Use `--allow-dirty` only when deliberately publishing committed `HEAD`.
- `409 Conflict` with `expected_commit`: push `HEAD` to the selected remote branch and retry. The CLI will not publish a different branch tip.
- `secure credential storage is unavailable`: restore Windows Credential Manager, macOS Keychain, or Linux Secret Service. The client does not fall back to a plaintext token file.
- `invalid API URL` for an older context: run `imprun --context <name> auth logout`, then `imprun context delete <name> --yes`. Use logout's `--local-only` override only if hosted revocation cannot be completed.
- A redirected API response is rejected. Configure the context with the final canonical Cell URL instead of relying on an HTTP redirect.

Diagnostics and JSON errors redact credential fields, Bearer values, OAuth codes, and Windforce token prefixes before writing standard error.

Released `imprun` archives and the documented command/exit-code behavior are the installed client contract.
