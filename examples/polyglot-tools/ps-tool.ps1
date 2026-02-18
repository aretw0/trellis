# 0. Setup: Suppress output noise
$ProgressPreference = 'SilentlyContinue'

try {
    # 1. Input: Read from TRELLIS_ARGS (2026-02-17: Tool Argument Evolution)
    # Trellis passes all arguments as a JSON object in TRELLIS_ARGS
    $RawArgs = if ($env:TRELLIS_ARGS) { $env:TRELLIS_ARGS } else { "{}" }
    $TrellisArgs = $RawArgs | ConvertFrom-Json
    
    $Name = if ($TrellisArgs.name) { $TrellisArgs.name } else { "Guest" }
    $Greeting = if ($TrellisArgs.greeting) { $TrellisArgs.greeting } else { "Greetings" }

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
