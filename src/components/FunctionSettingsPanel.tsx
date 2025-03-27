import React, { useState } from 'react';
import { X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { api, FunctionItem, ComponentItem } from '@/services/api';

interface FunctionSettingsProps {
  selectedFunction: FunctionItem;
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

// Handle executing functions through the unified API
const handleExecuteFunction = async (
  func: FunctionItem,
  options: {
    parameters?: Record<string, any>;
    inputData?: any;
    config?: any;
    workflowId?: string;
  }
) => {
  const { parameters, inputData, config, workflowId } = options;
  
  try {
    // Check if this is a chain analysis function
    if (func.id === 'analysis-chain') {
      // For chain analysis, use the specialized endpoint
      return await api.performChainAnalysis(
        workflowId || 'temp-workflow',
        inputData,
        config
      );
    } else {
      // For regular analysis functions, use the unified endpoint with the analysis type
      const analysisType = func.analysisType || func.id.replace('analysis-', '');
      
      // Determine text vs data parameter based on input type
      const isTextInput = typeof inputData === 'string';
      
      return await api.performAnalysis(
        analysisType,
        parameters || {},
        isTextInput ? undefined : inputData,
        isTextInput ? inputData : undefined,
        workflowId
      );
    }
  } catch (error: unknown) {
    console.error('Error executing function:', error);
    throw error;
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

  // Render different configuration UI based on function ID
  const renderConfigurationUI = () => {
    switch (selectedFunction.id) {
      case 'analysis-chain':
        return <ChainAnalysisConfig function={selectedFunction} />;
      case 'analysis-trends':
        return <TrendsAnalysisConfig function={selectedFunction} />;
      case 'analysis-patterns':
        return <PatternsAnalysisConfig function={selectedFunction} />;
      case 'analysis-findings':
        return <FindingsAnalysisConfig function={selectedFunction} />;
      default:
        return <DefaultFunctionConfig function={selectedFunction} />;
    }
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
          <div className="flex flex-wrap gap-2 mt-2">
            <span className="text-xs px-2 py-1 bg-emerald-100 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-400 rounded-full">
              {selectedFunction.id.replace('analysis-', '').replace(/-/g, ' ')}
            </span>
            {selectedFunction.type && (
              <span className="text-xs px-2 py-1 bg-blue-100 text-blue-700 dark:bg-blue-900/20 dark:text-blue-400 rounded-full">
                {selectedFunction.type}
              </span>
            )}
            {(selectedFunction as any).analysisType && (
              <span className="text-xs px-2 py-1 bg-purple-100 text-purple-700 dark:bg-purple-900/20 dark:text-purple-400 rounded-full">
                {(selectedFunction as any).analysisType}
              </span>
            )}
          </div>
          <p className="text-xs text-muted-foreground mt-2">
            <span className="font-medium">Endpoint:</span> <code className="bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded text-emerald-700 dark:text-emerald-400">{selectedFunction.endpoint}</code>
            {selectedFunction.analysisType && (
              <>
                <br />
                <span className="font-medium">Analysis Type:</span> <code className="bg-slate-100 dark:bg-slate-800 px-1.5 py-0.5 rounded text-purple-700 dark:text-purple-400">{selectedFunction.analysisType}</code>
              </>
            )}
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

// Default configuration UI for functions
function DefaultFunctionConfig({ function: func }: { function: FunctionItem }) {
  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">
        Configure the parameters for this function.
      </p>
      
      {/* Render parameters if they exist */}
      {(func as any).parameters && Array.isArray((func as any).parameters) && 
       (func as any).parameters.map((param: any, index: number) => (
        <div key={index} className="space-y-2">
          <Label htmlFor={`param-${param.name}`}>{param.name}</Label>
          <Input 
            id={`param-${param.name}`} 
            placeholder={param.description} 
          />
          <p className="text-xs text-muted-foreground">{param.description}</p>
        </div>
      ))}
    </div>
  );
}

// ChainAnalysisConfig component for the Chain Analysis function
function ChainAnalysisConfig({ function: func }: { function: FunctionItem }) {
  const [useAttributes, setUseAttributes] = useState(true);
  const [inputType, setInputType] = useState<'text' | 'json'>('text');
  const [inputText, setInputText] = useState('');
  const [focusAreas, setFocusAreas] = useState('customer_satisfaction, resolution_time, agent_effectiveness');
  const [patternTypes, setPatternTypes] = useState('conversation_flow, resolution_patterns, customer_behavior');
  const [questions, setQuestions] = useState('What are common issues?\nHow effective are our agents?\nWhat can be improved?');
  const [isExecuting, setIsExecuting] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);
  
  // Handle execution of the chain analysis
  const executeChainAnalysis = async () => {
    try {
      setIsExecuting(true);
      setError(null);
      
      // Parse input data based on type
      let inputData: any;
      if (inputType === 'text') {
        inputData = inputText;
      } else {
        try {
          inputData = JSON.parse(inputText);
        } catch (e: unknown) {
          const errorMessage = e instanceof Error ? e.message : 'Invalid JSON input data';
          setError(errorMessage);
          setIsExecuting(false);
          return;
        }
      }
      
      // Parse string lists into arrays
      const focusAreasArray = focusAreas.split(',').map(item => item.trim()).filter(Boolean);
      const patternTypesArray = patternTypes.split(',').map(item => item.trim()).filter(Boolean);
      const questionsArray = questions.split('\n').map(item => item.trim()).filter(Boolean);
      
      // Create config object
      const config = {
        use_attributes: useAttributes,
        focus_areas: focusAreasArray,
        pattern_types: patternTypesArray,
        questions: questionsArray
      };
      
      // Execute the function
      try {
        const result = await handleExecuteFunction(func, {
          inputData,
          config,
          workflowId: 'current-workflow' // This could be passed from props in a real implementation
        });
        
        setResult(result);
      } catch (e: unknown) {
        const errorMessage = e instanceof Error ? e.message : 'Error executing chain analysis';
        setError(`Error executing chain analysis: ${errorMessage}`);
      }
    } finally {
      setIsExecuting(false);
    }
  };
  
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h3 className="font-medium">Chain Configuration</h3>
        <p className="text-sm text-muted-foreground">
          This function chains multiple analysis steps together into a workflow.
        </p>
      </div>
      
      <div className="space-y-4">
        <div className="flex items-center space-x-2">
          <Switch 
            id="use-attributes" 
            checked={useAttributes}
            onCheckedChange={setUseAttributes}
          />
          <Label htmlFor="use-attributes">Extract attributes from text first</Label>
        </div>
        
        <div className="space-y-2">
          <Label>Input Type</Label>
          <div className="flex space-x-2">
            <Button 
              variant={inputType === 'text' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setInputType('text')}
              className="flex-1"
            >
              Text Input
            </Button>
            <Button 
              variant={inputType === 'json' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setInputType('json')}
              className="flex-1"
            >
              JSON Input
            </Button>
          </div>
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="input-data">Input Data</Label>
          <Textarea 
            id="input-data" 
            placeholder={inputType === 'text' 
              ? 'Enter conversation text here...' 
              : '{ "key": "value" }'}
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            className="min-h-[100px]"
          />
          <p className="text-xs text-muted-foreground">
            {inputType === 'text' 
              ? 'Conversation text to analyze' 
              : 'JSON data must be properly formatted'}
          </p>
        </div>
      </div>
      
      <div className="pt-4 border-t space-y-4">
        <h3 className="font-medium">Analysis Steps</h3>
        
        <div className="space-y-2">
          <Label htmlFor="focus-areas">Focus Areas for Trend Analysis</Label>
          <Input 
            id="focus-areas" 
            placeholder="customer_satisfaction, resolution_time" 
            value={focusAreas}
            onChange={(e) => setFocusAreas(e.target.value)}
          />
          <p className="text-xs text-muted-foreground">
            Comma-separated list of areas to focus on for trend analysis
          </p>
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="pattern-types">Pattern Types to Identify</Label>
          <Input 
            id="pattern-types" 
            placeholder="conversation_flow, resolution_patterns" 
            value={patternTypes}
            onChange={(e) => setPatternTypes(e.target.value)}
          />
          <p className="text-xs text-muted-foreground">
            Comma-separated list of pattern types to look for
          </p>
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="questions">Questions for Findings Analysis</Label>
          <Textarea 
            id="questions" 
            placeholder="Enter questions on separate lines..." 
            value={questions}
            onChange={(e) => setQuestions(e.target.value)}
            className="min-h-[100px]"
          />
          <p className="text-xs text-muted-foreground">
            Enter each question on a new line
          </p>
        </div>
      </div>
      
      <div className="pt-4 border-t">
        <Button 
          className="w-full" 
          onClick={executeChainAnalysis}
          disabled={isExecuting}
        >
          {isExecuting ? 'Executing...' : 'Execute Chain Analysis'}
        </Button>
        
        {error && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-600 text-sm">
            {error}
          </div>
        )}
        
        {result && !error && (
          <div className="mt-4 space-y-2">
            <h3 className="font-medium text-sm">Result Summary:</h3>
            <div className="p-3 bg-slate-50 border border-slate-200 rounded-md text-sm max-h-[300px] overflow-y-auto">
              <pre className="whitespace-pre-wrap break-words">
                {JSON.stringify(result, null, 2)}
              </pre>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// TrendsAnalysisConfig component for the Trends Analysis function
function TrendsAnalysisConfig({ function: func }: { function: FunctionItem }) {
  const [inputType, setInputType] = useState<'text' | 'json'>('text');
  const [inputText, setInputText] = useState('');
  const [focusAreas, setFocusAreas] = useState('customer_satisfaction, resolution_time, agent_effectiveness');
  const [isExecuting, setIsExecuting] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);
  
  // Handle execution of trends analysis
  const executeTrendsAnalysis = async () => {
    try {
      setIsExecuting(true);
      setError(null);
      
      // Parse focus areas
      const focusAreasArray = focusAreas.split(',').map(item => item.trim()).filter(Boolean);
      
      // Prepare parameters
      const parameters = {
        focus_areas: focusAreasArray
      };
      
      // Execute the function
      try {
        const result = await handleExecuteFunction(func, {
          parameters,
          inputData: inputType === 'text' ? inputText : JSON.parse(inputText)
        });
        
        setResult(result);
      } catch (e: unknown) {
        const errorMessage = e instanceof Error ? e.message : 'Error executing trends analysis';
        setError(`Error executing trends analysis: ${errorMessage}`);
      }
    } catch (parseError) {
      setError('Invalid JSON input data');
    } finally {
      setIsExecuting(false);
    }
  };
  
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h3 className="font-medium">Trends Analysis Configuration</h3>
        <p className="text-sm text-muted-foreground">
          Identify trends and patterns in conversation data.
        </p>
      </div>
      
      <div className="space-y-4">
        <div className="space-y-2">
          <Label>Input Type</Label>
          <div className="flex space-x-2">
            <Button 
              variant={inputType === 'text' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setInputType('text')}
              className="flex-1"
            >
              Text Input
            </Button>
            <Button 
              variant={inputType === 'json' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setInputType('json')}
              className="flex-1"
            >
              JSON Input
            </Button>
          </div>
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="input-data-trends">Input Data</Label>
          <Textarea 
            id="input-data-trends" 
            placeholder={inputType === 'text' 
              ? 'Enter conversation text here...' 
              : '{ "key": "value" }'}
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            className="min-h-[100px]"
          />
          <p className="text-xs text-muted-foreground">
            {inputType === 'text' 
              ? 'Conversation text to analyze' 
              : 'JSON data must be properly formatted'}
          </p>
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="focus-areas-trends">Focus Areas</Label>
          <Input 
            id="focus-areas-trends" 
            placeholder="customer_satisfaction, resolution_time" 
            value={focusAreas}
            onChange={(e) => setFocusAreas(e.target.value)}
          />
          <p className="text-xs text-muted-foreground">
            Comma-separated list of areas to focus on
          </p>
        </div>
      </div>
      
      <div className="pt-4 border-t">
        <Button 
          className="w-full" 
          onClick={executeTrendsAnalysis}
          disabled={isExecuting}
        >
          {isExecuting ? 'Executing...' : 'Execute Trends Analysis'}
        </Button>
        
        {error && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-600 text-sm">
            {error}
          </div>
        )}
        
        {result && !error && (
          <div className="mt-4 space-y-2">
            <h3 className="font-medium text-sm">Result Summary:</h3>
            <div className="p-3 bg-slate-50 border border-slate-200 rounded-md text-sm max-h-[300px] overflow-y-auto">
              <pre className="whitespace-pre-wrap break-words">
                {JSON.stringify(result, null, 2)}
              </pre>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// PatternsAnalysisConfig component for the Patterns Analysis function
function PatternsAnalysisConfig({ function: func }: { function: FunctionItem }) {
  const [inputType, setInputType] = useState<'text' | 'json'>('text');
  const [inputText, setInputText] = useState('');
  const [patternTypes, setPatternTypes] = useState('conversation_flow, resolution_patterns, customer_behavior');
  const [isExecuting, setIsExecuting] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);
  
  // Handle execution of patterns analysis
  const executePatternsAnalysis = async () => {
    try {
      setIsExecuting(true);
      setError(null);
      
      // Parse pattern types
      const patternTypesArray = patternTypes.split(',').map(item => item.trim()).filter(Boolean);
      
      // Prepare parameters
      const parameters = {
        pattern_types: patternTypesArray
      };
      
      // Execute the function
      try {
        const result = await handleExecuteFunction(func, {
          parameters,
          inputData: inputType === 'text' ? inputText : JSON.parse(inputText)
        });
        
        setResult(result);
      } catch (e: unknown) {
        const errorMessage = e instanceof Error ? e.message : 'Error executing patterns analysis';
        setError(`Error executing patterns analysis: ${errorMessage}`);
      }
    } catch (parseError) {
      setError('Invalid JSON input data');
    } finally {
      setIsExecuting(false);
    }
  };
  
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h3 className="font-medium">Patterns Analysis Configuration</h3>
        <p className="text-sm text-muted-foreground">
          Identify patterns in conversation data.
        </p>
      </div>
      
      <div className="space-y-4">
        <div className="space-y-2">
          <Label>Input Type</Label>
          <div className="flex space-x-2">
            <Button 
              variant={inputType === 'text' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setInputType('text')}
              className="flex-1"
            >
              Text Input
            </Button>
            <Button 
              variant={inputType === 'json' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setInputType('json')}
              className="flex-1"
            >
              JSON Input
            </Button>
          </div>
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="input-data-patterns">Input Data</Label>
          <Textarea 
            id="input-data-patterns" 
            placeholder={inputType === 'text' 
              ? 'Enter conversation text here...' 
              : '{ "key": "value" }'}
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            className="min-h-[100px]"
          />
          <p className="text-xs text-muted-foreground">
            {inputType === 'text' 
              ? 'Conversation text to analyze' 
              : 'JSON data must be properly formatted'}
          </p>
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="pattern-types">Pattern Types</Label>
          <Input 
            id="pattern-types" 
            placeholder="conversation_flow, resolution_patterns" 
            value={patternTypes}
            onChange={(e) => setPatternTypes(e.target.value)}
          />
          <p className="text-xs text-muted-foreground">
            Comma-separated list of pattern types to identify
          </p>
        </div>
      </div>
      
      <div className="pt-4 border-t">
        <Button 
          className="w-full" 
          onClick={executePatternsAnalysis}
          disabled={isExecuting}
        >
          {isExecuting ? 'Executing...' : 'Execute Patterns Analysis'}
        </Button>
        
        {error && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-600 text-sm">
            {error}
          </div>
        )}
        
        {result && !error && (
          <div className="mt-4 space-y-2">
            <h3 className="font-medium text-sm">Result Summary:</h3>
            <div className="p-3 bg-slate-50 border border-slate-200 rounded-md text-sm max-h-[300px] overflow-y-auto">
              <pre className="whitespace-pre-wrap break-words">
                {JSON.stringify(result, null, 2)}
              </pre>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// FindingsAnalysisConfig component for the Findings Analysis function
function FindingsAnalysisConfig({ function: func }: { function: FunctionItem }) {
  const [inputType, setInputType] = useState<'text' | 'json'>('text');
  const [inputText, setInputText] = useState('');
  const [questions, setQuestions] = useState('What are common issues?\nHow effective are our agents?\nWhat can be improved?');
  const [isExecuting, setIsExecuting] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);
  
  // Handle execution of findings analysis
  const executeFindingsAnalysis = async () => {
    try {
      setIsExecuting(true);
      setError(null);
      
      // Parse questions
      const questionsArray = questions.split('\n').map(item => item.trim()).filter(Boolean);
      
      // Prepare parameters
      const parameters = {
        questions: questionsArray
      };
      
      // Execute the function
      try {
        const result = await handleExecuteFunction(func, {
          parameters,
          inputData: inputType === 'text' ? inputText : JSON.parse(inputText)
        });
        
        setResult(result);
      } catch (e: unknown) {
        const errorMessage = e instanceof Error ? e.message : 'Error executing findings analysis';
        setError(`Error executing findings analysis: ${errorMessage}`);
      }
    } catch (parseError) {
      setError('Invalid JSON input data');
    } finally {
      setIsExecuting(false);
    }
  };
  
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h3 className="font-medium">Findings Analysis Configuration</h3>
        <p className="text-sm text-muted-foreground">
          Extract findings and insights from conversation data.
        </p>
      </div>
      
      <div className="space-y-4">
        <div className="space-y-2">
          <Label>Input Type</Label>
          <div className="flex space-x-2">
            <Button 
              variant={inputType === 'text' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setInputType('text')}
              className="flex-1"
            >
              Text Input
            </Button>
            <Button 
              variant={inputType === 'json' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setInputType('json')}
              className="flex-1"
            >
              JSON Input
            </Button>
          </div>
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="input-data-findings">Input Data</Label>
          <Textarea 
            id="input-data-findings" 
            placeholder={inputType === 'text' 
              ? 'Enter conversation text here...' 
              : '{ "key": "value" }'}
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            className="min-h-[100px]"
          />
          <p className="text-xs text-muted-foreground">
            {inputType === 'text' 
              ? 'Conversation text to analyze' 
              : 'JSON data must be properly formatted'}
          </p>
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="questions">Questions</Label>
          <Textarea 
            id="questions" 
            placeholder="Enter questions on separate lines..." 
            value={questions}
            onChange={(e) => setQuestions(e.target.value)}
            className="min-h-[100px]"
          />
          <p className="text-xs text-muted-foreground">
            Enter each question on a new line
          </p>
        </div>
      </div>
      
      <div className="pt-4 border-t">
        <Button 
          className="w-full" 
          onClick={executeFindingsAnalysis}
          disabled={isExecuting}
        >
          {isExecuting ? 'Executing...' : 'Execute Findings Analysis'}
        </Button>
        
        {error && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-600 text-sm">
            {error}
          </div>
        )}
        
        {result && !error && (
          <div className="mt-4 space-y-2">
            <h3 className="font-medium text-sm">Result Summary:</h3>
            <div className="p-3 bg-slate-50 border border-slate-200 rounded-md text-sm max-h-[300px] overflow-y-auto">
              <pre className="whitespace-pre-wrap break-words">
                {JSON.stringify(result, null, 2)}
              </pre>
            </div>
          </div>
        )}
      </div>
    </div>
  );
} 