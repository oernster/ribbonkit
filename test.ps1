# Verifies ribbonkit: formatting, vet, staticcheck, the whole suite, the web half and the coverage
# floors.
#
#   ./test.ps1              run everything
#   ./test.ps1 -Floor 95    run with a different floor over domain and application
#
# Every floor is the measured number, not a target. A floor picked from an aspiration only teaches
# people to lower it; one at the measured number fails the moment cover is lost.
param(
    [double]$Floor = 100
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

# Domain and application are the layers a test reaches with no filesystem, clock or display, so
# anything short of 100 percent there is a decision nobody made.
$gated = './domain/...', './application/...'

# The Go tools are pointed at this list rather than at ./..., which would reach into node_modules.
$packages = go list ./... | Where-Object { $_ -notmatch '/node_modules/' }
if ($LASTEXITCODE -ne 0) { throw "go list failed with exit code $LASTEXITCODE" }

Write-Host 'Checking formatting...'
$unformatted = gofmt -l . | Where-Object { $_ -notmatch '^node_modules' }
if ($unformatted) { throw "gofmt reports unformatted files:`n$($unformatted -join "`n")" }

Write-Host 'Vetting...'
go vet $packages
if ($LASTEXITCODE -ne 0) { throw "go vet failed with exit code $LASTEXITCODE" }

# Pinned so a new release of the checker cannot fail a change that touched nothing it reads. Raise
# it on purpose, having read what the new version reports.
$staticcheckVersion = 'v0.8.1'
Write-Host "Running staticcheck $staticcheckVersion..."
go run "honnef.co/go/tools/cmd/staticcheck@$staticcheckVersion" $packages
if ($LASTEXITCODE -ne 0) { throw "staticcheck failed with exit code $LASTEXITCODE" }

Write-Host 'Running the whole suite...'
go test -count=1 $packages
if ($LASTEXITCODE -ne 0) { throw "go test failed with exit code $LASTEXITCODE" }

# The web half is held to the same bar in its own runner. A missing node_modules stops the gate
# rather than skipping it: a check that quietly does not run is the one that is not there on the day.
Write-Host 'Checking the web half...'
if (-not (Test-Path (Join-Path $root 'node_modules'))) {
    throw "the web half's tools are not installed: run npm install in $root, then run this again"
}
foreach ($check in 'lint', 'typecheck', 'test') {
    npm run $check
    if ($LASTEXITCODE -ne 0) { throw "npm run $check failed with exit code $LASTEXITCODE" }
}

Write-Host "Measuring coverage of $($gated -join ', ')..."
$profilePath = Join-Path ([System.IO.Path]::GetTempPath()) 'ribbonkit-coverage.out'
try {
    go test -count=1 "-coverprofile=$profilePath" @gated
    if ($LASTEXITCODE -ne 0) { throw "the coverage run failed with exit code $LASTEXITCODE" }
    $summary = go tool cover "-func=$profilePath"
    if ($LASTEXITCODE -ne 0) { throw "go tool cover failed with exit code $LASTEXITCODE" }
    $total = ($summary | Select-Object -Last 1)
    if ($total -notmatch '([0-9]+(?:\.[0-9]+)?)%\s*$') { throw "could not read a total from: $total" }
    $percent = [double]$Matches[1]
    if ($percent -lt $Floor) {
        Write-Host 'Not covered:'
        $summary | Where-Object { $_ -notmatch '100\.0%\s*$' -and $_ -notmatch '^total:' } | ForEach-Object { Write-Host "  $_" }
        throw "coverage is $percent%, below the floor of $Floor%"
    }
    Write-Host "Coverage $percent%, floor $Floor%."
} finally {
    if (Test-Path $profilePath) { Remove-Item $profilePath -Force }
}

# The rest of the kit, each package held at the number it reaches. TESTING.md names what each
# shortfall is: error returns that only a failing disk, registry or display driver can produce; in
# the window, the calls that reach Wails and the start, listening and stop that only Wails runs; in
# the setup program, the window it opens, the calls into Wails and Main's reading of the real machine; in structure,
# the checks themselves, which the structural suite calls and planted violations prove, while Go
# counts only the package's own tests of its recognisers and arithmetic.
$measured = [ordered]@{
    './ui/window'                  = 93
    './infrastructure/appdata'     = 100
    './infrastructure/atomicfile'  = 85
    './infrastructure/heldfile'    = 100
    './infrastructure/settingsfile' = 97
    './infrastructure/desktop'     = 49
    './infrastructure/iconscale'   = 100
    './infrastructure/runlog'      = 77
    './infrastructure/monitors'    = 82
    './infrastructure/occupancy'   = 90
    './infrastructure/setup'       = 84
    './infrastructure/delivery'    = 96
    './installer'                  = 67
    './infrastructure/startup'     = 80
    './infrastructure/system'      = 100
    './infrastructure/update'      = 100
    './infrastructure/zones'       = 100
    './structure'                  = 33
}

Write-Host 'Measuring the rest of the kit...'
foreach ($package in $measured.Keys) {
    $floorHere = $measured[$package]
    $reported = go test -count=1 -cover $package
    # The report is held to read the coverage from, so a failing test's own words would otherwise
    # never reach the log.
    if ($LASTEXITCODE -ne 0) {
        $reported | ForEach-Object { Write-Host $_ }
        throw "$package failed with exit code $LASTEXITCODE"
    }
    $line = $reported | Where-Object { $_ -match 'coverage: ' } | Select-Object -First 1
    if ($line -notmatch 'coverage: ([0-9]+(?:\.[0-9]+)?)%') { throw "could not read a coverage figure for ${package}: $line" }
    $reached = [double]$Matches[1]
    if ($reached -lt $floorHere) { throw "$package is at $reached%, below its floor of $floorHere%" }
    Write-Host ("  {0,-30} {1,5}%  floor {2}%" -f $package, $reached, $floorHere)
}

Write-Host 'All green.'
