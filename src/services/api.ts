import { Node, Edge } from 'reactflow';

export interface FlowData {
  id: string;
  name: string;
  nodes: Node[];
  edges: Edge[];
}

const API_URL = 'http://localhost:8080/api';

export const api = {
  // Get all flows
  getFlows: async (): Promise<FlowData[]> => {
    const response = await fetch(`${API_URL}/flows`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch flows: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  // Get a specific flow
  getFlow: async (id: string): Promise<FlowData> => {
    const response = await fetch(`${API_URL}/flows/${id}`);
    
    if (!response.ok) {
      throw new Error(`Failed to fetch flow with id ${id}: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  // Create a new flow
  createFlow: async (flow: FlowData): Promise<FlowData> => {
    const response = await fetch(`${API_URL}/flows`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(flow),
    });
    
    if (!response.ok) {
      throw new Error(`Failed to create flow: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  // Update a flow
  updateFlow: async (id: string, flow: FlowData): Promise<FlowData> => {
    const response = await fetch(`${API_URL}/flows/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(flow),
    });
    
    if (!response.ok) {
      throw new Error(`Failed to update flow with id ${id}: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  // Delete a flow
  deleteFlow: async (id: string): Promise<void> => {
    const response = await fetch(`${API_URL}/flows/${id}`, {
      method: 'DELETE',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to delete flow with id ${id}: ${response.statusText}`);
    }
  },
}; 