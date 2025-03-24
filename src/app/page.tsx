"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import FlowEditor from "@/components/FlowEditor";
import { Plus, Save, Share2 } from "lucide-react";

export default function Home() {
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
      </div>

      <main className="flex-grow">
        <Card className="h-full rounded-none border-0">
          <CardContent className="p-0 h-[calc(100vh-124px)]">
            <FlowEditor />
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
