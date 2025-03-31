"use client";

import { useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import FlowEditor, { FlowEditorHandle } from "@/components/FlowEditor";
import { Save, Sparkles } from "lucide-react";
import dynamic from "next/dynamic";

// Dynamically import the WorkflowGeneratorModal
const WorkflowGeneratorModal = dynamic(
  () => import("@/components/WorkflowGeneratorModal"),
  { ssr: false }
);

export default function Home() {
  // Create a ref to access the FlowEditor's methods
  const flowEditorRef = useRef<FlowEditorHandle>(null);
  const [isGeneratorOpen, setIsGeneratorOpen] = useState(false);

  const handleSaveFlow = async () => {
    if (flowEditorRef.current) {
      const result = await flowEditorRef.current.handleSaveWorkflow();
      if (result.success) {
        // Show success notification with workflow name
        alert(`"${result.workflowName}" saved successfully!`);
      }
    }
  };

  return (
    <div className="min-h-screen flex flex-col overflow-hidden">
      <header className="px-4 py-4 border-b">
        <div className="container mx-auto">
          <h1 className="text-2xl font-bold">Agent Flow Designer</h1>
        </div>
      </header>

      <div className="px-4 py-2 border-b">
        <div className="container mx-auto flex justify-between">
          <div className="flex gap-2">
            <Button 
              variant="outline" 
              size="sm"
              onClick={handleSaveFlow}
            >
              <Save className="h-4 w-4 mr-2" />
              Save Flow
            </Button>
            <Button 
              variant="outline" 
              size="sm"
              onClick={() => setIsGeneratorOpen(true)}
            >
              <Sparkles className="h-4 w-4 mr-2" />
              Generate Workflow
            </Button>
          </div>
        </div>
      </div>

      <main className="flex-grow">
        <Card className="h-full rounded-none border-0">
          <CardContent className="p-0 h-[calc(100vh-124px)]">
            <FlowEditor ref={flowEditorRef} />
          </CardContent>
        </Card>
      </main>

      {isGeneratorOpen && (
        <WorkflowGeneratorModal 
          onClose={() => setIsGeneratorOpen(false)}
          onWorkflowGenerated={(workflowId: string) => {
            // After workflow is generated, close the modal
            setIsGeneratorOpen(false);
            // Optionally: redirect or load the new workflow
            if (flowEditorRef.current) {
              flowEditorRef.current.loadWorkflow(workflowId);
            }
          }}
        />
      )}
    </div>
  );
}
