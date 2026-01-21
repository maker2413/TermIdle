#!/bin/bash

# The Ralph Wiggum Loop for OpenCode
# Iterates to make the Implementation Plan.

while :; do
    echo "--- Starting new Ralph iteration ---"
    
    # -p: Run a single prompt in non-interactive mode
    # -q: Quiet mode (removes the spinner for cleaner logs)
    # The prompt is piped from your PROMPT.md file
    opencode run "$(cat PLAN_PROMPT.md)" -m opencode/big-pickle
    
    # Check if the JTBD.md still has unchecked tasks
    # If no unchecked [ ] boxes remain, the project is complete.
    if ! grep -q "\[ \]" JTBD.md; then
        echo "Mission accomplished. All the plans are made."
        break
    fi

    # Optional: Safety sleep to prevent API rate limiting
    sleep 2
done
