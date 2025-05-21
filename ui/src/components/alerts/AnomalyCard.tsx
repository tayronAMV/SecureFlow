import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { formatDistanceToNow } from 'date-fns';
import { AlertTriangle, AlertCircle, ArrowRightIcon } from 'lucide-react';

interface Alert {
  id: string;
  containerId: string;
  timestamp: string;
  severity: string;
  message: string;
  details: string;
}

interface AnomalyCardProps {
  alert: Alert;
}

export default function AnomalyCard({ alert }: AnomalyCardProps) {
  return (
    <Card className={cn(
      alert.severity === 'critical' ? "border-red-500/50" : "border-orange-500/50"
    )}>
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <CardTitle className="flex items-center gap-2">
              {alert.severity === 'critical' ? (
                <AlertCircle className="h-5 w-5 text-red-500" />
              ) : (
                <AlertTriangle className="h-5 w-5 text-orange-500" />
              )}
              <span>Security Anomaly Detected</span>
            </CardTitle>
            <CardDescription>Container: {alert.containerId.replace('c-', '')}</CardDescription>
          </div>
          <Badge 
            className={cn(
              "ml-2 px-2",
              alert.severity === 'critical' 
                ? "bg-red-500/20 text-red-600 hover:bg-red-500/30 dark:bg-red-500/20 dark:text-red-400 dark:hover:bg-red-500/30" 
                : "bg-orange-500/20 text-orange-600 hover:bg-orange-500/30 dark:bg-orange-500/20 dark:text-orange-400 dark:hover:bg-orange-500/30"
            )}
          >
            {alert.severity}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <h4 className="font-semibold">{alert.message}</h4>
          <p className="text-sm text-muted-foreground">{alert.details}</p>
        </div>
      </CardContent>
      <CardFooter className="flex items-center justify-between border-t p-4 pt-3 text-xs text-muted-foreground">
        <span>{formatDistanceToNow(new Date(alert.timestamp), { addSuffix: true })}</span>
        <Button variant="ghost" size="sm" className="h-7 gap-1 text-xs">
          Investigate <ArrowRightIcon className="h-3 w-3" />
        </Button>
      </CardFooter>
    </Card>
  );
}