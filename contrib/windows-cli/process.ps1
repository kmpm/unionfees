# SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
#
# SPDX-License-Identifier: MIT

# Check for unionfees-cli.exe
$exepath = "$PSScriptRoot\unionfees-cli.exe"

if ($null -eq (Get-Command "$exepath" -ErrorAction SilentlyContinue)) 
{
    Write-Host "Kan inte hitta  '$exepath'"
    Exit 1
}

$count = (Get-ChildItem -Filter *.pdf | Measure-Object).Count

if ($count -gt 1)
{
    Write-Host 'Mer än 1 pdf-dokument i mappen!' -fore red
    Exit 1
}

if ($count -eq 0)
{
    Write-Host 'Inga pdf-dokument hittades. Inget att göra'
    Exit 0
}

while(1){
	Try{
		$d = [datetime](read-host 'Ange utbetalningsdatum ÅÅÅÅ-MM-DD')
		break
    }
	Catch{
		Write-Host 'Inte giltigt datum' -fore red
    }
}

$ds = $d.ToString('yyMMdd')

Write-Host "d: '$ds' $PSScriptRoot"

New-Item -ItemType Directory -Force -Path ".\arkiv"

Get-ChildItem -Filter *.pdf | 
Foreach-Object {
    Write-Host "Bearbetar $($_.FullName)"
    & $exepath -d $ds $_.FullName
    if ($LASTEXITCODE -eq 0){
        Write-Host "Flyttar $($_) till arkiv."
        Move-Item $_ -Destination ".\arkiv"
    }
}

Write-Host "Klar"