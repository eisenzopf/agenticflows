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

export interface FunctionItem {
  id: string;
  type: string;
  label: string;
  endpoint: string;
  description: string;
  analysisType?: string;
}

const API_URL = 'http://localhost:8080/api';

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
      // Create request object with analysis type and parameters
      const request: {
        analysis_type: string;
        parameters: Record<string, any>;
        workflow_id?: string;
        text?: string;
        data?: Record<string, any>;
      } = {
        analysis_type: analysisType,
        parameters: parameters
      };
      
      // Add optional fields if provided
      if (workflowId) request.workflow_id = workflowId;
      if (text) request.text = text;
      if (data) request.data = data;
      
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
}; 