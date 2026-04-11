package actions

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"gitea.com/gitea/gitea-mcp/pkg/errors"
	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	MonitorWorkflowDispatchToolName = "monitor_workflow_dispatch"
	DefaultPollInterval             = 10 * time.Second
	DefaultTimeout                  = 5 * time.Minute
)

var (
	MonitorWorkflowDispatchTool = mcp.NewTool(
		MonitorWorkflowDispatchToolName,
		mcp.WithDescription("Dispatch a workflow and monitor its execution until completion. Returns full execution summary including run ID, status, conclusion, duration, and logs."),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("workflow_id", mcp.Required(), mcp.Description("workflow ID or filename")),
		mcp.WithString("ref", mcp.Required(), mcp.Description("git ref (branch/tag) to run workflow on")),
		mcp.WithObject("inputs", mcp.Description("workflow inputs object")),
		mcp.WithNumber("timeout_seconds", mcp.Description("polling timeout in seconds (default: 300 = 5 minutes)"), mcp.DefaultNumber(300), mcp.Min(10)),
		mcp.WithNumber("poll_interval_seconds", mcp.Description("poll interval in seconds (default: 10)"), mcp.DefaultNumber(10), mcp.Min(5)),
	)
)

type MonitorResult struct {
	RunID        int64              `json:"run_id"`
	Status       string             `json:"status"`
	Conclusion   string             `json:"conclusion"`
	WorkflowID   string             `json:"workflow_id"`
	WorkflowName string             `json:"workflow_name,omitempty"`
	Branch       string             `json:"branch"`
	CommitSHA    string             `json:"commit_sha"`
	Duration     string             `json:"duration"`
	DurationSec  float64            `json:"duration_seconds"`
	StartedAt    string             `json:"started_at,omitempty"`
	CompletedAt  string             `json:"completed_at,omitempty"`
	Jobs         []JobSummary       `json:"jobs"`
	Logs         map[string]JobLogs `json:"logs,omitempty"`
	Error        string             `json:"error,omitempty"`
	TimedOut     bool               `json:"timed_out"`
}

type JobSummary struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Conclusion  string     `json:"conclusion"`
	StartedAt   string     `json:"started_at,omitempty"`
	CompletedAt string     `json:"completed_at,omitempty"`
	Steps       []StepInfo `json:"steps,omitempty"`
}

type StepInfo struct {
	Name       string `json:"name"`
	Number     int    `json:"number"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion,omitempty"`
}

type JobLogs struct {
	JobID     int64  `json:"job_id"`
	JobName   string `json:"job_name"`
	Log       string `json:"log,omitempty"`
	Bytes     int    `json:"bytes"`
	Truncated bool   `json:"truncated"`
}

func init() {
	Tool.RegisterWrite(server.ServerTool{
		Tool:    MonitorWorkflowDispatchTool,
		Handler: monitorWorkflowDispatchFn,
	})
}

func monitorWorkflowDispatchFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called monitorWorkflowDispatchFn")

	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil || owner == "" {
		return to.ErrorResult(errors.TranslateError(
			stderrors.New("owner is required"),
			map[string]string{"operation": "MonitorWorkflowDispatch"},
		))
	}

	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil || repo == "" {
		return to.ErrorResult(errors.TranslateError(
			stderrors.New("repo is required"),
			map[string]string{"operation": "MonitorWorkflowDispatch", "owner": owner},
		))
	}

	workflowID, err := params.GetString(req.GetArguments(), "workflow_id")
	if err != nil || workflowID == "" {
		return to.ErrorResult(errors.TranslateError(
			stderrors.New("workflow_id is required"),
			map[string]string{"operation": "MonitorWorkflowDispatch", "owner": owner, "repo": repo},
		))
	}

	ref, err := params.GetString(req.GetArguments(), "ref")
	if err != nil || ref == "" {
		return to.ErrorResult(errors.TranslateError(
			stderrors.New("ref is required"),
			map[string]string{"operation": "MonitorWorkflowDispatch", "owner": owner, "repo": repo, "workflow_id": workflowID},
		))
	}

	timeoutSec := int(params.GetOptionalInt(req.GetArguments(), "timeout_seconds", 300))
	pollIntervalSec := int(params.GetOptionalInt(req.GetArguments(), "poll_interval_seconds", 10))

	var inputs map[string]any
	if raw, exists := req.GetArguments()["inputs"]; exists {
		if m, ok := raw.(map[string]any); ok {
			inputs = m
		}
	}

	if err := checkActionsAPIAvailable(ctx); err != nil {
		return to.ErrorResult(err)
	}

	log.Infof("Dispatching workflow %s for %s/%s on ref %s", workflowID, owner, repo, ref)
	_, err = dispatchWorkflow(ctx, owner, repo, workflowID, ref, inputs)
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "DispatchWorkflow",
			"owner":     owner,
			"repo":      repo,
			"workflow":  workflowID,
		}))
	}

	timeout := time.Duration(timeoutSec) * time.Second
	pollInterval := time.Duration(pollIntervalSec) * time.Second

	log.Infof("Waiting for workflow run to start (timeout: %v)...", timeout)
	runID, err := waitForRunToStart(ctx, owner, repo, workflowID, ref, pollInterval, timeout)
	if err != nil {
		result := MonitorResult{
			WorkflowID: workflowID,
			Branch:     ref,
			Error:      fmt.Sprintf("Workflow dispatched but run never started: %v", err),
		}
		return toErrorResultWithJSON(result)
	}

	log.Infof("Run %d started, monitoring until completion...", runID)
	monitorResult, err := monitorRunUntilComplete(ctx, owner, repo, runID, pollInterval, timeout)
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "MonitorRun",
			"owner":     owner,
			"repo":      repo,
			"run_id":    fmt.Sprintf("%d", runID),
		}))
	}

	log.Infof("Retrieving logs for run %d...", runID)
	logs, err := retrieveJobLogs(ctx, owner, repo, monitorResult.Jobs)
	if err != nil {
		log.Warnf("Failed to retrieve some job logs: %v", err)
	}
	monitorResult.Logs = logs

	return toTextResultWithJSON(monitorResult)
}

func checkActionsAPIAvailable(ctx context.Context) error {
	var versionResp struct {
		Version string `json:"version"`
	}
	status, err := gitea.DoJSON(ctx, "GET", "version", nil, nil, &versionResp)
	if err != nil {
		return errors.TranslateError(
			fmt.Errorf("failed to check Gitea version: status=%d, err=%v", status, err),
			map[string]string{"operation": "CheckGiteaVersion"},
		)
	}

	major, minor, patch, err := parseVersionForCheck(versionResp.Version)
	if err != nil {
		return errors.TranslateError(
			fmt.Errorf("failed to parse version '%s': %v", versionResp.Version, err),
			map[string]string{"operation": "ParseVersion"},
		)
	}

	if major < 1 || (major == 1 && minor < 23) {
		return errors.NewEnhancedError(
			stderrors.New("Actions API not available"),
			fmt.Sprintf("Actions API requires Gitea 1.23+, found %d.%d.%d", major, minor, patch),
			errors.CategoryActions,
		).WithOperation("CheckActionsAPIAvailable")
	}

	return nil
}

func parseVersionForCheck(version string) (int, int, int, error) {
	version = trimVersionPrefix(version)
	parts := splitVersion(version)
	if len(parts) < 2 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch := 0
	if len(parts) >= 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			patch = 0
		}
	}

	return major, minor, patch, nil
}

func trimVersionPrefix(v string) string {
	v = trimOnePrefix(v, "v")
	v = trimOnePrefix(v, "V")
	return v
}

func trimOnePrefix(s, prefix string) string {
	if len(s) > 0 && s[0] == prefix[0] {
		return s[1:]
	}
	return s
}

func splitVersion(v string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(v); i++ {
		if v[i] == '.' {
			if start < i {
				parts = append(parts, v[start:i])
			}
			start = i + 1
		}
	}
	if start < len(v) {
		parts = append(parts, v[start:])
	}
	return parts
}

func dispatchWorkflow(ctx context.Context, owner, repo, workflowID, ref string, inputs map[string]any) (map[string]any, error) {
	body := map[string]any{
		"ref": ref,
	}
	if inputs != nil {
		body["inputs"] = inputs
	}

	err := doJSONWithFallback(ctx, "POST",
		[]string{
			fmt.Sprintf("repos/%s/%s/actions/workflows/%s/dispatches", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(workflowID)),
			fmt.Sprintf("repos/%s/%s/actions/workflows/%s/dispatch", url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(workflowID)),
		},
		nil, body, nil,
	)
	if err != nil {
		var httpErr *gitea.HTTPError
		if stderrors.As(err, &httpErr) && (httpErr.StatusCode == http.StatusNotFound || httpErr.StatusCode == http.StatusMethodNotAllowed) {
			return nil, errors.NewEnhancedError(
				err,
				fmt.Sprintf("workflow dispatch not supported on this Gitea version (endpoint returned %d)", httpErr.StatusCode),
				errors.CategoryActions,
			).WithOperation("DispatchWorkflow")
		}
		return nil, err
	}

	return map[string]any{"message": "workflow dispatched"}, nil
}

