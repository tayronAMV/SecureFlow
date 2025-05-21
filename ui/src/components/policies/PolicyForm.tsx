import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { AlertTriangle, CheckCircle2 } from 'lucide-react';
import { toast } from 'sonner';

const policySchema = z.object({
  name: z.string().min(3, { message: 'Policy name must be at least 3 characters' }),
  description: z.string().min(5, { message: 'Description is required' }),
  scope: z.string().min(1, { message: 'Scope is required' }),
  severity: z.string().min(1, { message: 'Severity is required' }),
  action: z.string().min(1, { message: 'Action is required' }),
  enabled: z.boolean().default(true),
});

type PolicyFormValues = z.infer<typeof policySchema>;

interface PolicyFormProps {
  policy?: any;
  onClose: () => void;
}

export default function PolicyForm({ policy, onClose }: PolicyFormProps) {
  const isEditing = !!policy;
  
  const defaultValues: Partial<PolicyFormValues> = {
    name: policy?.name || '',
    description: policy?.description || '',
    scope: policy?.scope || 'global',
    severity: policy?.severity || 'medium',
    action: policy?.action || 'alert',
    enabled: policy?.enabled !== undefined ? policy.enabled : true,
  };

  const { register, handleSubmit, formState, control, setValue, watch } = useForm<PolicyFormValues>({
    resolver: zodResolver(policySchema),
    defaultValues,
  });

  const { errors } = formState;
  const enabled = watch('enabled');

  const onSubmit = (data: PolicyFormValues) => {
    // In a real app, this would save to backend
    console.log('Form submitted:', data);
    
    setTimeout(() => {
      toast.success(
        isEditing ? 'Policy updated successfully' : 'Policy created successfully', 
        { 
          icon: <CheckCircle2 className="h-5 w-5" />,
        }
      );
      onClose();
    }, 500);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{isEditing ? 'Edit Policy' : 'Create New Policy'}</CardTitle>
        <CardDescription>
          {isEditing 
            ? 'Update your security policy settings' 
            : 'Configure a new security policy to enforce container security'}
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit(onSubmit)}>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Policy Name</Label>
            <Input
              id="name"
              {...register('name')}
              placeholder="Block unauthorized network access"
            />
            {errors.name && (
              <p className="text-xs text-destructive">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              {...register('description')}
              placeholder="Prevent containers from establishing unauthorized network connections"
              rows={3}
            />
            {errors.description && (
              <p className="text-xs text-destructive">{errors.description.message}</p>
            )}
          </div>

          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="scope">Scope</Label>
              <Select 
                defaultValue={defaultValues.scope}
                onValueChange={(value) => setValue('scope', value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select scope" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="global">Global</SelectItem>
                  <SelectItem value="production">Production</SelectItem>
                  <SelectItem value="development">Development</SelectItem>
                  <SelectItem value="custom">Custom</SelectItem>
                </SelectContent>
              </Select>
              {errors.scope && (
                <p className="text-xs text-destructive">{errors.scope.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="severity">Severity</Label>
              <Select 
                defaultValue={defaultValues.severity}
                onValueChange={(value) => setValue('severity', value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select severity" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="critical">Critical</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="low">Low</SelectItem>
                </SelectContent>
              </Select>
              {errors.severity && (
                <p className="text-xs text-destructive">{errors.severity.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="action">Action</Label>
              <Select 
                defaultValue={defaultValues.action}
                onValueChange={(value) => setValue('action', value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select action" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="alert">Alert Only</SelectItem>
                  <SelectItem value="block">Block</SelectItem>
                  <SelectItem value="isolate">Isolate</SelectItem>
                </SelectContent>
              </Select>
              {errors.action && (
                <p className="text-xs text-destructive">{errors.action.message}</p>
              )}
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="enabled"
              checked={enabled}
              onCheckedChange={(checked) => setValue('enabled', checked)}
            />
            <Label htmlFor="enabled">Enable this policy</Label>
          </div>

          <div className="rounded-md bg-muted p-3 text-sm">
            <div className="flex gap-2">
              <AlertTriangle className="h-5 w-5 text-yellow-500" />
              <div>
                <p className="font-medium">Important: Runtime Policy Impact</p>
                <p className="text-muted-foreground">
                  Blocking policies may temporarily impact container operations. 
                  Test in non-production environments first.
                </p>
              </div>
            </div>
          </div>
        </CardContent>
        <CardFooter className="flex justify-between">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit">
            {isEditing ? 'Update Policy' : 'Create Policy'}
          </Button>
        </CardFooter>
      </form>
    </Card>
  );
}