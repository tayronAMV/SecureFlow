import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { PlusCircle, Users } from 'lucide-react';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';
import { Label } from '@/components/ui/label';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';

const mockUsers = [
  {
    id: '1',
    name: 'Admin User',
    email: 'admin@secureflow.io',
    role: 'admin',
  },
  {
    id: '2',
    name: 'Security Manager',
    email: 'security@secureflow.io',
    role: 'manager',
  },
  {
    id: '3',
    name: 'DevOps Engineer',
    email: 'devops@secureflow.io',
    role: 'editor',
  },
  {
    id: '4',
    name: 'Junior Analyst',
    email: 'analyst@secureflow.io',
    role: 'viewer',
  },
];

const permissions = [
  { id: 'view-dashboard', name: 'View Dashboard', description: 'View main dashboard' },
  { id: 'view-alerts', name: 'View Alerts', description: 'View security alerts and logs' },
  { id: 'manage-policies', name: 'Manage Policies', description: 'Create, edit and delete policies' },
  { id: 'manage-users', name: 'Manage Users', description: 'Add and remove users and roles' },
  { id: 'run-scans', name: 'Run Security Scans', description: 'Initiate on-demand security scans' },
];

export default function RBACPanel() {
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>User Management</CardTitle>
              <CardDescription>Manage user access and permissions</CardDescription>
            </div>
            <Button>
              <PlusCircle className="mr-2 h-4 w-4" /> Add User
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>User</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Permissions</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {mockUsers.map((user) => (
                <TableRow key={user.id}>
                  <TableCell>
                    <div>
                      <div className="font-medium">{user.name}</div>
                      <div className="text-xs text-muted-foreground">{user.email}</div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant={user.role === 'admin' ? 'default' : 'secondary'}>
                      {user.role === 'admin' ? 'Administrator' : 
                      user.role === 'manager' ? 'Security Manager' : 
                      user.role === 'editor' ? 'Editor' : 'Viewer'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1">
                      {user.role === 'admin' && (
                        <Badge variant="outline">All Permissions</Badge>
                      )}
                      {user.role === 'manager' && (
                        <>
                          <Badge variant="outline">View</Badge>
                          <Badge variant="outline">Manage Policies</Badge>
                          <Badge variant="outline">Run Scans</Badge>
                        </>
                      )}
                      {user.role === 'editor' && (
                        <>
                          <Badge variant="outline">View</Badge>
                          <Badge variant="outline">Run Scans</Badge>
                        </>
                      )}
                      {user.role === 'viewer' && (
                        <Badge variant="outline">View Only</Badge>
                      )}
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    <Button variant="ghost" size="sm">
                      Edit
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Role Permissions</CardTitle>
          <CardDescription>Configure permissions for each role</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="rounded-md border">
              <div className="bg-muted/50 p-4">
                <div className="flex items-center gap-3">
                  <Users className="h-5 w-5" />
                  <div>
                    <h3 className="font-medium">Role: Security Manager</h3>
                    <p className="text-xs text-muted-foreground">
                      Can view security data and manage policies
                    </p>
                  </div>
                </div>
              </div>
              <div className="p-4">
                <div className="space-y-4">
                  {permissions.map((permission) => (
                    <div key={permission.id} className="flex items-start space-x-3">
                      <Checkbox 
                        id={permission.id} 
                        checked={permission.id !== 'manage-users'} 
                      />
                      <div className="grid gap-1.5">
                        <Label 
                          htmlFor={permission.id}
                          className="font-medium"
                        >
                          {permission.name}
                        </Label>
                        <p className="text-xs text-muted-foreground">
                          {permission.description}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}