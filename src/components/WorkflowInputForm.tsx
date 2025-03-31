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
      try {
        setIsLoading(true);
        let config: WorkflowConfig;
        
        if (!workflowId) {
          // Use default configuration if no workflowId
          config = getDefaultWorkflowConfig();
        } else {
          // Fetch workflow configuration from the API
          try {
            config = await api.getWorkflowExecutionConfig(workflowId);
          } catch (error) {
            console.error("Error loading workflow configuration:", error);
            // Fall back to default config
            config = getDefaultWorkflowConfig();
          }
        }
        
        // Ensure the database source is always available
        ensureDatabaseSourceExists(config);
        
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
        
        // Pre-select database source if available
        setSelectedDataSources({
          databaseSource: true
        });
      } catch (error) {
        console.error("Error in workflow configuration setup:", error);
      } finally {
        setIsLoading(false);
      }
    };
    
    loadWorkflowConfig();
  }, [workflowId]);
  
  // Helper function to ensure database source exists
  const ensureDatabaseSourceExists = (config: WorkflowConfig) => {
    if (!config || !config.inputTabs || config.inputTabs.length === 0) return;
    
    // Check if database source already exists
    const hasDatabaseSource = config.inputTabs.some(tab => 
      tab.dataSourceConfigs.some(source => source.id === 'databaseSource')
    );
    
    if (!hasDatabaseSource) {
      // Get the data sources tab (usually the first tab)
      const dataTab = config.inputTabs.find(tab => 
        tab.label.toLowerCase().includes('data') || 
        tab.dataSourceConfigs.length > 0
      ) || config.inputTabs[0];
      
      // Add the database source
      if (dataTab) {
        const dbSource: DataSourceConfig = {
          id: 'databaseSource',
          name: 'SQLite Database',
          description: 'Use a SQLite database file for dispute analysis',
          fields: [
            {
              id: 'dbPath',
              label: 'Database Path',
              type: 'text' as 'text',
              placeholder: '/path/to/your/database.db',
              defaultValue: '',
              description: 'Path to the SQLite database file to analyze',
              required: true
            },
            {
              id: 'maxDisputes',
              label: 'Maximum Disputes',
              type: 'number' as 'number',
              placeholder: '100',
              defaultValue: '100',
              description: 'Maximum number of disputes to analyze',
              required: false
            },
            {
              id: 'conversationLimit',
              label: 'Conversation Limit',
              type: 'number' as 'number',
              placeholder: '5',
              defaultValue: '5',
              description: 'Number of example conversations to include',
              required: false
            }
          ]
        };
        
        // Insert at the beginning
        dataTab.dataSourceConfigs.unshift(dbSource);
      }
    }
  };
  
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
            
            // Special handling for database source to match analyze_fee_disputes script
            if (dataSource.id === 'databaseSource') {
              const dbPath = formValues[`databaseSource_dbPath`];
              const maxDisputes = parseInt(formValues[`databaseSource_maxDisputes`] || '100');
              const conversationLimit = parseInt(formValues[`databaseSource_conversationLimit`] || '5');
              const batchSize = parseInt(formValues.batchSize || '10');
              const enableDebug = !!formValues.enableDebug;
              
              // Format for API in the same structure as the Go script
              data.dbConfig = {
                dbPath,
                maxDisputes,
                conversationLimit,
                batchSize,
                enableDebug,
                workflowId: workflowId || '',
              };
              
              // Add the SQLite query parameters that are used in the script
              data.sqlQueries = {
                disputesQuery: `
                SELECT 
                  conversation_id,
                  text,
                  COALESCE(date_time, CURRENT_TIMESTAMP) as date_time
                FROM conversations
                WHERE text IS NOT NULL 
                AND LENGTH(text) > 100
                AND (
                  text LIKE '%fee%'
                  OR text LIKE '%charge%'
                  OR text LIKE '%billing%'
                  OR text LIKE '%refund%'
                  OR text LIKE '%dispute%'
                )
                ORDER BY RANDOM()
                LIMIT ?
                `,
                conversationsQuery: `
                SELECT 
                  conversation_id,
                  text,
                  COALESCE(date_time, CURRENT_TIMESTAMP) as date_time
                FROM conversations
                WHERE text IS NOT NULL 
                AND LENGTH(text) > 200
                ORDER BY RANDOM()
                LIMIT ?
                `
              };
              
              // Advanced parameters
              if (formValues.focusAreas) {
                data.focusAreas = formValues.focusAreas.split(',').map((item: string) => item.trim());
              }
              
              if (formValues.patternTypes) {
                data.patternTypes = formValues.patternTypes.split(',').map((item: string) => item.trim());
              }
            }
          }
        });
      });
    }
    
    // Add backward compatibility for old code
    if (formValues.disputeText || (formValues.manualInput_disputeText && selectedDataSources.manualInput)) {
      // Get the dispute text from either old format or new format
      const disputeText = formValues.disputeText || formValues.manualInput_disputeText;
      const disputeAmount = formValues.disputeAmount || formValues.manualInput_disputeAmount;
      const disputeCount = formValues.disputeCount || formValues.manualInput_disputeCount;
      
      data.text = disputeText;
      
      if (disputeAmount) {
        const amount = parseFloat(disputeAmount);
        if (!isNaN(amount)) {
          data.disputes = [{
            id: 'generated-dispute-1',
            text: disputeText,
            amount: amount,
            created_at: new Date().toISOString()
          }];
        }
      }
      
      if (disputeCount) {
        const count = parseInt(disputeCount);
        if (!isNaN(count)) {
          data.count = count;
          data.attributes = {
            dispute_count: count,
            avg_amount: parseFloat(disputeAmount) || 0,
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
    
    // Special handling for database path field to add file browser
    if (prefix === 'databaseSource' && field.id === 'dbPath') {
      return (
        <div key={fieldId}>
          <Label htmlFor={fieldId}>{field.label}</Label>
          {field.description && <p className="text-xs text-muted-foreground mb-1">{field.description}</p>}
          <div className="flex space-x-2">
            <Input 
              id={fieldId} 
              type="text" 
              placeholder={field.placeholder || ''}
              value={value || ''}
              onChange={(e) => handleInputChange(fieldId, e.target.value)}
              className="flex-1"
            />
            <Button 
              type="button"
              variant="outline"
              size="sm"
              onClick={() => {
                // Create a file input element
                const fileInput = document.createElement('input');
                fileInput.type = 'file';
                fileInput.accept = '.db,.sqlite,.sqlite3';
                
                // Handle file selection
                fileInput.onchange = (e) => {
                  const files = (e.target as HTMLInputElement).files;
                  if (files && files[0]) {
                    // Get the file name since path isn't available for security reasons
                    handleInputChange(fieldId, files[0].name);
                  }
                };
                
                // Trigger the file dialog
                fileInput.click();
              }}
            >
              Browse...
            </Button>
          </div>
        </div>
      );
    }
    
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
          label: 'Data Sources',
          dataSourceConfigs: [
            {
              id: 'databaseSource',
              name: 'SQLite Database',
              description: 'Use a SQLite database file for dispute analysis',
              fields: [
                {
                  id: 'dbPath',
                  label: 'Database Path',
                  type: 'text',
                  placeholder: '/path/to/your/database.db',
                  defaultValue: '',
                  description: 'Path to the SQLite database file to analyze',
                  required: true
                },
                {
                  id: 'maxDisputes',
                  label: 'Maximum Disputes',
                  type: 'number',
                  placeholder: '100',
                  defaultValue: '100',
                  description: 'Maximum number of disputes to analyze',
                  required: false
                },
                {
                  id: 'conversationLimit',
                  label: 'Conversation Limit',
                  type: 'number',
                  placeholder: '5',
                  defaultValue: '5',
                  description: 'Number of example conversations to include',
                  required: false
                }
              ]
            },
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
        },
        {
          id: 'feeDisputeParams',
          label: 'Fee Dispute Analysis',
          fields: [
            {
              id: 'focusAreas',
              label: 'Focus Areas',
              type: 'text',
              description: 'Comma-separated list of areas to focus on',
              defaultValue: 'fee_dispute_trends,customer_impact,financial_impact',
              required: false
            },
            {
              id: 'patternTypes',
              label: 'Pattern Types',
              type: 'text',
              description: 'Comma-separated list of pattern types to identify',
              defaultValue: 'fee_dispute_patterns,resolution_patterns,customer_behavior_patterns',
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
                        {dataSource.id === 'databaseSource' && (
                          <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-md mb-3 text-sm">
                            <h4 className="font-medium text-blue-800 dark:text-blue-300 mb-1">Using SQLite Database</h4>
                            <p className="text-blue-700 dark:text-blue-400 mb-2">
                              This option allows you to analyze fee disputes from a SQLite database, similar to how the command-line tool works.
                            </p>
                            <div className="text-xs text-blue-600 dark:text-blue-400">
                              <p className="mb-1"><strong>Expected Database Schema:</strong> The database should have a <code>conversations</code> table with at least these columns:</p>
                              <ul className="list-disc pl-5 mb-2">
                                <li><code>conversation_id</code>: Unique identifier for the conversation</li>
                                <li><code>text</code>: The conversation text content</li>
                                <li><code>date_time</code>: When the conversation occurred</li>
                              </ul>
                              <p><strong>Example Path:</strong> For MacOS/Linux: <code>/path/to/your/database.db</code> or for Windows: <code>C:\path\to\your\database.db</code></p>
                            </div>
                          </div>
                        )}
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