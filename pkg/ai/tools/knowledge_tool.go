package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pixelvide/kube-sentinel/pkg/model"
	openai "github.com/sashabaranov/go-openai"
	"gorm.io/datatypes"
)

type KnowledgeTool struct{}

func (t *KnowledgeTool) Name() string { return "manage_knowledge" }

func (t *KnowledgeTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "manage_knowledge",
			Description: "Manage the AI knowledge base for the current cluster. Use this to store patterns, rules, or important observations.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"action": {
						"type": "string",
						"enum": ["add", "list", "delete"],
						"description": "The action to perform."
					},
					"content": {
						"type": "string",
						"description": "The knowledge content (required for 'add'). Max 2 lines. Include resource pattern if applicable (e.g. 'Deployment *-prod: ...')."
					},
					"id": {
						"type": "integer",
						"description": "The ID of the knowledge entry (required for 'delete')."
					}
				},
				"required": ["action"]
			}`),
		},
	}
}

func (t *KnowledgeTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Action  string `json:"action"`
		Content string `json:"content"`
		ID      uint   `json:"id"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	// Cluster Name is injected into context by middleware
	clusterName, ok := ctx.Value(ClusterNameKey{}).(string)
	if !ok || clusterName == "" {
		return "", fmt.Errorf("cluster context is missing")
	}

	// We need Cluster ID. Look it up.
	cluster, err := model.GetClusterByName(clusterName)
	if err != nil {
		return "", fmt.Errorf("failed to get cluster: %w", err)
	}

	switch params.Action {
	case "add":
		if params.Content == "" {
			return "", fmt.Errorf("content is required for add action")
		}

		// Check for dups? For now, just add.
		kb := model.ClusterKnowledgeBase{
			ClusterID: cluster.ID,
			Content:   params.Content,
			AddedBy:   "AI",
			Metadata:  datatypes.JSON([]byte(`{"source":"ai_tool"}`)),
		}
		if err := model.AddKnowledge(&kb); err != nil {
			return "", err
		}
		return fmt.Sprintf("Knowledge added (ID: %d).", kb.ID), nil

	case "list":
		items, err := model.ListKnowledge(cluster.ID)
		if err != nil {
			return "", err
		}
		if len(items) == 0 {
			return "No knowledge found.", nil
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Knowledge for cluster %s:\n", clusterName))
		for _, item := range items {
			sb.WriteString(fmt.Sprintf("- [%d] %s (Added by: %s)\n", item.ID, item.Content, item.AddedBy))
		}
		return sb.String(), nil

	case "delete":
		if params.ID == 0 {
			return "", fmt.Errorf("id is required for delete action")
		}
		// Verify ownership or just allow AI to delete anything?
		// For safety, maybe check if it belongs to this cluster at least.
		item, err := model.GetKnowledgeByID(params.ID)
		if err != nil {
			return "", fmt.Errorf("knowledge not found")
		}
		if item.ClusterID != cluster.ID {
			return "", fmt.Errorf("cannot delete knowledge from another cluster")
		}

		if err := model.DeleteKnowledge(params.ID); err != nil {
			return "", err
		}
		return "Knowledge deleted.", nil

	default:
		return "", fmt.Errorf("unknown action: %s", params.Action)
	}
}
