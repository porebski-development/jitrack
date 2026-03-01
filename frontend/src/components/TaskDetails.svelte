<script lang="ts">
  import { onMount } from 'svelte';
  import { GetWorklogs, GetTask } from '../../wailsjs/go/main/App';
  import { formatTimeShort } from '../stores/timer';
  import { BrowserOpenURL } from '../../wailsjs/runtime';
  
  export let task: any;
  export let onClose: () => void;
  export let onStart: () => void;
  
  let worklogs: any[] = [];
  let fullTask: any = null;
  
  onMount(async () => {
    fullTask = await GetTask(task.issue_key);
    worklogs = await GetWorklogs(task.issue_key) || [];
  });
  
  function openURL(url: string) {
    if (url) {
      BrowserOpenURL(url);
    }
  }
  
  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleString();
  }
</script>

<div class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4" on:click={onClose}>
  <div class="bg-[var(--color-bg-card)] rounded-2xl border border-[var(--color-accent)]/30 w-full max-w-2xl max-h-[85vh] overflow-hidden shadow-2xl" on:click|stopPropagation>
    <!-- Header -->
    <div class="flex items-center justify-between p-6 border-b border-[var(--color-border)] bg-gradient-to-r from-[var(--color-bg-surface)] to-transparent">
      <div>
        <div class="flex items-center gap-3 mb-1">
          <span class="text-2xl font-bold text-[var(--color-accent)] font-mono">{task.issue_key}</span>
          <span class="px-3 py-1 bg-[var(--color-accent-soft)] text-[var(--color-accent)] text-xs rounded-full font-semibold">{task.project}</span>
        </div>
        <p class="text-[var(--color-text-secondary)] text-sm">{task.summary}</p>
      </div>
      <button 
        on:click={onClose} 
        class="w-10 h-10 rounded-xl bg-[var(--color-bg-surface)] hover:bg-[var(--color-accent)]/20 text-[var(--color-text-muted)] hover:text-[var(--color-accent)] transition-all flex items-center justify-center text-2xl border border-[var(--color-border)]"
      >
        ×
      </button>
    </div>
    
    <div class="p-6 space-y-6 overflow-y-auto max-h-[60vh]">
      <!-- Summary -->
      <div class="bg-[var(--color-bg-surface)] rounded-xl p-4 border border-[var(--color-border)]">
        <label class="text-xs font-bold text-[var(--color-accent)] uppercase tracking-wider block mb-2">Summary</label>
        <p class="text-[var(--color-text-primary)] text-lg font-medium">{task.summary}</p>
      </div>
      
      <!-- URL -->
      {#if task.url}
        <div class="bg-[var(--color-bg-surface)] rounded-xl p-4 border border-[var(--color-border)]">
          <label class="text-xs font-bold text-[var(--color-accent)] uppercase tracking-wider block mb-2">JIRA URL</label>
          <button
            on:click={() => openURL(task.url)}
            class="text-[var(--color-accent)] hover:text-[var(--color-accent-dark)] hover:underline flex items-center gap-2 transition-colors text-left break-all"
          >
            <span>🔗</span>
            <span>{task.url}</span>
          </button>
        </div>
      {/if}
      
      <!-- Estimated Hours -->
      <div class="bg-[var(--color-bg-surface)] rounded-xl p-4 border border-[var(--color-border)]">
        <label class="text-xs font-bold text-[var(--color-accent)] uppercase tracking-wider block mb-2">Estimated Hours</label>
        <div class="text-3xl font-bold text-[var(--color-text-primary)]">{task.estimated_hours || 'Not set'}<span class="text-lg text-[var(--color-text-muted)] ml-1">h</span></div>
      </div>
      
      <!-- Worklog History -->
      <div>
        <label class="text-xs font-bold text-[var(--color-accent)] uppercase tracking-wider block mb-4">Worklog History</label>
        {#if worklogs.length === 0}
          <div class="text-center py-8 bg-[var(--color-bg-surface)] rounded-xl border border-[var(--color-border)] border-dashed">
            <div class="text-4xl mb-2 opacity-30">⏱️</div>
            <p class="text-[var(--color-text-muted)]">No time logged yet</p>
          </div>
        {:else}
          <div class="space-y-3">
            {#each worklogs as worklog}
              <div class="bg-[var(--color-bg-surface)] rounded-xl p-4 flex items-center justify-between border border-[var(--color-border)] hover:border-[var(--color-accent)]/30 transition-colors">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-xl bg-[var(--color-accent)]/20 flex items-center justify-center">
                    <span class="text-[var(--color-accent)]">⏱️</span>
                  </div>
                  <div>
                    <div class="text-sm text-[var(--color-text-secondary)]">
                      {formatDate(worklog.started_at)}
                    </div>
                    {#if worklog.notes}
                      <div class="text-xs text-[var(--color-text-muted)] mt-1">{worklog.notes}</div>
                    {/if}
                  </div>
                </div>
                <div class="text-xl font-mono font-bold text-[var(--color-accent)]">
                  {formatTimeShort(worklog.duration_seconds || 0)}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
    
    <!-- Actions -->
    <div class="flex gap-4 p-6 border-t border-[var(--color-border)] bg-[var(--color-bg-surface)]">
      <button
        on:click={onStart}
        class="flex-1 px-6 py-3 bg-gradient-to-r from-[var(--color-accent)] to-[var(--color-accent-dark)] hover:opacity-90 text-white font-bold rounded-xl transition-all duration-200 shadow-lg flex items-center justify-center gap-2"
      >
        <span>▶</span>
        <span>Start Tracking</span>
      </button>
      <button
        on:click={onClose}
        class="px-6 py-3 bg-[var(--color-bg-card)] hover:bg-[var(--color-border)] text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] rounded-xl transition-all duration-200 border border-[var(--color-border)] font-medium"
      >
        Close
      </button>
    </div>
  </div>
</div>
