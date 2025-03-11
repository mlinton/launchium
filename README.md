# Launchium - A Chromium Profile Manager

A terminal-based application for managing multiple Chromium/Chrome browser profiles with different settings, proxies, and configurations. This tool helps you maintain separate browsing environments for different purposes.

Primary purpose and inspiration was to regain similar functionality to the NCC Group's autochrome tool (https://github.com/nccgroup/autochrome).

![Launchium Screenshot](https://github.com/user-attachments/assets/2f5adc64-cf61-463c-ad67-819b2b828493)


## Features

- **Multiple Profile Support**: Create and manage separate browser profiles
- **Proxy Configurations**: Set different proxy settings for each profile
- **Custom Flags**: Configure browser launch flags for each profile
- **Profile Cleaning**: Reset profiles to a clean state
- **GPU Artifact Suppression**: Eliminate visual glitches with optimized settings
- **API Warning Suppression**: Remove annoying Google API key warnings
- **Terminal UI**: Beautiful, keyboard-driven interface

## To-Do
- **Chromium Management**: Download/manage and use different versions of chromium
- **Extension Management**: Download/manage extensions for use in different profiles

## Installation

### Prerequisites

- Go 1.16 or later
- Chromium or Google Chrome browser

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/mlinton/launchium.git
cd launchium
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o launchium
```

4. (Optional) Install to your system:
```bash
sudo mv launchium /usr/local/bin/
```

## Usage

### Starting the Application

```bash
./launchium
```

### Navigation

- Use arrow keys to navigate menus
- Press Enter to select an option
- Press Esc to go back
- Press Ctrl+C to quit

### Managing Profiles

1. **Launch Browser**: Start Chromium/Chrome with a selected profile
2. **Manage Profiles**:
   - Add New Profile: Create a new browser profile
   - Edit Profile: Modify settings for an existing profile
   - Delete Profile: Remove a profile
3. **Clean Profile**: Reset a profile to a clean state
4. **Quit**: Exit the application

### Profile Settings

Each profile has the following settings:

- **Name**: Unique identifier for the profile
- **Proxy**: Server address and port (or "none" for direct connection)
- **Proxy Type**: Connection type (http, socks5, or none)
- **Flags**: Custom command-line flags for Chromium/Chrome

### Default Profiles

Two profiles are created by default:
- **default**: Standard profile with minimal settings
- **clean**: Profile with enhanced GPU artifact suppression

## Configuration

Profiles are stored in `~/.chrome_profiles/`:
- Profile definitions: `~/.chrome_profiles/profiles.conf`
- Profile data: `~/.chrome_profiles/<profile-name>/`

## Advanced Usage

### Custom Proxy Configuration

To set up a proxy, edit a profile and specify:
1. Proxy address (e.g., "127.0.0.1:8080" or "none")
2. Proxy type ("http", "socks5", or "none")

Example for Tor proxy:
- Proxy: "127.0.0.1:9050"
- Type: "socks5"

### Custom Browser Flags

Common useful flags:
- `--incognito`: Start in incognito mode
- `--disable-features=...`: Disable specific Chrome features
- `--enable-features=...`: Enable specific Chrome features
- `--disable-extensions`: Run without extensions

## Troubleshooting

### Browser Won't Launch

If the browser doesn't launch:
1. Verify the browser path is correct
2. Check if your profile directory exists and has proper permissions
3. Try cleaning the profile and launching again

### Visual Artifacts

If you see graphics glitches:
1. Edit the profile
2. Add the following flags: `--disable-gpu --disable-gpu-compositing`
3. Save and launch again

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for terminal UI
- Inspired by the need for better browser profile management
