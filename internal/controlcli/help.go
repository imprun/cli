package controlcli

import (
	"fmt"
	"io"
	"strings"
)

var imprunCommandHelp = map[string]string{
	"auth": `Authenticate imprun with the selected context.

USAGE
  imprun auth login [--web | --with-token] [--no-browser] [--account label]
  imprun auth switch <account>
  imprun auth status
  imprun auth logout [--local-only]`,
	"auth login": `Authenticate with a hosted Identity provider or a direct Cell credential.

USAGE
  imprun auth login [flags]

FLAGS
  --web             Use hosted Device Authorization
  --with-token      Read one direct Cell credential from standard input
  --no-browser      Print the verification URL instead of opening a browser
  --account string  Local account label`,
	"auth logout": `Remove the selected account credential.

Hosted credentials are revoked at the Identity provider before local state is
removed. A revocation failure preserves the local credential so the operation
can be retried. This does not end the central browser session.

USAGE
  imprun auth logout [--local-only]

FLAGS
  --local-only  Skip hosted token revocation and remove only local state`,
	"context": `Manage non-secret Cell connection contexts.

USAGE
  imprun context list
  imprun context show [name]
  imprun context set <name> --api-url <url> [flags]
  imprun context use <name>
  imprun context delete <name> --yes`,
	"context set": `Create or update a connection context.

USAGE
  imprun context set <name> --api-url <url> [flags]

FLAGS
  --workspace string  Workspace ID
  --actor string      Direct Cell audit actor
  --token-env string  Compatibility bearer-token environment variable name
  --use               Select the context after saving`,
	"context delete": `Delete one non-secret Cell connection context.

An authenticated context must be logged out first so its credential lifecycle
is explicit. Select another context before deleting the current one when more
than one context exists.

USAGE
  imprun context delete <name> --yes`,
	"workspace": `Inspect and select a workspace in the current context.

USAGE
  imprun workspace list
  imprun workspace show
  imprun workspace view <workspace>
  imprun workspace use <workspace>

Workspace switching verifies access before updating the context. Listing and
viewing the global workspace registry require instance-admin or equivalent
hosted delegation.`,
	"source": `Manage low-level Git source connections.

USAGE
  imprun source list
  imprun source register [flags]
  imprun source probe [flags]
  imprun source sync <source-id> [--expected-commit sha]
  imprun source publish <source-id> [--expected-commit sha] [--message text]`,
	"source register": `Register a Git source. Credentials are read from an environment variable and stored server-side.

USAGE
  imprun source register --name <name> --repo-url <url> [flags]

FLAGS
  --branch string            Remote branch (default "main")
  --subpath string           App directory inside the repository
  --creds-ref string         Existing server-side credential reference
  --auth-method string       Credential type for a new reference
  --access-token-env string  Environment variable containing a Git token`,
	"source probe": `Validate a Git source without registering it.

USAGE
  imprun source probe --repo-url <url> [flags]`,
	"source sync": `Synchronize and validate the selected remote branch.

USAGE
  imprun source sync <source-id> [--expected-commit sha]`,
	"source publish": `Publish the latest synchronized source candidate as an immutable release.

USAGE
  imprun source publish <source-id> [--expected-commit sha] [--message text]`,
	"app": `Manage apps in the selected workspace.

USAGE
  imprun app publish [path] [flags]
  imprun app list [--summary]
  imprun app show <app>
  imprun app history <app>
  imprun app source <app>
  imprun app openapi <app>`,
	"app publish": `Publish the exact Git commit containing a Windforce app.

The command finds windforce.json, resolves the Git repository, branch, subpath,
and HEAD commit, finds or registers a matching source, synchronizes with an
exact-commit precondition, and publishes an immutable release.

USAGE
  imprun app publish [path] [flags]

FLAGS
  --source-id int       Use an existing Git source ID
  --source-name string  Select or register a source by name
  --creds-ref string    Server-side credential reference for a new private source
  --remote string       Git remote name
  --branch string       Remote branch
  --message string      Release audit message
  --allow-dirty         Ignore uncommitted files and publish HEAD only
  --quiet               Suppress progress messages`,
	"release": `Inspect and change immutable app releases.

USAGE
  imprun release list <app>
  imprun release view <app> <release-id>
  imprun release activate <app> <release-id> --reason <text> --yes
  imprun release rollback <app> <release-id> --reason <text> --yes`,
	"release activate": `Make an existing immutable release active.

USAGE
  imprun release activate <app> <release-id> --reason <text> --yes`,
	"release rollback": `Restore an earlier immutable release.

USAGE
  imprun release rollback <app> <release-id> --reason <text> --yes`,
	"run": `Create, wait for, inspect, and cancel Runs.

USAGE
  imprun run create <app> <action> [flags]
  imprun run wait <app> <action> [flags]
  imprun run show <run-id>
  imprun run watch <run-id> [flags]
  imprun run result <run-id>
  imprun run cancel <run-id> [--reason text]`,
	"run create": `Create a Run and return immediately.

USAGE
  imprun run create <app> <action> [flags]

FLAGS
  --input string            JSON input (default "{}")
  --input-file string       JSON input file, or - for standard input
  --idempotency-key string  Principal-scoped idempotency key
  --correlation-id string   Caller correlation ID`,
	"run wait": `Create a Run and wait server-side for completion.

USAGE
  imprun run wait <app> <action> [flags]

FLAGS
  --input string            JSON input (default "{}")
  --input-file string       JSON input file, or - for standard input
  --idempotency-key string  Principal-scoped idempotency key
  --correlation-id string   Caller correlation ID
  --timeout duration        Server wait duration`,
	"run watch": `Poll a Run until it reaches a terminal state.

USAGE
  imprun run watch <run-id> [flags]

FLAGS
  --interval duration  Polling interval (default 2s)
  --timeout duration   Maximum wait duration (default 10m)
  --result             Print the result after a successful Run
  --quiet              Suppress state-change progress`,
	"job": `Inspect low-level Jobs.

USAGE
  imprun job list [flags]
  imprun job show <job-id>
  imprun job result <job-id>
  imprun job logs <job-id> [--tail-bytes n]
  imprun job cancel <job-id> [--reason text]`,
	"action": `Inspect an app action and its schemas.

USAGE
  imprun action show <app> <action>
  imprun action schema <app> <action>`,
	"provisioning": `Export or apply workspace provisioning documents.

USAGE
  imprun provisioning export [--format json|yaml] [--output path]
  imprun provisioning apply --file <path> [--dry-run]`,
	"completion": `Generate shell completion code.

USAGE
  imprun completion bash|zsh|fish|powershell`,
	"version": `Print the imprun version.

USAGE
  imprun version`,
	"openapi": `Print the selected workspace Control Plane OpenAPI document.

USAGE
  imprun openapi`,
	"api": `Call a Control Plane endpoint using the selected context and credential.

Relative endpoints are resolved below the selected workspace. Paths beginning
with / are resolved on the same context host. Absolute URLs are rejected.

USAGE
  imprun api <endpoint> [flags]

FLAGS
  --method string     GET, POST, PUT, PATCH, or DELETE
  --field key=value   Typed JSON field; repeatable
  --raw-field key=value
                      String JSON field; repeatable
  --input path        JSON request body file, or - for standard input`,
}

func requestedCommandHelp(args []string) ([]string, bool) {
	if len(args) == 0 {
		return nil, false
	}
	if args[0] == "help" {
		return args[1:], true
	}
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return commandHelpPath(args), true
		}
	}
	return nil, false
}

func commandHelpPath(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	path := []string{args[0]}
	if len(args) > 1 && !strings.HasPrefix(args[1], "-") {
		candidate := strings.Join(args[:2], " ")
		if _, ok := imprunCommandHelp[candidate]; ok {
			path = append(path, args[1])
		}
	}
	return path
}

func printCommandHelp(writer io.Writer, program string, path []string) bool {
	if len(path) == 0 {
		printUsage(writer, program)
		return true
	}
	if program != imprunProgram.Name {
		printUsage(writer, program)
		return true
	}
	key := strings.Join(path, " ")
	help, ok := imprunCommandHelp[key]
	if !ok && len(path) > 1 {
		help, ok = imprunCommandHelp[path[0]]
	}
	if !ok {
		return false
	}
	_, _ = fmt.Fprintln(writer, help)
	return true
}
