import { writable } from 'svelte/store';

export interface Task {
  issue_key: string;
  summary: string;
  project: string;
  url: string;
  estimated_hours?: number;
  is_completed: boolean;
  imported_at: string;
  last_updated: string;
}

export const tasksStore = writable<Task[]>([]);
export const tasksFilter = writable<'all' | 'active' | 'completed'>('all');

export const setTasks = (tasks: Task[]) => {
  tasksStore.set(tasks);
};

export const addTask = (task: Task) => {
  tasksStore.update((tasks) => {
    // Check if task already exists
    const exists = tasks.some((t) => t.issue_key === task.issue_key);
    if (exists) {
      // Replace existing task
      return tasks.map((t) => (t.issue_key === task.issue_key ? task : t));
    }
    // Add new task at the beginning
    return [task, ...tasks];
  });
};

export const updateTask = (issueKey: string, updates: Partial<Task>) => {
  tasksStore.update((tasks) =>
    tasks.map((t) => (t.issue_key === issueKey ? { ...t, ...updates } : t))
  );
};

export const removeTask = (issueKey: string) => {
  tasksStore.update((tasks) => tasks.filter((t) => t.issue_key !== issueKey));
};
