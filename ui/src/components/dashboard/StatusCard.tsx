import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { ArrowDown, ArrowUp } from 'lucide-react';

interface StatusCardProps {
  title: string;
  value: string;
  description: string;
  icon: React.ReactNode;
  trend: string;
  trendUp: boolean;
  trendIsGood?: boolean;
}

export default function StatusCard({
  title,
  value,
  description,
  icon,
  trend,
  trendUp,
  trendIsGood = true,
}: StatusCardProps) {
  // Determine if the trend is good or bad
  const showGoodTrend = (trendUp && trendIsGood) || (!trendUp && !trendIsGood);

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <div className="h-8 w-8 rounded-full bg-primary/10 p-1.5 text-primary">{icon}</div>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        <p className="text-xs text-muted-foreground">{description}</p>
        <div className="mt-2 flex items-center gap-1 text-xs">
          {trendUp ? <ArrowUp className="h-3 w-3" /> : <ArrowDown className="h-3 w-3" />}
          <span
            className={cn(
              showGoodTrend ? 'text-green-500 dark:text-green-400' : 'text-red-500 dark:text-red-400'
            )}
          >
            {trend}
          </span>
        </div>
      </CardContent>
    </Card>
  );
}