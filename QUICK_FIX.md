# Quick Fix for Build Errors

## Problem
When cloning the repository, `admin.go` is missing, causing build errors.

## Solution

### Step 1: Fix notify.go (Already Fixed)
The variable redeclaration error in `notify.go` has been fixed.

### Step 2: Add admin.go to Your VPS

You have two options:

#### Option A: Copy admin.go from Local Machine to VPS

On your **local machine**, copy admin.go to VPS:

```bash
# From your local machine
scp admin.go root@your-vps-ip:/root/monitssd/
```

Then on VPS:
```bash
cd ~/monitssd
go build
```

#### Option B: Add admin.go to Git Repository (Recommended)

On your **local machine**:

```bash
cd C:\Users\USER\OneDrive\Documents\Dev\Pyhton\monitor
git add admin.go
git commit -m "Add admin web interface"
git push origin main
```

Then on **VPS**:
```bash
cd ~/monitssd
git pull
go build
```

### Step 3: Build

```bash
cd ~/monitssd
go build
./evilginx_monitor
```

## If You Don't Need Admin UI

If you don't need the admin web interface, you can build without it:

```bash
# The code will work, but --admin flag won't work
go build
./evilginx_monitor  # Run without --admin flag
```

## Summary

✅ **Fixed**: `notify.go` line 276 - variable redeclaration  
⚠️ **Missing**: `admin.go` file needs to be added to repository or copied to VPS

