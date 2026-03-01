<script lang="ts">
  import { timerStore, formatDuration } from '../stores/timer';
  
  export let onStop: () => void;
  
  $: displayTime = formatDuration($timerStore.elapsedSeconds);
</script>

<div class="timer-widget flex items-center gap-4">
  {#if $timerStore.isRunning}
    <div class="flex items-center gap-4 bg-[var(--color-bg-surface)] border border-[var(--color-accent)]/30 px-6 py-2.5 rounded-2xl shadow-glow-orange">
      <!-- Pulsing indicator -->
      <div class="relative">
        <div class="w-3 h-3 bg-[var(--color-accent)] rounded-full animate-pulse"></div>
        <div class="absolute inset-0 w-3 h-3 bg-[var(--color-accent)] rounded-full animate-ping opacity-75"></div>
      </div>
      
      <div class="text-3xl font-mono font-bold text-[var(--color-accent)] tracking-wider">
        {displayTime}
      </div>
      
      <div class="flex flex-col min-w-[120px]">
        <span class="text-sm font-bold text-[var(--color-accent)]">{$timerStore.issueKey}</span>
        <span class="text-xs text-[var(--color-text-muted)] truncate max-w-[180px]">
          {$timerStore.summary || 'No summary'}
        </span>
      </div>
      
      <button
        on:click={onStop}
        class="ml-2 px-5 py-2 bg-gradient-to-r from-red-500 to-red-600 hover:from-red-600 hover:to-red-700 text-white rounded-xl font-semibold transition-all duration-200 shadow-lg hover:shadow-red-500/30 flex items-center gap-2 transform hover:scale-105"
      >
        <span>⏹</span>
        <span>Stop</span>
      </button>
    </div>
  {:else}
    <div class="flex items-center gap-3 px-6 py-3 bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-2xl">
      <span class="text-2xl opacity-60 grayscale">⏱️</span>
      <span class="text-sm font-medium text-[var(--color-text-muted)]">Ready to track</span>
    </div>
  {/if}
</div>
