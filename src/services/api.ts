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