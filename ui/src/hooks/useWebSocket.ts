import { useEffect, useRef, useCallback } from 'react';

export interface FlowEvent {
  id: string;
  timestamp: string;
  level: 'info' | 'warning' | 'error' | 'critical';
  source: string;
  message: string;
}

export interface Anomaly {
  id: string;
  containerId: string;
  timestamp: string;
  severity: 'critical' | 'high' | 'medium' | 'low';
  message: string;
  details: string;
}

export interface ResourceMetric {
  containerId: string;
  timestamp: string;
  cpu: number;
  memory: number;
  network: {
    rx_bytes: number;
    tx_bytes: number;
  };
}

type IncomingMessage =
  | { type: 'log'; data: FlowEvent }
  | { type: 'anomaly'; data: Anomaly }
  | { type: 'resource'; data: ResourceMetric };

export function useWebSocket(
  url: string,
  onMessage: (msg: IncomingMessage) => void,
  maxRetries = 5
) {
  const socketRef = useRef<WebSocket | null>(null);
  const retriesRef = useRef(0);

  const connect = useCallback(() => {
    try {
      const ws = new WebSocket(url);
      socketRef.current = ws;

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as IncomingMessage;
          onMessage(message);
        } catch (err) {
          console.error('Invalid WebSocket message', err);
        }
      };

      ws.onopen = () => {
        console.log('WebSocket connected');
        retriesRef.current = 0; // Reset retry counter on successful connection
      };

      ws.onerror = (err) => {
        console.error('WebSocket error:', err);
      };

      ws.onclose = () => {
        console.warn('WebSocket closed');
        
        // Attempt to reconnect if we haven't exceeded max retries
        if (retriesRef.current < maxRetries) {
          const timeout = Math.min(1000 * Math.pow(2, retriesRef.current), 10000);
          retriesRef.current++;
          
          console.log(`Reconnecting in ${timeout}ms... (Attempt ${retriesRef.current})`);
          setTimeout(connect, timeout);
        }
      };
    } catch (err) {
      console.error('Failed to create WebSocket connection:', err);
    }
  }, [url, onMessage, maxRetries]);

  useEffect(() => {
    connect();

    return () => {
      if (socketRef.current) {
        socketRef.current.close();
      }
    };
  }, [connect]);

  // Return a function to manually reconnect if needed
  return {
    reconnect: () => {
      if (socketRef.current) {
        socketRef.current.close();
      }
      retriesRef.current = 0;
      connect();
    }
  };
}