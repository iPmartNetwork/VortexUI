Get-ChildItem "$PSScriptRoot\..\src\pages" -Filter "*.tsx" -Recurse | ForEach-Object {
  $c = [IO.File]::ReadAllText($_.FullName)
  $n = $c
  $n = $n.Replace("animate-fade-in", "animate-page-enter")
  if ($n -notmatch "animate-page-enter" -and $n -match 'className="space-y-6"') {
    $n = $n.Replace('className="space-y-6"', 'className="space-y-6 animate-page-enter"')
  }
  if ($n -notmatch "animate-page-enter" -and $n -match 'className="space-y-8"') {
    $n = $n.Replace('className="space-y-8"', 'className="space-y-8 animate-page-enter"')
  }
  if ($n -notmatch "animate-page-enter" -and $n -match 'className="space-y-5"') {
    $n = $n.Replace('className="space-y-5"', 'className="space-y-5 animate-page-enter"')
  }
  $n = $n.Replace("animate-page-enter animate-page-enter", "animate-page-enter")
  if ($n -ne $c) {
    [IO.File]::WriteAllText($_.FullName, $n)
    Write-Host "Updated $($_.Name)"
  }
}
