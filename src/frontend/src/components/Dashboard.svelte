<script>
  import ActionCard from './ActionCard.svelte'

  let actions = [
    {
      title: 'Desktop Cleanup',
      description: 'Clear and restore the desktop folder from a zip archive',
      icon: '🗂',
      dangerous: false,
      action: () => window.go.main.App.RunDesktopCleanup(),
    },
    {
      title: 'Chrome Extension',
      description: 'Force-install the Chrome extension via registry policy',
      icon: '🧩',
      dangerous: false,
      action: () => window.go.main.App.RunChromeInstall(),
    },
    {
      title: 'System Repair',
      description: 'Run DISM, SFC, and other system repair utilities',
      icon: '🔧',
      dangerous: true,
      action: () => window.go.main.App.RunSystemRepair(),
    },
    {
      title: 'Policy Settings',
      description: 'Apply group policy and registry settings',
      icon: '📋',
      dangerous: true,
      action: () => window.go.main.App.ApplyPolicySettings(),
    },
  ]
</script>

<div class="dashboard">
  <header class="header">
    <h1>bots-util</h1>
    <span class="subtitle">Desktop Utility</span>
  </header>
  <div class="grid">
    {#each actions as action}
      <ActionCard
        title={action.title}
        description={action.description}
        icon={action.icon}
        dangerous={action.dangerous}
        action={action.action}
      />
    {/each}
  </div>
</div>

<style>
  .dashboard {
    max-width: 900px;
    margin: 0 auto;
  }
  .header {
    text-align: center;
    margin-bottom: 32px;
  }
  .header h1 {
    font-size: 28px;
    font-weight: 700;
    color: #e0e0e0;
  }
  .subtitle {
    font-size: 14px;
    color: #8892b0;
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(380px, 1fr));
    gap: 20px;
  }
</style>