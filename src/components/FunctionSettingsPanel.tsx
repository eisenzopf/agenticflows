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
  responseExample?: string;
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
}`,
    responseExample: `{
  "trends": [
    {
      "focus_area": "customer satisfaction",
      "trend": "Increasing positivity in customer feedback",
      "supporting_data": {
        "sentiment_scores": [0.2, 0.4, 0.6, 0.7],
        "time_periods": ["Q1", "Q2", "Q3", "Q4"]
      },
      "confidence": 0.85
    },
    {
      "focus_area": "response time",
      "trend": "Decreasing average resolution time",
      "supporting_data": {
        "average_times": [45, 42, 38, 32],
        "time_periods": ["Q1", "Q2", "Q3", "Q4"]
      },
      "confidence": 0.92
    }
  ],
  "overall_insights": [
    "Customer satisfaction is improving consistently",
    "Response times have decreased by 29% over the last year",
    "Issue resolution rates show positive correlation with satisfaction scores"
  ],
  "data_quality": {
    "completeness": 0.87,
    "limitations": ["Limited data for enterprise customers", "Seasonal variations not fully accounted for"]
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
}`,
    responseExample: `{
  "patterns": [
    {
      "type": "user behavior",
      "description": "Users frequently check order status multiple times within an hour",
      "occurrences": 487,
      "examples": ["User #1242 checked status 6 times", "User #2155 checked status 4 times"],
      "significance": 0.78
    },
    {
      "type": "error frequency",
      "description": "Payment processing errors peak on Monday mornings",
      "occurrences": 142,
      "examples": ["Error #4011 on Monday 9:15 AM", "Error #4011 on Monday 8:45 AM"],
      "significance": 0.85
    }
  ],
  "unexpected_patterns": [
    {
      "description": "High rate of cart abandonment after successful coupon application",
      "potential_causes": ["UI confusion", "Expected larger discount", "Changed mind about purchase"],
      "significance": 0.65
    }
  ]
}`
  },
  'analysis-recommendations': {
    inputs: [
      { name: 'focus_area', type: 'string', required: true, description: 'Area to focus recommendations on' },
      { name: 'criteria', type: 'object', required: false, description: 'Criteria for prioritizing recommendations' },
      { name: 'data', type: 'object', required: true, description: 'Analysis results to base recommendations on' }
    ],
    outputs: [
      { name: 'immediate_actions', type: 'array', description: 'List of recommended actions with rationale, impact, and priority' },
      { name: 'implementation_notes', type: 'string[]', description: 'Notes on implementing the recommendations' },
      { name: 'success_metrics', type: 'string[]', description: 'Metrics to measure success of implementations' }
    ],
    example: `{
  "focus_area": "customer retention",
  "criteria": {
    "impact": 0.6,
    "implementation_ease": 0.4
  },
  "data": {
    "trends": [
      {"focus_area": "customer satisfaction", "trend": "Declining satisfaction scores in Q3"},
      {"focus_area": "response time", "trend": "Increasing average resolution time"}
    ],
    "findings": [
      {"question": "What are the main pain points?", "answer": "Long wait times and complex processes"}
    ]
  }
}`,
    responseExample: `{
  "immediate_actions": [
    {
      "action": "Implement callback option for customers on hold for more than 2 minutes",
      "rationale": "Reduces customer frustration during peak call times",
      "expected_impact": "15% reduction in call abandonment rate",
      "priority": 5
    },
    {
      "action": "Simplify the refund process from 5 steps to 2 steps",
      "rationale": "Current process is overly complex and leads to customer frustration",
      "expected_impact": "30% reduction in repeat calls about refunds",
      "priority": 4
    },
    {
      "action": "Proactively notify customers about known service issues",
      "rationale": "Prevents unnecessary inbound contacts and shows proactive service",
      "expected_impact": "20% reduction in calls during service incidents",
      "priority": 3
    }
  ],
  "implementation_notes": [
    "Begin with highest priority items requiring minimal IT changes",
    "Schedule implementation during low-volume periods",
    "Ensure customer service agents receive training on new processes"
  ],
  "success_metrics": [
    "Customer satisfaction scores (target: 15% improvement in 90 days)",
    "First call resolution rate (target: increase from 65% to 80%)",
    "Average handle time (target: reduce by 45 seconds)"
  ]
}`
  },
  'analysis-plan': {
    inputs: [
      { name: 'generate_timeline', type: 'boolean', required: false, description: 'Whether to generate a timeline instead of a full plan' },
      { name: 'constraints', type: 'object', required: false, description: 'Constraints for the plan such as budget and timeline' },
      { name: 'data', type: 'object', required: true, description: 'Recommendations or action plan to base the plan on' }
    ],
    outputs: [
      { name: 'goals', type: 'string[]', description: 'Goals of the action plan' },
      { name: 'immediate_actions', type: 'array', description: 'Actions to take immediately' },
      { name: 'short_term_actions', type: 'array', description: 'Actions to take in the short term' },
      { name: 'long_term_actions', type: 'array', description: 'Actions to take in the long term' },
      { name: 'timeline', type: 'array', description: 'Timeline of implementation phases and milestones' },
      { name: 'risks_mitigations', type: 'array', description: 'Risks and mitigation strategies' }
    ],
    example: `{
  "constraints": {
    "budget": 50000,
    "timeline": "6 months",
    "resources": ["customer_support", "engineering", "marketing"]
  },
  "data": {
    "recommendations": {
      "immediate_actions": [
        {
          "action": "Implement callback option for customers on hold",
          "rationale": "Reduces customer frustration during peak call times",
          "expected_impact": "15% reduction in call abandonment rate",
          "priority": 5
        },
        {
          "action": "Simplify the refund process from 5 steps to 2 steps",
          "rationale": "Current process is overly complex and leads to customer frustration",
          "expected_impact": "30% reduction in repeat calls about refunds",
          "priority": 4
        }
      ],
      "implementation_notes": ["Begin with highest priority items", "Schedule during low-volume periods"],
      "success_metrics": ["Customer satisfaction scores", "First call resolution rate"]
    }
  }
}`,
    responseExample: `{
  "goals": [
    "Improve customer satisfaction by 20% within 6 months",
    "Reduce average handle time by 15% within 3 months",
    "Increase first-call resolution rate to 80% within 4 months"
  ],
  "immediate_actions": [
    {
      "action": "Implement callback option",
      "description": "Set up system to capture callbacks for customers on hold over 2 minutes",
      "priority": 5,
      "estimated_effort": "2 weeks",
      "dependencies": [],
      "responsible_role": "IT Manager"
    },
    {
      "action": "Train agents on new system",
      "description": "Develop and deliver training on the callback system",
      "priority": 5,
      "estimated_effort": "1 week",
      "dependencies": ["Implement callback option"],
      "responsible_role": "Training Manager"
    }
  ],
  "short_term_actions": [
    {
      "action": "Simplify refund process",
      "description": "Redesign refund workflow to reduce steps from 5 to 2",
      "priority": 4,
      "estimated_effort": "3 weeks",
      "dependencies": [],
      "responsible_role": "Process Manager"
    }
  ],
  "long_term_actions": [
    {
      "action": "Implement proactive notification system",
      "description": "Develop system to alert customers of service issues",
      "priority": 3,
      "estimated_effort": "8 weeks",
      "dependencies": ["Analyze customer contact patterns"],
      "responsible_role": "IT Director"
    }
  ],
  "responsible_parties": [
    "Customer Support Director",
    "IT Manager",
    "Training Manager",
    "Process Manager"
  ],
  "timeline": [
    {
      "phase": "Phase 1: Quick Wins",
      "description": "Implement high-impact, low-effort solutions",
      "duration": "4 weeks",
      "milestones": [
        "Callback system implemented",
        "Agent training completed",
        "Initial metrics dashboard created"
      ]
    },
    {
      "phase": "Phase 2: Process Improvements",
      "description": "Redesign core customer processes",
      "duration": "8 weeks",
      "milestones": [
        "Refund process simplified",
        "Knowledge base updated",
        "Process documentation completed"
      ]
    }
  ],
  "success_metrics": [
    "Customer satisfaction score (target: 85%)",
    "Average handle time (target: 8 minutes)",
    "First call resolution rate (target: 80%)",
    "Call abandonment rate (target: <5%)"
  ],
  "risks_mitigations": [
    {
      "risk": "System integration delays",
      "impact": "High",
      "probability": "Medium",
      "mitigation_plan": "Engage IT early, use phased approach with interim manual solutions",
      "contingency_plan": "Prepare manual workaround processes",
      "responsible_party": "IT Manager"
    },
    {
      "risk": "Agent resistance to new processes",
      "impact": "Medium",
      "probability": "Medium",
      "mitigation_plan": "Early agent involvement in design, comprehensive training",
      "contingency_plan": "Identify champions to support peers",
      "responsible_party": "Training Manager"
    }
  ]
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
}`,
    responseExample: `{
  "answers": [
    {
      "question": "What are the most common customer complaints?",
      "answer": "The most common customer complaints relate to product delivery delays and inconsistent quality control",
      "metrics": {
        "delivery_complaints": 42.7,
        "quality_complaints": 31.2,
        "support_complaints": 18.5,
        "pricing_complaints": 7.6
      },
      "confidence": 0.89,
      "supporting_data": {
        "sample_size": 2456,
        "time_period": "Last 6 months"
      }
    },
    {
      "question": "How has response time changed over time?",
      "answer": "Response times have steadily improved by approximately 15% quarterly over the past year",
      "metrics": {
        "q1_avg_minutes": 42,
        "q2_avg_minutes": 36,
        "q3_avg_minutes": 31,
        "q4_avg_minutes": 26
      },
      "confidence": 0.93,
      "supporting_data": {
        "total_tickets": 12845,
        "methodology": "Average time to first response"
      }
    }
  ],
  "data_gaps": [
    "Insufficient data for enterprise customer segment",
    "Limited historical data beyond 18 months",
    "Inconsistent categorization of complaints before Q2"
  ]
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
}`,
    responseExample: `{
  "attribute_values": {
    "order_id": "12345",
    "delay_duration": "2 days",
    "issue_type": "delivery delay",
    "customer_sentiment": "frustrated",
    "priority": "medium"
  },
  "attributes": [
    {
      "field_name": "order_id",
      "title": "Order ID",
      "description": "The order number mentioned",
      "value_type": "string",
      "confidence": 0.98
    },
    {
      "field_name": "delay_duration",
      "title": "Delay Duration",
      "description": "How long the order was delayed",
      "value_type": "duration",
      "confidence": 0.95
    },
    {
      "field_name": "issue_type",
      "title": "Issue Type",
      "description": "Category of the customer issue",
      "value_type": "category",
      "confidence": 0.87
    }
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
}`,
    responseExample: `{
  "label_name": "order_status_inquiry",
  "label": "Order Status Inquiry",
  "description": "Customer is asking about the current status of their existing order",
  "confidence": 0.94,
  "extracted_entities": {
    "order_id": "AB123"
  },
  "alternative_intents": [
    {
      "label_name": "order_tracking",
      "confidence": 0.68
    },
    {
      "label_name": "delivery_inquiry",
      "confidence": 0.42
    }
  ]
}`
  },
  'analysis-chain': {
    inputs: [
      { name: 'input_data', type: 'string | object', required: true, description: 'Text or structured data to analyze' },
      { name: 'config', type: 'object', required: true, description: 'Configuration for the chain analysis' },
      { name: 'workflow_id', type: 'string', required: false, description: 'ID of the workflow to save results to' }
    ],
    outputs: [
      { name: 'results', type: 'object', description: 'Combined results from all analysis steps' },
      { name: 'meta', type: 'object', description: 'Metadata about the analysis process' }
    ],
    example: `{
  "input_data": "Customer messaged about their recent purchase: 'I ordered the premium headphones two weeks ago and they still haven't arrived. The tracking hasn't updated in 5 days. This is very frustrating as I needed them for an upcoming trip.'",
  "config": {
    "use_attributes": true,
    "focus_areas": ["delivery experience", "customer satisfaction"],
    "pattern_types": ["delivery issues", "communication gaps"],
    "questions": ["What is the customer's main concern?", "How urgent is this issue?"]
  },
  "workflow_id": "workflow-1234"
}`,
    responseExample: `{
  "results": {
    "attributes": {
      "product": "premium headphones",
      "order_age": "two weeks",
      "tracking_status": "stalled",
      "tracking_last_updated": "5 days ago",
      "customer_sentiment": "frustrated",
      "urgency": "high",
      "purpose": "upcoming trip"
    },
    "trends": {
      "identified_trends": [
        {
          "focus_area": "delivery experience",
          "trend": "Delayed delivery without status updates",
          "confidence": 0.89
        },
        {
          "focus_area": "customer satisfaction",
          "trend": "Increasing frustration over time",
          "confidence": 0.92
        }
      ]
    },
    "patterns": {
      "identified_patterns": [
        {
          "type": "delivery issues",
          "description": "Premium product delivery delays",
          "significance": 0.82
        },
        {
          "type": "communication gaps",
          "description": "Lack of proactive shipping updates",
          "significance": 0.78
        }
      ]
    },
    "findings": {
      "answers": [
        {
          "question": "What is the customer's main concern?",
          "answer": "The customer's main concern is the lack of updates on their delayed order, especially since they need the item for an upcoming trip.",
          "confidence": 0.94
        },
        {
          "question": "How urgent is this issue?",
          "answer": "This issue is highly urgent due to the customer's upcoming trip and the extended period without tracking updates.",
          "confidence": 0.87
        }
      ]
    }
  },
  "meta": {
    "processing_time": 2.45,
    "steps_completed": ["attributes", "trends", "patterns", "findings"],
    "overall_confidence": 0.86,
    "result_id": "res-78912",
    "timestamp": "2023-07-15T14:32:18Z"
  }
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
}`,
    responseExample: `{
  "results": [
    {
      "id": "result-5678",
      "analysis_type": "trends",
      "timestamp": "2023-07-15T08:42:56Z",
      "data": {
        "trends": [
          {
            "focus_area": "customer satisfaction",
            "trend": "Upward trend in positive feedback"
          }
        ]
      }
    },
    {
      "id": "result-5679",
      "analysis_type": "patterns",
      "timestamp": "2023-07-15T09:12:23Z",
      "data": {
        "patterns": [
          {
            "type": "conversation_flow",
            "description": "Customers asking about delivery status after ordering"
          }
        ]
      }
    }
  ]
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
  try {
    // Structure the request according to StandardAnalysisRequest format
    const request = {
      workflow_id: options.workflowId,
      analysis_type: func.analysisType || func.id.split('-')[1],
      parameters: options.parameters || {},
      data: options.inputData || {},
      text: options.inputData?.text || ''
    };

    // Special handling for chain analysis
    if (func.id === 'analysis-chain') {
      return await api.performChainAnalysis(
        options.workflowId || '',
        options.inputData || {},
        options.config || {}
      );
    }

    // Use the standardized analysis endpoint for all other functions
    return await api.performAnalysis(
      request.analysis_type,
      request.parameters,
      request.data,
      request.text,
      request.workflow_id
    );
  } catch (error) {
    console.error(`Error executing function ${func.id}:`, error);
    throw error;
  }
};

export default function FunctionSettingsPanel({ selectedFunction, onClose }: FunctionSettingsProps) {
  const [activeTab, setActiveTab] = useState<string>('inputs');

  if (!selectedFunction) return null;

  const metadata = functionMetadata[selectedFunction.id] || {
    inputs: [],
    outputs: [],
    example: "No example available",
    responseExample: "No example response available"
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
      case 'analysis-recommendations':
        return <RecommendationsConfig function={selectedFunction} />;
      case 'analysis-plan':
        return <ActionPlanConfig function={selectedFunction} />;
      default:
        return <DefaultFunctionConfig function={selectedFunction} />;
    }
  };

  return (
    <div className="settings-panel bg-background border-l border-border h-full w-[512px] overflow-y-auto">
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

            <TabsContent value="example" className="px-4 mt-0 pb-4">
              {/* Request Example */}
              <div className="mb-6">
                <h3 className="text-sm font-medium mb-2 flex items-center">
                  <span className="inline-block w-2 h-2 rounded-full bg-emerald-500 mr-2"></span>
                  Example Request
                </h3>
                <pre className="bg-slate-100 dark:bg-slate-800 p-4 rounded text-sm overflow-x-auto border border-slate-200 dark:border-slate-700 font-mono whitespace-pre-wrap">
                  {metadata.example}
                </pre>
              </div>
              
              {/* Response Example */}
              <div>
                <h3 className="text-sm font-medium mb-2 flex items-center">
                  <span className="inline-block w-2 h-2 rounded-full bg-blue-500 mr-2"></span>
                  Example Response
                </h3>
                <pre className="bg-slate-100 dark:bg-slate-800 p-4 rounded text-sm overflow-x-auto border border-slate-200 dark:border-slate-700 font-mono whitespace-pre-wrap">
                  {metadata.responseExample}
                </pre>
              </div>
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

// RecommendationsConfig component for the Recommendations function
function RecommendationsConfig({ function: func }: { function: FunctionItem }) {
  const [focusArea, setFocusArea] = useState('customer retention');
  const [inputType, setInputType] = useState<'json'>('json');
  const [inputData, setInputData] = useState('{\n  "trends": [\n    {"focus_area": "customer satisfaction", "trend": "Declining satisfaction scores"},\n    {"focus_area": "response time", "trend": "Increasing average resolution time"}\n  ],\n  "findings": [\n    {"question": "What are the main pain points?", "answer": "Long wait times and complex processes"}\n  ]\n}');
  const [usePrioritization, setUsePrioritization] = useState(true);
  const [criteria, setCriteria] = useState('{\n  "impact": 0.6,\n  "implementation_ease": 0.4\n}');
  const [isExecuting, setIsExecuting] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);
  
  // Handle execution of recommendations generation
  const executeRecommendations = async () => {
    try {
      setIsExecuting(true);
      setError(null);
      
      // Parse input data
      let data: any;
      try {
        data = JSON.parse(inputData);
      } catch (e) {
        setError('Invalid JSON input data');
        setIsExecuting(false);
        return;
      }
      
      // Prepare parameters
      const parameters: Record<string, any> = {
        focus_area: focusArea
      };
      
      // Add prioritization criteria if enabled
      if (usePrioritization) {
        try {
          parameters.criteria = JSON.parse(criteria);
        } catch (e) {
          setError('Invalid JSON criteria');
          setIsExecuting(false);
          return;
        }
      }
      
      // Execute the function
      try {
        const result = await handleExecuteFunction(func, {
          parameters,
          inputData: data
        });
        
        setResult(result);
      } catch (e: unknown) {
        const errorMessage = e instanceof Error ? e.message : 'Error generating recommendations';
        setError(`Error: ${errorMessage}`);
      }
    } finally {
      setIsExecuting(false);
    }
  };
  
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h3 className="font-medium">Recommendations Configuration</h3>
        <p className="text-sm text-muted-foreground">
          Generate actionable recommendations based on analysis results.
        </p>
      </div>
      
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="focus-area">Focus Area</Label>
          <Input 
            id="focus-area" 
            placeholder="customer retention" 
            value={focusArea}
            onChange={(e) => setFocusArea(e.target.value)}
          />
          <p className="text-xs text-muted-foreground">
            The area to focus recommendations on (e.g., customer retention, agent performance)
          </p>
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="input-data-recommendations">Analysis Data (JSON)</Label>
          <Textarea 
            id="input-data-recommendations"
            placeholder="{ ... }"
            value={inputData}
            onChange={(e) => setInputData(e.target.value)}
            className="min-h-[150px] font-mono text-sm"
          />
          <p className="text-xs text-muted-foreground">
            Analysis results to base recommendations on (JSON format)
          </p>
        </div>
        
        <div className="flex items-center space-x-2">
          <Switch 
            id="use-prioritization" 
            checked={usePrioritization}
            onCheckedChange={setUsePrioritization}
          />
          <Label htmlFor="use-prioritization">Prioritize recommendations with criteria</Label>
        </div>
        
        {usePrioritization && (
          <div className="space-y-2">
            <Label htmlFor="criteria">Prioritization Criteria (JSON)</Label>
            <Textarea 
              id="criteria"
              placeholder="{ 'impact': 0.6, 'implementation_ease': 0.4 }"
              value={criteria}
              onChange={(e) => setCriteria(e.target.value)}
              className="min-h-[100px] font-mono text-sm"
            />
            <p className="text-xs text-muted-foreground">
              Criteria and weights for prioritization (must sum to 1.0)
            </p>
          </div>
        )}
      </div>
      
      <div className="pt-4 border-t">
        <Button 
          className="w-full" 
          onClick={executeRecommendations}
          disabled={isExecuting}
        >
          {isExecuting ? 'Generating...' : 'Generate Recommendations'}
        </Button>
        
        {error && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-600 text-sm">
            {error}
          </div>
        )}
        
        {result && !error && (
          <div className="mt-4 space-y-2">
            <h3 className="font-medium text-sm">Recommendations:</h3>
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

// ActionPlanConfig component for the Action Plan function
function ActionPlanConfig({ function: func }: { function: FunctionItem }) {
  const [planMode, setPlanMode] = useState<'full_plan' | 'timeline'>('full_plan');
  const [inputType, setInputType] = useState<'json'>('json');
  const [inputData, setInputData] = useState('{\n  "recommendations": {\n    "immediate_actions": [\n      {\n        "action": "Implement callback option",\n        "rationale": "Reduces customer frustration",\n        "expected_impact": "15% reduction in abandonment",\n        "priority": 5\n      }\n    ],\n    "implementation_notes": ["Begin with highest priority items"],\n    "success_metrics": ["Customer satisfaction scores"]\n  }\n}');
  const [constraints, setConstraints] = useState('{\n  "budget": 50000,\n  "timeline": "6 months",\n  "resources": ["customer_support", "engineering", "marketing"]\n}');
  const [resourceData, setResourceData] = useState('{\n  "staff": 5,\n  "start_date": "2023-10-01"\n}');
  const [isExecuting, setIsExecuting] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState<string | null>(null);
  
  // Handle execution of action plan generation
  const executeActionPlan = async () => {
    try {
      setIsExecuting(true);
      setError(null);
      
      // Parse input data
      let data: any;
      try {
        data = JSON.parse(inputData);
      } catch (e) {
        setError('Invalid JSON input data');
        setIsExecuting(false);
        return;
      }
      
      // Prepare parameters
      const parameters: Record<string, any> = {};
      
      // Set up parameters based on plan mode
      if (planMode === 'timeline') {
        parameters.generate_timeline = true;
        
        // For timeline mode, we need the action plan and resources
        try {
          const resources = JSON.parse(resourceData);
          data = {
            action_plan: data.action_plan || data, // Support both formats
            resources: resources
          };
        } catch (e) {
          setError('Invalid JSON resource data');
          setIsExecuting(false);
          return;
        }
      } else {
        // For full plan mode, add constraints
        try {
          parameters.constraints = JSON.parse(constraints);
        } catch (e) {
          setError('Invalid JSON constraints');
          setIsExecuting(false);
          return;
        }
      }
      
      // Execute the function
      try {
        const result = await handleExecuteFunction(func, {
          parameters,
          inputData: data
        });
        
        setResult(result);
      } catch (e: unknown) {
        const errorMessage = e instanceof Error ? e.message : 'Error generating action plan';
        setError(`Error: ${errorMessage}`);
      }
    } finally {
      setIsExecuting(false);
    }
  };
  
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h3 className="font-medium">Action Plan Configuration</h3>
        <p className="text-sm text-muted-foreground">
          Create implementation plans and timelines from recommendations.
        </p>
      </div>
      
      <div className="space-y-4">
        <div className="space-y-2">
          <Label>Plan Type</Label>
          <div className="flex space-x-2">
            <Button 
              variant={planMode === 'full_plan' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setPlanMode('full_plan')}
              className="flex-1"
            >
              Full Action Plan
            </Button>
            <Button 
              variant={planMode === 'timeline' ? 'default' : 'outline'} 
              size="sm"
              onClick={() => setPlanMode('timeline')}
              className="flex-1"
            >
              Timeline Only
            </Button>
          </div>
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="input-data-plan">
            {planMode === 'timeline' ? 'Action Plan Data (JSON)' : 'Recommendations Data (JSON)'}
          </Label>
          <Textarea 
            id="input-data-plan"
            placeholder="{ ... }"
            value={inputData}
            onChange={(e) => setInputData(e.target.value)}
            className="min-h-[150px] font-mono text-sm"
          />
          <p className="text-xs text-muted-foreground">
            {planMode === 'timeline' 
              ? 'Action plan to generate a timeline for' 
              : 'Recommendations to base the action plan on'}
          </p>
        </div>
        
        {planMode === 'full_plan' ? (
          <div className="space-y-2">
            <Label htmlFor="constraints">Constraints (JSON)</Label>
            <Textarea 
              id="constraints"
              placeholder="{ 'budget': 50000, 'timeline': '6 months' }"
              value={constraints}
              onChange={(e) => setConstraints(e.target.value)}
              className="min-h-[100px] font-mono text-sm"
            />
            <p className="text-xs text-muted-foreground">
              Constraints for the plan implementation
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            <Label htmlFor="resources">Resources (JSON)</Label>
            <Textarea 
              id="resources"
              placeholder="{ 'staff': 5, 'start_date': '2023-10-01' }"
              value={resourceData}
              onChange={(e) => setResourceData(e.target.value)}
              className="min-h-[100px] font-mono text-sm"
            />
            <p className="text-xs text-muted-foreground">
              Resources available for timeline generation
            </p>
          </div>
        )}
      </div>
      
      <div className="pt-4 border-t">
        <Button 
          className="w-full" 
          onClick={executeActionPlan}
          disabled={isExecuting}
        >
          {isExecuting ? 'Generating...' : planMode === 'timeline' ? 'Generate Timeline' : 'Create Action Plan'}
        </Button>
        
        {error && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-600 text-sm">
            {error}
          </div>
        )}
        
        {result && !error && (
          <div className="mt-4 space-y-2">
            <h3 className="font-medium text-sm">
              {planMode === 'timeline' ? 'Timeline:' : 'Action Plan:'}
            </h3>
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