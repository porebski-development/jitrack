import { writable } from 'svelte/store';

export type Theme = 'dark' | 'light';

// Get initial theme from localStorage or default to dark
const getInitialTheme = (): Theme => {
  if (typeof window !== 'undefined' && window.localStorage) {
    const stored = localStorage.getItem('jitrack-theme');
    if (stored === 'light' || stored === 'dark') {
      return stored;
    }
  }
  return 'dark';
};

export const themeStore = writable<Theme>(getInitialTheme());

// Subscribe to changes and update localStorage + document class
themeStore.subscribe((theme) => {
  if (typeof document !== 'undefined') {
    if (theme === 'dark') {
      document.documentElement.classList.add('dark');
      document.documentElement.classList.remove('light');
    } else {
      document.documentElement.classList.add('light');
      document.documentElement.classList.remove('dark');
    }
    
    // Store in localStorage
    if (typeof window !== 'undefined' && window.localStorage) {
      localStorage.setItem('jitrack-theme', theme);
    }
  }
});

export const toggleTheme = () => {
  themeStore.update((current) => (current === 'dark' ? 'light' : 'dark'));
};

export const setTheme = (theme: Theme) => {
  themeStore.set(theme);
};
