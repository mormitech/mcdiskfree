import './style.css';
import {hierarchy, partition} from 'd3-hierarchy';
import {arc} from 'd3-shape';
import {Home, ChooseFolder, Scan, Cancel, View, Reveal, Volumes, FreeSpace, Trash} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

const $ = (id) => document.getElementById(id);
const DEPTH = 4;

let scanRoot = '';    // folder that was scanned
let current = null;   // node shown now
let usage = null;     // size and free space of the scanned folder's volume
let volumes = [];
let busy = false;
const basket = new Map();   // path -> {name, size}
let basketState = 'idle';   // idle | confirm | working
let basketError = '';

function bytes(n) {
    const u = ['B', 'KiB', 'MiB', 'GiB', 'TiB'];
    let i = 0;
    while (n >= 1024 && i < u.length - 1) { n /= 1024; i++; }
    return (i === 0 ? n : n.toFixed(n >= 100 ? 0 : 1)) + ' ' + u[i];
}

const esc = (s) => s.replace(/[&<>"]/g, (c) => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;'}[c]));
const sepOf = (p) => (p.includes('\\') ? '\\' : '/');
const isUnder = (p, dir) => p.startsWith(dir.endsWith(sepOf(dir)) ? dir : dir + sepOf(dir));

// ---- start screen: all disks ------------------------------------------------------------------------------------

async function showStart() {
    if (busy) return;
    volumes = (await Volumes()) || [];
    $('splash').hidden = false;
    $('splash').innerHTML = `<div id="start"><h2>Disks</h2>` + volumes.map((v, i) => {
        const used = v.total - v.free;
        const pct = v.total ? (used / v.total) * 100 : 0;
        return `<div class="disk" data-i="${i}"><div class="t"><b>${esc(v.name)}</b><span>${bytes(used)} used</span></div>` +
            `<div class="p">${esc(v.path)}</div><div class="meter"><i class="${pct > 90 ? 'full' : ''}" style="width:${pct}%"></i></div>` +
            `<div class="f">${bytes(v.free)} free of ${bytes(v.total)}</div></div>`;
    }).join('') + `<p>Click a disk to scan it, or use “Choose folder…” for just one folder.</p></div>`;
    $('splash').onclick = (e) => {
        const d = e.target.closest('.disk');
        if (d) scan(volumes[+d.dataset.i].path);
    };
}

// ---- colours and chart --------------------------------------------------------------------------------------------

function colour(d, hue) {
    if (!d.data.path) return 'hsl(215 10% 45%)';
    const l = 52 + (d.depth - 1) * 6;
    return d.data.isDir ? `hsl(${hue} 62% ${l}%)` : `hsl(${hue} 22% ${l - 4}%)`;
}

function topHue(d, tops) {
    let t = d;
    while (t.depth > 1) t = t.parent;
    return (tops.indexOf(t.data) * 137.5 + 210) % 360;
}

function render(node) {
    current = node;
    $('splash').hidden = true;
    drawCrumbs(node.path);
    drawChart(node);
    drawList(node);
    drawBasket();
    drawSpace();
}

// At the top of a whole-disk scan the chart also shows the free space as a dashed slice.
function withFree(node) {
    const wholeDisk = node.path === scanRoot && volumes.some((v) => v.path === scanRoot);
    if (!wholeDisk || !usage || !usage.free) return node;
    const free = {name: 'Free space', path: '', isDir: false, free: true, size: usage.free, items: 0};
    return {...node, size: node.size + usage.free, children: [...(node.children || []), free]};
}

function drawChart(node) {
    const shown = withFree(node);
    const box = $('chart').getBoundingClientRect();
    const R = (Math.min(box.width, box.height) - 16) / 2;
    const hole = R * 0.26;
    const ring = (R - hole) / DEPTH;
    const svg = $('sun');
    svg.setAttribute('viewBox', `${-box.width / 2} ${-box.height / 2} ${box.width} ${box.height}`);

    const root = hierarchy(shown, (d) => d.children).sum((d) => (d.children ? 0 : d.size));
    partition().size([2 * Math.PI, 1])(root);
    const tops = shown.children || [];
    const gen = arc()
        .startAngle((d) => d.x0).endAngle((d) => d.x1)
        .padAngle(0.002)
        .innerRadius((d) => hole + (d.depth - 1) * ring)
        .outerRadius((d) => hole + d.depth * ring - 1);

    const arcs = root.descendants().filter((d) => d.depth > 0 && d.depth <= DEPTH && d.x1 - d.x0 > 0.003);
    svg.innerHTML = arcs.map((d, i) => {
        const cls = d.data.free ? 'free' : d.data.isDir ? 'dir' : d.data.path ? 'file' : 'rest';
        const gone = basket.has(d.data.path) ? ' style="opacity:.35"' : '';
        return `<path data-i="${i}" class="${cls}" d="${gen(d)}" fill="${colour(d, topHue(d, tops))}"${gone}/>`;
    }).join('');

    svg.onmousemove = (e) => {
        const p = e.target.closest('path');
        if (!p) return hideTip();
        showTip(e, arcs[+p.dataset.i].data, shown.size);
    };
    svg.onmouseleave = hideTip;
    svg.onclick = (e) => {
        const p = e.target.closest('path');
        if (!p) return;
        const d = arcs[+p.dataset.i].data;
        if (e.shiftKey && d.path) toggleBasket(d);
        else if (d.isDir) open(d.path);
    };
    svg.oncontextmenu = (e) => { e.preventDefault(); goUp(); };

    const up = node.path !== scanRoot;
    $('center').innerHTML = `<div class="name">${esc(node.name || node.path)}</div><div class="size">${bytes(node.size)}</div>` +
        (usage && node.path === scanRoot ? `<div class="size">${bytes(usage.free)} free</div>` : '') +
        (up ? '<div class="up">click ring centre ↑</div>' : '');
    const hit = document.getElementById('hit') || document.createElementNS('http://www.w3.org/2000/svg', 'circle');
    hit.id = 'hit';
    hit.setAttribute('r', hole);
    hit.setAttribute('fill', 'transparent');
    hit.style.cursor = up ? 'pointer' : 'default';
    hit.onclick = goUp;
    svg.appendChild(hit);
}

function drawSpace() {
    const el = $('space');
    if (!usage || !usage.total) { el.hidden = true; return; }
    const pct = ((usage.total - usage.free) / usage.total) * 100;
    el.hidden = false;
    el.innerHTML = `${bytes(usage.free)} free of ${bytes(usage.total)}<div class="meter"><i class="${pct > 90 ? 'full' : ''}" style="width:${pct}%"></i></div>`;
}

// ---- list ---------------------------------------------------------------------------------------------------------

function drawList(node) {
    const kids = node.children || [];
    $('list').innerHTML = kids.map((c, i) => {
        const pct = node.size ? (c.size / node.size) * 100 : 0;
        const hue = c.path ? (i * 137.5 + 210) % 360 : 215;
        const inBasket = basket.has(c.path);
        return `<div class="row${inBasket ? ' in-basket' : ''}" data-i="${i}"><span class="n">${c.isDir ? '▸ ' : ''}${esc(c.name)}</span><span class="s">${bytes(c.size)}</span>` +
            (c.path ? `<div class="acts"><button data-a="reveal">Reveal</button><button data-a="basket">${inBasket ? 'Remove' : '+ Basket'}</button></div>` : '') +
            `<div class="bar"><i style="width:${pct}%;background:hsl(${hue} 62% 55%)"></i></div></div>`;
    }).join('');
    $('list').onclick = (e) => {
        const row = e.target.closest('.row');
        if (!row) return;
        const c = kids[+row.dataset.i];
        if (!c.path) return;
        const act = e.target.dataset.a;
        if (act === 'reveal') Reveal(c.path);
        else if (act === 'basket') toggleBasket(c);
        else if (c.isDir) open(c.path);
        else Reveal(c.path);
    };
}

// ---- basket and delete-to-Trash -------------------------------------------------------------------------------------

function toggleBasket(n) {
    if (basket.has(n.path)) {
        basket.delete(n.path);
    } else if (![...basket.keys()].some((p) => isUnder(n.path, p))) {   // a parent folder is already in
        for (const p of [...basket.keys()]) if (isUnder(p, n.path)) basket.delete(p);   // drop what it now contains
        basket.set(n.path, {name: n.name, size: n.size});
    }
    basketState = 'idle';
    basketError = '';
    drawList(current);
    drawBasket();
    drawChart(current);
}

function drawBasket() {
    const el = $('basket');
    if (!basket.size) { el.hidden = true; return; }
    el.hidden = false;
    const total = [...basket.values()].reduce((s, b) => s + b.size, 0);
    const items = [...basket].map(([p, b]) => `<div class="it"><span title="${esc(p)}">${esc(b.name)}</span><span>${bytes(b.size)} <a data-rm="${esc(p)}">×</a></span></div>`).join('');
    let btns;
    if (basketState === 'confirm') {
        btns = `<div class="note">Move ${basket.size} item${basket.size > 1 ? 's' : ''} (${bytes(total)}) to the Trash? You can still put them back from there.</div>` +
            `<div class="btns"><button class="danger" data-b="yes">Yes, move to Trash</button><button data-b="no">Cancel</button></div>`;
    } else if (basketState === 'working') {
        btns = '<div class="note">Moving to the Trash…</div>';
    } else {
        btns = `<div class="btns"><button class="danger" data-b="ask">Move to Trash</button><button data-b="clear">Clear</button></div>`;
    }
    el.innerHTML = `<div class="h"><span>Basket · ${basket.size}</span><span>${bytes(total)}</span></div>${items}${btns}` +
        (basketError ? `<div class="err">${esc(basketError)}</div>` : '');
    el.onclick = async (e) => {
        if (e.target.dataset.rm) { basket.delete(e.target.dataset.rm); basketError = ''; return refreshBasketViews(); }
        const b = e.target.dataset.b;
        if (b === 'clear') { basket.clear(); basketError = ''; refreshBasketViews(); }
        else if (b === 'ask') { basketState = 'confirm'; drawBasket(); }
        else if (b === 'no') { basketState = 'idle'; drawBasket(); }
        else if (b === 'yes') await moveToTrash();
    };
}

function refreshBasketViews() {
    basketState = 'idle';
    drawList(current);
    drawBasket();
    drawChart(current);
}

async function moveToTrash() {
    basketState = 'working';
    drawBasket();
    const res = await Trash([...basket.keys()]);
    for (const p of res.trashed || []) basket.delete(p);
    basketState = 'idle';
    basketError = (res.failed || []).join('\n');
    try { usage = await FreeSpace(scanRoot); } catch (_) { /* keep the old numbers */ }
    await open(current.path);
}

// ---- navigation -----------------------------------------------------------------------------------------------------

function drawCrumbs(path) {
    const sep = sepOf(scanRoot);
    const rel = path.slice(scanRoot.length).split(/[\\/]/).filter(Boolean);
    let acc = scanRoot;
    const parts = [`<span data-p="${esc(scanRoot)}">${esc(scanRoot)}</span>`];
    for (const r of rel) {
        acc += (acc.endsWith(sep) ? '' : sep) + r;
        parts.push(`<span data-p="${esc(acc)}">${esc(r)}</span>`);
    }
    $('crumbs').innerHTML = parts.join(' <span>/</span> ');
    $('crumbs').scrollLeft = $('crumbs').scrollWidth;   // keep the current folder in view
    $('up').disabled = path === scanRoot;
}
$('crumbs').onclick = (e) => { if (e.target.dataset.p) open(e.target.dataset.p); };

async function open(path) {
    try { render(await View(path)); } catch (err) { console.error(err); }
}

function goUp() {
    if (!current || current.path === scanRoot) return;
    const sep = sepOf(current.path);
    const cut = current.path.lastIndexOf(sep);
    open(cut <= 0 ? sep : current.path.slice(0, cut));
}

function showTip(e, n, total) {
    const t = $('tip');
    t.hidden = false;
    t.innerHTML = `<div><b>${esc(n.name)}</b></div><div>${bytes(n.size)} · ${total ? ((n.size / total) * 100).toFixed(1) : 0}%` +
        (n.isDir ? ` · ${n.items.toLocaleString()} files` : '') + `</div>` + (n.path ? `<div class="p">${esc(n.path)}</div>` : '') +
        (n.path ? '<div class="p">shift-click: add to basket</div>' : '');
    t.style.left = Math.min(e.clientX + 14, innerWidth - 380) + 'px';
    t.style.top = Math.min(e.clientY + 14, innerHeight - 110) + 'px';
}
function hideTip() { $('tip').hidden = true; }

async function scan(path) {
    if (!path || busy) return;
    busy = true;
    scanRoot = path;
    basket.clear();
    basketError = '';
    $('sun').innerHTML = '';
    $('splash').hidden = false;
    $('splash').onclick = null;
    $('splash').innerHTML = `<div>Scanning ${esc(path)}…<br><span id="count">0</span> files<br><br><button id="stop">Stop</button></div>`;
    $('stop').onclick = () => Cancel();
    try {
        volumes = (await Volumes()) || [];
        usage = await FreeSpace(path).catch(() => null);
        render(await Scan(path));
        $('rescan').disabled = false;
    } catch (err) {
        $('splash').innerHTML = `<div>${esc(String(err))}<br><br><button id="back">Back</button></div>`;
        $('back').onclick = showStart;
    } finally {
        busy = false;
    }
}

EventsOn('progress', (n) => { const c = $('count'); if (c) c.textContent = n.toLocaleString(); });
$('disks').onclick = showStart;
$('choose').onclick = async () => { const p = await ChooseFolder(); if (p) scan(p); };
$('up').onclick = goUp;
$('rescan').onclick = () => scan(scanRoot);
addEventListener('resize', () => { if (current) drawChart(current); });
showStart();
