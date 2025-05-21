import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { dashboardStats, containers, scanResults, alerts, vulnerabilityTrend } from '@/lib/mockData';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, Legend } from 'recharts';
import { Shield, AlertTriangle, Bug, Server } from 'lucide-react';
import StatusCard from '@/components/dashboard/StatusCard';
import ContainerTable from '@/components/dashboard/ContainerTable';
import ScanReportTable from '@/components/dashboard/ScanReportTable';
import AlertsPreview from '@/components/dashboard/AlertsPreview';
import { useWebSocket, ResourceMetric } from '@/hooks/useWebSocket';

export default function DashboardPage() {
  const [activeTab, setActiveTab] = useState('overview');
  const [metrics, setMetrics] = useState<ResourceMetric[]>([]);
  const [containerStats, setContainerStats] = useState(containers);

  // Use wss:// for secure WebSocket connection
  const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;

  useWebSocket(wsUrl, (message) => {
    if (message.type === 'resource') {
      // Keep last 100 metrics for charts
      setMetrics(prev => [...prev.slice(-99), message.data]);
      
      // Update container stats
      setContainerStats(prev => prev.map(container => {
        if (container.id === message.data.containerId) {
          return {
            ...container,
            cpu: message.data.cpu,
            memory: message.data.memory,
          };
        }
        return container;
      }));
    }
  });

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 md:flex-row md:items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
          <p className="text-muted-foreground">
            Monitor and manage your container security from one place
          </p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatusCard
          title="Total Containers"
          value={dashboardStats.totalContainers.toString()}
          description={`${dashboardStats.runningContainers} running, ${dashboardStats.stoppedContainers} stopped`}
          icon={<Server className="h-5 w-5" />}
          trend="+2 since yesterday"
          trendUp={true}
        />
        <StatusCard
          title="Active Alerts"
          value={dashboardStats.totalAlerts.toString()}
          description={`${dashboardStats.criticalAlerts} critical, ${dashboardStats.highAlerts} high`}
          icon={<AlertTriangle className="h-5 w-5" />}
          trend="+5 since yesterday"
          trendUp={true}
          trendIsGood={false}
        />
        <StatusCard
          title="Total Vulnerabilities"
          value={dashboardStats.totalVulnerabilities.toString()}
          description={`${dashboardStats.criticalVulnerabilities} critical, ${dashboardStats.highVulnerabilities} high`}
          icon={<Bug className="h-5 w-5" />}
          trend="-8 since yesterday"
          trendUp={false}
          trendIsGood={true}
        />
        <StatusCard
          title="Security Score"
          value="75%"
          description="Good - up from 70% last week"
          icon={<Shield className="h-5 w-5" />}
          trend="+5% improvement"
          trendUp={true}
          trendIsGood={true}
        />
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle>Vulnerability Trend</CardTitle>
            <CardDescription>Last 7 days</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-[300px]">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={vulnerabilityTrend} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id="colorCritical" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="hsl(var(--destructive))" stopOpacity={0.8} />
                      <stop offset="95%" stopColor="hsl(var(--destructive))" stopOpacity={0.1} />
                    </linearGradient>
                    <linearGradient id="colorHigh" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="hsl(var(--chart-5))" stopOpacity={0.8} />
                      <stop offset="95%" stopColor="hsl(var(--chart-5))" stopOpacity={0.1} />
                    </linearGradient>
                    <linearGradient id="colorMedium" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="hsl(var(--chart-4))" stopOpacity={0.8} />
                      <stop offset="95%" stopColor="hsl(var(--chart-4))" stopOpacity={0.1} />
                    </linearGradient>
                    <linearGradient id="colorLow" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="hsl(var(--chart-2))" stopOpacity={0.8} />
                      <stop offset="95%" stopColor="hsl(var(--chart-2))" stopOpacity={0.1} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                  <XAxis dataKey="date" className="text-xs" />
                  <YAxis className="text-xs" />
                  <Tooltip 
                    contentStyle={{ 
                      backgroundColor: 'hsl(var(--card))', 
                      borderColor: 'hsl(var(--border))',
                      color: 'hsl(var(--foreground))'
                    }} 
                  />
                  <Area type="monotone" dataKey="critical" name="Critical" stroke="hsl(var(--destructive))" fillOpacity={1} fill="url(#colorCritical)" />
                  <Area type="monotone" dataKey="high" name="High" stroke="hsl(var(--chart-5))" fillOpacity={1} fill="url(#colorHigh)" />
                  <Area type="monotone" dataKey="medium" name="Medium" stroke="hsl(var(--chart-4))" fillOpacity={1} fill="url(#colorMedium)" />
                  <Area type="monotone" dataKey="low" name="Low" stroke="hsl(var(--chart-2))" fillOpacity={1} fill="url(#colorLow)" />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <AlertsPreview alerts={alerts} />
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="overview">Container Status</TabsTrigger>
          <TabsTrigger value="scan">Scan Reports</TabsTrigger>
        </TabsList>
        <TabsContent value="overview" className="mt-4">
          <ContainerTable containers={containerStats} />
        </TabsContent>
        <TabsContent value="scan" className="mt-4">
          <ScanReportTable scanResults={scanResults} />
        </TabsContent>
      </Tabs>
    </div>
  );
}