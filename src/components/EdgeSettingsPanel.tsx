import { useState, useEffect } from 'react';
import { Edge } from 'reactflow';
import { Button } from '@/components/ui/button';
import { X, Plus, Trash2 } from 'lucide-react';
import { FunctionItem } from '@/services/api';

// Import interfaces from FlowEditor
interface DataFlowMapping {
  sourceOutput: string;
  targetInput: string;
}

interface EdgeSettingsPanelProps {
  edge: Edge;
  sourceFunction: FunctionItem;
  targetFunction: FunctionItem;
  onClose: () => void;
  updateMappings: (edgeId: string, mappings: DataFlowMapping[]) => void;
}

export default function EdgeSettingsPanel({ edge, sourceFunction, targetFunction, onClose, updateMappings }: EdgeSettingsPanelProps) {
  // Extract the current mappings from the edge data
  const initialMappings = edge.data?.mappings || [];
  const [mappings, setMappings] = useState<DataFlowMapping[]>(initialMappings);
  
  // Define source outputs and target inputs based on function schemas
  const sourceOutputs = getSourceOutputs(sourceFunction);
  const targetInputs = getTargetInputs(targetFunction);
  
  // Update edge data when mappings change
  useEffect(() => {
    updateMappings(edge.id, mappings);
  }, [mappings, edge.id, updateMappings]);
  
  // Add a new mapping
  const addMapping = () => {
    // Default to first available source output and target input
    const newMapping: DataFlowMapping = {
      sourceOutput: sourceOutputs.length > 0 ? sourceOutputs[0].name : '',
      targetInput: targetInputs.length > 0 ? targetInputs[0].name : ''
    };
    
    setMappings([...mappings, newMapping]);
  };
  
  // Update a mapping
  const updateMapping = (index: number, field: 'sourceOutput' | 'targetInput', value: string) => {
    const updatedMappings = [...mappings];
    updatedMappings[index][field] = value;
    setMappings(updatedMappings);
  };
  
  // Delete a mapping
  const deleteMapping = (index: number) => {
    const updatedMappings = [...mappings];
    updatedMappings.splice(index, 1);
    setMappings(updatedMappings);
  };
  
  return (
    <div className="h-full bg-card border-l flex flex-col">
      <div className="flex justify-between items-center p-4 border-b">
        <h2 className="text-lg font-semibold">Data Flow Configuration</h2>
        <Button variant="ghost" size="sm" onClick={onClose}>
          <X size={16} />
        </Button>
      </div>
      
      <div className="p-4 flex-grow overflow-y-auto">
        <div className="mb-6">
          <div className="flex justify-between items-center mb-2">
            <h3 className="text-md font-medium">Source Function</h3>
            <span className="text-sm bg-blue-100 text-blue-800 px-2 py-1 rounded">
              {sourceFunction.label}
            </span>
          </div>
          <p className="text-sm text-muted-foreground mb-2">{sourceFunction.description}</p>
          
          <h4 className="text-sm font-medium mt-4 mb-2">Available Outputs:</h4>
          <ul className="text-sm text-muted-foreground ml-4 list-disc">
            {sourceOutputs.map(output => (
              <li key={output.name}>
                <span className="font-medium">{output.name}</span>: {output.description}
              </li>
            ))}
          </ul>
        </div>
        
        <div className="mb-6">
          <div className="flex justify-between items-center mb-2">
            <h3 className="text-md font-medium">Target Function</h3>
            <span className="text-sm bg-emerald-100 text-emerald-800 px-2 py-1 rounded">
              {targetFunction.label}
            </span>
          </div>
          <p className="text-sm text-muted-foreground mb-2">{targetFunction.description}</p>
          
          <h4 className="text-sm font-medium mt-4 mb-2">Required Inputs:</h4>
          <ul className="text-sm text-muted-foreground ml-4 list-disc">
            {targetInputs.map(input => (
              <li key={input.name}>
                <span className="font-medium">{input.name}</span>: {input.description}
              </li>
            ))}
          </ul>
        </div>
        
        <div className="mb-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-md font-medium">Data Mappings</h3>
            <Button 
              variant="outline" 
              size="sm" 
              onClick={addMapping}
              className="h-7"
            >
              <Plus size={14} className="mr-1" /> Add Mapping
            </Button>
          </div>
          
          {mappings.length === 0 ? (
            <div className="text-sm text-muted-foreground text-center p-4 border rounded-md">
              No data mappings configured yet. Click "Add Mapping" to create one.
            </div>
          ) : (
            <div className="space-y-3">
              {mappings.map((mapping, index) => (
                <div key={index} className="flex items-center gap-2 p-3 border rounded-md">
                  <div className="flex-grow">
                    <label className="text-xs text-muted-foreground block mb-1">
                      Source Output
                    </label>
                    <select 
                      value={mapping.sourceOutput}
                      onChange={(e) => updateMapping(index, 'sourceOutput', e.target.value)}
                      className="w-full text-sm p-2 border rounded-md"
                    >
                      {sourceOutputs.map(output => (
                        <option key={output.name} value={output.name}>
                          {output.name}
                        </option>
                      ))}
                    </select>
                  </div>
                  
                  <div className="self-center text-center">
                    → 
                  </div>
                  
                  <div className="flex-grow">
                    <label className="text-xs text-muted-foreground block mb-1">
                      Target Input
                    </label>
                    <select 
                      value={mapping.targetInput}
                      onChange={(e) => updateMapping(index, 'targetInput', e.target.value)}
                      className="w-full text-sm p-2 border rounded-md"
                    >
                      {targetInputs.map(input => (
                        <option key={input.name} value={input.name}>
                          {input.name}
                        </option>
                      ))}
                    </select>
                  </div>
                  
                  <Button 
                    variant="ghost" 
                    size="sm" 
                    onClick={() => deleteMapping(index)}
                    className="self-end h-8 w-8"
                  >
                    <Trash2 size={14} />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// Helper function to get source outputs based on function type
function getSourceOutputs(func: FunctionItem): { name: string; description: string; }[] {
  // This would typically come from the function's schema
  // For now, we'll hardcode some example outputs based on function type
  
  switch (func.id) {
    case 'analysis-trends':
      return [
        { name: 'trend_descriptions', description: 'Descriptions of identified trends' },
        { name: 'recommended_actions', description: 'Recommended actions based on trends' }
      ];
    case 'analysis-patterns':
      return [
        { name: 'patterns', description: 'List of identified patterns' }
      ];
    case 'analysis-findings':
      return [
        { name: 'findings', description: 'Analytical findings' },
        { name: 'recommendations', description: 'Recommended actions' }
      ];
    case 'analysis-attributes':
      return [
        { name: 'attribute_values', description: 'Extracted attribute values' }
      ];
    case 'analysis-intent':
      return [
        { name: 'label_name', description: 'Machine-readable intent label' },
        { name: 'label', description: 'Human-readable intent label' },
        { name: 'description', description: 'Intent description' }
      ];
    case 'analysis-recommendations':
      return [
        { name: 'recommendations', description: 'Recommended actions' },
        { name: 'priorities', description: 'Priority ratings for recommendations' }
      ];
    case 'analysis-plan':
      return [
        { name: 'action_plan', description: 'Full action plan' },
        { name: 'timeline', description: 'Implementation timeline' },
        { name: 'resources', description: 'Required resources' }
      ];
    default:
      return [
        { name: 'results', description: 'Analysis results' }
      ];
  }
}

// Helper function to get target inputs based on function type
function getTargetInputs(func: FunctionItem): { name: string; description: string; required?: boolean }[] {
  // This would typically come from the function's schema
  // For now, we'll hardcode some example inputs based on function type
  
  switch (func.id) {
    case 'analysis-trends':
      return [
        { name: 'disputes', description: 'Dispute data to analyze', required: true },
        { name: 'conversations', description: 'Example conversations' },
        { name: 'attributes', description: 'Additional attributes' }
      ];
    case 'analysis-patterns':
      return [
        { name: 'disputes', description: 'Dispute data to analyze', required: true },
        { name: 'conversations', description: 'Example conversations' },
        { name: 'attributes', description: 'Attributes to look for patterns' }
      ];
    case 'analysis-findings':
      return [
        { name: 'disputes', description: 'Dispute data to analyze', required: true },
        { name: 'trends', description: 'Trend analysis results' },
        { name: 'patterns', description: 'Pattern analysis results' },
        { name: 'conversations', description: 'Example conversations' }
      ];
    case 'analysis-attributes':
      return [
        { name: 'text', description: 'Text to extract attributes from', required: true }
      ];
    case 'analysis-intent':
      return [
        { name: 'text', description: 'Text to extract intent from', required: true }
      ];
    case 'analysis-recommendations':
      return [
        { name: 'data', description: 'Data for recommendation generation', required: true },
        { name: 'findings', description: 'Analysis findings' },
        { name: 'patterns', description: 'Identified patterns' }
      ];
    case 'analysis-plan':
      return [
        { name: 'recommendations', description: 'Recommendations to build plan from', required: true },
        { name: 'context', description: 'Context information' }
      ];
    default:
      return [
        { name: 'data', description: 'Input data', required: true }
      ];
  }
} 