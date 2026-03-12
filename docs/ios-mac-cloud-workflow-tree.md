# iOS + macOS Cloud Sync Workflow Tree

## Goal
Use one repo as the control plane for iCloud Drive, OneDrive, and Google Drive, with shared status strings across devices.

## Repo
- Public remote: `https://github.com/allancordoba2-crypto/IDH_brazil.git`
- Local repo path: `<LOCAL_REPO_PATH>`

## Workflow Tree
1. Source layer
- iCloud Drive
- OneDrive
- Google Drive

2. Normalize layer (macOS)
- LaunchAgent scans provider folders.
- Allowed extensions mirror into `cloud_sources_mirror/`.

3. Validate layer
- Generate `reports/cloud_sources_sync_report.md` with `SYNCED`, `FAILED`, `SKIPPED`.

4. Commit + Push layer
- Commit only real deltas.
- Emit upload status strings before and after push.

5. Monitor layer (iOS + macOS)
- iOS Shortcut polls latest commit from GitHub API.
- macOS monitors local launch log.

## Unified Status String Protocol
`<EVENT>|<ISO8601>|<SCOPE>|<DETAILS>`

Examples:
- `SYNC_START|2026-03-12T09:20:00-03:00|cloud_sources|scan started`
- `SYNC_PROVIDER|2026-03-12T09:20:05-03:00|onedrive|FAILED:permission`
- `UPLOAD_START|2026-03-12T09:21:10-03:00|origin/main|git push started`
- `UPLOAD_DONE|2026-03-12T09:21:18-03:00|origin/main|ok`
- `UPLOAD_FAIL|2026-03-12T09:21:18-03:00|origin/main|auth_or_network`

## iOS Shortcut (Recommended)
1. Trigger hourly.
2. Request: `https://api.github.com/repos/allancordoba2-crypto/IDH_brazil/commits/main`
3. Compare SHA with last stored value.
4. Notify on change.

## macOS Runtime Log
- Log path: `<LOCAL_LAUNCH_LOG_PATH>`
- Watch for: `UPLOAD_START|...`, `UPLOAD_DONE|...`, `UPLOAD_FAIL|...`
