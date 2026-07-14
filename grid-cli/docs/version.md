# Version

Display version information about the tfcmd binary.

## Version Command

```bash
tfcmd version
```

Shows the current version, build information, and git commit details.

## Example Output

```console
$ tfcmd version
tfcmd version 1.2.3
Build date: 2023-12-17T10:30:00Z
Git commit: a1b2c3d4e5f6789abcdef1234567890abcdef12
Go version: go1.21.1
```

## Output Fields

- **Version**: Semantic version number (e.g., 1.2.3)
- **Build date**: When the binary was built
- **Git commit**: Full git commit hash of the source
- **Go version**: Go language version used to build

## Use Cases

### Checking Installation
```bash
# Verify tfcmd is installed and working
tfcmd version
```

### Bug Reporting
Include version information when reporting issues:
```bash
$ tfcmd version
# Include this output in bug reports
```

### Version Compatibility
Check if you have the latest version:
```bash
# Compare your version with latest releases
tfcmd version
# Check: https://github.com/threefoldtech/grid_agent/releases
```

### Troubleshooting
Version information helps diagnose:
- Build-specific issues
- Feature availability
- Compatibility problems

## Version Format

The version follows [semantic versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

Example: `1.2.3`
- `1`: Major version
- `2`: Minor version  
- `3`: Patch version

## Update Process

To update to the latest version:

1. **Check current version**:
   ```bash
   tfcmd version
   ```

2. **Download latest release**:
   ```bash
   # Visit: https://github.com/threefoldtech/grid_agent/releases
   # Download the latest binary for your platform
   ```

3. **Replace binary**:
   ```bash
   # Backup old version
   mv /usr/local/bin/tfcmd /usr/local/bin/tfcmd.backup
   
   # Install new version
   mv tfcmd /usr/local/bin/tfcmd
   chmod +x /usr/local/bin/tfcmd
   ```

4. **Verify update**:
   ```bash
   tfcmd version
   ```

## Build from Source

If you build from source, version information is automatically included:

```bash
# Clone repository
git clone https://github.com/threefoldtech/grid_agent.git
cd grid_agent/grid-cli

# Build
make build

# Check version
./tfcmd version
```

## Development Versions

Development builds may include:
- Pre-release tags (e.g., `1.2.3-alpha.1`)
- Commit hash instead of version
- Build metadata

Example development output:
```console
$ tfcmd version
tfcmd version 1.2.3-alpha.1+dev.a1b2c3d
Build date: 2023-12-17T10:30:00Z
Git commit: a1b2c3d4e5f6789abcdef1234567890abcdef12
Go version: go1.21.1
```

## Troubleshooting

### Version Not Showing
```bash
$ tfcmd version
Error: version information not available
```
**Solutions:**
- Binary may be corrupted - reinstall
- Build may not include version info - rebuild with proper flags

### Old Version
If you have an old version and need features:
1. Check current version: `tfcmd version`
2. Download latest release
3. Install as shown in update process

### Version Conflicts
Multiple tfcmd installations can cause conflicts:
```bash
# Check which tfcmd is being used
which tfcmd
# Expected: /usr/local/bin/tfcmd

# Check all versions
/usrusr/local/bin/tfcmd version
~/bin/tfcmd version
```

## Integration with Scripts

Use version information in scripts for compatibility checks:

```bash
#!/bin/bash
# Check tfcmd version compatibility

REQUIRED_VERSION="1.2.0"
CURRENT_VERSION=$(tfcmd version | head -1 | cut -d' ' -f3)

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$CURRENT_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "Error: tfcmd version $CURRENT_VERSION is too old. Required: $REQUIRED_VERSION or newer"
    exit 1
fi

echo "tfcmd version $CURRENT_VERSION is compatible"
```

## Release Notes

Version information corresponds to official releases:
- **Main releases**: Stable features and bug fixes
- **Pre-releases**: Alpha/beta testing versions
- **Development builds**: Latest features (may be unstable)

Check the [GitHub releases](https://github.com/threefoldtech/grid_agent/releases) for:
- Release notes
- Breaking changes
- New features
- Bug fixes

## Getting Help

If version-related issues occur:
1. Verify you have the latest version
2. Check release notes for breaking changes
3. Report issues with version information included
4. Consider building from source for latest features

For the most up-to-date information, always check the official ThreeFold documentation and GitHub releases.
