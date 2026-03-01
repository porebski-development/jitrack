<script lang="ts">
  import { onMount } from 'svelte';
  import { tasksStore, tasksFilter } from '../stores/tasks';
  import { timerStore, formatDuration, formatTimeShort } from '../stores/timer';
  import { GetTasks, CompleteTask, DeleteTask, GetWorklogs, StopTimer } from '../../wailsjs/go/main/App';
  
  export let onViewTask: (task: any) => void;
  export let onStartTask: (issueKey: string) => void;
  
  let worklogTimes: Record<string, number> = {};
  let showBlockModal = false;
  let blockMessage = '';
  let showDeleteModal = false;
  let deleteIssueKey: string | null = null;
  
  $: incompleteTasks = $tasksStore.filter(t => !t.is_completed);
  $: completedTasks = $tasksStore.filter(t => t.is_completed).slice(0, 5);
  
  // Get timer state
  let timerState: any = { isRunning: false, issueKey: null };
  timerStore.subscribe(state => {
    timerState = state;
  });
  
  // Check if timer is running for a specific task
  function isTimerRunningForTask(issueKey: string) {
    return timerState.isRunning && timerState.issueKey === issueKey;
  }
  
  onMount(async () => {
    // Load worklog times for progress bars
    for (const task of $tasksStore) {
      const worklogs = await GetWorklogs(task.issue_key);
      const totalSeconds = worklogs?.reduce((sum: number, w: any) => sum + (w.duration_seconds || 0), 0) || 0;
      worklogTimes[task.issue_key] = totalSeconds;
    }
  });
  
  function showBlockDialog(message: string) {
    blockMessage = message;
    showBlockModal = true;
  }
  
  async function handleComplete(issueKey: string, event: Event) {
    event.stopPropagation();
    
    // Check if timer is running for this task
    if (isTimerRunningForTask(issueKey)) {
      showBlockDialog('Cannot complete task while timer is running. Please stop the timer first.');
      return;
    }
    
    await CompleteTask(issueKey);
    const tasks = await GetTasks();
    tasksStore.set(tasks);
  }
  
  function showDeleteDialog(issueKey: string) {
    deleteIssueKey = issueKey;
    showDeleteModal = true;
  }
  
  async function confirmDelete() {
    if (!deleteIssueKey) return;
    
    showDeleteModal = false;
    await DeleteTask(deleteIssueKey);
    const tasks = await GetTasks();
    tasksStore.set(tasks);
    deleteIssueKey = null;
  }
  
  function cancelDelete() {
    showDeleteModal = false;
    deleteIssueKey = null;
  }
  
  async function handleDelete(issueKey: string, event: Event) {
    event.stopPropagation();
    
    // Check if timer is running for this task
    if (isTimerRunningForTask(issueKey)) {
      showBlockDialog('Cannot delete task while timer is running. Please stop the timer first.');
      return;
    }
    
    showDeleteDialog(issueKey);
  }
  
  // Use formatTimeShort from timer store for consistent time display
  
  function getProgressPercent(task: any): number {
    if (!task.estimated_hours || task.estimated_hours === 0) return 0;
    const actualHours = (worklogTimes[task.issue_key] || 0) / 3600;
    const percent = (actualHours / task.estimated_hours) * 100;
    return Math.min(percent, 100);
  }
  
  function isOverEstimate(task: any): boolean {
    if (!task.estimated_hours || task.estimated_hours === 0) return false;
    const actualHours = (worklogTimes[task.issue_key] || 0) / 3600;
    return actualHours > task.estimated_hours;
  }
  
  async function handleStopAndAction(action: () => Promise<void>) {
    await StopTimer();
    timerStore.update(s => ({ ...s, isRunning: false, issueKey: null }));
    await action();
    showBlockModal = false;
  }
  
  async function handleStartTaskWithRefresh(issueKey: string, event: Event) {
    event.stopPropagation();
    await onStartTask(issueKey);
  }
