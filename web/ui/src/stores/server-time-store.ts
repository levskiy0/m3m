import { create } from 'zustand';
import type { ServerTime } from '@/lib/websocket';

interface ServerTimeState {
  serverTime: ServerTime | null;
  setServerTime: (time: ServerTime) => void;
}

export const useServerTimeStore = create<ServerTimeState>((set) => ({
  serverTime: null,
  setServerTime: (time) => set({ serverTime: time }),
}));
