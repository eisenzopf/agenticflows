"use client";

import React, { useState } from 'react';
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api } from "@/services/api";

export default function QuestionAskerPage() {
  const [questions, setQuestions] = useState('');
  const [databasePath, setDatabasePath] = useState('/Users/jonathan/Documents/Work/discourse_ai/Research/corpora/banking_2025/db/standard_charter_bank.db');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [answers, setAnswers] = useState<Array<{ question: string, answer: string }>>([]);
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!questions.trim()) {
      setError('Please enter at least one question');
      return;
    }
    
    // Split questions by newline
    const questionsList = questions.split('\n')
      .map(q => q.trim())
      .filter(q => q.length > 0);
    
    if (questionsList.length === 0) {
      setError('Please enter at least one question');
      return;
    }
    
    setIsSubmitting(true);
    setError('');
    
    try {
      // Call the API to get answers
      const response = await api.answerQuestions(questionsList, databasePath);
      setAnswers(response.answers);
    } catch (err: any) {
      console.error('Error getting answers:', err);
      setError(err.message || 'Failed to get answers. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };
  
  return (
    <div className="container mx-auto max-w-3xl py-8">
      <div className="space-y-8">
        <div>
          <h1 className="text-2xl font-bold mb-2">Ask Questions About Banking Data</h1>
          <p className="text-muted-foreground">
            Ask questions about the banking conversations in the Standard Chartered Bank database.
          </p>
        </div>
        
        {error && (
          <div className="bg-destructive/10 text-destructive p-4 rounded-md">
            {error}
          </div>
        )}
        
        <form onSubmit={handleSubmit} className="space-y-6 border rounded-lg p-6">
          <div className="space-y-2">
            <Label htmlFor="databasePath">Database Path</Label>
            <Input 
              id="databasePath"
              value={databasePath}
              onChange={(e) => setDatabasePath(e.target.value)}
              disabled={isSubmitting}
            />
            <p className="text-sm text-muted-foreground">
              Path to the SQLite database containing banking conversations.
            </p>
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="questions">Your Questions</Label>
            <Textarea
              id="questions"
              value={questions}
              onChange={(e) => setQuestions(e.target.value)}
              placeholder="Enter your questions here, one per line. For example:

What are the most common customer complaints?
What banking features are customers requesting most often?
How does customer sentiment vary across different banking services?"
              rows={6}
              disabled={isSubmitting}
            />
            <p className="text-sm text-muted-foreground">
              Enter each question on a new line. Be specific to get better answers.
            </p>
          </div>
          
          <div className="flex justify-end">
            <Button 
              type="submit" 
              disabled={isSubmitting}
              className="min-w-[100px]"
            >
              {isSubmitting ? 'Processing...' : 'Get Answers'}
            </Button>
          </div>
        </form>
        
        {answers.length > 0 && (
          <div className="space-y-6">
            <h2 className="text-xl font-bold">Answers</h2>
            {answers.map((item, index) => (
              <div key={index} className="border rounded-lg p-6 space-y-4">
                <div className="font-medium text-lg">Q: {item.question}</div>
                <div className="text-muted-foreground whitespace-pre-wrap">{item.answer}</div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
} 