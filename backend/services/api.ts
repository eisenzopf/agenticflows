import axios from 'axios';

// API base URL
const API_BASE_URL = 'http://localhost:8080/api';

// Component item interface
export interface ComponentItem {
  id: string;
  type: string;
  label: string;
  description?: string;
}

// Function item interface (extends ComponentItem)
export interface FunctionItem extends ComponentItem {
  endpoint: string;
  analysisType?: string;
  parameters?: any[];
  returns?: any[];
  documentation?: string;
}

// Workflow data interface
export interface WorkflowData {
  id: string;
  name: string;
  date: string;
  nodes: any;
  edges: any;
}

// API class
export const api = {
  // Get all components (agents and tools)
  async getComponents(): Promise<{ agents: ComponentItem[], tools: ComponentItem[] }> {
    try {
      const agentsResponse = await axios.get(`${API_BASE_URL}/agents`);
      const toolsResponse = await axios.get(`${API_BASE_URL}/tools`);
      
      return {
        agents: agentsResponse.data || [],
        tools: toolsResponse.data || []
      };
    } catch (error) {
      console.error('Error fetching components:', error);
      return { agents: [], tools: [] };
    }
  },
  
  // Get all functions
  async getFunctions(): Promise<FunctionItem[]> {
    // For now, we'll return hardcoded functions
    // In a real implementation, this would fetch from the API
    return [
      {
        id: 'function-trends-analysis',
        type: 'function',
        label: 'Trends Analysis',
        description: 'Analyze trends in conversation data',
        endpoint: '/api/analysis',
        analysisType: 'trends',
        parameters: [
          { name: 'focus_areas', type: 'string[]', description: 'Areas to focus analysis on' },
          { name: 'data', type: 'object', description: 'Conversation data to analyze' }
        ],
        returns: [
          { name: 'trends', type: 'object[]', description: 'Identified trends' },
          { name: 'insights', type: 'string[]', description: 'Overall insights' }
        ]
      },
      {
        id: 'function-patterns-analysis',
        type: 'function',
        label: 'Patterns Analysis',
        description: 'Identify patterns in conversation data',
        endpoint: '/api/analysis',
        analysisType: 'patterns',
        parameters: [
          { name: 'pattern_types', type: 'string[]', description: 'Types of patterns to look for' },
          { name: 'data', type: 'object', description: 'Conversation data to analyze' }
        ],
        returns: [
          { name: 'patterns', type: 'object[]', description: 'Identified patterns' }
        ]
      },
      {
        id: 'function-findings-analysis',
        type: 'function',
        label: 'Findings Analysis',
        description: 'Analyze findings from attribute extraction',
        endpoint: '/api/analysis',
        analysisType: 'findings',
        parameters: [
          { name: 'questions', type: 'string[]', description: 'Questions to answer' },
          { name: 'data', type: 'object', description: 'Attribute data to analyze' }
        ],
        returns: [
          { name: 'answers', type: 'object[]', description: 'Answers to questions' },
          { name: 'data_gaps', type: 'string[]', description: 'Identified data gaps' }
        ]
      },
      {
        id: 'function-attributes-analysis',
        type: 'function',
        label: 'Extract Attributes',
        description: 'Extract attributes from text or generate required attributes',
        endpoint: '/api/analysis',
        analysisType: 'attributes',
        parameters: [
          { name: 'text', type: 'string', description: 'Text to extract attributes from' },
          { name: 'attributes', type: 'array', description: 'Attribute definitions to extract' },
          { name: 'generate_required', type: 'boolean', description: 'Whether to generate required attributes' }
        ],
        returns: [
          { name: 'attribute_values', type: 'object', description: 'Extracted attribute values' }
        ]
      },
      {
        id: 'function-intent-analysis',
        type: 'function',
        label: 'Generate Intent',
        description: 'Generate intent from text',
        endpoint: '/api/analysis',
        analysisType: 'intent',
        parameters: [
          { name: 'text', type: 'string', description: 'Text to extract intent from' }
        ],
        returns: [
          { name: 'label_name', type: 'string', description: 'Human-readable intent label' },
          { name: 'label', type: 'string', description: 'Machine-readable intent label' }
        ]
      },
      {
        id: 'function-chain-analysis',
        type: 'function',
        label: 'Chain Analysis',
        description: 'Perform a complete chain of analysis steps (attributes, trends, patterns, findings)',
        endpoint: '/api/analysis/chain',
        parameters: [
          { name: 'input_data', type: 'object', description: 'Input data for analysis (text or structured data)' },
          { name: 'config', type: 'object', description: 'Configuration for analysis chain' },
          { name: 'use_attributes', type: 'boolean', description: 'Whether to extract attributes from text first' },
          { name: 'focus_areas', type: 'string[]', description: 'Areas to focus trends analysis on' },
          { name: 'pattern_types', type: 'string[]', description: 'Types of patterns to look for' },
          { name: 'questions', type: 'string[]', description: 'Questions for findings analysis' }
        ],
        returns: [
          { name: 'attributes', type: 'object', description: 'Extracted attributes (if enabled)' },
          { name: 'trends', type: 'object', description: 'Trend analysis results' },
          { name: 'patterns', type: 'string[]', description: 'Identified patterns' },
          { name: 'findings', type: 'string[]', description: 'Analysis findings' },
          { name: 'recommendations', type: 'string[]', description: 'Recommendations based on analysis' }
        ]
      }
    ];
  },
  
  // Get all workflows
  async getWorkflows(): Promise<WorkflowData[]> {
    try {
      const response = await axios.get(`${API_BASE_URL}/workflows`);
      return response.data || [];
    } catch (error) {
      console.error('Error fetching workflows:', error);
      return [];
    }
  },
  
  // Get a specific workflow
  async getWorkflow(id: string): Promise<WorkflowData> {
    try {
      const response = await axios.get(`${API_BASE_URL}/workflows/${id}`);
      return response.data;
    } catch (error) {
      console.error(`Error fetching workflow ${id}:`, error);
      throw error;
    }
  },
  
  // Create a new workflow
  async createWorkflow(workflow: WorkflowData): Promise<WorkflowData> {
    try {
      const response = await axios.post(`${API_BASE_URL}/workflows`, workflow);
      return response.data;
    } catch (error) {
      console.error('Error creating workflow:', error);
      throw error;
    }
  },
  
  // Update a workflow
  async updateWorkflow(id: string, workflow: WorkflowData): Promise<WorkflowData> {
    try {
      const response = await axios.put(`${API_BASE_URL}/workflows/${id}`, workflow);
      return response.data;
    } catch (error) {
      console.error(`Error updating workflow ${id}:`, error);
      throw error;
    }
  },
  
  // Delete a workflow
  async deleteWorkflow(id: string): Promise<void> {
    try {
      await axios.delete(`${API_BASE_URL}/workflows/${id}`);
    } catch (error) {
      console.error(`Error deleting workflow ${id}:`, error);
      throw error;
    }
  },
  
  // Perform standard analysis using the unified API
  async performAnalysis(
    analysisType: string, 
    parameters: Record<string, any>, 
    data?: Record<string, any>, 
    text?: string,
    workflowId?: string
  ): Promise<any> {
    try {
      // Create a properly typed request object
      const request: {
        analysis_type: string;
        parameters: Record<string, any>;
        workflow_id?: string;
        text?: string;
        data?: Record<string, any>;
      } = {
        analysis_type: analysisType,
        parameters: parameters,
        workflow_id: workflowId
      };
      
      // Add text if provided
      if (text) {
        request.text = text;
      }
      
      // Add data if provided
      if (data) {
        request.data = data;
      }
      
      const response = await axios.post(`${API_BASE_URL}/analysis`, request);
      return response.data;
    } catch (error) {
      console.error(`Error performing ${analysisType} analysis:`, error);
      throw error;
    }
  },
  
  // Perform chain analysis
  async performChainAnalysis(workflowId: string, inputData: any, config: any): Promise<any> {
    try {
      const response = await axios.post(`${API_BASE_URL}/analysis/chain`, {
        workflow_id: workflowId,
        input_data: inputData,
        config: config
      });
      return response.data;
    } catch (error) {
      console.error('Error performing chain analysis:', error);
      throw error;
    }
  }
}; 