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
import { Check, AlertTriangle, AlertCircle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

interface Vulnerability {
  id: string;
  severity: string;
  package: string;
  description: string;
}

interface ScanResult {
  id: string;
  containerId: string;
  timestamp: string;
  status: string;
  details: string;
  vulnerabilities: Vulnerability[];
}

interface ScanReportTableProps {
  scanResults: ScanResult[];
}

export default function ScanReportTable({ scanResults }: ScanReportTableProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Container</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Vulnerabilities</TableHead>
            <TableHead>Details</TableHead>
            <TableHead>Scanned</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {scanResults.map((scan) => (
            <TableRow key={scan.id}>
              <TableCell className="font-medium">
                {scan.containerId.replace('c-', '')}
              </TableCell>
              <TableCell>
                <div className="flex items-center gap-2">
                  {scan.status === 'OK' ? (
                    <Badge className="bg-green-500/20 text-green-600 hover:bg-green-500/30 dark:bg-green-500/20 dark:text-green-400 dark:hover:bg-green-500/30">
                      <Check className="mr-1 h-3 w-3" />
                      {scan.status}
                    </Badge>
                  ) : scan.status === 'Warning' ? (
                    <Badge className="bg-yellow-500/20 text-yellow-600 hover:bg-yellow-500/30 dark:bg-yellow-500/20 dark:text-yellow-400 dark:hover:bg-yellow-500/30">
                      <AlertTriangle className="mr-1 h-3 w-3" />
                      {scan.status}
                    </Badge>
                  ) : (
                    <Badge className="bg-red-500/20 text-red-600 hover:bg-red-500/30 dark:bg-red-500/20 dark:text-red-400 dark:hover:bg-red-500/30">
                      <AlertCircle className="mr-1 h-3 w-3" />
                      {scan.status}
                    </Badge>
                  )}
                </div>
              </TableCell>
              <TableCell>
                {scan.vulnerabilities.length > 0 ? (
                  <div className="flex flex-wrap gap-1">
                    {scan.vulnerabilities.map((vuln) => (
                      <Badge 
                        key={vuln.id} 
                        variant="outline" 
                        className={cn(
                          "text-xs",
                          vuln.severity === 'critical' ? "border-red-500 text-red-500" :
                          vuln.severity === 'high' ? "border-orange-500 text-orange-500" :
                          vuln.severity === 'medium' ? "border-yellow-500 text-yellow-500" :
                          "border-blue-500 text-blue-500"
                        )}
                      >
                        {vuln.id}
                      </Badge>
                    ))}
                  </div>
                ) : (
                  <span className="text-xs text-muted-foreground">None found</span>
                )}
              </TableCell>
              <TableCell className="max-w-[200px] truncate text-xs">
                {scan.details}
              </TableCell>
              <TableCell className="text-xs">
                {formatDistanceToNow(new Date(scan.timestamp), { addSuffix: true })}
              </TableCell>
              <TableCell className="text-right">
                <Button size="sm" variant="outline" className="text-xs">
                  View Details
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}