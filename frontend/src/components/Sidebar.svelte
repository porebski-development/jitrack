<script lang="ts">
  import { onMount } from 'svelte';
  
  export let currentView: string;
  export let onNavigate: (view: string) => void;
  
  let version = 'v2.0';
  
  const menuItems = [
    { id: 'dashboard', label: 'Dashboard', icon: '📊' },
    { id: 'tasks', label: 'Tasks', icon: '📋' },
    { id: 'reports', label: 'Reports', icon: '📈' },
    { id: 'settings', label: 'Settings', icon: '⚙️' },
  ];
  
  onMount(async () => {
    try {
      // Dynamic import to handle missing binding
      const app: any = await import('../../wailsjs/go/main/App');
      // @ts-ignore
      if (app.GetVersion) {
        // @ts-ignore
        const v = await app.GetVersion();
        if (v && v !== 'dev') {
          version = v;
        }
      }
    } catch (e) {
      // Use default version if GetVersion is not available
      console.log('GetVersion not available, using default');
    }
  });
</script>

<aside class="w-[220px] bg-[var(--color-bg-secondary)] border-r border-[var(--color-border-strong)] flex flex-col">
  <!-- Navigation -->
  <nav class="flex-1 py-4 px-3 space-y-1">
    {#each menuItems as item}
      <button
        on:click={() => onNavigate(item.id)}
        class="w-full flex items-center gap-3 px-4 py-3 text-left rounded-xl transition-all duration-200 group
          {currentView === item.id 
            ? 'bg-gradient-to-r from-[var(--color-accent)] to-[var(--color-accent-dark)] text-white shadow-lg shadow-[var(--color-accent)]/30' 
            : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-surface)] hover:text-[var(--color-accent)]'}"
      >
        <span class="text-xl transition-transform duration-200 group-hover:scale-110">
          {item.icon}
        </span>
        <span class="font-medium">{item.label}</span>
        
        {#if currentView === item.id}
          <div class="ml-auto w-1.5 h-1.5 rounded-full bg-white animate-pulse"></div>
        {/if}
      </button>
    {/each}
  </nav>
  
  <!-- Footer info -->
  <div class="p-4 border-t border-[var(--color-border-strong)]">
    <div class="text-xs text-[var(--color-text-muted)] text-center">
      <span class="text-[var(--color-accent)] font-semibold">JITRACK</span> {version}
    </div>
  </div>
</aside>
