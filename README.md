# Go File Server & JSON DB

A powerful, standalone static file server written in Go, featuring virtual hosting, automatic hosts file management, a built-in JSON database API, and system tray integration.

## Features

- **Virtual Hosting**: Serve different folders for different domains (e.g., `myserver.local`, `docs.local`).
- **Auto-Hosts Configuration**: Automatically adds configured domains to your Windows `hosts` file (requires Administrator privileges).
- **JSON Database API**: A built-in REST API to read and write JSON files, enabling backend-less frontend development.
- **System Tray Integration**: Runs silently in the background with a system tray icon and "Quit" option.
- **Auto-Open Browser**: Optionally opens your default browser to the configured sites on startup.
- **Zero-Config Deployment**: compiles to a single `start.exe` (plus `config.xml`).

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) (if building from source)
- Windows OS (for hosts file automation and system tray)

### Building the Project

To build the project effectively, you need to create two executables: one for the main server (background) and one for the stop utility.

1.  **Build the Server (Headless/Background)**
    ```powershell
    go build -ldflags "-H=windowsgui" -o start.exe main.go icon.go
    ```

2.  **Build the Stop Utility**
    ```powershell
    go build -o stop.exe stop.go
    ```

## Configuration

Server behavior is controlled via `config.xml` in the same directory.

```xml
<Config>
    <Port>9000</Port>
    <Sites>
        <Site>
            <Domain>myserver.local</Domain>
            <Path>.</Path>
            <AutoOpen>true</AutoOpen>
        </Site>
        <Site>
            <Domain>docs.local</Domain>
            <Path>./docs</Path>
            <!-- AutoOpen defaults to false if omitted -->
        </Site>
    </Sites>
</Config>
```

- **Port**: The HTTP port to listen on.
- **Sites**: List of site configurations.
    - **Domain**: The hostname to listen for.
    - **Path**: The local directory (or file) to serve for this domain.
    - **AutoOpen**: If `true`, opens this URL in the browser on startup.

## Usage

1.  **Start the Server**: Double-click `start.exe`.
    - It will appear in your **System Tray**.
    - If running as **Administrator**, it will update `C:\Windows\System32\drivers\etc\hosts` automatically.
    - Logs are written to `server.log`.

2.  **Stop the Server**:
    - Right-click the System Tray icon and select **Quit**.
    - OR run `stop.exe`.

## JSON Database API

The server exposes a simple API to read/write JSON files located in the `db/` directory.

### Endpoint: `/api/db/<filename>`

- **GET** `/api/db/users`
    - Returns the content of `db/users.json`.
    - Returns `{}` if the file does not exist.

- **POST** `/api/db/users`
    - Overwrites `db/users.json` with the request body (JSON).
    - returns `{"status": "ok"}` on success.

### Example Usage (JavaScript)

```javascript
// Read
fetch('/api/db/guestbook')
  .then(res => res.json())
  .then(data => console.log(data));

// Write
fetch('/api/db/guestbook', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ entries: [...] })
});
```
