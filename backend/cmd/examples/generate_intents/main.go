package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"agenticflows/backend/cmd/examples/client"
	"agenticflows/backend/cmd/examples/utils"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Parse command-line flags
	dbPath := flag.String("db", "", "Path to the SQLite database")
	limit := flag.Int("limit", 10, "Number of conversations to analyze")
	debugFlag := flag.Bool("debug", false, "Enable debug mode")
	workflowID := flag.String("workflow", "", "Workflow ID for persisting results")
	mockFlag := flag.Bool("mock", false, "Use mock data instead of database")
	flag.Parse()

	// Validate required flags
	if *dbPath == "" && !*mockFlag {
		fmt.Println("Error: --db flag is required unless --mock is used")
		flag.Usage()
		os.Exit(1)
	}

	startTime := time.Now()

	// Create API client using the standardized client package
	apiClient := client.NewClient("http://localhost:8080", *workflowID, *debugFlag)

	// Print debug information if debug flag is enabled
	if *debugFlag {
		fmt.Println("Debug mode enabled: LLM inputs and outputs will be printed")
	}

	// Step 1: Fetch sample conversations from database or use mock data
	var conversations []utils.Conversation
	var err error

	if *mockFlag {
		fmt.Printf("Using mock data for %d sample conversations...\n", *limit)
		conversations = createMockConversations(*limit)
	} else {
		fmt.Printf("Fetching %d sample conversations from database...\n", *limit)
		conversations, err = fetchSampleConversations(*dbPath, *limit)
		if err != nil {
			fmt.Printf("Error fetching conversations: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Found %d conversations\n", len(conversations))

	// Step 2: Generate intents for each conversation
	fmt.Println("\nGenerating intents for conversations...")
	results := make([]map[string]interface{}, 0)

	for _, conv := range conversations {
		fmt.Printf("\nAnalyzing conversation %s...\n", conv.ID)

		// Use standardized API to generate intent
		req := client.StandardAnalysisRequest{
			AnalysisType: "intent",
			Text:         conv.Text,
			Parameters:   map[string]interface{}{},
		}

		resp, err := apiClient.PerformAnalysis(req)
		if err != nil {
			fmt.Printf("Error generating intent: %v\n", err)
			continue
		}

		// Extract intent from response
		if intentResults, ok := resp.Results.(map[string]interface{}); ok {
			result := map[string]interface{}{
				"conversation_id": conv.ID,
				"intent":          intentResults["intent"],
				"confidence":      resp.Confidence,
				"explanation":     intentResults["explanation"],
			}
			results = append(results, result)
		}
	}

	// Print results
	fmt.Println("\n=== Generated Intents ===")
	for _, result := range results {
		fmt.Printf("\nConversation ID: %s\n", result["conversation_id"])
		fmt.Printf("Intent: %s\n", result["intent"])
		fmt.Printf("Confidence: %.2f\n", result["confidence"])
		fmt.Printf("Explanation: %s\n", result["explanation"])
	}

	utils.PrintTimeTaken(startTime, "Generate intents")
}

// createMockConversations creates mock sample conversations
func createMockConversations(count int) []utils.Conversation {
	mockConversations := []utils.Conversation{
		{
			ID: "mock-conv-1",
			Text: `Customer: I'm having a problem with my account, I can't log in.
Agent: I'm sorry to hear that. I'd be happy to help you with your login issue. Could you please verify your email address?
Customer: It's john.smith@example.com
Agent: Thank you. I can see your account here. It looks like your account was temporarily locked due to multiple failed login attempts. I can reset it for you.
Customer: Yes, please unlock it. I really need to access my account today.
Agent: I've unlocked your account. Please try logging in again. You should also receive an email with instructions to reset your password.
Customer: Great, thank you so much for your help!
Agent: You're welcome! Is there anything else I can assist you with today?
Customer: No, that's all I needed. Have a good day.
Agent: Thank you for contacting us. Have a wonderful day!`,
		},
		{
			ID: "mock-conv-2",
			Text: `Customer: I'd like to cancel my subscription.
Agent: I'm sorry to hear you'd like to cancel. May I ask what's prompting you to cancel today?
Customer: It's just too expensive for what I'm getting.
Agent: I understand price is a concern. We do have some more affordable options that might better suit your needs. Would you be interested in hearing about those?
Customer: No, I've already decided to cancel.
Agent: I understand. I've gone ahead and processed your cancellation. Your service will remain active until the end of your current billing cycle on June 15th.
Customer: When will I get my refund?
Agent: Since you've used the service this month, there won't be a refund for the current period, but you won't be charged again. Is there anything else I can help with?
Customer: No, that's all.
Agent: Thank you for being our customer. If you decide to return in the future, we'll be happy to have you back.`,
		},
		{
			ID: "mock-conv-3",
			Text: `Customer: I ordered a product 5 days ago and it still hasn't arrived.
Agent: I apologize for the delay. I'd be happy to look into this for you. May I have your order number please?
Customer: It's #ORD-12345-67890
Agent: Thank you. I see your order is currently in transit. According to the tracking information, it should be delivered by tomorrow.
Customer: But I was promised it would arrive within 3 days when I placed the order.
Agent: I apologize for the miscommunication. I see there was a delay at our warehouse. As a goodwill gesture, I'd like to offer you a 15% discount on your next purchase.
Customer: Well, I really needed it for an event this evening.
Agent: I understand your frustration. Let me expedite this with our delivery team to see if we can get it to you today. Can I have your phone number to update you?
Customer: Yes, it's 555-123-4567
Agent: Thank you. I'll call you back within 30 minutes with an update on the delivery.`,
		},
		{
			ID: "mock-conv-4",
			Text: `Customer: I've been charged twice for my last payment.
Agent: I apologize for the duplicate charge. Let me look into that for you right away. May I have your account number?
Customer: It's ACT-987654
Agent: Thank you. I can see the duplicate charge on your account. I'll process a refund immediately. The funds should return to your account within 3-5 business days.
Customer: That's too long. I need that money now.
Agent: I understand your concern. While standard refunds take 3-5 days, I can process this as an expedited refund which should appear within 24 hours. Would that work better for you?
Customer: Yes, please do that.
Agent: I've processed the expedited refund. You'll receive a confirmation email shortly, and the funds should be back in your account within 24 hours.
Customer: Thank you for fixing this quickly.
Agent: You're welcome. I apologize again for the inconvenience. Is there anything else I can assist you with today?`,
		},
		{
			ID: "mock-conv-5",
			Text: `Customer: Hi, I'm trying to upgrade my plan but getting an error.
Agent: I'd be happy to help you upgrade your plan. What error message are you seeing?
Customer: It says "Unable to process request at this time."
Agent: Thank you for that information. Let me check what's happening. Can I have your account email, please?
Customer: It's sarah@example.com
Agent: Thank you, Sarah. I see the issue. There appears to be a temporary problem with our upgrade system. I can process this upgrade manually for you instead.
Customer: That would be great. I want to upgrade from the Basic to the Premium plan.
Agent: Perfect. I've manually upgraded your account to the Premium plan. The changes are effective immediately, and you should now have access to all Premium features.
Customer: Wonderful! How much will I be charged?
Agent: The Premium plan is $29.99 per month, but I've applied a 10% discount for the first three months due to the inconvenience. You'll see the prorated charge of $26.99 on your next statement.
Customer: Thank you so much for your help!`,
		},
	}

	// Return only the requested number of conversations
	if count < len(mockConversations) {
		return mockConversations[:count]
	}

	// If more conversations are requested than we have mock data for,
	// duplicate some conversations to reach the desired count
	result := make([]utils.Conversation, count)
	for i := 0; i < count; i++ {
		result[i] = mockConversations[i%len(mockConversations)]
		// Modify the ID to make it unique
		if i >= len(mockConversations) {
			result[i].ID = fmt.Sprintf("%s-dup-%d", result[i].ID, i/len(mockConversations))
		}
	}

	return result
}

// fetchSampleConversations fetches random sample conversations
func fetchSampleConversations(dbPath string, limit int) ([]utils.Conversation, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Query for random conversations
	query := `
	SELECT conversation_id, text
	FROM conversations
	WHERE text IS NOT NULL AND LENGTH(text) > 100
	ORDER BY RANDOM()
	LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	// Format conversations as objects
	conversations := make([]utils.Conversation, 0)
	for rows.Next() {
		var conv utils.Conversation
		if err := rows.Scan(&conv.ID, &conv.Text); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		conversations = append(conversations, conv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return conversations, nil
}
