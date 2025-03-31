import { useState, useEffect } from 'react';
import { Edge } from 'reactflow';
import { Button } from '@/components/ui/button';
import { X, Plus, Trash2, AlertTriangle } from 'lucide-react';
import { FunctionItem, api, ParameterDefinition, OutputDefinition } from '@/services/api';
import { Alert, AlertDescription } from '@/components/ui/alert';

// Import interfaces from FlowEditor
interface DataFlowMapping {
  sourceOutput: string;
  targetInput: string;
  transform?: string; // Added transform function name
}

interface EdgeSettingsPanelProps {
  edge: Edge;
  sourceFunction: FunctionItem;
  targetFunction: FunctionItem;
  onClose: () => void;
  updateMappings: (edgeId: string, mappings: DataFlowMapping[]) => void;
}

// Helper function to determine if two paths might be compatible
const arePathsCompatible = (sourcePath: string, targetPath: string, sourceType: string, targetType: string): boolean => {
  // Direct matches
  if (sourcePath === targetPath) return true;
  
  // Common naming patterns
  const sourceKey = sourcePath.split('.').pop() || '';
  const targetKey = targetPath.split('.').pop() || '';
  
  if (sourceKey === targetKey) return true;
  
  // Type compatibility (simplified)
  if (sourceType === targetType) return true;
  if (sourceType === 'object' && targetType === 'object') return true;
  if (sourceType === 'object[]' && targetType === 'object[]') return true;
  
  // Array to single item compatibility
  if (sourceType === 'object[]' && targetType === 'object') return true;
  if (sourceType === 'string[]' && targetType === 'string') return true;
  
  return false;
};

// List of available transform functions
const transformFunctions = [
  { name: 'TransformForTrends', description: 'Transforms data for trend analysis', sourceType: 'attributes', targetType: 'trends' },
  { name: 'TransformForPatterns', description: 'Transforms data for pattern identification', sourceType: 'attributes', targetType: 'patterns' },
  { name: 'TransformForFindings', description: 'Transforms data for findings analysis', sourceType: 'patterns', targetType: 'findings' },
  { name: 'TransformForIntent', description: 'Converts attribute data to text for intent analysis', sourceType: 'attributes', targetType: 'intent' },
  { name: 'TransformIntentForFindings', description: 'Prepares intent data for findings analysis', sourceType: 'intent', targetType: 'findings' },
  { name: 'TransformFindingsForRecommendations', description: 'Prepares findings for recommendations', sourceType: 'findings', targetType: 'recommendations' },
  { name: 'TransformRecommendationsForPlan', description: 'Prepares recommendations for action planning', sourceType: 'recommendations', targetType: 'plan' },
  { name: 'TransformPlanForTimeline', description: 'Prepares action plan for timeline generation', sourceType: 'plan', targetType: 'plan' },
];

// Function to get suggested transform for a source-target function pair
const getSuggestedTransform = (sourceId: string, targetId: string): string | null => {
  // Extract function types from IDs
  const sourceType = sourceId.split('-').pop() || '';
  const targetType = targetId.split('-').pop() || '';
  
  // Find matching transform
  const transform = transformFunctions.find(
    t => t.sourceType === sourceType && t.targetType === targetType
  );
  
  return transform ? transform.name : null;
};

