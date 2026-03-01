<script lang="ts">
  import { onMount } from 'svelte';
  import { GetDailyReport } from '../../wailsjs/go/main/App';
  import { formatTimeShort } from '../stores/timer';
  import Chart from 'chart.js/auto';
  
  let selectedDate = new Date().toISOString().split('T')[0];
  let reportData: any[] = [];
  let chartCanvas: HTMLCanvasElement;
  let chart: Chart;
  let activePreset: 'today' | 'yesterday' | 'custom' = 'today';
  
  $: totalSeconds = reportData.reduce((sum, row) => sum + row.total_seconds, 0);
  $: totalSessions = reportData.reduce((sum, row) => sum + row.sessions, 0);
  $: totalTasks = reportData.length;
  
  onMount(async () => {
    // Small delay to ensure canvas is bound
    setTimeout(async () => {
      await loadReport();
    }, 0);
  });
  
  // Reactive update when reportData changes
  $: if (chartCanvas && reportData !== undefined) {
    updateChart();
  }
  
  async function loadReport() {
    reportData = await GetDailyReport(selectedDate) || [];
  }
  
  // Get computed CSS color value
  function getCssColor(varName: string): string {
    if (typeof window === 'undefined') return '#f97316';
    const style = getComputedStyle(document.documentElement);
    // Try to get from CSS variable
    const value = style.getPropertyValue(varName).trim();
    if (value) return value;
    // Fallback colors
    const fallbacks: Record<string, string> = {
      '--color-accent': '#f97316',
      '--color-accent-light': '#fb923c',
      '--color-border': '#3f2c22',
      '--color-text-secondary': '#d6d3d1',
    };
    return fallbacks[varName] || '#f97316';
  }
  
  function updateChart() {
    if (chart) {
      chart.destroy();
    }
    
    if (reportData.length === 0 || !chartCanvas) return;
    
    const ctx = chartCanvas.getContext('2d');
    if (!ctx) return;
    
    const accentColor = getCssColor('--color-accent');
    const borderColor = getCssColor('--color-border');
    const textColor = getCssColor('--color-text-secondary');
    
    // Format time for tooltip
    const formatTooltip = (minutes: number) => {
      const h = Math.floor(minutes / 60);
      const m = Math.round(minutes % 60);
      if (h === 0) return `${m} min`;
      if (m === 0) return `${h}h`;
      return `${h}h ${m}min`;
    };
    
    chart = new Chart(ctx, {
      type: 'bar',
      data: {
        labels: reportData.map(r => r.issue_key),
        datasets: [{
          label: 'Time',
          data: reportData.map(r => Math.round(r.total_seconds / 60)), // minutes
          backgroundColor: accentColor,
          borderRadius: 8,
          borderSkipped: false,
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              label: (context) => `Time: ${formatTooltip(context.parsed.y)}`
            }
          }
        },
        scales: {
          y: {
            beginAtZero: true,
            grid: { color: borderColor },
            ticks: { 
              color: textColor,
              callback: (value) => {
                const minutes = Number(value);
                const h = Math.floor(minutes / 60);
                const m = minutes % 60;
                if (h === 0) return `${m}m`;
                if (m === 0) return `${h}h`;
                return `${h}h${m}m`;
              }
            }
          },
          x: {
            grid: { display: false },
            ticks: { color: textColor }
          }
        }
      }
    });
  }
  
  // Use formatTimeShort from timer store for consistent time display
  
  function getPercentOfTotal(seconds: number): string {
    if (totalSeconds === 0) return '0%';
    return ((seconds / totalSeconds) * 100).toFixed(1) + '%';
  }
  
  function setPreset(preset: 'today' | 'yesterday') {
    const today = new Date();
    if (preset === 'today') {
      selectedDate = today.toISOString().split('T')[0];
      activePreset = 'today';
    } else if (preset === 'yesterday') {
      today.setDate(today.getDate() - 1);
      selectedDate = today.toISOString().split('T')[0];
      activePreset = 'yesterday';
    }
    loadReport();
  }
  
  function onDateChange() {
    activePreset = 'custom';
    loadReport();
  }
</script>

