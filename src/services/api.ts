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
    // Since we don't have an actual endpoint for functions yet, 
    // we'll hard-code the available analysis endpoints
    const functions: FunctionItem[] = [
      {
        id: 'analysis-trends',
        type: 'function',
        label: 'Analyze Trends',
        endpoint: '/api/analysis/trends',
        description: 'Analyze trends based on focus areas'
      },
      {
        id: 'analysis-patterns',
        type: 'function',
        label: 'Identify Patterns',
        endpoint: '/api/analysis/patterns',
        description: 'Identify patterns based on pattern types'
      },
      {
        id: 'analysis-findings',
        type: 'function',
        label: 'Analyze Findings',
        endpoint: '/api/analysis/findings',
        description: 'Analyze findings based on questions and attribute values'
      },
      {
        id: 'analysis-attributes',
        type: 'function',
        label: 'Extract Attributes',
        endpoint: '/api/analysis/attributes',
        description: 'Extract attribute values from text or generate required attributes'
      },
      {
        id: 'analysis-intent',
        type: 'function',
        label: 'Generate Intent',
        endpoint: '/api/analysis/intent',
        description: 'Generate intent from text'
      },
      {
        id: 'analysis-results',
        type: 'function',
        label: 'Get Analysis Results',
        endpoint: '/api/analysis/results',
        description: 'Get or delete analysis results for a workflow'
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
}; 