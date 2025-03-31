import { Node, Edge } from 'reactflow';

export interface FlowData {
  id: string;
  name: string;
  nodes: Node[];
  edges: Edge[];
}

export interface WorkflowData extends FlowData {
  date: string;
}

export interface ComponentItem {
  id: string;
  type: string;
  label: string;
}

// Define types for workflow execution
interface WorkflowNode {
  id: string;
  data: {
    nodeType?: string;
    functionId?: string;
    [key: string]: any;
  };
}

interface WorkflowEdge {
  id: string;
  source: string;
  target: string;
  data?: {
    mappings?: Array<{
      sourceOutput: string;
      targetInput: string;
    }>;
    [key: string]: any;
  };
}

// Define types for workflow input configuration
export interface WorkflowInputField {
  id: string;
  label: string;
  type: 'text' | 'number' | 'textarea' | 'checkbox' | 'select' | 'fileUpload';
  description?: string;
  placeholder?: string;
  defaultValue?: any;
  required?: boolean;
  options?: { value: string; label: string }[]; // For select fields
}

export interface DataSourceConfig {
  id: string;
  name: string;
  description: string;
  fields: WorkflowInputField[];
}

export interface WorkflowInputConfig {
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

const API_URL = (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080') + '/api';

export interface FunctionItem {
  id: string;
  label: string;
  type: string;
  description: string;
  endpoint?: string;
  analysisType?: string;
  inputs?: Array<{
    name: string;
    description: string;
    required?: boolean;
  }>;
  outputs?: Array<{
    name: string;
    description: string;
  }>;
}

// Add new types for function metadata
export interface ParameterDefinition {
  name: string;
  path: string;
  description: string;
  required: boolean;
  type: string;
}

export interface OutputDefinition {
  name: string;
  path: string;
  description: string;
  type: string;
}

export interface FunctionMetadata {
  id: string;
  label: string;
  description: string;
  inputs: ParameterDefinition[];
  outputs: OutputDefinition[];
  example?: Record<string, any>;
}

// Initialize cache as empty object instead of null
let functionMetadataCache: Record<string, FunctionMetadata> = {};
let workflowConfigCache: Record<string, WorkflowInputConfig> = {};

export const api = {
  // Get all components (agents and tools)
  getComponents: async (): Promise<{ agents: ComponentItem[], tools: ComponentItem[] }> => {
    const [agentsResponse, toolsResponse] = await Promise.all([
      fetch(`${API_URL}/agents`),
      fetch(`${API_URL}/tools`),
    ]);
    
    if (!agentsResponse.ok) {
      throw new Error(`Failed to fetch agents: ${agentsResponse.statusText}`);
    }
    
    if (!toolsResponse.ok) {
      throw new Error(`Failed to fetch tools: ${toolsResponse.statusText}`);
    }
    
    const agents = await agentsResponse.json();
    const tools = await toolsResponse.json();
    
    return { agents, tools };
  },
  
  // Get all functions
  getFunctions: async (): Promise<FunctionItem[]> => {
    // Return functions using the consolidated /api/analysis endpoint
    const functions: FunctionItem[] = [
      {
        id: 'analysis-trends',
        type: 'function',
        label: 'Analyze Trends',
        endpoint: '/api/analysis',
        description: 'Analyze trends based on focus areas',
        analysisType: 'trends'
      },
      {
        id: 'analysis-patterns',
        type: 'function',
        label: 'Identify Patterns',
        endpoint: '/api/analysis',
        description: 'Identify patterns based on pattern types',
        analysisType: 'patterns'
      },
      {
        id: 'analysis-findings',
        type: 'function',
        label: 'Analyze Findings',
        endpoint: '/api/analysis',
        description: 'Analyze findings based on questions and attribute values',
        analysisType: 'findings'
      },
      {
        id: 'analysis-attributes',
        type: 'function',
        label: 'Extract Attributes',
        endpoint: '/api/analysis',
        description: 'Extract attribute values from text or generate required attributes',
        analysisType: 'attributes'
      },
      {
        id: 'analysis-intent',
        type: 'function',
        label: 'Generate Intent',
        endpoint: '/api/analysis',
        description: 'Generate intent from text',
        analysisType: 'intent'
      },
      {
        id: 'analysis-recommendations',
        type: 'function',
        label: 'Generate Recommendations',
        endpoint: '/api/analysis',
        description: 'Generate actionable recommendations based on analysis results',
        analysisType: 'recommendations'
      },
      {
        id: 'analysis-plan',
        type: 'function',
        label: 'Create Action Plan',
        endpoint: '/api/analysis',
        description: 'Create implementation plan from recommendations with timeline and resources',
        analysisType: 'plan'
      },
      {
        id: 'analysis-chain',
        type: 'function',
        label: 'Chain Analysis',
        endpoint: '/api/analysis/chain',
        description: 'Perform a complete chain of analysis steps'
      }
    ];
    
    return functions;
  },
  
  // Get all agents
  getAgents: async (): Promise<ComponentItem[]> => {
    const response = await fetch(`${API_URL}/agents`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch agents: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  // Get all tools
  getTools: async (): Promise<ComponentItem[]> => {
    const response = await fetch(`${API_URL}/tools`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch tools: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  // Add a new agent
  addAgent: async (agent: ComponentItem): Promise<ComponentItem> => {
    const response = await fetch(`${API_URL}/agents`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(agent),
    });
    
    if (!response.ok) {
      throw new Error(`Failed to add agent: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  // Add a new tool
  addTool: async (tool: ComponentItem): Promise<ComponentItem> => {
    const response = await fetch(`${API_URL}/tools`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(tool),
    });
    
    if (!response.ok) {
      throw new Error(`Failed to add tool: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  // Get all workflows
  getWorkflows: async (): Promise<WorkflowData[]> => {
    const response = await fetch(`${API_URL}/workflows`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch workflows: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  // Get a specific workflow
  getWorkflow: async (id: string): Promise<WorkflowData> => {
    const response = await fetch(`${API_URL}/workflows/${id}`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch workflow with id ${id}: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  // Create a new workflow
  createWorkflow: async (workflow: WorkflowData): Promise<WorkflowData> => {
    const response = await fetch(`${API_URL}/workflows`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(workflow),
    });
    
    if (!response.ok) {
      throw new Error(`Failed to create workflow: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  // Update a workflow
  updateWorkflow: async (id: string, workflow: WorkflowData): Promise<WorkflowData> => {
    const response = await fetch(`${API_URL}/workflows/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(workflow),
    });
    
    if (!response.ok) {
      throw new Error(`Failed to update workflow with id ${id}: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  // Delete a workflow
  deleteWorkflow: async (id: string): Promise<void> => {
    const response = await fetch(`${API_URL}/workflows/${id}`, {
      method: 'DELETE',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to delete workflow with id ${id}: ${response.statusText}`);
    }
  },

  // Perform standard analysis using the unified API
  performAnalysis: async (
    analysisType: string, 
    parameters: Record<string, any>, 
    data?: Record<string, any>, 
    text?: string,
    workflowId?: string
  ): Promise<any> => {
    try {
      // Create request object with analysis type
      const request: Record<string, any> = {
        analysis_type: analysisType
      };
      
      // Add parameters or attributes directly based on what's provided
      if (parameters) {
        // For attributes analysis, handle the attributes array directly
        if (analysisType === 'attributes' && parameters.attributes) {
          request.attributes = parameters.attributes; 
        } else {
          request.parameters = parameters;
        }
      }
      
      // Add optional fields if provided
      if (workflowId) request.workflow_id = workflowId;
      if (text) request.text = text;
      if (data) request.data = data;
      
      console.log(`Sending ${analysisType} request:`, JSON.stringify(request, null, 2));
      
      // Make the API call to the unified endpoint
      const response = await fetch(`${API_URL}/analysis`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
      });
      
      if (!response.ok) {
        throw new Error(`Analysis failed: ${response.statusText}`);
      }
      
      return response.json();
    } catch (error) {
      console.error(`Error performing ${analysisType} analysis:`, error);
      throw error;
    }
  },

  // Perform chain analysis using the dedicated endpoint
  performChainAnalysis: async (workflowId: string, inputData: any, config: any): Promise<any> => {
    try {
      const response = await fetch(`${API_URL}/analysis/chain`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          workflow_id: workflowId,
          input_data: inputData,
          config: config
        }),
      });
      
      if (!response.ok) {
        throw new Error(`Chain analysis failed: ${response.statusText}`);
      }
      
      return response.json();
    } catch (error) {
      console.error('Error performing chain analysis:', error);
      throw error;
    }
  },

  // Execute a workflow of connected functions
  executeWorkflow: async (workflowId: string, inputData: Record<string, any> = {}): Promise<any> => {
    try {
      console.log(`Executing workflow ${workflowId} with input data:`, inputData);
      
      // Call the backend API to execute the workflow
      const response = await fetch(`${API_URL}/workflows/${workflowId}/execute`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          data: inputData,
          text: inputData.text,
          parameters: inputData.parameters
        }),
      });
      
      if (!response.ok) {
        throw new Error(`Failed to execute workflow: ${response.statusText}`);
      }
      
      return response.json();
    } catch (error) {
      console.error('Error executing workflow:', error);
      throw error;
    }
  },

  // Get metadata for all analysis functions
  getFunctionMetadata: async (): Promise<Record<string, FunctionMetadata>> => {
    if (Object.keys(functionMetadataCache).length > 0) {
      return functionMetadataCache;
    }

    const response = await fetch(`${API_URL}/analysis/metadata`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch function metadata: ${response.statusText}`);
    }
    
    const metadata = await response.json();
    functionMetadataCache = metadata;
    return metadata;
  },

  // Get metadata for a specific function
  getFunctionMetadataById: async (functionId: string): Promise<FunctionMetadata | null> => {
    console.log('Getting metadata for function ID:', functionId);
    
    const metadata = await api.getFunctionMetadata();
    const analysisType = functionId.split('-')[1];
    
    console.log('Looking up analysis type:', analysisType, 'Available types:', Object.keys(metadata));
    
    const result = metadata[analysisType] || null;
    console.log('Found metadata:', result ? 'yes' : 'no');
    
    return result;
  },

  // Get workflow execution configuration
  getWorkflowExecutionConfig: async (workflowId: string): Promise<WorkflowInputConfig> => {
    // Check cache first
    if (workflowConfigCache[workflowId]) {
      return workflowConfigCache[workflowId];
    }
    
    try {
      // Try to fetch from the backend
      const response = await fetch(`${API_URL}/workflows/${workflowId}/execution-config`);
      
      if (response.ok) {
        const config = await response.json();
        // Store in cache
        workflowConfigCache[workflowId] = config;
        return config;
      }
    } catch (error) {
      console.error("Error fetching workflow execution config:", error);
    }
    
    // If no config exists or there was an error, generate based on workflow analysis
    const workflow = await api.getWorkflow(workflowId);
    const generatedConfig = await generateWorkflowConfig(workflow);
    
    // Store in cache
    workflowConfigCache[workflowId] = generatedConfig;
    return generatedConfig;
  },

  // Generate a workflow from description
  generateWorkflow: async (name: string, description: string): Promise<WorkflowData> => {
    try {
      const response = await fetch(`${API_URL}/workflows/generate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name, description }),
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Failed to generate workflow (${response.status})`);
      }
      
      return response.json();
    } catch (error) {
      console.error('Error generating workflow:', error);
      throw error;
    }
  },

  // Generate a dynamic workflow from description
  generateDynamicWorkflow: async (name: string, description: string): Promise<WorkflowData> => {
    try {
      const response = await fetch(`${API_URL}/workflows/generate-dynamic`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name, description }),
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Failed to generate dynamic workflow (${response.status})`);
      }
      
      return response.json();
    } catch (error) {
      console.error('Error generating dynamic workflow:', error);
      throw error;
    }
  },