func waitForRunToStart(ctx context.Context, owner, repo, workflowID, ref string, pollInterval, timeout time.Duration) (int64, error) {
	startTime := time.Now()
	seenRunIDs := make(map[int64]bool)

	for time.Since(startTime) < timeout {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}

		runs, err := listRecentRuns(ctx, owner, repo, workflowID, ref, 10)
		if err != nil {
			log.Warnf("Failed to list runs: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		for _, run := range runs {
			runID := int64(run["id"].(float64))
			if seenRunIDs[runID] {
				continue
			}
			seenRunIDs[runID] = true

			status := getString(run, "status")
			if status == "queued" || status == "in_progress" || status == "waiting" {
				return runID, nil
			}
			createdAt := getString(run, "created_at")
			if createdAt != "" {
				runTime, err := time.Parse(time.RFC3339, createdAt)
				if err == nil && time.Since(runTime) < 2*time.Minute {
					return runID, nil
				}
			}
		}

		time.Sleep(pollInterval)
	}

	return 0, fmt.Errorf("timeout waiting for run to start after %v", timeout)
}

func listRecentRuns(ctx context.Context, owner, repo, workflowID, ref string, limit int) ([]map[string]any, error) {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))

	var result struct {
		WorkflowRuns []map[string]any `json:"workflow_runs"`
	}

	err := doJSONWithFallback(ctx, "GET",
		[]string{
			fmt.Sprintf("repos/%s/%s/actions/runs", url.PathEscape(owner), url.PathEscape(repo)),
		},
		query, nil, &result,
	)
	if err != nil {
		return nil, err
	}

	var matchingRuns []map[string]any
	for _, run := range result.WorkflowRuns {
		if workflowID != "" {
			runWorkflowID := getString(run, "workflow_id")
			runPath := getString(run, "path")
			if runWorkflowID != workflowID && runPath != workflowID &&
				runWorkflowID != "" && !containsPath(runPath, workflowID) {
				continue
			}
		}
		if ref != "" {
			headBranch := getString(run, "head_branch")
			if headBranch != "" && headBranch != ref {
				continue
			}
		}
		matchingRuns = append(matchingRuns, run)
	}

	return matchingRuns, nil
}

func containsPath(path, substring string) bool {
	return len(path) > 0 && len(substring) > 0 &&
		(path == substring ||
			(len(path) > len(substring) && (path[len(path)-len(substring):] == substring ||
				path[:len(substring)] == substring)))
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	if v, ok := m[key].(float64); ok {
		return strconv.FormatInt(int64(v), 10)
	}
	return ""
}

func monitorRunUntilComplete(ctx context.Context, owner, repo string, runID int64, pollInterval, timeout time.Duration) (*MonitorResult, error) {
	startTime := time.Now()
	var firstSeen time.Time

	for time.Since(startTime) < timeout {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		run, err := getRun(ctx, owner, repo, runID)
		if err != nil {
			return nil, fmt.Errorf("failed to get run %d: %v", runID, err)
		}

		status := getString(run, "status")
		conclusion := getString(run, "conclusion")

		if firstSeen.IsZero() {
			firstSeen = time.Now()
		}

		if status == "completed" {
			return buildMonitorResult(run, runID, owner, repo, firstSeen, time.Now())
		}

		log.Debugf("Run %d status: %s (conclusion: %s), waiting...", runID, status, conclusion)
		time.Sleep(pollInterval)
	}

	run, err := getRun(ctx, owner, repo, runID)
	if err != nil {
		return nil, fmt.Errorf("timeout after %v and failed to get final status: %v", timeout, err)
	}

	result, _ := buildMonitorResult(run, runID, owner, repo, firstSeen, time.Now())
	result.TimedOut = true
	result.Error = fmt.Sprintf("Polling timed out after %v", timeout)
	return result, nil
}

