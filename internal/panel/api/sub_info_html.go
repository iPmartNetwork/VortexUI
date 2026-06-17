package api

const subInfoHTML = `<!DOCTYPE html>
<html lang="en" dir="ltr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Username}} — VortexUI</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
:root{--bg:#0f1724;--surface:#1a2332;--surface2:#243044;--border:#2d3f55;--fg:#e8edf5;--fg2:#94a3b8;--accent:#6366f1;--accent2:#818cf8;--success:#22c55e;--warning:#f59e0b;--danger:#ef4444;--radius:14px}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:var(--bg);color:var(--fg);min-height:100vh;padding:20px}
.container{max-width:480px;margin:0 auto}
.header{text-align:center;padding:28px 0 20px}
.logo{font-family:'Orbitron',sans-serif;font-size:1.4rem;font-weight:800;background:linear-gradient(135deg,var(--accent),var(--accent2));-webkit-background-clip:text;-webkit-text-fill-color:transparent}
.username{margin-top:8px;font-size:1.1rem;font-weight:600;color:var(--fg)}
.status{display:inline-block;margin-top:6px;padding:3px 12px;border-radius:20px;font-size:.7rem;font-weight:600;text-transform:uppercase;letter-spacing:.5px}
.status-active{background:rgba(34,197,94,.15);color:var(--success)}
.status-limited{background:rgba(245,158,11,.15);color:var(--warning)}
.status-expired{background:rgba(239,68,68,.15);color:var(--danger)}
.status-disabled{background:rgba(148,163,184,.15);color:var(--fg2)}
.card{background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);padding:18px;margin-top:14px}
.card-title{font-size:.7rem;font-weight:600;text-transform:uppercase;letter-spacing:.8px;color:var(--fg2);margin-bottom:12px}
.stats-grid{display:grid;grid-template-columns:1fr 1fr;gap:12px}
.stat{text-align:center;padding:12px;background:var(--surface2);border-radius:10px}
.stat-value{font-size:1.3rem;font-weight:700;color:var(--fg)}
.stat-label{font-size:.65rem;color:var(--fg2);margin-top:3px;text-transform:uppercase;letter-spacing:.3px}
.progress-wrap{margin-top:14px}
.progress-bar{height:8px;background:var(--surface2);border-radius:4px;overflow:hidden}
.progress-fill{height:100%;border-radius:4px;background:linear-gradient(90deg,var(--accent),var(--accent2));transition:width .6s ease}
.progress-labels{display:flex;justify-content:space-between;margin-top:6px;font-size:.7rem;color:var(--fg2)}
.links-section{margin-top:14px}
.link-row{display:flex;align-items:center;gap:8px;padding:10px 12px;background:var(--surface2);border-radius:10px;margin-bottom:8px;cursor:pointer;transition:background .2s}
.link-row:hover{background:var(--border)}
.link-row:active{transform:scale(.98)}
.link-text{flex:1;font-size:.72rem;font-family:monospace;color:var(--fg2);white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.link-badge{font-size:.6rem;font-weight:700;padding:2px 7px;border-radius:6px;text-transform:uppercase;background:var(--accent);color:#fff;white-space:nowrap}
.copy-btn{padding:5px 10px;border-radius:7px;border:1px solid var(--border);background:transparent;color:var(--fg2);font-size:.65rem;cursor:pointer;transition:all .2s}
.copy-btn:hover{border-color:var(--accent);color:var(--accent)}
.sub-links{margin-top:14px}
.sub-link{display:flex;align-items:center;justify-content:space-between;padding:10px 14px;background:var(--surface2);border-radius:10px;margin-bottom:8px}
.sub-link-label{font-size:.75rem;font-weight:600;color:var(--fg)}
.qr-section{text-align:center;margin-top:14px;padding:18px;background:var(--surface2);border-radius:10px}
.qr-section img{width:200px;height:200px;border-radius:8px}
.qr-label{font-size:.7rem;color:var(--fg2);margin-top:8px}
.footer{text-align:center;padding:24px 0;font-size:.65rem;color:var(--fg2)}
.toast{position:fixed;bottom:20px;left:50%;transform:translateX(-50%);background:var(--success);color:#fff;padding:8px 20px;border-radius:8px;font-size:.75rem;font-weight:600;opacity:0;transition:opacity .3s;pointer-events:none;z-index:99}
.toast.show{opacity:1}
</style>
<link href="https://fonts.googleapis.com/css2?family=Orbitron:wght@700;800&display=swap" rel="stylesheet">
</head>
<body>
<div class="container">
  <div class="header">
    <div class="logo">VortexUI</div>
    <div class="username">{{.Username}}</div>
    <span class="status status-{{.Status}}">{{.Status}}</span>
  </div>

  <!-- Usage Card -->
  <div class="card">
    <div class="card-title">Traffic Usage</div>
    <div class="stats-grid">
      <div class="stat"><div class="stat-value">{{.UsedGB}}</div><div class="stat-label">Used (GB)</div></div>
      <div class="stat"><div class="stat-value">{{.LimitGB}}</div><div class="stat-label">Limit (GB)</div></div>
      {{if ge .DaysLeft 0}}<div class="stat"><div class="stat-value">{{.DaysLeft}}</div><div class="stat-label">Days Left</div></div>{{end}}
      <div class="stat"><div class="stat-value">{{.DeviceCount}}/{{.DeviceLimit}}</div><div class="stat-label">Devices</div></div>
    </div>
    <div class="progress-wrap">
      <div class="progress-bar"><div class="progress-fill" style="width:{{.UsedPercent}}%"></div></div>
      <div class="progress-labels"><span>{{.UsedPercent}}% used</span><span>{{.LimitGB}} GB</span></div>
    </div>
  </div>

  <!-- Subscription Links -->
  <div class="card">
    <div class="card-title">Subscription Links</div>
    <div class="sub-links">
      <div class="sub-link"><span class="sub-link-label">Auto (Universal)</span><button class="copy-btn" onclick="copy('{{.SubURL}}')">Copy</button></div>
      <div class="sub-link"><span class="sub-link-label">Clash / Mihomo</span><button class="copy-btn" onclick="copy('{{.ClashURL}}')">Copy</button></div>
      <div class="sub-link"><span class="sub-link-label">Sing-Box</span><button class="copy-btn" onclick="copy('{{.SingboxURL}}')">Copy</button></div>
      <div class="sub-link"><span class="sub-link-label">Base64</span><button class="copy-btn" onclick="copy('{{.Base64URL}}')">Copy</button></div>
    </div>
  </div>

  <!-- QR Code -->
  <div class="card">
    <div class="card-title">QR Code</div>
    <div class="qr-section">
      <img src="https://api.qrserver.com/v1/create-qr-code/?size=200x200&data={{.SubURL}}" alt="QR">
      <div class="qr-label">Scan in your proxy client</div>
    </div>
  </div>

  <!-- Configs -->
  <div class="card">
    <div class="card-title">Configs ({{.ConfigCount}})</div>
    <div class="links-section">
      {{range $i, $link := .Links}}
      <div class="link-row" onclick="copy('{{$link}}')">
        <span class="link-text">{{$link}}</span>
        <button class="copy-btn">Copy</button>
      </div>
      {{end}}
    </div>
  </div>

  <!-- Traffic Chart (7-day) -->
  <div class="card">
    <div class="card-title">Traffic (last 7 days)</div>
    <canvas id="trafficChart" height="120" style="width:100%"></canvas>
  </div>

  <!-- Renew / Purchase -->
  <div class="card">
    <div class="card-title">Renew / Upgrade</div>
    <div style="text-align:center;padding:10px 0">
      <a href="/api/shop/plans" target="_blank" style="display:inline-block;padding:10px 24px;background:linear-gradient(135deg,var(--accent),var(--accent2));color:#fff;border-radius:10px;text-decoration:none;font-size:.8rem;font-weight:600">View Plans & Purchase</a>
    </div>
  </div>

  <div class="footer">© 2026 iPmart Network. All rights reserved.</div>
</div>

<div class="toast" id="toast">Copied!</div>
<script>
function copy(t){navigator.clipboard.writeText(t).then(()=>{const e=document.getElementById('toast');e.classList.add('show');setTimeout(()=>e.classList.remove('show'),1500)})}

// Multi-language support (detect browser language)
(function(){
  const labels={
    fa:{usage:'مصرف ترافیک',used:'مصرف (GB)',limit:'سقف (GB)',days:'روز باقیمانده',devices:'دستگاه',sub:'لینک اشتراک',configs:'کانفیگ‌ها',qr:'QR Code',scan:'با اپ پروکسی اسکن کنید',traffic:'ترافیک (۷ روز اخیر)'},
    ar:{usage:'استخدام الحركة',used:'مستخدم (GB)',limit:'الحد (GB)',days:'أيام متبقية',devices:'أجهزة',sub:'رابط الاشتراك',configs:'الإعدادات',qr:'رمز QR',scan:'امسح بتطبيق البروكسي',traffic:'حركة المرور (٧ أيام)'},
    tr:{usage:'Trafik Kullanımı',used:'Kullanılan (GB)',limit:'Limit (GB)',days:'Kalan Gün',devices:'Cihaz',sub:'Abonelik Bağlantıları',configs:'Yapılandırmalar',qr:'QR Kod',scan:'Proxy uygulamanızla tarayın',traffic:'Trafik (son 7 gün)'},
    ru:{usage:'Использование трафика',used:'Использовано (ГБ)',limit:'Лимит (ГБ)',days:'Дней осталось',devices:'Устройства',sub:'Ссылки подписки',configs:'Конфигурации',qr:'QR-код',scan:'Сканируйте прокси-приложением',traffic:'Трафик (7 дней)'},
    zh:{usage:'流量使用',used:'已用 (GB)',limit:'限制 (GB)',days:'剩余天数',devices:'设备',sub:'订阅链接',configs:'配置',qr:'二维码',scan:'用代理应用扫描',traffic:'流量（7天）'},
  };
  const lang=navigator.language?.slice(0,2)||'en';
  if(labels[lang]){
    document.documentElement.lang=lang;
    if(lang==='fa'||lang==='ar')document.documentElement.dir='rtl';
    const l=labels[lang];
    document.querySelectorAll('.card-title').forEach(el=>{
      const t=el.textContent.trim().toLowerCase();
      if(t.includes('traffic usage'))el.textContent=l.usage;
      else if(t.includes('subscription'))el.textContent=l.sub;
      else if(t.includes('qr'))el.textContent=l.qr;
      else if(t.includes('configs'))el.textContent=l.configs;
      else if(t.includes('traffic'))el.textContent=l.traffic;
    });
    document.querySelectorAll('.stat-label').forEach(el=>{
      const t=el.textContent.trim().toLowerCase();
      if(t.includes('used'))el.textContent=l.used;
      else if(t.includes('limit'))el.textContent=l.limit;
      else if(t.includes('days'))el.textContent=l.days;
      else if(t.includes('device'))el.textContent=l.devices;
    });
    document.querySelectorAll('.qr-label').forEach(el=>{el.textContent=l.scan});
  }
})();

// Traffic chart
fetch(window.location.pathname.replace('/info','/usage')).then(r=>r.json()).then(d=>{
  const canvas=document.getElementById('trafficChart');
  if(!canvas||!d.points||!d.points.length)return;
  const ctx=canvas.getContext('2d');
  const W=canvas.width=canvas.offsetWidth;
  const H=canvas.height=120;
  const pts=d.points;
  const vals=pts.map(p=>(p.up||0)+(p.down||0));
  const max=Math.max(...vals,1);
  const barW=Math.floor((W-20)/vals.length)-2;
  ctx.fillStyle='rgba(99,102,241,0.3)';
  ctx.strokeStyle='rgba(99,102,241,0.8)';
  ctx.lineWidth=1;
  vals.forEach((v,i)=>{
    const x=10+i*(barW+2);
    const h=(v/max)*(H-30);
    ctx.fillRect(x,H-20-h,barW,h);
    ctx.strokeRect(x,H-20-h,barW,h);
  });
  ctx.fillStyle='rgba(148,163,184,0.7)';
  ctx.font='9px sans-serif';
  pts.forEach((p,i)=>{
    if(i%2===0){
      const x=10+i*(barW+2);
      const day=new Date(p.time).toLocaleDateString(undefined,{month:'short',day:'numeric'});
      ctx.fillText(day,x,H-5);
    }
  });
}).catch(()=>{});
</script>
</body>
</html>`
