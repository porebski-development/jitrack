<script lang="ts">
  import { onMount } from 'svelte';
  import { EventsOn } from '../wailsjs/runtime';
  import { GetTasks, GetRunningTimer, StartTimer, StopTimer, GetServerPort, GetAuthToken } from '../wailsjs/go/main/App';
  
  import Header from './components/Header.svelte';
  import Sidebar from './components/Sidebar.svelte';
  import Dashboard from './components/Dashboard.svelte';
  import TaskTable from './components/TaskTable.svelte';
  import Reports from './components/Reports.svelte';
  import Settings from './components/Settings.svelte';
  import TaskDetails from './components/TaskDetails.svelte';
  
  import { tasksStore, addTask, setTasks } from './stores/tasks';
  import { timerStore, setTimerRunning, setTimerStopped, updateElapsedSeconds } from './stores/timer';
  import { BrowserOpenURL, WindowMinimise, Quit } from '../wailsjs/runtime';
  
  let currentView = 'dashboard';
  let selectedTask: any = null;
  
  // Load initial data
  onMount(async () => {
    // Load tasks
    const tasks = await GetTasks();
    if (tasks) {
      tasksStore.set(tasks);
    }
    
    // Check for running timer
    const running = await GetRunningTimer();
    if (running) {
      const elapsed = running.elapsed_seconds || 0;
      setTimerRunning(running.issue_key, '', new Date(running.started_at), elapsed);
    }
    
    // Listen for events from backend
    EventsOn('timer:tick', (data: { elapsedSeconds: number }) => {
      updateElapsedSeconds(data.elapsedSeconds);
    });
    
    EventsOn('timer:started', (data: { issueKey: string; summary: string; startedAt: string }) => {
      setTimerRunning(data.issueKey, data.summary, new Date(data.startedAt), 0);
    });
    
    EventsOn('timer:stopped', () => {
      setTimerStopped();
    });
    
    EventsOn('task:imported', async (data: { task: any }) => {
      addTask(data.task);
      // Refresh all tasks to ensure consistency
      await refreshTasks();
    });
    
    // Listen for navigation events from menu bar
    EventsOn('navigate', (view: string) => {
      currentView = view;
    });
  });
  
  async function handleStopTimer() {
    await StopTimer();
    setTimerStopped();
  }
  
  async function refreshTasks() {
    const tasks = await GetTasks();
    if (tasks) {
      setTasks(tasks);
    }
  }
  
  function handleNavigate(view: string) {
    currentView = view;
  }
  
  function handleViewTask(task: any) {
    selectedTask = task;
  }
  
  function handleCloseTaskDetails() {
    selectedTask = null;
  }
  
  async function handleStartTask(issueKey: string) {
    await StartTimer(issueKey);
    const task = $tasksStore.find(t => t.issue_key === issueKey);
    setTimerRunning(issueKey, task?.summary || '', new Date(), 0);
  }
</script>

<div class="h-screen flex flex-col bg-[var(--color-bg-primary)] text-[var(--color-text-primary)]">
  <Header 
    onStopTimer={handleStopTimer}
  />
  
  <div class="flex-1 flex overflow-hidden">
    <Sidebar 
      currentView={currentView}
      onNavigate={handleNavigate}
    />
    
    <main class="flex-1 overflow-auto p-6">
      {#if currentView === 'dashboard'}
        <Dashboard 
          onViewTask={handleViewTask}
          onStartTask={handleStartTask}
        />
      {:else if currentView === 'tasks'}
        <TaskTable 
          onViewTask={handleViewTask}
          onStartTask={handleStartTask}
        />
      {:else if currentView === 'reports'}
        <Reports />
      {:else if currentView === 'settings'}
        <Settings />
      {/if}
    </main>
  </div>
  
  {#if selectedTask}
    <TaskDetails 
      task={selectedTask}
      onClose={handleCloseTaskDetails}
      onStart={() => handleStartTask(selectedTask.issue_key)}
    />
  {/if}
</div>
