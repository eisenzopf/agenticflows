import { useRef, useState, useCallback, forwardRef, useImperativeHandle, useEffect, useMemo } from 'react';
import ReactFlow, {
  Node,
  Edge,
  addEdge,
  Background,
  Controls,
  MiniMap,
  Connection,
  useNodesState,
  useEdgesState,
  ReactFlowInstance,
  XYPosition,
  ConnectionLineType,
  Position,
  NodeMouseHandler,
  EdgeMouseHandler
} from 'reactflow';
import 'reactflow/dist/style.css';
import ComponentsPanel from './ComponentsPanel';
import FunctionSettingsPanel from './FunctionSettingsPanel';
import EdgeSettingsPanel from './EdgeSettingsPanel';
import WorkflowInputForm from './WorkflowInputForm';
import WorkflowResultsViewer from './WorkflowResultsViewer';
import { api, FunctionItem } from '@/services/api';
import { Button } from '@/components/ui/button';

// Data flow interface for function connections
interface DataFlowMapping {
  sourceOutput: string;
  targetInput: string;
}

// Custom edge data for function connections
interface FunctionEdgeData {
  label?: string;
  mappings: DataFlowMapping[];
}

// Define a custom Edge with our data type
type CustomEdge = Edge<FunctionEdgeData>;

// Start with an empty canvas
const initialNodes: Node[] = [];
const initialEdges: Edge[] = [];

// Default edge options for horizontal flow
const defaultEdgeOptions = {
  type: 'smoothstep', 
  animated: true,
  style: { stroke: '#64748B' },
};

// Node types based on component type
const getNodeStyle = (type: string) => {
  switch (type) {
    case 'agent':
      return { 
        background: 'rgba(59, 130, 246, 0.1)', 
        borderColor: '#3B82F6',
        borderWidth: '2px',
        padding: '10px',
        borderRadius: '8px',
        color: '#1E40AF',
        fontWeight: 500,
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
      };
    case 'tool':
      return { 
        background: 'rgba(245, 158, 11, 0.1)', 
        borderColor: '#F59E0B',
        borderWidth: '2px',
        padding: '10px',
        borderRadius: '8px',
        color: '#B45309',
        fontWeight: 500,
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
      };
    case 'function':
      return { 
        background: 'rgba(16, 185, 129, 0.1)', 
        borderColor: '#10B981',
        borderWidth: '2px',
        padding: '10px',
        borderRadius: '8px',
        color: '#065F46',
        fontWeight: 500,
        boxShadow: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
      };
    default:
      return {};
  }
};

// Define the handle types we want to expose
export interface FlowEditorHandle {
  handleSaveWorkflow: () => Promise<{success: boolean, workflowName: string}>;
  loadWorkflow: (workflowId: string) => Promise<boolean>;
  executeWorkflow: (initialData?: Record<string, any>) => Promise<any>;
}

interface SavedWorkflowData {
  nodes: Node[];
  edges: Edge[];
  viewport?: {
    x: number;
    y: number;
    zoom: number;
  };
}

// Remove the inline EdgeSettingsPanel component definition since we're importing it
// const EdgeSettingsPanel = ({ edge, sourceFunction, targetFunction, onClose, updateMappings }: {
//   edge: CustomEdge;
//   sourceFunction: FunctionItem;
//   targetFunction: FunctionItem;
//   onClose: () => void;
//   updateMappings: (edgeId: string, mappings: DataFlowMapping[]) => void;
// }) => {
//   return <div>Edge Settings Panel Placeholder</div>;
// };

