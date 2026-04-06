# pomo — Lock In From Your Terminal

Start structured work sessions, block distractions, and stay locked in — right from your CLI. No fluff, no fancy UI stuff.

## Installation

### Using `go install`

```bash
go install github.com/JayYoon0412/pomo@latest
```

### Manual build

```bash
git clone https://github.com/JayYoon0412/pomo.git
cd pomo
go build -o pomo .
mv pomo /usr/local/bin/pomo   # or any directory on your $PATH
```

## Usage

Start a session with defaults (25 min focus, 5 min break):

```bash
pomo start
```

Custom durations:

```bash
pomo start --focus 45 --break 10
```

Block websites during focus time:

```bash
sudo pomo start --block youtube.com,twitter.com,reddit.com
```

> **Note:** Website blocking requires `sudo` because it modifies `/etc/hosts`.

### Flags

| Flag      | Default | Description                                         |
| --------- | ------- | --------------------------------------------------- |
| `--focus` | `25`    | Focus duration in minutes                           |
| `--break` | `5`     | Break duration in minutes                           |
| `--block` | —       | Comma-separated list of sites to block during focus |

## How website blocking works

pomo redirects blocked domains to `127.0.0.1` by adding entries to `/etc/hosts`. It also blocks the `www.` variant of each domain automatically, so blocking `youtube.com` covers `www.youtube.com` as well.

### Why it's safe

**Only pomo's own entries are ever modified.** Entries are wrapped in clearly marked delimiters:

```
# pomo-block-start
127.0.0.1 youtube.com
127.0.0.1 www.youtube.com
# pomo-block-end
```

Every line outside these markers is left completely untouched. Removal works by stripping only the content between the markers — it never parses, rewrites, or reformats any of your existing hosts entries.

**Blocks are always removed automatically.** Sites are unblocked when the focus phase ends (before your break begins), or immediately if you exit early with `Ctrl+C`. The signal handler runs cleanup before the process exits.

**Leftover blocks from a crash are cleaned up on the next run.** If pomo is killed before it can clean up (e.g. power loss), the markers remain in the file. The next time you run `pomo start`, it detects and removes any leftover pomo block before writing a new one — so the file never accumulates stale entries.

To manually remove pomo's entries at any time, delete the lines between and including the `# pomo-block-start` and `# pomo-block-end` markers.
