import { useState, useEffect } from 'react';
import { ChevronLeft, ChevronRight, ChevronDown, ChevronUp, Plus, Pencil, Copy, Trash2, Save } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { api, ComponentItem, WorkflowData, FunctionItem } from '@/services/api';

// Define the workflow type
interface Workflow {
  id: string;
  name: string;
  date: string;
}

// Component card styling based on type
const getComponentStyle = (type: string) => {
  switch (type) {
    case 'agent':
      return 'border-l-4 border-l-blue-400 bg-blue-50 dark:bg-blue-950/30 dark:border-l-blue-600';
    case 'tool':
      return 'border-l-4 border-l-amber-400 bg-amber-50 dark:bg-amber-950/30 dark:border-l-amber-600';
    case 'function':
      return 'border-l-4 border-l-emerald-400 bg-emerald-50 dark:bg-emerald-950/30 dark:border-l-emerald-600';
    default:
      return '';
  }
};

interface ComponentsPanelProps {
  onSaveWorkflow?: () => Promise<{success: boolean, workflowName: string}>;
  onLoadWorkflow?: (workflowId: string) => Promise<boolean>;
  activeWorkflowId?: string | null;
}

export default function ComponentsPanel({ onSaveWorkflow, onLoadWorkflow, activeWorkflowId }: ComponentsPanelProps) {
  const [isExpanded, setIsExpanded] = useState(true);
  const [agentComponents, setAgentComponents] = useState<ComponentItem[]>([]);
  const [toolComponents, setToolComponents] = useState<ComponentItem[]>([]);
  const [functionComponents, setFunctionComponents] = useState<FunctionItem[]>([]);
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [activeWorkflow, setActiveWorkflow] = useState<string>(activeWorkflowId || '');
  
  // Update activeWorkflow when activeWorkflowId changes from parent
  if (activeWorkflowId && activeWorkflowId !== activeWorkflow) {
    setActiveWorkflow(activeWorkflowId);
  }
  
  // Expandable sections state
  const [sectionsExpanded, setSectionsExpanded] = useState({
    workflows: true,
    agents: true,
    tools: true,
    functions: true
  });

  useEffect(() => {
    // Fetch components when the component mounts
    async function fetchData() {
      setIsLoading(true);
      try {
        // Fetch components (agents, tools, and functions)
        const componentsData = await api.getComponents();
        setAgentComponents(componentsData.agents || []);
        setToolComponents(componentsData.tools || []);
        
        // Fetch functions
        const functionsData = await api.getFunctions();
        setFunctionComponents(functionsData || []);
        
        // Fetch workflows
        const workflowsData = await api.getWorkflows();
        
        if (workflowsData && workflowsData.length > 0) {
          // Convert API response to expected workflow format
          const formattedWorkflows = workflowsData.map(workflow => ({
            id: workflow.id,
            name: workflow.name,
            date: workflow.date
          }));
          
          setWorkflows(formattedWorkflows);
          
          // Set active workflow if none is selected
          if (!activeWorkflow) {
            setActiveWorkflow(formattedWorkflows[0].id);
          }
        } else {
          // If no workflows exist, create a default one
          const defaultWorkflow = {
            id: `workflow-${Date.now()}`,
            name: 'Default Workflow',
            date: new Date().toLocaleDateString()
          };
          setWorkflows([defaultWorkflow]);
          setActiveWorkflow(defaultWorkflow.id);
        }
      } catch (error) {
        console.error("Error fetching components:", error);
      } finally {
        setIsLoading(false);
      }
    }
    
    fetchData();
  }, []);

  const toggleSection = (section: 'workflows' | 'agents' | 'tools' | 'functions') => {
    setSectionsExpanded(prev => ({
      ...prev,
      [section]: !prev[section]
    }));
  };

  const onDragStart = (event: React.DragEvent<HTMLDivElement>, nodeType: string, nodeLabel: string) => {
    const data = JSON.stringify({ type: nodeType, label: nodeLabel });
    event.dataTransfer.setData('application/reactflow', data);
    event.dataTransfer.effectAllowed = 'move';
  };

  const handleAddWorkflow = async () => {
    const newWorkflow = {
      id: `workflow-${Date.now()}`,
      name: `New Workflow ${workflows.length + 1}`,
      date: new Date().toLocaleDateString()
    };
    
    try {
      // Create the workflow in the API with empty nodes and edges
      await api.createWorkflow({
        ...newWorkflow,
        nodes: [],
        edges: []
      });
      
      // Update local state
      setWorkflows([...workflows, newWorkflow]);
      setActiveWorkflow(newWorkflow.id);
      
      // Notify user of successful creation
      alert(`"${newWorkflow.name}" created successfully`);
      
      // Load the empty workflow if load function exists
      if (onLoadWorkflow) {
        await onLoadWorkflow(newWorkflow.id);
      }
    } catch (error) {
      console.error("Error creating workflow:", error);
    }
  };

  const handleSaveWorkflow = async () => {
    // This would typically save the current flow state
    if (onSaveWorkflow) {
      const result = await onSaveWorkflow();
      
      if (result.success) {
        // If the save was successful and returned a name, use it
        const workflowName = result.workflowName || 'Current workflow';
        
        // Update UI with notification (you might want to use a toast notification library in a real app)
        alert(`${workflowName} saved successfully`);
      }
    }
    
    // For demo purposes, we'll just update the date
    const updatedWorkflows = workflows.map(wf => 
      wf.id === activeWorkflow 
        ? { ...wf, date: new Date().toLocaleDateString() } 
        : wf
    );
    setWorkflows(updatedWorkflows);
  };

  const handleCloneWorkflow = async (id: string) => {
    const workflowToClone = workflows.find(wf => wf.id === id);
    if (workflowToClone) {
      const newWorkflow: Workflow = {
        id: `workflow-${Date.now()}`,
        name: `${workflowToClone.name} (Copy)`,
        date: new Date().toLocaleDateString()
      };
      
      try {
        // Fetch the original workflow to get nodes and edges
        const originalWorkflow = await api.getWorkflow(id);
        
        // Create a copy in the database
        await api.createWorkflow({
          ...newWorkflow,
          nodes: originalWorkflow.nodes,
          edges: originalWorkflow.edges
        });
        
        // Update local state
        setWorkflows([...workflows, newWorkflow]);
        
        // Notify user of successful clone
        alert(`"${workflowToClone.name}" cloned successfully as "${newWorkflow.name}"`);
      } catch (error) {
        console.error("Error cloning workflow:", error);
      }
    }
  };

  const handleRenameWorkflow = async (id: string) => {
    // Find current workflow name before renaming
    const workflowToRename = workflows.find(wf => wf.id === id);
    const oldName = workflowToRename?.name || '';
    
    const newName = prompt('Enter new workflow name:', oldName);
    if (newName && newName.trim() !== '') {
      try {
        // Find the workflow to update
        const workflowToUpdate = workflows.find(wf => wf.id === id);
        if (workflowToUpdate) {
          // Get the full workflow data
          const fullWorkflow = await api.getWorkflow(id);
          
          // Update the workflow in the API
          const updatedWorkflow = {
            ...fullWorkflow,
            name: newName.trim()
          };
          
          await api.updateWorkflow(id, updatedWorkflow);
          
          // Update local state
          setWorkflows(workflows.map(wf => 
            wf.id === id ? { ...wf, name: newName.trim() } : wf
          ));
          
          // Notify user of successful rename
          alert(`Workflow renamed from "${oldName}" to "${newName.trim()}"`);
        }
      } catch (error) {
        console.error("Error renaming workflow:", error);
      }
    }
  };

  const handleDeleteWorkflow = async (id: string) => {
    if (workflows.length <= 1) {
      alert('Cannot delete the only workflow.');
      return;
    }
    
    // Find the workflow name before deleting
    const workflowToDelete = workflows.find(wf => wf.id === id);
    const workflowName = workflowToDelete?.name || 'Selected workflow';
    
    const confirmed = confirm(`Are you sure you want to delete "${workflowName}"?`);
    if (confirmed) {
      try {
        // Delete the workflow from the API
        await api.deleteWorkflow(id);
        
        // Update local state
        const newWorkflows = workflows.filter(wf => wf.id !== id);
        setWorkflows(newWorkflows);
        
        // Show confirmation message
        alert(`"${workflowName}" was deleted successfully`);
        
        // If the active workflow was deleted, select the first one
        if (activeWorkflow === id) {
          setActiveWorkflow(newWorkflows[0].id);
        }
      } catch (error) {
        console.error("Error deleting workflow:", error);
      }
    }
  };
  
  const handleWorkflowDoubleClick = async (id: string) => {
    if (onLoadWorkflow) {
      try {
        const success = await onLoadWorkflow(id);
        if (success) {
          setActiveWorkflow(id);
        }
      } catch (error) {
        console.error("Error loading workflow:", error);
      }
    }
  };

  if (isLoading) {
    return (
      <div className="component-panel transition-all duration-300 bg-card border-r h-full flex flex-col w-64">
        <div className="flex justify-between items-center p-2 border-b">
          <span className="font-medium">Components</span>
          <Button variant="ghost" size="sm" className="ml-auto">
            <ChevronLeft size={16} />
          </Button>
        </div>
        <div className="p-4 text-center">
          Loading components...
        </div>
      </div>
    );
  }

  return (
    <div 
      className={`component-panel transition-all duration-300 bg-card border-r h-full flex flex-col ${isExpanded ? 'w-64' : 'w-12'}`} 
      style={{ position: 'absolute', top: 0, left: 0, zIndex: 10 }}
    >
      <div className="flex justify-between items-center p-2 border-b">
        {isExpanded && <span className="font-medium">Components</span>}
        <Button 
          variant="ghost" 
          size="sm" 
          onClick={() => setIsExpanded(!isExpanded)}
          className="ml-auto"
        >
          {isExpanded ? <ChevronLeft size={16} /> : <ChevronRight size={16} />}
        </Button>
      </div>

      {isExpanded && (
        <div className="overflow-y-auto p-2 flex-grow">
          {/* Workflows Section */}
          <div className="mb-4">
            <div 
              className="flex items-center justify-between cursor-pointer mb-2"
              onClick={() => toggleSection('workflows')}
            >
              <h3 className="font-medium text-sm text-purple-600 dark:text-purple-400">Workflows</h3>
              <div className="flex items-center">
                <Button 
                  variant="ghost" 
                  size="sm" 
                  onClick={(e) => {
                    e.stopPropagation();
                    handleAddWorkflow();
                  }}
                  className="h-6 w-6 p-0 mr-1"
                  title="Add new workflow"
                >
                  <Plus size={14} />
                </Button>
                <Button 
                  variant="ghost" 
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleSaveWorkflow();
                  }}
                  className="h-6 w-6 p-0 mr-1"
                  title="Save current workflow"
                >
                  <Save size={14} />
                </Button>
                {sectionsExpanded.workflows ? (
                  <ChevronUp size={16} className="text-muted-foreground" />
                ) : (
                  <ChevronDown size={16} className="text-muted-foreground" />
                )}
              </div>
            </div>
            
            {sectionsExpanded.workflows && (
              <div className="space-y-1.5">
                {workflows.map((workflow) => (
                  <div
                    key={workflow.id}
                    onClick={() => setActiveWorkflow(workflow.id)}
                    onDoubleClick={() => handleWorkflowDoubleClick(workflow.id)}
                    className={`p-2 rounded text-sm flex items-center justify-between group cursor-pointer ${
                      activeWorkflow === workflow.id 
                        ? 'bg-purple-100 dark:bg-purple-900/30 border-l-4 border-l-purple-400 dark:border-l-purple-600' 
                        : 'hover:bg-purple-50 dark:hover:bg-purple-900/10'
                    }`}
                    title="Double-click to load this workflow"
                  >
                    <div className="overflow-hidden">
                      <div className="font-medium truncate">{workflow.name}</div>
                      <div className="text-xs text-muted-foreground">{workflow.date}</div>
                    </div>
                    
                    <div className="flex opacity-0 group-hover:opacity-100 transition-opacity">
                      <Button 
                        variant="ghost" 
                        size="sm" 
                        onClick={(e) => {
                          e.stopPropagation();
                          handleRenameWorkflow(workflow.id);
                        }}
                        className="h-6 w-6 p-0"
                        title="Rename workflow"
                      >
                        <Pencil size={12} />
                      </Button>
                      <Button 
                        variant="ghost" 
                        size="sm" 
                        onClick={(e) => {
                          e.stopPropagation();
                          handleCloneWorkflow(workflow.id);
                        }}
                        className="h-6 w-6 p-0"
                        title="Clone workflow"
                      >
                        <Copy size={12} />
                      </Button>
                      <Button 
                        variant="ghost" 
                        size="sm" 
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDeleteWorkflow(workflow.id);
                        }}
                        className="h-6 w-6 p-0"
                        title="Delete workflow"
                      >
                        <Trash2 size={12} />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Agents Section */}
          <div className="mb-4">
            <div 
              className="flex items-center justify-between cursor-pointer mb-2"
              onClick={() => toggleSection('agents')}
            >
              <h3 className="font-medium text-sm text-blue-600 dark:text-blue-400">Agents</h3>
              {sectionsExpanded.agents ? (
                <ChevronUp size={16} className="text-muted-foreground" />
              ) : (
                <ChevronDown size={16} className="text-muted-foreground" />
              )}
            </div>
            
            {sectionsExpanded.agents && (
              <div className="space-y-1.5">
                {agentComponents.map((component) => (
                  <div
                    key={component.id}
                    draggable
                    onDragStart={(event) => onDragStart(event, component.type, component.label)}
                    className={`p-2.5 rounded shadow-sm cursor-move hover:shadow-md transition-all text-sm ${getComponentStyle(component.type)}`}
                  >
                    {component.label}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Tools Section */}
          <div className="mb-4">
            <div 
              className="flex items-center justify-between cursor-pointer mb-2"
              onClick={() => toggleSection('tools')}
            >
              <h3 className="font-medium text-sm text-amber-600 dark:text-amber-400">Tools</h3>
              {sectionsExpanded.tools ? (
                <ChevronUp size={16} className="text-muted-foreground" />
              ) : (
                <ChevronDown size={16} className="text-muted-foreground" />
              )}
            </div>
            
            {sectionsExpanded.tools && (
              <div className="space-y-1.5">
                {toolComponents.map((component) => (
                  <div
                    key={component.id}
                    draggable
                    onDragStart={(event) => onDragStart(event, component.type, component.label)}
                    className={`p-2.5 rounded shadow-sm cursor-move hover:shadow-md transition-all text-sm ${getComponentStyle(component.type)}`}
                  >
                    {component.label}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Functions Section */}
          <div>
            <div 
              className="flex items-center justify-between cursor-pointer mb-2"
              onClick={() => toggleSection('functions')}
            >
              <h3 className="font-medium text-sm text-emerald-600 dark:text-emerald-400">Functions</h3>
              {sectionsExpanded.functions ? (
                <ChevronUp size={16} className="text-muted-foreground" />
              ) : (
                <ChevronDown size={16} className="text-muted-foreground" />
              )}
            </div>
            
            {sectionsExpanded.functions && (
              <div className="space-y-1.5">
                {functionComponents.map((component) => (
                  <div
                    key={component.id}
                    draggable
                    onDragStart={(event) => onDragStart(event, component.type, component.label)}
                    className={`p-2.5 rounded shadow-sm cursor-move hover:shadow-md transition-all text-sm ${getComponentStyle(component.type)}`}
                    title={component.description}
                  >
                    {component.label}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
} 