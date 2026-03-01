import { writable } from 'svelte/store';

export interface TimerState {
  isRunning: boolean;
  issueKey: string | null;
  summary: string | null;
  elapsedSeconds: number;
  startedAt: Date | null;
}

const initialState: TimerState = {
  isRunning: false,
  issueKey: null,
  summary: null,
  elapsedSeconds: 0,
  startedAt: null,
};

export const timerStore = writable<TimerState>(initialState);

export const setTimerRunning = (
  issueKey: string,
  summary: string,
  startedAt: Date,
  elapsedSeconds: number = 0
) => {
  timerStore.set({
    isRunning: true,
    issueKey,
    summary,
    startedAt,
    elapsedSeconds,
  });
};

export const setTimerStopped = () => {
  timerStore.set(initialState);
};

export const updateElapsedSeconds = (elapsedSeconds: number) => {
  timerStore.update((state) => ({
    ...state,
    elapsedSeconds,
  }));
};

// Format seconds to HH:MM:SS (for timer display)
export const formatDuration = (seconds: number): string => {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;
  return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
};

// Format seconds to human readable (e.g. "40 min" or "1h 40min")
export const formatTimeShort = (seconds: number): string => {
  if (seconds === 0) return '-';
  
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  
  if (hours === 0) {
    return `${minutes} min`;
  } else if (minutes === 0) {
    return `${hours}h`;
  } else {
    return `${hours}h ${minutes}min`;
  }
};
