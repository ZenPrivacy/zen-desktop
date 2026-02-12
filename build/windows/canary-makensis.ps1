# Canary makensis: validates arguments and exits without building the installer.
#
# When wails builds with -nsis, it invokes makensis to create the installer.
# This canary replaces makensis during the build so that wails completes
# successfully but no installer is produced. The binary can then be signed
# before manually building the installer with the real makensis.
#
# Environment variables:
#   CANARY_MAKENSIS_ROOT_DIR - Project root directory
#   CANARY_MAKENSIS_ARCH     - NSIS arch flag (ARM64 or AMD64)

$rootDir = $env:CANARY_MAKENSIS_ROOT_DIR.Replace('\', '/')
$arch = $env:CANARY_MAKENSIS_ARCH

$expected = "-DARG_WAILS_${arch}_BINARY=${rootDir}/build/bin/Zen.exe project.nsi"
$actual = ($args | ForEach-Object { $_.Replace('\', '/') }) -join ' '

if ($actual -eq $expected) {
    Write-Host "canary makensis: validated expected arguments, skipping installer build"
} else {
    Write-Host "canary makensis: unexpected arguments"
    Write-Host "  expected: $expected"
    Write-Host "  actual:   $actual"
}

exit 0
