import { useState } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

// Define the component types
interface ComponentItem {
  id: string;
  type: string;
  label: string;
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

export default function ComponentsPanel() {
  const [isExpanded, setIsExpanded] = useState(true);

  const onDragStart = (event: React.DragEvent<HTMLDivElement>, nodeType: string, nodeLabel: string) => {
    const data = JSON.stringify({ type: nodeType, label: nodeLabel });
    event.dataTransfer.setData('application/reactflow', data);
    event.dataTransfer.effectAllowed = 'move';
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
          <div className="mb-4">
            <h3 className="font-medium mb-2 text-sm text-blue-600 dark:text-blue-400">Agents</h3>
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
          </div>

          <div>
            <h3 className="font-medium mb-2 text-sm text-amber-600 dark:text-amber-400">Tools</h3>
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
          </div>
        </div>
      )}
    </div>
  );
} 