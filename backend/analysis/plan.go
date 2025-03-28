package analysis

import (
	"context"
	"encoding/json"
	"fmt"
)

// Planner creates action plans based on analysis and recommendations
type Planner struct {
	llmClient *LLMClient
	debug     bool
}

// NewPlanner creates a new Planner instance
func NewPlanner(apiKey string, debug bool) (*Planner, error) {
	llmClient, err := NewLLMClient(apiKey, debug)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}
	return &Planner{
		llmClient: llmClient,
		debug:     debug,
	}, nil
}

// CreateActionPlan creates an implementation plan based on recommendations
func (p *Planner) CreateActionPlan(
	ctx context.Context,
	recommendations *RecommendationResponse,
	constraints map[string]interface{},
) (*ActionPlan, error) {
	// Validate input
	if recommendations == nil || len(recommendations.ImmediateActions) == 0 {
		return nil, fmt.Errorf("recommendations are required")
	}

	// Format recommendations for the prompt
	recsBytes, err := json.Marshal(recommendations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recommendations: %w", err)
	}

	// Format constraints for the prompt
	constraintsBytes, err := json.Marshal(constraints)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal constraints: %w", err)
	}

	prompt := fmt.Sprintf(`Create a comprehensive implementation plan for these recommendations:

Recommendations:
%s

Implementation Constraints:
%s

Develop a structured action plan that addresses all recommendations while considering the constraints.
Include short-term and long-term actions, timeline, responsible parties, and success metrics.

Format as JSON:
{
  "goals": [str],
  "immediate_actions": [
    {
      "action": str,
      "description": str,
      "priority": int,
      "estimated_effort": str,
      "dependencies": [str],
      "responsible_role": str
    }
  ],
  "short_term_actions": [
    {
      "action": str,
      "description": str,
      "priority": int,
      "estimated_effort": str,
      "dependencies": [str],
      "responsible_role": str
    }
  ],
  "long_term_actions": [
    {
      "action": str,
      "description": str,
      "priority": int,
      "estimated_effort": str,
      "dependencies": [str],
      "responsible_role": str
    }
  ],
  "responsible_parties": [str],
  "timeline": [
    {
      "phase": str,
      "description": str,
      "duration": str,
      "milestones": [str]
    }
  ],
  "success_metrics": [str],
  "risks_mitigations": [
    {
      "risk": str,
      "impact": str,
      "probability": str,
      "mitigation_plan": str,
      "contingency_plan": str,
      "responsible_party": str
    }
  ]
}`, string(recsBytes), string(constraintsBytes))

	expectedFormat := map[string]interface{}{
		"goals":               []interface{}{},
		"immediate_actions":   []interface{}{},
		"short_term_actions":  []interface{}{},
		"long_term_actions":   []interface{}{},
		"responsible_parties": []interface{}{},
		"timeline":            []interface{}{},
		"success_metrics":     []interface{}{},
		"risks_mitigations":   []interface{}{},
	}

	result, err := p.llmClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Parse the result into ActionPlan
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	plan := &ActionPlan{}

	// Extract goals
	if goalsRaw, ok := resultMap["goals"].([]interface{}); ok {
		for _, goalRaw := range goalsRaw {
			if goal, ok := goalRaw.(string); ok && goal != "" {
				plan.Goals = append(plan.Goals, goal)
			}
		}
	}

	// Extract immediate actions
	plan.ImmediateActions = p.extractActionItems(resultMap, "immediate_actions")

	// Extract short-term actions
	plan.ShortTermActions = p.extractActionItems(resultMap, "short_term_actions")

	// Extract long-term actions
	plan.LongTermActions = p.extractActionItems(resultMap, "long_term_actions")

	// Extract responsible parties
	if partiesRaw, ok := resultMap["responsible_parties"].([]interface{}); ok {
		for _, partyRaw := range partiesRaw {
			if party, ok := partyRaw.(string); ok && party != "" {
				plan.ResponsibleParties = append(plan.ResponsibleParties, party)
			}
		}
	}

	// Extract timeline
	if timelineRaw, ok := resultMap["timeline"].([]interface{}); ok {
		for _, eventRaw := range timelineRaw {
			if eventMap, ok := eventRaw.(map[string]interface{}); ok {
				event := TimelineEvent{
					Phase:       getString(eventMap, "phase"),
					Description: getString(eventMap, "description"),
					Duration:    getString(eventMap, "duration"),
				}

				// Extract milestones
				if milestonesRaw, ok := eventMap["milestones"].([]interface{}); ok {
					for _, milestoneRaw := range milestonesRaw {
						if milestone, ok := milestoneRaw.(string); ok && milestone != "" {
							event.Milestones = append(event.Milestones, milestone)
						}
					}
				}

				plan.Timeline = append(plan.Timeline, event)
			}
		}
	}

	// Extract success metrics
	if metricsRaw, ok := resultMap["success_metrics"].([]interface{}); ok {
		for _, metricRaw := range metricsRaw {
			if metric, ok := metricRaw.(string); ok && metric != "" {
				plan.SuccessMetrics = append(plan.SuccessMetrics, metric)
			}
		}
	}

	// Extract risks and mitigations
	if risksRaw, ok := resultMap["risks_mitigations"].([]interface{}); ok {
		for _, riskRaw := range risksRaw {
			if riskMap, ok := riskRaw.(map[string]interface{}); ok {
				risk := RiskItem{
					Risk:             getString(riskMap, "risk"),
					Impact:           getString(riskMap, "impact"),
					Probability:      getString(riskMap, "probability"),
					MitigationPlan:   getString(riskMap, "mitigation_plan"),
					ContingencyPlan:  getString(riskMap, "contingency_plan"),
					ResponsibleParty: getString(riskMap, "responsible_party"),
				}
				plan.RisksMitigations = append(plan.RisksMitigations, risk)
			}
		}
	}

	return plan, nil
}

