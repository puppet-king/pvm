#!/usr/bin/env pwsh
param(
  [String]$Version = "latest",
  [Switch]$NoPathUpdate = $false,
  [Switch]$DownloadWithoutCurl = $false
)

$ErrorActionPreference = "Stop"

$Arch = (Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment').PROCESSOR_ARCHITECTURE
if ($Arch -ne "AMD64") {
  Write-Output "Install Failed:"
  Write-Output "pvm for Windows is only available for x64 Windows.`n"
  exit 1
}

function Publish-Env {
  if (-not ("Win32.NativeMethods" -as [Type])) {
    Add-Type -Namespace Win32 -Name NativeMethods -MemberDefinition @"
[DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
public static extern IntPtr SendMessageTimeout(
    IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam,
    uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);
"@
  }

  $HWND_BROADCAST = [IntPtr]0xffff
  $WM_SETTINGCHANGE = 0x1a
  $result = [UIntPtr]::Zero

  [Win32.NativeMethods]::SendMessageTimeout(
    $HWND_BROADCAST,
    $WM_SETTINGCHANGE,
    [UIntPtr]::Zero,
    "Environment",
    2,
    5000,
    [ref]$result
  ) | Out-Null
}

function Write-Env {
  param([String]$Key, [String]$Value)

  $RegisterKey = Get-Item -Path 'HKCU:'
  $EnvRegisterKey = $RegisterKey.OpenSubKey('Environment', $true)

  if ($null -eq $Value) {
    $EnvRegisterKey.DeleteValue($Key)
  } else {
    $RegistryValueKind = if ($Value.Contains('%')) {
      [Microsoft.Win32.RegistryValueKind]::ExpandString
    } elseif ($EnvRegisterKey.GetValue($Key)) {
      $EnvRegisterKey.GetValueKind($Key)
    } else {
      [Microsoft.Win32.RegistryValueKind]::String
    }

    $EnvRegisterKey.SetValue($Key, $Value, $RegistryValueKind)
  }

  Publish-Env
}

function Get-Env {
  param([String]$Key)

  $RegisterKey = Get-Item -Path 'HKCU:'
  $EnvRegisterKey = $RegisterKey.OpenSubKey('Environment')
  $EnvRegisterKey.GetValue($Key, $null, [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
}

function Install-Pvm {
  param([String]$Version)

  $NormalizedVersion = $Version.Trim()
  if ($NormalizedVersion -match '^v') {
    $NormalizedVersion = $NormalizedVersion.Substring(1)
  }

  $PvmRoot = if ($env:PVM_INSTALL) { $env:PVM_INSTALL } else { "${Home}\.pvm" }
  $PvmBin = Join-Path $PvmRoot 'bin'
  $PvmExe = Join-Path $PvmBin 'pvm.exe'
  $DownloadVersion = if ($NormalizedVersion -eq "latest") { "latest/download" } else { "download/$NormalizedVersion" }
  $Url = "https://github.com/puppet-king/pvm/releases/$DownloadVersion/pvm.exe"
  $TempExe = Join-Path $PvmBin 'pvm.exe.tmp'

  New-Item -ItemType Directory -Force -Path $PvmBin | Out-Null
  Remove-Item -Force $TempExe -ErrorAction SilentlyContinue

  if (-not $DownloadWithoutCurl) {
    curl.exe "-#SfLo" "$TempExe" "$Url"
  }

  if ($DownloadWithoutCurl -or ($LASTEXITCODE -ne 0)) {
    if (-not $DownloadWithoutCurl) {
      Write-Warning "The command 'curl.exe $Url -o $TempExe' exited with code ${LASTEXITCODE}`nTrying an alternative download method..."
    }
    try {
      Invoke-RestMethod -Uri $Url -OutFile $TempExe
    } catch {
      Write-Output "Install Failed:"
      Write-Output "Could not download $Url"
      exit 1
    }
  }

  if (!(Test-Path $TempExe)) {
    Write-Output "Install Failed:"
    Write-Output "The downloaded file '$TempExe' does not exist.`n"
    exit 1
  }

  try {
    Remove-Item -Force $PvmExe -ErrorAction Stop
  } catch [System.Management.Automation.ItemNotFoundException] {
  } catch [System.UnauthorizedAccessException] {
    Write-Output "Install Failed:"
    Write-Output "An existing pvm.exe could not be replaced. Please close any running pvm processes and try again.`n"
    exit 1
  } catch {
    Write-Output "Install Failed:"
    Write-Output "An unexpected error occurred while removing the existing installation."
    Write-Output $_
    exit 1
  }

  Move-Item -Force $TempExe $PvmExe

  $PathValue = Get-Env -Key 'Path'
  $Path = @()
  if ($PathValue) {
    $Path = $PathValue -split ';'
  }

  if ($Path -notcontains $PvmBin) {
    if (-not $NoPathUpdate) {
      $Path += $PvmBin
      Write-Env -Key 'Path' -Value ($Path -join ';')
      $CurrentProcessPath = @()
      if ($env:PATH) {
        $CurrentProcessPath = $env:PATH -split ';'
      }

      if ($CurrentProcessPath -notcontains $PvmBin) {
        $env:PATH = if ($env:PATH) { "$env:PATH;$PvmBin" } else { $PvmBin }
      }
    } else {
      Write-Output "Skipping adding '${PvmBin}' to the user's %PATH%`n"
    }
  }

  $C_RESET = [char]27 + "[0m"
  $C_GREEN = [char]27 + "[1;32m"

  Write-Output "${C_GREEN}pvm was installed successfully!${C_RESET}"
  Write-Output "The binary is located at ${PvmExe}`n"
  Write-Output "Restart your terminal/editor, then run 'pvm help'. You should see pvm!`n"
}

Install-Pvm -Version $Version
