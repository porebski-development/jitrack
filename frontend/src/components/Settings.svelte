<script lang="ts">
  import { themeStore, setTheme } from '../stores/theme';
  import { GetServerPort, GetAuthToken, GetDeveloperInfo, OpenFolder } from '../../wailsjs/go/main/App';
  import { BrowserOpenURL } from '../../wailsjs/runtime';
  
  let serverPort = 0;
  let authToken = '';
  let devInfo: any = null;
  let showCopied = false;
  
  $: dbLocation = devInfo?.database_path || 'Loading...';
  $: logsLocation = devInfo?.logs_path || 'Loading...';
  
  async function loadSettings() {
    serverPort = await GetServerPort();
    authToken = await GetAuthToken();
    devInfo = await GetDeveloperInfo();
  }
  
  loadSettings();
  
  async function openFolder(path: string) {
    await OpenFolder(path);
  }
  
  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
    showCopied = true;
    setTimeout(() => showCopied = false, 2000);
  }
  
  function openExternal(url: string) {
    BrowserOpenURL(url);
  }
</script>

<div class="space-y-8 max-w-3xl">
  <!-- Header -->
  <div class="flex items-center gap-4">
    <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-[var(--color-accent)] to-[var(--color-accent-dark)] flex items-center justify-center shadow-glow-orange">
      <span class="text-2xl">⚙️</span>
    </div>
    <div>
      <h1 class="text-3xl font-bold text-[var(--color-text-primary)]">Settings</h1>
      <p class="text-[var(--color-text-muted)] text-sm">Configure your JITRACK preferences</p>
    </div>
  </div>
  
  <!-- Theme Section -->
  <section class="bg-[var(--color-bg-card)] rounded-2xl p-6 border border-[var(--color-border)] shadow-sm">
    <div class="flex items-center gap-3 mb-6">
      <div class="w-10 h-10 rounded-xl bg-[var(--color-accent)]/20 flex items-center justify-center">
        <span class="text-xl">🎨</span>
      </div>
      <div>
        <h2 class="text-xl font-bold text-[var(--color-text-primary)]">Appearance</h2>
        <p class="text-[var(--color-text-muted)] text-sm">Choose your preferred theme</p>
      </div>
    </div>
    
    <div class="grid grid-cols-2 gap-4">
      <button
        on:click={() => setTheme('dark')}
        class="relative p-6 rounded-xl border-2 transition-all duration-200 text-left group
          {$themeStore === 'dark' 
            ? 'border-[var(--color-accent)] bg-[var(--color-accent)]/10 shadow-glow-orange' 
            : 'border-[var(--color-border)] hover:border-[var(--color-accent)]/50 hover:bg-[var(--color-bg-surface)]'}"
      >
        <div class="text-4xl mb-3">🌙</div>
        <div class="font-semibold text-[var(--color-text-primary)]">Dark Mode</div>
        <div class="text-sm text-[var(--color-text-muted)]">Easier on the eyes</div>
        {#if $themeStore === 'dark'}
          <div class="absolute top-4 right-4 w-6 h-6 rounded-full bg-[var(--color-accent)] flex items-center justify-center">
            <span class="text-white text-sm">✓</span>
          </div>
        {/if}
      </button>
      
      <button
        on:click={() => setTheme('light')}
        class="relative p-6 rounded-xl border-2 transition-all duration-200 text-left group
          {$themeStore === 'light' 
            ? 'border-[var(--color-accent)] bg-[var(--color-accent-soft)] shadow-glow-orange' 
            : 'border-[var(--color-border)] hover:border-[var(--color-accent)]/50 hover:bg-[var(--color-bg-surface)]'}"
      >
        <div class="text-4xl mb-3">☀️</div>
        <div class="font-semibold text-[var(--color-text-primary)]">Light Mode</div>
        <div class="text-sm text-[var(--color-text-muted)]">Bright and clean</div>
        {#if $themeStore === 'light'}
          <div class="absolute top-4 right-4 w-6 h-6 rounded-full bg-[var(--color-accent)] flex items-center justify-center">
            <span class="text-white text-sm">✓</span>
          </div>
        {/if}
      </button>
    </div>
  </section>
  
  <!-- Chrome Extension Section -->
  <section class="bg-[var(--color-bg-card)] rounded-2xl p-6 border border-[var(--color-border)] shadow-sm">
    <div class="flex items-center gap-3 mb-6">
      <div class="w-10 h-10 rounded-xl bg-[var(--color-accent)]/20 flex items-center justify-center">
        <span class="text-xl">🔌</span>
      </div>
      <div>
        <h2 class="text-xl font-bold text-[var(--color-text-primary)]">Chrome Extension</h2>
        <p class="text-[var(--color-text-muted)] text-sm">Connection details</p>
      </div>
    </div>
    
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="bg-[var(--color-bg-surface)] rounded-xl p-4 border border-[var(--color-border)]">
        <div class="text-xs font-bold text-[var(--color-accent)] uppercase tracking-wider mb-1">Server Port</div>
        <div class="text-2xl font-mono font-bold text-[var(--color-text-primary)]">{serverPort || 'Not running'}</div>
      </div>
      
      <div class="bg-[var(--color-bg-surface)] rounded-xl p-4 border border-[var(--color-border)]">
        <div class="text-xs font-bold text-[var(--color-accent)] uppercase tracking-wider mb-1">Auth Token</div>
        <code class="text-sm font-mono text-[var(--color-text-secondary)] break-all">{authToken ? authToken.substring(0, 16) + '...' : 'Not available'}</code>
      </div>
    </div>
  </section>
  
  <!-- Developer Tools Section - NEW! -->
  <section class="bg-[var(--color-bg-card)] rounded-2xl p-6 border border-[var(--color-border)] shadow-sm">
    <div class="flex items-center gap-3 mb-6">
      <div class="w-10 h-10 rounded-xl bg-purple-500/20 flex items-center justify-center">
        <span class="text-xl">🛠️</span>
      </div>
      <div>
        <h2 class="text-xl font-bold text-[var(--color-text-primary)]">Developer Tools</h2>
        <p class="text-[var(--color-text-muted)] text-sm">Advanced settings & debugging</p>
      </div>
    </div>
    
    <div class="space-y-4">
      <!-- Database Path -->
      <div class="bg-[var(--color-bg-surface)] rounded-xl p-4 border border-[var(--color-border)]">
        <div class="flex items-center justify-between mb-2">
          <div class="text-xs font-bold text-purple-400 uppercase tracking-wider">Database Location</div>
          <button
            on:click={() => openFolder(dbLocation)}
            class="text-xs px-3 py-1.5 bg-[var(--color-accent)]/20 hover:bg-[var(--color-accent)] text-[var(--color-accent)] hover:text-white rounded-lg transition-all"
          >
            📁 Open Folder
          </button>
        </div>
        <code class="block text-sm font-mono text-[var(--color-text-secondary)] break-all bg-[var(--color-bg-primary)] p-3 rounded-lg border border-[var(--color-border)]">
          {dbLocation}
        </code>
      </div>
      
      <!-- Logs Path -->
      <div class="bg-[var(--color-bg-surface)] rounded-xl p-4 border border-[var(--color-border)]">
        <div class="flex items-center justify-between mb-2">
          <div class="text-xs font-bold text-purple-400 uppercase tracking-wider">Application Logs</div>
          <div class="flex gap-2">
            <button
              on:click={() => copyToClipboard(logsLocation)}
              class="text-xs px-3 py-1.5 bg-[var(--color-bg-card)] hover:bg-[var(--color-border)] text-[var(--color-text-secondary)] rounded-lg transition-all border border-[var(--color-border)]"
            >
              {showCopied ? '✓ Copied!' : '📋 Copy Path'}
            </button>
            <button
              on:click={() => openFolder(logsLocation)}
              class="text-xs px-3 py-1.5 bg-[var(--color-accent)]/20 hover:bg-[var(--color-accent)] text-[var(--color-accent)] hover:text-white rounded-lg transition-all"
            >
              📁 Open Folder
            </button>
          </div>
        </div>
        <code class="block text-sm font-mono text-[var(--color-text-secondary)] break-all bg-[var(--color-bg-primary)] p-3 rounded-lg border border-[var(--color-border)]">
          {logsLocation}
        </code>
      </div>
      
      <!-- Version Info -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="bg-[var(--color-bg-surface)] rounded-xl p-4 border border-[var(--color-border)] text-center">
          <div class="text-2xl mb-1">📦</div>
          <div class="text-xs text-[var(--color-text-muted)] uppercase">App Version</div>
          <div class="font-mono font-bold text-[var(--color-text-primary)]">{devInfo?.app_version || '-'}</div>
        </div>
        <div class="bg-[var(--color-bg-surface)] rounded-xl p-4 border border-[var(--color-border)] text-center">
          <div class="text-2xl mb-1">🐹</div>
          <div class="text-xs text-[var(--color-text-muted)] uppercase">Go Version</div>
          <div class="font-mono font-bold text-[var(--color-text-primary)]">{devInfo?.go_version || '-'}</div>
        </div>
        <div class="bg-[var(--color-bg-surface)] rounded-xl p-4 border border-[var(--color-border)] text-center">
          <div class="text-2xl mb-1">⚡</div>
          <div class="text-xs text-[var(--color-text-muted)] uppercase">Wails Version</div>
          <div class="font-mono font-bold text-[var(--color-text-primary)]">{devInfo?.wails_version || '-'}</div>
        </div>
        <div class="bg-[var(--color-bg-surface)] rounded-xl p-4 border border-[var(--color-border)] text-center">
          <div class="text-2xl mb-1">💻</div>
          <div class="text-xs text-[var(--color-text-muted)] uppercase">Platform</div>
          <div class="font-mono font-bold text-[var(--color-text-primary)]">{devInfo?.platform || '-'}</div>
        </div>
      </div>
      
      <!-- Quick Actions -->
      <div class="flex flex-wrap gap-3 pt-4 border-t border-[var(--color-border)]">
        <button
          on:click={() => openExternal('https://wails.io/docs/introduction')}
          class="px-4 py-2 bg-[var(--color-bg-surface)] hover:bg-[var(--color-border)] text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] rounded-lg transition-all text-sm border border-[var(--color-border)] flex items-center gap-2"
        >
          <span>📖</span>
          <span>Wails Docs</span>
        </button>
        <button
          on:click={() => openExternal('https://github.com/wailsapp/wails')}
          class="px-4 py-2 bg-[var(--color-bg-surface)] hover:bg-[var(--color-border)] text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] rounded-lg transition-all text-sm border border-[var(--color-border)] flex items-center gap-2"
        >
          <span>🐙</span>
          <span>GitHub</span>
        </button>
      </div>
    </div>
  </section>
  
  <!-- About Section -->
  <section class="bg-[var(--color-bg-card)] rounded-2xl p-6 border border-[var(--color-border)] shadow-sm">
    <div class="flex items-center gap-3 mb-6">
      <div class="w-10 h-10 rounded-xl bg-[var(--color-accent)]/20 flex items-center justify-center">
        <span class="text-xl">ℹ️</span>
      </div>
      <div>
        <h2 class="text-xl font-bold text-[var(--color-text-primary)]">About</h2>
        <p class="text-[var(--color-text-muted)] text-sm">Application information</p>
      </div>
    </div>
    
    <div class="mt-6 p-4 bg-[var(--color-accent-soft)] rounded-xl border border-[var(--color-accent)]/20">
      <p class="text-[var(--color-text-secondary)] text-center">
        <span class="text-[var(--color-accent)] font-semibold">JITRACK</span> — JIRA Time Tracker
        <br>
        <span class="text-sm text-[var(--color-text-muted)]">Built with ❤️ for developers</span>
      </p>
    </div>
  </section>
</div>