func getRun(ctx context.Context, owner, repo string, runID int64) (map[string]any, error) {
	var result map[string]any
	err := doJSONWithFallback(ctx, "GET",
		[]string{
			fmt.Sprintf("repos/%s/%s/actions/runs/%d", url.PathEscape(owner), url.PathEscape(repo), runID),
		},
		nil, nil, &result,
	)
	return result, err
}

func buildMonitorResult(run map[string]any, runID int64, owner, repo string, started, ended time.Time) (*MonitorResult, error) {
	result := &MonitorResult{
		RunID:        runID,
		Status:       getString(run, "status"),
		Conclusion:   getString(run, "conclusion"),
		WorkflowID:   getString(run, "workflow_id"),
		WorkflowName: getString(run, "name"),
		Branch:       getString(run, "head_branch"),
		CommitSHA:    getString(run, "head_sha"),
		StartedAt:    getString(run, "created_at"),
		CompletedAt:  getString(run, "updated_at"),
	}

	if createdAt := getString(run, "run_started_at"); createdAt != "" {
		result.StartedAt = createdAt
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			started = t
		}
	}
	if updatedAt := getString(run, "updated_at"); updatedAt != "" {
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			ended = t
		}
	}

	duration := ended.Sub(started)
	if duration < 0 {
		duration = 0
	}
	result.Duration = duration.String()
	result.DurationSec = duration.Seconds()

	jobs, err := listRunJobs(runID, owner, repo, run)
	if err == nil {
		result.Jobs = jobs
	}

	return result, nil
}

func listRunJobs(runID int64, owner, repo string, run map[string]any) ([]JobSummary, error) {
	if jobsData, ok := run["jobs"].([]any); ok && len(jobsData) > 0 {
		return parseJobSummaries(jobsData), nil
	}
	return nil, nil
}

func parseJobSummaries(jobsData []any) []JobSummary {
	var summaries []JobSummary
	for _, j := range jobsData {
		job, ok := j.(map[string]any)
		if !ok {
			continue
		}

		summary := JobSummary{
			ID:          int64(job["id"].(float64)),
			Name:        getString(job, "name"),
			Status:      getString(job, "status"),
			Conclusion:  getString(job, "conclusion"),
			StartedAt:   getString(job, "started_at"),
			CompletedAt: getString(job, "completed_at"),
		}

		if stepsData, ok := job["steps"].([]any); ok {
			summary.Steps = parseSteps(stepsData)
		}

		summaries = append(summaries, summary)
	}
	return summaries
}

func parseSteps(stepsData []any) []StepInfo {
	var steps []StepInfo
	for _, s := range stepsData {
		step, ok := s.(map[string]any)
		if !ok {
			continue
		}
		steps = append(steps, StepInfo{
			Name:       getString(step, "name"),
			Number:     int(step["number"].(float64)),
			Status:     getString(step, "status"),
			Conclusion: getString(step, "conclusion"),
		})
	}
	return steps
}

func retrieveJobLogs(ctx context.Context, owner, repo string, jobs []JobSummary) (map[string]JobLogs, error) {
	logs := make(map[string]JobLogs)

	for _, job := range jobs {
		if job.ID == 0 {
			continue
		}

		logData, _, err := fetchJobLogBytes(ctx, owner, repo, job.ID)
		if err != nil {
			log.Warnf("Failed to fetch logs for job %d: %v", job.ID, err)
			continue
		}

		maxLogBytes := 100 * 1024
		truncated := false
		if len(logData) > maxLogBytes {
			logData = logData[len(logData)-maxLogBytes:]
			truncated = true
		}

		logs[job.Name] = JobLogs{
			JobID:     job.ID,
			JobName:   job.Name,
			Log:       string(logData),
			Bytes:     len(logData),
			Truncated: truncated,
		}
	}

	return logs, nil
}

func toTextResultWithJSON(result *MonitorResult) (*mcp.CallToolResult, error) {
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return to.ErrorResult(fmt.Errorf("failed to marshal result: %v", err))
	}
	return to.TextResult(string(jsonBytes))
}

func toErrorResultWithJSON(result MonitorResult) (*mcp.CallToolResult, error) {
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return to.ErrorResult(fmt.Errorf("%s (marshal error: %v)", result.Error, err))
	}
	return to.TextResult(fmt.Sprintf("Error: %s\n\nPartial Result:\n%s", result.Error, string(jsonBytes)))
}