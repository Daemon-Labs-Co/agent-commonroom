# Run from PowerShell. The checkout and Go installation belong inside WSL2.
param(
    [Parameter(Mandatory = $true)][string]$LinuxRepoPath,
    [ValidateSet('--check', '--build', '--migrate', '--start')]
    [string]$Mode = '--build'
)
$ErrorActionPreference = 'Stop'
if (-not (Get-Command wsl.exe -ErrorAction SilentlyContinue)) {
    throw 'Install WSL2 and a Linux distribution first; see docs/installation.md.'
}
if (-not $LinuxRepoPath.StartsWith('/')) { throw 'Use an absolute Linux checkout path.' }
# Pass arguments directly; never interpolate the path into shell source.
& wsl.exe --cd $LinuxRepoPath --exec bash scripts/setup-linux.sh $Mode
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
