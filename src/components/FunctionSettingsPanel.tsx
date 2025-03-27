import { useState } from 'react';
import { X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { FunctionItem } from '@/services/api';

interface FunctionSettingsProps {
  selectedFunction: FunctionItem | null;
  onClose: () => void;
}

// Function metadata including input and output formats
const functionMetadata: Record<string, { 
  inputs: { name: string; type: string; required: boolean; description: string }[];
  outputs: { name: string; type: string; description: string }[];
  example?: string;
}> = {
  'analysis-trends': {
    inputs: [
      { name: 'focus_areas', type: 'string[]', required: true, description: 'Areas of interest to analyze trends for' },
      { name: 'attribute_values', type: 'object', required: false, description: 'Extracted attributes to analyze' }
    ],
    outputs: [
      { name: 'trends', type: 'array', description: 'List of identified trends with focus area, trend, supporting data, and confidence' },
      { name: 'overall_insights', type: 'string[]', description: 'General insights derived from the data' },
      { name: 'data_quality', type: 'object', description: 'Assessment of data quality and limitations' }
    ],
    example: `{
  "focus_areas": ["customer satisfaction", "response time", "issue resolution"],
  "attribute_values": {
    "sentiments": [{"id": 1, "value": "positive"}, ...],
    "response_times": [{"id": 1, "value": 120}, ...]
  }
}`
  },
  'analysis-patterns': {
    inputs: [
      { name: 'pattern_types', type: 'string[]', required: true, description: 'Types of patterns to identify' },
      { name: 'attribute_values', type: 'object', required: false, description: 'Extracted attributes to analyze' }
    ],
    outputs: [
      { name: 'patterns', type: 'array', description: 'List of identified patterns with type, description, occurrences, examples, and significance' },
      { name: 'unexpected_patterns', type: 'array', description: 'Unexpected patterns and potential causes' }
    ],
    example: `{
  "pattern_types": ["user behavior", "error frequency", "intent_groups"],
  "attribute_values": {
    "intents": ["check status", "report issue", "cancel order", ...]
  }
}`
  },
  'analysis-findings': {
    inputs: [
      { name: 'questions', type: 'string[]', required: true, description: 'Questions to answer based on the analysis' },
      { name: 'attribute_values', type: 'object', required: true, description: 'Extracted attributes to analyze' }
    ],
    outputs: [
      { name: 'answers', type: 'array', description: 'Answers to each question with metrics, confidence level, and supporting data' },
      { name: 'data_gaps', type: 'string[]', description: 'Areas where data is insufficient' }
    ],
    example: `{
  "questions": ["What are the most common customer complaints?", "How has response time changed over time?"],
  "attribute_values": {
    "complaints": ["slow response", "product quality", ...],
    "response_times": [{"date": "2023-01", "time": 45}, ...]
  }
}`
  },
  'analysis-attributes': {
    inputs: [
      { name: 'text', type: 'string', required: true, description: 'Text to extract attributes from' },
      { name: 'attributes', type: 'array', required: true, description: 'Attribute definitions to extract' },
      { name: 'generate_required', type: 'boolean', required: false, description: 'Whether to generate required attributes' },
      { name: 'questions', type: 'string[]', required: false, description: 'Questions to generate attributes for (required if generate_required is true)' }
    ],
    outputs: [
      { name: 'attribute_values', type: 'object', description: 'Extracted attribute values' },
      { name: 'attributes', type: 'array', description: 'Generated attribute definitions (when generate_required is true)' }
    ],
    example: `{
  "text": "Customer called about their order #12345 which was delayed by 2 days...",
  "attributes": [
    {"field_name": "order_id", "title": "Order ID", "description": "The order number mentioned"}
  ]
}`
  },
  'analysis-intent': {
    inputs: [
      { name: 'text', type: 'string', required: true, description: 'Text to extract intent from' }
    ],
    outputs: [
      { name: 'label_name', type: 'string', description: 'Machine-readable intent label' },
      { name: 'label', type: 'string', description: 'Human-readable intent label' },
      { name: 'description', type: 'string', description: 'Detailed description of the intent' }
    ],
    example: `{
  "text": "I'd like to check on the status of my order #AB123"
}`
  },
  'analysis-results': {
    inputs: [
      { name: 'workflow_id', type: 'string', required: true, description: 'ID of the workflow to get results for' }
    ],
    outputs: [
      { name: 'results', type: 'array', description: 'Array of analysis results for the workflow' }
    ],
    example: `{
  "workflow_id": "workflow-12345"
}`
  }
};

export default function FunctionSettingsPanel({ selectedFunction, onClose }: FunctionSettingsProps) {
  const [activeTab, setActiveTab] = useState<string>('inputs');

  if (!selectedFunction) return null;

  const metadata = functionMetadata[selectedFunction.id] || {
    inputs: [],
    outputs: [],
    example: "No example available"
  };

  return (
    <div className="settings-panel bg-background border-l border-border h-full w-96 overflow-y-auto">
      <Card className="border-0 rounded-none h-full shadow-none">
        <CardHeader className="flex flex-row items-center justify-between p-4 pb-2 border-b">
          <CardTitle className="text-lg font-medium text-emerald-600">{selectedFunction.label}</CardTitle>
          <Button variant="ghost" size="icon" onClick={onClose} className="h-8 w-8">
            <X className="h-4 w-4" />
          </Button>
        </CardHeader>
        <div className="px-4 py-3 border-b">
          <p className="text-sm text-foreground/80">{selectedFunction.description}</p>
          <p className="text-xs text-muted-foreground mt-2">
            <span className="font-medium">Endpoint:</span> <code className="bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded text-emerald-700 dark:text-emerald-400">{selectedFunction.endpoint}</code>
          </p>
        </div>
        <CardContent className="p-0">
          <Tabs defaultValue="inputs" className="w-full" value={activeTab} onValueChange={setActiveTab}>
            <div className="px-4 pt-3 pb-2">
              <TabsList className="grid grid-cols-3 mb-2 w-full">
                <TabsTrigger value="inputs">Inputs</TabsTrigger>
                <TabsTrigger value="outputs">Outputs</TabsTrigger>
                <TabsTrigger value="example">Example</TabsTrigger>
              </TabsList>
            </div>

            <TabsContent value="inputs" className="px-4 mt-0 pb-4">
              <h3 className="text-sm font-medium mb-3 flex items-center">
                <span className="inline-block w-2 h-2 rounded-full bg-emerald-500 mr-2"></span>
                Required Parameters
              </h3>
              {metadata.inputs.filter(input => input.required).map(input => (
                <div key={input.name} className="mb-4 pl-4 border-l-2 border-emerald-200 dark:border-emerald-900">
                  <div className="flex items-center mb-1">
                    <div className="font-medium text-sm">{input.name}</div>
                    <div className="ml-2 text-xs px-1.5 py-0.5 bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400 rounded">
                      {input.type}
                    </div>
                  </div>
                  <p className="text-sm text-muted-foreground">{input.description}</p>
                </div>
              ))}

              {metadata.inputs.some(input => !input.required) && (
                <>
                  <h3 className="text-sm font-medium mb-3 mt-5 flex items-center">
                    <span className="inline-block w-2 h-2 rounded-full bg-slate-400 mr-2"></span>
                    Optional Parameters
                  </h3>
                  {metadata.inputs.filter(input => !input.required).map(input => (
                    <div key={input.name} className="mb-4 pl-4 border-l-2 border-slate-200 dark:border-slate-700">
                      <div className="flex items-center mb-1">
                        <div className="font-medium text-sm">{input.name}</div>
                        <div className="ml-2 text-xs px-1.5 py-0.5 bg-slate-100 text-slate-800 dark:bg-slate-800 dark:text-slate-300 rounded">
                          {input.type}
                        </div>
                      </div>
                      <p className="text-sm text-muted-foreground">{input.description}</p>
                    </div>
                  ))}
                </>
              )}
            </TabsContent>

            <TabsContent value="outputs" className="px-4 mt-0 pb-4">
              <h3 className="text-sm font-medium mb-3 flex items-center">
                <span className="inline-block w-2 h-2 rounded-full bg-blue-500 mr-2"></span>
                Response Structure
              </h3>
              {metadata.outputs.map(output => (
                <div key={output.name} className="mb-4 pl-4 border-l-2 border-blue-200 dark:border-blue-900">
                  <div className="flex items-center mb-1">
                    <div className="font-medium text-sm">{output.name}</div>
                    <div className="ml-2 text-xs px-1.5 py-0.5 bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-400 rounded">
                      {output.type}
                    </div>
                  </div>
                  <p className="text-sm text-muted-foreground">{output.description}</p>
                </div>
              ))}
            </TabsContent>

            <TabsContent value="example" className="px-4 mt-0">
              <h3 className="text-sm font-medium mb-2">Example Request</h3>
              <pre className="bg-slate-100 dark:bg-slate-800 p-4 rounded text-sm overflow-x-auto border border-slate-200 dark:border-slate-700 font-mono">
                {metadata.example}
              </pre>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
} 