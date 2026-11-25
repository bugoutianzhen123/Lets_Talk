import React, { useState, useEffect, useRef } from 'react';
import {
  Layout, Input, Button, Avatar, List, Typography, Dropdown, Space,
  Empty, Modal, Form, Switch, message, Divider, Popconfirm, Tooltip, Badge,
} from 'antd';
import {
  MessageOutlined,
  UserOutlined,
  SettingOutlined,
  SendOutlined,
  LogoutOutlined,
  TeamOutlined,
  PlusOutlined,
  UserAddOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useChatStore } from '../store';
import moment from 'moment';

const { Header, Sider, Content } = Layout;
const { Title, Text } = Typography;

// 接口统一使用相对路径：本地由 Vite 代理转发，线上由 Nginx 反代到后端，避免写死后端地址
const API_BASE = '';
// WebSocket 需要绝对地址，基于当前页面 origin 推导（http→ws / https→wss）
const wsUrl = (path: string) => {
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
  return `${proto}://${window.location.host}${path}`;
};

interface GroupItem {
  id: number;
  name: string;
  joined?: boolean;
  last_message?: string;
  last_time?: string;
  last_sender_id?: number;
}

interface FriendItem {
  id: number;
  username: string;
  online?: boolean;
  last_message?: string;
  last_time?: string;
  last_sender_id?: number;
}

interface ApplicationItem {
  id: number;
  message: string;
  sender_id: number;
  username: string;
  created: string;
}

interface ChatMessage {
  user_id: number;
  target_id?: number;
  username: string;
  content: string;
  send_time: string;
}

type NavKey = 'msg' | 'contacts';

