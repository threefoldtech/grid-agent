<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fade, scale } from "svelte/transition";

  const dispatch = createEventDispatcher();
  export let show = false;

  function close() {
    dispatch("close");
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      close();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if show}
  <div class="overlay" transition:fade={{ duration: 200 }}>
    <div
      class="docs-container"
      transition:scale={{ duration: 200, start: 0.95 }}
    >
      <header class="docs-header">
        <h2>User Guide</h2>
        <p>Learn how to use Grid Agent effectively</p>
        <button class="close-btn" on:click={close}>
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </header>

      <main class="docs-content">
        <!-- Getting Started -->
        <div class="docs-section">
          <h4>🚀 Getting Started</h4>
          <p><strong>Grid Agent</strong> is an AI-powered interface for managing ThreeFold Grid resources. It combines natural language processing with direct Grid operations.</p>

          <h5>Initial Setup</h5>
          <ul>
            <li><strong>Mnemonic Phrase:</strong> Your ThreeFold Grid wallet credentials</li>
            <li><strong>Network:</strong> Choose mainnet for production, testnet for testing</li>
            <li><strong>API Key:</strong> Get a free Gemini API key from Google AI Studio</li>
          </ul>
        </div>

        <!-- Using the Chat Interface -->
        <div class="docs-section">
          <h4>💬 Using the Chat Interface</h4>
          <p>Interact with your Grid resources using natural language commands.</p>

          <h5>Examples</h5>
          <div class="code-example">
            <p>"Deploy a VM with 2 CPUs and 4GB RAM"</p>
            <p>"Show me all my active contracts"</p>
            <p>"Deploy a wordpress instance"</p>
          </div>

          <h5>Understanding Responses</h5>
          <ul>
            <li><strong>Steps:</strong> Each action is broken down into executable steps</li>
            <li><strong>Commands:</strong> Actual CLI commands that will be executed</li>
            <li><strong>Output:</strong> Real-time execution results</li>
            <li><strong>Export:</strong> Save conversations for documentation</li>
          </ul>
        </div>

        <!-- Managing Personas -->
        <div class="docs-section">
          <h4>🎭 Managing Personas</h4>
          <p>Personas customize how the AI agent responds and behaves for different tasks.</p>

          <h5>How to Create and Use Personas</h5>
          <ol>
            <li>Click <strong>"Personas"</strong> in the sidebar</li>
            <li>Click the <strong>"+"</strong> button to create a new persona</li>
            <li>Give it a descriptive name (e.g., "DevOps Engineer")</li>
            <li>Add specific instructions for how the AI should behave</li>
            <li>Click the activate button on the persona card to activate it</li>
          </ol>

          <h5>What are Personas?</h5>
          <p>Personas are your <strong>custom preferences and recurring instructions</strong>. They let you define behaviors that apply to every conversation when activated.</p>

          <h5>Example Use Cases</h5>
          <ul>
            <li><strong>Deployment Shortcuts:</strong> "When I say 'deploy dev env', use docker with 4GB RAM, 2 CPUs, IPv4 access, and my SSH key from ~/.ssh/id_rsa"</li>
            <li><strong>Cost Monitoring:</strong> "At the start of each conversation, check my twin's last 24h consumption in USD and alert me if it exceeds $20"</li>
            <li><strong>Default Preferences:</strong> "Always use nodes in germany for my deployments. prefer farm 1"</li>
            <li><strong>Output Format:</strong> "When showing contract info, always include the last bill cost in USD"</li>
          </ul>

          <h5>Example Persona</h5>
          <div class="code-example">
            <p><strong>Name:</strong> TurboNerd</p>
            <p><strong>Instructions:</strong> When I ask to deploy a dev environment, create a VM with 4GB RAM, 2 CPUs, IPv4 access, and add my SSH key from ~/.ssh/id_rsa.pub. At the start of conversations, check my last 24h billing and warn me if it exceeds $20.</p>
          </div>
        </div>

        <!-- AI Configuration -->
        <div class="docs-section">
          <h4>🤖 AI Configuration</h4>
          <p>Configure the AI model and behavior settings.</p>

          <h5>Provider Options</h5>
          <ul>
            <li><strong>Google Gemini:</strong> Cloud-based AI with API key (default)</li>
            <li><strong>Ollama (Local):</strong> Run AI models locally for privacy and no API costs</li>
          </ul>

          <h5>How to Configure AI Settings</h5>
          <ol>
            <li>Click <strong>"AI Configuration"</strong> in the sidebar</li>
            <li><strong>Provider Selection:</strong> Choose between Google Gemini or Ollama</li>
            <li><strong>Model Selection:</strong> Pick from available models for your chosen provider</li>
            <li><strong>API Key:</strong> Required for Gemini, not needed for Ollama</li>
            <li><strong>Export Summary:</strong> Enable AI-generated summaries when exporting conversations</li>
          </ol>

          <h5>Gemini Models</h5>
          <ul>
            <li><strong>gemini-3-flash-preview:</strong> Latest and fastest (recommended)</li>
            <li><strong>gemini-3-pro-preview:</strong> Most capable Gemini 3 model</li>
            <li><strong>gemini-2.5-pro:</strong> Good balance of capability and speed</li>
            <li><strong>gemini-2.5-flash:</strong> Fast responses, lower token usage</li>
            <li><strong>gemini-2.5-flash-lite:</strong> Lightest model, minimal tokens</li>
          </ul>

          <h5>Export Options</h5>
          <ul>
            <li><strong>With Summary:</strong> AI generates a summary of your conversation</li>
            <li><strong>Without Summary:</strong> Saves tokens, just exports the raw conversation</li>
          </ul>
        </div>

        <!-- Ollama Local AI Setup -->
        <div class="docs-section">
          <h4>🏠 Ollama Local AI Setup</h4>
          <p>Run AI models locally on your machine for maximum privacy and zero API costs.</p>

          <h5>Why Use Ollama?</h5>
          <ul>
            <li><strong>Complete Privacy:</strong> Your conversations never leave your machine</li>
            <li><strong>No API Costs:</strong> Unlimited usage with no monthly fees</li>
            <li><strong>Offline Capable:</strong> Works without internet connection</li>
            <li><strong>High Performance:</strong> Instant responses with local processing</li>
          </ul>

          <h5>Step 1: Install Ollama</h5>
          <div class="code-example">
            <p><strong>For macOS:</strong></p>
            <p>brew install ollama</p>
            <br>
            <p><strong>For Linux:</strong></p>
            <p>curl -fsSL https://ollama.ai/install.sh | sh</p>
            <br>
            <p><strong>For Windows:</strong></p>
            <p>Download from <a href="https://ollama.ai/download" target="_blank" rel="noopener noreferrer">ollama.ai/download</a></p>
          </div>

          <h5>Step 2: Quick Setup (Choose One)</h5>

          <h6>Native Installation (Recommended)</h6>
          <div class="code-example">
            <p># 1. Start Ollama in background</p>
            <p>ollama serve &</p>
            <br>
            <p># 2. Verify it's running</p>
            <p>curl http://localhost:11434/api/version</p>
            <br>
            <p># 3. Pull a model</p>
            <p>ollama pull llama3.1:8b</p>
            <br>
            <p># 4. Check models</p>
            <p>ollama list</p>
            <br>
            <p># 5. Use Grid Agent!</p>
            <p># (configure to use Ollama)</p>
            <br>
            <p># 6. Stop when done</p>
            <p>pkill -f "ollama serve"</p>
          </div>

          <h6>Docker Installation</h6>
          <div class="code-example">
            <p># 1. Start Ollama container</p>
            <p>docker run -d -v ollama:/root/.ollama -p 11434:11434 --name ollama ollama/ollama</p>
            <br>
            <p># 2. Verify it's running</p>
            <p>curl http://localhost:11434/api/version</p>
            <br>
            <p># 3. Pull a model</p>
            <p>docker exec -it ollama ollama pull llama3.1:8b</p>
            <br>
            <p># 4. Check models</p>
            <p>docker exec -it ollama ollama list</p>
            <br>
            <p># 5. Use Grid Agent!</p>
            <p># (configure to use Ollama)</p>
            <br>
            <p># 6. Stop when done</p>
            <p>docker stop ollama</p>
          </div>

          <h5>Step 4: Configure Grid Agent</h5>
          <ol>
            <li>Open Grid Agent and go to <strong>"AI Configuration"</strong></li>
            <li>Select <strong>"Ollama (Local)"</strong> as the provider</li>
            <li>Choose your downloaded model from the dropdown</li>
            <li>Click <strong>"Save Configuration"</strong></li>
          </ol>

          <h5>Troubleshooting</h5>
          <ul>
            <li><strong>"Cannot connect to Ollama":</strong> Make sure Ollama is running with <code>ollama serve</code></li>
            <li><strong>"Model not found":</strong> Pull the model first with <code>ollama pull modelname</code></li>
            <li><strong>Slow responses:</strong> Larger models take more resources but give better results</li>
            <li><strong>Out of memory:</strong> Try smaller models like <code>llama3.1:8b</code> or <code>qwen2.5:7b</code></li>
          </ul>

          <h5>Model Recommendations</h5>
          <ul>
            <li><strong>llama3.1:8b:</strong> Best balance of capability and resource usage</li>
            <li><strong>qwen2.5:7b:</strong> Excellent reasoning and tool calling</li>
            <li><strong>mistral:7b:</strong> Fast and reliable for most tasks</li>
            <li><strong>llama3.1:70b:</strong> Maximum capability (requires more RAM)</li>
          </ul>

          <h5>System Requirements</h5>
          <ul>
            <li><strong>RAM:</strong> Minimum 8GB, recommended 16GB+</li>
            <li><strong>Storage:</strong> 5-40GB free space depending on model size</li>
            <li><strong>CPU/GPU:</strong> Modern processor (GPU recommended for better performance)</li>
          </ul>
        </div>

        <!-- Grid Configuration -->
        <div class="docs-section">
          <h4>🌐 Grid Configuration</h4>
          <p>Set up your ThreeFold Grid connection and credentials.</p>

          <h5>How to Configure Grid Settings</h5>
          <ol>
            <li>Click <strong>"Grid Configuration"</strong> in the sidebar</li>
            <li><strong>Network:</strong> Choose mainnet for production, testnet for testing</li>
            <li><strong>Mnemonic:</strong> Click "Show" to reveal your wallet phrase</li>
            <li>Click <strong>"Save Configuration"</strong> when done</li>
          </ol>

          <h5>Network Options</h5>
          <ul>
            <li><strong>Mainnet:</strong> Production network with real TFT tokens</li>
            <li><strong>Testnet:</strong> Testing network with discounted costs</li>
            <li><strong>Devnet:</strong> Development network</li>
            <li><strong>QA:</strong> Quality assurance network</li>
          </ul>

          <h5>Mnemonic Security</h5>
          <ul>
            <li>Your mnemonic is stored locally</li>
            <li>Never share your mnemonic with anyone</li>
            <li>Keep backups of your mnemonic in a secure location</li>
          </ul>
        </div>

        <!-- Saving Conversations -->
        <div class="docs-section">
          <h4>💾 Saving Conversations</h4>
          <p>Export your AI conversations for documentation and reference.</p>

          <h5>How to Save Discussions</h5>
          <ol>
            <li>Look for the <strong>download button</strong> (📥) in the top-right corner of the main screen</li>
            <li>Click it to open the save dialog</li>
            <li>Choose a filename (defaults to current date/time)</li>
            <li>Select save location</li>
            <li>Conversation saves as a Markdown file with full details</li>
          </ol>

          <h5>What Gets Saved</h5>
          <ul>
            <li>All messages in the conversation</li>
            <li>AI responses and reasoning steps</li>
            <li>Commands executed and their outputs</li>
            <li>Version information and active persona</li>
            <li>Optional AI-generated summary (if enabled)</li>
          </ul>
        </div>

        <!-- Grid Operations -->
        <div class="docs-section">
          <h4>🌐 Grid Operations</h4>
          <p>Understanding ThreeFold Grid concepts and operations.</p>

          <h5>Key Concepts</h5>
          <ul>
            <li><strong>Contracts:</strong> Agreements for resource usage</li>
            <li><strong>Nodes:</strong> Physical servers providing resources</li>
            <li><strong>Farms:</strong> Groups of nodes managed together</li>
            <li><strong>Flists:</strong> Container images for deployments</li>
          </ul>

          <h5>Common Operations</h5>
          <ul>
            <li><strong>VM Deployment:</strong> Virtual machines with custom specs</li>
            <li><strong>Kubernetes:</strong> Container orchestration clusters</li>
            <li><strong>Gateways:</strong> Load balancers and reverse proxies</li>
            <li><strong>ZDB:</strong> Distributed database storage</li>
          </ul>
        </div>

        <!-- Troubleshooting -->
        <div class="docs-section">
          <h4>🔧 Troubleshooting</h4>

          <h5>Common Issues</h5>
          <ul>
            <li><strong>"Agent not initialized":</strong> Check API key and network settings</li>
            <li><strong>"Command failed":</strong> Verify network connectivity and credentials</li>
            <li><strong>"No nodes available":</strong> Try different farm or adjust resources specs</li>
            <li><strong>"tfcmd not found":</strong> Grid CLI is missing - make sure to download tfcmd from <a href="https://github.com/threefoldtech/grid-agent/releases" target="_blank" rel="noopener noreferrer">github.com/threefoldtech/grid-agent/releases</a></li>
          </ul>

          <h5>Getting Help</h5>
          <ul>
            <li><strong>ThreeFold Forum:</strong> Community support at <a href="https://forum.threefold.io" target="_blank" rel="noopener noreferrer">forum.threefold.io</a></li>
            <li><strong>Grid Manual:</strong> Comprehensive docs at <a href="https://manual.grid.tf" target="_blank" rel="noopener noreferrer">manual.grid.tf</a></li>
            <li><strong>GitHub Issues:</strong> Report bugs or request features at <a href="https://github.com/threefoldtech/grid-agent" target="_blank" rel="noopener noreferrer">github.com/threefoldtech/grid-agent</a></li>
          </ul>
        </div>

        <!-- Tips -->
        <div class="docs-section">
          <h4>💡 Pro Tips</h4>
          <ul>
            <li>Use specific resource requirements for better deployment success</li>
            <li>Check contract billing regularly to monitor resource usage</li>
            <li>Export important conversations for documentation</li>
            <li>Use personas for specialized workflows</li>
            <li>Test on testnet before deploying to mainnet</li>
          </ul>
        </div>
      </main>
    </div>
  </div>
{/if}