const FlowEditor = forwardRef<FlowEditorHandle, {}>((props, ref) => {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState<FunctionEdgeData>(initialEdges);
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const [activeWorkflowId, setActiveWorkflowId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [selectedFunction, setSelectedFunction] = useState<FunctionItem | null>(null);
  const [functionComponents, setFunctionComponents] = useState<FunctionItem[]>([]);
  const [selectedEdge, setSelectedEdge] = useState<CustomEdge | null>(null);
  const [sourceFunctionItem, setSourceFunctionItem] = useState<FunctionItem | null>(null);
  const [targetFunctionItem, setTargetFunctionItem] = useState<FunctionItem | null>(null);
  const [executionResults, setExecutionResults] = useState<Record<string, any> | null>(null);
  const [isExecuting, setIsExecuting] = useState(false);
  const [showInputForm, setShowInputForm] = useState(false);
  const [showResultsViewer, setShowResultsViewer] = useState(false);

  // Fetch function components when component mounts
  useEffect(() => {
    async function fetchFunctions() {
      try {
        const functions = await api.getFunctions();
        setFunctionComponents(functions);
      } catch (error) {
        console.error("Error fetching functions:", error);
      }
    }
    
    fetchFunctions();
  }, []);

  // Handle keyboard events for node deletion
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Delete' && selectedNode) {
        // Remove the selected node
        setNodes((nds) => nds.filter((node) => node.id !== selectedNode.id));
        
        // Also remove any connected edges
        setEdges((eds) => eds.filter(
          (edge) => edge.source !== selectedNode.id && edge.target !== selectedNode.id
        ));
        
        // Clear the selected node
        setSelectedNode(null);
        setSelectedFunction(null);
      }
    };

    // Add event listener
    window.addEventListener('keydown', handleKeyDown);
    
    // Clean up
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [selectedNode, setNodes, setEdges]);

  // Apply horizontal orientation to any nodes
  const ensureHorizontalOrientation = useCallback((nodes: Node[]) => {
    return nodes.map(node => ({
      ...node,
      sourcePosition: Position.Right,
      targetPosition: Position.Left,
    }));
  }, []);

  // Load initial workflows from backend when component mounts
  useEffect(() => {
    const fetchInitialWorkflow = async () => {
      if (!reactFlowInstance) return;
      
      try {
        setIsLoading(true);
        // Fetch all workflows
        const workflows = await api.getWorkflows();
        
        // If there are workflows, load the first one
        if (workflows && workflows.length > 0) {
          const firstWorkflow = workflows[0];
          await loadWorkflow(firstWorkflow.id);
        } else {
          // Start with an empty canvas if no workflows exist
          setNodes([]);
          setEdges([]);
        }
      } catch (error) {
        console.error('Error loading initial workflow:', error);
        // Start with an empty canvas on error
        setNodes([]);
        setEdges([]);
      } finally {
        setIsLoading(false);
      }
    };
    
    fetchInitialWorkflow();
  }, [reactFlowInstance]);

  // Handle node click to show the settings panel
  const onNodeClick: NodeMouseHandler = useCallback((event, node) => {
    // Get the node type from data
    const nodeType = node.data?.nodeType || node.type;
    
    if (nodeType === 'function') {
      // For function nodes, we need to find the corresponding function item
      const functionId = node.data?.functionId;
      
      if (functionId) {
        const functionItem = functionComponents.find(f => f.id === functionId);
        if (functionItem) {
          setSelectedFunction(functionItem);
          setSelectedNode(node);
          return;
        }
      }
      
      // Extract function ID from node ID or data
      let extractedId = null;
      
      // Try to extract from node ID if it has a prefix like "function-"
      if (node.id.includes('function-')) {
        // Get the part after "function-" prefix and before any potential numbering
        const match = node.id.match(/function-([^-\d]+)/);
        if (match && match[1]) {
          extractedId = `analysis-${match[1]}`;
        }
      }
      
      // If we couldn't extract from ID, try to extract from label
      if (!extractedId && node.data?.label) {
        const label = node.data.label.toLowerCase();
        // Find function with matching label or name
        const functionItem = functionComponents.find(f => 
          f.label.toLowerCase() === label
        );
        
        if (functionItem) {
          setSelectedFunction(functionItem);
          setSelectedNode(node);
          return;
        }
      }
      
      // Try to find by extracted ID
      if (extractedId) {
        const functionItem = functionComponents.find(f => f.id === extractedId);
        if (functionItem) {
          setSelectedFunction(functionItem);
          setSelectedNode(node);
          return;
        }
      }
    } else {
      // For other node types, just set the selected node
      setSelectedNode(node);
      setSelectedFunction(null);
    }
  }, [functionComponents]);

  // Handle background click to deselect nodes
  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
    setSelectedFunction(null);
    setSelectedEdge(null);
    setSourceFunctionItem(null);
    setTargetFunctionItem(null);
  }, []);

  // Close the function settings panel
  const closeFunctionSettings = useCallback(() => {
    setSelectedFunction(null);
  }, []);

  // Load workflow by ID from the backend API
  const loadWorkflow = useCallback(async (workflowId: string) => {
    if (!reactFlowInstance) return false;
    setIsLoading(true);
    
    try {
      // Fetch workflow from the backend API
      const workflow = await api.getWorkflow(workflowId);
      
      if (workflow) {
        // Parse nodes and edges if they're strings
        const parsedNodes = Array.isArray(workflow.nodes) 
          ? workflow.nodes 
          : (typeof workflow.nodes === 'string' ? JSON.parse(workflow.nodes) : []);
          
        const parsedEdges = Array.isArray(workflow.edges) 
          ? workflow.edges 
          : (typeof workflow.edges === 'string' ? JSON.parse(workflow.edges) : []);
        
        // Update nodes with horizontal orientation
        const nodesWithHorizontalFlow = ensureHorizontalOrientation(parsedNodes || []);
        
        // Set nodes and edges
        setNodes(nodesWithHorizontalFlow);
        setEdges(parsedEdges || []);
        setActiveWorkflowId(workflowId);
        
        // Apply consistent zoom level and fit view
        reactFlowInstance.setViewport({ x: 0, y: 0, zoom: 0.8 });
        setTimeout(() => reactFlowInstance.fitView({ padding: 0.4 }), 50);
        
        return true;
      }
      return false;
    } catch (error) {
      console.error(`Error loading workflow ${workflowId}:`, error);
      return false;
    } finally {
      setIsLoading(false);
    }
  }, [reactFlowInstance, setNodes, setEdges, ensureHorizontalOrientation]);

  // Handle saving the workflow to the backend API
  const handleSaveWorkflow = useCallback(async () => {
    if (!reactFlowInstance) return {success: false, workflowName: ''};
    setIsLoading(true);
    
    try {
      // Get current flow state
      const flowData = reactFlowInstance.toObject();
      
      // Make sure all nodes have horizontal orientation before saving
      flowData.nodes = ensureHorizontalOrientation(flowData.nodes);
      
      const currentDate = new Date().toLocaleDateString();
      let workflowName = '';
      
      if (activeWorkflowId) {
        // Update existing workflow
        try {
          // Get the current workflow data first
          const existingWorkflow = await api.getWorkflow(activeWorkflowId);
          workflowName = existingWorkflow.name;
          
          // Update the workflow with new nodes and edges
          await api.updateWorkflow(activeWorkflowId, {
            ...existingWorkflow,
            nodes: flowData.nodes,
            edges: flowData.edges,
            date: currentDate
          });
          
          console.log(`Updated workflow ${activeWorkflowId}`);
          return {success: true, workflowName};
        } catch (error) {
          // If the workflow doesn't exist yet, create a new one
          console.log("Workflow not found, creating new one");
        }
      }
      
      // Create a new workflow if no active ID or update failed
      const newWorkflowId = activeWorkflowId || `workflow-${Date.now()}`;
      workflowName = `Workflow ${new Date().toLocaleTimeString()}`;
      
      await api.createWorkflow({
        id: newWorkflowId,
        name: workflowName,
        date: currentDate,
        nodes: flowData.nodes,
        edges: flowData.edges
      });
      
      setActiveWorkflowId(newWorkflowId);
      console.log(`Created new workflow ${newWorkflowId}`);
      return {success: true, workflowName};
    } catch (error) {
      console.error('Error saving workflow:', error);
      return {success: false, workflowName: ''};
    } finally {
      setIsLoading(false);
    }
  }, [reactFlowInstance, ensureHorizontalOrientation, activeWorkflowId]);

  // Execute the current workflow
  const executeWorkflow = useCallback(async (initialData: Record<string, any> = {}) => {
    if (!activeWorkflowId) {
      alert('Please save the workflow first before executing.');
      return null;
    }
    
    setIsExecuting(true);
    setExecutionResults(null);
    
    try {
      // Save the workflow first
      await handleSaveWorkflow();
      
      // Execute the workflow
      const results = await api.executeWorkflow(activeWorkflowId, initialData);
      
      // Update state with results
      setExecutionResults(results);
      setShowResultsViewer(true);
      
      // Apply visual indicators for successful nodes
      const updatedNodes = nodes.map(node => {
        if (results.results[node.id]) {
          // Node was executed successfully
          return {
            ...node,
            data: {
              ...node.data,
              executionResult: results.results[node.id]
            },
            style: {
              ...node.style,
              boxShadow: '0 0 0 2px green'
            }
          };
        } else if (node.data?.nodeType === 'function') {
          // Function node but not executed
          return {
            ...node,
            style: {
              ...node.style,
              boxShadow: '0 0 0 2px gray'
            }
          };
        }
        return node;
      });
      
      setNodes(updatedNodes);
      return results;
    } catch (error) {
      console.error('Error executing workflow:', error);
      alert(`Error executing workflow: ${error}`);
      return null;
    } finally {
      setIsExecuting(false);
    }
  }, [activeWorkflowId, handleSaveWorkflow, nodes, setNodes]);

  // Expose methods to the parent component
  useImperativeHandle(ref, () => ({
    handleSaveWorkflow,
    loadWorkflow,
    executeWorkflow
  }));

  // Memoize React Flow props to prevent recreating objects on each render
  const onConnect = useCallback(
    (connection: Connection) => {
      // Get source and target nodes to determine function types
      const sourceNode = nodes.find(node => node.id === connection.source);
      const targetNode = nodes.find(node => node.id === connection.target);
      
      if (sourceNode && targetNode && 
          sourceNode.data?.nodeType === 'function' && 
          targetNode.data?.nodeType === 'function') {
        
        // Find source and target function items
        const sourceFuncId = sourceNode.data?.functionId;
        const targetFuncId = targetNode.data?.functionId;
        
        const sourceFunctionItem = functionComponents.find(f => f.id === sourceFuncId);
        const targetFunctionItem = functionComponents.find(f => f.id === targetFuncId);
        
        const newEdge = {
          ...connection,
          id: `edge-${Date.now()}`,
          type: 'smoothstep',
          animated: true,
          style: { stroke: '#64748B' },
          data: {
            label: 'Configure data flow',
            mappings: [] as DataFlowMapping[]
          }
        } as CustomEdge;
        
        // Add the edge
        setEdges((eds) => addEdge(newEdge, eds));
        
        // If we have function information, select the edge for configuration
        if (sourceFunctionItem && targetFunctionItem) {
          setSelectedEdge(newEdge);
          setSourceFunctionItem(sourceFunctionItem);
          setTargetFunctionItem(targetFunctionItem);
        }
      } else {
        // Regular connection for non-function nodes
        setEdges((eds) => addEdge({
          ...connection,
          type: 'smoothstep',
          animated: true,
          style: { stroke: '#64748B' }
        }, eds));
      }
    },
    [nodes, setEdges, functionComponents]
  );

  // Handle edge click to show edge settings panel
  const onEdgeClick: EdgeMouseHandler = useCallback((event, edge) => {
    // Find source and target nodes
    const sourceNode = nodes.find(node => node.id === edge.source);
    const targetNode = nodes.find(node => node.id === edge.target);
    
    if (sourceNode && targetNode && 
        sourceNode.data?.nodeType === 'function' && 
        targetNode.data?.nodeType === 'function') {
      
      // Find source and target function items
      const sourceFuncId = sourceNode.data?.functionId;
      const targetFuncId = targetNode.data?.functionId;
      
      const sourceFunctionItem = functionComponents.find(f => f.id === sourceFuncId);
      const targetFunctionItem = functionComponents.find(f => f.id === targetFuncId);
      
      if (sourceFunctionItem && targetFunctionItem) {
        setSelectedEdge(edge as CustomEdge);
        setSourceFunctionItem(sourceFunctionItem);
        setTargetFunctionItem(targetFunctionItem);
        // Deselect any selected node
        setSelectedNode(null);
        setSelectedFunction(null);
      }
    }
  }, [nodes, functionComponents]);

  // Close the edge settings panel
  const closeEdgeSettings = useCallback(() => {
    setSelectedEdge(null);
    setSourceFunctionItem(null);
    setTargetFunctionItem(null);
  }, []);

  // Update edge data mappings
  const updateEdgeDataMappings = useCallback((edgeId: string, mappings: DataFlowMapping[]) => {
    setEdges(eds => 
      eds.map(edge => {
        if (edge.id === edgeId) {
          return {
            ...edge,
            data: {
              ...edge.data,
              mappings
            }
          };
        }
        return edge;
      })
    );
  }, [setEdges]);

  const onDragOver = useCallback((event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent<HTMLDivElement>) => {
      event.preventDefault();

      if (!reactFlowWrapper.current || !reactFlowInstance) return;

      const reactFlowBounds = reactFlowWrapper.current.getBoundingClientRect();
      const dataStr = event.dataTransfer.getData('application/reactflow');
      
      if (!dataStr) return;

      try {
        const { type, label, id } = JSON.parse(dataStr);
        
        // Get drop position in the canvas
        const position: XYPosition = reactFlowInstance.project({
          x: event.clientX - reactFlowBounds.left,
          y: event.clientY - reactFlowBounds.top,
        });

        // For function nodes, use the id from drag data if available
        let functionId = id;
        
        // If no id was provided and it's a function, try to find by label
        if (!functionId && type === 'function') {
          const functionItem = functionComponents.find(f => f.label === label);
          functionId = functionItem?.id;
        }

        // Create a unique node ID
        const nodeId = `${type}-${Date.now()}`;

        // Create a new node
        const newNode: Node = {
          id: nodeId,
          type: 'default',
          position,
          data: { 
            label,
            nodeType: type,
            functionId
          },
          style: getNodeStyle(type),
          // Set source handles on the right, target handles on the left for horizontal flow
          sourcePosition: Position.Right,
          targetPosition: Position.Left,
        };

        setNodes((nds) => nds.concat(newNode));
      } catch (error) {
        console.error('Error adding new node:', error);
      }
    },
    [reactFlowInstance, setNodes, functionComponents]
  );
  
  // Memoize connectionLineStyle to prevent recreation on each render
  const connectionLineStyle = useMemo(() => ({ stroke: '#64748B' }), []);
  
  // Memoize defaultViewport to prevent recreation on each render
  const defaultViewport = useMemo(() => ({ x: 0, y: 0, zoom: 0.8 }), []);

  // Show input form when executing workflow
  const handleExecuteClick = useCallback(() => {
    setShowInputForm(true);
  }, []);
  
  // Handle workflow input form submission
  const handleWorkflowInputSubmit = useCallback((inputData: Record<string, any>) => {
    setShowInputForm(false);
    executeWorkflow(inputData);
  }, [executeWorkflow]);

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      <ComponentsPanel 
        onSaveWorkflow={handleSaveWorkflow} 
        onLoadWorkflow={loadWorkflow}
        activeWorkflowId={activeWorkflowId}
      />
      
      {/* Add workflow execution button */}
      {activeWorkflowId && (
        <div 
          style={{ 
            position: 'absolute', 
            top: '16px', 
            right: (selectedFunction || selectedEdge) ? '528px' : '16px',
            zIndex: 5,
            transition: 'right 0.3s ease-in-out'
          }}
        >
          <Button
            onClick={handleExecuteClick}
            disabled={isExecuting}
            className="bg-green-600 hover:bg-green-700 text-white"
          >
            {isExecuting ? 'Executing...' : 'Execute Workflow'}
          </Button>
        </div>
      )}
      
      {/* Workflow input form modal */}
      {showInputForm && (
        <div 
          style={{ 
            position: 'fixed', 
            top: 0, 
            left: 0, 
            right: 0, 
            bottom: 0, 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center', 
            backgroundColor: 'rgba(0, 0, 0, 0.5)',
            zIndex: 20
          }}
        >
          <WorkflowInputForm 
            onSubmit={handleWorkflowInputSubmit} 
            onClose={() => setShowInputForm(false)} 
          />
        </div>
      )}
      
      {/* Workflow results viewer */}
      {showResultsViewer && executionResults && (
        <div 
          style={{ 
            position: 'fixed', 
            top: 0, 
            left: 0, 
            right: 0, 
            bottom: 0, 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center', 
            backgroundColor: 'rgba(0, 0, 0, 0.5)',
            zIndex: 20
          }}
        >
          <WorkflowResultsViewer 
            results={executionResults} 
            onClose={() => setShowResultsViewer(false)} 
          />
        </div>
      )}
      
      <div 
        className="reactflow-wrapper" 
        ref={reactFlowWrapper} 
        style={{ 
          width: '100%', 
          height: '100%', 
          marginLeft: '64px', 
          marginRight: (selectedFunction || selectedEdge) ? '512px' : '0', 
          transition: 'margin-right 0.3s ease-in-out'
        }}
      >
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onInit={setReactFlowInstance}
          onDrop={onDrop}
          onDragOver={onDragOver}
          onNodeClick={onNodeClick}
          onEdgeClick={onEdgeClick}
          onPaneClick={onPaneClick}
          defaultViewport={defaultViewport}
          minZoom={0.5}
          maxZoom={4}
          defaultEdgeOptions={defaultEdgeOptions}
          connectionLineType={ConnectionLineType.SmoothStep}
          connectionLineStyle={connectionLineStyle}
          fitView
        >
          <Controls position="bottom-right" />
          <MiniMap 
            nodeStrokeWidth={3}
            zoomable
            pannable
          />
          <Background color="#aaa" gap={16} />
        </ReactFlow>
      </div>

      {selectedFunction && (
        <div 
          style={{ 
            position: 'absolute', 
            top: 0, 
            right: 0, 
            height: '100%', 
            width: '512px',
            zIndex: 10,
            boxShadow: '-2px 0px 10px rgba(0,0,0,0.1)'
          }}
        >
          <FunctionSettingsPanel 
            selectedFunction={selectedFunction} 
            onClose={closeFunctionSettings} 
          />
        </div>
      )}

      {selectedEdge && sourceFunctionItem && targetFunctionItem && (
        <div 
          style={{ 
            position: 'absolute', 
            top: 0, 
            right: 0, 
            height: '100%', 
            width: '512px',
            zIndex: 10,
            boxShadow: '-2px 0px 10px rgba(0,0,0,0.1)'
          }}
        >
          <EdgeSettingsPanel 
            edge={selectedEdge}
            sourceFunction={sourceFunctionItem}
            targetFunction={targetFunctionItem}
            onClose={closeEdgeSettings}
            updateMappings={updateEdgeDataMappings}
          />
        </div>
      )}
    </div>
  );
});

FlowEditor.displayName = 'FlowEditor';

export default FlowEditor; 