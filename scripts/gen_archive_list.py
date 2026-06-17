#!/usr/bin/env python3
import csv
import json
import sys
import urllib.request
from pathlib import Path

sys.stdout.reconfigure(encoding="utf-8")

USER = "iPmartNetwork"
OUT = Path(__file__).parent / "archive-repos.csv"

KEEP_EXACT = {
    "VortexUI", "RatholePro", "TunnelMaster", "RatholeTunnel",
    "Install-MikroTik-CHR", "iPmartNetwork", "Songbird", "BirdX",
    "BirdX.Chat", "wgpanelbot", "XMPlus", "Veltrix", "passwall2",
    "iPShadowT", "iPBotS",
}
KEEP_STARS_GE = 5
KEEP_UPDATED_SINCE = "2025-06-01"

repos = []
for page in (1, 2):
    url = f"https://api.github.com/users/{USER}/repos?per_page=100&page={page}&sort=updated"
    req = urllib.request.Request(url, headers={"User-Agent": "archive-csv"})
    batch = json.loads(urllib.request.urlopen(req, timeout=30).read())
    if not batch:
        break
    repos.extend(batch)

rows = []
for r in repos:
    if r.get("archived"):
        continue
    name = r["name"]
    stars = r["stargazers_count"]
    updated = r["updated_at"][:10]
    is_fork = r.get("fork", False)
    desc = (r.get("description") or "").strip()

    if name in KEEP_EXACT or stars >= KEEP_STARS_GE or updated >= KEEP_UPDATED_SINCE:
        action = "KEEP"
        phase = ""
        reason = "flagship, stars, or recent"
    else:
        score = 0
        reasons = []
        if is_fork:
            score += 3
            reasons.append("fork")
        if updated < "2025-01-01":
            score += 2
            reasons.append("stale")
        if stars == 0 and r["forks_count"] == 0:
            score += 2
            reasons.append("0 engagement")
        if not desc:
            score += 1
            reasons.append("no desc")
        if score >= 3:
            action = "ARCHIVE"
            phase = "1" if stars == 0 and updated < "2025-01-01" else "2"
            reason = ", ".join(reasons)
        else:
            action = "REVIEW"
            phase = "3"
            reason = ", ".join(reasons) or "manual check"

    rows.append({
        "action": action,
        "phase": phase,
        "name": name,
        "stars": stars,
        "forks": r["forks_count"],
        "updated": updated,
        "is_fork": is_fork,
        "reason": reason,
        "settings_url": f"https://github.com/{USER}/{name}/settings",
        "repo_url": f"https://github.com/{USER}/{name}",
    })

order = {"ARCHIVE": 0, "REVIEW": 1, "KEEP": 2}
rows.sort(key=lambda x: (order[x["action"]], x["phase"], x["stars"], x["name"]))

with OUT.open("w", newline="", encoding="utf-8-sig") as f:
    w = csv.DictWriter(f, fieldnames=list(rows[0].keys()))
    w.writeheader()
    w.writerows(rows)

archive = [r for r in rows if r["action"] == "ARCHIVE"]
print(f"Wrote {OUT}")
print(f"ARCHIVE: {len(archive)} | REVIEW: {sum(1 for r in rows if r['action']=='REVIEW')} | KEEP: {sum(1 for r in rows if r['action']=='KEEP')}")
print("\nPhase 1 (open these):")
for r in archive:
    if r["phase"] == "1":
        print(f"  {r['settings_url']}")
