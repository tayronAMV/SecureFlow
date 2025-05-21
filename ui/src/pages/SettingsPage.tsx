import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import * as z from 'zod';

import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { toast } from 'sonner';
import { useTheme } from '@/contexts/ThemeContext';
import { CheckCircle2 } from 'lucide-react';

const settingsFormSchema = z.object({
  websocketEndpoint: z.string().url({ message: 'Please enter a valid URL' }),
  agentLogInterval: z.string(),
  autoScanEnabled: z.boolean().default(true),
  scanInterval: z.string(),
  slackWebhook: z.string().url({ message: 'Please enter a valid URL' }).or(z.string().length(0)),
  emailNotifications: z.boolean().default(false),
  emailAddress: z.string().email({ message: 'Please enter a valid email' }).or(z.string().length(0)),
});

type SettingsFormValues = z.infer<typeof settingsFormSchema>;

// Update the default WebSocket URL to use the current host
const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsHost = window.location.host;
const defaultWsUrl = `${wsProtocol}//${wsHost}/ws`;

const defaultValues: Partial<SettingsFormValues> = {
  websocketEndpoint: defaultWsUrl,
  agentLogInterval: '30',
  autoScanEnabled: true,
  scanInterval: 'daily',
  slackWebhook: '',
  emailNotifications: false,
  emailAddress: '',
};

export default function SettingsPage() {
  const { theme, setTheme } = useTheme();
  
  const form = useForm<SettingsFormValues>({
    resolver: zodResolver(settingsFormSchema),
    defaultValues,
  });

  function onSubmit(data: SettingsFormValues) {
    toast.success('Settings saved successfully', { 
      icon: <CheckCircle2 className="h-5 w-5" />,
    });
    console.log(data);
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Settings</h2>
        <p className="text-muted-foreground">
          Configure your SecureFlow platform settings
        </p>
      </div>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <Card>
            <CardHeader>
              <CardTitle>Agent Configuration</CardTitle>
              <CardDescription>
                Configure how SecureFlow agents connect and report data
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <FormField
                control={form.control}
                name="websocketEndpoint"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>WebSocket Endpoint</FormLabel>
                    <FormControl>
                      <Input placeholder={defaultWsUrl} {...field} />
                    </FormControl>
                    <FormDescription>
                      The endpoint where agents will connect to send real-time data
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="agentLogInterval"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Agent Log Interval (seconds)</FormLabel>
                    <Select 
                      onValueChange={field.onChange} 
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select interval" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="10">10 seconds</SelectItem>
                        <SelectItem value="30">30 seconds</SelectItem>
                        <SelectItem value="60">1 minute</SelectItem>
                        <SelectItem value="300">5 minutes</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      How frequently agents should send log data
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Security Scanning</CardTitle>
              <CardDescription>
                Configure automated security scanning settings
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <FormField
                control={form.control}
                name="autoScanEnabled"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                      <FormLabel className="text-base">
                        Automatic Security Scanning
                      </FormLabel>
                      <FormDescription>
                        Regularly scan containers for vulnerabilities
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="scanInterval"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Scan Frequency</FormLabel>
                    <Select 
                      onValueChange={field.onChange} 
                      defaultValue={field.value}
                      disabled={!form.watch('autoScanEnabled')}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select frequency" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="hourly">Hourly</SelectItem>
                        <SelectItem value="daily">Daily</SelectItem>
                        <SelectItem value="weekly">Weekly</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      How often to perform security scans
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Notifications</CardTitle>
              <CardDescription>
                Configure how you receive security alerts
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <FormField
                control={form.control}
                name="slackWebhook"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Slack Webhook URL</FormLabel>
                    <FormControl>
                      <Input placeholder="https://hooks.slack.com/services/..." {...field} />
                    </FormControl>
                    <FormDescription>
                      Send critical alerts to a Slack channel (optional)
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="emailNotifications"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                      <FormLabel className="text-base">
                        Email Notifications
                      </FormLabel>
                      <FormDescription>
                        Receive security alerts via email
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="emailAddress"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email Address</FormLabel>
                    <FormControl>
                      <Input 
                        placeholder="security@example.com" 
                        {...field} 
                        disabled={!form.watch('emailNotifications')}
                      />
                    </FormControl>
                    <FormDescription>
                      Where to send email notifications
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Appearance</CardTitle>
              <CardDescription>
                Customize the look and feel of SecureFlow
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <FormLabel className="text-base">Theme</FormLabel>
                <div className="grid grid-cols-3 gap-2">
                  <Button 
                    type="button" 
                    variant={theme === 'light' ? 'default' : 'outline'} 
                    onClick={() => setTheme('light')}
                    className="justify-start"
                  >
                    Light
                  </Button>
                  <Button 
                    type="button" 
                    variant={theme === 'dark' ? 'default' : 'outline'} 
                    onClick={() => setTheme('dark')}
                    className="justify-start"
                  >
                    Dark
                  </Button>
                  <Button 
                    type="button" 
                    variant={theme === 'system' ? 'default' : 'outline'} 
                    onClick={() => setTheme('system')}
                    className="justify-start"
                  >
                    System
                  </Button>
                </div>
                <FormDescription>
                  Select a theme preference for the dashboard
                </FormDescription>
              </div>
            </CardContent>
            <CardFooter className="border-t px-6 py-4">
              <Button type="submit">Save Settings</Button>
            </CardFooter>
          </Card>
        </form>
      </Form>
    </div>
  );
}