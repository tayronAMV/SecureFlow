import { useState, useEffect, useRef } from 'react';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { format } from 'date-fns';
import { Card } from '@/components/ui/card';
import { Check, AlertTriangle, AlertCircle, Info } from 'lucide-react';
import { ScrollArea } from '@/components/ui/scroll-area';

interface Log {
  id: string;
  timestamp: string;
  level: string;
  source: string;
  message: string;
}

interface LogStreamProps {
  logs: Log[];
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
          {logs.map((log) => (
            <div 
              key={log.id} 
              className={cn(
                "flex items-start rounded px-2 py-1.5",
                log.level === 'critical' ? "bg-red-500/10" :
                log.level === 'error' ? "bg-red-500/5" :
                log.level === 'warning' ? "bg-yellow-500/5" :
                "hover:bg-muted/50"
              )}
            >
              <div className="mr-2 mt-0.5">
                {log.level === 'critical' ? (
                  <AlertCircle className="h-3 w-3 text-red-500" />
                ) : log.level === 'error' ? (
                  <AlertCircle className="h-3 w-3 text-red-400" />
                ) : log.level === 'warning' ? (
                  <AlertTriangle className="h-3 w-3 text-yellow-500" />
                ) : (
                  <Info className="h-3 w-3 text-blue-400" />
                )}
              </div>
              <div className="flex-1">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-muted-foreground">
                    {format(new Date(log.timestamp), 'HH:mm:ss')}
                  </span>
                  <Badge 
                    variant="outline" 
                    className={cn(
                      "text-[10px]",
                      log.level === 'critical' ? "border-red-500 text-red-500" :
                      log.level === 'error' ? "border-red-400 text-red-400" :
                      log.level === 'warning' ? "border-yellow-500 text-yellow-500" :
                      "border-blue-400 text-blue-400"
                    )}
                  >
                    {log.level.toUpperCase()}
                  </Badge>
                  <Badge variant="secondary" className="text-[10px]">
                    {log.source}
                  </Badge>
                </div>
                <div className="mt-1 pr-4">{log.message}</div>
              </div>
            </div>
          ))}
          <div ref={bottomRef} />
        </div>
      </ScrollArea>
    </Card>
  );
}