import React, { createContext, useContext, useEffect, useRef, useState } from 'react';

export interface LogMessage {
  id: number;
  message: string;
}

type SubscriberCallback = (message: LogMessage) => void;

interface Subscriber {
  callback: SubscriberCallback;
  channels: number[] | null; // null means subscribe to all channels
}

interface WebSocketContextType {
  subscribe: (callback: SubscriberCallback, channels: number[] | null) => () => void;
  connectionStatus: 'connected' | 'connecting' | 'disconnected';
  getMessageHistory: (channelId: number) => string[];
  clearMessageHistory: (channelId: number) => void;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

const WS_URL = "ws://localhost:8080/ws";
const MAX_RETRIES = 5;
const MAX_HISTORY_MESSAGES = 1000; // Maximum number of messages to keep per channel

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const [connectionStatus, setConnectionStatus] = useState<'connected' | 'connecting' | 'disconnected'>('disconnected');
  const wsRef = useRef<WebSocket | null>(null);
  const subscribersRef = useRef<Set<Subscriber>>(new Set());
  const retriesRef = useRef(0);
  
  // Store message history by channel ID
  const messageHistoryRef = useRef<Map<number, string[]>>(new Map());

  // Helper function to add a message to history for a specific channel
  const addMessageToHistory = (channelId: number, message: string) => {
    const currentHistory = messageHistoryRef.current.get(channelId) || [];
    const newHistory = [...currentHistory, message].slice(-MAX_HISTORY_MESSAGES);
    messageHistoryRef.current.set(channelId, newHistory);
  };

  const connect = () => {
    try {
      if (wsRef.current) {
        wsRef.current.close();
      }

      setConnectionStatus('connecting');
      const ws = new WebSocket(WS_URL);
      wsRef.current = ws;

      ws.onmessage = (event) => {
        try {
          // Try to parse as JSON first
          const data = JSON.parse(event.data) as LogMessage;
          
          // Store message in history for its channel
          addMessageToHistory(data.id, data.message);
          
          // Route message only to subscribers that care about this channel
          subscribersRef.current.forEach(subscriber => {
            if (
              subscriber.channels === null || // Subscriber wants all messages
              subscriber.channels.includes(data.id) // Subscriber wants this specific channel
            ) {
              subscriber.callback(data);
            }
          });
        } catch (err) {
          // If parsing fails, treat as plain string with default id of 0
          console.warn('Received non-JSON message:', event.data);
          const fallbackMessage: LogMessage = { 
            id: 0, 
            message: event.data 
          };
          
          // Store fallback message in history
          addMessageToHistory(0, event.data);
          
          // Send fallback message to everyone who subscribes to channel 0 or all channels
          subscribersRef.current.forEach(subscriber => {
            if (subscriber.channels === null || subscriber.channels.includes(0)) {
              subscriber.callback(fallbackMessage);
            }
          });
        }
      };

      ws.onopen = () => {
        console.log('WebSocket connected');
        setConnectionStatus('connected');
        retriesRef.current = 0;
      };

      ws.onerror = (err) => {
        console.error('WebSocket error:', err);
      };

      ws.onclose = () => {
        console.warn('WebSocket closed');
        setConnectionStatus('disconnected');
        if (retriesRef.current < MAX_RETRIES) {
          const timeout = Math.min(1000 * 2 ** retriesRef.current, 10000);
          retriesRef.current++;
          console.log(`Reconnecting in ${timeout}ms... (Attempt ${retriesRef.current})`);
          setTimeout(connect, timeout);
        }
      };
    } catch (err) {
      console.error('Failed to create WebSocket connection:', err);
      setConnectionStatus('disconnected');
    }
  };

  useEffect(() => {
    connect();
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const subscribe = (callback: SubscriberCallback, channels: number[] | null) => {
    const subscriber: Subscriber = { callback, channels };
    subscribersRef.current.add(subscriber);
    
    return () => {
      subscribersRef.current.delete(subscriber);
    };
  };

  // Get message history for a specific channel
  const getMessageHistory = (channelId: number): string[] => {
    return messageHistoryRef.current.get(channelId) || [];
  };

  // Clear message history for a specific channel
  const clearMessageHistory = (channelId: number): void => {
    messageHistoryRef.current.set(channelId, []);
  };

  return (
    <WebSocketContext.Provider value={{ 
      subscribe, 
      connectionStatus, 
      getMessageHistory,
      clearMessageHistory 
    }}>
      {children}
    </WebSocketContext.Provider>
  );
}

export function useWebSocketSubscription(
  onMessage: (message: LogMessage) => void,
  channels: number[] | null = null // null means subscribe to all channels
) {
  const context = useContext(WebSocketContext);
  
  if (!context) {
    throw new Error('useWebSocketSubscription must be used within a WebSocketProvider');
  }

  useEffect(() => {
    return context.subscribe(onMessage, channels);
  }, [onMessage, channels, context]);

  return context.connectionStatus;
}

export function useMessageHistory(channelId: number) {
  const context = useContext(WebSocketContext);
  
  if (!context) {
    throw new Error('useMessageHistory must be used within a WebSocketProvider');
  }

  const [messages, setMessages] = useState<string[]>([]);

  useEffect(() => {
    // Initialize with current history
    setMessages(context.getMessageHistory(channelId));
  }, [channelId, context]);

  return {
    messages,
    clearHistory: () => {
      context.clearMessageHistory(channelId);
      setMessages([]);
    }
  };
} 