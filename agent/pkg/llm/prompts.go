package llm

// JSONFormatInstructions defines the mandatory JSON response format for the agent framework
const JSONFormatInstructions = `
### RESPONSE FORMAT STANDARDS
You must reply with a SINGLE or ARRAY of JSON objects.

For tool calls (when you need to use a tool):
{
  "toolName": "<tool_name>",
  "arguments": <string_or_array>,
  "explanation": "Why you need to use this tool..."
}
The available tools and their specific formats are described below.

For questions (when you need more information):
{
  "question": "What specific information do I need?",
  "explanation": "Why this information is required..."
}

For answers (final response to user):
{
  "answer": "Your complete response here...",
  "explanation": "Context or reasoning..."
}

CRITICAL: 
- Use the exact tool call format specified in each tool's description
- Never guess required parameters - ask if information is missing
- When reporting data, include the FULL output in the answer field, don't summarize
`
