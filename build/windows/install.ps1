# Registers an installed katdown.exe with Windows for the current user:
# a Start Menu shortcut, and katdown as an "Open with" handler for .md files.
# Windows won't let an app make itself the default; pick katdown once via
# Open with -> Always.
param([Parameter(Mandatory)][string]$Exe)

$ErrorActionPreference = 'Stop'

$lnk = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs\katdown.lnk'
$s = (New-Object -ComObject WScript.Shell).CreateShortcut($lnk)
$s.TargetPath = $Exe
$s.WorkingDirectory = $env:USERPROFILE
$s.Description = 'Markdown viewer'
$s.Save()

$classes = 'HKCU:\Software\Classes'
$open = "`"$Exe`" `"%1`""

New-Item -Force "$classes\katdown.md\DefaultIcon" | Set-ItemProperty -Name '(default)' -Value "`"$Exe`",0"
New-Item -Force "$classes\katdown.md\shell\open\command" | Set-ItemProperty -Name '(default)' -Value $open
Set-ItemProperty "$classes\katdown.md" -Name '(default)' -Value 'Markdown document'

foreach ($ext in '.md', '.markdown') {
    New-Item -Force "$classes\$ext\OpenWithProgids" | New-ItemProperty -Force -Name 'katdown.md' -Value '' | Out-Null
}

$app = "$classes\Applications\katdown.exe"
New-Item -Force "$app\shell\open\command" | Set-ItemProperty -Name '(default)' -Value $open
New-Item -Force "$app\SupportedTypes" | Out-Null
foreach ($ext in '.md', '.markdown') {
    New-ItemProperty -Force "$app\SupportedTypes" -Name $ext -Value '' | Out-Null
}
Set-ItemProperty $app -Name 'FriendlyAppName' -Value 'katdown'

# Tell Explorer the associations changed so the Open with list refreshes.
Add-Type -Namespace Win32 -Name Shell -MemberDefinition '[DllImport("shell32.dll")] public static extern void SHChangeNotify(int e, uint f, IntPtr a, IntPtr b);'
[Win32.Shell]::SHChangeNotify(0x08000000, 0, [IntPtr]::Zero, [IntPtr]::Zero)

Write-Host "Registered $Exe"
