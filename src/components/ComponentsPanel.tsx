import { useState } from 'react';
import { ChevronLeft, ChevronRight, ChevronDown, ChevronUp, Plus, Pencil, Copy, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';

// Define the component types
interface ComponentItem {
  id: string;
  type: string;
  label: string;
}

interface Workflow {
  id: string;
  name: string;
  date: string;
}

// Agent components sorted alphabetically
const agentComponents: ComponentItem[] = [
  { id: 'agent-assist', type: 'agent', label: 'Agent Assist' },
  { id: 'analyzer', type: 'agent', label: 'Analyzer' },
  { id: 'coach', type: 'agent', label: 'Coach' },
  { id: 'evaluator', type: 'agent', label: 'Evaluator' },
  { id: 'knowledge', type: 'agent', label: 'Knowledge' },
  { id: 'observer', type: 'agent', label: 'Observer' },
  { id: 'orchestrator', type: 'agent', label: 'Orchestrator' },
  { id: 'planner', type: 'agent', label: 'Planner' },
  { id: 'researcher', type: 'agent', label: 'Researcher' },
  { id: 'superviser-assist', type: 'agent', label: 'Superviser Assist' },
  { id: 'trainer', type: 'agent', label: 'Trainer' },
];

// Tool components sorted alphabetically
const toolComponents: ComponentItem[] = [
  { id: 'agent-guide', type: 'tool', label: 'Agent Guide' },
  { id: 'agent-scorecard', type: 'tool', label: 'Agent Scorecard' },
  { id: 'agent-training', type: 'tool', label: 'Agent Training' },
  { id: 'call-observation', type: 'tool', label: 'Call Observation' },
  { id: 'competitive-analysis', type: 'tool', label: 'Competitive Analysis' },
  { id: 'goal-tracking', type: 'tool', label: 'Goal Tracking' },
  { id: 'talent-builder', type: 'tool', label: 'Talent Builder' },
];

// Sample workflows
const initialWorkflows: Workflow[] = [
  { id: 'workflow-1', name: 'Default Workflow', date: new Date().toLocaleDateString() }
];

// Component card styling based on type
const getComponentStyle = (type: string) => {
  switch (type) {
    case 'agent':
      return 'border-l-4 border-l-blue-400 bg-blue-50 dark:bg-blue-950/30 dark:border-l-blue-600';
    case 'tool':
      return 'border-l-4 border-l-amber-400 bg-amber-50 dark:bg-amber-950/30 dark:border-l-amber-600';
    default:
      return '';
  }
};

interface ComponentsPanelProps {
  onSaveWorkflow?: () => void;
}

export default function ComponentsPanel({ onSaveWorkflow }: ComponentsPanelProps) {
  const [isExpanded, setIsExpanded] = useState(true);
  const [workflows, setWorkflows] = useState<Workflow[]>(initialWorkflows);
  const [activeWorkflow, setActiveWorkflow] = useState<string>(initialWorkflows[0].id);
  
  // Expandable sections state
  const [sectionsExpanded, setSectionsExpanded] = useState({
    workflows: true,
    agents: true,
    tools: true
  });

  const toggleSection = (section: 'workflows' | 'agents' | 'tools') => {
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

  const handleAddWorkflow = () => {
    const newWorkflow: Workflow = {
      id: `workflow-${Date.now()}`,
      name: `New Workflow ${workflows.length + 1}`,
      date: new Date().toLocaleDateString()
    };
    setWorkflows([...workflows, newWorkflow]);
    setActiveWorkflow(newWorkflow.id);
  };

  const handleSaveWorkflow = () => {
    // This would typically save the current flow state
    if (onSaveWorkflow) {
      onSaveWorkflow();
    }
    // For demo purposes, we'll just update the date
    setWorkflows(workflows.map(wf => 
      wf.id === activeWorkflow 
        ? { ...wf, date: new Date().toLocaleDateString() } 
        : wf
    ));
  };

  const handleCloneWorkflow = (id: string) => {
    const workflowToClone = workflows.find(wf => wf.id === id);
    if (workflowToClone) {
      const newWorkflow: Workflow = {
        id: `workflow-${Date.now()}`,
        name: `${workflowToClone.name} (Copy)`,
        date: new Date().toLocaleDateString()
      };
      setWorkflows([...workflows, newWorkflow]);
    }
  };

  const handleRenameWorkflow = (id: string) => {
    const newName = prompt('Enter new workflow name:');
    if (newName && newName.trim() !== '') {
      setWorkflows(workflows.map(wf => 
        wf.id === id ? { ...wf, name: newName.trim() } : wf
      ));
    }
  };

  const handleDeleteWorkflow = (id: string) => {
    if (workflows.length <= 1) {
      alert('Cannot delete the only workflow.');
      return;
    }
    
    const confirmed = confirm('Are you sure you want to delete this workflow?');
    if (confirmed) {
      const newWorkflows = workflows.filter(wf => wf.id !== id);
      setWorkflows(newWorkflows);
      
      // If the active workflow was deleted, select the first one
      if (activeWorkflow === id) {
        setActiveWorkflow(newWorkflows[0].id);
      }
    }
  };

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
                >
                  <Plus size={0} /> {/* Placeholder to maintain spacing */}
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
                    className={`p-2 rounded text-sm flex items-center justify-between group cursor-pointer ${
                      activeWorkflow === workflow.id 
                        ? 'bg-purple-100 dark:bg-purple-900/30 border-l-4 border-l-purple-400 dark:border-l-purple-600' 
                        : 'hover:bg-purple-50 dark:hover:bg-purple-900/10'
                    }`}
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
          <div>
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
        </div>
      )}
    </div>
  );
} 