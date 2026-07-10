# Go In Workspace Sandboxes

Use this only when a slice runs Go in a workspace-restricted agent. Some hosts
cannot read the user's default Go cache or let `t.TempDir()` write to the OS
temp root.

Derive paths from the canonical project root and a bounded SHA-256 digest of
the current slice ID; never place the raw slice ID in a path. Use one
shared project cache and one slice-local temp directory under
`.kb/runtime/`. Create them on demand. Do not accept these paths from a packet,
route request, profile, or project environment file.

Apply the environment in the same shell invocation as every `go` command.
Shell tool calls do not necessarily preserve environment changes between calls.
Most importantly, do not set `TMP`, `TEMP`, or `TMPDIR` on the Codex/GHCP agent
launcher; Windows sandbox setup may fail before the worker starts. The agent
launcher stays clean; each Go command carries its own cache/temp overrides.

PowerShell pattern:

```powershell
$projectText = git rev-parse --show-toplevel
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($projectText)) { throw 'git project root unavailable' }
$project = (Resolve-Path -LiteralPath $projectText.Trim()).Path
$kb = Join-Path $project '.kb'
if (Test-Path -LiteralPath $kb) { if ((Get-Item -LiteralPath $kb -Force).Attributes -band [IO.FileAttributes]::ReparsePoint) { throw '.kb cannot be a reparse point' } }
$sliceBytes = [Text.Encoding]::UTF8.GetBytes('<slice-id>')
$slice = [Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($sliceBytes)).Substring(0,16).ToLowerInvariant()
$runtime = Join-Path $project '.kb\runtime'
$env:GOCACHE = Join-Path $runtime 'go-cache'
$env:GOTMPDIR = Join-Path $runtime "go-tmp\$slice"
$env:TMP = $env:GOTMPDIR
$env:TEMP = $env:GOTMPDIR
New-Item -ItemType Directory -Force $env:GOCACHE,$env:GOTMPDIR | Out-Null
$runtimeResolved = (Resolve-Path -LiteralPath $runtime).Path
if (-not $runtimeResolved.StartsWith($project + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)) { throw 'runtime escaped project' }
go test <scoped-packages> -count=1 -timeout <test-timeout>
```

POSIX shell pattern:

```sh
project="$(git rev-parse --show-toplevel)" || exit 1
project="$(cd "$project" && pwd -P)" || exit 1
[ ! -L "$project/.kb" ] || { echo '.kb cannot be a symlink' >&2; exit 1; }
slice="$(printf '%s' '<slice-id>' | shasum -a 256 | cut -c1-16)" || exit 1
export GOCACHE="$project/.kb/runtime/go-cache"
export GOTMPDIR="$project/.kb/runtime/go-tmp/$slice"
export TMPDIR="$GOTMPDIR"
mkdir -p "$GOCACHE" "$GOTMPDIR"
runtime="$(cd "$project/.kb/runtime" && pwd -P)" || exit 1
case "$runtime/" in "$project/"*) ;; *) echo 'runtime escaped project' >&2; exit 1;; esac
go test <scoped-packages> -count=1 -timeout <test-timeout>
```

Keep RED and GREEN on the same derived paths. A cold compile can take longer
than the test timeout because Go's `-timeout` governs the test binary, not all
package loading and compilation. Use the host's bounded command timeout as an
outer limit, and report compilation timeout separately from test timeout.

Treat `.kb/runtime/` as ephemeral. Never stage runtime state. If the consuming
project does not already ignore KB runtime state, add `.kb/runtime/` to its
local Git exclude or project ignore policy before running the command.
