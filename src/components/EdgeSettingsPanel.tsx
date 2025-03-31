import { useState, useEffect } from 'react';
import { Edge } from 'reactflow';
import { Button } from '@/components/ui/button';
import { X, Plus, Trash2 } from 'lucide-react';
import { FunctionItem, api, ParameterDefinition, OutputDefinition } from '@/services/api';

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
  
  // State for function metadata
  const [sourceMetadata, setSourceMetadata] = useState<{ outputs: OutputDefinition[] }>({ outputs: [] });
  const [targetMetadata, setTargetMetadata] = useState<{ inputs: ParameterDefinition[] }>({ inputs: [] });
  const [loading, setLoading] = useState(true);
  
  // Fetch function metadata on mount
  useEffect(() => {
    const fetchMetadata = async () => {
      try {
        const [sourceMeta, targetMeta] = await Promise.all([
          api.getFunctionMetadataById(sourceFunction.id),
          api.getFunctionMetadataById(targetFunction.id)
        ]);
        
        if (sourceMeta) setSourceMetadata(sourceMeta);
        if (targetMeta) setTargetMetadata(targetMeta);
        setLoading(false);
      } catch (error) {
        console.error('Error fetching function metadata:', error);
        setLoading(false);
      }
    };
    
    fetchMetadata();
  }, [sourceFunction.id, targetFunction.id]);
  
  // Update edge data when mappings change
  useEffect(() => {
    updateMappings(edge.id, mappings);
  }, [mappings, edge.id, updateMappings]);
  
  // Add a new mapping
  const addMapping = () => {
    const newMapping: DataFlowMapping = {
      sourceOutput: sourceMetadata.outputs[0]?.path || '',
      targetInput: targetMetadata.inputs[0]?.path || ''
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
  
  if (loading) {
    return (
      <div className="h-full bg-card border-l flex items-center justify-center">
        <div className="text-sm text-muted-foreground">Loading function metadata...</div>
      </div>
    );
  }
  
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
            {sourceMetadata.outputs.map(output => (
              <li key={output.path}>
                <span className="font-medium">{output.name}</span>: {output.description}
                <span className="text-xs text-muted-foreground ml-1">({output.type})</span>
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
            {targetMetadata.inputs.map(input => (
              <li key={input.path}>
                <span className="font-medium">{input.name}</span>: {input.description}
                <span className="text-xs text-muted-foreground ml-1">({input.type})</span>
                {input.required && <span className="text-xs text-red-500 ml-1">(Required)</span>}
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
                      {sourceMetadata.outputs.map(output => (
                        <option key={output.path} value={output.path}>
                          {output.name} ({output.type})
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
                      {targetMetadata.inputs.map(input => (
                        <option key={input.path} value={input.path}>
                          {input.name} ({input.type})
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