export default function EdgeSettingsPanel({ edge, sourceFunction, targetFunction, onClose, updateMappings }: EdgeSettingsPanelProps) {
  // Extract the current mappings from the edge data
  const initialMappings = edge.data?.mappings || [];
  const [mappings, setMappings] = useState<DataFlowMapping[]>(initialMappings);
  
  // State for function metadata
  const [sourceMetadata, setSourceMetadata] = useState<{ outputs: OutputDefinition[] }>({ outputs: [] });
  const [targetMetadata, setTargetMetadata] = useState<{ inputs: ParameterDefinition[] }>({ inputs: [] });
  const [loading, setLoading] = useState(true);
  const [incompatibleMappings, setIncompatibleMappings] = useState<number[]>([]);
  
  // Get suggested transform based on function types
  const suggestedTransform = getSuggestedTransform(sourceFunction.id, targetFunction.id);
  
  // Fetch function metadata on mount
  useEffect(() => {
    const fetchMetadata = async () => {
      try {
        console.log('Fetching metadata for', sourceFunction.id, targetFunction.id);
        
        const [sourceMeta, targetMeta] = await Promise.all([
          api.getFunctionMetadataById(sourceFunction.id),
          api.getFunctionMetadataById(targetFunction.id)
        ]);
        
        console.log('Retrieved metadata:', { 
          sourceId: sourceFunction.id,
          sourceMeta, 
          targetId: targetFunction.id,
          targetMeta 
        });
        
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
  
  // Check for incompatible mappings when metadata is loaded
  useEffect(() => {
    if (!loading && mappings.length > 0) {
      const incompatible: number[] = [];
      
      mappings.forEach((mapping, index) => {
        const sourceOutput = sourceMetadata.outputs.find(o => o.path === mapping.sourceOutput);
        const targetInput = targetMetadata.inputs.find(i => i.path === mapping.targetInput);
        
        if (sourceOutput && targetInput) {
          const isCompatible = arePathsCompatible(
            mapping.sourceOutput, 
            mapping.targetInput,
            sourceOutput.type,
            targetInput.type
          );
          
          if (!isCompatible && !mapping.transform) {
            incompatible.push(index);
          }
        }
      });
      
      setIncompatibleMappings(incompatible);
    }
  }, [loading, mappings, sourceMetadata, targetMetadata]);
  
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
    
    // Add suggested transform if available
    if (suggestedTransform) {
      newMapping.transform = suggestedTransform;
    }
    
    setMappings([...mappings, newMapping]);
  };
  
  // Add an automatic mapping (best guess based on name/type compatibility)
  const addAutomaticMapping = () => {
    const availableMappings: DataFlowMapping[] = [];
    
    // Find mappings between outputs and inputs with similar names or compatible types
    for (const output of sourceMetadata.outputs) {
      for (const input of targetMetadata.inputs) {
        if (arePathsCompatible(output.path, input.path, output.type, input.type)) {
          availableMappings.push({
            sourceOutput: output.path,
            targetInput: input.path
          });
        }
      }
    }
    
    // If we found compatible mappings, add them
    if (availableMappings.length > 0) {
      // Add transform suggestion if available
      if (suggestedTransform) {
        availableMappings.forEach(mapping => {
          mapping.transform = suggestedTransform;
        });
      }
      
      setMappings([...mappings, ...availableMappings]);
    } else if (suggestedTransform) {
      // If no compatible mappings but we have a transform, add a mapping with the transform
      const newMapping: DataFlowMapping = {
        sourceOutput: sourceMetadata.outputs[0]?.path || '',
        targetInput: targetMetadata.inputs[0]?.path || '',
        transform: suggestedTransform
      };
      setMappings([...mappings, newMapping]);
    }
  };
  
  // Update a mapping
  const updateMapping = (index: number, field: 'sourceOutput' | 'targetInput' | 'transform', value: string) => {
    const updatedMappings = [...mappings];
    updatedMappings[index][field] = value;
    
    // If setting a transform, remove from incompatible list
    if (field === 'transform' && value) {
      setIncompatibleMappings(prev => prev.filter(i => i !== index));
    }
    
    setMappings(updatedMappings);
  };
  
  // Delete a mapping
  const deleteMapping = (index: number) => {
    const updatedMappings = [...mappings];
    updatedMappings.splice(index, 1);
    setMappings(updatedMappings);
    
    // Update incompatible mappings indices
    setIncompatibleMappings(prev => 
      prev.filter(i => i !== index).map(i => i > index ? i - 1 : i)
    );
  };
  
  // Add suggested transform to all mappings
  const addSuggestedTransformToAll = () => {
    if (!suggestedTransform) return;
    
    const updatedMappings = mappings.map(mapping => ({
      ...mapping,
      transform: suggestedTransform
    }));
    
    setMappings(updatedMappings);
    setIncompatibleMappings([]);
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
        
        {suggestedTransform && (
          <Alert className="mb-4 bg-blue-50 border-blue-200">
            <AlertTriangle className="h-4 w-4 text-blue-500" />
            <AlertDescription className="text-sm">
              Recommended transformation function: <strong>{suggestedTransform}</strong>
              {incompatibleMappings.length > 0 && (
                <Button 
                  variant="outline" 
                  size="sm" 
                  className="ml-2"
                  onClick={addSuggestedTransformToAll}
                >
                  Apply to All
                </Button>
              )}
            </AlertDescription>
          </Alert>
        )}
        
        <div className="mb-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-md font-medium">Data Mappings</h3>
            <div className="space-x-2">
              <Button 
                variant="outline" 
                size="sm"
                onClick={addAutomaticMapping}
                className="h-7"
              >
                Auto Map
              </Button>
              <Button 
                variant="outline" 
                size="sm" 
                onClick={addMapping}
                className="h-7"
              >
                <Plus size={14} className="mr-1" /> Add Mapping
              </Button>
            </div>
          </div>
          
          {mappings.length === 0 ? (
            <div className="text-sm text-muted-foreground text-center p-4 border rounded-md">
              No data mappings configured yet. Click "Add Mapping" to create one or "Auto Map" to automatically connect compatible fields.
            </div>
          ) : (
            <div className="space-y-3">
              {mappings.map((mapping, index) => (
                <div 
                  key={index} 
                  className={`flex items-center gap-2 p-3 border rounded-md ${
                    incompatibleMappings.includes(index) ? 'border-amber-500 bg-amber-50' : ''
                  }`}
                >
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
                  
                  {(incompatibleMappings.includes(index) || mapping.transform) && (
                    <div className="flex-grow">
                      <label className="text-xs text-muted-foreground block mb-1">
                        Transform Function
                      </label>
                      <select 
                        value={mapping.transform || ''}
                        onChange={(e) => updateMapping(index, 'transform', e.target.value)}
                        className={`w-full text-sm p-2 border rounded-md ${
                          incompatibleMappings.includes(index) && !mapping.transform ? 'border-amber-500' : ''
                        }`}
                      >
                        <option value="">None</option>
                        {transformFunctions.map(tf => (
                          <option 
                            key={tf.name} 
                            value={tf.name}
                            disabled={tf.sourceType !== sourceFunction.id.split('-').pop() || 
                                      tf.targetType !== targetFunction.id.split('-').pop()}
                          >
                            {tf.name}
                          </option>
                        ))}
                      </select>
                    </div>
                  )}
                  
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
          
          {incompatibleMappings.length > 0 && (
            <Alert className="mt-4 bg-amber-50 border-amber-200">
              <AlertTriangle className="h-4 w-4 text-amber-500" />
              <AlertDescription className="text-sm">
                {incompatibleMappings.length} incompatible mapping{incompatibleMappings.length > 1 ? 's' : ''} detected. 
                Select a transform function to convert the data formats.
              </AlertDescription>
            </Alert>
          )}
        </div>
      </div>
    </div>
  );
} 