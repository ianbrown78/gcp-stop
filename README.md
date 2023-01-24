# GCP-STOP

## Background
Inspired by gcp-nuke (https://github.com.au/arehmandev/gcp-nuke)  
This tool was created to stop/shutdown resources in a GCP project instead of nuking them.

### Why??
Many reasons:  
1. Shutting down resources to save money is the main driver.
2. Terraform destroy doesn't always fit the bill when it comes to saving $$$.

## Usage
```
NAME:
    gcp-stop - The GCP resource shutdown tool
    
USAGE:
    gcp-stop --project test-stop-123456 --dryrun
    
VERSION:
    v0.0.1
    
COMMANDS:
    help, h     Shows a list of commands or help for one command
    
GLOBAL OPTIONS:
    --project <string>      GCP project id (required)
    --dryrun                Perform a dryrun instead
    --timeout <integer>     Timeout for shutdown of a single resource in seconds. (default: 400)
    --polltime <integer>    Time for polling resource shutdown status in seconds. (default: 10)
    --help, -h              Show help
    --version, -v           Print the version of gcp-stop
```

## Roadmap
- Unit tests would be helpful
- Add Cloud SQL