<style>
  /* Overlay */
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(8px);
    z-index: 999;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem;
  }

  /* Main Container */
  .docs-container {
    width: 100%;
    max-width: 800px;
    height: 85vh;
    max-height: 800px;
    background: var(--bg-primary);
    border-radius: var(--radius-lg);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    border: 1px solid var(--border);
  }

  /* Header */
  .docs-header {
    padding: 2rem;
    border-bottom: 1px solid var(--border);
    position: relative;
  }

  .docs-header h2 {
    margin: 0 0 0.5rem 0;
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .docs-header p {
    margin: 0;
    font-size: 1rem;
    color: var(--text-secondary);
  }

  .close-btn {
    position: absolute;
    top: 1.5rem;
    right: 1.5rem;
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 0.5rem;
    border-radius: var(--radius-md);
    transition: all 0.2s;
  }

  .close-btn:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  /* Content */
  .docs-content {
    flex: 1;
    overflow-y: auto;
    padding: 2rem;
    line-height: 1.6;
    text-align: left;
  }

  .docs-section {
    margin-bottom: 3rem;
    padding-bottom: 2rem;
    border-bottom: 1px solid var(--border);
  }

  .docs-section:last-child {
    border-bottom: none;
    margin-bottom: 0;
    padding-bottom: 0;
  }

  .docs-section h4 {
    margin: 0 0 1rem 0;
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .docs-section h5 {
    margin: 1.5rem 0 0.75rem 0;
    font-size: 1rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .docs-section p {
    margin: 0 0 1rem 0;
    color: var(--text-primary);
  }

  .docs-section ul,
  .docs-section ol {
    margin: 0 0 1rem 0;
    padding-left: 1.5rem;
  }

  .docs-section li {
    margin-bottom: 0.5rem;
    color: var(--text-secondary);
  }

  .docs-section li strong {
    color: var(--text-primary);
    font-weight: 600;
  }

  .code-example {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: 1rem;
    margin: 1rem 0;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 0.9rem;
  }

  .code-example p {
    margin: 0.5rem 0;
    color: var(--accent);
    font-weight: 500;
  }

  /* Link styling for both light and dark themes */
  .docs-section a {
    color: #22d3ee; /* Cyan-400 - visible on both themes */
    text-decoration: none;
    font-weight: 500;
  }

  .docs-section a:hover {
    color: #67e8f9; /* Cyan-300 - lighter on hover */
    text-decoration: underline;
  }
</style>
