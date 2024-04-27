# LightSpeed: Multithreaded Download Manager

LightSpeed is a powerful multithreaded download manager written in Go. It enables users to download files from the internet concurrently, utilizing multiple threads to accelerate the download process. With LightSpeed, you can download large files swiftly and efficiently, making it ideal for various downloading scenarios.

## Features

- **Multithreaded Downloads**: LightSpeed utilizes multiple threads to download files concurrently, speeding up the download process.
- **Resumable Downloads**: If a download is interrupted, LightSpeed supports resuming from where it left off, saving time and bandwidth.
- **Customizable Thread Count**: Users can specify the number of threads to use for downloading, allowing for flexibility based on network conditions and system resources.
- **Progress Tracking**: LightSpeed provides real-time progress updates during downloads, including download speed, completion percentage, and estimated time remaining.

## Usage

LightSpeed can be easily integrated into your Go applications to enable multithreaded downloading capabilities.

### Example

```go
package main

import (
    "fmt"
    "github.com/yourusername/lightspeed"
)

func main() {
    // Initialize LightSpeed with URL and output file path
    downloader := lightspeed.NewDownloader("https://example.com/largefile.zip", "largefile.zip")

    // Set the number of threads (optional)
    downloader.SetThreads(16)

    // Start the download
    err := downloader.Start()
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Download completed successfully!")
}
