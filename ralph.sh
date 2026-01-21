#!/bin/bash

# The Ralph Wiggum Loop for OpenCode
# Iterates until the Implementation Plan is finished or a task fails.

while :; do
    echo "--- Starting new Ralph iteration ---"
    
    # -p: Run a single prompt in non-interactive mode
    # -q: Quiet mode (removes the spinner for cleaner logs)
    # The prompt is piped from your PROMPT.md file
    opencode run "$(cat PROMPT.md)" -m opencode/big-pickle
    
    # Check if the IMPLEMENTATION_PLAN.md still has unchecked tasks
    # If no unchecked [ ] boxes remain, the project is complete.
    if ! grep -q "\[ \]" IMPLEMENTATION_PLAN.md; then
        echo "Mission accomplished. All tasks in the plan are marked complete."
        break
    fi

    # Optional: Safety sleep to prevent API rate limiting
    sleep 2
done