<div class="space-y-8">
  <!-- Header -->
  <div class="flex items-center gap-4">
    <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-[var(--color-accent)] to-[var(--color-accent-dark)] flex items-center justify-center shadow-glow-orange">
      <span class="text-2xl">📈</span>
    </div>
    <div>
      <h1 class="text-3xl font-bold text-[var(--color-text-primary)]">Reports</h1>
      <p class="text-[var(--color-text-muted)] text-sm">Analyze your time tracking data</p>
    </div>
  </div>
  
  <!-- Date Picker -->
  <div class="flex flex-wrap gap-4 items-center">
    <div class="flex bg-[var(--color-bg-card)] rounded-xl overflow-hidden border border-[var(--color-border)] p-1">
      <button 
        on:click={() => setPreset('today')} 
        class="px-6 py-2.5 text-sm font-medium rounded-lg transition-all duration-200
          {activePreset === 'today' 
            ? 'bg-[var(--color-accent)] text-white shadow-md' 
            : 'hover:bg-[var(--color-accent)]/10 text-[var(--color-text-secondary)] hover:text-[var(--color-accent)]'}"
      >
        Today
      </button>
      <button 
        on:click={() => setPreset('yesterday')} 
        class="px-6 py-2.5 text-sm font-medium rounded-lg transition-all duration-200
          {activePreset === 'yesterday' 
            ? 'bg-[var(--color-accent)] text-white shadow-md' 
            : 'hover:bg-[var(--color-accent)]/10 text-[var(--color-text-secondary)] hover:text-[var(--color-accent)]'}"
      >
        Yesterday
      </button>
    </div>
    
    <input
      type="date"
      bind:value={selectedDate}
      on:change={onDateChange}
      class="px-4 py-3 bg-[var(--color-bg-card)] border border-[var(--color-border)] rounded-xl text-[var(--color-text-primary)] focus:outline-none focus:border-[var(--color-accent)] focus:ring-2 focus:ring-[var(--color-accent)]/20"
    />
  </div>
  
  <!-- Summary Cards -->
  <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
    <div class="bg-[var(--color-bg-card)] rounded-2xl p-6 border border-[var(--color-border)] shadow-sm relative overflow-hidden">
      <div class="absolute top-0 right-0 w-32 h-32 bg-[var(--color-accent)]/10 rounded-full blur-2xl -translate-y-1/2 translate-x-1/2"></div>
      <div class="relative z-10">
        <div class="text-sm font-medium text-[var(--color-accent)] uppercase tracking-wider mb-2">Total Hours</div>
        <div class="text-4xl font-bold text-[var(--color-text-primary)]">{formatTimeShort(totalSeconds)}</div>
      </div>
    </div>
    <div class="bg-[var(--color-bg-card)] rounded-2xl p-6 border border-[var(--color-border)] shadow-sm relative overflow-hidden">
      <div class="absolute top-0 right-0 w-32 h-32 bg-[var(--color-accent)]/10 rounded-full blur-2xl -translate-y-1/2 translate-x-1/2"></div>
      <div class="relative z-10">
        <div class="text-sm font-medium text-[var(--color-accent)] uppercase tracking-wider mb-2">Total Sessions</div>
        <div class="text-4xl font-bold text-[var(--color-text-primary)]">{totalSessions}</div>
      </div>
    </div>
    <div class="bg-[var(--color-bg-card)] rounded-2xl p-6 border border-[var(--color-border)] shadow-sm relative overflow-hidden">
      <div class="absolute top-0 right-0 w-32 h-32 bg-[var(--color-accent)]/10 rounded-full blur-2xl -translate-y-1/2 translate-x-1/2"></div>
      <div class="relative z-10">
        <div class="text-sm font-medium text-[var(--color-accent)] uppercase tracking-wider mb-2">Tasks Worked</div>
        <div class="text-4xl font-bold text-[var(--color-text-primary)]">{totalTasks}</div>
      </div>
    </div>
  </div>
  
  <!-- Chart -->
  {#if reportData.length > 0}
    <div class="bg-[var(--color-bg-card)] rounded-2xl p-6 border border-[var(--color-border)] shadow-sm h-[350px]">
      <canvas bind:this={chartCanvas}></canvas>
    </div>
  {/if}
  
  <!-- Table -->
  <div class="bg-[var(--color-bg-card)] rounded-2xl border border-[var(--color-border)] overflow-hidden shadow-sm">
    <table class="w-full">
      <thead class="bg-[var(--color-bg-surface)] border-b border-[var(--color-border)]">
        <tr>
          <th class="px-6 py-4 text-left text-sm font-bold text-[var(--color-accent)] uppercase tracking-wider">Task</th>
          <th class="px-6 py-4 text-left text-sm font-bold text-[var(--color-accent)] uppercase tracking-wider">Sessions</th>
          <th class="px-6 py-4 text-left text-sm font-bold text-[var(--color-accent)] uppercase tracking-wider">Time</th>
          <th class="px-6 py-4 text-left text-sm font-bold text-[var(--color-accent)] uppercase tracking-wider">% of Total</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-[var(--color-border)]">
        {#each reportData as row}
          <tr class="hover:bg-[var(--color-bg-surface)]/50 transition-colors">
            <td class="px-6 py-4">
              <div class="font-mono font-bold text-[var(--color-accent)]">{row.issue_key}</div>
              <div class="text-sm text-[var(--color-text-secondary)] truncate max-w-[300px]">{row.summary}</div>
            </td>
            <td class="px-6 py-4 text-[var(--color-text-primary)] font-semibold">{row.sessions}</td>
            <td class="px-6 py-4 text-[var(--color-accent)] font-mono font-bold">{formatTimeShort(row.total_seconds)}</td>
            <td class="px-6 py-4">
              <div class="flex items-center gap-3">
                <div class="flex-1 h-2 bg-[var(--color-bg-surface)] rounded-full overflow-hidden max-w-[100px]">
                  <div class="h-full bg-[var(--color-accent)] rounded-full" style="width: {getPercentOfTotal(row.total_seconds)}"></div>
                </div>
                <span class="text-[var(--color-text-secondary)] text-sm">{getPercentOfTotal(row.total_seconds)}</span>
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
    
    {#if reportData.length === 0}
      <div class="text-center py-16">
        <div class="text-5xl mb-4 opacity-30">📊</div>
        <p class="text-[var(--color-text-secondary)] text-lg">No data for selected date</p>
        <p class="text-[var(--color-text-muted)] text-sm mt-2">Try selecting a different date</p>
      </div>
    {/if}
  </div>
</div>
