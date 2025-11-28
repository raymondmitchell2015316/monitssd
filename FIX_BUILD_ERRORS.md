# Fix Build Errors

## Issues Found

1. **notify.go:276** - Variable redeclaration error (FIXED)
2. **main.go:479** - Missing `StartAdminServer` function (admin.go not in repo)

## Solution

### Option 1: Add admin.go to Repository (Recommended)

The `admin.go` file exists locally but needs to be added to the repository:

```bash
# On your local machine
git add admin.go
git commit -m "Add admin web interface"
git push origin main
```

Then on VPS:
```bash
cd ~/monitssd
git pull
go build
```

### Option 2: Copy admin.go to VPS Manually

If you can't push to the repo right now, copy admin.go directly to your VPS:

```bash
# On VPS, create admin.go with the content
nano ~/monitssd/admin.go
```

Then paste the admin.go content (see below or copy from your local machine).

### Option 3: Build Without Admin Feature

The code has been updated to work without admin.go. You can build and run without the `--admin` flag:

```bash
go build
./evilginx_monitor
```

The `--admin` flag will show a message but won't crash.

## Fixed Issues

✅ **notify.go line 276** - Changed `sessionKey :=` to reuse existing variable (removed redeclaration)

## Next Steps

1. Add `admin.go` to your repository
2. Pull the latest code on VPS
3. Rebuild the application