  // Answer questions about banking data
  answerQuestions: async (
    questions: string[], 
    databasePath: string = '/Users/jonathan/Documents/Work/discourse_ai/Research/corpora/banking_2025/db/standard_charter_bank.db', 
    context?: string
  ): Promise<{ answers: Array<{ question: string, answer: string }> }> => {
    try {
      const response = await fetch(`${API_URL}/questions/answer`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 
          questions, 
          databasePath, 
          context
        }),
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `Failed to answer questions (${response.status})`);
      }
      
      return response.json();
    } catch (error) {
      console.error('Error answering questions:', error);
      throw error;
    }
  },
};

// Helper function to determine execution order of nodes based on dependencies
function getExecutionOrder(nodes: WorkflowNode[], edges: WorkflowEdge[]): WorkflowNode[] {
  // Create a map of node dependencies
  const dependencies: Record<string, string[]> = {};
  const nodeMap: Record<string, WorkflowNode> = {};
  
  // Initialize with empty dependencies
  nodes.forEach(node => {
    dependencies[node.id] = [];
    nodeMap[node.id] = node;
  });
  
  // Add dependencies based on edges
  edges.forEach(edge => {
    if (dependencies[edge.target]) {
      dependencies[edge.target].push(edge.source);
    }
  });
  
  // Topological sort
  const visited: Record<string, boolean> = {};
  const temp: Record<string, boolean> = {}; // For cycle detection
  const result: WorkflowNode[] = [];
  
  // DFS function for topological sort
  function dfs(nodeId: string) {
    // Skip if already visited
    if (visited[nodeId]) return;
    
    // Check for cycles
    if (temp[nodeId]) {
      throw new Error('Workflow contains cycles, which are not supported');
    }
    
    // Mark as temporarily visited
    temp[nodeId] = true;
    
    // Visit all dependencies first
    for (const dep of dependencies[nodeId]) {
      dfs(dep);
    }
    
    // Mark as visited
    visited[nodeId] = true;
    temp[nodeId] = false;
    
    // Add to result
    if (nodeMap[nodeId]) {
      result.push(nodeMap[nodeId]);
    }
  }
  
  // Start DFS from all nodes
  for (const nodeId of Object.keys(dependencies)) {
    if (!visited[nodeId]) {
      dfs(nodeId);
    }
  }
  
  return result;
}