// GenerateTimeline generates an implementation timeline for an action plan
func (p *Planner) GenerateTimeline(
	ctx context.Context,
	actionPlan *ActionPlan,
	resources map[string]interface{},
) ([]TimelineEvent, error) {
	// Validate input
	if actionPlan == nil {
		return nil, fmt.Errorf("action plan is required")
	}

	// Format action plan for the prompt
	planBytes, err := json.Marshal(actionPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal action plan: %w", err)
	}

	// Format resources for the prompt
	resourcesBytes, err := json.Marshal(resources)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resources: %w", err)
	}

	prompt := fmt.Sprintf(`Generate a detailed implementation timeline for this action plan:

Action Plan:
%s

Available Resources:
%s

Create a realistic implementation timeline considering dependencies between actions and available resources.
Include key phases, milestones, and estimated durations.

Format as JSON:
[
  {
    "phase": str,
    "description": str,
    "duration": str,
    "milestones": [str],
    "start_date": str,
    "end_date": str,
    "dependencies": [str],
    "resources_required": [str]
  }
]`, string(planBytes), string(resourcesBytes))

	result, err := p.llmClient.GenerateContent(ctx, prompt, []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Parse the result into TimelineEvents
	resultArray, ok := result.([]interface{})
	if !ok {
		resultMap, isMap := result.(map[string]interface{})
		if !isMap {
			return nil, fmt.Errorf("unexpected result format")
		}

		// Check if result is wrapped in a 'timeline' field
		if timeline, ok := resultMap["timeline"].([]interface{}); ok {
			resultArray = timeline
		} else {
			return nil, fmt.Errorf("unexpected result format, missing timeline array")
		}
	}

	// Convert to TimelineEvent objects
	timeline := make([]TimelineEvent, 0, len(resultArray))
	for _, eventRaw := range resultArray {
		if eventMap, ok := eventRaw.(map[string]interface{}); ok {
			event := TimelineEvent{
				Phase:       getString(eventMap, "phase"),
				Description: getString(eventMap, "description"),
				Duration:    getString(eventMap, "duration"),
			}

			// Extract milestones
			if milestonesRaw, ok := eventMap["milestones"].([]interface{}); ok {
				for _, milestoneRaw := range milestonesRaw {
					if milestone, ok := milestoneRaw.(string); ok && milestone != "" {
						event.Milestones = append(event.Milestones, milestone)
					}
				}
			}

			timeline = append(timeline, event)
		}
	}

	return timeline, nil
}

// extractActionItems extracts action items from a result map for a given key
func (p *Planner) extractActionItems(resultMap map[string]interface{}, key string) []ActionItem {
	items := []ActionItem{}
	if actionsRaw, ok := resultMap[key].([]interface{}); ok {
		for _, actionRaw := range actionsRaw {
			if actionMap, ok := actionRaw.(map[string]interface{}); ok {
				item := ActionItem{
					Action:          getString(actionMap, "action"),
					Description:     getString(actionMap, "description"),
					Priority:        int(getFloat(actionMap, "priority")),
					EstimatedEffort: getString(actionMap, "estimated_effort"),
					ResponsibleRole: getString(actionMap, "responsible_role"),
				}

				// Extract dependencies
				if depsRaw, ok := actionMap["dependencies"].([]interface{}); ok {
					for _, depRaw := range depsRaw {
						if dep, ok := depRaw.(string); ok && dep != "" {
							item.Dependencies = append(item.Dependencies, dep)
						}
					}
				}

				items = append(items, item)
			}
		}
	}
	return items
}
