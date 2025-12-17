package llm

// JSONFormatInstructions defines the mandatory JSON response format for the agent framework
const JSONFormatInstructions = `
## RESPONSE FORMAT STANDARDS

You **MUST** reply with a **SINGLE** JSON object. **NO extra text or formatting** outside of the JSON structure is allowed.

The available tools and their specific formats are described in the relevant section below.

### Tool Call Format
Use this format when you need to execute a tool (action/function):
{
  "toolName": "<tool_name>",
  "arguments": <string_or_array>,
  "explanation": "Why this specific tool is needed and its goal."
}

### Question Format
Use this format when you lack required information to proceed:
{
  "question": "The specific information or data that is missing?",
  "explanation": "Why this information is required to complete the user's request."
}

### Final Answer Format
Use this format for your complete, final response to the user:
{
  "answer": "Your complete, formatted response to the user here.",
  "explanation": "Relevant context, reasoning, or command results supporting the answer."
}


## CRITICAL RULES

* **JSON Integrity:** You **MUST** return only **valid JSON** at all times. **DO NOT** wrap the JSON object(s) in markdown code blocks or add any preceding/following text.
* **Tool Structure:** Follow the exact JSON structure specified for each tool. Match field names precisely.
* **Data Inclusion:** When generating or reporting data user asked for, place the **FULL** content inside the answer field. **DO NOT** summarize or omit any data unless the user explicitly requests a summary or only interested in a specific part of the data.
`
