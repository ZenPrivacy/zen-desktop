# Installs or uninstalls the canary makensis script.
#
# Usage:
#   setup-canary-makensis.ps1 install <path-to-canary-makensis.ps1>
#   setup-canary-makensis.ps1 uninstall

param(
    [Parameter(Mandatory, Position=0)]
    [ValidateSet('install', 'uninstall')]
    [string]$Action,

    [Parameter(Position=1)]
    [string]$CanaryScript
)

switch ($Action) {
    'install' {
        if (-not $CanaryScript) {
            Write-Error "install requires the path to the canary script"
            exit 1
        }

        $makensisExe = (Get-Command makensis.exe).Source
        $dir = Split-Path $makensisExe

        # Rename the real makensis so it can be restored later.
        Rename-Item $makensisExe (Join-Path $dir 'makensis-real.exe')

        # Copy the canary PowerShell script into the NSIS directory.
        Copy-Item $CanaryScript (Join-Path $dir 'canary-makensis.ps1')

        # Create a .cmd wrapper that forwards all arguments to the canary.
        # With the real .exe renamed, Windows resolves "makensis" to this .cmd.
        $wrapper = "@echo off`r`npowershell -NoProfile -ExecutionPolicy Bypass -File `"%~dp0canary-makensis.ps1`" %*`r`nexit /b 0"
        Set-Content (Join-Path $dir 'makensis.cmd') $wrapper

        Write-Host "canary makensis installed at $dir"
    }
    'uninstall' {
        $found = Get-Command makensis -ErrorAction SilentlyContinue
        $dir = Split-Path $found.Source

        Remove-Item (Join-Path $dir 'makensis.cmd') -ErrorAction SilentlyContinue
        Remove-Item (Join-Path $dir 'canary-makensis.ps1') -ErrorAction SilentlyContinue
        Rename-Item (Join-Path $dir 'makensis-real.exe') 'makensis.exe'

        Write-Host "real makensis restored at $dir"
    }
}
