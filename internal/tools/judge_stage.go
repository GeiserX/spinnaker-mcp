package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/geiserx/spinnaker-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewJudgeStage(gate *client.GateClient) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("judge_stage",
		mcp.WithDescription("Approve (continue) or reject (stop) a Manual Judgment stage that is waiting for input. continue lets the pipeline go on, usually into a production deploy; stop fails the stage. Confirm the decision with the user first. Find waiting judgments in get_execution's awaitingJudgment list, which gives the execution_id and stage_id. Refuses any stage that is not a running manual judgment."),
		mcp.WithString("execution_id",
			mcp.Required(),
			mcp.Description("Pipeline execution ID"),
		),
		mcp.WithString("stage_id",
			mcp.Required(),
			mcp.Description("Stage ID of the manual judgment (the id field in get_execution's awaitingJudgment, not the refId)"),
		),
		mcp.WithString("decision",
			mcp.Required(),
			mcp.Description("continue to approve, stop to reject"),
			mcp.Enum("continue", "stop"),
		),
		mcp.WithString("input",
			mcp.Description("Judgment input option to record, when the stage offers a list of options (e.g. a target environment)"),
		),
		destructive("Approve or reject manual judgment", false),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		execID, err := req.RequireString("execution_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		stageID, err := req.RequireString("stage_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		decision, err := req.RequireString("decision")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if decision != "continue" && decision != "stop" {
			return mcp.NewToolResultError(fmt.Sprintf("invalid decision %q: must be continue or stop", decision)), nil
		}

		// PATCH /pipelines/{id}/stages/{stageId} merges any map into any stage's context,
		// so check first that this really is a manual judgment waiting for input.
		body, err := gate.GetExecution(ctx, execID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		var ex struct {
			Stages []struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Type   string `json:"type"`
				Status string `json:"status"`
			} `json:"stages"`
		}
		if err := json.Unmarshal(body, &ex); err != nil {
			return mcp.NewToolResultError("could not read the execution from Gate"), nil
		}
		found := false
		for _, st := range ex.Stages {
			if st.ID != stageID {
				continue
			}
			found = true
			if st.Type != "manualJudgment" {
				return mcp.NewToolResultError(fmt.Sprintf("stage %s (%q) is a %s stage, not a manual judgment", stageID, st.Name, st.Type)), nil
			}
			if st.Status != "RUNNING" {
				return mcp.NewToolResultError(fmt.Sprintf("manual judgment %q is %s, not waiting for input", st.Name, st.Status)), nil
			}
		}
		if !found {
			return mcp.NewToolResultError(fmt.Sprintf("execution %s has no stage with id %s; use the id from get_execution's awaitingJudgment", execID, stageID)), nil
		}

		if _, err := gate.JudgeStage(ctx, execID, stageID, decision, req.GetString("input", "")); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Manual judgment %s recorded as %q in execution %s. Poll get_execution to follow the pipeline.", stageID, decision, execID)), nil
	}

	return tool, handler
}
