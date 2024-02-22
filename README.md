# WatchZap

WatchZap is a tool designed to monitor a folder for changes in files containing messages and send those messages via WhatsApp using the WhatsApp API. Alternatively
you can also serve an HTTP server which accepts requests and send them to their recipients.

## Features

- Monitor a folder for changes in files.
- Parse messages from JSON or YAML files.
- Set up an HTTP server for receiving message requests.

## Install

Download the binaries from the release section (Windows, Linux) only

## Build

1. Clone the repository:

   ```bash
   git clone https://github.com/your_username/watchzap.git
   ```

2. Navigate to the project directory:

   ```bash
   cd watchzap
   ```

3. Build the project:

   ```bash
   go build
   ```

4. Run the executable:

   ```bash
   ./watchzap
   ```

## Usage

WatchZap provides the following functionalities:

- **Watch Folder**: Monitor a folder for changes in files containing messages.
- **Enable HTTP Server**: Set up an HTTP server for receiving message requests.
- **Both**: Perform both actions concurrently.
- **Logout**: Log out from WhatsApp and clear the local database.

The message request formats are the following:

JSON

```JSON
    [
        {
            "recipient": "Recipient 1",
            "content": "Content for Recipient 1",
            "attachment": "base64string" // Optional field
        },
        {
            "recipient": "Recipient 2",
            "content": "Content for Recipient 2",
            "attachment": "base64string" // Optional field
        }
    ]
```

YAML

```YAML
---
- recipient: Recipient1
  content: Content for Recipient1
  attachment: |
  base64string
- recipient: Recipient2
  content: Content for Recipient2
  attachment: |
  base64string
```

Currently only JSON and YAML is supported
## Configuration

WatchZap can be configured using command-line flags:

- `-debug`: Enable debug mode for WhatsApp API.

Example:

```bash
./watchzap --debug
```

## Dependencies

WatchZap relies on the following third-party libraries:

- `github.com/radovskyb/watcher`: For monitoring file changes.
- `github.com/rs/zerolog`: For logging.
- `github.com/mattn/go-sqlite3`: Sqlite driver for database interaction.
- `go.mau.fi/whatsmeow`: WhatsApp Library for API integration.
