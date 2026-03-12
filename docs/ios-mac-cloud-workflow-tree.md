# iOS + macOS Cloud Sync Workflow Tree

## Goal
Use one repo as the control plane for iCloud Drive, OneDrive, and Google Drive, with clear status strings that iOS and macOS can both read.

## Repo
- Public remote: `https://github.com/allancordoba2-crypto/IDH_brazil.git`
- Local repo path: `/Users/bateuopontococandoosaco/IDH_brazil_launchd`

## Workflow Tree (Best Practice)
1. Source Layer (Cloud Providers)
- iCloud Drive
- OneDrive
- Google Drive

2. Normalize Layer (macOS)
- LaunchAgent script scans provider folders.
- Only selected extensions are mirrored to `cloud_sources_mirror/`.

3. Validate Layer (macOS)
- Generate `reports/cloud_sources_sync_report.md` with `SYNCED`, `FAILED`, `SKIPPED`.
- Detect permission errors early.

4. Commit + Push Layer (macOS)
- Commit only real file/report deltas.
- Emit upload strings before/after push.

5. Monitor Layer (iOS + macOS)
- Read status strings from logs and GitHub latest commit.
- iOS Shortcut notifies success/failure.

## Unified Status String Protocol
Use this exact format for all providers and git events:

`<EVENT>|<ISO8601>|<SCOPE>|<DETAILS>`

Examples:
- `SYNC_START|2026-03-12T09:20:00-03:00|cloud_sources|scan started`
- `SYNC_PROVIDER|2026-03-12T09:20:03-03:00|icloud_drive|SYNCED`
- `SYNC_PROVIDER|2026-03-12T09:20:05-03:00|onedrive|FAILED:permission`
- `UPLOAD_START|2026-03-12T09:21:10-03:00|origin/main|git push started`
- `UPLOAD_DONE|2026-03-12T09:21:18-03:00|origin/main|ok`
- `UPLOAD_FAIL|2026-03-12T09:21:18-03:00|origin/main|auth_or_network`

## Folder Tree (Recommended)

```text
IDH_brazil_launchd/
  .github/workflows/
    proposals-fast.yml
    proposals-mirror-sync.yml
  cloud_sources_mirror/
    icloud_drive/
    onedrive_*/
    googledrive_*/
  reports/
    cloud_sources_sync_report.md
    proposals_summary.md
  docs/
    ios-mac-cloud-workflow-tree.md
  scripts/
    update_proposals_mirror.py
```

## iOS Shortcut (Recommended)
1. Trigger: Every 1 hour.
2. Action: `Get Contents of URL` for GitHub latest commit API:
- `https://api.github.com/repos/allancordoba2-crypto/IDH_brazil/commits/main`
3. Extract commit SHA + date.
4. Compare with last SHA saved in iCloud Shortcuts file.
5. If changed, show notification:
- `Repo updated: <short_sha> at <date>`
6. Optional: open repo URL on tap.

## macOS Runtime Signals
- Start-of-upload signal already enabled in script:
- `UPLOAD_START|<timestamp>|origin/<branch>`

Log path:
- `/Users/bateuopontococandoosaco/Library/Logs/com.bateuopontococandoosaco.cloud-sources-github-sync.out.log`

## Operational Notes
- If provider paths show `Operation not permitted`, grant Full Disk Access for the runner context or run sync through a user-approved foreground automation.
- Keep remote URL on HTTPS for public repo workflows.
