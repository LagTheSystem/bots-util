<script>
  export let lines = []

  let container

  $: if (container) {
    container.scrollTop = container.scrollHeight
  }
</script>

<div class="progress-log" bind:this={container}>
  {#each lines as line}
    <div class="log-line">{line}</div>
  {/each}
</div>

<style>
  .progress-log {
    background: #0d0d1a;
    border-radius: 6px;
    padding: 10px 14px;
    max-height: 200px;
    overflow-y: auto;
    font-family: 'Consolas', 'Courier New', monospace;
    font-size: 12px;
    line-height: 1.6;
  }
  .log-line {
    color: #8be9fd;
    white-space: pre-wrap;
    word-break: break-all;
  }
</style>