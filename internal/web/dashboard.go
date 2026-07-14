package web

var dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>DevKit — Universal CLI Dashboard</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#09090b;color:#e4e4e7;min-height:100vh}
.app{display:flex;height:100vh}
.sidebar{width:260px;background:#18181b;border-right:1px solid #27272a;padding:20px 0;flex-shrink:0;display:flex;flex-direction:column}
.sidebar-header{padding:0 20px 20px;border-bottom:1px solid #27272a}
.sidebar-header h1{font-size:18px;font-weight:700}
.sidebar-header h1 span{color:#22c55e}
.sidebar-header p{color:#52525b;font-size:12px;margin-top:4px}
.nav{flex:1;padding:12px 10px;overflow-y:auto}
.nav-item{display:flex;align-items:center;gap:10px;padding:10px 14px;border-radius:8px;cursor:pointer;color:#a1a1aa;font-size:14px;font-weight:500;transition:all .15s;margin-bottom:2px}
.nav-item:hover{background:#27272a;color:#e4e4e7}
.nav-item.active{background:#22c55e15;color:#22c55e;border:1px solid #22c55e30}
.nav-icon{font-size:18px;width:24px;text-align:center}
.nav-section{padding:16px 14px 6px;font-size:11px;text-transform:uppercase;letter-spacing:1px;color:#52525b;font-weight:600}
.main{flex:1;overflow-y:auto;padding:24px}
.topbar{display:flex;align-items:center;justify-content:space-between;margin-bottom:24px}
.topbar h2{font-size:22px;font-weight:600}
.topbar .actions{display:flex;gap:8px}
.btn{padding:8px 16px;border:none;border-radius:8px;font-size:13px;font-weight:500;cursor:pointer;transition:all .15s}
.btn-primary{background:#22c55e;color:#000}.btn-primary:hover{background:#16a34a}
.btn-danger{background:#ef4444;color:#fff}.btn-danger:hover{background:#dc2626}
.btn-secondary{background:#27272a;color:#e4e4e7}.btn-secondary:hover{background:#3f3f46}
.btn-sm{padding:5px 10px;font-size:12px}
.card{background:#18181b;border:1px solid #27272a;border-radius:12px;padding:20px;margin-bottom:16px}
.card h3{font-size:14px;font-weight:600;color:#a1a1aa;margin-bottom:12px;text-transform:uppercase;letter-spacing:.5px}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:12px}
.stat-card{background:#09090b;border:1px solid #27272a;border-radius:10px;padding:16px;text-align:center}
.stat-card .value{font-size:28px;font-weight:700;color:#22c55e}
.stat-card .label{font-size:12px;color:#52525b;margin-top:4px}
.cmd-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(220px,1fr));gap:10px}
.cmd-btn{background:#09090b;border:1px solid #27272a;border-radius:10px;padding:14px;cursor:pointer;transition:all .15s;text-align:left}
.cmd-btn:hover{border-color:#22c55e;background:#22c55e08}
.cmd-btn .cmd-title{font-size:14px;font-weight:600;color:#e4e4e7;margin-bottom:4px}
.cmd-btn .cmd-desc{font-size:12px;color:#52525b}
.output-box{background:#09090b;border:1px solid #27272a;border-radius:10px;padding:16px;font-family:'Fira Code',Consolas,monospace;font-size:13px;line-height:1.6;max-height:500px;overflow:auto;white-space:pre-wrap;color:#a1a1aa}
.service-row{display:flex;align-items:center;justify-content:space-between;padding:12px 16px;background:#09090b;border:1px solid #27272a;border-radius:8px;margin-bottom:8px}
.service-info{display:flex;align-items:center;gap:12px}
.dot{width:10px;height:10px;border-radius:50%}
.dot-on{background:#22c55e;box-shadow:0 0 8px #22c55e55}
.dot-off{background:#52525b}
.param-input{width:100%;padding:10px 14px;background:#09090b;border:1px solid #27272a;border-radius:8px;color:#e4e4e7;font-size:14px;margin-bottom:8px;font-family:monospace}
.param-input:focus{outline:none;border-color:#22c55e}
.param-input.textarea{min-height:80px;resize:vertical}
.hidden{display:none}
.toast{position:fixed;bottom:24px;right:24px;background:#22c55e;color:#000;padding:12px 20px;border-radius:8px;font-size:14px;font-weight:500;opacity:0;transition:opacity .3s;pointer-events:none;z-index:100}
.toast.show{opacity:1}
.spinner{display:inline-block;width:14px;height:14px;border:2px solid #27272a;border-top-color:#22c55e;border-radius:50%;animation:spin .6s linear infinite}
@keyframes spin{to{transform:rotate(360deg)}}
.modal-overlay{position:fixed;inset:0;background:#000a;display:flex;align-items:center;justify-content:center;z-index:50}
.modal{background:#18181b;border:1px solid #27272a;border-radius:12px;padding:24px;width:400px;max-width:90vw}
.modal h3{margin-bottom:16px;font-size:16px}
.modal .actions{display:flex;gap:8px;justify-content:flex-end;margin-top:16px}
</style>
</head>
<body>
<div class="app">
  <div class="sidebar">
    <div class="sidebar-header">
      <h1><span>Dev</span>Kit</h1>
      <p>Universal CLI Dashboard</p>
    </div>
    <div class="nav">
      <div class="nav-section">General</div>
      <div class="nav-item active" onclick="showPage('dashboard')"><span class="nav-icon">📊</span>Dashboard</div>
      <div class="nav-item" onclick="showPage('services')"><span class="nav-icon">🔌</span>Services</div>
      <div class="nav-section">Tools</div>
      <div id="pluginNav"></div>
    </div>
  </div>
  <div class="main" id="main-content"></div>
</div>
<div class="toast" id="toast"></div>

<script>
let currentPage = 'dashboard';
let currentPlugin = null;
let plugins = [];
let services = [];

function toast(msg) {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.classList.add('show');
  setTimeout(() => t.classList.remove('show'), 2500);
}

async function api(url, method, body) {
  const opts = { method: method || 'GET', headers: {'Content-Type':'application/json'} };
  if (body) opts.body = JSON.stringify(body);
  const r = await fetch(url, opts);
  return r.json();
}

async function loadPlugins() {
  const data = await api('/api/plugins');
  plugins = data.plugins || [];
  const nav = document.getElementById('pluginNav');
  nav.innerHTML = plugins.map(p =>
    '<div class="nav-item" id="nav-'+p.name+'" onclick="showPlugin(\''+p.name+'\')"><span class="nav-icon">'+p.icon+'</span>'+p.label+'</div>'
  ).join('');
}

async function loadServices() {
  const data = await api('/api/services');
  services = data.services || [];
}

function showPage(page) {
  currentPage = page;
  currentPlugin = null;
  document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
  event.target.closest('.nav-item')?.classList.add('active');
  render();
}

function showPlugin(name) {
  currentPage = 'plugin';
  currentPlugin = plugins.find(p => p.name === name);
  document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
  document.getElementById('nav-'+name)?.classList.add('active');
  render();
}

async function execCommand(plugin, command, params) {
  toast('Running ' + command + '...');
  const data = await api('/api/exec/' + plugin + '/' + command, 'POST', params || {});
  return data.output;
}

async function startService(name) {
  toast('Starting ' + name + '...');
  await api('/api/start/' + name, 'POST');
  await loadServices();
  render();
  toast(name + ' started!');
}

async function stopService(name) {
  toast('Stopping ' + name + '...');
  await api('/api/stop/' + name, 'POST');
  await loadServices();
  render();
  toast(name + ' stopped!');
}

function render() {
  const main = document.getElementById('main-content');
  if (currentPage === 'dashboard') renderDashboard(main);
  else if (currentPage === 'services') renderServices(main);
  else if (currentPage === 'plugin' && currentPlugin) renderPlugin(main, currentPlugin);
}

function renderDashboard(el) {
  const running = services.length;
  const enabled = plugins.length;
  el.innerHTML =
    '<div class="topbar"><h2>Dashboard</h2></div>' +
    '<div class="grid" style="margin-bottom:24px">' +
      '<div class="stat-card"><div class="value">'+running+'</div><div class="label">Running Services</div></div>' +
      '<div class="stat-card"><div class="value">'+enabled+'</div><div class="label">Enabled Tools</div></div>' +
      '<div class="stat-card"><div class="value">'+plugins.reduce((a,p)=>a+p.commands.length,0)+'</div><div class="label">Available Commands</div></div>' +
    '</div>' +
    '<div class="card"><h3>Quick Actions</h3><div class="cmd-grid">' +
      plugins.flatMap(p => p.commands.slice(0,3).map(c =>
        '<div class="cmd-btn" onclick="runFromDash(\''+p.name+'\',\''+c.name+'\')"><div class="cmd-title">'+p.icon+' '+c.label+'</div><div class="cmd-desc">'+p.label+'</div></div>'
      )).join('') +
    '</div></div>' +
    '<div class="card"><h3>Output</h3><div class="output-box" id="dashOutput">Click a command above to see output</div></div>';
}

function renderServices(el) {
  el.innerHTML =
    '<div class="topbar"><h2>Services</h2><div class="actions"><button class="btn btn-danger" onclick="stopAll()">Stop All</button></div></div>' +
    '<div class="card">' +
    (services.length ? services.map(s =>
      '<div class="service-row"><div class="service-info"><div class="dot dot-on"></div><span style="font-weight:500">'+s.name+'</span><span style="color:#52525b;font-family:monospace">:'+s.port+'</span></div><div><button class="btn btn-sm btn-secondary" onclick="viewLogs(\''+s.name+'\')">Logs</button> <button class="btn btn-sm btn-danger" onclick="stopService(\''+s.name+'\')">Stop</button></div></div>'
    ).join('') : '<div style="text-align:center;padding:40px;color:#52525b">No services running</div>') +
    '<div class="grid" style="margin-top:16px">' +
      '<button class="btn btn-primary" onclick="startService(\'postgres\')">+ PostgreSQL</button>' +
      '<button class="btn btn-primary" onclick="startService(\'redis\')">+ Redis</button>' +
      '<button class="btn btn-primary" onclick="startService(\'mysql\')">+ MySQL</button>' +
      '<button class="btn btn-primary" onclick="startService(\'mongo\')">+ MongoDB</button>' +
    '</div></div>' +
    '<div class="card"><h3>Logs</h3><div class="output-box" id="logOutput">Select a service and click Logs</div></div>';
}

function renderPlugin(el, plugin) {
  const isRunning = plugin.name === 'postgres' || plugin.name === 'redis' || services.some(s => s.name === plugin.name);
  el.innerHTML =
    '<div class="topbar"><h2>'+plugin.icon+' '+plugin.label+'</h2><div class="actions">' +
      (isRunning ? '<button class="btn btn-danger btn-sm" onclick="stopService(\''+plugin.name+'\')">Stop</button>' :
      ['postgres','redis','mysql','mongo'].includes(plugin.name) ? '<button class="btn btn-primary btn-sm" onclick="startService(\''+plugin.name+'\')">Start</button>' : '') +
    '</div></div>' +
    '<div class="card"><h3>'+plugin.description+'</h3><div class="cmd-grid">' +
    plugin.commands.map(c =>
      '<div class="cmd-btn" onclick="runCommand2(\''+plugin.name+'\',\''+c.name+'\')"><div class="cmd-title">'+c.label+'</div><div class="cmd-desc">'+(c.params && c.params.length ? 'Needs input' : 'Click to run')+'</div></div>'
    ).join('') +
    '</div></div>' +
    '<div class="card"><h3>Output</h3><div class="output-box" id="cmdOutput">Click a command above</div></div>';
}

async function runFromDash(plugin, cmdName) {
  const out = document.getElementById('dashOutput');
  out.innerHTML = '<span class="spinner"></span> Running...';
  const output = await execCommand(plugin, cmdName);
  out.textContent = output || '(no output)';
}

async function runCommand2(pluginName, cmdName) {
  const plugin = plugins.find(p => p.name === pluginName);
  const cmd = plugin.commands.find(c => c.name === cmdName);
  const out = document.getElementById('cmdOutput');

  if (cmd.params && cmd.params.length) {
    showParamModal(pluginName, cmdName, cmd.params, out);
    return;
  }

  out.innerHTML = '<span class="spinner"></span> Running...';
  const output = await execCommand(pluginName, cmdName);
  out.textContent = output || '(no output)';
}

function showParamModal(pluginName, cmdName, params, outEl) {
  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay';
  overlay.innerHTML =
    '<div class="modal"><h3>Enter Parameters</h3>' +
    params.map(p =>
      '<label style="display:block;font-size:13px;color:#a1a1aa;margin-bottom:4px">'+p.label+'</label>' +
      (p.type === 'textarea' ?
        '<textarea class="param-input textarea" id="param-'+p.name+'" placeholder="'+p.label+'"></textarea>' :
        '<input class="param-input" id="param-'+p.name+'" placeholder="'+(p.default||p.label)+'" />')
    ).join('') +
    '<div class="actions"><button class="btn btn-secondary" onclick="this.closest(\'.modal-overlay\').remove()">Cancel</button><button class="btn btn-primary" id="paramSubmit">Run</button></div></div>';

  document.body.appendChild(overlay);

  overlay.querySelector('#paramSubmit').onclick = async () => {
    const body = {};
    params.forEach(p => { body[p.name] = overlay.querySelector('#param-'+p.name).value; });
    overlay.remove();
    outEl.innerHTML = '<span class="spinner"></span> Running...';
    const output = await execCommand(pluginName, cmdName, body);
    outEl.textContent = output || '(no output)';
  };
}

async function viewLogs(name) {
  const data = await api('/api/logs/' + name);
  const el = document.getElementById('logOutput');
  el.textContent = (data.lines || []).join('\n') || 'No logs';
}

async function stopAll() {
  await api('/api/stop-all', 'POST');
  await loadServices();
  render();
  toast('All services stopped');
}

async function init() {
  await loadPlugins();
  await loadServices();
  render();
}

init();
</script>
</body>
</html>`
