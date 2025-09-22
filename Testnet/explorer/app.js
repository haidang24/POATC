const $ = (s) => document.querySelector(s);
const fmt = (v) => v?.toString();

let RPC = localStorage.getItem('rpc') || 'http://localhost:8547';

async function rpc(method, params = []) {
  const res = await fetch(RPC, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ jsonrpc: '2.0', id: 1, method, params })
  });
  const json = await res.json();
  if (json.error) throw new Error(json.error.message);
  return json.result;
}

function shortHash(h, n = 10) { return h ? h.slice(0, 2 + n) + '…' + h.slice(-n) : ''; }
function toInt(hex) { return parseInt(hex, 16); }
function weiToEth(hex) { return (toInt(hex) / 1e18).toFixed(6); }

async function loadLatest() {
  try {
    const [latestHex, chainId, peers, gasPrice] = await Promise.all([
      rpc('eth_blockNumber'), rpc('eth_chainId'), rpc('net_peerCount'), rpc('eth_gasPrice')
    ]);
    if (document.getElementById('statBlock')) {
      document.getElementById('statBlock').textContent = parseInt(latestHex, 16);
      document.getElementById('statChain').textContent = parseInt(chainId, 16);
      document.getElementById('statPeers').textContent = parseInt(peers, 16);
      document.getElementById('statGas').textContent = (parseInt(gasPrice, 16)/1e9).toFixed(3) + ' Gwei';
    }
    const latest = toInt(latestHex);
    const blocks = [];
    for (let i = 0; i < 12; i++) {
      const num = '0x' + (latest - i).toString(16);
      const b = await rpc('eth_getBlockByNumber', [num, true]);
      if (b) blocks.push(b);
    }
    renderBlocks(blocks);
    renderTxs(blocks.flatMap(b => b.transactions).slice(0, 12));
  } catch (e) {
    $('#blocksList').innerHTML = `<div class="rowi">${e.message}</div>`;
  }
}

function renderBlocks(blocks) {
  const html = blocks.map(b => `
    <div class="rowi">
      <div class="mono">#${toInt(b.number)}</div>
      <div>
        <div>Hash: <a href="#/block/${b.hash}" class="link mono">${shortHash(b.hash)}</a></div>
        <div class="muted mono">Miner: ${b.miner}</div>
      </div>
      <div class="pill">${b.transactions.length} txs</div>
    </div>
  `).join('');
  $('#blocksList').innerHTML = html;
}

function renderTxs(txs) {
  const html = txs.map(t => `
    <div class="rowi tx">
      <div class="mono">${shortHash(t.hash)}</div>
      <div>
        <div class="mono">${t.from} → ${t.to || '(contract)'} </div>
        <div class="muted">${(parseInt(t.value,16)/1e18).toFixed(6)} ETH</div>
      </div>
      <div class="pill">gas: ${toInt(t.gas)}</div>
    </div>
  `).join('');
  $('#txsList').innerHTML = html || '<div class="rowi">Chưa có giao dịch</div>';
}

async function showBlock(hashOrNumber) {
  let block;
  if (/^0x[0-9a-fA-F]{64}$/.test(hashOrNumber)) block = await rpc('eth_getBlockByHash', [hashOrNumber, true]);
  else block = await rpc('eth_getBlockByNumber', ['0x' + parseInt(hashOrNumber).toString(16), true]);
  if (!block) return;
  $('#detailView').style.display = 'block';
  $('#detailView').innerHTML = `
    <div class="card-header"><h3>Block #${toInt(block.number)}</h3></div>
    <div class="kv"><div>Hash</div><div class="mono">${block.hash}</div></div>
    <div class="kv"><div>Parent</div><div class="mono"><a class="link" href="#/block/${block.parentHash}">${block.parentHash}</a></div></div>
    <div class="kv"><div>Miner</div><div class="mono">${block.miner}</div></div>
    <div class="kv"><div>Txs</div><div>${block.transactions.length}</div></div>
  `;
}

async function showTx(hash) {
  const tx = await rpc('eth_getTransactionByHash', [hash]);
  if (!tx) return;
  $('#detailView').style.display = 'block';
  $('#detailView').innerHTML = `
    <div class="card-header"><h3>Transaction</h3></div>
    <div class="kv"><div>Hash</div><div class="mono">${tx.hash}</div></div>
    <div class="kv"><div>From</div><div class="mono">${tx.from}</div></div>
    <div class="kv"><div>To</div><div class="mono">${tx.to}</div></div>
    <div class="kv"><div>Value</div><div>${(parseInt(tx.value,16)/1e18).toFixed(6)} ETH</div></div>
  `;
}

