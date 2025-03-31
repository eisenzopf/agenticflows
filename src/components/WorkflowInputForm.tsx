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
          description: 'Use a SQLite database with conversation data',
          fields: [
            {
              id: 'dbPath',
              label: 'Database Path',
              type: 'text' as 'text',
              placeholder: '/path/to/your/database.db',
              defaultValue: '',
              description: 'Path to the SQLite database file',
              required: true
            },
            {
              id: 'maxConversations',
              label: 'Maximum Conversations',
              type: 'number' as 'number',
              placeholder: '100',
              defaultValue: '100',
              description: 'Maximum number of conversations to analyze',
              required: false
            },
            {
              id: 'conversationLimit',
              label: 'Example Limit',
              type: 'number' as 'number',
              placeholder: '5',
              defaultValue: '5',
              description: 'Number of example conversations to include in output',
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
    
    // Get conversation text from manual input if selected
    let conversationText = '';
    if (formValues.conversationText || (formValues.manualInput_conversationText && selectedDataSources.manualInput)) {
      // Get the conversation text from either old format or new format
      conversationText = formValues.conversationText || formValues.manualInput_conversationText;
    }
    
    // If no conversation text was provided, use a sample
    if (!conversationText) {
      conversationText = "This is a sample conversation about a customer service interaction. The customer was asking about order status and delivery timeframe. The agent was able to provide the expected delivery date and helped track the package.";
    }
    
    // Prepare questions list from form input
    const questions = formValues.questions 
      ? formValues.questions.split('\n').filter((q: string) => q.trim().length > 0)
      : ['What is the main topic?', 'What is the customer sentiment?', 'What are the key points?'];
    
    // Define default attributes
    const defaultAttributes = [
      {
        field_name: "topic",
        title: "Topic",
        description: "Main topic of the conversation"
      },
      {
        field_name: "sentiment",
        title: "Sentiment",
        description: "Overall sentiment of the conversation"
      },
      {
        field_name: "key_points",
        title: "Key Points",
        description: "Important points mentioned in the conversation"
      }
    ];
    
    // Add question-based attributes
    const questionAttributes = questions.map((question: string, index: number) => ({
      field_name: `question_${index + 1}`,
      title: `Question ${index + 1}`,
      description: question
    }));
    
    // Combine all attributes
    const allAttributes = [...defaultAttributes, ...questionAttributes];
    
    // Determine if we're inside a workflow with specific node types
    const hasWorkflowId = !!workflowId;
    
    // Create the appropriate request based on context
    let finalData: Record<string, any>;
    
    if (hasWorkflowId) {
      // We're in a workflow context, so include whatever the workflow nodes need
      finalData = {
        // Add the user-provided data 
        text: conversationText,
        
        // Include attributes at the root level for any direct attributes analysis
        attributes: allAttributes,
        
        // Add additional structured data for the server-side execution
        data: {
          text: conversationText,
          conversations: [{
            id: 'manual-input-1',
            text: conversationText,
            created_at: new Date().toISOString()
          }]
        },
        
        // Also include parameters for backward compatibility
        parameters: {
          attributes: allAttributes,
          questions: questions
        }
      };
    } else {
      // Direct API call (not in workflow)
      finalData = {
        analysis_type: "attributes",
        text: conversationText,
        attributes: allAttributes
      };
    }
    
    // Log the data being sent
    console.log("Submitting workflow data:", JSON.stringify(finalData, null, 2));
    
    onSubmit(finalData);
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
      name: 'Conversation Analysis Workflow',
      description: 'Configure data sources and analysis parameters',
      inputTabs: [
        {
          id: 'basicData',
          label: 'Data Sources',
          dataSourceConfigs: [
            {
              id: 'databaseSource',
              name: 'SQLite Database',
              description: 'Use a SQLite database with conversation data',
              fields: [
                {
                  id: 'dbPath',
                  label: 'Database Path',
                  type: 'text' as 'text',
                  placeholder: '/path/to/your/database.db',
                  defaultValue: '',
                  description: 'Path to the SQLite database file',
                  required: true
                },
                {
                  id: 'maxConversations',
                  label: 'Maximum Conversations',
                  type: 'number' as 'number',
                  placeholder: '100',
                  defaultValue: '100',
                  description: 'Maximum number of conversations to analyze',
                  required: false
                },
                {
                  id: 'conversationLimit',
                  label: 'Example Limit',
                  type: 'number' as 'number',
                  placeholder: '5',
                  defaultValue: '5',
                  description: 'Number of example conversations to include in output',
                  required: false
                }
              ]
            },
            {
              id: 'manualInput',
              name: 'Manual Input',
              description: 'Enter conversation data manually',
              fields: [
                {
                  id: 'conversationText',
                  label: 'Conversation Text',
                  type: 'textarea' as 'textarea',
                  placeholder: 'Enter conversation text here...',
                  defaultValue: '',
                  description: 'Paste one or more conversations to analyze',
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
          label: 'Analysis Settings',
          fields: [
            {
              id: 'batchSize',
              label: 'Batch Size',
              type: 'number' as 'number',
              description: 'Number of items to process in each batch',
              defaultValue: '10',
              required: false
            },
            {
              id: 'enableDebug',
              label: 'Enable Debug Mode',
              type: 'checkbox' as 'checkbox',
              defaultValue: false,
              required: false
            }
          ]
        },
        {
          id: 'questionParams',
          label: 'Analysis Questions',
          fields: [
            {
              id: 'questions',
              label: 'Questions to Answer',
              type: 'textarea' as 'textarea',
              description: 'Enter questions for the analysis (one per line)',
              placeholder: 'What are the common patterns?\nWhat recommendations can be made?',
              defaultValue: 'What are the most common topics discussed?\nWhat patterns can be identified?\nWhat are the key areas for improvement?\nWhat recommendations would you suggest?',
              required: true
            },
            {
              id: 'focusAreas',
              label: 'Focus Areas (Optional)',
              type: 'text' as 'text',
              description: 'Comma-separated list of areas to focus on',
              defaultValue: 'customer_satisfaction,response_time,resolution_effectiveness',
              required: false
            },
            {
              id: 'patternTypes',
              label: 'Pattern Types (Optional)',
              type: 'text' as 'text',
              description: 'Comma-separated list of pattern types to identify',
              defaultValue: 'conversation_flow,resolution_patterns,customer_behavior',
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
                              This option allows you to analyze conversations from a SQLite database.
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