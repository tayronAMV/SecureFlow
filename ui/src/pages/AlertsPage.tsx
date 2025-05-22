import { useState, useCallback, useEffect } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Search, Filter } from 'lucide-react';
import LogStream from '@/components/alerts/LogStream';
import AnomalyCard from '@/components/alerts/AnomalyCard';
import { useWebSocketSubscription, LogMessage, useMessageHistory } from '@/contexts/WebSocketContext';

export default function AlertsPage() {
  const [activeTab, setActiveTab] = useState('logs');
  const [searchQuery, setSearchQuery] = useState('');
  
  // Use the message history hook to get persistent logs for channel 1
  const { messages: logHistory } = useMessageHistory(1);
  const [logs, setLogs] = useState<string[]>([]);

  // Initialize logs from history when the component mounts or history changes
  useEffect(() => {
    setLogs(logHistory);
  }, [logHistory]);

  const [anomalies, setAnomalies] = useState<any[]>([]); // Anomaly logic placeholder for later

  const handleMessage = useCallback((msg: LogMessage) => {
    // With channel filtering in the context, we only receive messages with id=1 here
    setLogs(prev => [...prev, msg.message]);
  }, []);

  // Subscribe only to messages with id=1
  const connectionStatus = useWebSocketSubscription(handleMessage, [1]);

  const filteredLogs = logs.filter(log =>
    log.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const criticalAnomalies = anomalies.filter(anomaly =>
    anomaly.severity === 'critical' || anomaly.severity === 'high'
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 md:flex-row md:items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Alerts & Logs</h2>
          <p className="text-muted-foreground">
            Monitor security events and system logs in real-time ({connectionStatus})
          </p>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="logs">Logs Stream</TabsTrigger>
          <TabsTrigger value="anomalies">Anomaly Updates</TabsTrigger>
        </TabsList>

        <TabsContent value="logs" className="mt-4 space-y-4">
          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search logs..."
                className="pl-8"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
            <Button variant="outline" size="icon">
              <Filter className="h-4 w-4" />
            </Button>
          </div>

          <LogStream logs={filteredLogs} />
        </TabsContent>

        <TabsContent value="anomalies" className="mt-4">
          <div className="grid gap-4 sm:grid-cols-1 lg:grid-cols-2">
            {criticalAnomalies.map(alert => (
              <AnomalyCard key={alert.id} alert={alert} />
            ))}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
