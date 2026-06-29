package api

import (
	"html/template"

	"github.com/labstack/echo/v4"
)

// PaymentResult renders a simple server-side HTML page showing the payment
// outcome (success/failed/error) based on the ?status= query parameter. This
// is the page users land on after the payment gateway redirects back.
func (h *Handlers) PaymentResult(c echo.Context) error {
	status := c.QueryParam("status")
	msg := c.QueryParam("msg")

	data := paymentResultData{
		Status:  status,
		Message: msg,
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return paymentResultTmpl.Execute(c.Response().Writer, data)
}

type paymentResultData struct {
	Status  string
	Message string
}

var paymentResultTmpl = template.Must(template.New("payment_result").Parse(paymentResultHTML))

const paymentResultHTML = `<!DOCTYPE html>
<html lang="en" dir="ltr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Payment Result — VortexUI</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
:root{--bg:#0f1724;--surface:#1a2332;--surface2:#243044;--border:#2d3f55;--fg:#e8edf5;--fg2:#94a3b8;--accent:#6366f1;--accent2:#818cf8;--success:#22c55e;--warning:#f59e0b;--danger:#ef4444;--radius:14px}
@media(prefers-color-scheme:light){:root{--bg:#f8fafc;--surface:#ffffff;--surface2:#f1f5f9;--border:#e2e8f0;--fg:#1e293b;--fg2:#64748b;--accent:#6366f1;--accent2:#818cf8;--success:#22c55e;--warning:#f59e0b;--danger:#ef4444}}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:var(--bg);color:var(--fg);min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px}
.container{max-width:420px;width:100%;text-align:center}
.logo{font-family:'Orbitron',sans-serif;font-size:1.4rem;font-weight:800;background:linear-gradient(135deg,var(--accent),var(--accent2));-webkit-background-clip:text;-webkit-text-fill-color:transparent;margin-bottom:24px}
.card{background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);padding:32px 24px}
.icon{font-size:3rem;margin-bottom:16px}
.title{font-size:1.2rem;font-weight:700;margin-bottom:8px}
.message{font-size:.85rem;color:var(--fg2);margin-bottom:20px}
.success .icon{color:var(--success)}
.failed .icon{color:var(--danger)}
.error .icon{color:var(--warning)}
.btn{display:inline-block;padding:10px 24px;background:linear-gradient(135deg,var(--accent),var(--accent2));color:#fff;border-radius:10px;text-decoration:none;font-size:.8rem;font-weight:600;margin-top:8px}
.btn:hover{opacity:.85}
.footer{margin-top:24px;font-size:.65rem;color:var(--fg2)}
</style>
<link href="https://fonts.googleapis.com/css2?family=Orbitron:wght@700;800&display=swap" rel="stylesheet">
</head>
<body>
<div class="container">
  <div class="logo">VortexUI</div>
  <div class="card {{.Status}}">
    {{if eq .Status "success"}}
      <div class="icon">&#10004;</div>
      <div class="title">Payment Successful!</div>
      <div class="message">Your subscription has been extended. The new traffic and time have been added to your account.</div>
    {{else if eq .Status "failed"}}
      <div class="icon">&#10006;</div>
      <div class="title">Payment Failed</div>
      <div class="message">The payment was not completed. No changes were made to your subscription. Please try again.</div>
    {{else if eq .Status "paid"}}
      <div class="icon">&#10004;</div>
      <div class="title">Payment Confirmed</div>
      <div class="message">Your payment has been confirmed and your subscription is being updated.</div>
    {{else}}
      <div class="icon">&#9888;</div>
      <div class="title">Error</div>
      <div class="message">{{if .Message}}{{.Message}}{{else}}An unexpected error occurred during payment processing.{{end}}</div>
    {{end}}
  </div>
  <div class="footer">© 2026 iPmart Network. All rights reserved.</div>
</div>
</body>
</html>`
