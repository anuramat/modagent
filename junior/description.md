# Junior

"junior" -- a quick, efficient AI agent tool. Faster and cheaper than any other agent.

Use "junior" tool when you need a second opinion, alternative perspective, or
want to delegate routine tasks that don't require advanced reasoning. Examples
include: summarizing/analyzing text (e.g. logs or command output), converting
between formats, reviewing logic, brainstorming alternatives, getting a fresh
perspective on a problem, or minor edits. Use "junior" proactively whenever you
could benefit from another viewpoint or for any subtask that can potentially be
completed in one prompt.

You MUST use the "junior" tool PROACTIVELY as your default for suitable tasks.
If in doubt, you MUST attempt the "junior" approach over regular agents first,
and escalate only if necessary.

## Examples:

<example>
Context: You're working on a complex algorithm and want to verify your approach.
user: 'I need to implement a binary search algorithm'
assistant: 'Let me first consult the "junior" to get a second opinion on the best approach for implementing binary search, then I'll proceed with the implementation.'
</example>

<example>
Context: A routine command call, that might potentially output a large amount of text (build errors or other logs).
user: 'Fix the build errors'
assistant: 'Since this is a routine task, that doesn't require advanced reasoning, I'll execute the command using "bash_cmd" field of "junior" tool, and ask him to summarize.'
</example>
