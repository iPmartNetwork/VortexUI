package api

const shopHTML = `<!DOCTYPE html>
<html lang="en" dir="ltr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Username}} — Plans — VortexUI</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
:root{--bg:#0f1724;--surface:#1a2332;--surface2:#243044;--border:#2d3f55;--fg:#e8edf5;--fg2:#94a3b8;--accent:#6366f1;--accent2:#818cf8;--success:#22c55e;--warning:#f59e0b;--danger:#ef4444;--radius:14px}
@media(prefers-color-scheme:light){:root{--bg:#f8fafc;--surface:#ffffff;--surface2:#f1f5f9;--border:#e2e8f0;--fg:#1e293b;--fg2:#64748b;--accent:#6366f1;--accent2:#818cf8;--success:#22c55e;--warning:#f59e0b;--danger:#ef4444}}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:var(--bg);color:var(--fg);min-height:100vh;padding:20px}
.container{max-width:560px;margin:0 auto}
.header{text-align:center;padding:28px 0 20px}
.logo{font-family:'Orbitron',sans-serif;font-size:1.4rem;font-weight:800;background:linear-gradient(135deg,var(--accent),var(--accent2));-webkit-background-clip:text;-webkit-text-fill-color:transparent}
.username{margin-top:8px;font-size:1rem;color:var(--fg2)}
.subtitle{margin-top:4px;font-size:.85rem;color:var(--fg2)}
.plans-grid{display:grid;gap:16px;margin-top:20px}
.plan-card{background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);padding:20px;transition:border-color .2s}
.plan-card:hover{border-color:var(--accent)}
.plan-name{font-size:1.05rem;font-weight:700;color:var(--fg)}
.plan-desc{font-size:.75rem;color:var(--fg2);margin-top:4px}
.plan-details{display:grid;grid-template-columns:1fr 1fr;gap:8px;margin-top:14px}
.plan-detail{padding:8px;background:var(--surface2);border-radius:8px;text-align:center}
.plan-detail-value{font-size:.9rem;font-weight:600;color:var(--fg)}
.plan-detail-label{font-size:.6rem;text-transform:uppercase;letter-spacing:.5px;color:var(--fg2);margin-top:2px}
.plan-price{margin-top:14px;display:flex;gap:10px;align-items:center;justify-content:center;flex-wrap:wrap}
.price-tag{font-size:.85rem;font-weight:700;color:var(--accent);background:rgba(99,102,241,.1);padding:4px 12px;border-radius:8px}
.plan-actions{margin-top:14px;display:flex;gap:8px;justify-content:center;flex-wrap:wrap}
.buy-btn{padding:10px 20px;border-radius:10px;border:none;font-size:.8rem;font-weight:600;cursor:pointer;color:#fff;background:linear-gradient(135deg,var(--accent),var(--accent2));transition:opacity .2s}
.buy-btn:hover{opacity:.85}
.buy-btn:disabled{opacity:.5;cursor:not-allowed}
.back-link{display:block;text-align:center;margin-top:24px;font-size:.8rem;color:var(--fg2);text-decoration:none}
.back-link:hover{color:var(--accent)}
.footer{text-align:center;padding:24px 0;font-size:.65rem;color:var(--fg2)}
.toast{position:fixed;bottom:20px;left:50%;transform:translateX(-50%);background:var(--accent);color:#fff;padding:10px 24px;border-radius:8px;font-size:.8rem;font-weight:600;opacity:0;transition:opacity .3s;pointer-events:none;z-index:99}
.toast.show{opacity:1}
</style>
<link href="https://fonts.googleapis.com/css2?family=Orbitron:wght@700;800&display=swap" rel="stylesheet">
</head>
<body>
<div class="container">
  <div class="header">
    <div class="logo">VortexUI</div>
    <div class="username">{{.Username}}</div>
    <div class="subtitle">Choose a plan to renew or upgrade your subscription</div>
  </div>

  <div class="plans-grid">
    {{range .Plans}}
    <div class="plan-card">
      <div class="plan-name">{{.Name}}</div>
      {{if .Description}}<div class="plan-desc">{{.Description}}</div>{{end}}
      <div class="plan-details">
        <div class="plan-detail">
          <div class="plan-detail-value">{{dataGB .DataLimit}}</div>
          <div class="plan-detail-label">Traffic</div>
        </div>
        <div class="plan-detail">
          <div class="plan-detail-value">{{.Duration}} days</div>
          <div class="plan-detail-label">Duration</div>
        </div>
        {{if gt .DeviceLimit 0}}
        <div class="plan-detail">
          <div class="plan-detail-value">{{.DeviceLimit}}</div>
          <div class="plan-detail-label">Devices</div>
        </div>
        {{end}}
      </div>
      <div class="plan-price">
        {{if gt .PriceToman 0}}<span class="price-tag">{{.PriceToman}} Toman</span>{{end}}
        {{if gt .PriceUSD 0.0}}<span class="price-tag">${{printf "%.2f" .PriceUSD}} USD</span>{{end}}
      </div>
      <div class="plan-actions">
        {{if gt .PriceToman 0}}
        <button class="buy-btn" onclick="purchase('{{.ID}}','zarinpal')">Pay with ZarinPal</button>
        {{end}}
        {{if gt .PriceUSD 0.0}}
        <button class="buy-btn" onclick="purchase('{{.ID}}','crypto')">Pay with Crypto</button>
        {{end}}
      </div>
    </div>
    {{end}}
  </div>

  <a href="/sub/{{$.Token}}/info" class="back-link">← Back to subscription info</a>
  <div class="footer">© 2026 iPmart Network. All rights reserved.</div>
</div>

<div class="toast" id="toast">Redirecting to payment...</div>
<script>
function purchase(planID, gateway) {
  const btn = event.target;
  btn.disabled = true;
  btn.textContent = 'Processing...';
  const toast = document.getElementById('toast');

  fetch('/api/shop/purchase', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      plan_id: planID,
      sub_token: '{{$.Token}}',
      gateway: gateway
    })
  })
  .then(r => r.json())
  .then(data => {
    if (data.redirect_url) {
      toast.classList.add('show');
      window.location.href = data.redirect_url;
    } else {
      alert(data.message || 'Purchase failed');
      btn.disabled = false;
      btn.textContent = gateway === 'zarinpal' ? 'Pay with ZarinPal' : 'Pay with Crypto';
    }
  })
  .catch(() => {
    alert('Network error, please try again');
    btn.disabled = false;
    btn.textContent = gateway === 'zarinpal' ? 'Pay with ZarinPal' : 'Pay with Crypto';
  });
}
</script>
</body>
</html>`
