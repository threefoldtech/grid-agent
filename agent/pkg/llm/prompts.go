package llm

// JSONFormatInstructions defines the mandatory JSON response format for the agent framework
const JSONFormatInstructions = `
### RESPONSE FORMAT STANDARDS
You must reply with a SINGLE or ARRAY of JSON objects.

The tools available to you will be described in the system prompt. Use the exact format specified for each tool.

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
