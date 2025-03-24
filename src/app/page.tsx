"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import FlowEditor from "@/components/FlowEditor";
import { Plus, Save, Share2 } from "lucide-react";

export default function Home() {
  return (
    <div className="min-h-screen p-4">
      <header className="container mx-auto py-4 border-b mb-4">
        <h1 className="text-2xl font-bold">Agent Flow Designer</h1>
      </header>

      <main className="container mx-auto">
        <div className="flex justify-between mb-4">
          <div className="flex gap-2">
            <Button variant="outline" size="sm">
              <Plus className="h-4 w-4 mr-2" />
              New Component
            </Button>
            <Button variant="outline" size="sm">
              <Save className="h-4 w-4 mr-2" />
              Save Flow
            </Button>
          </div>
          <Button size="sm">
            <Share2 className="h-4 w-4 mr-2" />
            Share
          </Button>
        </div>

        <Card className="mb-4">
          <CardHeader>
            <CardTitle>Flow Editor</CardTitle>
          </CardHeader>
          <CardContent>
            <FlowEditor />
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
