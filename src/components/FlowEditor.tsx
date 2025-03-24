import { useRef, useState, useCallback, forwardRef, useImperativeHandle, useEffect } from 'react';
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
} from 'reactflow';
import 'reactflow/dist/style.css';
import ComponentsPanel from './ComponentsPanel';

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
    default:
      return {};
  }
};

// Define the handle types we want to expose
export interface FlowEditorHandle {
  handleSaveWorkflow: () => boolean;
}

const FlowEditor = forwardRef<FlowEditorHandle, {}>((props, ref) => {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);
  const reactFlowWrapper = useRef<HTMLDivElement>(null);

  // Load saved flow if available and set horizontal orientation
  useEffect(() => {
    try {
      const savedFlow = localStorage.getItem('savedFlow');
      
      if (savedFlow && reactFlowInstance) {
        const flow = JSON.parse(savedFlow);
        
        // Update nodes with horizontal orientation
        const nodesWithHorizontalFlow = flow.nodes.map((node: Node) => ({
          ...node,
          sourcePosition: Position.Right,
          targetPosition: Position.Left,
        }));
        
        // Set nodes and edges from saved state
        setNodes(nodesWithHorizontalFlow || []);
        setEdges(flow.edges || []);
      }
    } catch (error) {
      console.error('Error loading saved flow:', error);
    }
  }, [reactFlowInstance, setNodes, setEdges]);

  // Apply horizontal orientation to any nodes
  const ensureHorizontalOrientation = useCallback((nodes: Node[]) => {
    return nodes.map(node => ({
      ...node,
      sourcePosition: Position.Right,
      targetPosition: Position.Left,
    }));
  }, []);

  // Handle saving the workflow
  const handleSaveWorkflow = useCallback(() => {
    if (reactFlowInstance) {
      // Get current flow state
      const flowData = reactFlowInstance.toObject();
      
      // Make sure all nodes have horizontal orientation before saving
      flowData.nodes = ensureHorizontalOrientation(flowData.nodes);
      
      console.log('Saving flow data:', flowData);
      
      // Store in localStorage for demo purposes
      localStorage.setItem('savedFlow', JSON.stringify(flowData));
      
      return true;
    }
    return false;
  }, [reactFlowInstance, ensureHorizontalOrientation]);

  // Expose methods to the parent component
  useImperativeHandle(ref, () => ({
    handleSaveWorkflow
  }));

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
        const { type, label } = JSON.parse(dataStr);
        
        // Get drop position in the canvas
        const position: XYPosition = reactFlowInstance.project({
          x: event.clientX - reactFlowBounds.left,
          y: event.clientY - reactFlowBounds.top,
        });

        // Create a new node
        const newNode: Node = {
          id: `${type}-${Date.now()}`,
          type: 'default',
          position,
          data: { label },
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
    [reactFlowInstance, setNodes]
  );

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      <ComponentsPanel onSaveWorkflow={handleSaveWorkflow} />
      <div 
        className="reactflow-wrapper" 
        ref={reactFlowWrapper} 
        style={{ width: '100%', height: '100%', marginLeft: '64px' }}
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
          defaultViewport={{ x: 0, y: 0, zoom: 1.5 }}
          minZoom={0.5}
          maxZoom={4}
          defaultEdgeOptions={defaultEdgeOptions}
          connectionLineType={ConnectionLineType.SmoothStep}
          connectionLineStyle={{ stroke: '#64748B' }}
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
    </div>
  );
});

FlowEditor.displayName = 'FlowEditor';

export default FlowEditor; 