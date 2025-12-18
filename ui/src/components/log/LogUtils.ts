import { useState, useEffect } from 'react';

// Debounce hook for search input
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
}

// Get default date range (yesterday 00:00 to today 23:59 in local time)
export function getDefaultDateRange(): { start_time: string; end_time: string } {
  const now = new Date();

  // Yesterday 00:00:00
  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  yesterday.setHours(0, 0, 0, 0);

  // Today 23:59:59
  const today = new Date(now);
  today.setHours(23, 59, 59, 999);

  return {
    start_time: yesterday.toISOString(),
    end_time: today.toISOString(),
  };
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

export function formatTime(timestamp: string): string {
  return new Date(timestamp).toLocaleString();
}
