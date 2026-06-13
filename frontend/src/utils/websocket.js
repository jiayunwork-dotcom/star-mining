import { MESSAGE_TYPES } from '../types/game';

const WS_URL = 'ws://localhost:8080/ws';
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
    this.playerId = null;
  }

  connect(nickname) {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(WS_URL);
        this.nickname = nickname;

        this.ws.onopen = () => {
          console.log('[WebSocket] 连接成功');
          this.reconnectAttempts = 0;
          this.startHeartbeat();
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
            setTimeout(() => this.connect(this.nickname), RECONNECT_INTERVAL);
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

    if (player_id) {
      this.playerId = player_id;
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
