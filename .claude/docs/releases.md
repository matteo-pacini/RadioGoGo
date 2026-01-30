# Releases

## Building a Release

```bash
./make_release.sh <version>
```

Version is injected via ldflags:
```
-ldflags="-s -w -X github.com/zi0p4tch0/radiogogo/data.Version=$1"
```

## Supported Platforms

**Built:**
- macOS: amd64, arm64
- Linux: 386, amd64, arm64, armv6, armv7
- Windows: 386, amd64

**Not supported** (modernc.org/sqlite libc constraints):
- FreeBSD, OpenBSD, NetBSD
- Windows ARM

## Release Notes Format

Use emoji-based categories:

```markdown
## v0.3.0 - Feature Release Title

- New features
- Audio-related
- UI improvements
- Bug fixes
- Customization
- Platform-specific

### SHA256 Checksums
[checksums here]
```

## Guidelines

- Only list changes visible to users
- Don't list bugs introduced and fixed in same release
- Focus on user impact, not implementation ("Bookmarks" not "SQLite backend")
- Include SHA256 checksums at end
