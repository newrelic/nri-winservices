<#
    .SYNOPSIS
        This script verifies, tests, builds and packages a New Relic Infrastructure Integration
#>
param (
    # Target architecture: amd64 (default) or 386
    [ValidateSet("amd64", "386")]
    [string]$arch="amd64",
    [string]$version="0.0.0",
    # Skip tests
    [switch]$skipTests=$false,
    [switch]$skipExporterCompile=$false
)
# Print errors in NormalView
$ErrorView = 'NormalView'

$integration = $(Split-Path -Leaf $PSScriptRoot)
$integrationName = $integration.Replace("nri-", "")
$executable = "nri-$integrationName.exe"
$commitHash = (git rev-parse HEAD)

$exporterRepo = "github.com\prometheus-community\windows_exporter"
$exporterBinaryName = "windows_exporter.exe"
# Commit used by v0.12.0 of windows_exporter
$exporterVersion = "eff5f2415398d89173a2203c47366e4882f3bc0e"
# Collector used by the Windows Service integration
$collectors = "collector.go","wmi.go","perflib.go","service.go","cs.go"

$env:GOPATH = go env GOPATH
$env:GOBIN = "$env:GOPATH\bin"
$env:GOOS = "windows"
$env:GOARCH = $arch
$env:GO111MODULE = "auto"

# verifying version number format
$v = $version.Split(".")

if ($v.Length -ne 3) {
    echo "-version must follow a numeric major.minor.patch semantic versioning schema (received: $version)"
    exit -1
}

$wrong = $v | ? { (-Not [System.Int32]::TryParse($_, [ref]0)) -or ( $_.Length -eq 0) -or ([int]$_ -lt 0)} | % { 1 }
if ($wrong.Length  -ne 0) {
    echo "-version major, minor and patch must be valid positive integers (received: $version)"
    exit -1
}

echo "--- Checking dependencies"
# We are running a job in a windows that calls a .ps1 experiencing this issue. 
# Basically when using git (and as well when running go get or go mod...) the powershell script fails misinterpreting the output
# https://stackoverflow.com/questions/57279007/error-when-pulling-from-powershell-script
$ErrorActionPreference = "SilentlyContinue"
go mod download
$ErrorActionPreference = "Stop"
go mod download

echo "--- Collecting files"
$goFiles = go list ./...

echo "--- Format check"

$wrongFormat = go fmt $goFiles

if ($wrongFormat -and ($wrongFormat.Length -gt 0))
{
    echo "ERROR: Wrong format for files:"
    echo $wrongFormat
    exit -1
}

if (-Not $skipTests) {
    echo "--- Running tests"

    go test $goFiles
    if (-not $?)
    {
        echo "Failed running tests"
        exit -1
    }    
}

echo "--- Running Build"

go build -v $goFiles
if (-not $?)
{
    echo "Failed building files"
    exit -1
}

echo "--- Collecting Go main files"

$packages = go list -f "{{.ImportPath}} {{.Name}}" ./...  | ConvertFrom-String -PropertyNames Path, Name
$mainPackage = $packages | ? { $_.Name -eq 'main' } | % { $_.Path }

echo "generating $integrationName"
go generate $mainPackage

$fileName = ([io.fileinfo]$mainPackage).BaseName

echo "creating $executable"
go build -ldflags "-X main.integrationVersion=$version -X main.commitHash=$commitHash" -o ".\target\bin\windows_$arch\$executable" $mainPackage

if (-Not $skipExporterCompile) 
{
    echo "--- Compiling exporter"
    Push-Location $env:GOPATH
    
    $ErrorActionPreference = "SilentlyContinue"
    # exporter is build using the Prometheus tool
    go get "github.com/prometheus/promu"
    go get -d "$exporterRepo"
    $ErrorActionPreference = "Stop"

    Set-Location "$env:GOPATH\src\$exporterRepo"

    $ErrorActionPreference = "SilentlyContinue"
    git checkout "$exporterVersion"
    $ErrorActionPreference = "Stop"
    $currentCommit = cat .git/HEAD
    if($currentCommit -ne $exporterVersion){
        echo "Failed checking out exporter version $exporterVersion"
        exit -1
    }

    # remove unused collectors 
    Remove-Item .\collector\* -Exclude $collectors
    $ErrorActionPreference = "SilentlyContinue"
    go mod download
    $ErrorActionPreference = "Stop"
    promu build --prefix=output\$arch

    Pop-Location
    Copy-Item "$env:GOPATH\src\$exporterRepo\output\$arch\$exporterBinaryName" -Destination ".\target\bin\windows_$arch\" -Force 
    if (-not $?)
    {
        echo "Failed compiling exporter"
        exit -1
    }

    if (-Not $skipTests) {
        echo "--- Running integrations tests"
        go test -v -tags=integration ./test/integration_test.go
        if (-not $?)
        {
            echo "Failed running integrations tests"
            exit -1
        }
    }

}