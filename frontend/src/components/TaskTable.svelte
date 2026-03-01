<script lang="ts">
  import { onMount } from 'svelte';
  import { tasksStore, tasksFilter } from '../stores/tasks';
  import { timerStore, formatTimeShort } from '../stores/timer';
  import { GetTasks, CompleteTask, UncompleteTask, DeleteTask, GetWorklogs, StopTimer } from '../../wailsjs/go/main/App';
  
  export let onViewTask: (task: any) => void;
  export let onStartTask: (issueKey: string) => void;
  
  let searchQuery = '';
  let worklogTimes: Record<string, number> = {};
  let showBlockModal = false;
  let blockMessage = '';
  let showDeleteModal = false;
  let deleteIssueKey: string | null = null;
  
  onMount(async () => {
    await loadTasks();
  });
  
  async function loadTasks() {
    const tasks = await GetTasks();
    tasksStore.set(tasks);
    
    // Load worklog times
    for (const task of tasks) {
      const worklogs = await GetWorklogs(task.issue_key);
      const totalSeconds = worklogs?.reduce((sum: number, w: any) => sum + (w.duration_seconds || 0), 0) || 0;
      worklogTimes[task.issue_key] = totalSeconds;
    }
  }
  
  $: filteredTasks = $tasksStore.filter(task => {
    // Filter by status
    if ($tasksFilter === 'active' && task.is_completed) return false;
    if ($tasksFilter === 'completed' && !task.is_completed) return false;
    
    // Filter by search
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      return task.issue_key.toLowerCase().includes(query) ||
             task.summary.toLowerCase().includes(query) ||
             task.project.toLowerCase().includes(query);
    }
    return true;
  });
  
  // Get timer state
  let timerState: any = { isRunning: false, issueKey: null };
  timerStore.subscribe(state => {
    timerState = state;
  });
  
  // Check if any timer is running for a specific task
  function isTimerRunningForTask(issueKey: string) {
    return timerState.isRunning && timerState.issueKey === issueKey;
  }
  
  // Use formatTimeShort from timer store for consistent time display
  
  function showBlockDialog(message: string) {
    blockMessage = message;
    showBlockModal = true;
  }
  
  async function handleComplete(task: any) {
    // Check if timer is running for this task
    if (isTimerRunningForTask(task.issue_key)) {
      showBlockDialog('Cannot complete task while timer is running. Please stop the timer first.');
      return;
    }
    
    if (task.is_completed) {
      await UncompleteTask(task.issue_key);
    } else {
      await CompleteTask(task.issue_key);
    }
    await loadTasks();
  }
  
  function showDeleteDialog(issueKey: string) {
    deleteIssueKey = issueKey;
    showDeleteModal = true;
  }
  
  async function confirmDelete() {
    if (!deleteIssueKey) return;
    
    showDeleteModal = false;
    try {
      await DeleteTask(deleteIssueKey);
      await loadTasks();
    } catch (err) {
      console.error('Error deleting task:', err);
    }
    deleteIssueKey = null;
  }
  
  function cancelDelete() {
    showDeleteModal = false;
    deleteIssueKey = null;
  }
  
  async function handleDelete(issueKey: string) {
    // Check if timer is running for this task
    if (isTimerRunningForTask(issueKey)) {
      showBlockDialog('Cannot delete task while timer is running. Please stop the timer first.');
      return;
    }
    
    // Show custom modal instead of confirm()
    showDeleteDialog(issueKey);
  }
  
  async function handleStartTask(issueKey: string) {
    await onStartTask(issueKey);
  }
  
  async function handleStopAndAction(action: () => Promise<void>) {
    await StopTimer();
    await action();
    showBlockModal = false;
  }
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-4">
      <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-[var(--color-accent)] to-[var(--color-accent-dark)] flex items-center justify-center shadow-glow-orange">
        <span class="text-2xl">📋</span>
      </div>
      <div>
        <h1 class="text-3xl font-bold text-[var(--color-text-primary)]">Tasks</h1>
        <p class="text-[var(--color-text-muted)] text-sm">{filteredTasks.length} tasks found</p>
      </div>
    </div>
  </div>
  
  <!-- Search and Filters -->
  <div class="flex flex-wrap gap-4">
    <div class="relative flex-1 min-w-[300px]">
      <span class="absolute left-4 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]">🔍</span>
      <input
        type="text"
        placeholder="Search tasks..."
        bind:value={searchQuery}
        class="w-full pl-12 pr-4 py-3 bg-[var(--color-bg-card)] border border-[var(--color-border)] rounded-xl text-[var(--color-text-primary)] placeholder-[var(--color-text-muted)] focus:outline-none focus:border-[var(--color-accent)] focus:ring-2 focus:ring-[var(--color-accent)]/20 transition-all"
      />
    </div>
    
    <div class="flex bg-[var(--color-bg-card)] rounded-xl overflow-hidden border border-[var(--color-border)] p-1">
      {#each ['all', 'active', 'completed'] as filter}
        <button
          on:click={() => tasksFilter.set(filter)}
          class="px-6 py-2.5 text-sm font-medium rounded-lg transition-all duration-200
            {$tasksFilter === filter 
              ? 'bg-[var(--color-accent)] text-white shadow-md' 
              : 'text-[var(--color-text-secondary)] hover:text-[var(--color-accent)] hover:bg-[var(--color-bg-surface)]'}"
        >
          {filter.charAt(0).toUpperCase() + filter.slice(1)}
        </button>
      {/each}
    </div>
  </div>
  
  <!-- Table -->
  <div class="bg-[var(--color-bg-card)] rounded-2xl border border-[var(--color-border)] overflow-hidden shadow-sm">
    <table class="w-full">
      <thead class="bg-[var(--color-bg-surface)] border-b border-[var(--color-border)]">
        <tr>
          <th class="px-6 py-4 text-left text-sm font-bold text-[var(--color-accent)] uppercase tracking-wider">Issue Key</th>
          <th class="px-6 py-4 text-left text-sm font-bold text-[var(--color-accent)] uppercase tracking-wider">Summary</th>
          <th class="px-6 py-4 text-left text-sm font-bold text-[var(--color-accent)] uppercase tracking-wider">Project</th>
          <th class="px-6 py-4 text-left text-sm font-bold text-[var(--color-accent)] uppercase tracking-wider">Est.</th>
          <th class="px-6 py-4 text-left text-sm font-bold text-[var(--color-accent)] uppercase tracking-wider">Actual</th>
          <th class="px-6 py-4 text-left text-sm font-bold text-[var(--color-accent)] uppercase tracking-wider">Status</th>
          <th class="px-6 py-4 text-left text-sm font-bold text-[var(--color-accent)] uppercase tracking-wider">Actions</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-[var(--color-border)]">
        {#each filteredTasks as task}
          <tr class="hover:bg-[var(--color-bg-surface)]/50 transition-colors group {isTimerRunningForTask(task.issue_key) ? 'bg-[var(--color-accent)]/5' : ''}">
            <td class="px-6 py-4">
              <div class="flex items-center gap-3">
                <span class="font-mono font-bold text-[var(--color-accent)] text-lg">{task.issue_key}</span>
                {#if isTimerRunningForTask(task.issue_key)}
                  <span class="flex items-center gap-1.5 text-xs bg-[var(--color-accent)] text-white px-2.5 py-1 rounded-full font-medium">
                    <span class="w-1.5 h-1.5 bg-white rounded-full animate-pulse"></span>
                    Tracking
                  </span>
                {/if}
              </div>
            </td>
            <td class="px-6 py-4">
              <span class="text-[var(--color-text-primary)] font-medium {task.is_completed ? 'line-through text-[var(--color-text-muted)]' : ''}">
                {task.summary}
              </span>
            </td>
            <td class="px-6 py-4 text-[var(--color-text-secondary)]">{task.project}</td>
            <td class="px-6 py-4 text-[var(--color-text-secondary)]">
              {task.estimated_hours ? `${task.estimated_hours}h` : '-'}
            </td>
            <td class="px-6 py-4">
              <span class="text-[var(--color-accent)] font-mono">{formatTimeShort(worklogTimes[task.issue_key] || 0)}</span>
            </td>
            <td class="px-6 py-4">
              {#if task.is_completed}
                <span class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-green-100 text-green-600 text-xs rounded-full font-semibold">
                  <span class="w-1.5 h-1.5 rounded-full bg-green-500"></span>
                  Completed
                </span>
              {:else}
                <span class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-[var(--color-accent-soft)] text-[var(--color-accent)] text-xs rounded-full font-semibold">
                  <span class="w-1.5 h-1.5 rounded-full bg-[var(--color-accent)]"></span>
                  Active
                </span>
              {/if}
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center gap-2">
                <button
                  on:click={() => onViewTask(task)}
                  class="p-2.5 hover:bg-[var(--color-bg-surface)] rounded-xl text-[var(--color-text-muted)] hover:text-[var(--color-accent)] transition-all"
                  title="View details"
                >
                  👁️
                </button>
                {#if !task.is_completed}
                  {#if isTimerRunningForTask(task.issue_key)}
                    <span class="p-2.5 text-[var(--color-accent)]" title="Timer running">⏱️</span>
                  {:else}
                    <button
                      on:click={() => handleStartTask(task.issue_key)}
                      class="p-2.5 hover:bg-green-100 rounded-xl text-[var(--color-text-muted)] hover:text-green-600 transition-all"
                      title="Start timer"
                    >
                      ▶️
                    </button>
                  {/if}
                {/if}
                <!-- Complete/Uncomplete button (checkmark icon) -->
                <button
                  on:click={() => handleComplete(task)}
                  class="p-2.5 hover:bg-[var(--color-accent-soft)] rounded-xl transition-all {task.is_completed ? 'text-green-600' : 'text-[var(--color-text-muted)] hover:text-[var(--color-accent)]'}"
                  title={task.is_completed ? "Mark as not completed" : "Mark as completed"}
                  disabled={isTimerRunningForTask(task.issue_key)}
                >
                  {task.is_completed ? '✅' : '☑️'}
                </button>
                <button
                  on:click={() => handleDelete(task.issue_key)}
                  class="p-2.5 hover:bg-red-100 rounded-xl text-[var(--color-text-muted)] hover:text-red-500 transition-all"
                  title="Delete"
                >
                  🗑️
                </button>
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
    
    {#if filteredTasks.length === 0}
      <div class="text-center py-16">
        <div class="text-5xl mb-4 opacity-30">🔍</div>
        <p class="text-[var(--color-text-secondary)] text-lg">No tasks found</p>
        <p class="text-[var(--color-text-muted)] text-sm mt-2">Try adjusting your search or filters</p>
      </div>
    {/if}
  </div>
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
