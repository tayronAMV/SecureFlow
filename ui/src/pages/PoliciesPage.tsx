import { useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { policies } from '@/lib/mockData';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';
import PoliciesTable from '@/components/policies/PoliciesTable';
import PolicyForm from '@/components/policies/PolicyForm';
import RBACPanel from '@/components/policies/RBACPanel';

export default function PoliciesPage() {
  const [activeTab, setActiveTab] = useState('policies');
  const [showPolicyForm, setShowPolicyForm] = useState(false);
  const [editingPolicy, setEditingPolicy] = useState<any>(null);

  const handleAddPolicy = () => {
    setEditingPolicy(null);
    setShowPolicyForm(true);
  };

  const handleEditPolicy = (policy: any) => {
    setEditingPolicy(policy);
    setShowPolicyForm(true);
  };

  const handleCloseForm = () => {
    setShowPolicyForm(false);
    setEditingPolicy(null);
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 md:flex-row md:items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Policy Management</h2>
          <p className="text-muted-foreground">
            Configure security policies and access controls
          </p>
        </div>
        <Button onClick={handleAddPolicy} className="shrink-0">
          <Plus className="mr-2 h-4 w-4" /> New Policy
        </Button>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="policies">Security Policies</TabsTrigger>
          <TabsTrigger value="rbac">RBAC Configuration</TabsTrigger>
        </TabsList>
        
        <TabsContent value="policies" className="mt-4 space-y-4">
          {showPolicyForm ? (
            <PolicyForm policy={editingPolicy} onClose={handleCloseForm} />
          ) : (
            <PoliciesTable 
              policies={policies} 
              onEdit={handleEditPolicy}
            />
          )}
        </TabsContent>
        
        <TabsContent value="rbac" className="mt-4">
          <RBACPanel />
        </TabsContent>
      </Tabs>
    </div>
  );
}