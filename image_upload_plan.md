# Personal File Upload Service - Implementation Guide

## Project Overview
A personal, lightweight file upload service written in Go. The service will receive general file uploads (images, text, documents) via a POST request, save them to local NVMe storage, and return a URL. The client script will prioritize `0x0.st` to save server bandwidth, falling back to this personal service if `0x0.st` fails. 

The client script will also include logic to strip EXIF data from images, troll metadata sniffers by injecting coordinates for Pyongyang, North Korea, and optionally preserve original filenames.

## Target Architecture

1.  **Backend Service:** Go (Standard Library `net/http`)
    *   **Why Stdlib?** Go's standard HTTP library is incredibly robust, fast, and requires zero external dependencies. For a single-endpoint upload service, frameworks like Gin or Echo add unnecessary bloat.
2.  **Reverse Proxy:** Caddy or Nginx (Recommended: Caddy)
    *   The Go service will run locally on the server (e.g., `http://localhost:8080`). A reverse proxy will listen on ports 80/443, automatically provision SSL/TLS certificates for `u.fftp.io`, and route traffic to the Go service. It can also be configured to serve the uploaded files statically for maximum performance to standard `GET` requests.
3.  **Storage:** Local server storage (e.g., `/var/www/uploads/`)
    *   *Note on Capacity:* A 25GB NVMe drive will typically have ~15-20GB of usable space after the OS. This provides plenty of space, especially since large files should be ideally uploaded to 0x0.st first anyway.

## Phase 1: Go Backend Implementation

**Requirements for the Go Service:**
*   **Endpoint:** `POST /upload`
*   **Authentication:** 
    *   Use a static "API Token" passed via the `Authorization: Bearer <token>` header.
*   **File Naming:**
    *   By default, files should be saved using the current date and time (e.g., `2026-03-15_00-04-13.jpg` or `2026-03-15_00-04-13.txt`).
    *   The service must accept an optional `preserve_filename=true` form field or URL parameter. If true, the file is saved using its original uploaded name.
    *   The service must extract the original file extension from the uploaded file and append it to the timestamp. If no extension exists, save without one or default to `.bin`.
    *   If a collision occurs on saving, append a suffix before the extension (e.g. `YYYY-MM-DD_HH-MM-SS-1.ext`).
*   **Response:** Return the final URL (e.g., `https://u.fftp.io/2026-03-15_00-04-13.txt`).

**Task List for Claude:**
1.  Initialize a new Go module (`go mod init u`).
2.  Write `main.go` setting up an HTTP server on a configurable port.
3.  Implement a middleware or check for the Authorization header comparing against an environment variable (e.g., `UPLOAD_API_KEY`).
4.  Implement the upload handler:
    *   Parse the multipart form upload.
    *   Check for the `preserve_filename` parameter.
    *   Determine the destination filename (either the original name or `time.Now().Format("2006-01-02_15-04-05")` + `extension`).
    *   Check if the file exists; if so, enter a loop appending `-1`, `-2` until a unique name is found.
    *   Save the file to a configured base directory.
    *   Return the constructed URL.
5.  Provide a systemd service file (`u.service` or `file-upload.service`) to keep the Go app running in the background.

## Phase 2: Server Configuration

**Task List for Claude:**
1.  Provide instructions to install and configure Caddy (or Nginx).
2.  Write the Caddyfile (or Nginx config) for `u.fftp.io`:
    *   Route `POST /upload` requests to the Go service (e.g., `localhost:8080`).
    *   Route `GET /*` requests to serve files directly from the storage directory.
    *   Ensure appropriate MIME types are returned for various file types.
3.  Provide steps to set up the directory and permissions on the Vultr server.

## Phase 3: Client Script Modifications

**Target Files:**
*   `/home/ff/.local/bin/upload-image` (Consider renaming to `upload-file` going forward)
*   `/home/ff/.local/bin/upload-selection`

**Requirements:**
*   **Argument Parsing:** Implement flags in the scripts (e.g., using `argparse` in Python).
    *   `--preserve-filename`: Passed to the Go service to prevent date/time renaming.
    *   `--keep-exif`: Bypasses EXIF stripping/trolling logic.
*   **EXIF Stripping & Trolling (Default Behavior):**
    *   Before uploading an image (JPEG, PNG, etc.), the scripts must strip existing EXIF data.
    *   Inject fake GPS coordinates for **Pyongyang, North Korea**. (e.g., Latitude: 39° 1' 27.48" N, Longitude: 125° 45' 12.6" E).
    *   *Implementation Note for Claude:* Recommend using a standard Python library like `Pillow` or heavily rely on triggering `exiftool` via subprocess if it is installed on the user's system to handle the metadata manipulation reliably before standard upload procedures.
*   **Upload Logic:**
    *   The scripts should first attempt to `curl` to `https://0x0.st`.
    *   If `0x0.st` fails, times out, or returns a non-200 status, **fallback** to uploading to `https://u.fftp.io/upload`.
    *   Pass the authentication token in the fallback request (`-H "Authorization: Bearer YOUR_TOKEN"`).
    *   Pass the preservation flag (`-F "preserve_filename=true"`) when requested.

**Task List for Claude:**
1.  Write a Python function to modify image EXIF data using either an external tool (`exiftool` via `subprocess`) or `Pillow`/`piexif`. The function should remove all metadata and insert Pyongyang GPS coordinates.
2.  Add argument parsing to the main script execution.
3.  Write a new Python function `upload_to_personal_service(filepath_or_content, preserve_filename=False)` that executes the `curl` command to `u.fftp.io` using the authentication token and optional flags.
4.  Update the existing upload logic to catch failures from 0x0.st and immediately trigger the new personal service function.
5.  Ensure notifications (`notify-send`) accurately reflect whether the item went to 0x0.st or the personal fallback server.
