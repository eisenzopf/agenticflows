import React, { useState } from 'react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { X } from "lucide-react";
import { api } from "@/services/api";

interface DynamicWorkflowGeneratorModalProps {
  onClose: () => void;
  onWorkflowGenerated: (workflowId: string) => void;
}

export default function DynamicWorkflowGeneratorModal({ onClose, onWorkflowGenerated }: DynamicWorkflowGeneratorModalProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!name.trim()) {
      setError('Please provide a name for your workflow');
      return;
    }
    
    if (!description.trim()) {
      setError('Please provide a description of what you want the workflow to do');
      return;
    }
    
    setIsGenerating(true);
    setError('');
    
    try {
      // Call the API to generate the dynamic workflow
      const newWorkflow = await api.generateDynamicWorkflow(name, description);
      
      // Call the callback with the new workflow ID
      onWorkflowGenerated(newWorkflow.id);
    } catch (err: any) {
      console.error('Error generating dynamic workflow:', err);
      setError(err.message || 'Failed to generate dynamic workflow. Please try again.');
      setIsGenerating(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-card border shadow-lg rounded-lg p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-semibold">Generate Dynamic Workflow</h2>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X size={16} />
          </Button>
        </div>
        
        <p className="text-muted-foreground mb-4">
          Describe what you want your dynamic workflow to do, and we'll automatically generate a custom workflow with specialized functions.
        </p>
        
        {error && (
          <div className="bg-destructive/10 text-destructive p-4 rounded-md mb-4">
            {error}
          </div>
        )}
        
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="name">Workflow Name</Label>
            <Input 
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter a name for your workflow"
              disabled={isGenerating}
            />
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Describe what you want this workflow to do. Be specific about the type of processing and analysis you need."
              rows={6}
              disabled={isGenerating}
            />
            <p className="text-sm text-muted-foreground">
              Tip: Be specific about the specialized functionality you need. This will create custom functions tailored to your requirements.
            </p>
          </div>
          
          <div className="bg-muted p-4 rounded-md">
            <h3 className="text-sm font-medium mb-2">About Dynamic Workflows</h3>
            <p className="text-sm text-muted-foreground">
              Dynamic workflows create custom functions specific to your needs, rather than using pre-existing functions.
              This gives you more flexibility but may take slightly longer to generate.
            </p>
          </div>
          
          <div className="flex gap-4 justify-end pt-2">
            <Button 
              type="button" 
              variant="outline" 
              onClick={onClose}
              disabled={isGenerating}
            >
              Cancel
            </Button>
            <Button 
              type="submit" 
              disabled={isGenerating}
              className="min-w-[100px]"
            >
              {isGenerating ? 'Generating...' : 'Generate'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
} 