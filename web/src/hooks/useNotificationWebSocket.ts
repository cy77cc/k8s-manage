import { useRef, useCallback, useEffect } from 'react';
import type { WSMessage, WSConnectionStatus } from '../types/notification';

interface UseNotificationWebSocketOptions {
  userId?: number | string;
  onMessage?: (message: WSMessage) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  reconnectInterval?: number;
  maxReconnectInterval?: number;
}

export const useNotificationWebSocket = (options: UseNotificationWebSocketOptions = {}) => {
  const {
    userId,
    onMessage,
    onConnect,
    onDisconnect,
    reconnectInterval = 1000,
    maxReconnectInterval = 30000,
  } = options;

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const currentReconnectIntervalRef = useRef(reconnectInterval);
  const statusRef = useRef<WSConnectionStatus>('disconnected');
  const broadcastChannelRef = useRef<BroadcastChannel | null>(null);

  // 多标签页同步
  const initBroadcastChannel = useCallback(() => {
    if (typeof BroadcastChannel !== 'undefined') {
      broadcastChannelRef.current = new BroadcastChannel('notification-sync');
      broadcastChannelRef.current.onmessage = (event) => {
        if (event.data.type === 'notification-update') {
          onMessage?.(event.data.message);
        }
      };
    }
  }, [onMessage]);

  // 广播消息到其他标签页
  const broadcast = useCallback((message: WSMessage) => {
    if (broadcastChannelRef.current) {
      broadcastChannelRef.current.postMessage({
        type: 'notification-update',
        message,
      });
    }
  }, []);

  // 连接 WebSocket
  const connect = useCallback(() => {
    if (!userId || statusRef.current === 'connecting') {
      return;
    }

    statusRef.current = 'connecting';

    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsHost = window.location.host;
    const wsUrl = `${wsProtocol}//${wsHost}/ws/notifications?user_id=${userId}`;

    try {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        statusRef.current = 'connected';
        currentReconnectIntervalRef.current = reconnectInterval;
        onConnect?.();
        console.log('WebSocket: 已连接');
      };

      ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data);
          onMessage?.(message);
          broadcast(message);
        } catch (error) {
          console.error('WebSocket: 解析消息失败', error);
        }
      };

      ws.onclose = () => {
        statusRef.current = 'disconnected';
        onDisconnect?.();
        console.log('WebSocket: 连接关闭');

        // 自动重连
        scheduleReconnect();
      };

      ws.onerror = (error) => {
        console.error('WebSocket: 连接错误', error);
        ws.close();
      };
    } catch (error) {
      console.error('WebSocket: 创建连接失败', error);
      statusRef.current = 'disconnected';
      scheduleReconnect();
    }
  }, [userId, onMessage, onConnect, onDisconnect, reconnectInterval, broadcast]);

  // 安排重连
  const scheduleReconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }

    const delay = currentReconnectIntervalRef.current;
    currentReconnectIntervalRef.current = Math.min(
      currentReconnectIntervalRef.current * 2,
      maxReconnectInterval
    );

    reconnectTimeoutRef.current = setTimeout(() => {
      console.log(`WebSocket: 尝试重连 (延迟 ${delay}ms)`);
      connect();
    }, delay);
  }, [connect, maxReconnectInterval]);

  // 断开连接
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    statusRef.current = 'disconnected';
  }, []);

  // 发送消息
  const send = useCallback((data: object) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data));
    }
  }, []);

  // 初始化
  useEffect(() => {
    initBroadcastChannel();
    if (userId) {
      connect();
    }

    return () => {
      disconnect();
      if (broadcastChannelRef.current) {
        broadcastChannelRef.current.close();
      }
    };
  }, [userId, connect, disconnect, initBroadcastChannel]);

  return {
    connect,
    disconnect,
    send,
    get status() {
      return statusRef.current;
    },
  };
};

export default useNotificationWebSocket;
