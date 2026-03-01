<script lang="ts">
  import { themeStore, toggleTheme } from '../stores/theme';
  import TimerWidget from './TimerWidget.svelte';
  
  export let onStopTimer: () => void;
</script>

<header class="h-[60px] bg-[var(--color-bg-secondary)] border-b border-[var(--color-border-strong)] flex items-center justify-between px-6 shadow-sm">
  <!-- Logo -->
  <div class="flex items-center gap-3">
    <img src="/logo.png" alt="JITRACK" class="w-10 h-10 rounded-xl shadow-glow-orange object-contain" />
    <span class="text-2xl font-bold text-[var(--color-text-primary)]">
      JITRACK
    </span>
  </div>
  
  <!-- Timer Widget (centered) -->
  <div class="flex-1 flex justify-center">
    <TimerWidget onStop={onStopTimer} />
  </div>
  
  <!-- Right controls -->
  <div class="flex items-center gap-3">
    <!-- Theme toggle -->
    <button
      on:click={toggleTheme}
      class="px-4 py-2 rounded-xl bg-[var(--color-bg-surface)] hover:bg-[var(--color-accent)]/10 border border-[var(--color-border)] hover:border-[var(--color-accent)]/50 transition-all duration-200 text-sm font-medium flex items-center gap-2"
      title="Toggle theme"
    >
      <span class="block transform transition-transform hover:scale-110">
        {$themeStore === 'dark' ? '☀️' : '🌙'}
      </span>
      <span>{$themeStore === 'dark' ? 'Timi mode' : 'Dark mode'}</span>
    </button>
  </div>
</header>
