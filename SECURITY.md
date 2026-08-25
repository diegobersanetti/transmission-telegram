# Security Policy

## Supported versions

Security fixes are applied to the latest release and the default branch. Older releases may not receive patches; users should upgrade to the newest release after reviewing its notes and checksums.

## Reporting a vulnerability

Do not disclose suspected vulnerabilities, credentials, tokens, magnet links, or private Transmission details in a public issue.

Use the repository's **Security** tab and **Report a vulnerability** option when it is available. If private vulnerability reporting is unavailable, open a public issue containing only a request for a private maintainer contact and no sensitive or exploit details.

Include, when possible:

- the affected version or commit;
- the impact and required preconditions;
- minimal reproduction steps or a proof of concept;
- any suggested mitigation or fix; and
- whether the issue is already public elsewhere.

Please allow time to investigate and coordinate a fix before public disclosure. If a real bot token or Transmission credential was exposed, revoke or rotate it immediately rather than waiting for a code change.

## Deployment considerations

This bot can control Transmission and can delete downloaded data. Restrict `-master` to trusted users, prefer numeric Telegram user IDs, protect the host and Transmission RPC endpoint, and store credentials outside source control.
