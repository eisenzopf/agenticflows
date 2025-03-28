import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { X } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

interface WorkflowResultsViewerProps {
  results: Record<string, any>;
  onClose: () => void;
}

export default function WorkflowResultsViewer({ results, onClose }: WorkflowResultsViewerProps) {
  const [selectedTab, setSelectedTab] = useState<string>('summary');
  
  // Extract summary information from results
  const getResultSummary = () => {
    const nodeIds = Object.keys(results.results || {});
    const nodesExecuted = nodeIds.length;
    
    let findings: string[] = [];
    let patterns: string[] = [];
    let trends: string[] = [];
    let recommendations: string[] = [];
    
    // Loop through all node results to extract information
    nodeIds.forEach(nodeId => {
      const nodeResult = results.results[nodeId];
      
      if (nodeResult.findings && Array.isArray(nodeResult.findings)) {
        findings = [...findings, ...nodeResult.findings];
      }
      
      if (nodeResult.patterns && Array.isArray(nodeResult.patterns)) {
        patterns = [...patterns, ...nodeResult.patterns];
      }
      
      if (nodeResult.trend_descriptions && Array.isArray(nodeResult.trend_descriptions)) {
        trends = [...trends, ...nodeResult.trend_descriptions];
      }
      
      if (nodeResult.recommended_actions && Array.isArray(nodeResult.recommended_actions)) {
        recommendations = [...recommendations, ...nodeResult.recommended_actions];
      }
      
      if (nodeResult.recommendations && Array.isArray(nodeResult.recommendations)) {
        recommendations = [...recommendations, ...nodeResult.recommendations];
      }
    });
    
    return {
      nodesExecuted,
      findings,
      patterns,
      trends,
      recommendations
    };
  };
  
  const summary = getResultSummary();
  
  return (
    <div className="bg-card border shadow-lg rounded-lg max-w-2xl w-full h-4/5 flex flex-col">
      <div className="flex justify-between items-center p-4 border-b">
        <h2 className="text-lg font-semibold">Workflow Results</h2>
        <Button variant="ghost" size="sm" onClick={onClose}>
          <X size={16} />
        </Button>
      </div>
      
      <Tabs 
        defaultValue="summary" 
        value={selectedTab}
        onValueChange={setSelectedTab}
        className="flex-grow flex flex-col"
      >
        <TabsList className="mx-4 mt-2 justify-start">
          <TabsTrigger value="summary">Summary</TabsTrigger>
          <TabsTrigger value="trends">Trends</TabsTrigger>
          <TabsTrigger value="patterns">Patterns</TabsTrigger>
          <TabsTrigger value="findings">Findings</TabsTrigger>
          <TabsTrigger value="recommendations">Recommendations</TabsTrigger>
          <TabsTrigger value="raw">Raw Data</TabsTrigger>
        </TabsList>
        
        <div className="flex-grow overflow-auto p-4">
          <TabsContent value="summary" className="h-full">
            <div className="space-y-4">
              <div className="p-4 bg-muted rounded-md">
                <h3 className="font-semibold text-lg mb-2">Workflow Summary</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  Successfully executed {summary.nodesExecuted} nodes in the workflow.
                </p>
                
                <div className="grid grid-cols-2 gap-4">
                  <div className="border rounded-md p-3">
                    <h4 className="font-medium text-sm">Trends Identified</h4>
                    <p className="text-2xl font-bold">{summary.trends.length}</p>
                  </div>
                  
                  <div className="border rounded-md p-3">
                    <h4 className="font-medium text-sm">Patterns Found</h4>
                    <p className="text-2xl font-bold">{summary.patterns.length}</p>
                  </div>
                  
                  <div className="border rounded-md p-3">
                    <h4 className="font-medium text-sm">Findings Generated</h4>
                    <p className="text-2xl font-bold">{summary.findings.length}</p>
                  </div>
                  
                  <div className="border rounded-md p-3">
                    <h4 className="font-medium text-sm">Recommendations</h4>
                    <p className="text-2xl font-bold">{summary.recommendations.length}</p>
                  </div>
                </div>
              </div>
            </div>
          </TabsContent>
          
          <TabsContent value="trends" className="h-full">
            <h3 className="font-semibold text-lg mb-2">Identified Trends</h3>
            {summary.trends.length > 0 ? (
              <ul className="space-y-2">
                {summary.trends.map((trend, index) => (
                  <li key={index} className="p-2 border rounded-md">
                    {trend}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-muted-foreground text-center p-4">
                No trends were identified.
              </p>
            )}
          </TabsContent>
          
          <TabsContent value="patterns" className="h-full">
            <h3 className="font-semibold text-lg mb-2">Identified Patterns</h3>
            {summary.patterns.length > 0 ? (
              <ul className="space-y-2">
                {summary.patterns.map((pattern, index) => (
                  <li key={index} className="p-2 border rounded-md">
                    {pattern}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-muted-foreground text-center p-4">
                No patterns were identified.
              </p>
            )}
          </TabsContent>
          
          <TabsContent value="findings" className="h-full">
            <h3 className="font-semibold text-lg mb-2">Generated Findings</h3>
            {summary.findings.length > 0 ? (
              <ul className="space-y-2">
                {summary.findings.map((finding, index) => (
                  <li key={index} className="p-2 border rounded-md">
                    {finding}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-muted-foreground text-center p-4">
                No findings were generated.
              </p>
            )}
          </TabsContent>
          
          <TabsContent value="recommendations" className="h-full">
            <h3 className="font-semibold text-lg mb-2">Recommendations</h3>
            {summary.recommendations.length > 0 ? (
              <ul className="space-y-2">
                {summary.recommendations.map((recommendation, index) => (
                  <li key={index} className="p-2 border rounded-md">
                    {recommendation}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-muted-foreground text-center p-4">
                No recommendations were generated.
              </p>
            )}
          </TabsContent>
          
          <TabsContent value="raw" className="h-full">
            <h3 className="font-semibold text-lg mb-2">Raw Results Data</h3>
            <div className="bg-muted p-4 rounded-md overflow-auto h-[calc(100%-2rem)]">
              <pre className="text-xs">{JSON.stringify(results, null, 2)}</pre>
            </div>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
} 