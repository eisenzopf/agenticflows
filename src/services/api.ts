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

  // Execute a workflow of connected functions
  executeWorkflow: async (workflowId: string, initialData: Record<string, any>): Promise<any> => {
    try {
      // Get the workflow definition
      const workflow = await api.getWorkflow(workflowId);
      
      if (!workflow || !workflow.nodes || !workflow.edges) {
        throw new Error('Invalid workflow');
      }
      
      // Parse nodes and edges if they're strings
      const nodes: WorkflowNode[] = Array.isArray(workflow.nodes) 
        ? workflow.nodes 
        : (typeof workflow.nodes === 'string' ? JSON.parse(workflow.nodes) : []);
        
      const edges: WorkflowEdge[] = Array.isArray(workflow.edges) 
        ? workflow.edges 
        : (typeof workflow.edges === 'string' ? JSON.parse(workflow.edges) : []);
      
      // Find all function nodes
      const functionNodes = nodes.filter(node => node.data?.nodeType === 'function');
      
      // Sort nodes into a dependency order for execution
      const sortedNodes = getExecutionOrder(functionNodes, edges);
      
      // Initialize results storage
      let results: Record<string, any> = { ...initialData };
      
      // Execute each node in order
      for (const node of sortedNodes) {
        const functionId = node.data?.functionId;
        const functionType = functionId?.split('-')[1] || '';
        
        // Get the incoming edges for this node
        const incomingEdges = edges.filter(edge => edge.target === node.id);
        
        // Prepare function parameters from edge mappings
        const parameters: Record<string, any> = {};
        let inputData: Record<string, any> = {};
        
        // Process each incoming edge to get inputs
        for (const edge of incomingEdges) {
          const mappings = edge.data?.mappings || [];
          const sourceNodeId = edge.source;
          
          // Get results from the source node
          const sourceResults = results[sourceNodeId];
          
          if (sourceResults) {
            // Apply the mappings
            for (const mapping of mappings) {
              if (mapping && sourceResults[mapping.sourceOutput]) {
                inputData[mapping.targetInput] = sourceResults[mapping.sourceOutput];
              }
            }
          }
        }
        
        // Merge with initial data if needed
        inputData = { ...inputData, ...initialData };
        
        // Execute the function based on its type
        let functionResult;
        try {
          functionResult = await api.performAnalysis(
            functionType,
            parameters,
            inputData,
            inputData.text,
            workflowId
          );
          
          // Store the results indexed by node id
          results[node.id] = functionResult.results || {};
        } catch (error) {
          console.error(`Error executing function node ${node.id}:`, error);
          results[node.id] = { error: `Failed to execute: ${error}` };
        }
      }
      
      return {
        workflow_id: workflowId,
        results
      };
    } catch (error) {
      console.error('Error executing workflow:', error);
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