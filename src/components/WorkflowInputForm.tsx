import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { X } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { api } from '@/services/api';
import { Checkbox } from '@/components/ui/checkbox';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface WorkflowInputField {
  id: string;
  label: string;
  type: 'text' | 'number' | 'textarea' | 'checkbox' | 'select' | 'fileUpload';
  description?: string;
  placeholder?: string;
  defaultValue?: any;
  required?: boolean;
  options?: { value: string; label: string }[]; // For select fields
}

interface DataSourceConfig {
  id: string;
  name: string;
  description: string;
  fields: WorkflowInputField[];
}

interface WorkflowConfig {
  id: string;
  name: string;
  description: string;
  inputTabs: {
    id: string;
    label: string;
    dataSourceConfigs: DataSourceConfig[];
  }[];
  parameters: {
    id: string;
    label: string;
    fields: WorkflowInputField[];
  }[];
}

interface WorkflowInputFormProps {
  workflowId?: string;
  onSubmit: (data: Record<string, any>) => void;
  onClose: () => void;
}

export default function WorkflowInputForm({ workflowId, onSubmit, onClose }: WorkflowInputFormProps) {
  const [activeTab, setActiveTab] = useState('data');
  const [workflowConfig, setWorkflowConfig] = useState<WorkflowConfig | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [formValues, setFormValues] = useState<Record<string, any>>({});
  const [selectedDataSources, setSelectedDataSources] = useState<Record<string, boolean>>({});
  
  // Load workflow configuration based on workflowId
  useEffect(() => {
    const loadWorkflowConfig = async () => {
      if (!workflowId) {
        // Use default configuration if no workflowId
        setWorkflowConfig(getDefaultWorkflowConfig());
        setIsLoading(false);
        return;
      }
      
      try {
        setIsLoading(true);
        // Fetch workflow configuration from the API
        const config = await api.getWorkflowExecutionConfig(workflowId);
        setWorkflowConfig(config);
        
        // Initialize form values with defaults
        const initialValues: Record<string, any> = {};
        
        // Set defaults for parameters
        config.parameters.forEach(paramGroup => {
          paramGroup.fields.forEach(field => {
            if (field.defaultValue !== undefined) {
              initialValues[field.id] = field.defaultValue;
            }
          });
        });
        
        // Set defaults for data sources
        config.inputTabs.forEach(tab => {
          tab.dataSourceConfigs.forEach(dataSource => {
            dataSource.fields.forEach(field => {
              if (field.defaultValue !== undefined) {
                initialValues[`${dataSource.id}_${field.id}`] = field.defaultValue;
              }
            });
          });
        });
        
        setFormValues(initialValues);
      } catch (error) {
        console.error("Error loading workflow configuration:", error);
        // Fall back to default config
        setWorkflowConfig(getDefaultWorkflowConfig());
      } finally {
        setIsLoading(false);
      }
    };
    
    loadWorkflowConfig();
  }, [workflowId]);
  
  const handleInputChange = (id: string, value: any) => {
    setFormValues(prev => ({
      ...prev,
      [id]: value
    }));
  };
  
  const handleDataSourceToggle = (sourceId: string, isSelected: boolean) => {
    setSelectedDataSources(prev => ({
      ...prev,
      [sourceId]: isSelected
    }));
  };
  
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Prepare data for workflow execution
    const data: Record<string, any> = {
      parameters: {},
      dataSources: {}
    };
    
    // Add parameters
    if (workflowConfig) {
      workflowConfig.parameters.forEach(paramGroup => {
        paramGroup.fields.forEach(field => {
          if (formValues[field.id] !== undefined) {
            data.parameters[field.id] = formValues[field.id];
          }
        });
      });
      
      // Add data from selected data sources
      workflowConfig.inputTabs.forEach(tab => {
        tab.dataSourceConfigs.forEach(dataSource => {
          if (selectedDataSources[dataSource.id]) {
            // This data source is selected
            const sourceData: Record<string, any> = {};
            
            dataSource.fields.forEach(field => {
              const fieldId = `${dataSource.id}_${field.id}`;
              if (formValues[fieldId] !== undefined) {
                sourceData[field.id] = formValues[fieldId];
              }
            });
            
            data.dataSources[dataSource.id] = sourceData;
          }
        });
      });
    }
    
    // Add backward compatibility for old code
    if (formValues.disputeText) {
      data.text = formValues.disputeText;
      
      if (formValues.disputeAmount) {
        const amount = parseFloat(formValues.disputeAmount);
        if (!isNaN(amount)) {
          data.disputes = [{
            id: 'generated-dispute-1',
            text: formValues.disputeText,
            amount: amount,
            created_at: new Date().toISOString()
          }];
        }
      }
      
      if (formValues.disputeCount) {
        const count = parseInt(formValues.disputeCount);
        if (!isNaN(count)) {
          data.count = count;
          data.attributes = {
            dispute_count: count,
            avg_amount: parseFloat(formValues.disputeAmount) || 0,
            dispute_timespan: '3 months'
          };
        }
      }
    }
    
    onSubmit(data);
  };
  
  // Render a single input field based on its type
  const renderField = (field: WorkflowInputField, prefix: string = '') => {
    const fieldId = prefix ? `${prefix}_${field.id}` : field.id;
    const value = formValues[fieldId] !== undefined ? formValues[fieldId] : field.defaultValue;
    
    switch (field.type) {
      case 'textarea':
        return (
          <div key={fieldId}>
            <Label htmlFor={fieldId}>{field.label}</Label>
            {field.description && <p className="text-xs text-muted-foreground mb-1">{field.description}</p>}
            <Textarea 
              id={fieldId} 
              placeholder={field.placeholder || ''}
              value={value || ''}
              onChange={(e) => handleInputChange(fieldId, e.target.value)}
              className="h-24"
            />
          </div>
        );
        
      case 'number':
        return (
          <div key={fieldId}>
            <Label htmlFor={fieldId}>{field.label}</Label>
            {field.description && <p className="text-xs text-muted-foreground mb-1">{field.description}</p>}
            <Input 
              id={fieldId} 
              type="number" 
              placeholder={field.placeholder || ''}
              value={value || ''}
              onChange={(e) => handleInputChange(fieldId, e.target.value)}
            />
          </div>
        );
        
      case 'checkbox':
        return (
          <div key={fieldId} className="flex items-center space-x-2">
            <Checkbox 
              id={fieldId} 
              checked={!!value}
              onCheckedChange={(checked) => handleInputChange(fieldId, checked)}
            />
            <Label htmlFor={fieldId}>{field.label}</Label>
            {field.description && <p className="text-xs text-muted-foreground">{field.description}</p>}
          </div>
        );
        
      case 'select':
        return (
          <div key={fieldId}>
            <Label htmlFor={fieldId}>{field.label}</Label>
            {field.description && <p className="text-xs text-muted-foreground mb-1">{field.description}</p>}
            <Select
              value={value || ''}
              onValueChange={(val: string) => handleInputChange(fieldId, val)}
            >
              <SelectTrigger id={fieldId}>
                <SelectValue placeholder={field.placeholder || 'Select an option'} />
              </SelectTrigger>
              <SelectContent>
                {field.options?.map(option => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        );
        
      case 'fileUpload':
        // Simplified file upload - in a real app, you'd implement proper file handling
        return (
          <div key={fieldId}>
            <Label htmlFor={fieldId}>{field.label}</Label>
            {field.description && <p className="text-xs text-muted-foreground mb-1">{field.description}</p>}
            <Input 
              id={fieldId} 
              type="file" 
              onChange={(e) => {
                const files = e.target.files;
                if (files && files.length > 0) {
                  // In a real implementation, you would process the file
                  handleInputChange(fieldId, files[0]);
                }
              }}
            />
            <p className="text-xs text-muted-foreground mt-1">
              Note: File upload is for demonstration purposes. Files aren't processed in this demo.
            </p>
          </div>
        );
        
      case 'text':
      default:
        return (
          <div key={fieldId}>
            <Label htmlFor={fieldId}>{field.label}</Label>
            {field.description && <p className="text-xs text-muted-foreground mb-1">{field.description}</p>}
            <Input 
              id={fieldId} 
              type="text" 
              placeholder={field.placeholder || ''}
              value={value || ''}
              onChange={(e) => handleInputChange(fieldId, e.target.value)}
            />
          </div>
        );
    }
  };
  
  // Get default workflow configuration when none is provided
  const getDefaultWorkflowConfig = (): WorkflowConfig => {
    return {
      id: 'default',
      name: 'Default Workflow',
      description: 'A simple workflow for dispute analysis',
      inputTabs: [
        {
          id: 'basicData',
          label: 'Basic Data',
          dataSourceConfigs: [
            {
              id: 'manualInput',
              name: 'Manual Input',
              description: 'Enter dispute information manually',
              fields: [
                {
                  id: 'disputeText',
                  label: 'Example Dispute Text',
                  type: 'textarea',
                  placeholder: 'I was charged a $35 overdraft fee but I had sufficient funds in my account.',
                  defaultValue: '',
                  required: true
                },
                {
                  id: 'disputeAmount',
                  label: 'Dispute Amount ($)',
                  type: 'number',
                  placeholder: '35.00',
                  defaultValue: '',
                  required: true
                },
                {
                  id: 'disputeCount',
                  label: 'Number of Disputes',
                  type: 'number',
                  placeholder: '10',
                  defaultValue: '10',
                  required: true
                }
              ]
            }
          ]
        }
      ],
      parameters: [
        {
          id: 'analysisParams',
          label: 'Analysis Parameters',
          fields: [
            {
              id: 'batchSize',
              label: 'Batch Size',
              type: 'number',
              description: 'Number of disputes to process in each batch',
              defaultValue: '10',
              required: false
            },
            {
              id: 'enableDebug',
              label: 'Enable Debug Mode',
              type: 'checkbox',
              defaultValue: false,
              required: false
            }
          ]
        }
      ]
    };
  };
  
  if (isLoading) {
    return (
      <div className="bg-card border shadow-lg rounded-lg p-6 max-w-2xl w-full">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-semibold">Loading Workflow Configuration...</h2>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X size={16} />
          </Button>
        </div>
        <div className="flex justify-center p-6">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </div>
    );
  }
  
  if (!workflowConfig) {
    return (
      <div className="bg-card border shadow-lg rounded-lg p-6 max-w-2xl w-full">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-semibold">Error Loading Configuration</h2>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X size={16} />
          </Button>
        </div>
        <p className="text-muted-foreground mb-4">
          Failed to load workflow configuration. Please try again or contact support.
        </p>
        <div className="flex justify-end pt-2">
          <Button 
            type="button" 
            variant="outline" 
            onClick={onClose}
            className="mr-2"
          >
            Close
          </Button>
        </div>
      </div>
    );
  }
  
  return (
    <div className="bg-card border shadow-lg rounded-lg p-6 max-w-2xl w-full overflow-y-auto max-h-[80vh]">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-lg font-semibold">{workflowConfig.name}</h2>
        <Button variant="ghost" size="sm" onClick={onClose}>
          <X size={16} />
        </Button>
      </div>
      
      {workflowConfig.description && (
        <p className="text-muted-foreground mb-4">{workflowConfig.description}</p>
      )}
      
      <form onSubmit={handleSubmit} className="space-y-4">
        <Tabs defaultValue={activeTab} value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="mb-4">
            <TabsTrigger value="data">Data Sources</TabsTrigger>
            <TabsTrigger value="parameters">Parameters</TabsTrigger>
          </TabsList>
          
          <TabsContent value="data" className="space-y-4">
            {workflowConfig.inputTabs.map(tab => (
              <div key={tab.id} className="space-y-4">
                <h3 className="font-medium text-sm">{tab.label}</h3>
                
                {tab.dataSourceConfigs.map(dataSource => (
                  <div key={dataSource.id} className="border rounded-md p-4">
                    <div className="flex items-start mb-3">
                      <Checkbox 
                        id={`select_${dataSource.id}`}
                        checked={!!selectedDataSources[dataSource.id]}
                        onCheckedChange={(checked) => 
                          handleDataSourceToggle(dataSource.id, !!checked)
                        }
                        className="mr-2 mt-1"
                      />
                      <div>
                        <Label htmlFor={`select_${dataSource.id}`} className="font-medium">
                          {dataSource.name}
                        </Label>
                        {dataSource.description && (
                          <p className="text-xs text-muted-foreground">{dataSource.description}</p>
                        )}
                      </div>
                    </div>
                    
                    {selectedDataSources[dataSource.id] && (
                      <div className="ml-6 space-y-3">
                        {dataSource.fields.map(field => renderField(field, dataSource.id))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            ))}
          </TabsContent>
          
          <TabsContent value="parameters" className="space-y-4">
            {workflowConfig.parameters.map(paramGroup => (
              <div key={paramGroup.id} className="space-y-3">
                <h3 className="font-medium text-sm">{paramGroup.label}</h3>
                <div className="border rounded-md p-4 space-y-3">
                  {paramGroup.fields.map(field => renderField(field))}
                </div>
              </div>
            ))}
          </TabsContent>
        </Tabs>
        
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