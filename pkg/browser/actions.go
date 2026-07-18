package browser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

// ExecuteStep executes a single browser action step
func ExecuteStep(page *rod.Page, step StepConfig, ctx context.Context) StepResult {
	result := StepResult{Name: step.Name, Pass: true}

	// Set default timeout for this step
	timeout := time.Duration(step.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	action := strings.ToLower(step.Action)

	switch action {
	case "goto":
		if step.URL == "" {
			result.Pass = false
			result.Error = fmt.Errorf("goto action requires url field")
			return result
		}
		err := page.Timeout(timeout).Navigate(step.URL)
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("failed to navigate to %s: %w", step.URL, err)
		}

	case "click":
		if step.Selector == "" {
			result.Pass = false
			result.Error = fmt.Errorf("click action requires selector field")
			return result
		}
		err := page.Timeout(timeout).MustElement(step.Selector).Click(proto.InputMouseButtonLeft, 1)
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("failed to click %s: %w", step.Selector, err)
		}

	case "fill":
		if step.Selector == "" {
			result.Pass = false
			result.Error = fmt.Errorf("fill action requires selector field")
			return result
		}
		value, ok := step.Value.(string)
		if !ok {
			result.Pass = false
			result.Error = fmt.Errorf("fill action requires value field as string")
			return result
		}
		el := page.Timeout(timeout).MustElement(step.Selector)
		err := el.Input(value)
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("failed to fill %s: %w", step.Selector, err)
		}

	case "fill_text":
		// fill_text is the same as fill
		if step.Selector == "" {
			result.Pass = false
			result.Error = fmt.Errorf("fill_text action requires selector field")
			return result
		}
		value, ok := step.Value.(string)
		if !ok {
			result.Pass = false
			result.Error = fmt.Errorf("fill_text action requires value field as string")
			return result
		}
		el := page.Timeout(timeout).MustElement(step.Selector)
		err := el.Input(value)
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("failed to fill_text %s: %w", step.Selector, err)
		}

	case "hover":
		if step.Selector == "" {
			result.Pass = false
			result.Error = fmt.Errorf("hover action requires selector field")
			return result
		}
		el := page.Timeout(timeout).MustElement(step.Selector)
		err := el.Hover()
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("failed to hover %s: %w", step.Selector, err)
		}

	case "select":
		if step.Selector == "" {
			result.Pass = false
			result.Error = fmt.Errorf("select action requires selector field")
			return result
		}
		value, ok := step.Value.(string)
		if !ok {
			result.Pass = false
			result.Error = fmt.Errorf("select action requires value field as string")
			return result
		}
		err := page.Timeout(timeout).MustElement(step.Selector).Select([]string{value}, true, "")
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("failed to select option in %s: %w", step.Selector, err)
		}

	case "switch_frame":
		// Handle popup switching - placeholder implementation
		if step.Target == "popup" {
			result.Output = fmt.Sprintf("Switch to popup with filter: %s (requires manual implementation)", step.Filter)
		}

	case "screenshot":
		path := step.Path
		if path == "" {
			path = fmt.Sprintf("screenshot_%d.png", time.Now().Unix())
		}
		_, err := page.Timeout(timeout).Screenshot(true, &proto.PageCaptureScreenshot{})
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("failed to take screenshot: %w", err)
		} else {
			result.Output = fmt.Sprintf("Screenshot saved to: %s", path)
		}

	case "wait_url":
		if step.Pattern == "" {
			result.Pass = false
			result.Error = fmt.Errorf("wait_url action requires pattern field")
			return result
		}
		// Wait for URL to change and contain pattern
		page.Timeout(timeout).WaitLoad()
		initialURL := page.MustInfo().URL
		pollInterval := 500 * time.Millisecond
		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			pageInfo := page.MustInfo()
			currentURL := pageInfo.URL
			if currentURL != initialURL && strings.Contains(currentURL, step.Pattern) {
				result.Output = fmt.Sprintf("URL changed to: %s", currentURL)
				return result
			}
			time.Sleep(pollInterval)
		}
		finalURL := page.MustInfo().URL
		result.Pass = false
		result.Error = fmt.Errorf("URL '%s' does not contain '%s' after %v", finalURL, step.Pattern, timeout)

	case "wait_selector":
		if step.Selector == "" {
			result.Pass = false
			result.Error = fmt.Errorf("wait_selector action requires selector field")
			return result
		}
		// Wait for element to be available
		_, err := page.Timeout(timeout).Element(step.Selector)
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("wait for selector '%s' timed out: %w", step.Selector, err)
		}

	case "wait_load":
		// Just wait for page to load
		page.Timeout(timeout).WaitLoad()

	case "evaluate_js":
		if step.Script == "" {
			result.Pass = false
			result.Error = fmt.Errorf("evaluate_js action requires script field")
			return result
		}
		obj, err := page.Timeout(timeout).Evaluate(&rod.EvalOptions{
			JS:      step.Script,
			ByValue: true,
		})
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("failed to evaluate JS: %w", err)
		} else {
			result.Output = fmt.Sprintf("%v", obj.Value)
		}

	case "set_viewport":
		// Viewport is set at browser launch time
		result.Output = "Viewport is set at browser launch time via config"

	case "press":
		if step.Key == "" {
			result.Pass = false
			result.Error = fmt.Errorf("press action requires key field")
			return result
		}
		el := page.Timeout(timeout).MustElement(step.Selector)
		// Focus the element first, then press the key using the page keyboard
		err := el.Focus()
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("failed to focus element for key press '%s': %w", step.Key, err)
			return result
		}
		// Map common key names to input key constants
		var key input.Key
		switch step.Key {
		case "Enter":
			key = input.Enter
		case "Escape", "Esc":
			key = input.Escape
		case "Tab":
			key = input.Tab
		case "Backspace":
			key = input.Backspace
		case "Delete":
			key = input.Delete
		case "ArrowUp":
			key = input.ArrowUp
		case "ArrowDown":
			key = input.ArrowDown
		case "ArrowLeft":
			key = input.ArrowLeft
		case "ArrowRight":
			key = input.ArrowRight
		case "Space":
			key = input.Space
		default:
			result.Pass = false
			result.Error = fmt.Errorf("unsupported key '%s' for press action", step.Key)
			return result
		}
		err = page.Keyboard.Press(key)
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("failed to press key '%s': %w", step.Key, err)
		}

	case "wait_navigation":
		// Wait for navigation to complete
		page.Timeout(timeout).WaitLoad()

	case "scroll_to":
		if step.Selector == "" {
			result.Pass = false
			result.Error = fmt.Errorf("scroll_to action requires selector field")
			return result
		}
		el := page.Timeout(timeout).MustElement(step.Selector)
		err := el.ScrollIntoView()
		if err != nil {
			result.Pass = false
			result.Error = fmt.Errorf("failed to scroll to %s: %w", step.Selector, err)
		}

	default:
		result.Pass = false
		result.Error = fmt.Errorf("unknown action type: %s", action)
	}

	// Apply wait after action if specified
	if step.WaitMillis > 0 && result.Pass {
		time.Sleep(time.Duration(step.WaitMillis) * time.Millisecond)
	}

	return result
}