// Helper function to generate workflow configuration based on workflow nodes and edges
async function generateWorkflowConfig(workflow: WorkflowData): Promise<WorkflowInputConfig> {
  const nodes = workflow.nodes;
  const edges = workflow.edges;
  
  // Default config
  const config: WorkflowInputConfig = {
    id: workflow.id,
    name: workflow.name || 'Workflow Execution',
    description: 'Input configuration for workflow execution',
    inputTabs: [
      {
        id: 'basicData',
        label: 'Data Sources',
        dataSourceConfigs: []
      }
    ],
    parameters: [
      {
        id: 'executionParams',
        label: 'Execution Parameters',
        fields: [
          {
            id: 'batchSize',
            label: 'Batch Size',
            type: 'number',
            description: 'Number of items to process in each batch',
            defaultValue: '10',
            required: false
          },
          {
            id: 'debugMode',
            label: 'Enable Debug Mode',
            type: 'checkbox',
            defaultValue: false,
            required: false
          }
        ]
      }
    ]
  };
  
  // Find database nodes if any
  const databaseNodes = nodes.filter(node => 
    node.data?.nodeType === 'tool' && 
    (node.data?.label?.toLowerCase().includes('database') || 
     node.data?.label?.toLowerCase().includes('db'))
  );
  
  // Find analysis functions if any
  const analysisNodes = nodes.filter(node => 
    node.data?.nodeType === 'function' && 
    node.data?.functionId?.includes('analysis-')
  );
  
  // Add database connection config if needed
  if (databaseNodes.length > 0) {
    config.inputTabs[0].dataSourceConfigs.push({
      id: 'databaseSource',
      name: 'Database Connection',
      description: 'Configure database connection for data retrieval',
      fields: [
        {
          id: 'dbPath',
          label: 'Database Path',
          type: 'text',
          description: 'Path to the SQLite database file',
          required: true
        },
        {
          id: 'maxItems',
          label: 'Maximum Items',
          type: 'number',
          description: 'Maximum number of items to retrieve',
          defaultValue: '100',
          required: false
        }
      ]
    });
  }
  
  // Add manual input for data
  config.inputTabs[0].dataSourceConfigs.push({
    id: 'manualInput',
    name: 'Manual Input',
    description: 'Enter data manually for workflow execution',
    fields: [
      {
        id: 'text',
        label: 'Input Text',
        type: 'textarea',
        placeholder: 'Enter text to analyze...',
        required: false
      }
    ]
  });
  
  // Customize for analysis types
  if (analysisNodes.length > 0) {
    // Check for trends analysis
    if (analysisNodes.some(node => node.data?.functionId === 'analysis-trends')) {
      config.parameters.push({
        id: 'trendsParams',
        label: 'Trends Analysis',
        fields: [
          {
            id: 'focusAreas',
            label: 'Focus Areas',
            type: 'text',
            description: 'Comma-separated list of focus areas for trend analysis',
            defaultValue: 'customer_impact,financial_impact',
            required: false
          }
        ]
      });
    }
    
    // Check for patterns analysis
    if (analysisNodes.some(node => node.data?.functionId === 'analysis-patterns')) {
      config.parameters.push({
        id: 'patternsParams',
        label: 'Patterns Analysis',
        fields: [
          {
            id: 'patternTypes',
            label: 'Pattern Types',
            type: 'text',
            description: 'Comma-separated list of pattern types to identify',
            defaultValue: 'behavior_patterns,resolution_patterns',
            required: false
          }
        ]
      });
    }
    
    // Check for findings analysis
    if (analysisNodes.some(node => node.data?.functionId === 'analysis-findings')) {
      config.parameters.push({
        id: 'findingsParams',
        label: 'Findings Analysis',
        fields: [
          {
            id: 'questions',
            label: 'Analysis Questions',
            type: 'textarea',
            description: 'Enter questions for findings analysis (one per line)',
            defaultValue: 'What are the most common patterns?\nWhat are the key areas for improvement?',
            required: false
          }
        ]
      });
    }
  }
  
  return config;
} 