# Checks, adds or removes a folder on the current user's PATH (HKCU\Environment).
# Used by the NSIS installer (project.nsi) and `task install-windows` (install.ps1).
#
#   -Mode check   exits 0 if the folder is on the user or system PATH, 1 if not
#   -Mode add     appends it to the user PATH (no-op if already there)
#   -Mode remove  drops it from the user PATH (no-op if absent)
#
# The value stays REG_EXPAND_SZ so entries like %USERPROFILE%\bin keep expanding,
# and running processes are told the environment changed, so new terminals see it.
param(
    [Parameter(Mandatory)][string]$Dir,
    [Parameter(Mandatory)][ValidateSet('check', 'add', 'remove')][string]$Mode
)

$ErrorActionPreference = 'Stop'
$Dir = $Dir.TrimEnd('\')

function Test-Entry([string]$p) {
    [bool]($p -split ';' | Where-Object { $_ -and [Environment]::ExpandEnvironmentVariables($_).TrimEnd('\') -eq $Dir })
}

$key = [Microsoft.Win32.Registry]::CurrentUser.CreateSubKey('Environment')
$userPath = [string]$key.GetValue('Path', '', 'DoNotExpandEnvironmentNames')

if ($Mode -eq 'check') {
    $machinePath = [string][Environment]::GetEnvironmentVariable('Path', 'Machine')
    if ((Test-Entry $userPath) -or (Test-Entry $machinePath)) { exit 0 } else { exit 1 }
}

$parts = @($userPath -split ';' | Where-Object { $_ -and [Environment]::ExpandEnvironmentVariables($_).TrimEnd('\') -ne $Dir })
if ($Mode -eq 'add') { $parts += $Dir }

if ($parts.Count) { $key.SetValue('Path', ($parts -join ';'), 'ExpandString') }
elseif ($null -ne $key.GetValue('Path')) { $key.DeleteValue('Path') }

# WM_SETTINGCHANGE to all top-level windows, so Explorer hands new terminals the new PATH.
Add-Type -Namespace Win32 -Name Env -MemberDefinition '[DllImport("user32.dll", CharSet = CharSet.Unicode)] public static extern IntPtr SendMessageTimeout(IntPtr h, uint m, UIntPtr w, string l, uint f, uint t, out UIntPtr r);'
$r = [UIntPtr]::Zero
[void][Win32.Env]::SendMessageTimeout([IntPtr]0xffff, 0x1a, [UIntPtr]::Zero, 'Environment', 2, 5000, [ref]$r)
