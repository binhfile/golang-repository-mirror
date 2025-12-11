# Todo List Tracking Guide

## Why Use Todo Lists?

When working on complex tasks in Claude Code, using a structured todo list helps you:
- **Stay organized** - Track multiple steps and dependencies
- **Maintain focus** - Know exactly what needs to be done next
- **Demonstrate progress** - Show clear completion status
- **Avoid missing steps** - Ensure all requirements are completed
- **Handle interruptions** - Quickly resume from where you left off
- **Communicate status** - Let users see real-time progress updates

## When to Create a Todo List

Create a todo list when:
- **Task has 3+ distinct steps** - Multi-phase projects benefit most
- **Complex dependencies** - Some tasks depend on others completing first
- **Multiple code changes** - Tracking changes across different files
- **Testing requirements** - Need to verify each component works
- **User-requested tracking** - User explicitly asks for progress tracking
- **Non-trivial implementation** - Task requires careful planning

## Don't use for:
- Single trivial tasks (< 3 steps)
- Quick bug fixes with obvious solutions
- Simple information requests
- Tasks that can complete in seconds

## Todo List Format

```json
{
  "content": "Task description in imperative form (e.g., 'Fix authentication bug')",
  "activeForm": "Present continuous form shown during work (e.g., 'Fixing authentication bug')",
  "status": "One of: pending, in_progress, completed"
}
```

### Status Meanings

- **pending** - Not yet started
- **in_progress** - Currently working on this task (only 1 at a time)
- **completed** - Task finished successfully

## Best Practices

### 1. Create Early
Create the todo list at the START of work, not after

```
❌ BAD: Work on task, then create todo
✅ GOOD: Create todo, then start work
```

### 2. Break Down Large Tasks
Split complex work into smaller, manageable items

```
❌ BAD: "Implement user authentication system"
✅ GOOD:
  - Set up database schema for users
  - Create login endpoint
  - Add password hashing
  - Create JWT token generation
  - Add token validation middleware
  - Write tests for authentication
```

### 3. Use Clear, Specific Names
Each todo should be a complete action item

```
❌ BAD: "Fix stuff", "More coding", "Testing"
✅ GOOD: "Fix null pointer exception in resolver",
          "Add validation for user input",
          "Write unit tests for packer module"
```

### 4. Maintain One Active Task
Only mark ONE task as `in_progress` at a time

```
❌ BAD: Multiple in_progress tasks
✅ GOOD: One in_progress, rest pending or completed
```

### 5. Complete Immediately After Finishing
Mark tasks as completed right after finishing

```
❌ BAD: Batch multiple completions at end
✅ GOOD: Mark complete as soon as each task finishes
```

### 6. Two-Form Requirement
Every todo must have both forms:

```
❌ BAD: Only one form
✅ GOOD:
  - content: "Fix authentication bug"
  - activeForm: "Fixing authentication bug"
```

The `content` is what the task IS, the `activeForm` is what you're DOING

## Real-World Examples

### Example 1: Feature Implementation

```json
[
  {
    "content": "Explore codebase to understand architecture",
    "activeForm": "Exploring codebase structure",
    "status": "pending"
  },
  {
    "content": "Design API endpoint structure",
    "activeForm": "Designing API endpoints",
    "status": "pending"
  },
  {
    "content": "Implement API handler functions",
    "activeForm": "Implementing API handlers",
    "status": "pending"
  },
  {
    "content": "Add input validation and error handling",
    "activeForm": "Adding validation and error handling",
    "status": "pending"
  },
  {
    "content": "Write unit tests for new endpoints",
    "activeForm": "Writing unit tests",
    "status": "pending"
  },
  {
    "content": "Test integration with database",
    "activeForm": "Testing database integration",
    "status": "pending"
  },
  {
    "content": "Document new API endpoints",
    "activeForm": "Documenting API endpoints",
    "status": "pending"
  }
]
```

### Example 2: Bug Fix with Investigation

```json
[
  {
    "content": "Reproduce the bug and gather error logs",
    "activeForm": "Reproducing bug",
    "status": "in_progress"
  },
  {
    "content": "Trace through code to find root cause",
    "activeForm": "Tracing code execution",
    "status": "pending"
  },
  {
    "content": "Implement fix for the issue",
    "activeForm": "Implementing fix",
    "status": "pending"
  },
  {
    "content": "Verify fix resolves the problem",
    "activeForm": "Verifying fix",
    "status": "pending"
  },
  {
    "content": "Check for regression in related features",
    "activeForm": "Checking for regressions",
    "status": "pending"
  }
]
```

