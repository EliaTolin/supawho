# supawho installer for Windows (PowerShell)
#
#   irm https://raw.githubusercontent.com/EliaTolin/supawho/main/install.ps1 | iex
#
# Installs the latest release into %LOCALAPPDATA%\Programs\supawho and adds it to
# the user PATH. Pass a version as the first arg to pin one, e.g.:
#   & ([scriptblock]::Create((irm .../install.ps1))) 1.4.0

param([string]$Version = "latest")

$ErrorActionPreference = "Stop"
$Repo = "EliaTolin/supawho"

Write-Host "`n  Installing supawho...`n"

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) {
  if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "x86_64" }
} else {
  throw "Unsupported architecture"
}

# Resolve the tag to install
if ($Version -eq "latest") {
  $rel = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
  $tag = $rel.tag_name
} else {
  $tag = "v" + ($Version -replace '^v', '')
}
$ver = $tag -replace '^v', ''

$archive = "supawho_${ver}_Windows_${arch}.zip"
$baseUrl = "https://github.com/$Repo/releases/download/$tag"
$tmp = Join-Path $env:TEMP ("supawho_" + [System.Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
  Write-Host "  -> Downloading $archive"
  Invoke-WebRequest "$baseUrl/$archive" -OutFile "$tmp\$archive"

  # Verify checksum
  try {
    Invoke-WebRequest "$baseUrl/checksums.txt" -OutFile "$tmp\checksums.txt"
    $expected = (Select-String -Path "$tmp\checksums.txt" -Pattern ([regex]::Escape($archive)) |
      ForEach-Object { ($_ -split '\s+')[0] } | Select-Object -First 1)
    if ($expected) {
      $actual = (Get-FileHash "$tmp\$archive" -Algorithm SHA256).Hash.ToLower()
      if ($actual -ne $expected.ToLower()) { throw "Checksum mismatch - aborting." }
    }
  } catch [System.Net.WebException] {
    # checksums.txt missing on very old releases; skip verification
  }

  Expand-Archive -Path "$tmp\$archive" -DestinationPath $tmp -Force

  $installDir = Join-Path $env:LOCALAPPDATA "Programs\supawho"
  New-Item -ItemType Directory -Path $installDir -Force | Out-Null
  Copy-Item "$tmp\supawho.exe" (Join-Path $installDir "supawho.exe") -Force

  # Add to the user PATH if not already present
  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "  -> Added $installDir to your PATH (restart your terminal)"
  }

  Write-Host "`n  supawho $tag installed to $installDir`n"
  Write-Host "  Run 'supawho help' to get started (open a new terminal first).`n"
}
finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
