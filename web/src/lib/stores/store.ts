import { writable } from 'svelte/store';
import type { RecordModel } from 'pocketbase';

export const user = writable<RecordModel | null>(null);

interface ToastData {
  id: number;
  message: string;
  type: 'info' | 'success' | 'warning' |'error';
}

export const toasts = writable<ToastData[]>([]);

let toastId = 0;

export function showToast(message: string, type: 'info' | 'success' | 'warning' |'error', duration = 3000) {
  const id = toastId++;
  toasts.update((allToasts) => [...allToasts, { id, message, type }]);
  setTimeout(() => {
    toasts.update((allToasts) => allToasts.filter((toast) => toast.id !== id));
  }, duration);
}