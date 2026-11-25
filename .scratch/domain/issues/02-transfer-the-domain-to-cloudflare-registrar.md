# 02: Transfer the domain to Cloudflare Registrar

**What to build:** `forgespark.dev` was registered at GoDaddy on 2026-11-24, for R$ 96 the first year. GoDaddy renews it for R$ 142 a year, on 2027-11-24. Cloudflare Registrar sells `.dev` at cost, about US$ 12.20 a year, and the maintainer wants the DNS there too (see ticket 01). The transfer has two parts:

1. **Now:** move the DNS to Cloudflare. Cloudflare accepts a transfer only for a domain that is already active on it.
2. **On 2027-01-24 or later:** transfer the registration. ICANN refuses a transfer within 60 days of registration. Do it well before 2027-11-24, or GoDaddy renews for R$ 142 first.

The transfer pays one year at Cloudflare, added to the current expiry, so the expiry moves to 2028-11-24. Nothing already paid to GoDaddy is lost. The domain has no record of the system yet, so nothing goes down.

The maintainer's step-by-step guide, in Portuguese: [Transferir forgespark.dev da GoDaddy para a Cloudflare](https://claude.ai/artifact/B6vQSGTHeZGS2NwihLxMfX). In short:

- **Part 1, the DNS.** Create the Cloudflare account, with two-factor login. Add `forgespark.dev` on the Free plan. Delete the records the scan imports: GoDaddy's parking page, two `A` records on `forgespark.dev` (`15.197.148.33`, `3.33.130.190`) and a `CNAME` from `www`. At GoDaddy, replace `ns25.domaincontrol.com` and `ns26.domaincontrol.com` with the two name servers Cloudflare gives. Wait for the zone to show **Active**. DNSSEC is off for this domain: no DS record, checked on 2026-11-25. So nothing has to be turned off first.
- **Part 2, the registration.** Check that the Cloudflare account has a valid card. At GoDaddy, turn off the domain lock and ask for the authorization code. At Cloudflare, go to **Domain Registration** → **Transfer Domains**, paste the code, pay the year, fill in the registrant contact and select **Confirm transfer**. Approve the e-mail GoDaddy sends, or wait up to 5 days.
- **Until Part 2 is done:** do not change the registrant's name, organisation or e-mail at GoDaddy. ICANN locks the transfer for 60 more days after such a change.

**Blocked by:** None for part 1. Part 2 cannot start before 2027-01-24.

**Status:** ready-for-human

- [ ] `forgespark.dev` is active on Cloudflare, and `nslookup -type=NS forgespark.dev` shows Cloudflare's name servers.
- [ ] `docs/providers.md` names Cloudflare as the DNS host and gives its two name servers. Ticket 01's choice of a DNS host is recorded as Cloudflare.
- [ ] On 2027-01-24 or later, the registration was transferred. Cloudflare lists `forgespark.dev` under **Registrations**, expiring on 2028-11-24, with auto-renew on.
- [ ] Nothing at GoDaddy still bills for this domain, such as privacy or e-mail.
- [ ] `docs/providers.md` has no GoDaddy left. Cloudflare is the registrar and the DNS host, with the yearly price it charged, in its section and its row of the summary table.
