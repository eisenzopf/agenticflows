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
} from 'reactflow';
import 'reactflow/dist/style.css';
import ComponentsPanel from './ComponentsPanel';
import FunctionSettingsPanel from './FunctionSettingsPanel';
import { api, FunctionItem } from '@/services/api';

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

const FlowEditor = forwardRef<FlowEditorHandle, {}>((props, ref) => {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const [activeWorkflowId, setActiveWorkflowId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [selectedFunction, setSelectedFunction] = useState<FunctionItem | null>(null);
  const [functionComponents, setFunctionComponents] = useState<FunctionItem[]>([]);

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
      
      // Fallback to trying to extract from node ID if functionId not in data
      const extractedId = node.id.startsWith('function-') 
        ? node.id.replace(/^function-\d+/, 'analysis') 
        : null;
        
      if (extractedId) {
        // Try to find by extracted ID
        const functionItem = functionComponents.find(f => 
          f.id.includes(extractedId) || extractedId.includes(f.id.split('-')[1])
        );
        
        if (functionItem) {
          setSelectedFunction(functionItem);
          setSelectedNode(node);
          return;
        }
      }
      
      // Last resort: try to match by label
      if (node.data?.label) {
        const functionItem = functionComponents.find(f => 
          f.label.toLowerCase() === node.data.label.toLowerCase()
        );
        
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

  // Expose methods to the parent component
  useImperativeHandle(ref, () => ({
    handleSaveWorkflow,
    loadWorkflow
  }));

  // Memoize React Flow props to prevent recreating objects on each render
  const onConnect = useCallback(
    (connection: Connection) => {
      setEdges((eds) => addEdge({
        ...connection,
        type: 'smoothstep',
        animated: true,
        style: { stroke: '#64748B' }
      }, eds));
    },
    [setEdges]
  );

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

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      <ComponentsPanel 
        onSaveWorkflow={handleSaveWorkflow} 
        onLoadWorkflow={loadWorkflow}
        activeWorkflowId={activeWorkflowId}
      />
      <div 
        className="reactflow-wrapper" 
        ref={reactFlowWrapper} 
        style={{ 
          width: '100%', 
          height: '100%', 
          marginLeft: '64px', 
          marginRight: selectedFunction ? '380px' : '0', 
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
            width: '380px',
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
    </div>
  );
});

FlowEditor.displayName = 'FlowEditor';

export default FlowEditor; 