// 本地生成默认头像（SVG data URI）：按名称哈希取色 + 首字符，无需联网
const AVATAR_COLORS = ['#1677ff', '#13c2c2', '#52c41a', '#faad14', '#eb2f96', '#722ed1', '#2f54eb', '#fa541c'];
const avatarOf = (seed: string) => {
  const name = (seed || '?').trim() || '?';
  const ch = name.charAt(0).toUpperCase();
  const safe = ch.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) >>> 0;
  const color = AVATAR_COLORS[hash % AVATAR_COLORS.length];
  const svg =
    `<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64">` +
    `<rect width="64" height="64" rx="8" fill="${color}"/>` +
    `<text x="50%" y="50%" dy="0.35em" text-anchor="middle" font-size="28" fill="#fff" font-family="sans-serif">${safe}</text>` +
    `</svg>`;
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`;
};

// 会话列表时间：今天显示 HH:mm，其余显示 MM-DD
const formatListTime = (t?: string) => {
  if (!t) return '';
  const d = moment(t);
  if (!d.isValid()) return '';
  return d.isSame(moment(), 'day') ? d.format('HH:mm') : d.format('MM-DD');
};

// 会话排序用：无最后消息的排到末尾
const lastTimeValue = (item: { last_time?: string }) =>
  item.last_time ? new Date(item.last_time).getTime() : 0;

// 会话预览文案：区分「我：」与他人消息
const previewText = (item: { last_message?: string; last_sender_id?: number }, myId: number) => {
  if (!item.last_message) return '';
  return item.last_sender_id === myId ? `我：${item.last_message}` : item.last_message;
};

export const ChatLayout: React.FC = () => {
  const { currentUser, setCurrentUser } = useChatStore();
  const navigate = useNavigate();

  // 群聊
  const [groups, setGroups] = useState<GroupItem[]>([]);
  const [activeGroupId, setActiveGroupId] = useState<string | null>(null);
  // 群聊已读时间戳（本地记录，用于计算未读红点）
  const [groupReadTs, setGroupReadTs] = useState<Record<number, number>>({});

  // 单聊
  const [activeFriendId, setActiveFriendId] = useState<number | null>(null);
  const [friendMessages, setFriendMessages] = useState<ChatMessage[]>([]);
  const [unread, setUnread] = useState<Record<number, number>>({});

  // 好友
  const [friends, setFriends] = useState<FriendItem[]>([]);
  const [applications, setApplications] = useState<ApplicationItem[]>([]);

  // 界面状态
  const [activeNav, setActiveNav] = useState<NavKey>('msg');
  const [searchText, setSearchText] = useState('');
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [autoScroll, setAutoScroll] = useState(true);

  // 弹窗
  const [profileOpen, setProfileOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [createGroupOpen, setCreateGroupOpen] = useState(false);
  const [addFriendOpen, setAddFriendOpen] = useState(false);
  const [userKeyword, setUserKeyword] = useState('');
  const [searchResults, setSearchResults] = useState<FriendItem[]>([]);
  const [searching, setSearching] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [groupForm] = Form.useForm();

  const wsRef = useRef<WebSocket | null>(null);
  const wsFriendRef = useRef<WebSocket | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 供 WebSocket 回调读取最新值，避免闭包捕获旧状态
  const activeFriendIdRef = useRef<number | null>(null);
  const activeNavRef = useRef<NavKey>('msg');
  useEffect(() => { activeFriendIdRef.current = activeFriendId; }, [activeFriendId]);
  useEffect(() => { activeNavRef.current = activeNav; }, [activeNav]);

  const activeGroup = groups.find(g => String(g.id) === activeGroupId) || null;

  useEffect(() => {
    if (!currentUser) {
      navigate('/login');
      return;
    }
    loadGroups();
    loadFriends();
    loadApplications();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentUser, navigate]);

  // ---------- 数据加载 ----------
  // silent=true 用于轮询静默刷新，失败不弹提示
  const loadGroups = async (selectId?: number, silent = false) => {
    try {
      const res = await fetch(`${API_BASE}/group/list?userid=${currentUser?.id ?? ''}`);
      const data = await res.json();
      const list: GroupItem[] = data.data || [];
      setGroups(list);
      setActiveGroupId(prev => {
        const target = selectId ?? (prev ? Number(prev) : undefined);
        // 只有已加入的群聊才能作为当前会话
        if (target && list.some(g => g.id === target && g.joined)) return String(target);
        const firstJoined = list.find(g => g.joined);
        return firstJoined ? String(firstJoined.id) : null;
      });
    } catch {
      if (!silent) message.error('加载群聊列表失败');
    }
  };

  const handleJoinGroup = async (groupId: number) => {
    try {
      const res = await fetch(`${API_BASE}/group/join`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ group_id: groupId, user_id: currentUser?.id }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || '加入群聊失败');
      message.success('已加入群聊');
      await loadGroups(groupId);
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const loadFriends = async (silent = false) => {
    if (!currentUser) return;
    try {
      const res = await fetch(`${API_BASE}/friend/list?userid=${currentUser.id}`);
      const data = await res.json();
      setFriends(data.data || []);
    } catch {
      if (!silent) message.error('加载好友列表失败');
    }
  };

  const loadApplications = async (silent = false) => {
    if (!currentUser) return;
    try {
      const res = await fetch(`${API_BASE}/friend/applications?userid=${currentUser.id}`);
      const data = await res.json();
      setApplications(data.data || []);
    } catch {
      if (!silent) message.error('加载好友申请失败');
    }
  };

  // 轮询 + 窗口聚焦刷新：另一窗口的好友申请/建群/加群操作无需手动刷新即可同步
  useEffect(() => {
    if (!currentUser) return;
    const refresh = () => {
      loadGroups(undefined, true);
      loadFriends(true);
      loadApplications(true);
    };
    const timer = setInterval(refresh, 5000);
    window.addEventListener('focus', refresh);
    return () => {
      clearInterval(timer);
      window.removeEventListener('focus', refresh);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentUser]);

  // ---------- 进入群聊：拉历史 + 建立实时连接 ----------
  useEffect(() => {
    if (!currentUser || !activeGroupId) {
      setMessages([]);
      return;
    }

    let cancelled = false;
    setMessages([]);
    setIsConnected(false);

    fetch(`${API_BASE}/group/history?groupid=${activeGroupId}`)
      .then(res => res.json())
      .then(res => {
        if (cancelled) return;
        const history: ChatMessage[] = (res.data || []).map((item: any) => ({
          user_id: item.sender_id,
          username: item.Sender?.username || `用户${item.sender_id}`,
          content: item.text,
          send_time: item.created,
        }));
        setMessages(history);
      })
      .catch(() => {
        if (!cancelled) setMessages([]);
      });

    const ws = new WebSocket(wsUrl(`/ws?groupid=${activeGroupId}&userid=${currentUser.id}`));

    ws.onopen = () => {
      if (wsRef.current === ws) setIsConnected(true);
    };
    ws.onclose = () => {
      if (wsRef.current === ws) setIsConnected(false);
    };
    ws.onmessage = (event) => {
      try {
        const msg: ChatMessage = JSON.parse(event.data);
        setMessages(prev => [...prev, msg]);
      } catch (e) {
        console.error(e);
      }
    };

    wsRef.current = ws;

    return () => {
      cancelled = true;
      ws.close();
    };
  }, [activeGroupId, currentUser]);

  // 打开群聊或收到该群新消息时，标记为已读（用于清除未读红点）
  useEffect(() => {
    if (!activeGroupId) return;
    setGroupReadTs(prev => ({ ...prev, [Number(activeGroupId)]: Date.now() }));
  }, [activeGroupId, messages]);

  // ---------- 私聊：常驻一条长连接，断线自动重连，按 target_id 路由消息 ----------
  useEffect(() => {
    if (!currentUser) return;

    let stopped = false;
    let current: WebSocket | null = null;
    let retryTimer: number | undefined;

    const connect = () => {
      if (stopped) return;
      const ws = new WebSocket(wsUrl(`/ws/friend?userid=${currentUser.id}`));
      current = ws;
      wsFriendRef.current = ws;

      ws.onmessage = (event) => {
        if (wsFriendRef.current !== ws) return;
        try {
          const raw = JSON.parse(event.data);
          // 后端私聊推送使用 sender_id，这里统一成与群聊一致的 user_id
          const msg: ChatMessage = {
            user_id: raw.sender_id,
            target_id: raw.target_id,
            username: raw.username,
            content: raw.content,
            send_time: raw.send_time,
          };
          const counterpart = msg.user_id === currentUser.id ? msg.target_id : msg.user_id;
          // 当前正与该好友聊天时直接追加；否则累计未读
          if (counterpart === activeFriendIdRef.current && activeNavRef.current === 'contacts') {
            setFriendMessages(prev => [...prev, msg]);
          } else {
            setUnread(prev => ({ ...prev, [counterpart as number]: (prev[counterpart as number] || 0) + 1 }));
          }
        } catch (e) {
          console.error(e);
        }
      };

      // 断线后 3 秒自动重连，避免长连接静默断开导致「在线状态/消息推送」失效
      ws.onclose = () => {
        if (wsFriendRef.current === ws) wsFriendRef.current = null;
        if (!stopped) retryTimer = window.setTimeout(connect, 3000);
      };
    };

    connect();

    return () => {
      stopped = true;
      if (retryTimer) window.clearTimeout(retryTimer);
      wsFriendRef.current = null;
      current?.close();
    };
  }, [currentUser]);

  // ---------- 进入某个好友的会话：拉取历史记录 ----------
  useEffect(() => {
    if (!currentUser || !activeFriendId) {
      setFriendMessages([]);
      return;
    }

    let cancelled = false;
    setFriendMessages([]);

    fetch(`${API_BASE}/friend/history?userid=${currentUser.id}&friendid=${activeFriendId}`)
      .then(res => res.json())
      .then(res => {
        if (cancelled) return;
        const history: ChatMessage[] = (res.data || []).map((item: any) => ({
          user_id: item.sender_id,
          target_id: item.target_id,
          username: item.Sender?.username || `用户${item.sender_id}`,
          content: item.text,
          send_time: item.created,
        }));
        setFriendMessages(history);
      })
      .catch(() => {
        if (!cancelled) setFriendMessages([]);
      });

    // 打开会话即清除该好友未读
    setUnread(prev => {
      if (!prev[activeFriendId]) return prev;
      const next = { ...prev };
      delete next[activeFriendId];
      return next;
    });

    return () => { cancelled = true; };
  }, [activeFriendId, currentUser]);

  useEffect(() => {
    if (autoScroll && (activeNav === 'msg' || activeFriendId)) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, friendMessages, autoScroll, activeNav, activeFriendId]);

  // ---------- 操作 ----------
  const handleSend = () => {
    if (!inputValue.trim() || !wsRef.current || !isConnected) return;
    wsRef.current.send(inputValue);
    setInputValue('');
  };

  const handleSendFriend = () => {
    const ws = wsFriendRef.current;
    if (!inputValue.trim() || !activeFriendId || !ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({ target_id: activeFriendId, content: inputValue }));
    setInputValue('');
  };

  const handleLogout = () => {
    setCurrentUser(null);
    navigate('/login');
  };

  const handleCreateGroup = async () => {
    const name = groupForm.getFieldValue('name');
    if (!name || !name.trim()) {
      message.warning('请输入群聊名称');
      return;
    }
    setSubmitting(true);
    try {
      const res = await fetch(`${API_BASE}/group/create`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name.trim(), creator_id: currentUser?.id }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || '创建群聊失败');
      message.success('群聊创建成功');
      setCreateGroupOpen(false);
      groupForm.resetFields();
      setActiveNav('msg');
      await loadGroups(data.data.id);
    } catch (e: any) {
      message.error(e.message);
    } finally {
      setSubmitting(false);
    }
  };

  const handleSearchUsers = async (keyword: string) => {
    setUserKeyword(keyword);
    if (!keyword.trim()) {
      setSearchResults([]);
      return;
    }
    setSearching(true);
    try {
      const res = await fetch(`${API_BASE}/user/search?keyword=${encodeURIComponent(keyword.trim())}`);
      const data = await res.json();
      setSearchResults((data.data || []).filter((u: FriendItem) => u.id !== currentUser?.id));
    } catch {
      message.error('搜索失败');
    } finally {
      setSearching(false);
    }
  };

  const handleApplyFriend = async (receiverId: number) => {
    try {
      const res = await fetch(`${API_BASE}/friend/apply`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sender_id: currentUser?.id,
          receiver_id: receiverId,
          message: '你好，我想加你为好友',
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || '申请失败');
      message.success('好友申请已发送');
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const handleAcceptApplication = async (applicationId: number) => {
    try {
      const res = await fetch(`${API_BASE}/friend/accept`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ application_id: applicationId }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || '操作失败');
      message.success('已添加为好友');
      await loadApplications();
      await loadFriends();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const switchNav = (key: string) => {
    setActiveNav(key as NavKey);
    setSearchText('');
  };

  const openAddFriend = () => {
    setAddFriendOpen(true);
    setUserKeyword('');
    setSearchResults([]);
  };

  // ---------- 派生数据 ----------
  const keyword = searchText.trim().toLowerCase();
  // 未搜索时只展示已加入的群聊；搜索时展示名称/群号匹配的全部群聊（含未加入，可点击「加入」）
  const visibleGroups = keyword
    ? groups.filter(g => g.name.toLowerCase().includes(keyword) || String(g.id).includes(keyword))
    : groups.filter(g => g.joined);
  // 按最后消息时间倒序，像 IM 一样把最近活跃的会话排在前面
  const filteredGroups = [...visibleGroups].sort((a, b) => lastTimeValue(b) - lastTimeValue(a));
  const filteredFriends = friends
    .filter(f => f.username.toLowerCase().includes(keyword))
    .sort((a, b) => lastTimeValue(b) - lastTimeValue(a));
  const isFriend = (id: number) => friends.some(f => f.id === id);

  const userMenuItems = [
    { key: 'profile', icon: <UserOutlined />, label: '个人信息', onClick: () => setProfileOpen(true) },
    { key: 'settings', icon: <SettingOutlined />, label: '设置', onClick: () => setSettingsOpen(true) },
    { type: 'divider' as const },
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', danger: true, onClick: handleLogout },
  ];

  if (!currentUser) return null;

  const activeFriend = friends.find(f => f.id === activeFriendId) || null;

  // 群聊未读：最后一条消息不是我发的，晚于本地已读时间，且当前未在看该群
  const hasGroupUnread = (g: GroupItem) => {
    if (!g.joined || !g.last_time) return false;
    if (g.last_sender_id === currentUser.id) return false;
    if (activeGroupId === String(g.id)) return false;
    return new Date(g.last_time).getTime() > (groupReadTs[g.id] || 0);
  };

  // 消息气泡列表（群聊 / 单聊共用）
  const renderMessageList = (list: ChatMessage[]) => {
    if (list.length === 0) {
      return <div style={{ textAlign: 'center', marginTop: 100, color: '#999' }}>暂无消息，来打个招呼吧</div>;
    }
    return list.map((msg, index) => {
      const isMe = msg.user_id.toString() === currentUser.id.toString();
      const displayName = msg.username || `用户${msg.user_id}`;
      return (
        <div key={index} style={{ display: 'flex', flexDirection: isMe ? 'row-reverse' : 'row', marginBottom: 20 }}>
          <Avatar src={avatarOf(displayName)} />
          <div style={{
            maxWidth: '60%',
            margin: isMe ? '0 12px 0 0' : '0 0 0 12px',
            display: 'flex',
            flexDirection: 'column',
            alignItems: isMe ? 'flex-end' : 'flex-start'
          }}>
            {!isMe && <Text type="secondary" style={{ fontSize: 12, marginBottom: 4 }}>{displayName}</Text>}
            <div style={{
              padding: '10px 16px',
              background: isMe ? '#1890ff' : '#fff',
              color: isMe ? '#fff' : '#333',
              borderRadius: isMe ? '12px 2px 12px 12px' : '2px 12px 12px 12px',
              boxShadow: '0 1px 2px rgba(0,0,0,0.05)',
              wordBreak: 'break-word'
            }}>
              {msg.content}
            </div>
            <Text type="secondary" style={{ fontSize: 12, marginTop: 4 }}>
              {moment(msg.send_time).format('HH:mm')}
            </Text>
          </div>
        </div>
      );
    });
  };

  // 输入区（群聊 / 单聊共用）
  const renderComposer = (placeholder: string, disabled: boolean, onSend: () => void) => (
    <div style={{ background: '#fff', borderTop: '1px solid #f0f0f0', padding: '16px 24px' }}>
      <div style={{ display: 'flex', gap: 16 }}>
        <Input.TextArea
          value={inputValue}
          onChange={e => setInputValue(e.target.value)}
          onPressEnter={(e) => {
            if (!e.shiftKey) {
              e.preventDefault();
              onSend();
            }
          }}
          placeholder={placeholder}
          disabled={disabled}
          autoSize={{ minRows: 2, maxRows: 6 }}
          variant="borderless"
          style={{ background: '#f5f7fa', borderRadius: 8, padding: '8px 12px' }}
        />
        <Button
          type="primary"
          shape="circle"
          icon={<SendOutlined />}
          size="large"
          style={{ alignSelf: 'flex-end' }}
          onClick={onSend}
          disabled={!inputValue.trim() || disabled}
        />
      </div>
    </div>
  );

  return (
    <Layout style={{ height: '100vh', overflow: 'hidden' }}>
      {/* 极简侧边导航：头像与下方图标统一由 Sider 的水平居中约束对齐 */}
      <Sider width={64} style={{ background: '#001529', display: 'flex', flexDirection: 'column', alignItems: 'center', paddingTop: 20 }}>
        <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
          <Avatar
            size={40}
            src={avatarOf(currentUser.username)}
            style={{ cursor: 'pointer', marginBottom: 24, flexShrink: 0 }}
          />
        </Dropdown>

        {[
          { key: 'msg' as const, icon: <MessageOutlined />, label: '消息' },
          { key: 'contacts' as const, icon: <TeamOutlined />, label: '联系人' },
        ].map(item => (
          <Tooltip key={item.key} title={item.label} placement="right">
            <div
              onClick={() => switchNav(item.key)}
              style={{
                width: 44,
                height: 44,
                marginBottom: 12,
                borderRadius: 10,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: 20,
                cursor: 'pointer',
                color: activeNav === item.key ? '#fff' : 'rgba(255,255,255,0.65)',
                background: activeNav === item.key ? '#1890ff' : 'transparent',
                transition: 'all 0.2s',
              }}
            >
              {item.icon}
            </div>
          </Tooltip>
        ))}
      </Sider>

      {/* 列表栏：群聊 / 联系人 */}
      <Sider width={280} theme="light" style={{ borderRight: '1px solid #f0f0f0', overflow: 'auto' }}>
        <div style={{ padding: '16px 20px', borderBottom: '1px solid #f0f0f0' }}>
          <Space.Compact style={{ width: '100%' }}>
            <Input
              placeholder={activeNav === 'msg' ? '搜索群名称或群号加入' : '搜索好友'}
              value={searchText}
              onChange={e => setSearchText(e.target.value)}
              allowClear
            />
            {activeNav === 'msg' ? (
              <Button icon={<PlusOutlined />} title="创建群聊" onClick={() => setCreateGroupOpen(true)} />
            ) : (
              <Button icon={<UserAddOutlined />} title="添加好友" onClick={openAddFriend} />
            )}
          </Space.Compact>
        </div>

        {activeNav === 'msg' ? (
          <List
            itemLayout="horizontal"
            dataSource={filteredGroups}
            locale={{ emptyText: keyword ? '未找到匹配的群聊' : '暂未加入群聊，搜索群名称或群号即可加入' }}
            renderItem={item => {
              const selected = item.joined && activeGroupId === String(item.id);
              const preview = item.joined
                ? (previewText(item, currentUser.id) || `群号：${item.id}`)
                : `未加入 · 群号：${item.id}`;
              return (
                <List.Item
                  style={{
                    padding: '12px 20px',
                    cursor: item.joined ? 'pointer' : 'default',
                    background: selected ? '#e6f7ff' : 'transparent',
                    borderLeft: selected ? '3px solid #1890ff' : '3px solid transparent'
                  }}
                  onClick={() => item.joined && setActiveGroupId(String(item.id))}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 12, width: '100%' }}>
                    <Badge dot={hasGroupUnread(item)} offset={[-2, 4]}>
                      <Avatar src={avatarOf(item.name)} shape="square" size="large" />
                    </Badge>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 8 }}>
                        <Text strong type={item.joined ? undefined : 'secondary'} ellipsis style={{ flex: 1 }}>
                          {item.name}
                        </Text>
                        {item.joined && (
                          <Text type="secondary" style={{ fontSize: 12, flexShrink: 0 }}>
                            {formatListTime(item.last_time)}
                          </Text>
                        )}
                      </div>
                      <Text type="secondary" ellipsis style={{ fontSize: 12 }}>
                        {preview}
                      </Text>
                    </div>
                    {!item.joined && (
                      <Button
                        type="primary"
                        size="small"
                        onClick={(e) => { e.stopPropagation(); handleJoinGroup(item.id); }}
                      >
                        加入
                      </Button>
                    )}
                  </div>
                </List.Item>
              );
            }}
          />
        ) : (
          <>
            {applications.length > 0 && (
              <>
                <div style={{ padding: '12px 20px 4px' }}>
                  <Text type="secondary" style={{ fontSize: 12 }}>好友申请（{applications.length}）</Text>
                </div>
                <List
                  itemLayout="horizontal"
                  dataSource={applications}
                  renderItem={app => (
                    <List.Item style={{ padding: '10px 20px' }}>
                      <List.Item.Meta
                        avatar={<Avatar src={avatarOf(app.username)} />}
                        title={<Text strong>{app.username}</Text>}
                        description={<Text type="secondary" style={{ fontSize: 12 }}>{app.message}</Text>}
                      />
                      <Button type="primary" size="small" onClick={() => handleAcceptApplication(app.id)}>
                        同意
                      </Button>
                    </List.Item>
                  )}
                />
                <Divider style={{ margin: 0 }} />
              </>
            )}

            <div style={{ padding: '12px 20px 4px' }}>
              <Text type="secondary" style={{ fontSize: 12 }}>我的好友（{friends.length}）</Text>
            </div>
            <List
              itemLayout="horizontal"
              dataSource={filteredFriends}
              locale={{ emptyText: '还没有好友，点击右上角添加' }}
              renderItem={item => {
                const count = unread[item.id] || 0;
                const selected = activeFriendId === item.id;
                const preview = previewText(item, currentUser.id) || (item.online ? '在线' : '离线');
                return (
                  <List.Item
                    style={{
                      padding: '12px 20px',
                      cursor: 'pointer',
                      background: selected ? '#e6f7ff' : 'transparent',
                      borderLeft: selected ? '3px solid #1890ff' : '3px solid transparent',
                    }}
                    onClick={() => setActiveFriendId(item.id)}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12, width: '100%' }}>
                      <Badge count={count} size="small">
                        <div style={{ position: 'relative' }}>
                          <Avatar size="large" src={avatarOf(item.username)} />
                          {/* 在线状态小圆点：绿=在线，灰=离线 */}
                          <span style={{
                            position: 'absolute',
                            right: -1,
                            bottom: -1,
                            width: 10,
                            height: 10,
                            borderRadius: '50%',
                            background: item.online ? '#52c41a' : '#d9d9d9',
                            border: '2px solid #fff',
                          }} />
                        </div>
                      </Badge>
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 8 }}>
                          <Text strong ellipsis style={{ flex: 1 }}>{item.username}</Text>
                          <Text type="secondary" style={{ fontSize: 12, flexShrink: 0 }}>
                            {formatListTime(item.last_time)}
                          </Text>
                        </div>
                        <Text type="secondary" ellipsis style={{ fontSize: 12 }}>{preview}</Text>
                      </div>
                    </div>
                  </List.Item>
                );
              }}
            />
          </>
        )}
      </Sider>

      {/* 主区域 */}
      <Layout>
        {activeNav === 'msg' ? (
          <>
            <Header style={{ background: '#fff', padding: '0 24px', borderBottom: '1px solid #f0f0f0', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <Title level={4} style={{ margin: 0 }}>{activeGroup ? activeGroup.name : '未选择群聊'}</Title>
              <Space>
                <Text type={isConnected ? 'success' : 'danger'}>
                  {isConnected ? '● 在线' : '○ 离线'}
                </Text>
              </Space>
            </Header>

            <Content style={{ background: '#f5f7fa', display: 'flex', flexDirection: 'column', position: 'relative' }}>
              <div style={{ flex: 1, overflowY: 'auto', padding: '24px' }}>
                {!activeGroup ? (
                  <div style={{ display: 'flex', justifyContent: 'center', marginTop: 120 }}>
                    <Empty description="暂无群聊，先创建一个吧">
                      <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateGroupOpen(true)}>创建群聊</Button>
                    </Empty>
                  </div>
                ) : renderMessageList(messages)}
                <div ref={messagesEndRef} />
              </div>
              {renderComposer(
                activeGroup ? '输入消息，按 Enter 发送，Shift + Enter 换行' : '请先选择或创建一个群聊',
                !activeGroup,
                handleSend,
              )}
            </Content>
          </>
        ) : activeFriend ? (
          <>
            <Header style={{ background: '#fff', padding: '0 24px', borderBottom: '1px solid #f0f0f0', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <Space>
                <Avatar src={avatarOf(activeFriend.username)} />
                <Title level={4} style={{ margin: 0 }}>{activeFriend.username}</Title>
              </Space>
            </Header>
            <Content style={{ background: '#f5f7fa', display: 'flex', flexDirection: 'column', position: 'relative' }}>
              <div style={{ flex: 1, overflowY: 'auto', padding: '24px' }}>
                {renderMessageList(friendMessages)}
                <div ref={messagesEndRef} />
              </div>
              {renderComposer('输入消息，按 Enter 发送，Shift + Enter 换行', false, handleSendFriend)}
            </Content>
          </>
        ) : (
          <>
            <Header style={{ background: '#fff', padding: '0 24px', borderBottom: '1px solid #f0f0f0', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <Title level={4} style={{ margin: 0 }}>联系人</Title>
              <Button type="primary" icon={<UserAddOutlined />} onClick={openAddFriend}>添加好友</Button>
            </Header>
            <Content style={{ background: '#f5f7fa', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <Empty description="从左侧选择好友开始聊天" />
            </Content>
          </>
        )}
      </Layout>

      {/* 个人信息 */}
      <Modal
        title="个人信息"
        open={profileOpen}
        onCancel={() => setProfileOpen(false)}
        footer={<Button onClick={() => setProfileOpen(false)}>关闭</Button>}
      >
        <div style={{ textAlign: 'center', padding: '16px 0' }}>
          <Avatar size={72} src={avatarOf(currentUser.username)} />
          <Title level={4} style={{ marginTop: 16, marginBottom: 4 }}>{currentUser.username}</Title>
          <Text type="secondary">用户 ID：{currentUser.id}</Text>
        </div>
      </Modal>

      {/* 设置 */}
      <Modal
        title="设置"
        open={settingsOpen}
        onCancel={() => setSettingsOpen(false)}
        footer={<Button onClick={() => setSettingsOpen(false)}>关闭</Button>}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '12px 0' }}>
          <div>
            <div>自动滚动到最新消息</div>
            <Text type="secondary" style={{ fontSize: 12 }}>收到新消息时自动滚动到底部</Text>
          </div>
          <Switch checked={autoScroll} onChange={setAutoScroll} />
        </div>
        <Divider style={{ margin: '8px 0' }} />
        <Popconfirm title="确定要退出登录吗？" onConfirm={handleLogout} okText="退出" cancelText="取消">
          <Button danger block icon={<LogoutOutlined />}>退出登录</Button>
        </Popconfirm>
      </Modal>

      {/* 创建群聊 */}
      <Modal
        title="创建群聊"
        open={createGroupOpen}
        onCancel={() => { setCreateGroupOpen(false); groupForm.resetFields(); }}
        onOk={handleCreateGroup}
        okText="创建"
        confirmLoading={submitting}
      >
        <Form form={groupForm} layout="vertical">
          <Form.Item
            name="name"
            label="群聊名称"
            rules={[{ required: true, message: '请输入群聊名称' }]}
          >
            <Input
              placeholder="例如：Web程序设计交流群"
              maxLength={30}
              showCount
            />
          </Form.Item>
        </Form>
      </Modal>

      {/* 添加好友 */}
      <Modal
        title="添加好友"
        open={addFriendOpen}
        onCancel={() => setAddFriendOpen(false)}
        footer={<Button onClick={() => setAddFriendOpen(false)}>关闭</Button>}
      >
        <Input.Search
          placeholder="输入用户名搜索"
          enterButton={<SearchOutlined />}
          loading={searching}
          onSearch={handleSearchUsers}
          allowClear
        />
        <List
          style={{ marginTop: 16, maxHeight: 300, overflowY: 'auto' }}
          dataSource={searchResults}
          locale={{ emptyText: userKeyword ? '未找到该用户' : '输入用户名开始搜索' }}
          renderItem={item => (
            <List.Item
              actions={[
                isFriend(item.id) ? (
                  <Text type="secondary" key="added">已是好友</Text>
                ) : (
                  <Button
                    key="add"
                    type="link"
                    disabled={applications.some(a => a.sender_id === item.id)}
                    onClick={() => handleApplyFriend(item.id)}
                  >
                    {applications.some(a => a.sender_id === item.id) ? '已申请' : '添加'}
                  </Button>
                ),
              ]}
            >
              <List.Item.Meta
                avatar={<Avatar src={avatarOf(item.username)} />}
                title={<Text strong>{item.username}</Text>}
                description={<Text type="secondary" style={{ fontSize: 12 }}>用户 ID：{item.id}</Text>}
              />
            </List.Item>
          )}
        />
      </Modal>
    </Layout>
  );
};
