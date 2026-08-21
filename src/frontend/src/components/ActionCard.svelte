<script>
  import { onMount, onDestroy } from 'svelte'
  import ProgressLog from './ProgressLog.svelte'

  export let title = ''
  export let description = ''
  export let icon = '⚙'
  export let dangerous = false
  export let action = null
  export let eventName = ''

  let state = 'idle' // idle | running | success | error | cancelled
  let logLines = []
  let percent = 0
  let errorMessage = ''
  let off = null

  function addLog(msg) {
    logLines = [...logLines, msg]
  }

  onMount(() => {
    if (window.runtime && window.runtime.EventsOn) {
      off = window.runtime.EventsOn('progress', (data) => {
        if (data.action === eventName) {
          addLog(data.msg)
          percent = Math.round(data.pct * 100)
        }
      })
    }
  })

  onDestroy(() => {
    if (off && typeof off === 'function') off()
  })

  async function handleRun() {
    state = 'running'
    logLines = []
    percent = 0
    errorMessage = ''

    try {
      await action()
      state = 'success'
      percent = 100
    } catch (e) {
      state = 'error'
      errorMessage = e.message || String(e)
      addLog('ERROR: ' + errorMessage)
    }
  }

  function handleCancel() {
    state = 'cancelled'
    addLog('Cancelled by user')
  }

  $: statusClass = state === 'running' ? 'running' :
                   state === 'success' ? 'success' :
                   state === 'error' ? 'error' :
                   state === 'cancelled' ? 'cancelled' : ''
</script>

<div class="card {statusClass}">
  <div class="card-header">
    <span class="icon">{icon}</span>
    <div class="header-text">
      <h3>{title}</h3>
      <p>{description}</p>
    </div>
  </div>

  {#if state === 'running'}
    <div class="progress-bar">
      <div class="progress-fill" style="width: {percent}%"></div>
    </div>
  {/if}

  {#if logLines.length > 0}
    <ProgressLog lines={logLines} />
  {/if}

  {#if state === 'error'}
    <div class="error-banner">{errorMessage}</div>
  {/if}

  <div class="card-actions">
    {#if state === 'idle' || state === 'success' || state === 'error' || state === 'cancelled'}
      <button
        class="btn {dangerous ? 'btn-danger' : 'btn-primary'}"
        on:click={handleRun}
      >
        {state === 'idle' ? 'Run' : 'Run Again'}
      </button>
    {/if}
    {#if state === 'running'}
      <button class="btn btn-cancel" on:click={handleCancel}>
        Cancel
      </button>
    {/if}
  </div>
</div>

<style>
  .card {
    background: #16213e;
    border: 1px solid #0f3460;
    border-radius: 10px;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    transition: border-color 0.2s;
  }
  .card.running { border-color: #e6a817; }
  .card.success { border-color: #2ecc71; }
  .card.error { border-color: #e74c3c; }
  .card.cancelled { border-color: #7f8c8d; }

  .card-header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
  }
  .icon { font-size: 24px; }
  .header-text h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }
  .header-text p {
    margin: 4px 0 0;
    font-size: 13px;
    color: #8892b0;
  }

  .progress-bar {
    background: #0d0d1a;
    border-radius: 4px;
    height: 6px;
    overflow: hidden;
  }
  .progress-fill {
    background: #e6a817;
    height: 100%;
    transition: width 0.3s;
  }

  .card-actions {
    display: flex;
    gap: 8px;
  }
  .btn {
    padding: 8px 20px;
    border: none;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
  }
  .btn-primary { background: #0f3460; color: #e0e0e0; }
  .btn-primary:hover { background: #1a4a7a; }
  .btn-danger { background: #c0392b; color: #e0e0e0; }
  .btn-danger:hover { background: #e74c3c; }
  .btn-cancel { background: #7f8c8d; color: #e0e0e0; }
  .btn-cancel:hover { background: #95a5a6; }

  .error-banner {
    background: rgba(231, 76, 60, 0.15);
    color: #e74c3c;
    padding: 8px 12px;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
  }
</style>