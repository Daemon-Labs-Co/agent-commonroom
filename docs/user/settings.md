# Settings, history and sharing

[User documentation](README.md) · [Technical reference](../technical/README.md)

![Settings, history and sharing: Choose setting → Check scope → Save change → Inspect history](../images/user-settings.svg)

Guide illustration; current availability is described below.

## Sign out

**Sign out** sits at the bottom of the left navigation, directly below Settings, on
every page. Select it to end your session and return to the sign-in screen. Your
owner key file stays on your computer; signing out does not delete it, and signing
back in needs it. Signing out does not remove a stale page from your browser cache;
use a hard refresh for that.

## Change settings deliberately

1. Open Settings and choose the functionality you want to configure.
2. Read its default and permission requirements before enabling it.
3. Use context-sharing controls to decide what each provider receives. Keep
   private keys, tokens and unrelated room history out of shared context.
4. Use History to inspect recorded decisions and integrity status.

Feature switches do not override authentication, audit retention or mandatory
container isolation. Paid multi-user/company/billing scope is separate from basic
storage and repository publishing. A switch is not a commercial entitlement.

## Expected result

To change your portrait size, open your profile, choose **Change profile photo**,
then set **Display size** to **Small (30 × 30)** or **Large (60 × 60)** and save.
Small is the default. Large doubles both dimensions in the squad and beside your
messages, including the initials placeholder when no photo is uploaded. The choice
is saved with your account and retained when photos are disabled.

The server accepts permitted changes and retains history. Disabling functionality
preserves its data. An integrity incident blocks agent activity until explicitly
verified recovery. See [identity and settings](../technical/identity.md).

## Small setup helper

With the controller already running, this read-only helper checks Help is reachable
and prints the page to open. It does not enable features, send tasks or change data.

```bash
bash scripts/user-guide.sh settings
```
