import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { X } from 'lucide-react';

interface WorkflowInputFormProps {
  onSubmit: (data: Record<string, any>) => void;
  onClose: () => void;
}

export default function WorkflowInputForm({ onSubmit, onClose }: WorkflowInputFormProps) {
  const [disputeText, setDisputeText] = useState<string>('');
  const [disputeAmount, setDisputeAmount] = useState<string>('');
  const [disputeCount, setDisputeCount] = useState<string>('10');
  
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Prepare data for workflow execution
    const data: Record<string, any> = {
      text: disputeText,
      disputes: [{
        id: 'generated-dispute-1',
        text: disputeText,
        amount: parseFloat(disputeAmount) || 0,
        created_at: new Date().toISOString()
      }],
      count: parseInt(disputeCount) || 10,
      attributes: {
        dispute_count: parseInt(disputeCount) || 10,
        avg_amount: parseFloat(disputeAmount) || 0,
        dispute_timespan: '3 months'
      }
    };
    
    onSubmit(data);
  };
  
  return (
    <div className="bg-card border shadow-lg rounded-lg p-6 max-w-lg w-full">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-lg font-semibold">Workflow Input</h2>
        <Button variant="ghost" size="sm" onClick={onClose}>
          <X size={16} />
        </Button>
      </div>
      
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <Label htmlFor="dispute-text">Example Dispute Text</Label>
          <Textarea 
            id="dispute-text" 
            placeholder="I was charged a $35 overdraft fee but I had sufficient funds in my account."
            value={disputeText}
            onChange={(e) => setDisputeText(e.target.value)}
            className="h-24"
          />
        </div>
        
        <div className="grid grid-cols-2 gap-4">
          <div>
            <Label htmlFor="dispute-amount">Dispute Amount ($)</Label>
            <Input 
              id="dispute-amount" 
              type="number" 
              placeholder="35.00"
              value={disputeAmount}
              onChange={(e) => setDisputeAmount(e.target.value)}
            />
          </div>
          
          <div>
            <Label htmlFor="dispute-count">Number of Disputes</Label>
            <Input 
              id="dispute-count" 
              type="number" 
              placeholder="10"
              value={disputeCount}
              onChange={(e) => setDisputeCount(e.target.value)}
            />
          </div>
        </div>
        
        <div className="flex justify-end pt-2">
          <Button 
            type="button" 
            variant="outline" 
            onClick={onClose}
            className="mr-2"
          >
            Cancel
          </Button>
          <Button type="submit">
            Execute Workflow
          </Button>
        </div>
      </form>
    </div>
  );
} 