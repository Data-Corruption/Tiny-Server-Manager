$DIST_DIR = "dist"
$SRC_DIR = "src"

# Print usage if no arguments are provided
if ($args.Length -eq 0) {
  Write-Host "Usage: build.ps1 [-all | -win | -linux]"
  exit 1
}

# Remove the dist directory if it exists, then create it
if (Test-Path -Path $DIST_DIR -PathType Container) {
  Remove-Item -Path $DIST_DIR -Recurse -Force
} elseif (Test-Path -Path $DIST_DIR) {
  Write-Host "Error: The path '$DIST_DIR' exists but is not a directory."
  exit 1
}
New-Item -Path $DIST_DIR -ItemType Directory

# Function that takes in an OS and architecture then builds an executable
function build($os, $arch) {
  $output = Join-Path $DIST_DIR "bin-$os-$arch"
  if ($os -eq "windows") {
    $output += ".exe"
  }
  # Set the GOOS and GOARCH environment variables
  $env:GOOS = $os
  $env:GOARCH = $arch
  # Build the executable
  go build -o $output .\src\main.go
  # Check if the build succeeded
  if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed for '$output'."
  } else {
    Write-Host "Successfully built '$output'."
  }
}

# Check the command line argument and build the appropriate executables
if ($args[0] -eq "-all") {
  build "windows" "amd64"
  build "windows" "arm64"
  build "linux" "amd64"
  build "linux" "arm64"
  build "linux" "riscv64"
}
if ($args[0] -eq "-win") {
  build "windows" "amd64"
  build "windows" "arm64"
}
if ($args[0] -eq "-linux") {
  build "linux" "amd64"
  build "linux" "arm64"
  build "linux" "riscv64"
}