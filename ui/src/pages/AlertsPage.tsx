import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { format } from 'date-fns';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { AlertCircle, AlertTriangle, Info, CheckCircle2, Search, Filter } from 'lucide-react';
import LogStream from '@/components/alerts/LogStream';
import AnomalyCard from '@/components/alerts/AnomalyCard';
import { useWebSocket, FlowEvent, Anomaly } from '@/hooks/useWebSocket';

export default function AlertsPage() {
  const [activeTab, setActiveTab] = useState('logs');
  const [searchQuery, setSearchQuery] = useState('');
  const [logs, setLogs] = useState<FlowEvent[]>([]);
  const [anomalies, setAnomalies] = useState<Anomaly[]>([]);

  // Use wss:// for secure WebSocket connection
  const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;
  
  useWebSocket(wsUrl, (message) => {
    if (message.type === 'log') {
      setLogs(prev => [...prev.slice(-999), message.data]);
    } else if (message.type === 'anomaly') {
      setAnomalies(prev => [...prev.slice(-99), message.data]);
    }
  });

  // Filter logs based on search query
  const filteredLogs = logs.filter(log => 
    log.message.toLowerCase().includes(searchQuery.toLowerCase()) ||
    log.source.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Filter critical and high severity anomalies
  const criticalAnomalies = anomalies.filter(anomaly => 
    anomaly.severity === 'critical' || anomaly.severity === 'high'
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 md:flex-row md:items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Alerts & Logs</h2>
          <p className="text-muted-foreground">
            Monitor security events and system logs in real-time
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