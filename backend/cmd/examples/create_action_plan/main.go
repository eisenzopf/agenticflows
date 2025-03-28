package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"agenticflows/backend/cmd/examples/client"
	"agenticflows/backend/cmd/examples/utils"
)

// Main function
func main() {
	// Parse command-line flags
	// Keeping recommendationsFile for future implementation
	_ = flag.String("recommendations", "", "Path to recommendations JSON file (if empty, uses sample data)")
	genTimeline := flag.Bool("timeline", false, "Generate timeline instead of full action plan")
	budget := flag.Int("budget", 50000, "Budget constraint for the action plan")
	timespan := flag.String("timespan", "6 months", "Timespan for implementation")
	debug := flag.Bool("debug", false, "Enable debug mode")
	workflowID := flag.String("workflow", "", "Workflow ID for persisting results")
	// Adding mock flag for consistency, though this script already uses sample data
	_ = flag.Bool("mock", false, "Use mock data (this script always uses sample data)")
	flag.Parse()

	// Initialize API client
	apiClient := client.NewClient("http://localhost:8080", *workflowID, *debug)

	// Step 1: Prepare recommendation data (either from file or sample)
	fmt.Println("Preparing recommendation data...")
	recommendations := prepareSampleRecommendations()

	// Step 2: Create action plan or timeline
	var result interface{}
	var err error

	if *genTimeline {
		fmt.Println("\nGenerating implementation timeline...")
		result, err = generateTimeline(apiClient, recommendations, *timespan)
	} else {
		fmt.Println("\nCreating action plan...")
		result, err = createActionPlan(apiClient, recommendations, *budget, *timespan)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Print results
	fmt.Println("\n=== Results ===")

	if *genTimeline {
		printTimeline(result.(TimelineResult))
	} else {
		printActionPlan(result.(ActionPlanResult))
	}

	fmt.Println("\nAction Plan Generation complete!")
}

// RecommendationAction represents a recommended action
type RecommendationAction struct {
	Action         string `json:"action"`
	Rationale      string `json:"rationale"`
	ExpectedImpact string `json:"expected_impact"`
	Priority       int    `json:"priority"`
}

// RecommendationData represents sample recommendation data
type RecommendationData struct {
	ImmediateActions    []RecommendationAction `json:"immediate_actions"`
	ImplementationNotes []string               `json:"implementation_notes"`
	SuccessMetrics      []string               `json:"success_metrics"`
}

// ActionItem represents an action item in the plan
type ActionItem struct {
	Action          string   `json:"action"`
	Description     string   `json:"description"`
	Priority        int      `json:"priority"`
	EstimatedEffort string   `json:"estimated_effort"`
	Dependencies    []string `json:"dependencies"`
	ResponsibleRole string   `json:"responsible_role"`
}

// TimelinePhase represents a phase in the implementation timeline
type TimelinePhase struct {
	Phase       string   `json:"phase"`
	Description string   `json:"description"`
	Duration    string   `json:"duration"`
	Milestones  []string `json:"milestones"`
}

// RiskItem represents a risk and its mitigation strategy
type RiskItem struct {
	Risk             string `json:"risk"`
	Impact           string `json:"impact"`
	Probability      string `json:"probability"`
	MitigationPlan   string `json:"mitigation_plan"`
	ContingencyPlan  string `json:"contingency_plan"`
	ResponsibleParty string `json:"responsible_party"`
}

// ActionPlanResult represents the full action plan result
type ActionPlanResult struct {
	Goals              []string        `json:"goals"`
	ImmediateActions   []ActionItem    `json:"immediate_actions"`
	ShortTermActions   []ActionItem    `json:"short_term_actions"`
	LongTermActions    []ActionItem    `json:"long_term_actions"`
	ResponsibleParties []string        `json:"responsible_parties"`
	Timeline           []TimelinePhase `json:"timeline"`
	SuccessMetrics     []string        `json:"success_metrics"`
	RisksMitigations   []RiskItem      `json:"risks_mitigations"`
}

// TimelineResult represents just the timeline portion
type TimelineResult struct {
	Timeline       []TimelinePhase `json:"timeline"`
	StartDate      string          `json:"start_date"`
	EndDate        string          `json:"end_date"`
	TotalDuration  string          `json:"total_duration"`
	MilestoneDates []interface{}   `json:"milestone_dates"`
}

// prepareSampleRecommendations prepares sample recommendation data
func prepareSampleRecommendations() RecommendationData {
	// This is sample data that would typically come from the recommendations API
	return RecommendationData{
		ImmediateActions: []RecommendationAction{
			{
				Action:         "Implement callback option for customers on hold for more than 2 minutes",
				Rationale:      "Reduces customer frustration during peak call times",
				ExpectedImpact: "15% reduction in call abandonment rate",
				Priority:       5,
			},
			{
				Action:         "Simplify the refund process from 5 steps to 2 steps",
				Rationale:      "Current process is overly complex and leads to customer frustration",
				ExpectedImpact: "30% reduction in repeat calls about refunds",
				Priority:       4,
			},
			{
				Action:         "Proactively notify customers about known service issues",
				Rationale:      "Prevents unnecessary inbound contacts and shows proactive service",
				ExpectedImpact: "20% reduction in calls during service incidents",
				Priority:       3,
			},
		},
		ImplementationNotes: []string{
			"Begin with highest priority items requiring minimal IT changes",
			"Schedule implementation during low-volume periods",
			"Ensure customer service agents receive training on new processes",
		},
		SuccessMetrics: []string{
			"Customer satisfaction scores (target: 15% improvement in 90 days)",
			"First call resolution rate (target: increase from 65% to 80%)",
			"Average handle time (target: reduce by 45 seconds)",
		},
	}
}

// createActionPlan creates a comprehensive action plan based on recommendations
func createActionPlan(apiClient *client.Client, recommendations RecommendationData, budget int, timespan string) (ActionPlanResult, error) {
	emptyResult := ActionPlanResult{}

	// Define constraints
	constraints := map[string]interface{}{
		"budget":    budget,
		"timeline":  timespan,
		"resources": []string{"customer_support", "engineering", "marketing"},
	}

	// Request action plan
	fmt.Println("Requesting action plan from API...")
	req := client.StandardAnalysisRequest{
		AnalysisType: "plan",
		Parameters: map[string]interface{}{
			"constraints": constraints,
		},
		Data: map[string]interface{}{
			"recommendations": recommendations,
		},
	}

	resp, err := apiClient.PerformAnalysis(req)
	if err != nil {
		return emptyResult, fmt.Errorf("error creating action plan: %w", err)
	}

	// Parse action plan results
	var result ActionPlanResult
	if results, ok := resp.Results.(map[string]interface{}); ok {
		// Extract goals
		if goals, ok := results["goals"].([]interface{}); ok {
			result.Goals = make([]string, 0, len(goals))
			for _, g := range goals {
				if goal, ok := g.(string); ok {
					result.Goals = append(result.Goals, goal)
				}
			}
		}

		// Extract immediate actions
		result.ImmediateActions = extractActionItems(results, "immediate_actions")

		// Extract short-term actions
		result.ShortTermActions = extractActionItems(results, "short_term_actions")

		// Extract long-term actions
		result.LongTermActions = extractActionItems(results, "long_term_actions")

		// Extract responsible parties
		if parties, ok := results["responsible_parties"].([]interface{}); ok {
			result.ResponsibleParties = make([]string, 0, len(parties))
			for _, p := range parties {
				if party, ok := p.(string); ok {
					result.ResponsibleParties = append(result.ResponsibleParties, party)
				}
			}
		}

		// Extract timeline
		result.Timeline = extractTimelinePhases(results)

		// Extract success metrics
		if metrics, ok := results["success_metrics"].([]interface{}); ok {
			result.SuccessMetrics = make([]string, 0, len(metrics))
			for _, m := range metrics {
				if metric, ok := m.(string); ok {
					result.SuccessMetrics = append(result.SuccessMetrics, metric)
				}
			}
		}

		// Extract risks and mitigations
		if risks, ok := results["risks_mitigations"].([]interface{}); ok {
			result.RisksMitigations = make([]RiskItem, 0, len(risks))
			for _, r := range risks {
				if riskMap, ok := r.(map[string]interface{}); ok {
					risk := RiskItem{
						Risk:             utils.GetString(riskMap, "risk"),
						Impact:           utils.GetString(riskMap, "impact"),
						Probability:      utils.GetString(riskMap, "probability"),
						MitigationPlan:   utils.GetString(riskMap, "mitigation_plan"),
						ContingencyPlan:  utils.GetString(riskMap, "contingency_plan"),
						ResponsibleParty: utils.GetString(riskMap, "responsible_party"),
					}
					result.RisksMitigations = append(result.RisksMitigations, risk)
				}
			}
		}
	}

	return result, nil
}

// generateTimeline creates an implementation timeline based on the action plan
func generateTimeline(apiClient *client.Client, recommendations RecommendationData, timespan string) (TimelineResult, error) {
	emptyResult := TimelineResult{}

	// Define resources for timeline generation
	resources := map[string]interface{}{
		"staff":      5,
		"start_date": time.Now().Format("2006-01-02"),
	}

	// Request timeline
	fmt.Println("Requesting timeline from API...")
	req := client.StandardAnalysisRequest{
		AnalysisType: "plan",
		Parameters: map[string]interface{}{
			"generate_timeline": true,
		},
		Data: map[string]interface{}{
			"action_plan": map[string]interface{}{
				"recommendations": recommendations,
				"timespan":        timespan,
			},
			"resources": resources,
		},
	}

	resp, err := apiClient.PerformAnalysis(req)
	if err != nil {
		return emptyResult, fmt.Errorf("error generating timeline: %w", err)
	}

	// Parse timeline results
	var result TimelineResult
	if results, ok := resp.Results.(map[string]interface{}); ok {
		// Extract timeline
		result.Timeline = extractTimelinePhases(results)

		// Extract other timeline-specific details
		result.StartDate = utils.GetString(results, "start_date")
		result.EndDate = utils.GetString(results, "end_date")
		result.TotalDuration = utils.GetString(results, "total_duration")

		// Extract milestone dates
		if milestoneDates, ok := results["milestone_dates"].([]interface{}); ok {
			result.MilestoneDates = milestoneDates
		}
	}

	return result, nil
}

// extractActionItems extracts action items from the results
func extractActionItems(results map[string]interface{}, key string) []ActionItem {
	var items []ActionItem

	if actions, ok := results[key].([]interface{}); ok {
		items = make([]ActionItem, 0, len(actions))
		for _, a := range actions {
			if actionMap, ok := a.(map[string]interface{}); ok {
				// Extract dependencies as string array
				var dependencies []string
				if deps, ok := actionMap["dependencies"].([]interface{}); ok {
					dependencies = make([]string, 0, len(deps))
					for _, d := range deps {
						if dep, ok := d.(string); ok {
							dependencies = append(dependencies, dep)
						}
					}
				}

				action := ActionItem{
					Action:          utils.GetString(actionMap, "action"),
					Description:     utils.GetString(actionMap, "description"),
					Priority:        utils.GetInt(actionMap, "priority"),
					EstimatedEffort: utils.GetString(actionMap, "estimated_effort"),
					Dependencies:    dependencies,
					ResponsibleRole: utils.GetString(actionMap, "responsible_role"),
				}
				items = append(items, action)
			}
		}
	}

	return items
}

// extractTimelinePhases extracts timeline phases from the results
func extractTimelinePhases(results map[string]interface{}) []TimelinePhase {
	var phases []TimelinePhase

	if timeline, ok := results["timeline"].([]interface{}); ok {
		phases = make([]TimelinePhase, 0, len(timeline))
		for _, t := range timeline {
			if phaseMap, ok := t.(map[string]interface{}); ok {
				// Extract milestones as string array
				var milestones []string
				if ms, ok := phaseMap["milestones"].([]interface{}); ok {
					milestones = make([]string, 0, len(ms))
					for _, m := range ms {
						if milestone, ok := m.(string); ok {
							milestones = append(milestones, milestone)
						}
					}
				}

				phase := TimelinePhase{
					Phase:       utils.GetString(phaseMap, "phase"),
					Description: utils.GetString(phaseMap, "description"),
					Duration:    utils.GetString(phaseMap, "duration"),
					Milestones:  milestones,
				}
				phases = append(phases, phase)
			}
		}
	}

	return phases
}

// printActionPlan prints the action plan in a readable format
func printActionPlan(plan ActionPlanResult) {
	fmt.Println("\nGoals:")
	for i, goal := range plan.Goals {
		fmt.Printf("%d. %s\n", i+1, goal)
	}

	fmt.Println("\nImmediate Actions:")
	for i, action := range plan.ImmediateActions {
		fmt.Printf("%d. %s\n", i+1, action.Action)
		fmt.Printf("   Description: %s\n", action.Description)
		fmt.Printf("   Priority: %d\n", action.Priority)
		fmt.Printf("   Effort: %s\n", action.EstimatedEffort)
		fmt.Printf("   Role: %s\n", action.ResponsibleRole)

		if len(action.Dependencies) > 0 {
			fmt.Printf("   Dependencies: %v\n", action.Dependencies)
		}
		fmt.Println()
	}

	if len(plan.ShortTermActions) > 0 {
		fmt.Println("\nShort-Term Actions:")
		for i, action := range plan.ShortTermActions {
			fmt.Printf("%d. %s\n", i+1, action.Action)
		}
	}

	if len(plan.LongTermActions) > 0 {
		fmt.Println("\nLong-Term Actions:")
		for i, action := range plan.LongTermActions {
			fmt.Printf("%d. %s\n", i+1, action.Action)
		}
	}

	fmt.Println("\nImplementation Timeline:")
	for i, phase := range plan.Timeline {
		fmt.Printf("Phase %d: %s (%s)\n", i+1, phase.Phase, phase.Duration)
		fmt.Printf("   %s\n", phase.Description)
		fmt.Printf("   Milestones:\n")
		for _, milestone := range phase.Milestones {
			fmt.Printf("   - %s\n", milestone)
		}
		fmt.Println()
	}

	fmt.Println("\nKey Risks and Mitigations:")
	for i, risk := range plan.RisksMitigations {
		fmt.Printf("%d. Risk: %s\n", i+1, risk.Risk)
		fmt.Printf("   Impact: %s, Probability: %s\n", risk.Impact, risk.Probability)
		fmt.Printf("   Mitigation: %s\n", risk.MitigationPlan)
		fmt.Println()
	}

	fmt.Println("\nSuccess Metrics:")
	for i, metric := range plan.SuccessMetrics {
		fmt.Printf("%d. %s\n", i+1, metric)
	}
}

// printTimeline prints the timeline in a readable format
func printTimeline(timeline TimelineResult) {
	fmt.Printf("\nImplementation Timeline:\n")
	fmt.Printf("Start Date: %s\n", timeline.StartDate)
	fmt.Printf("End Date: %s\n", timeline.EndDate)
	fmt.Printf("Total Duration: %s\n\n", timeline.TotalDuration)

	for i, phase := range timeline.Timeline {
		fmt.Printf("Phase %d: %s (%s)\n", i+1, phase.Phase, phase.Duration)
		fmt.Printf("   %s\n", phase.Description)
		fmt.Printf("   Milestones:\n")
		for _, milestone := range phase.Milestones {
			fmt.Printf("   - %s\n", milestone)
		}
		fmt.Println()
	}
}
