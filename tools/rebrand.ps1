# Rebranding script: WuzAPI -> Miauzap
$rootPath = Get-Location

$filePatterns = @("*.go", "*.html", "*.js", "*.md", "*.yml", "*.yaml", ".env*", "Dockerfile*")

$replacements = @(
    @{Old = "WUZAPI"; New = "MIAUZAP"},
    @{Old = "WuzAPI"; New = "Miauzap"},
    @{Old = "Wuzapi"; New = "Miauzap"},
    @{Old = "wuzapi"; New = "miauzap"}
)

Write-Host "Starting rebranding process..."
Write-Host "Root path: $rootPath"

$totalFiles = 0
$totalReplacements = 0

foreach ($pattern in $filePatterns) {
    Write-Host "Processing files matching: $pattern"
    
    $files = Get-ChildItem -Path $rootPath -Filter $pattern -Recurse -File -ErrorAction SilentlyContinue | Where-Object { $_.FullName -notmatch '\\\.git\\' -and $_.FullName -notmatch '\\node_modules\\' -and $_.FullName -notmatch '\\vendor\\' -and $_.FullName -notmatch '\\tools\\' }
    
    foreach ($file in $files) {
        try {
            $content = Get-Content -Path $file.FullName -Raw -Encoding UTF8
            $originalContent = $content
            $fileChanged = $false
            
            foreach ($replacement in $replacements) {
                if ($content -cmatch [regex]::Escape($replacement.Old)) {
                    $content = $content -creplace [regex]::Escape($replacement.Old), $replacement.New
                    $fileChanged = $true
                }
            }
            
            if ($fileChanged) {
                Set-Content -Path $file.FullName -Value $content -Encoding UTF8 -NoNewline
                Write-Host "  Updated: $($file.Name)"
                $totalFiles++
                
                foreach ($replacement in $replacements) {
                    $count = ([regex]::Matches($originalContent, [regex]::Escape($replacement.Old))).Count
                    $totalReplacements += $count
                }
            }
        } catch {
            Write-Host "  Error processing $($file.Name): $_"
        }
    }
}

Write-Host ""
Write-Host "Rebranding complete!"
Write-Host "Files updated: $totalFiles"
Write-Host "Total replacements: $totalReplacements"
