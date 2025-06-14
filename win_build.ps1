<#
    .SYNOPSIS
        This script verifies, tests, builds and packages a New Relic Infrastructure Integration
#>
param (
    # Target architecture: amd64 (default) or 386
    [ValidateSet("amd64", "386")]
    [string]$arch="amd64",
    [string]$version="v0.0.0",
    # Skip tests
    [switch]$skipTests=$false,
    [switch]$skipExporterCompile=$false
)
# Print errors in NormalView
$ErrorView = 'NormalView'

$integrationName = "winservices"
$executable = "nri-winservices.exe"
$commitHash = (git rev-parse HEAD)

$exporterRepo = "github.com/prometheus-community/windows_exporter"
$exporterBinaryName = "windows_exporter.exe"
# Commit used by v0.30.6 of windows_exporter
$exporterVersion = "db60c78f32185083354f16fb9d534a021f0d85f9"


$env:GOPATH = go env GOPATH
$env:GOBIN = "$env:GOPATH\bin"
$env:GOOS = "windows"
$env:GOARCH = $arch
$env:GO111MODULE = "auto"

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
    git clone "https://$exporterRepo.git" "$env:GOPATH\src\$exporterRepo"
    $env:GO111MODULE = "on"
    go install "github.com/prometheus/promu@latest"
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

    $ErrorActionPreference = "SilentlyContinue"
    go mod download
    $ErrorActionPreference = "Stop"
    promu build --prefix=output\$arch

    Pop-Location
    Copy-Item "$env:GOPATH\src\$exporterRepo\output\$arch\$exporterBinaryName" -Destination ".\target\bin\windows_$arch\" -Force 
    Copy-Item "$env:GOPATH\src\$exporterRepo\LICENSE" -Destination ".\target\bin\windows_$arch\windows_exporter-LICENSE" -Force 
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