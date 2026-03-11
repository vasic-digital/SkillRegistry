// Package agents provides the CLI agent registry for HelixAgent.
//
// This package manages the 48 supported CLI agents, providing configuration
// generation and validation capabilities powered by LLMsVerifier.
//
// # Supported Agents (48)
//
// Original Agents (18):
//   - OpenCode, Crush, HelixCode, Kiro, Aider
//   - ClaudeCode, Cline, CodenameGoose, DeepSeekCLI, Forge
//   - GeminiCLI, GPTEngineer, KiloCode, MistralCode, OllamaCode
//   - Plandex, QwenCode, AmazonQ
//
// Extended Agents (30):
//   - AgentDeck, Bridle, CheshireCat, ClaudePlugins, ClaudeSquad
//   - Codai, Codex, CodexSkills, Conduit, Emdash
//   - FauxPilot, GetShitDone, GitHubCopilotCLI, GitHubSpecKit, GitMCP
//   - GPTME, MobileAgent, MultiagentCoding, Nanocoder, Noi
//   - Octogen, OpenHands, PostgresMCP, Shai, SnowCLI
//   - TaskWeaver, UIUXProMax, VTCode, Warp, Continue
//
// # Agent Registry
//
// The registry manages agent metadata:
//
//	registry := agents.NewRegistry()
//
//	// List all agents
//	allAgents := registry.List()
//
//	// Get agent info
//	info, ok := registry.Get("codex")
//
//	// Check if agent is supported
//	if registry.IsSupported("openhands") {
//	    // Generate config
//	}
//
// # Configuration Generation
//
// Generate configurations for any supported agent:
//
//	config, err := registry.GenerateConfig("codex", options)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Write to file
//	if err := config.WriteToFile("/path/to/codex.json"); err != nil {
//	    log.Fatal(err)
//	}
//
// # Configuration Validation
//
// Validate existing agent configurations:
//
//	validator := agents.NewValidator(registry)
//	if err := validator.ValidateFile("codex", "/path/to/config.json"); err != nil {
//	    log.Fatal("Invalid config:", err)
//	}
//
// # CLI Commands
//
// HelixAgent supports agent configuration via CLI:
//
//	# List all agents
//	./bin/helixagent --list-agents
//
//	# Generate config for specific agent
//	./bin/helixagent --generate-agent-config=codex
//	./bin/helixagent --generate-agent-config=openhands --agent-config-output=~/openhands.toml
//
//	# Validate config
//	./bin/helixagent --validate-agent-config=codex:/path/to/codex.json
//
//	# Generate all configs
//	./bin/helixagent --generate-all-agents --all-agents-output-dir=~/agent-configs/
//
// # Agent Capabilities
//
// Each agent has defined capabilities:
//
//	type AgentCapabilities struct {
//	    SupportsStreaming    bool
//	    SupportsToolCalling  bool
//	    SupportsMultimodal   bool
//	    SupportedProtocols   []string
//	    MaxContextLength     int
//	}
//
// # LLMsVerifier Integration
//
// Configuration generation is powered by LLMsVerifier:
//
//	// Uses LLMsVerifier's pkg/cliagents/ for unified generation
//	generator := verifier.NewAgentConfigGenerator()
//	config := generator.Generate("codex", providerConfig)
//
// # Key Files
//
//   - registry.go: Agent registry and metadata
//   - config_generator.go: Configuration generation
//   - validator.go: Configuration validation
//   - capabilities.go: Agent capability definitions
//
// # Example: Generate Multiple Configs
//
//	registry := agents.NewRegistry()
//	outputDir := "/path/to/configs"
//
//	for _, agent := range registry.List() {
//	    config, err := registry.GenerateConfig(agent.Name, nil)
//	    if err != nil {
//	        log.Printf("Failed to generate %s: %v", agent.Name, err)
//	        continue
//	    }
//
//	    path := filepath.Join(outputDir, agent.Name+".json")
//	    if err := config.WriteToFile(path); err != nil {
//	        log.Printf("Failed to write %s: %v", agent.Name, err)
//	    }
//	}
package agents
