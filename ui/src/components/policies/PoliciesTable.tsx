import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { formatDistanceToNow } from 'date-fns';
import { Edit, Trash2 } from 'lucide-react';
import { cn } from '@/lib/utils';

interface Policy {
  id: string;
  name: string;
  description: string;
  scope: string;
  severity: string;
  action: string;
  enabled: boolean;
  createdAt: string;
  createdBy: string;
}

interface PoliciesTableProps {
  policies: Policy[];
  onEdit: (policy: Policy) => void;
}

export default function PoliciesTable({ policies, onEdit }: PoliciesTableProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-12">Status</TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Scope</TableHead>
            <TableHead>Severity</TableHead>
            <TableHead>Action</TableHead>
            <TableHead>Created</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {policies.map((policy) => (
            <TableRow key={policy.id}>
              <TableCell>
                <Switch 
                  checked={policy.enabled} 
                  onCheckedChange={() => {}} 
                  aria-label="Toggle policy"
                />
              </TableCell>
              <TableCell className="font-medium">
                <div>
                  <div className="font-medium">{policy.name}</div>
                  <div className="text-xs text-muted-foreground truncate max-w-[300px]">
                    {policy.description}
                  </div>
                </div>
              </TableCell>
              <TableCell>
                <Badge variant="outline">{policy.scope}</Badge>
              </TableCell>
              <TableCell>
                <Badge 
                  className={cn(
                    policy.severity === 'critical' ? "bg-red-500/20 text-red-600 hover:bg-red-500/30 dark:bg-red-500/20 dark:text-red-400 dark:hover:bg-red-500/30" :
                    policy.severity === 'high' ? "bg-orange-500/20 text-orange-600 hover:bg-orange-500/30 dark:bg-orange-500/20 dark:text-orange-400 dark:hover:bg-orange-500/30" :
                    "bg-yellow-500/20 text-yellow-600 hover:bg-yellow-500/30 dark:bg-yellow-500/20 dark:text-yellow-400 dark:hover:bg-yellow-500/30"
                  )}
                >
                  {policy.severity}
                </Badge>
              </TableCell>
              <TableCell>
                <Badge 
                  variant="secondary" 
                  className={policy.action === 'block' ? 'bg-blue-500/20 text-blue-600 dark:bg-blue-500/20 dark:text-blue-400' : ''}
                >
                  {policy.action}
                </Badge>
              </TableCell>
              <TableCell className="text-xs">
                {formatDistanceToNow(new Date(policy.createdAt), { addSuffix: true })}
              </TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                  <Button
                    size="icon"
                    variant="ghost"
                    onClick={() => onEdit(policy)}
                  >
                    <Edit className="h-4 w-4" />
                  </Button>
                  <Button
                    size="icon"
                    variant="ghost"
                    className="text-destructive hover:text-destructive"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}