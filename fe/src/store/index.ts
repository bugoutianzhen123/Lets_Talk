import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface User {
  id: number;
  username: string;
}

interface ChatStore {
  currentUser: User | null;
  setCurrentUser: (user: User | null) => void;
  token: string | null;
  setToken: (token: string | null) => void;
}

// 登录态持久化到 localStorage，刷新页面后无需重新登录
export const useChatStore = create<ChatStore>()(
  persist(
    (set) => ({
      currentUser: null,
      setCurrentUser: (user) => set({ currentUser: user }),
      token: null,
      setToken: (token) => set({ token }),
    }),
    { name: 'lets-talk-auth' },
  ),
);
