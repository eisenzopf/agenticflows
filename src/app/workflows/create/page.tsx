"use client";

import React, { useState } from 'react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { useRouter } from "next/navigation";
import { api } from "@/services/api";

export default function WorkflowGeneratorPage() {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState('');
  const router = useRouter();

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
      // Call the API to generate the workflow
      const newWorkflow = await api.generateWorkflow(name, description);
      
      // Redirect to the workflow editor
      router.push(`/workflows/${newWorkflow.id}`);
    } catch (err: any) {
      console.error('Error generating workflow:', err);
      setError(err.message || 'Failed to generate workflow. Please try again.');
      setIsGenerating(false);
    }
  };

  return (
    <div className="container mx-auto max-w-2xl py-8">
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold mb-2">Create Workflow</h1>
          <p className="text-muted-foreground">
            Describe what you want your workflow to do, and we'll automatically generate it using AI.
          </p>
        </div>
        
        {error && (
          <div className="bg-destructive/10 text-destructive p-4 rounded-md">
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
              placeholder="Describe what you want this workflow to do. For example: 'Analyze customer conversations from the banking database, identify common issues, and generate recommendations for improving service.'"
              rows={6}
              disabled={isGenerating}
            />
            <p className="text-sm text-muted-foreground">
              Tip: Be specific about what data you want to analyze and what insights you're looking for.
            </p>
          </div>
          
          <div className="flex gap-4 justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={() => router.push('/workflows')}
              disabled={isGenerating}
            >
              Cancel
            </Button>
            <Button 
              type="submit" 
              disabled={isGenerating}
              className="min-w-[100px]"
            >
              {isGenerating ? 'Generating...' : 'Generate Workflow'}
            </Button>
          </div>
        </form>
        
        <div className="bg-muted p-4 rounded-md">
          <h3 className="text-sm font-medium mb-2">About this feature</h3>
          <p className="text-sm text-muted-foreground">
            This workflow will automatically connect to the Standard Chartered Bank database at:
            <code className="bg-muted-foreground/10 rounded-sm px-1 py-0.5 block mt-1 mb-2 text-xs overflow-x-auto">
              /Users/jonathan/Documents/Work/discourse_ai/Research/corpora/banking_2025/db/standard_charter_bank.db
            </code>
            It will analyze conversations from the database and provide insights based on your description.
          </p>
        </div>
      </div>
    </div>
  );
} 