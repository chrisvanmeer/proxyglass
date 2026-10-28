package server

const dashboardHTML = `<!DOCTYPE html>
<html lang="en" class="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Proxyglass</title>
<link rel="icon" type="image/svg+xml" href="/static/img/proxyglass-mark.svg">
<link rel="stylesheet" href="/static/css/output.css?v=0.1.2">
<script src="/static/js/htmx.min.js"></script>
<style>
html,body{min-height:100%;margin:0}body{font-family:Inter,ui-sans-serif,system-ui,sans-serif;background:#070b16;color:#e2e8f0}header{position:sticky;top:0;background:rgba(10,16,30,.88);backdrop-filter:blur(18px);border-bottom:1px solid rgba(255,255,255,.1)}main{width:min(1800px,calc(100% - 48px));margin:0 auto;padding:28px 0}#content>.grid{display:grid;grid-template-columns:minmax(220px,280px) minmax(0,1fr);gap:24px}#content aside form,#content section>div{background:linear-gradient(145deg,rgba(21,31,51,.94),rgba(12,19,34,.9));border:1px solid rgba(148,163,184,.18);border-radius:18px;box-shadow:0 18px 50px rgba(0,0,0,.24)}#content aside form{padding:20px}#content section{display:grid;gap:20px;align-content:start}#content section>div{overflow:hidden}#content h2{font-size:1.05rem}#content h3{color:#dbeafe}#content table{border-collapse:collapse}#content th,#content td{padding:12px 16px;border-bottom:1px solid rgba(255,255,255,.06);text-align:left}#content th{font-size:.7rem;color:#94a3b8;text-transform:uppercase;letter-spacing:.08em}#content tr:hover{background:rgba(255,255,255,.035)}input{color:inherit}input[type=checkbox]{accent-color:#22d3ee}a,button{cursor:pointer}.pg-select{position:relative;min-width:190px}.pg-select-trigger{width:100%;display:flex;align-items:center;justify-content:space-between;gap:12px;padding:10px 12px;color:#e2e8f0;text-align:left;background:linear-gradient(135deg,rgba(30,41,59,.92),rgba(15,23,42,.98));border:1px solid rgba(148,163,184,.24);border-radius:12px;box-shadow:inset 0 1px rgba(255,255,255,.06);transition:border-color .18s,box-shadow .18s}.pg-select-trigger:hover,.pg-select-trigger[aria-expanded=true]{border-color:rgba(103,232,249,.62);box-shadow:0 0 0 3px rgba(34,211,238,.08)}.pg-select-menu{position:absolute;z-index:100;top:calc(100% + 8px);left:0;right:0;padding:6px;background:rgba(10,16,30,.98);border:1px solid rgba(148,163,184,.24);border-radius:14px;box-shadow:0 18px 48px rgba(0,0,0,.55),0 0 24px rgba(34,211,238,.08);backdrop-filter:blur(20px)}.pg-select-menu[hidden]{display:none!important}.pg-select-option{width:100%;display:flex;align-items:center;justify-content:space-between;gap:10px;padding:10px;border:0;border-radius:9px;color:#cbd5e1;background:transparent;text-align:left;cursor:pointer}.pg-select-option:hover,.pg-select-option[aria-selected=true]{color:#f0fdff;background:linear-gradient(90deg,rgba(8,145,178,.2),rgba(124,58,237,.12))}.pg-select-option small{display:block;margin-top:2px;color:#8492a8;font-size:11px}.pg-select-option[aria-selected=true] .pg-select-check{color:#67e8f9;opacity:1}.pg-select-check{opacity:0;color:#67e8f9}.pg-refresh{min-width:132px}.pg-refresh .pg-select-trigger{padding:7px 10px;border-radius:9px;font-size:13px}.pg-refresh .pg-select-menu{width:238px;left:auto}.pg-metric{min-width:0}.pg-metric-value{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}@media(max-width:700px){main{width:calc(100% - 28px);padding:18px 0}#content>.grid{grid-template-columns:1fr}#content aside{order:0}#content section{order:1}header>div{padding-left:14px;padding-right:14px}}
.pg-fleet-grid{display:grid!important;grid-template-columns:repeat(4,minmax(0,1fr));gap:16px;width:100%}.pg-fleet-card{min-width:0}@media(max-width:900px){.pg-fleet-grid{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:560px){.pg-fleet-grid{grid-template-columns:1fr}}
.pg-brand-status{display:inline-block;width:7px;height:7px;margin-left:5px;vertical-align:middle;border-radius:50%;background:#34d399;box-shadow:0 0 10px rgba(52,211,153,.85),0 0 20px rgba(52,211,153,.3)}
</style>
<style>
.pg-topology-panel{margin:0 0 24px;padding:20px;border:1px solid rgba(148,163,184,.18);border-radius:18px;background:linear-gradient(145deg,rgba(21,31,51,.94),rgba(12,19,34,.9));box-shadow:0 18px 50px rgba(0,0,0,.24)}.pg-topology-heading{display:flex;align-items:center;justify-content:space-between;gap:16px;margin-bottom:12px}.pg-topology-toolbar{display:flex;align-items:center;gap:8px}.pg-topology-toolbar button{min-width:36px;height:34px;padding:0 10px;color:#cbd5e1;background:rgba(255,255,255,.05);border:1px solid rgba(148,163,184,.2);border-radius:9px}.pg-topology-toolbar button:hover{border-color:rgba(103,232,249,.55);color:#a5f3fc}.pg-topology-zoom{min-width:48px;color:#94a3b8;text-align:center;font-size:12px}.pg-topology-help{margin:0 0 12px;color:#94a3b8;font-size:12px}.pg-topology-viewport{height:min(68vh,720px);min-height:340px;overflow:hidden;border:1px solid rgba(148,163,184,.12);border-radius:14px;background:radial-gradient(ellipse at top,rgba(34,211,238,.045),transparent 70%),rgba(2,6,23,.4);touch-action:none;cursor:grab}.pg-topology-viewport.is-dragging{cursor:grabbing;user-select:none}.pg-topology-svg{display:block;width:100%;height:100%;user-select:none}.pg-topology-edge{fill:none;stroke:rgba(148,163,184,.38);stroke-width:2;vector-effect:non-scaling-stroke}.pg-topology-edge-target{stroke:rgba(148,163,184,.28);stroke-dasharray:4 6}.pg-topology-edge-route{stroke:#22d3ee;stroke-width:2.5}.pg-topology-edge-server{stroke:rgba(103,232,249,.45)}.pg-topology-node{cursor:pointer}.pg-topology-node rect{fill:rgba(15,23,42,.96);stroke:rgba(148,163,184,.3);stroke-width:1.5;vector-effect:non-scaling-stroke}.pg-topology-node:hover rect,.pg-topology-node:focus rect{stroke:#67e8f9;stroke-width:2.5}.pg-topology-node circle{fill:#fbbf24}.pg-topology-state-healthy circle{fill:#34d399}.pg-topology-state-unhealthy circle{fill:#fb7185}.pg-topology-node text{pointer-events:none;font-family:Inter,ui-sans-serif,system-ui,sans-serif}.pg-topology-node-title{fill:#e2e8f0;font-size:13px;font-weight:600}.pg-topology-node-meta{fill:#94a3b8;font-size:10px}.pg-topology-legend{display:flex;flex-wrap:wrap;gap:16px;margin-top:12px;color:#94a3b8;font-size:11px}.pg-topology-legend span{display:inline-flex;align-items:center;gap:6px}.pg-topology-legend i{width:8px;height:8px;border-radius:50%;background:#67e8f9}.pg-topology-legend .legend-unhealthy{background:#fb7185}.pg-topology-empty{padding:32px;color:#94a3b8;text-align:center}.pg-topology-highlight{outline:2px solid #22d3ee;outline-offset:3px;transition:outline-color .2s}@media(max-width:700px){.pg-topology-panel{padding:14px}.pg-topology-heading{align-items:flex-start;flex-direction:column}.pg-topology-viewport{height:55vh;min-height:300px}}
</style>
<style>
.pg-topology-scrim{position:fixed;z-index:46;top:72px;left:0;right:0;bottom:0;padding:0 24px 24px;background:rgba(2,6,23,.52);backdrop-filter:blur(6px);overflow-y:auto}.pg-topology-drawer{width:min(1680px,100%);margin:12px auto 24px;padding:22px;background:linear-gradient(145deg,rgba(15,23,42,.98),rgba(8,15,29,.98));border:1px solid rgba(103,232,249,.2);border-radius:20px;box-shadow:0 28px 90px rgba(0,0,0,.65),0 0 40px rgba(34,211,238,.06);backdrop-filter:blur(20px)}@media(max-width:700px){.pg-topology-scrim{top:64px;padding:0 10px 12px}.pg-topology-drawer{padding:14px;margin-top:8px}}
</style>
<style>
[hidden]{display:none!important}.pg-filter-scrim{position:fixed;z-index:45;top:72px;left:0;right:0;bottom:0;padding:0 24px 24px;background:rgba(2,6,23,.52);backdrop-filter:blur(5px);overflow-y:auto}.pg-filter-drawer{width:min(1680px,100%);margin:12px auto 24px;padding:22px;background:linear-gradient(145deg,rgba(15,23,42,.98),rgba(8,15,29,.98));border:1px solid rgba(103,232,249,.2);border-radius:20px;box-shadow:0 28px 90px rgba(0,0,0,.65),0 0 40px rgba(34,211,238,.06)}.pg-filter-header{display:flex;align-items:center;justify-content:space-between;margin-bottom:18px}.pg-filter-groups{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:14px}.pg-filter-group{min-width:0;padding:16px;background:rgba(255,255,255,.025);border:1px solid rgba(148,163,184,.13);border-radius:14px}.pg-filter-options{display:grid;gap:3px;max-height:36vh;overflow-y:auto;padding-right:5px}.pg-filter-option{display:flex;align-items:center;gap:9px;min-height:34px;padding:6px 8px;color:#cbd5e1;font-size:13px;border-radius:8px}.pg-filter-option:hover{background:rgba(255,255,255,.05)}.pg-filter-option span{overflow:hidden;text-overflow:ellipsis}.pg-filter-option small{margin-left:auto;color:#718096;font-size:10px;white-space:nowrap}.pg-filter-server{padding-left:29px;color:#9aa9bd;font-size:12px}.pg-filter-scrim .pg-select{min-width:0}.pg-filter-scrim .pg-select-menu{min-width:250px}@media(max-width:1050px){.pg-filter-groups{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:700px){.pg-filter-scrim{top:64px;padding:0 10px 12px}.pg-filter-drawer{padding:14px;margin-top:8px}.pg-filter-groups{grid-template-columns:1fr}.pg-filter-options{max-height:24vh}main{width:calc(100% - 20px);padding:16px 0}header>div{padding-left:12px;padding-right:12px}}
</style>
<style>.pg-empty-row td{padding:28px!important;color:#94a3b8;text-align:center!important;font-size:13px}</style>
</head>
<body class="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 text-slate-100 font-sans">
<div class="fixed inset-0 overflow-hidden pointer-events-none">
<div class="absolute -top-40 -right-40 w-80 h-80 bg-cyan-500/10 rounded-full blur-3xl"></div>
<div class="absolute -bottom-40 -left-40 w-80 h-80 bg-violet-500/10 rounded-full blur-3xl"></div>
</div>
<div class="relative z-10">
<header class="backdrop-blur-xl bg-white/5 border-b border-white/10 sticky top-0 z-50">
<div class="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
<div class="flex items-center gap-3">
<img src="/static/img/proxyglass-mark.svg" alt="" width="38" height="38" class="rounded-xl shadow-lg shadow-cyan-500/20">
<h1 class="text-xl font-bold bg-gradient-to-r from-cyan-400 to-violet-400 bg-clip-text text-transparent">Proxyglass<span class="pg-brand-status" aria-hidden="true"></span></h1>
</div>
<div class="flex items-center gap-4">
<button id="filterToggle" type="button" aria-expanded="false" aria-controls="filterPanel" class="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-cyan-400/10 border border-cyan-300/20 hover:bg-cyan-400/15 text-sm text-cyan-100 transition">
<svg width="16" height="16" viewBox="0 0 20 20" fill="none" aria-hidden="true"><path d="M3 5h14M5.5 10h9M8 15h4" stroke="currentColor" stroke-width="1.7" stroke-linecap="round"/></svg>
Filters
</button>
<button id="topologyToggle" type="button" aria-expanded="false" aria-controls="topologyPanel" class="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-violet-400/10 border border-violet-300/20 hover:bg-violet-400/15 text-sm text-violet-100 transition">
<svg width="16" height="16" viewBox="0 0 20 20" fill="none" aria-hidden="true"><circle cx="4" cy="10" r="2" stroke="currentColor" stroke-width="1.5"/><circle cx="15.5" cy="4" r="2" stroke="currentColor" stroke-width="1.5"/><circle cx="15.5" cy="16" r="2" stroke="currentColor" stroke-width="1.5"/><path d="m5.8 9 7.8-4M5.8 11l7.8 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
Topology
</button>
<form id="refreshForm" class="flex items-center gap-2 backdrop-blur bg-white/5 px-3 py-1.5 rounded-lg border border-white/10">
<span class="text-sm text-slate-400">Auto refresh</span>
<div class="pg-select pg-refresh" data-select data-persist="refresh">
<input type="hidden" name="refresh" value="{{.Prefs.Refresh}}" data-select-value>
<button type="button" class="pg-select-trigger" data-select-trigger aria-haspopup="listbox" aria-expanded="false"><span data-select-label>{{if eq .Prefs.Refresh 0}}Off{{else}}{{.Prefs.Refresh}}s{{end}}</span><svg width="14" height="14" viewBox="0 0 20 20" fill="none" aria-hidden="true"><path d="m5 7.5 5 5 5-5" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/></svg></button>
<div class="pg-select-menu" data-select-menu role="listbox" hidden>
<button type="button" class="pg-select-option" data-select-option data-value="0" data-label="Off" aria-selected="{{if eq .Prefs.Refresh 0}}true{{else}}false{{end}}"><span><strong>Off</strong><small>Refresh manually</small></span><span class="pg-select-check">✓</span></button>
<button type="button" class="pg-select-option" data-select-option data-value="2" data-label="2s" aria-selected="{{if eq .Prefs.Refresh 2}}true{{else}}false{{end}}"><span><strong>Every 2 seconds</strong><small>Live updates</small></span><span class="pg-select-check">✓</span></button>
<button type="button" class="pg-select-option" data-select-option data-value="5" data-label="5s" aria-selected="{{if eq .Prefs.Refresh 5}}true{{else}}false{{end}}"><span><strong>Every 5 seconds</strong><small>Balanced refresh</small></span><span class="pg-select-check">✓</span></button>
<button type="button" class="pg-select-option" data-select-option data-value="10" data-label="10s" aria-selected="{{if eq .Prefs.Refresh 10}}true{{else}}false{{end}}"><span><strong>Every 10 seconds</strong><small>Lower request rate</small></span><span class="pg-select-check">✓</span></button>
<button type="button" class="pg-select-option" data-select-option data-value="30" data-label="30s" aria-selected="{{if eq .Prefs.Refresh 30}}true{{else}}false{{end}}"><span><strong>Every 30 seconds</strong><small>Low-frequency updates</small></span><span class="pg-select-check">✓</span></button>
{{if and (ne .Prefs.Refresh 0) (ne .Prefs.Refresh 2) (ne .Prefs.Refresh 5) (ne .Prefs.Refresh 10) (ne .Prefs.Refresh 30)}}<button type="button" class="pg-select-option" data-select-option data-value="{{.Prefs.Refresh}}" data-label="{{.Prefs.Refresh}}s" aria-selected="true"><span><strong>{{.Prefs.Refresh}} seconds</strong><small>Configured default</small></span><span class="pg-select-check">✓</span></button>{{end}}
</div>
</div>
</form>
<button hx-post="/api/refresh" hx-target="#content" hx-swap="innerHTML" class="px-3 py-1.5 rounded-lg bg-white/5 border border-white/10 hover:bg-white/10 text-sm transition">Refresh</button>
<a href="/logout" class="px-3 py-1.5 rounded-lg bg-white/5 border border-white/10 hover:bg-white/10 text-sm transition">Logout</a>
</div>
</div>
</header>
<div id="filterPanel" class="pg-filter-scrim" hidden>
<form id="filterForm" hx-post="/api/prefs" hx-target="#content" hx-swap="innerHTML" class="pg-filter-drawer">
<input type="hidden" name="filters" value="1">
<input type="hidden" name="targets_filter" value="1">
<input type="hidden" name="frontends_filter" value="1">
<input type="hidden" name="backends_filter" value="1">
<input type="hidden" name="servers_filter" value="1">
<input type="hidden" name="refresh" value="{{.Prefs.Refresh}}">
<div class="pg-filter-header">
<div><div class="text-xs uppercase tracking-[.22em] text-cyan-300/70">View controls</div><h2 class="mt-1 text-xl font-semibold">Filter &amp; sort</h2></div>
<button id="filterClose" type="button" class="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-300 hover:bg-white/10">Close <span aria-hidden="true">×</span></button>
</div>
<div class="pg-filter-groups">
<div class="pg-filter-group">
<h3 class="font-semibold mb-3">Sort by</h3>
<div class="pg-select" data-select data-submit="true">
<input type="hidden" name="sort" value="{{.Prefs.SortKey}}" data-select-value>
<button type="button" class="pg-select-trigger" data-select-trigger aria-haspopup="listbox" aria-expanded="false"><span data-select-label>{{if eq .Prefs.SortKey "config"}}Configuration order{{else if eq .Prefs.SortKey "status"}}Status{{else if eq .Prefs.SortKey "scur"}}Current sessions{{else if eq .Prefs.SortKey "bin"}}Bytes in{{else}}Server name{{end}}</span><svg width="14" height="14" viewBox="0 0 20 20" fill="none" aria-hidden="true"><path d="m5 7.5 5 5 5-5" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/></svg></button>
<div class="pg-select-menu" data-select-menu role="listbox" hidden>
<button type="button" class="pg-select-option" data-select-option data-value="config" aria-selected="{{if eq .Prefs.SortKey "config"}}true{{else}}false{{end}}"><span><strong>Configuration order</strong><small>Targets in config.yml order</small></span><span class="pg-select-check">✓</span></button>
<button type="button" class="pg-select-option" data-select-option data-value="name" aria-selected="{{if eq .Prefs.SortKey "name"}}true{{else}}false{{end}}"><span><strong>Server name</strong><small>Alphabetical order</small></span><span class="pg-select-check">✓</span></button>
<button type="button" class="pg-select-option" data-select-option data-value="status" aria-selected="{{if eq .Prefs.SortKey "status"}}true{{else}}false{{end}}"><span><strong>Status</strong><small>Group healthy and unavailable servers</small></span><span class="pg-select-check">✓</span></button>
<button type="button" class="pg-select-option" data-select-option data-value="scur" aria-selected="{{if eq .Prefs.SortKey "scur"}}true{{else}}false{{end}}"><span><strong>Current sessions</strong><small>Compare active connections</small></span><span class="pg-select-check">✓</span></button>
<button type="button" class="pg-select-option" data-select-option data-value="bin" aria-selected="{{if eq .Prefs.SortKey "bin"}}true{{else}}false{{end}}"><span><strong>Bytes in</strong><small>Compare inbound traffic</small></span><span class="pg-select-check">✓</span></button>
</div>
</div>
<div id="sortDirection" class="pg-select mt-3" data-select data-submit="true" {{if eq .Prefs.SortKey "config"}}hidden{{end}}>
<input type="hidden" name="asc" value="{{if .Prefs.SortAsc}}1{{else}}0{{end}}" data-select-value>
<button type="button" class="pg-select-trigger" data-select-trigger aria-haspopup="listbox" aria-expanded="false"><span data-select-label>{{if .Prefs.SortAsc}}Ascending{{else}}Descending{{end}}</span><svg width="14" height="14" viewBox="0 0 20 20" fill="none" aria-hidden="true"><path d="m5 7.5 5 5 5-5" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/></svg></button>
<div class="pg-select-menu" data-select-menu role="listbox" hidden>
<button type="button" class="pg-select-option" data-select-option data-value="1" aria-selected="{{if .Prefs.SortAsc}}true{{else}}false{{end}}"><span><strong>Ascending</strong><small>A to Z · low to high</small></span><span class="pg-select-check">✓</span></button>
<button type="button" class="pg-select-option" data-select-option data-value="0" aria-selected="{{if not .Prefs.SortAsc}}true{{else}}false{{end}}"><span><strong>Descending</strong><small>Z to A · high to low</small></span><span class="pg-select-check">✓</span></button>
</div>
</div>
</div>
<div class="pg-filter-group">
<h3 class="font-semibold mb-3">HAProxy servers</h3>
<div class="pg-filter-options">
{{range .Targets}}
<label class="pg-filter-option"><input type="checkbox" name="tg_target:{{.}}" data-target-filter {{if or (index $.Prefs.ShowTargets (printf "target:%s" .)) (not $.Prefs.TargetFilterActive)}}checked{{end}}><span>{{.}}</span></label>
{{end}}
</div>
</div>
<div class="pg-filter-group">
<h3 class="font-semibold mb-3">Frontends</h3>
<div class="pg-filter-options">
{{range .Frontends}}
{{if .TargetShown}}
<label class="pg-filter-option"><input type="checkbox" name="fe_{{.Key}}" {{if or (index $.Prefs.ShowFrontends .Key) (not $.Prefs.FrontendFilterActive)}}checked{{end}} onchange="this.form.requestSubmit()"><span>{{.Name}}</span><small>{{.Target}}</small></label>
{{else if or (index $.Prefs.ShowFrontends .Key) (not $.Prefs.FrontendFilterActive)}}
<input type="hidden" name="fe_{{.Key}}" value="1">
{{end}}
{{end}}
</div>
</div>
<div class="pg-filter-group">
<h3 class="font-semibold mb-3">Backends &amp; servers</h3>
<div class="pg-filter-options">
{{range .Backends}}
{{$backend := .}}
{{if .TargetShown}}
<label class="pg-filter-option"><input type="checkbox" name="be_{{.Key}}" data-backend-toggle="{{.Key}}" {{if or (index $.Prefs.ShowBackends .Key) (not $.Prefs.BackendFilterActive)}}checked{{end}}><span>{{.Name}}</span><small>{{.Target}}</small></label>
{{range .Servers}}
<label class="pg-filter-option pg-filter-server"><input type="checkbox" name="sv_{{.Key}}" data-backend-server="{{$backend.Key}}" {{if or (index $.Prefs.ShowServers .Key) (not $.Prefs.ServerFilterActive)}}checked{{end}}><span>{{.Name}}</span></label>
{{end}}
{{else}}
{{if or (index $.Prefs.ShowBackends .Key) (not $.Prefs.BackendFilterActive)}}<input type="hidden" name="be_{{.Key}}" value="1">{{end}}
{{range .Servers}}
{{if or (index $.Prefs.ShowServers .Key) (not $.Prefs.ServerFilterActive)}}<input type="hidden" name="sv_{{.Key}}" value="1">{{end}}
{{end}}
{{end}}
{{end}}
</div>
</div>
</div>
</form>
</div>
<div id="topologyPanel" class="pg-topology-scrim" hidden>
<section class="pg-topology-drawer" role="dialog" aria-modal="true" aria-labelledby="pgTopologyHeading">
<div class="pg-topology-heading">
<div><div class="text-xs uppercase tracking-[.22em] text-cyan-300/70">Environment map</div><h2 id="pgTopologyHeading" class="mt-1 text-xl font-semibold">Topology overview</h2></div>
<button id="topologyClose" type="button" class="rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-slate-300 hover:bg-white/10">Close <span aria-hidden="true">×</span></button>
</div>
<div id="topologyContent">{{template "topology-content" .}}</div>
</section>
</div>
<main class="mx-auto w-full px-6 py-6">
<div id="content" hx-get="/api/refresh" {{if gt .Prefs.Refresh 0}}hx-trigger="every {{.Prefs.Refresh}}s"{{end}} hx-target="#content" hx-swap="innerHTML">
{{template "content" .}}
</div>
</main>
</div>
<script>
let topologyViewState=null;
let topologyDrag=null;
let topologySuppressClickUntil=0;
function topologySvg(){return document.getElementById('topologySvg')}
function clampTopologyView(){
  const state=topologyViewState;
  state.width=Math.min(state.baseWidth,Math.max(state.baseWidth/4,state.width));
  state.height=state.width*state.baseHeight/state.baseWidth;
  state.x=Math.max(0,Math.min(state.baseWidth-state.width,state.x));
  state.y=Math.max(0,Math.min(state.baseHeight-state.height,state.y));
}
function applyTopologyView(svg){
  if(!svg||!topologyViewState)return;
  const state=topologyViewState;
  svg.setAttribute('viewBox',[state.x,state.y,state.width,state.height].map(value=>value.toFixed(2)).join(' '));
  const level=document.querySelector('[data-topology-zoom-level]');
  if(level)level.textContent=Math.round(state.baseWidth/state.width*100)+'%';
}
function initTopology(svg){
  if(!svg){topologyViewState=null;return}
  const baseWidth=Number(svg.dataset.baseWidth);
  const baseHeight=Number(svg.dataset.baseHeight);
  if(!topologyViewState||topologyViewState.baseWidth!==baseWidth||topologyViewState.baseHeight!==baseHeight){
    topologyViewState={x:0,y:0,width:baseWidth,height:baseHeight,baseWidth:baseWidth,baseHeight:baseHeight};
  }
  applyTopologyView(svg);
}
function zoomTopology(factor){
  const svg=topologySvg();
  if(!svg)return;
  initTopology(svg);
  const state=topologyViewState;
  const width=Math.min(state.baseWidth,Math.max(state.baseWidth/4,state.width/factor));
  const ratio=width/state.width;
  const centerX=state.x+state.width/2;
  const centerY=state.y+state.height/2;
  state.x=centerX-(centerX-state.x)*ratio;
  state.y=centerY-(centerY-state.y)*ratio;
  state.width=width;
  clampTopologyView();
  applyTopologyView(svg);
}
function resetTopology(){
  const svg=topologySvg();
  if(!svg)return;
  initTopology(svg);
  topologyViewState.x=0;
  topologyViewState.y=0;
  topologyViewState.width=topologyViewState.baseWidth;
  topologyViewState.height=topologyViewState.baseHeight;
  applyTopologyView(svg);
}
initTopology(topologySvg());
document.addEventListener('click',async function(event){
  const topologyAction=event.target.closest('[data-topology-action]');
  if(topologyAction){
    if(topologyAction.dataset.topologyAction==='zoom-in')zoomTopology(1.2);
    else if(topologyAction.dataset.topologyAction==='zoom-out')zoomTopology(1/1.2);
    else resetTopology();
    return;
  }
  const topologyNode=event.target.closest('[data-topology-node]');
  if(topologyNode){
    if(Date.now()<topologySuppressClickUntil)return;
    const anchor=topologyNode.dataset.topologyAnchor;
    const destination=anchor&&document.getElementById(anchor);
    if(destination){
      setTopologyPanel(false);
      destination.scrollIntoView({behavior:'smooth',block:'center'});
      destination.classList.add('pg-topology-highlight');
      setTimeout(()=>destination.classList.remove('pg-topology-highlight'),1400);
    }
    return;
  }
  const option=event.target.closest('[data-select-option]');
  if(option){
    const root=option.closest('[data-select]');
    const value=root.querySelector('[data-select-value]');
    const label=root.querySelector('[data-select-label]');
    value.value=option.dataset.value;
    label.textContent=option.dataset.label||option.querySelector('strong').textContent;
    root.querySelectorAll('[data-select-option]').forEach(item=>item.setAttribute('aria-selected',item===option?'true':'false'));
    root.querySelector('[data-select-menu]').hidden=true;
    root.querySelector('[data-select-trigger]').setAttribute('aria-expanded','false');
    if(root.dataset.persist==='refresh'){
      const response=await fetch('/api/prefs',{method:'POST',body:new FormData(root.closest('form')),credentials:'same-origin'});
      if(!response.ok){throw new Error('Unable to save refresh preference')}
      const content=document.getElementById('content');
      const seconds=Number(value.value);
      if(content){
        if(seconds>0){content.setAttribute('hx-trigger','every '+seconds+'s')}
        else{content.removeAttribute('hx-trigger')}
        htmx.process(content);
      }
    }else if(root.dataset.submit==='true'){
      if(value.name==='sort'){
        const direction=document.getElementById('sortDirection');
        if(direction){direction.hidden=value.value==='config'}
      }
      root.closest('form').requestSubmit();
    }
    return;
  }
  const trigger=event.target.closest('[data-select-trigger]');
  document.querySelectorAll('[data-select-menu]').forEach(menu=>{if(!trigger||menu!==trigger.parentElement.querySelector('[data-select-menu]')){menu.hidden=true;menu.parentElement.querySelector('[data-select-trigger]').setAttribute('aria-expanded','false')}});
  if(trigger){const menu=trigger.parentElement.querySelector('[data-select-menu]');menu.hidden=!menu.hidden;trigger.setAttribute('aria-expanded',menu.hidden?'false':'true')}
});
document.addEventListener('keydown',function(event){if(event.key==='Escape'){document.querySelectorAll('[data-select-menu]').forEach(menu=>{menu.hidden=true;menu.parentElement.querySelector('[data-select-trigger]').setAttribute('aria-expanded','false')})}});
document.addEventListener('keydown',function(event){if((event.key==='Enter'||event.key===' ')&&event.target.matches('[data-topology-node]')){event.preventDefault();event.target.click()}});
document.addEventListener('wheel',function(event){
  if(!event.target.closest('#pgTopologyViewport'))return;
  event.preventDefault();
  zoomTopology(event.deltaY<0?1.12:1/1.12);
},{passive:false});
document.addEventListener('pointerdown',function(event){
  if(event.button!==0||!event.target.closest('#pgTopologyViewport'))return;
  const svg=topologySvg();
  if(!svg)return;
  initTopology(svg);
  topologyDrag={startX:event.clientX,startY:event.clientY,lastX:event.clientX,lastY:event.clientY,moved:false};
});
document.addEventListener('pointermove',function(event){
  if(!topologyDrag)return;
  const svg=topologySvg();
  const viewport=document.getElementById('pgTopologyViewport');
  const rect=svg&&svg.getBoundingClientRect();
  if(!svg||!rect||!viewport)return;
  const dx=event.clientX-topologyDrag.lastX;
  const dy=event.clientY-topologyDrag.lastY;
  if(Math.abs(event.clientX-topologyDrag.startX)+Math.abs(event.clientY-topologyDrag.startY)>4)topologyDrag.moved=true;
  if(topologyDrag.moved){
    const fitScale=Math.min(rect.width/topologyViewState.width,rect.height/topologyViewState.height);
    topologyViewState.x-=dx/fitScale;
    topologyViewState.y-=dy/fitScale;
    clampTopologyView();
    applyTopologyView(svg);
    viewport.classList.add('is-dragging');
    event.preventDefault();
  }
  topologyDrag.lastX=event.clientX;
  topologyDrag.lastY=event.clientY;
});
function finishTopologyDrag(){
  if(!topologyDrag)return;
  if(topologyDrag.moved)topologySuppressClickUntil=Date.now()+200;
  topologyDrag=null;
  document.getElementById('pgTopologyViewport')?.classList.remove('is-dragging');
}
document.addEventListener('pointerup',finishTopologyDrag);
document.addEventListener('pointercancel',finishTopologyDrag);
document.addEventListener('htmx:afterSwap',function(event){if(event.detail&&event.detail.target&&['content','topologyContent'].includes(event.detail.target.id))initTopology(topologySvg())});
const filterPanel=document.getElementById('filterPanel');
const filterToggle=document.getElementById('filterToggle');
const topologyPanel=document.getElementById('topologyPanel');
const topologyToggle=document.getElementById('topologyToggle');
function setTopologyPanel(open){
  const wasOpen=!topologyPanel.hidden;
  if(open){filterPanel.hidden=true;filterToggle.setAttribute('aria-expanded','false')}
  topologyPanel.hidden=!open;
  topologyToggle.setAttribute('aria-expanded',open?'true':'false');
  if(open){document.getElementById('topologyClose').focus()}
  else if(wasOpen){topologyToggle.focus()}
}
function setFilterPanel(open){
  if(open)setTopologyPanel(false);
  filterPanel.hidden=!open;
  filterToggle.setAttribute('aria-expanded',open?'true':'false');
  if(open){document.getElementById('filterClose').focus()}
}
filterToggle.addEventListener('click',function(){setFilterPanel(filterPanel.hidden)});
topologyToggle.addEventListener('click',function(){setTopologyPanel(topologyPanel.hidden)});
document.getElementById('filterClose').addEventListener('click',function(){setFilterPanel(false)});
document.getElementById('topologyClose').addEventListener('click',function(){setTopologyPanel(false)});
filterPanel.addEventListener('click',function(event){if(event.target===filterPanel)setFilterPanel(false)});
document.addEventListener('keydown',function(event){if(event.key==='Escape'&&!filterPanel.hidden)setFilterPanel(false)});
topologyPanel.addEventListener('click',function(event){if(event.target===topologyPanel)setTopologyPanel(false)});
document.addEventListener('keydown',function(event){if(event.key==='Escape'&&!topologyPanel.hidden)setTopologyPanel(false)});
document.addEventListener('change',function(event){
  const input=event.target;
  if(input.matches('[data-target-filter]')){
    const form=input.form;
    form.addEventListener('htmx:afterRequest',function(requestEvent){
      const xhr=requestEvent.detail&&requestEvent.detail.xhr;
      if(xhr&&xhr.status>=200&&xhr.status<300){try{sessionStorage.setItem('proxyglass-reopen-filters','1')}catch(_e){}location.reload()}
    },{once:true});
    form.requestSubmit();
    return;
  }
  if(input.matches('[data-backend-toggle]')){
    const key=input.dataset.backendToggle;
    filterPanel.querySelectorAll('[data-backend-server]').forEach(function(server){if(server.dataset.backendServer===key)server.checked=input.checked});
    input.indeterminate=false;
    input.form.requestSubmit();
    return;
  }
  if(input.matches('[data-backend-server]')){
    const key=input.dataset.backendServer;
    const parent=Array.from(filterPanel.querySelectorAll('[data-backend-toggle]')).find(function(item){return item.dataset.backendToggle===key});
    const servers=Array.from(filterPanel.querySelectorAll('[data-backend-server]')).filter(function(server){return server.dataset.backendServer===key});
    const selected=servers.filter(function(server){return server.checked}).length;
    parent.checked=selected>0;
    parent.indeterminate=selected>0&&selected<servers.length;
    input.form.requestSubmit();
  }
});
filterPanel.querySelectorAll('[data-backend-toggle]').forEach(function(parent){
  const key=parent.dataset.backendToggle;
  const servers=Array.from(filterPanel.querySelectorAll('[data-backend-server]')).filter(function(server){return server.dataset.backendServer===key});
  const selected=servers.filter(function(server){return server.checked}).length;
  parent.checked=selected>0;
  parent.indeterminate=selected>0&&selected<servers.length;
});
try{if(sessionStorage.getItem('proxyglass-reopen-filters')==='1'){sessionStorage.removeItem('proxyglass-reopen-filters');setFilterPanel(true)}}catch(_e){}
</script>
</body>
</html>

{{define "content"}}
{{if .Errors}}
<div class="mb-5 rounded-xl border border-rose-400/20 bg-rose-500/10 p-4">
<h2 class="mb-2 text-sm font-semibold text-rose-200">Some HAProxy targets could not be read</h2>
{{range $target, $err := .Errors}}<p class="text-xs text-rose-200/80">{{$target}}: {{$err}}</p>{{end}}
</div>
{{end}}
<div class="mb-6 w-full">
<div class="flex items-center justify-between mb-4">
<div>
<div class="text-xs uppercase tracking-[.22em] text-cyan-300/70">Fleet overview</div>
<h2 class="mt-1 text-xl font-semibold">HAProxy at a glance</h2>
</div>
<div class="text-sm text-slate-400">Live view</div>
</div>
<div class="pg-fleet-grid">
<div class="pg-fleet-card rounded-2xl backdrop-blur-xl bg-gradient-to-br from-white/5 to-white/[0.02] border border-white/10 p-5"><div class="text-xs text-slate-400">HAProxy proxies</div><div class="mt-1 text-2xl font-semibold">{{.Fleet.Proxies}}</div><div class="mt-1 text-xs text-slate-500">Configured targets</div></div>
<div class="pg-fleet-card rounded-2xl backdrop-blur-xl bg-gradient-to-br from-white/5 to-white/[0.02] border border-white/10 p-5"><div class="text-xs text-slate-400">Active sessions</div><div class="mt-1 text-2xl font-semibold">{{.Fleet.ActiveSess}}</div><div class="mt-1 text-xs text-slate-500">Across frontends</div></div>
<div class="pg-fleet-card rounded-2xl backdrop-blur-xl bg-gradient-to-br from-white/5 to-white/[0.02] border border-white/10 p-5"><div class="text-xs text-slate-400">Healthy servers</div><div class="mt-1 text-2xl font-semibold">{{.Fleet.ServersUp}} <span class="text-base text-slate-500">/ {{.Fleet.ServersTotal}}</span></div><div class="mt-1 text-xs text-slate-500">{{if gt .Fleet.ServersTotal 0}}{{$pct := divf (mulf (toFloat64 .Fleet.ServersUp) 100.0) (toFloat64 .Fleet.ServersTotal)}}{{printf "%.1f" $pct}}% available{{else}}No servers{{end}}</div></div>
<div class="pg-fleet-card rounded-2xl backdrop-blur-xl bg-gradient-to-br from-white/5 to-white/[0.02] border border-white/10 p-5"><div class="text-xs text-slate-400">Bytes in (frontends)</div><div class="mt-1 text-2xl font-semibold">{{humanBytes .Fleet.BytesInTotal}}</div><div class="mt-1 text-xs text-slate-500">Total received</div></div>
</div>
</div>
<section class="space-y-6">
{{template "frontends" .}}
{{range .Backends}}
{{if .Shown}}
<div id="backend-{{.Key}}" class="backdrop-blur-xl bg-gradient-to-br from-white/5 to-white/[0.02] border border-white/10 rounded-2xl overflow-hidden shadow-xl">
<div class="px-6 py-4 border-b border-white/10 flex items-center justify-between">
<div class="flex items-center gap-3"><div class="w-2 h-2 rounded-full {{if eq .Status "UP"}}bg-emerald-400 shadow-lg shadow-emerald-400/50{{else if or (eq .Status "DOWN") (eq .Status "DRAIN")}}bg-red-400{{else}}bg-amber-400{{end}}"></div><h2 class="text-lg font-semibold">{{.Name}}</h2><span class="text-xs px-2 py-0.5 rounded-full bg-white/5">{{.Target}}</span><span class="text-xs px-2 py-0.5 rounded-full {{if eq .Status "UP"}}bg-emerald-500/20 text-emerald-300{{else if eq .Status "DOWN"}}bg-red-500/20 text-red-300{{else}}bg-amber-500/20 text-amber-300{{end}}">{{.Status}}</span></div>
<div class="text-sm text-slate-400">Sessions {{.SCur}}/{{.SMax}} • In {{humanBytes .BIn}} • Out {{humanBytes .BOut}} <span class="ml-2 text-cyan-200/75" title="{{if .LastActivityAt}}{{.LastActivityAt}}{{else}}No activity recorded{{end}}">Last activity {{.LastActivity}}</span></div>
</div>
<div class="overflow-x-auto">
<table class="w-full">
<thead class="bg-white/5">
<tr>
<th class="text-left px-6 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider">Server</th>
<th class="text-left px-6 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider">Status</th>
<th class="text-left px-6 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider">Check</th>
<th class="text-left px-6 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider">Sessions</th>
<th class="text-left px-6 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider">Bytes In/Out</th>
<th class="text-left px-6 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider">Last activity</th>
<th class="text-left px-6 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider">LastChk</th>
<th class="text-left px-6 py-3 text-xs font-medium text-slate-400 uppercase tracking-wider">LastChkDur</th>
</tr>
</thead>
<tbody class="divide-y divide-white/5">
{{if not .Servers}}<tr class="pg-empty-row"><td colspan="8">No backend server rows were returned by this HAProxy stats endpoint.</td></tr>{{else if not .VisibleServers}}<tr class="pg-empty-row"><td colspan="8">All servers in this backend are hidden by your current filters.</td></tr>{{end}}
{{range .Servers}}
{{if .Shown}}
<tr id="server-{{.Key}}" class="hover:bg-white/5 transition-colors">
<td class="px-6 py-3"><div class="flex items-center gap-2"><div class="w-1.5 h-1.5 rounded-full {{if eq .Status "UP"}}bg-emerald-400{{else if eq .Status "DOWN"}}bg-red-400{{else}}bg-amber-400{{end}}"></div><span class="font-medium">{{.Name}}</span></div></td>
<td class="px-6 py-3"><span class="text-xs px-2 py-1 rounded-full {{if eq .Status "UP"}}bg-emerald-500/20 text-emerald-300{{else if eq .Status "DOWN"}}bg-red-500/20 text-red-300{{else if eq .Status "MAINT"}}bg-orange-500/20 text-orange-300{{else}}bg-amber-500/20 text-amber-300{{end}}">{{.Status}}</span></td>
<td class="px-6 py-3 text-sm text-slate-400">{{.CheckStatus}}</td>
<td class="px-6 py-3 text-sm">{{.SCur}}/{{.SMax}}{{if .SLim}} ({{.SLim}}){{end}}</td>
<td class="px-6 py-3 text-sm text-slate-400">{{humanBytes .BIn}} / {{humanBytes .BOut}}</td>
<td class="px-6 py-3 text-sm text-cyan-100" title="HAProxy last-session age">{{.LastActivity}}</td>
<td class="px-6 py-3 text-sm text-slate-400">{{.LastChk}}{{if .CheckCode}} <span class="text-amber-300">({{.CheckCode}})</span>{{end}}</td>
<td class="px-6 py-3 text-sm text-slate-400">{{.LastChkDur}}</td>
</tr>
{{end}}
{{end}}
</tbody>
</table>
</div>
</div>
{{end}}
{{end}}
</section>
{{end}}
{{define "topology-content"}}
<p class="pg-topology-help">Drag to pan, scroll to zoom, and click a node to jump to its stats. Frontend routes appear when configured in <code>topology</code>.</p>
{{if .Topology.Nodes}}
<div id="pgTopologyViewport" class="pg-topology-viewport">
<svg id="topologySvg" class="pg-topology-svg" viewBox="0 0 {{.Topology.Width}} {{.Topology.Height}}" data-base-width="{{.Topology.Width}}" data-base-height="{{.Topology.Height}}" preserveAspectRatio="xMinYMin meet" role="group" aria-label="HAProxy environment topology">
<g data-topology-viewport>
{{range .Topology.Edges}}<path class="pg-topology-edge pg-topology-edge-{{.Kind}}" d="{{.Path}}"/>{{end}}
{{range .Topology.Nodes}}
<g class="pg-topology-node pg-topology-node-{{.Kind}} pg-topology-state-{{.StateClass}}" data-topology-node data-topology-anchor="{{.Anchor}}" transform="translate({{printf "%.1f" .X}} {{printf "%.1f" .Y}})" tabindex="0" role="button" aria-label="{{.Kind}} {{.Label}}, {{.Status}}">
<title>{{.Kind}}: {{.Label}} · {{.Target}} · {{.Status}}</title>
<rect width="190" height="70" rx="13"/>
<circle cx="16" cy="19" r="5"/>
<text class="pg-topology-node-title" x="29" y="24">{{.Label}}</text>
<text class="pg-topology-node-meta" x="15" y="49">{{.Target}} · {{.Status}}</text>
</g>
{{end}}
</g>
</svg>
</div>
<div class="pg-topology-legend" aria-label="Topology legend"><span><i></i>HAProxy / frontend / backend</span><span><i class="legend-unhealthy"></i>Unavailable</span><span>Solid cyan line: configured frontend route</span></div>
{{else}}
<div class="pg-topology-empty">No HAProxy targets are visible. Add targets to the Proxyglass configuration or adjust the active filters.</div>
{{end}}
{{end}}
{{define "topology-oob"}}<div id="topologyContent" hx-swap-oob="innerHTML">{{template "topology-content" .}}</div>{{end}}
{{define "frontends"}}
{{range .Frontends}}
{{if .Shown}}
<div id="frontend-{{.Key}}" class="backdrop-blur-xl bg-gradient-to-br from-white/5 to-white/[0.02] border border-white/10 rounded-2xl overflow-hidden shadow-xl">
<div class="px-6 py-4 border-b border-white/10 flex items-center justify-between">
<div class="flex items-center gap-3"><div class="w-2 h-2 rounded-full {{if eq .Status "UP"}}bg-emerald-400 shadow-lg shadow-emerald-400/50{{else if eq .Status "OPEN"}}bg-emerald-400{{else}}bg-amber-400{{end}}"></div><h2 class="text-lg font-semibold">{{.Name}}</h2><span class="text-xs px-2 py-0.5 rounded-full bg-white/5">{{.Target}}</span><span class="text-xs px-2 py-0.5 rounded-full {{if eq .Status "UP"}}bg-emerald-500/20 text-emerald-300{{else}}bg-amber-500/20 text-amber-300{{end}}">{{.Status}}</span></div>
</div>
<div class="grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-7 gap-3 p-5">
<div class="rounded-xl bg-white/[0.035] border border-white/[0.06] p-3"><div class="text-xs text-slate-400">Sessions now</div><div class="mt-1 text-lg font-semibold">{{.SCur}} <span class="text-xs font-normal text-slate-500">/ {{.SMax}} max</span></div></div>
<div class="rounded-xl bg-white/[0.035] border border-white/[0.06] p-3"><div class="text-xs text-slate-400">Total sessions</div><div class="mt-1 text-lg font-semibold">{{.STot}}</div></div>
<div class="rounded-xl bg-white/[0.035] border border-white/[0.06] p-3"><div class="text-xs text-slate-400">Request rate</div><div class="mt-1 text-lg font-semibold">{{.Rate}} <span class="text-xs font-normal text-slate-500">/s</span></div><div class="text-[11px] text-slate-500">Peak {{.RateMax}}/s</div></div>
<div class="rounded-xl bg-white/[0.035] border border-white/[0.06] p-3"><div class="text-xs text-slate-400">Bytes in</div><div class="mt-1 text-lg font-semibold">{{humanBytes .BIn}}</div></div>
<div class="rounded-xl bg-white/[0.035] border border-white/[0.06] p-3"><div class="text-xs text-slate-400">Bytes out</div><div class="mt-1 text-lg font-semibold">{{humanBytes .BOut}}</div></div>
<div class="rounded-xl bg-white/[0.035] border border-white/[0.06] p-3"><div class="text-xs text-slate-400">Denied / errors</div><div class="mt-1 text-lg font-semibold">{{.Denied}} <span class="text-xs font-normal text-slate-500">/ {{.Errors}}</span></div></div>
<div class="rounded-xl bg-white/[0.035] border border-cyan-300/[0.12] p-3" title="Last traffic-counter change observed by Proxyglass"><div class="text-xs text-slate-400">Last activity</div><div class="mt-1 text-sm font-semibold text-cyan-100"{{if .LastActivityAt}} title="{{.LastActivityAt}}"{{end}}>{{.LastActivity}}</div></div>
</div>
</div>
{{end}}
{{end}}
{{end}}
`
