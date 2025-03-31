import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { X, Download, ExternalLink, ChevronDown, ChevronRight } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";

interface WorkflowResultsViewerProps {
  results: Record<string, any>;
  onClose: () => void;
}

export default function WorkflowResultsViewer({ results, onClose }: WorkflowResultsViewerProps) {
  const [selectedTab, setSelectedTab] = useState<string>('summary');
  const [categories, setCategories] = useState<string[]>(['summary']);
  
  // Extract categories and summary information from results
  useEffect(() => {
    if (!results) return;
    
    // Determine available categories
    const availableCategories = ['summary'];
    const nodeResults = results.results || {};
    
    // Check all node results for different analysis types
    Object.values(nodeResults).forEach((nodeResult: any) => {
      if (!nodeResult) return;
      
      // Add trends if present
      if (nodeResult.trend_descriptions || nodeResult.trends) {
        if (!availableCategories.includes('trends')) {
          availableCategories.push('trends');
        }
      }
      
      // Add patterns if present
      if (nodeResult.patterns) {
        if (!availableCategories.includes('patterns')) {
          availableCategories.push('patterns');
        }
      }
      
      // Add findings if present
      if (nodeResult.findings) {
        if (!availableCategories.includes('findings')) {
          availableCategories.push('findings');
        }
      }
      
      // Add recommendations if present
      if (nodeResult.recommendations || nodeResult.recommended_actions) {
        if (!availableCategories.includes('recommendations')) {
          availableCategories.push('recommendations');
        }
      }
      
      // Add disputes if present
      if (nodeResult.disputes) {
        if (!availableCategories.includes('disputes')) {
          availableCategories.push('disputes');
        }
      }
    });
    
    // Always include raw data
    if (!availableCategories.includes('raw')) {
      availableCategories.push('raw');
    }
    
    setCategories(availableCategories);
    
    // If current tab isn't available, select summary
    if (!availableCategories.includes(selectedTab)) {
      setSelectedTab('summary');
    }
  }, [results, selectedTab]);
  
  // Extract summary information from results
  const getSummary = () => {
    const nodeIds = Object.keys(results.results || {});
    const nodesExecuted = nodeIds.length;
    
    // Counters for different result types
    let trendsCount = 0;
    let patternsCount = 0;
    let findingsCount = 0;
    let recommendationsCount = 0;
    let disputesCount = 0;
    
    // Lists for collected items
    let allTrends: string[] = [];
    let allPatterns: string[] = [];
    let allFindings: string[] = [];
    let allRecommendations: string[] = [];
    
    // Loop through all node results to extract information
    nodeIds.forEach(nodeId => {
      const nodeResult = results.results[nodeId];
      if (!nodeResult) return;
      
      // Extract trends
      if (nodeResult.trend_descriptions && Array.isArray(nodeResult.trend_descriptions)) {
        trendsCount += nodeResult.trend_descriptions.length;
        allTrends = [...allTrends, ...nodeResult.trend_descriptions];
      } else if (nodeResult.trends && Array.isArray(nodeResult.trends)) {
        trendsCount += nodeResult.trends.length;
        // Extract trend text from possible object format
        const extractedTrends = nodeResult.trends.map((trend: any) => {
          if (typeof trend === 'string') return trend;
          if (trend.trend) return trend.trend;
          if (trend.description) return trend.description;
          return JSON.stringify(trend);
        });
        allTrends = [...allTrends, ...extractedTrends];
      }
      
      // Extract patterns
      if (nodeResult.patterns && Array.isArray(nodeResult.patterns)) {
        const extractedPatterns = nodeResult.patterns.map((pattern: any) => {
          if (typeof pattern === 'string') return pattern;
          if (pattern.pattern_description) return pattern.pattern_description;
          if (pattern.description) return pattern.description;
          return JSON.stringify(pattern);
        });
        patternsCount += extractedPatterns.length;
        allPatterns = [...allPatterns, ...extractedPatterns];
      }
      
      // Extract findings
      if (nodeResult.findings && Array.isArray(nodeResult.findings)) {
        findingsCount += nodeResult.findings.length;
        allFindings = [...allFindings, ...nodeResult.findings];
      }
      
      // Extract recommendations
      if (nodeResult.recommendations && Array.isArray(nodeResult.recommendations)) {
        recommendationsCount += nodeResult.recommendations.length;
        allRecommendations = [...allRecommendations, ...nodeResult.recommendations];
      } else if (nodeResult.recommended_actions && Array.isArray(nodeResult.recommended_actions)) {
        recommendationsCount += nodeResult.recommended_actions.length;
        allRecommendations = [...allRecommendations, ...nodeResult.recommended_actions];
      }
      
      // Count disputes if present
      if (nodeResult.disputes && Array.isArray(nodeResult.disputes)) {
        disputesCount += nodeResult.disputes.length;
      }
    });
    
    return {
      nodesExecuted,
      trendsCount,
      patternsCount,
      findingsCount,
      recommendationsCount,
      disputesCount,
      allTrends,
      allPatterns,
      allFindings,
      allRecommendations
    };
  };
  
  // Helper function to extract results by category from all nodes
  const getResultsByCategory = (category: string): any[] => {
    if (!results || !results.results) return [];
    
    const nodeIds = Object.keys(results.results);
    let extractedResults: any[] = [];
    
    nodeIds.forEach(nodeId => {
      const nodeResult = results.results[nodeId];
      if (!nodeResult) return;
      
      switch (category) {
        case 'trends':
          if (nodeResult.trend_descriptions && Array.isArray(nodeResult.trend_descriptions)) {
            extractedResults = [...extractedResults, ...nodeResult.trend_descriptions.map((item: string) => ({
              text: item,
              nodeId,
              type: 'trend'
            }))];
          } else if (nodeResult.trends && Array.isArray(nodeResult.trends)) {
            extractedResults = [...extractedResults, ...nodeResult.trends.map((trend: any) => {
              const text = typeof trend === 'string' ? trend : 
                trend.trend || trend.description || JSON.stringify(trend);
              return {
                text,
                details: typeof trend === 'object' ? trend : undefined,
                nodeId,
                type: 'trend'
              };
            })];
          }
          break;
          
        case 'patterns':
          if (nodeResult.patterns && Array.isArray(nodeResult.patterns)) {
            extractedResults = [...extractedResults, ...nodeResult.patterns.map((pattern: any) => {
              const text = typeof pattern === 'string' ? pattern : 
                pattern.pattern_description || pattern.description || JSON.stringify(pattern);
              return {
                text,
                details: typeof pattern === 'object' ? pattern : undefined,
                nodeId,
                type: 'pattern'
              };
            })];
          }
          break;
          
        case 'findings':
          if (nodeResult.findings && Array.isArray(nodeResult.findings)) {
            extractedResults = [...extractedResults, ...nodeResult.findings.map((item: string) => ({
              text: item,
              nodeId,
              type: 'finding'
            }))];
          }
          break;
          
        case 'recommendations':
          if (nodeResult.recommendations && Array.isArray(nodeResult.recommendations)) {
            extractedResults = [...extractedResults, ...nodeResult.recommendations.map((item: string) => ({
              text: item,
              nodeId,
              type: 'recommendation'
            }))];
          } else if (nodeResult.recommended_actions && Array.isArray(nodeResult.recommended_actions)) {
            extractedResults = [...extractedResults, ...nodeResult.recommended_actions.map((item: string) => ({
              text: item,
              nodeId,
              type: 'recommendation'
            }))];
          }
          break;
          
        case 'disputes':
          if (nodeResult.disputes && Array.isArray(nodeResult.disputes)) {
            extractedResults = [...extractedResults, ...nodeResult.disputes.map((dispute: any) => ({
              ...dispute,
              nodeId,
              type: 'dispute'
            }))];
          }
          break;
      }
    });
    
    return extractedResults;
  };
  
  // Handle download of results
  const handleDownloadResults = () => {
    try {
      const dataStr = JSON.stringify(results, null, 2);
      const dataUri = 'data:application/json;charset=utf-8,'+ encodeURIComponent(dataStr);
      
      const downloadAnchorNode = document.createElement('a');
      downloadAnchorNode.setAttribute('href', dataUri);
      downloadAnchorNode.setAttribute('download', `workflow-results-${new Date().toISOString().slice(0,10)}.json`);
      document.body.appendChild(downloadAnchorNode);
      downloadAnchorNode.click();
      downloadAnchorNode.remove();
    } catch (error) {
      console.error('Error downloading results:', error);
      alert('Failed to download results');
    }
  };
  
  const summary = getSummary();
  
  return (
    <div className="bg-card border shadow-lg rounded-lg max-w-3xl w-full h-4/5 flex flex-col">
      <div className="flex justify-between items-center p-4 border-b">
        <h2 className="text-lg font-semibold">Workflow Results</h2>
        <div className="flex gap-2">
          <Button 
            variant="outline" 
            size="sm" 
            onClick={handleDownloadResults}
            title="Download results as JSON"
          >
            <Download size={16} className="mr-1" /> Export
          </Button>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X size={16} />
          </Button>
        </div>
      </div>
      
      <Tabs 
        defaultValue="summary" 
        value={selectedTab}
        onValueChange={setSelectedTab}
        className="flex-grow flex flex-col"
      >
        <TabsList className="mx-4 mt-2 justify-start overflow-x-auto">
          {categories.map(category => (
            <TabsTrigger key={category} value={category} className="capitalize">
              {category}
            </TabsTrigger>
          ))}
        </TabsList>
        
        <div className="flex-grow overflow-auto p-4">
          <TabsContent value="summary" className="h-full">
            <div className="space-y-4">
              <div className="p-4 bg-muted rounded-md">
                <h3 className="font-semibold text-lg mb-2">Workflow Summary</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  Successfully executed {summary.nodesExecuted} nodes in the workflow.
                </p>
                
                <div className="grid grid-cols-2 gap-4 md:grid-cols-3">
                  {summary.trendsCount > 0 && (
                    <div className="border rounded-md p-3">
                      <h4 className="font-medium text-sm">Trends Identified</h4>
                      <p className="text-2xl font-bold">{summary.trendsCount}</p>
                    </div>
                  )}
                  
                  {summary.patternsCount > 0 && (
                    <div className="border rounded-md p-3">
                      <h4 className="font-medium text-sm">Patterns Found</h4>
                      <p className="text-2xl font-bold">{summary.patternsCount}</p>
                    </div>
                  )}
                  
                  {summary.findingsCount > 0 && (
                    <div className="border rounded-md p-3">
                      <h4 className="font-medium text-sm">Findings Generated</h4>
                      <p className="text-2xl font-bold">{summary.findingsCount}</p>
                    </div>
                  )}
                  
                  {summary.recommendationsCount > 0 && (
                    <div className="border rounded-md p-3">
                      <h4 className="font-medium text-sm">Recommendations</h4>
                      <p className="text-2xl font-bold">{summary.recommendationsCount}</p>
                    </div>
                  )}
                  
                  {summary.disputesCount > 0 && (
                    <div className="border rounded-md p-3">
                      <h4 className="font-medium text-sm">Disputes Analyzed</h4>
                      <p className="text-2xl font-bold">{summary.disputesCount}</p>
                    </div>
                  )}
                </div>
              </div>
              
              {/* Key findings and recommendations */}
              {summary.allFindings.length > 0 && (
                <Card>
                  <CardContent className="pt-6">
                    <h3 className="font-semibold text-md mb-2">Key Findings</h3>
                    <ul className="space-y-2">
                      {summary.allFindings.slice(0, 3).map((finding, index) => (
                        <li key={index} className="text-sm">
                          • {finding}
                        </li>
                      ))}
                      {summary.allFindings.length > 3 && (
                        <li className="text-sm text-muted-foreground">
                          <Button 
                            variant="link" 
                            className="p-0 h-auto" 
                            onClick={() => setSelectedTab('findings')}
                          >
                            See {summary.allFindings.length - 3} more findings
                          </Button>
                        </li>
                      )}
                    </ul>
                  </CardContent>
                </Card>
              )}
              
              {summary.allRecommendations.length > 0 && (
                <Card>
                  <CardContent className="pt-6">
                    <h3 className="font-semibold text-md mb-2">Top Recommendations</h3>
                    <ul className="space-y-2">
                      {summary.allRecommendations.slice(0, 3).map((rec, index) => (
                        <li key={index} className="text-sm">
                          • {rec}
                        </li>
                      ))}
                      {summary.allRecommendations.length > 3 && (
                        <li className="text-sm text-muted-foreground">
                          <Button 
                            variant="link" 
                            className="p-0 h-auto" 
                            onClick={() => setSelectedTab('recommendations')}
                          >
                            See {summary.allRecommendations.length - 3} more recommendations
                          </Button>
                        </li>
                      )}
                    </ul>
                  </CardContent>
                </Card>
              )}
            </div>
          </TabsContent>
          
          <TabsContent value="trends" className="h-full">
            <h3 className="font-semibold text-lg mb-2">Identified Trends</h3>
            <div className="space-y-3">
              {getResultsByCategory('trends').length > 0 ? (
                getResultsByCategory('trends').map((trend, index) => (
                  <Collapsible key={index} className="border rounded-md overflow-hidden">
                    <CollapsibleTrigger className="flex w-full items-center justify-between p-3 font-medium hover:bg-muted/50">
                      <div className="flex items-center">
                        <span className="mr-2">{trend.text}</span>
                        {trend.details && <Badge variant="outline" className="ml-2">Details Available</Badge>}
                      </div>
                      <ChevronDown className="h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-200" />
                    </CollapsibleTrigger>
                    {trend.details && (
                      <CollapsibleContent className="px-4 pb-3 pt-0">
                        <div className="rounded-md bg-muted p-2 text-sm">
                          {Object.entries(trend.details).map(([key, value]) => {
                            if (key === 'trend' || key === 'description') return null;
                            return (
                              <div key={key} className="flex py-1">
                                <span className="font-medium capitalize">{key.replace(/_/g, ' ')}:</span>
                                <span className="ml-2">{typeof value === 'string' ? value : JSON.stringify(value)}</span>
                              </div>
                            );
                          })}
                        </div>
                      </CollapsibleContent>
                    )}
                  </Collapsible>
                ))
              ) : (
                <p className="text-muted-foreground text-center p-4">
                  No trends were identified.
                </p>
              )}
            </div>
          </TabsContent>
          
          <TabsContent value="patterns" className="h-full">
            <h3 className="font-semibold text-lg mb-2">Identified Patterns</h3>
            <div className="space-y-3">
              {getResultsByCategory('patterns').length > 0 ? (
                getResultsByCategory('patterns').map((pattern, index) => (
                  <Collapsible key={index} className="border rounded-md overflow-hidden">
                    <CollapsibleTrigger className="flex w-full items-center justify-between p-3 font-medium hover:bg-muted/50">
                      <div className="flex items-center">
                        <span className="mr-2">{pattern.text}</span>
                        {pattern.details && <Badge variant="outline" className="ml-2">Details Available</Badge>}
                      </div>
                      <ChevronDown className="h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-200" />
                    </CollapsibleTrigger>
                    {pattern.details && (
                      <CollapsibleContent className="px-4 pb-3 pt-0">
                        <div className="rounded-md bg-muted p-2 text-sm">
                          {Object.entries(pattern.details).map(([key, value]) => {
                            if (key === 'pattern_description' || key === 'description') return null;
                            return (
                              <div key={key} className="flex py-1">
                                <span className="font-medium capitalize">{key.replace(/_/g, ' ')}:</span>
                                <span className="ml-2">{typeof value === 'string' ? value : JSON.stringify(value)}</span>
                              </div>
                            );
                          })}
                        </div>
                      </CollapsibleContent>
                    )}
                  </Collapsible>
                ))
              ) : (
                <p className="text-muted-foreground text-center p-4">
                  No patterns were identified.
                </p>
              )}
            </div>
          </TabsContent>
          
          <TabsContent value="findings" className="h-full">
            <h3 className="font-semibold text-lg mb-2">Generated Findings</h3>
            <div className="space-y-3">
              {getResultsByCategory('findings').length > 0 ? (
                getResultsByCategory('findings').map((finding, index) => (
                  <div key={index} className="p-3 border rounded-md">
                    {finding.text}
                  </div>
                ))
              ) : (
                <p className="text-muted-foreground text-center p-4">
                  No findings were generated.
                </p>
              )}
            </div>
          </TabsContent>
          
          <TabsContent value="recommendations" className="h-full">
            <h3 className="font-semibold text-lg mb-2">Recommendations</h3>
            <div className="space-y-3">
              {getResultsByCategory('recommendations').length > 0 ? (
                getResultsByCategory('recommendations').map((recommendation, index) => (
                  <div key={index} className="p-3 border rounded-md">
                    {recommendation.text}
                  </div>
                ))
              ) : (
                <p className="text-muted-foreground text-center p-4">
                  No recommendations were generated.
                </p>
              )}
            </div>
          </TabsContent>
          
          <TabsContent value="disputes" className="h-full">
            <h3 className="font-semibold text-lg mb-2">Analyzed Disputes</h3>
            <div className="space-y-3">
              {getResultsByCategory('disputes').length > 0 ? (
                getResultsByCategory('disputes').map((dispute, index) => (
                  <Collapsible key={index} className="border rounded-md overflow-hidden">
                    <CollapsibleTrigger className="flex w-full items-center justify-between p-3 font-medium hover:bg-muted/50">
                      <div className="flex items-center truncate">
                        <span className="mr-2 truncate">
                          {dispute.id ? `ID: ${dispute.id.substring(0, 8)}... ` : ''}
                          {dispute.amount ? `$${dispute.amount} ` : ''}
                          {dispute.text && dispute.text.length > 50
                            ? dispute.text.substring(0, 50) + '...'
                            : dispute.text || 'No text available'}
                        </span>
                      </div>
                      <ChevronDown className="h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-200" />
                    </CollapsibleTrigger>
                    <CollapsibleContent className="px-4 pb-3 pt-0">
                      <div className="rounded-md bg-muted p-2 text-sm">
                        {dispute.text && (
                          <div className="py-1">
                            <span className="font-medium">Text:</span>
                            <p className="mt-1">{dispute.text}</p>
                          </div>
                        )}
                        {dispute.amount && (
                          <div className="py-1">
                            <span className="font-medium">Amount:</span>
                            <span className="ml-2">${dispute.amount}</span>
                          </div>
                        )}
                        {dispute.created_at && (
                          <div className="py-1">
                            <span className="font-medium">Date:</span>
                            <span className="ml-2">{new Date(dispute.created_at).toLocaleString()}</span>
                          </div>
                        )}
                        {dispute.sentiment && (
                          <div className="py-1">
                            <span className="font-medium">Sentiment:</span>
                            <span className="ml-2">{dispute.sentiment}</span>
                          </div>
                        )}
                      </div>
                    </CollapsibleContent>
                  </Collapsible>
                ))
              ) : (
                <p className="text-muted-foreground text-center p-4">
                  No disputes were analyzed.
                </p>
              )}
            </div>
          </TabsContent>
          
          <TabsContent value="raw" className="h-full">
            <h3 className="font-semibold text-lg mb-2">Raw Results Data</h3>
            <div className="bg-muted p-4 rounded-md overflow-auto h-[calc(100%-2rem)]">
              <pre className="text-xs whitespace-pre-wrap">{JSON.stringify(results, null, 2)}</pre>
            </div>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
} 