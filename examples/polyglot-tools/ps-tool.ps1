# 0. Setup: Suppress output noise
$ProgressPreference = 'SilentlyContinue'

try {
    # 1. Input: Read from Environment Variables
    # PowerShell environment variables are case-insensitive
    $Name = if ($env:TRELLIS_ARG_NAME) { $env:TRELLIS_ARG_NAME } else { "Guest" }
    $Greeting = if ($env:TRELLIS_ARG_GREETING) { $env:TRELLIS_ARG_GREETING } else { "Greetings" }

    # 2. Logic
    $Message = "$Greeting, $Name! [PowerShell]"

    # 3. Output: JSON to Stdout
    $Output = @{
        message = $Message
        runtime = "PowerShell $($PSVersionTable.PSVersion.ToString())"
        status  = "success"
    }

    $Output | ConvertTo-Json -Compress
}
catch {
    # Error: Write to Error Stream
    Write-Error "Error in pwsh script: $_"
    exit 1
}
