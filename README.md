# Portmap

A small CLI for seeing what's running on your ports.

```
portmap ls
```

```
PORT   PID    PROCESS  CWD
3000   84312  node     /home/user/projects/myapp
8080   5678   python   /home/user/projects/api
5432   1234   postgres
```

---

## Install

**Homebrew (macOS/Linux):**

```bash
brew tap bhpcv252/portmap
brew install portmap
```

**Go:**

```bash
go install github.com/bhpcv252/portmap@latest
```

**From source:**

```bash
git clone https://github.com/bhpcv252/portmap
cd portmap
go build -o portmap .
```

---

## Commands

### `ls`

List all active listening ports with their PID, process name, and working directory.

```bash
portmap ls
portmap ls --json
```

### `check <port>`

Show what's on a specific port. Exits 0 if nothing is running, 1 if something is.

```bash
portmap check 3000
```

```
port 3000
  status:   ● running
  pid:      84312
  process:  node
  cwd:      /home/user/projects/myapp
```

Useful in scripts:

```bash
portmap check 3000 || npm start
```

### `kill <port>`

Kill whatever is running on a port. Asks for confirmation unless you pass `--yes`.

```bash
portmap kill 3000
portmap kill 3000 --yes
```

### `suggest`

Find a free port in a range. Useful when you just need something available.

```bash
portmap suggest
# suggested port: 3000

portmap suggest --from 8000 --to 9000
# suggested port: 8001

portmap suggest --count 3
# suggested ports: 3000, 3001, 3002
```

### `version`

```bash
portmap version
```

---

## Platform support

| Platform | Port detection | CWD               |
| -------- | -------------- | ----------------- |
| macOS    | lsof           | lsof              |
| Linux    | /proc/net/tcp  | /proc/\<pid\>/cwd |
| Windows  | netstat        | not supported     |

On macOS, system processes protected by SIP won't show a working directory. That's a macOS restriction, not a bug.