async function showAddress(addr) {
  const bal = await rpc('eth_getBalance', [addr, 'latest']);
  $('#detailView').style.display = 'block';
  $('#detailView').innerHTML = `
    <div class="card-header"><h3>Address</h3></div>
    <div class="kv"><div>Address</div><div class="mono">${addr}</div></div>
    <div class="kv"><div>Balance</div><div>${weiToEth(bal)} ETH</div></div>
  `;
}

function handleRoute() {
  const h = location.hash.slice(2);
  document.querySelectorAll('.nav-link').forEach(a=>a.classList.remove('active'));
  if (!h) { showHome(); return; }
  const [type, value] = h.split('/');
  if (type === '') showHome();
  else if (type === 'validators') showValidators();
  else if (type === 'timedyn') showTimeDynamic();
  else if (type === 'block') showBlock(value);
  else if (type === 'tx') showTx(value);
  else if (type === 'address') showAddress(value);
}

function showHome(){
  $('#homeView').style.display='block';
  $('#rightPanel').style.display='block';
  $('#detailView').style.display='none';
  const v1=document.getElementById('validatorsView'); if(v1) v1.style.display='none';
  const v2=document.getElementById('timeDynView'); if(v2) v2.style.display='none';
  const tab = document.querySelector('[data-tab="home"]'); if(tab) tab.classList.add('active');
}

async function showValidators(){
  $('#homeView').style.display='none';
  $('#rightPanel').style.display='none';
  $('#detailView').style.display='none';
  document.getElementById('validatorsView').style.display='block';
  const v2=document.getElementById('timeDynView'); if(v2) v2.style.display='none';
  const tab = document.querySelector('[data-tab="validators"]'); if(tab) tab.classList.add('active');
  try {
    const stats = await rpc('poatc_getValidatorSelectionStats', []);
    document.getElementById('validatorsStats').innerHTML = `
      <div class="kv"><div>Method</div><div>${stats.config.selection_method}</div></div>
      <div class="kv"><div>Small set size</div><div>${stats.validators.small_set_size}</div></div>
      <div class="kv"><div>Active validators</div><div>${stats.validators.active}</div></div>
      <div class="kv"><div>Current set</div><div class="mono">${(stats.selection.current_set||[]).join(', ')}</div></div>
    `;
    const html = (stats.selection.current_set||[]).map(v=>`<div class="rowi"><div class="mono">Validator</div><div class="mono">${v}</div><div class="pill">active</div></div>`).join('');
    document.getElementById('validatorsTable').innerHTML = html || '<div class="rowi">Không có dữ liệu</div>';
  } catch(e) {
    document.getElementById('validatorsTable').innerHTML = `<div class="rowi">${e.message}</div>`;
  }
}

async function showTimeDynamic(){
  $('#homeView').style.display='none';
  $('#rightPanel').style.display='none';
  $('#detailView').style.display='none';
  const v1=document.getElementById('validatorsView'); if(v1) v1.style.display='none';
  document.getElementById('timeDynView').style.display='block';
  const tab = document.querySelector('[data-tab="timedyn"]'); if(tab) tab.classList.add('active');
  try {
    const s = await rpc('poatc_getTimeDynamicStats', []);
    document.getElementById('tdStats').innerHTML = `
      <div class="kv"><div>Block time</div><div>${s.dynamic_block_time.current_block_time}s</div></div>
      <div class="kv"><div>Base</div><div>${s.dynamic_block_time.base_block_time}s</div></div>
      <div class="kv"><div>Tx recent</div><div>${(s.dynamic_block_time.recent_tx_counts||[]).join(', ')}</div></div>
      <div class="kv"><div>Decay rate</div><div>${s.dynamic_reputation_decay.decay_rate}</div></div>
    `;
  } catch(e) {
    document.getElementById('tdStats').innerHTML = `<div class="rowi">${e.message}</div>`;
  }
}

$('#searchBtn').onclick = () => {
  const q = $('#searchInput').value.trim();
  if (!q) return;
  if (/^0x[0-9a-fA-F]{64}$/.test(q)) location.hash = `#/tx/${q}`;
  else if (/^0x[0-9a-fA-F]{40}$/.test(q)) location.hash = `#/address/${q}`;
  else if (/^\d+$/.test(q)) location.hash = `#/block/${q}`;
};

$('#rpcSave').onclick = () => {
  const v = $('#rpcInput').value.trim();
  if (!v) return; RPC = v; localStorage.setItem('rpc', v); $('#rpcStatus').textContent = 'Đã lưu'; loadLatest();
};

$('#refreshBtn').onclick = loadLatest;
window.addEventListener('hashchange', handleRoute);

// init
$('#rpcInput').value = RPC;
loadLatest();
handleRoute();


