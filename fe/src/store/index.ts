import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';

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

// 登录态持久化到 sessionStorage：刷新页面不会退出登录，且各浏览器窗口/标签互不影响，
// 便于同一浏览器开两个窗口分别登录不同账号进行演示。
export const useChatStore = create<ChatStore>()(
  persist(
    (set) => ({
      currentUser: null,
      setCurrentUser: (user) => set({ currentUser: user }),
      token: null,
      setToken: (token) => set({ token }),
    }),
    { name: 'lets-talk-auth', storage: createJSONStorage(() => sessionStorage) },
  ),
);
