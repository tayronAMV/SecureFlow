import { useState, useCallback, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Shield, AlertTriangle, Bug, Server } from 'lucide-react';
import StatusCard from '@/components/dashboard/StatusCard';
import ContainerTable from '@/components/dashboard/ContainerTable';
import { useWebSocketSubscription, LogMessage, useMessageHistory } from '@/contexts/WebSocketContext';

export default function DashboardPage() {
  const [activeTab, setActiveTab] = useState('overview');
  
  // Use message history hook for channel 2 (scan reports)
  const { messages: reportHistory } = useMessageHistory(2);
  const [scanReports, setScanReports] = useState<string[]>([]);
  
  // Initialize scan reports from history when component mounts or history changes
  useEffect(() => {
    setScanReports(reportHistory);
  }, [reportHistory]);

  const handleMessage = useCallback((msg: LogMessage) => {
    // Channel filtering is done in the WebSocketContext
    // Here we assume messages with ID 2 are for scan reports
    setScanReports(prev => [...prev, msg.message]);
  }, []);

  // Subscribe only to messages with id=2 (for scan reports)
  // Note: this assumes that scan reports will have id=2
  const connectionStatus = useWebSocketSubscription(handleMessage, [2]);

  // Placeholder data for containers (this would be replaced with real data)
  const containers = scanReports.map((log, i) => ({
    id: `container-${i}`,
    name: `container-${i}`,
    status: 'running',
    image: 'N/A',
    created: new Date().toISOString(),
    cpu: 0,
    memory: 0,
    anomalyScore: 0,
    vulnerabilities: 0,
    rawLog: log
  }));

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 md:flex-row md:items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
          <p className="text-muted-foreground">
            Monitor and manage your container security from one place ({connectionStatus})
          </p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatusCard
          title="Total Containers"
          value={containers.length.toString()}
          description={`${containers.length} running`}
          icon={<Server className="h-5 w-5" />}
          trend="Real-time monitoring"
          trendUp={true}
        />
        <StatusCard
          title="CPU Usage"
          value="N/A"
          description="Simplified log mode"
          icon={<AlertTriangle className="h-5 w-5" />}
          trend="Static"
          trendUp={false}
        />
        <StatusCard
          title="Memory Usage"
          value="N/A"
          description="Simplified log mode"
          icon={<Bug className="h-5 w-5" />}
          trend="Static"
          trendUp={false}
        />
        <StatusCard
          title="Security Status"
          value="Monitoring"
          description="Real-time security analysis"
          icon={<Shield className="h-5 w-5" />}
          trend="Active"
          trendUp={true}
        />
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="overview">Container Status</TabsTrigger>
          <TabsTrigger value="metrics">Scan Reports</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="mt-4">
          <ContainerTable containers={containers} />
        </TabsContent>

        <TabsContent value="metrics" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Scan Reports</CardTitle>
              <CardDescription>Security scan results will appear here</CardDescription>
            </CardHeader>
            <CardContent>
              {scanReports.length === 0 ? (
                <div className="text-center text-muted-foreground py-8">
                  No scan reports available...
                </div>
              ) : (
                <div className="space-y-2 font-mono text-sm">
                  {scanReports.map((report, i) => (
                    <div key={i} className="bg-muted px-3 py-1 rounded">{report}</div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
