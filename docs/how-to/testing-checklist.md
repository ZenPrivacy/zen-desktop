# Manual Checklist

## Home

- Start proxy
- Start proxy without CA installed
- Stop proxy

## Filter lists

- Filter lists counters are correct (enabled/total)
- Add custom filter list
- Add custom filter list with invalid URL (error)
- Add custom filter list without name (name becomes same as URL)
- Add trusted custom filter list
- Add filter list with duplicate URL (error)
- Delete custom filter list
- Import filter list
- Import filter list in incorrect format
- Export filter lists
- Enable filter list
- Disable filter list

## My rules

- Add My rule (e.g. `0.0.0.0 www.google.com`)
- Remove My rule
- "Help" button opens guide on How to write rules

## Preferences

- Change language
- Enable "Autostart" (Windows/macOS/Linux)
- Disable "Autostart" (Windows/macOS/Linux)
- Export logs
- Set fixed Port
- Add ignored hosts
- Remove ignored hosts
- Uninstall CA
- Uninstall CA while proxy is running

## Self‑update (Windows/macOS/Linux)

- Disable updates
- Select "Ask before updating"
- Select "Automatic updates"

## Windows installation

- On clean Windows 10 profiles A and B, install as A
- A's Start menu shortcut is under A's roaming `Microsoft\Windows\Start Menu\Programs` folder, targets A's installed executable, and launches Zen
- No Zen Start menu shortcut is created in ProgramData or in B's personal Start menu
- Desktop shortcut and uninstall registration match the previous installer
- Install for a user profile whose path contains spaces, then reinstall as the same user under the installer's normal elevation flow
- Shortcut and executable refer to the same profile
- Record alternate-credential elevation separately if the shortcut and executable refer to different profiles

## Windows uninstallation

- Install independently as B, then uninstall A's installation
- A's personal shortcut is removed; B's shortcut and installed executable remain
- Uninstall with A's Start menu shortcut already absent completes normally

## Windows upgrade

- Upgrade from a version that created a shared Start menu shortcut
- The new personal shortcut launches Zen
- Record the leftover shared shortcut; do not delete another user's entry
