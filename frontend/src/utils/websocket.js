import { MESSAGE_TYPES } from '../types/game';

const API_BASE_URL = 'http://localhost:8080/api';
const WS_BASE_URL = 'ws://localhost:8080/ws';
const RECONNECT_INTERVAL = 3000;
const MAX_RECONNECT_ATTEMPTS = 5;

class WebSocketManager {
  constructor() {
    this.ws = null;
    this.listeners = new Map();
    this.messageId = 0;
    this.pendingRequests = new Map();
    this.reconnectAttempts = 0;
    this.isManualClose = false;
    this.heartbeatInterval = null;
    this.roomId = null;
    this.roomName = null;
    this.playerId = null;
    this.nickname = null;
  }

  async login(nickname) {
    const response = await fetch(`${API_BASE_URL}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ player_name: nickname }),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: '登录失败' }));
      throw new Error(error.error || '登录失败');
    }
    const data = await response.json();
    this.playerId = data.player_id;
    this.nickname = data.player_name;
    return data;
  }

  async listRooms() {
    const response = await fetch(`${API_BASE_URL}/rooms`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: '获取房间列表失败' }));
      throw new Error(error.error || '获取房间列表失败');
    }
    return response.json();
  }

  async getRoom(roomId) {
    const response = await fetch(`${API_BASE_URL}/rooms/${roomId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: '获取房间信息失败' }));
      throw new Error(error.error || '获取房间信息失败');
    }
    return response.json();
  }

  async createRoom(roomName) {
    const response = await fetch(`${API_BASE_URL}/rooms`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        room_name: roomName,
        player_name: this.nickname,
        player_id: this.playerId,
      }),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: '创建房间失败' }));
      throw new Error(error.error || '创建房间失败');
    }
    const data = await response.json();
    this.roomId = data.room_id;
    this.roomName = roomName;

    const roomDetail = await this.getRoom(data.room_id);
    return {
      ...data,
      room_name: roomName,
      players: roomDetail.players || [],
      status: roomDetail.status,
    };
  }

  async joinRoom(roomId) {
    const response = await fetch(`${API_BASE_URL}/rooms/${roomId}/join`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        player_name: this.nickname,
        player_id: this.playerId,
      }),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: '加入房间失败' }));
      throw new Error(error.error || '加入房间失败');
    }
    const data = await response.json();
    this.roomId = data.room_id;

    const roomDetail = await this.getRoom(data.room_id);
    this.roomName = roomDetail.name || '游戏房间';
    return {
      ...data,
      room_name: this.roomName,
      players: roomDetail.players || [],
      status: roomDetail.status,
    };
  }

  async startGame(roomId) {
    const response = await fetch(`${API_BASE_URL}/rooms/${roomId}/start`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: '开始游戏失败' }));
      throw new Error(error.error || '开始游戏失败');
    }
    return response.json();
  }

  connect() {
    return new Promise((resolve, reject) => {
      try {
        if (!this.roomId || !this.playerId) {
          reject(new Error('需要先加入房间才能连接WebSocket'));
          return;
        }

        const wsUrl = `${WS_BASE_URL}/${this.roomId}/${this.playerId}`;
        console.log('[WebSocket] 连接到:', wsUrl);
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          console.log('[WebSocket] 连接成功');
          this.reconnectAttempts = 0;
          this.isManualClose = false;
          this.startHeartbeat();
          this.emit('open', { success: true });
          this.send(MESSAGE_TYPES.GAME_STATE, {});
          resolve({ success: true });
        };

        this.ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (e) {
            console.error('[WebSocket] 消息解析失败:', e);
          }
        };

        this.ws.onerror = (error) => {
          console.error('[WebSocket] 连接错误:', error);
        };

        this.ws.onclose = (event) => {
          console.log('[WebSocket] 连接关闭:', event.code, event.reason);
          this.stopHeartbeat();

          if (!this.isManualClose && this.reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
            this.reconnectAttempts++;
            console.log(`[WebSocket] 尝试重连 (${this.reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})`);
            setTimeout(() => this.connect(), RECONNECT_INTERVAL);
          }

          this.emit('close', event);
        };
      } catch (e) {
        reject(e);
      }
    });
  }

  disconnect() {
    this.isManualClose = true;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.stopHeartbeat();
  }

  startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.send(MESSAGE_TYPES.HEARTBEAT, {});
      }
    }, 30000);
  }

  stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  setRoomId(roomId) {
    this.roomId = roomId;
  }

  setPlayerId(playerId) {
    this.playerId = playerId;
  }

  send(type, data = {}) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn('[WebSocket] 连接未就绪，无法发送消息');
      return null;
    }

    const message = {
      type,
      room_id: this.roomId,
      player_id: this.playerId,
      data,
    };

    this.ws.send(JSON.stringify(message));
    return ++this.messageId;
  }

  sendPlayerAction(action, params = {}) {
    return this.send(MESSAGE_TYPES.PLAYER_ACTION, {
      action,
      params,
    });
  }

  sendWithAck(type, ackType, data = {}, timeout = 10000) {
    return new Promise((resolve, reject) => {
      const messageId = this.send(type, data);
      if (!messageId) {
        reject(new Error('WebSocket未连接'));
        return;
      }

      const timer = setTimeout(() => {
        this.pendingRequests.delete(messageId);
        reject(new Error('请求超时'));
      }, timeout);

      this.pendingRequests.set(messageId, {
        ackType,
        resolve,
        reject,
        timer,
      });
    });
  }

  handleMessage(message) {
    const { type, data, room_id, player_id } = message;

    if (type === MESSAGE_TYPES.HEARTBEAT) {
      return;
    }

    if (room_id) {
      this.roomId = room_id;
    }

    if (this.pendingRequests.size > 0) {
      for (const [id, pending] of this.pendingRequests) {
        if (pending.ackType === type) {
          clearTimeout(pending.timer);
          this.pendingRequests.delete(id);
          if (data && data.success === false) {
            pending.reject(new Error(data.message || '操作失败'));
          } else {
            pending.resolve(data);
          }
          return;
        }
      }
    }

    this.emit(type, data);
  }

  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event).add(callback);
    return () => this.off(event, callback);
  }

  once(event, callback) {
    const handler = (data) => {
      this.off(event, handler);
      callback(data);
    };
    return this.on(event, handler);
  }

  off(event, callback) {
    if (this.listeners.has(event)) {
      this.listeners.get(event).delete(callback);
    }
  }

  emit(event, data) {
    if (this.listeners.has(event)) {
      this.listeners.get(event).forEach((callback) => {
        try {
          callback(data);
        } catch (e) {
          console.error('[WebSocket] 事件处理错误:', e);
        }
      });
    }
  }

  getReadyState() {
    return this.ws ? this.ws.readyState : WebSocket.CLOSED;
  }

  isConnected() {
    return this.ws && this.ws.readyState === WebSocket.OPEN;
  }
}

const wsManager = new WebSocketManager();

export default wsManager;
