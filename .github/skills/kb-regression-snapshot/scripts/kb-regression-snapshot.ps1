param(
  [Parameter(Position=0)]
  [ValidateSet("capture", "verify")]
  [string]$Mode = "verify",
  [string]$SliceId,
  [string]$SpecPath,
  [string]$SnapshotDir = ".kb/snapshots",
  [string]$BaseUrl = $env:DEV_SERVER_URL
)

$ErrorActionPreference = "Stop"

function Resolve-Url([string]$Url) {
  if ($Url -match '^https?://') { return $Url }
  if (-not $BaseUrl) { $script:BaseUrl = "http://localhost:3000" }
  return ($script:BaseUrl.TrimEnd("/") + "/" + $Url.TrimStart("/"))
}

function Assert-RouteStatus($Check) {
  $url = Resolve-Url $Check.url
  try {
    $response = Invoke-WebRequest -Uri $url -UseBasicParsing -Method GET
    $status = [int]$response.StatusCode
  } catch {
    $status = [int]$_.Exception.Response.StatusCode
  }
  if ($status -ne [int]$Check.expected_status) {
    throw "route-status $url expected $($Check.expected_status) observed $status"
  }
}

function Assert-ApiSchema($Check) {
  Assert-RouteStatus $Check
  if ($Check.required_fields) {
    $json = (Invoke-WebRequest -Uri (Resolve-Url $Check.url) -UseBasicParsing).Content | ConvertFrom-Json
    foreach ($field in $Check.required_fields) {
      if (-not ($json.PSObject.Properties.Name -contains $field)) {
        throw "api-schema $($Check.url) missing field $field"
      }
    }
  }
}

function Assert-FileChecksum($Check) {
  if (-not (Test-Path -LiteralPath $Check.path)) { throw "file-checksum missing $($Check.path)" }
  $actual = (Get-FileHash -LiteralPath $Check.path -Algorithm SHA256).Hash.ToLowerInvariant()
  if ($actual -ne [string]$Check.sha256.ToLowerInvariant()) {
    throw "file-checksum $($Check.path) expected $($Check.sha256) observed $actual"
  }
}

function Assert-Cli($Check) {
  $output = Invoke-Expression $Check.command 2>&1 | Out-String
  $exit = if ($null -eq $LASTEXITCODE) { 0 } else { $LASTEXITCODE }
  if ($exit -ne [int]$Check.expected_exit_code) {
    throw "cli $($Check.command) expected exit $($Check.expected_exit_code) observed $exit"
  }
  if ($Check.expected_output_substring -and -not $output.Contains([string]$Check.expected_output_substring)) {
    throw "cli $($Check.command) missing output substring $($Check.expected_output_substring)"
  }
}

function Assert-DomElement($Check) {
  $node = @"
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  const errors = [];
  page.on('console', msg => { if (msg.type() === 'error') errors.push(msg.text()); });
  await page.goto(process.argv[2], { waitUntil: 'domcontentloaded' });
  const loc = page.locator(process.argv[3]).first();
  await loc.waitFor({ state: 'visible', timeout: 10000 });
  const text = (await loc.textContent()) || '';
  const expected = process.argv[4] || '';
  const pattern = process.argv[5] || '';
  if (expected && text.trim() !== expected) throw new Error(`expected text ${expected} observed ${text.trim()}`);
  if (pattern && !(new RegExp(pattern).test(text))) throw new Error(`text ${text} did not match ${pattern}`);
  if (errors.length !== 0) throw new Error(`console errors: ${errors.join(' | ')}`);
  await browser.close();
})().catch(err => { console.error(err.message); process.exit(1); });
"@
  $tmp = Join-Path ([System.IO.Path]::GetTempPath()) "kb-snapshot-dom-$([guid]::NewGuid()).js"
  Set-Content -LiteralPath $tmp -Value $node -Encoding UTF8
  try {
    node $tmp (Resolve-Url $Check.url) $Check.selector $Check.expected_text $Check.expected_text_pattern
    if ($LASTEXITCODE -ne 0) { throw "dom-element $($Check.url) $($Check.selector) failed" }
  } finally {
    Remove-Item -LiteralPath $tmp -Force -ErrorAction SilentlyContinue
  }
}

function Invoke-Check($Check) {
  switch ($Check.type) {
    "route-status" { Assert-RouteStatus $Check }
    "api-schema" { Assert-ApiSchema $Check }
    "file-checksum" { Assert-FileChecksum $Check }
    "cli" { Assert-Cli $Check }
    "dom-element" { Assert-DomElement $Check }
    default { throw "unknown snapshot check type: $($Check.type)" }
  }
}

New-Item -ItemType Directory -Force -Path $SnapshotDir | Out-Null

if ($Mode -eq "capture") {
  if (-not $SliceId -or -not $SpecPath) { throw "capture requires -SliceId and -SpecPath" }
  $snapshot = Get-Content -LiteralPath $SpecPath -Raw | ConvertFrom-Json
  $snapshot | Add-Member -NotePropertyName slice_id -NotePropertyValue $SliceId -Force
  $snapshot | Add-Member -NotePropertyName captured_at -NotePropertyValue ([DateTimeOffset]::UtcNow.ToString("o")) -Force
  foreach ($check in $snapshot.checks) { Invoke-Check $check }
  $out = Join-Path $SnapshotDir "$SliceId.json"
  $snapshot | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $out -Encoding UTF8
  "snapshot-capture: PASS $SliceId -> $out"
  exit 0
}

$files = Get-ChildItem -Path $SnapshotDir -Filter "*.json" -File -ErrorAction SilentlyContinue | Sort-Object Name
$count = 0
foreach ($file in $files) {
  $snapshot = Get-Content -LiteralPath $file.FullName -Raw | ConvertFrom-Json
  foreach ($check in $snapshot.checks) { Invoke-Check $check }
  $count++
}
"snapshot-verify: PASS $count/$count snapshots"
exit 0
