# {Page Title} (v{VERSION})

<div style="text-align: center; margin: 1.5rem 0;">
  <strong style="font-size: 1.25rem;">VortexUI v{VERSION}</strong><br/>
  <em style="font-size: 1rem; color: var(--md-default-fg-color--light);">{One-line subtitle}</em>
</div>

!!! info "{Release type}"
    {Short note: e.g. requires migration 0030, or frontend-only rebuild.}

---

## Highlights

<div class="grid cards" markdown>

- :material-view-dashboard: **{Feature 1}**

    {One sentence description.}

- :material-cog: **{Feature 2}**

    {One sentence description.}

- :material-shield-account: **{Feature 3}**

    {One sentence description.}

</div>

---

## {Section heading}

| Item | Description |
|------|-------------|
| {Row 1} | {Detail} |
| {Row 2} | {Detail} |

!!! tip "{Tip title}"
    {Actionable tip — command, keyboard shortcut, or navigation path.}

---

## Screenshots (optional)

| Light | Dark |
|-------|------|
| ![](../assets/panel/{name}_light.png) | ![](../assets/panel/{name}_dark.png) |

---

## Related

| Doc | Link |
|-----|------|
| Changelog | [CHANGELOG.md](https://github.com/iPmartNetwork/VortexUI/blob/master/CHANGELOG.md) |
| {Related chapter} | [{title}]({path}.md) |

---

## Upgrade

=== "Panel only (no DB change)"

    ```bash
    cd web && npm run build
    # redeploy dist + restart panel service
    ```

=== "With migration"

    ```bash
    vortexui migrate   # or your deploy path
    go build -o vortexui-panel ./cmd/panel
    cd web && npm run build
    systemctl restart vortexui-panel
    ```
