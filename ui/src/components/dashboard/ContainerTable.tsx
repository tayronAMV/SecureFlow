import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { formatDistanceToNow } from 'date-fns';
import { cn } from '@/lib/utils';

interface Container {
  id: string;
  name: string;
  status: string;
  image: string;
  created: string;
  cpu: number;
  memory: number;
  anomalyScore: number;
  vulnerabilities: number;
}

interface ContainerTableProps {
  containers: Container[];
}

export default function ContainerTable({ containers }: ContainerTableProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Container</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>CPU</TableHead>
            <TableHead>Memory</TableHead>
            <TableHead>Anomaly Score</TableHead>
            <TableHead>Age</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {containers.map((container) => (
            <TableRow key={container.id}>
              <TableCell className="font-medium">
                <div>
                  <div className="font-medium">{container.name}</div>
                  <div className="text-xs text-muted-foreground">{container.image}</div>
                </div>
              </TableCell>
              <TableCell>
                <Badge 
                  variant={container.status === 'running' ? 'default' : 'secondary'}
                  className={cn(
                    container.status === 'running' ? 'bg-green-500/20 text-green-600 hover:bg-green-500/30 dark:bg-green-500/20 dark:text-green-400 dark:hover:bg-green-500/30' : 
                    'bg-yellow-500/20 text-yellow-600 hover:bg-yellow-500/30 dark:bg-yellow-500/20 dark:text-yellow-400 dark:hover:bg-yellow-500/30'
                  )}
                >
                  {container.status}
                </Badge>
              </TableCell>
              <TableCell>
                <div className="flex items-center gap-2">
                  <div className="h-2 w-full rounded-full bg-muted">
                    <div 
                      className={cn(
                        "h-full rounded-full",
                        container.cpu < 50 ? "bg-green-500" : 
                        container.cpu < 80 ? "bg-yellow-500" : 
                        "bg-red-500"
                      )}
                      style={{ width: `${container.cpu}%` }}
                    />
                  </div>
                  <span className="text-xs">{container.cpu}%</span>
                </div>
              </TableCell>
              <TableCell>
                <div className="flex items-center gap-2">
                  <div className="h-2 w-full rounded-full bg-muted">
                    <div 
                      className={cn(
                        "h-full rounded-full",
                        container.memory < 60 ? "bg-green-500" : 
                        container.memory < 85 ? "bg-yellow-500" : 
                        "bg-red-500"
                      )}
                      style={{ width: `${container.memory}%` }}
                    />
                  </div>
                  <span className="text-xs">{container.memory}%</span>
                </div>
              </TableCell>
              <TableCell>
                <div className={cn(
                  "rounded px-1 py-0.5 text-center text-xs font-medium",
                  container.anomalyScore === 0 ? "bg-green-500/20 text-green-600 dark:bg-green-500/20 dark:text-green-400" :
                  container.anomalyScore < 10 ? "bg-blue-500/20 text-blue-600 dark:bg-blue-500/20 dark:text-blue-400" :
                  container.anomalyScore < 30 ? "bg-yellow-500/20 text-yellow-600 dark:bg-yellow-500/20 dark:text-yellow-400" :
                  "bg-red-500/20 text-red-600 dark:bg-red-500/20 dark:text-red-400"
                )}>
                  {container.anomalyScore}
                </div>
              </TableCell>
              <TableCell className="text-xs">
                {formatDistanceToNow(new Date(container.created), { addSuffix: true })}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}