</script>

<div class="space-y-8">
  <!-- Page Title -->
  <div class="flex items-center gap-4">
    <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-[var(--color-accent)] to-[var(--color-accent-dark)] flex items-center justify-center shadow-glow-orange">
      <span class="text-2xl">📊</span>
    </div>
    <div>
      <h1 class="text-3xl font-bold text-[var(--color-text-primary)]">Dashboard</h1>
      <p class="text-[var(--color-text-muted)] text-sm">Overview of your tasks and time tracking</p>
    </div>
  </div>
  
  <!-- Currently Tracking -->
  {#if $timerStore.isRunning}
    <section class="relative overflow-hidden rounded-2xl bg-gradient-to-r from-[var(--color-accent)]/20 to-[var(--color-accent-dark)]/10 p-6 border border-[var(--color-accent)]/30 shadow-glow-orange">
      <div class="absolute top-0 right-0 w-64 h-64 bg-[var(--color-accent)]/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2"></div>
      
      <h2 class="text-sm font-bold text-[var(--color-accent)] uppercase tracking-wider mb-4 flex items-center gap-2">
        <span class="w-2 h-2 rounded-full bg-[var(--color-accent)] animate-pulse"></span>
        Currently Tracking
      </h2>
      
      <div class="flex items-center justify-between relative z-10">
        <div>
          <div class="text-2xl font-bold text-[var(--color-text-primary)]">{$timerStore.issueKey}</div>
          <div class="text-[var(--color-text-secondary)]">{$timerStore.summary || 'No summary'}</div>
        </div>
        <div class="text-4xl font-mono font-bold text-[var(--color-accent)] tabular-nums">
          {formatDuration($timerStore.elapsedSeconds)}
        </div>
      </div>
    </section>
  {/if}
  
  <!-- Stats Row -->
  <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
    <div class="bg-[var(--color-bg-card)] rounded-2xl p-6 border border-[var(--color-border)] shadow-sm">
      <div class="text-[var(--color-text-muted)] text-sm font-medium mb-1">Active Tasks</div>
      <div class="text-4xl font-bold text-[var(--color-accent)]">{incompleteTasks.length}</div>
    </div>
    <div class="bg-[var(--color-bg-card)] rounded-2xl p-6 border border-[var(--color-border)] shadow-sm">
      <div class="text-[var(--color-text-muted)] text-sm font-medium mb-1">Completed Today</div>
      <div class="text-4xl font-bold text-green-500">{completedTasks.length}</div>
    </div>
    <div class="bg-[var(--color-bg-card)] rounded-2xl p-6 border border-[var(--color-border)] shadow-sm">
      <div class="text-[var(--color-text-muted)] text-sm font-medium mb-1">Total Tracked</div>
      <div class="text-4xl font-bold text-[var(--color-accent)]">
        {formatTimeShort(Object.values(worklogTimes).reduce((a, b) => a + b, 0))}
      </div>
    </div>
  </div>
  
  <!-- Active Tasks Grid -->
  <section>
    <h2 class="text-xl font-bold text-[var(--color-text-primary)] mb-6 flex items-center gap-2">
      <span class="w-1 h-6 bg-[var(--color-accent)] rounded-full"></span>
      Active Tasks
    </h2>
    
    {#if incompleteTasks.length === 0}
      <div class="text-center py-16 bg-[var(--color-bg-card)] rounded-2xl border border-[var(--color-border)] border-dashed">
        <div class="text-6xl mb-4 opacity-50">📋</div>
        <p class="text-[var(--color-text-secondary)] text-lg">No active tasks</p>
        <p class="text-[var(--color-text-muted)] text-sm mt-2">Import tasks from Chrome extension</p>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each incompleteTasks as task}
          <div 
            class="group relative bg-[var(--color-bg-card)] rounded-2xl p-5 border border-[var(--color-border)] hover:border-[var(--color-accent)]/50 transition-all duration-300 cursor-pointer hover:shadow-lg {isTimerRunningForTask(task.issue_key) ? 'ring-2 ring-[var(--color-accent)]/50 shadow-glow-orange' : ''}"
            on:click={() => onViewTask(task)}
          >
            <!-- Glow effect on hover -->
            <div class="absolute inset-0 bg-gradient-to-br from-[var(--color-accent)]/5 to-transparent rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
            
            <div class="relative z-10">
              <div class="flex items-start justify-between mb-3">
                <div class="flex items-center gap-2">
                  <span class="text-lg font-bold text-[var(--color-accent)]">{task.issue_key}</span>
                  {#if isTimerRunningForTask(task.issue_key)}
                    <span class="flex items-center gap-1 text-xs bg-[var(--color-accent)] text-white px-2 py-0.5 rounded-full font-medium">
                      <span class="w-1.5 h-1.5 bg-white rounded-full animate-pulse"></span>
                      Tracking
                    </span>
                  {/if}
                </div>
                <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-all duration-200">
                  <!-- Complete button (checkmark) -->
                  <button
                    on:click={(e) => handleComplete(task.issue_key, e)}
                    class="p-2 hover:bg-green-500/20 text-[var(--color-text-muted)] hover:text-green-500 rounded-lg transition-colors disabled:opacity-30"
                    title="Mark complete"
                    disabled={isTimerRunningForTask(task.issue_key)}
                  >
                    ☑️
                  </button>
                  <button
                    on:click={(e) => handleDelete(task.issue_key, e)}
                    class="p-2 hover:bg-red-500/20 text-[var(--color-text-muted)] hover:text-red-500 rounded-lg transition-colors disabled:opacity-30"
                    title="Delete"
                    disabled={isTimerRunningForTask(task.issue_key)}
                  >
                    🗑️
                  </button>
                </div>
              </div>
              
              <p class="text-[var(--color-text-primary)] text-sm mb-4 line-clamp-2 font-medium">{task.summary}</p>
              
              <!-- Progress Bar -->
              {#if task.estimated_hours}
                <div class="mb-3">
                  <div class="h-2.5 bg-[var(--color-bg-surface)] rounded-full overflow-hidden">
                    <div 
                      class="h-full rounded-full transition-all duration-500 {isOverEstimate(task) ? 'bg-red-500' : 'bg-gradient-to-r from-[var(--color-accent)] to-[var(--color-accent-light)]'}"
                      style="width: {getProgressPercent(task)}%"
                    ></div>
                  </div>
                </div>
              {/if}
              
              <div class="flex items-center justify-between text-xs">
                <span class="text-[var(--color-text-muted)]">{formatTimeShort(worklogTimes[task.issue_key] || 0)} / {task.estimated_hours ? formatTimeShort(task.estimated_hours * 3600) : '?'}</span>
                {#if isTimerRunningForTask(task.issue_key)}
                  <span class="text-[var(--color-accent)] font-semibold animate-pulse">Tracking...</span>
                {:else}
                  <button
                    on:click={(e) => handleStartTaskWithRefresh(task.issue_key, e)}
                    class="px-4 py-1.5 bg-[var(--color-accent)]/10 hover:bg-[var(--color-accent)] text-[var(--color-accent)] hover:text-white rounded-lg font-semibold transition-all duration-200 text-xs"
                  >
                    ▶ Start
                  </button>
                {/if}
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </section>
  
  <!-- Recently Completed -->
  {#if completedTasks.length > 0}
    <section>
      <h2 class="text-xl font-bold text-[var(--color-text-primary)] mb-6 flex items-center gap-2">
        <span class="w-1 h-6 bg-green-500 rounded-full"></span>
        Recently Completed
      </h2>
      <div class="space-y-3">
        {#each completedTasks as task}
          <div class="bg-[var(--color-bg-card)] rounded-xl p-4 border border-[var(--color-border)] flex items-center justify-between opacity-70 hover:opacity-100 transition-opacity">
            <div class="flex items-center gap-4">
              <div class="w-8 h-8 rounded-full bg-green-500/20 flex items-center justify-center">
                <span class="text-green-500">✓</span>
              </div>
              <div>
                <span class="text-[var(--color-text-muted)] line-through font-mono text-sm">{task.issue_key}</span>
                <span class="text-[var(--color-text-secondary)] line-through text-sm ml-3">{task.summary}</span>
              </div>
            </div>
            <span class="text-xs text-green-500 font-medium bg-green-500/10 px-3 py-1 rounded-full">Done</span>
          </div>
        {/each}
      </div>
    </section>
  {/if}
</div>

<!-- Block Modal -->
{#if showBlockModal}
  <div class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50" on:click={() => showBlockModal = false}>
    <div class="bg-[var(--color-bg-card)] rounded-2xl border border-[var(--color-accent)]/30 p-8 max-w-md w-full mx-4 shadow-2xl" on:click|stopPropagation>
      <div class="flex items-center gap-4 mb-6">
        <div class="w-14 h-14 rounded-2xl bg-[var(--color-accent)]/20 flex items-center justify-center">
          <span class="text-3xl">⏱️</span>
        </div>
        <div>
          <h2 class="text-xl font-bold text-[var(--color-text-primary)]">Timer Running</h2>
          <p class="text-[var(--color-text-muted)] text-sm">Action blocked</p>
        </div>
      </div>
      <p class="text-[var(--color-text-secondary)] mb-8 leading-relaxed">{blockMessage}</p>
      <div class="flex gap-3">
        <button
          on:click={() => handleStopAndAction(async () => {})}
          class="flex-1 px-6 py-3 bg-gradient-to-r from-[var(--color-accent)] to-[var(--color-accent-dark)] hover:opacity-90 text-white font-semibold rounded-xl transition-all duration-200 shadow-lg"
        >
          Stop Timer
        </button>
        <button
          on:click={() => showBlockModal = false}
          class="px-6 py-3 bg-[var(--color-bg-surface)] hover:bg-[var(--color-border)] text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] rounded-xl transition-all duration-200 border border-[var(--color-border)]"
        >
          Cancel
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Modal -->
{#if showDeleteModal}
  <div class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50" on:click={cancelDelete}>
    <div class="bg-[var(--color-bg-card)] rounded-2xl border border-red-500/30 p-8 max-w-md w-full mx-4 shadow-2xl" on:click|stopPropagation>
      <div class="flex items-center gap-4 mb-6">
        <div class="w-14 h-14 rounded-2xl bg-red-500/20 flex items-center justify-center">
          <span class="text-3xl">🗑️</span>
        </div>
        <div>
          <h2 class="text-xl font-bold text-[var(--color-text-primary)]">Delete Task</h2>
          <p class="text-[var(--color-text-muted)] text-sm">This action cannot be undone</p>
        </div>
      </div>
      <p class="text-[var(--color-text-secondary)] mb-8 leading-relaxed">
        Are you sure you want to delete task <strong class="text-[var(--color-accent)]">{deleteIssueKey}</strong>?
      </p>
      <div class="flex gap-3">
        <button
          on:click={confirmDelete}
          class="flex-1 px-6 py-3 bg-gradient-to-r from-red-500 to-red-600 hover:opacity-90 text-white font-semibold rounded-xl transition-all duration-200 shadow-lg"
        >
          Delete
        </button>
        <button
          on:click={cancelDelete}
          class="px-6 py-3 bg-[var(--color-bg-surface)] hover:bg-[var(--color-border)] text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] rounded-xl transition-all duration-200 border border-[var(--color-border)]"
        >
          Cancel
        </button>
      </div>
    </div>
  </div>
{/if}
