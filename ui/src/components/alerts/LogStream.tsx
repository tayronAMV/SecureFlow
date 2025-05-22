import { useState, useEffect, useRef } from 'react';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';

interface LogStreamProps {
  logs: string[];
}

export default function LogStream({ logs }: LogStreamProps) {
  const [autoScroll, setAutoScroll] = useState(true);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (autoScroll && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs, autoScroll]);

  return (
    <Card className="overflow-hidden">
      <div className="border-b bg-muted/50 px-4 py-2">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-medium">Real-time logs</h3>
          <Badge 
            variant="outline" 
            className="cursor-pointer"
            onClick={() => setAutoScroll(!autoScroll)}
          >
            Autoscroll: {autoScroll ? 'ON' : 'OFF'}
          </Badge>
        </div>
      </div>
      
      <ScrollArea className="h-[500px] w-full">
        <div className="space-y-1 p-4 font-mono text-xs">
          {logs.map((log, index) => (
            <div 
              key={index} 
              className="px-2 py-1.5 hover:bg-muted/50 break-all"
            >
              {log}
            </div>
          ))}
          <div ref={bottomRef} />
        </div>
      </ScrollArea>
    </Card>
  );
}