### Example 3: Code Refactoring

```json
[
  {
    "content": "Identify code to refactor and understand dependencies",
    "activeForm": "Analyzing code structure",
    "status": "pending"
  },
  {
    "content": "Create backup tests to verify current behavior",
    "activeForm": "Creating baseline tests",
    "status": "pending"
  },
  {
    "content": "Refactor first module",
    "activeForm": "Refactoring module A",
    "status": "pending"
  },
  {
    "content": "Refactor second module",
    "activeForm": "Refactoring module B",
    "status": "pending"
  },
  {
    "content": "Update integration points between modules",
    "activeForm": "Updating integrations",
    "status": "pending"
  },
  {
    "content": "Run all tests to ensure no breakage",
    "activeForm": "Running test suite",
    "status": "pending"
  },
  {
    "content": "Update documentation for refactored code",
    "activeForm": "Updating documentation",
    "status": "pending"
  }
]
```

## Workflow Pattern

### Step 1: Create Todo List at Start
As soon as you understand the task, create a comprehensive todo list

```
User: "Please add user authentication to the app"
Assistant: "I'll help with that. Let me create a todo list..."
[Creates 6-8 specific todos]
```

### Step 2: Work on First Task
Mark the first task as `in_progress` and start work

```
[Updates first task to in_progress]
"Now starting step 1: Analyzing the authentication requirements..."
```

### Step 3: Complete and Move Forward
As soon as you finish a task, mark it complete and move to the next

```
[Marks task 1 as completed]
[Updates task 2 to in_progress]
"Task 1 complete. Moving to next task..."
```

### Step 4: Final Summary
When all tasks are done, show the completion

```
[All tasks marked completed]
"All tasks complete! Summary:
- 8 tasks completed
- 0 tasks failed
- Full implementation ready for use"
```

## Integration with Claude Code

The todo list works seamlessly with Claude Code:

```bash
# Run the task
claude-code work "Add authentication feature"

# Claude will:
1. Create a todo list at the start
2. Update it as work progresses
3. Show real-time status to user
4. Complete all items systematically
```

## Tips for Success

### ✅ Do This
- Create todos BEFORE starting work
- Use specific, measurable descriptions
- Break large tasks into small steps
- Mark tasks complete immediately
- Have only one task in_progress
- Maintain clear task dependencies
- Review progress periodically

### ❌ Avoid This
- Creating todos after starting work
- Vague task descriptions
- Too many pending tasks without structure
- Forgetting to mark tasks complete
- Multiple in_progress tasks
- Ambiguous task names
- Batching completions at the end

## Common Mistakes

### Mistake 1: Too Generic Tasks
```
❌ "Work on feature"
✅ "Create API endpoint for user registration"
```

### Mistake 2: Not Breaking Down Enough
```
❌ "Implement authentication" (single todo)
✅ "Implement authentication" broken into:
   - Create database schema
   - Implement login API
   - Add password hashing
   - Create JWT tokens
   - Add auth middleware
   - Write tests
```

### Mistake 3: Forgetting Active Form
```
❌ "content": "Fix bug"
✅ "content": "Fix null pointer bug",
   "activeForm": "Fixing null pointer bug"
```

### Mistake 4: Forgetting to Update Status
```
❌ Create todos but don't update status as you work
✅ Update status immediately after each task
```

## Real Example from This Project

When I noticed the missing github.com dependencies issue, I should have created a todo list like:

```json
[
  {
    "content": "Identify why github.com dependencies aren't being resolved",
    "activeForm": "Investigating missing dependencies",
    "status": "completed"
  },
  {
    "content": "Test go list and go mod download behavior",
    "activeForm": "Testing Go module commands",
    "status": "completed"
  },
  {
    "content": "Discover that go mod download all is required",
    "activeForm": "Researching Go module resolution",
    "status": "completed"
  },
  {
    "content": "Update resolver to use go mod download all",
    "activeForm": "Updating resolver code",
    "status": "completed"
  },
  {
    "content": "Test fix to verify all 39 modules are resolved",
    "activeForm": "Testing resolver fix",
    "status": "completed"
  },
  {
    "content": "Verify Athens storage has all modules with proper files",
    "activeForm": "Verifying Athens storage structure",
    "status": "completed"
  }
]
```

This would have made the investigation more organized and trackable.

## Conclusion

Todo lists are a simple but powerful tool for:
- Staying focused on complex tasks
- Making progress visible
- Avoiding lost work when interrupted
- Communicating status effectively
- Ensuring no steps are skipped

**Start using them from the first moment of any complex task!**
