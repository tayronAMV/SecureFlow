import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { formatDistanceToNow } from 'date-fns';
import { AlertCircle, AlertTriangle, Info, CheckCircle2 } from 'lucide-react';

interface Alert {
  id: string;
  containerId: string;
  timestamp: string;
  severity: string;
  message: string;
  details: string;
}

interface AlertsPreviewProps {
  alerts: Alert[];
}

export default function AlertsPreview({ alerts }: AlertsPreviewProps) {
  // Sort alerts by timestamp (most recent first) and limit to 4
  const recentAlerts = [...alerts]
    .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
    .slice(0, 4);

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="flex items-center justify-between">
          <span>Recent Alerts</span>
          <Badge variant="outline" className="ml-2 text-xs font-normal">
            {alerts.length} total
          </Badge>
        </CardTitle>
        <CardDescription>Latest security alerts detected by SecureFlow</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {recentAlerts.map((alert) => (
            <div key={alert.id} className="flex gap-3">
              <div 
                className={cn(
                  "mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full",
                  alert.severity === 'critical' ? "bg-red-500/20 text-red-600 dark:bg-red-500/20 dark:text-red-400" :
                  alert.severity === 'high' ? "bg-orange-500/20 text-orange-600 dark:bg-orange-500/20 dark:text-orange-400" :
                  alert.severity === 'medium' ? "bg-yellow-500/20 text-yellow-600 dark:bg-yellow-500/20 dark:text-yellow-400" :
                  "bg-blue-500/20 text-blue-600 dark:bg-blue-500/20 dark:text-blue-400"
                )}
              >
                {alert.severity === 'critical' ? (
                  <AlertCircle className="h-4 w-4" />
                ) : alert.severity === 'high' ? (
                  <AlertTriangle className="h-4 w-4" />
                ) : alert.severity === 'medium' ? (
                  <AlertTriangle className="h-4 w-4" />
                ) : (
                  <Info className="h-4 w-4" />
                )}
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h4 className="font-medium leading-none">{alert.message}</h4>
                  <Badge 
                    variant="outline" 
                    className={cn(
                      "text-xs",
                      alert.severity === 'critical' ? "border-red-500 text-red-500" :
                      alert.severity === 'high' ? "border-orange-500 text-orange-500" :
                      alert.severity === 'medium' ? "border-yellow-500 text-yellow-500" :
                      "border-blue-500 text-blue-500"
                    )}
                  >
                    {alert.severity}
                  </Badge>
                </div>
                <p className="mt-1 text-sm text-muted-foreground">{alert.details}</p>
                <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                  <span>Container: {alert.containerId.replace('c-', '')}</span>
                  <span>•</span>
                  <span>{formatDistanceToNow(new Date(alert.timestamp), { addSuffix: true })}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}