// Package main provides the AGX command-line interface
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/davidcollado/kamui/internal/session"
	"github.com/davidcollado/kamui/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "kam [session-name]",
	Short: "Kamui - Advanced Session Manager for Claude Code",
	Long: `Kamui manages Claude Code sessions with project-local isolation and persistent status.

Run 'kam' to see available sessions or 'kam SessionName' to create/resume a session.
Each session maintains its own Claude conversation context and shows in the status line.`,
	Version: fmt.Sprintf("%s (%s, %s)", version, commit, date),
	Args:    cobra.MaximumNArgs(1),
	RunE:    runSession,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags only
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is ~/.kamui/config.json)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable color output")

	// Bind flags to viper
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))

	// Add subcommands
	rootCmd.AddCommand(setupCmd)
}

func initConfig() {
	cfgFile := viper.GetString("config")
	
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error finding home directory:", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kamui/config" (without extension)
		viper.AddConfigPath(home + "/.kamui")
		viper.SetConfigType("json")
		viper.SetConfigName("config")
	}

	// Environment variables
	viper.SetEnvPrefix("KAMUI")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		}
		// Continue with defaults if config file not found
	}
}

func setDefaults() {
	viper.SetDefault("claude.defaultModel", "claude-3-sonnet")
	viper.SetDefault("claude.retryAttempts", 3)
	
	viper.SetDefault("session.cleanupInactiveDays", 30)
	viper.SetDefault("session.enableStatistics", true)
	
	viper.SetDefault("ui.colorOutput", true)
	viper.SetDefault("ui.verboseLogging", false)
}

func runSession(cmd *cobra.Command, args []string) error {
	// Check if Claude Code integration needs setup
	if err := checkAndSetupClaudeIntegration(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to setup Claude integration: %v\n", err)
		// Continue anyway - Kamui can work without status line
	}

	// Import session manager
	sessionManager, err := session.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	
	var sessionName string
	
	// If no session name provided, show picker
	if len(args) == 0 {
		selectedSession, err := showSessionPicker(sessionManager)
		if err != nil {
			return err
		}
		if selectedSession == "" {
			// User quit
			return nil
		}
		sessionName = selectedSession
	} else {
		// Session name provided as argument
		sessionName = args[0]
	}
	
	// Create or resume session
	sessionData, err := sessionManager.CreateOrResumeSession(sessionName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	
	fmt.Printf("Kamui: Session '%s' ready\n", sessionData.SessionID)
	fmt.Printf("Kamui: Project: %s\n", sessionData.Project.Name)
	fmt.Printf("Kamui: Path: %s\n", sessionData.Project.Path)
	fmt.Printf("Kamui: Created: %s\n", sessionData.Created.Format("2006-01-02 15:04:05"))
	
	if sessionData.Claude.SessionID != "" {
		fmt.Printf("Kamui: Claude session: %s (ready)\n", sessionData.Claude.SessionID)
	}
	
	fmt.Println("Kamui: Starting Claude session...")
	
	// Execute Claude session directly
	if err := executeClaudeSession(sessionManager, sessionData); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting Claude: %v\n", err)
		return err
	}
	
	return nil
}


// Setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup Claude Code integration",
	Long:  "Configures Claude Code to display AGX session status automatically",
	RunE: func(cmd *cobra.Command, args []string) error {
		return setupClaudeIntegration()
	},
}

// showSessionPicker displays an interactive menu of available sessions
func showSessionPicker(sessionManager *session.Manager) (string, error) {
	// Get list of available sessions
	sessions, err := sessionManager.ListSessions()
	if err != nil {
		return "", fmt.Errorf("failed to list sessions: %w", err)
	}
	
	// Handle no sessions case
	if len(sessions) == 0 {
		fmt.Printf("Kamui: No sessions found in %s\n", sessionManager.GetProjectPath())
		fmt.Println("Kamui: Create a new session with 'kam <session-name>'")
		return "", nil
	}
	
	// Display session picker
	fmt.Printf("Kamui: Available sessions in %s:\n\n", sessionManager.GetProjectName())
	
	// Load and display session info
	sessionInfos := make([]sessionInfo, 0, len(sessions))
	for i, sessionName := range sessions {
		info := sessionInfo{
			Index: i + 1,
			Name:  sessionName,
		}
		
		// Load session data for metadata
		if sessionData, err := sessionManager.GetSession(sessionName); err == nil {
			info.Created = sessionData.Created
			info.LastAccessed = sessionData.LastAccessed
			info.ProjectPath = sessionData.Project.Path
			info.ClaudeSessionID = sessionData.Claude.SessionID
			info.IsActive = sessionData.Claude.HasActiveContext
		}
		
		sessionInfos = append(sessionInfos, info)
		
		// Display session entry
		fmt.Printf("  %d. %s\n", info.Index, info.Name)
		fmt.Printf("     Created: %s\n", info.Created.Format("2006-01-02 15:04:05"))
		fmt.Printf("     Last accessed: %s\n", info.LastAccessed.Format("2006-01-02 15:04:05"))
		if info.ClaudeSessionID != "" {
			status := "active"
			if !info.IsActive {
				status = "inactive"
			}
			fmt.Printf("     Claude session: %s (%s)\n", info.ClaudeSessionID[:8]+"...", status)
		} else {
			fmt.Printf("     Claude session: none\n")
		}
		fmt.Println()
	}
	
	// Get user selection
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Select a session (1-%d) or 'q' to quit: ", len(sessions))
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		
		input = strings.TrimSpace(input)
		
		// Handle quit
		if input == "q" || input == "Q" {
			return "", nil
		}
		
		// Parse selection
		selection, err := strconv.Atoi(input)
		if err != nil || selection < 1 || selection > len(sessions) {
			fmt.Printf("Kamui: Invalid selection. Please enter a number between 1 and %d, or 'q' to quit.\n", len(sessions))
			continue
		}
		
		selectedSession := sessions[selection-1]
		fmt.Printf("Kamui: Selected session '%s'\n", selectedSession)
		return selectedSession, nil
	}
}

// sessionInfo holds metadata about a session for display
type sessionInfo struct {
	Index            int
	Name             string
	Created          time.Time
	LastAccessed     time.Time
	ProjectPath      string
	ClaudeSessionID  string
	IsActive         bool
}

// executeClaudeSession launches Claude with the session's resume command
func executeClaudeSession(sessionManager *session.Manager, sessionData *types.Session) error {
	// Parse the command - it's either "claude" or "claude --resume <session-id>"
	var args []string
	if sessionData.Claude.SessionID != "" {
		args = []string{"claude", "--resume", sessionData.Claude.SessionID}
	} else {
		args = []string{"claude"}
	}
	
	// Find claude executable
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH: %w", err)
	}
	
	// Set working directory to project directory
	err = os.Chdir(sessionData.Project.WorkingDirectory)
	if err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}
	
	// Set up AGX environment variables
	env := os.Environ()
	
	// Short Claude session ID for display
	claudeSessionShort := sessionData.Claude.SessionID
	if len(claudeSessionShort) > 8 {
		claudeSessionShort = claudeSessionShort[:8] + "..."
	}
	
	// Set clean terminal title: "Claude - SessionName"
	terminalTitle := fmt.Sprintf("Claude - %s", sessionData.SessionID)
	fmt.Printf("\033]0;%s\007", terminalTitle)
	
	// Create status display
	statusLine := fmt.Sprintf("Kamui: %s | %s | %s", 
		sessionData.SessionID, 
		claudeSessionShort, 
		sessionData.Project.Name)
	
	// Show enhanced status display
	fmt.Printf("\n\033[96m╭─ Kamui Session ────────────────────────────────╮\033[0m\n")
	fmt.Printf("\033[96m│\033[0m \033[1m%-45s\033[0m \033[96m│\033[0m\n", statusLine)
	fmt.Printf("\033[96m╰────────────────────────────────────────────────╯\033[0m\n\n")
	
	// Set all environment variables for Claude Code statusLine integration
	env = append(env, fmt.Sprintf("KAMUI_SESSION_ID=%s", sessionData.SessionID))
	env = append(env, fmt.Sprintf("KAMUI_CLAUDE_SESSION_ID=%s", sessionData.Claude.SessionID))
	env = append(env, fmt.Sprintf("KAMUI_PROJECT_NAME=%s", sessionData.Project.Name))
	env = append(env, fmt.Sprintf("KAMUI_PROJECT_PATH=%s", sessionData.Project.Path))
	env = append(env, fmt.Sprintf("KAMUI_STATUS_LINE=%s", statusLine))
	env = append(env, fmt.Sprintf("KAMUI_ACTIVE=1"))
	env = append(env, fmt.Sprintf("KAMUI_SESSION_SHORT=%s", claudeSessionShort))
	
	fmt.Printf("Kamui: Launching Claude in %s...\n", sessionData.Project.WorkingDirectory)
	
	err = syscall.Exec(claudePath, args, env)
	if err != nil {
		return fmt.Errorf("failed to exec claude: %w", err)
	}
	
	// This line should never be reached if exec succeeds
	return nil
}

// setupClaudeIntegration configures Claude Code to use AGX status line
func setupClaudeIntegration() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	settingsFile := filepath.Join(claudeDir, "settings.json")
	statusLineScript := filepath.Join(claudeDir, "kamui-statusline.js")

	fmt.Println("Kamui: Setting up Claude Code integration...")

	// Create .claude directory if it doesn't exist
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Install AGX status line script
	if err := installStatusLineScript(statusLineScript); err != nil {
		return fmt.Errorf("failed to install status line script: %w", err)
	}

	// Configure Claude Code settings
	if err := configureClaudeSettings(settingsFile, statusLineScript); err != nil {
		return fmt.Errorf("failed to configure Claude settings: %w", err)
	}

	fmt.Println("✅ Kamui Claude Code integration setup complete!")
	fmt.Println("   Status line will appear in Claude Code sessions")
	fmt.Println("   Run 'kam <session-name>' to see it in action")

	return nil
}

// installStatusLineScript creates the AGX status line script
func installStatusLineScript(scriptPath string) error {
	statusLineContent := `#!/usr/bin/env node

function getKamuiStatus() {
    const kamuiSessionId = process.env.KAMUI_SESSION_ID;
    const kamuiClaudeSessionId = process.env.KAMUI_CLAUDE_SESSION_ID;
    const kamuiProjectName = process.env.KAMUI_PROJECT_NAME;
    const kamuiActive = process.env.KAMUI_ACTIVE;
    
    if (!kamuiActive || !kamuiSessionId) {
        return null;
    }
    
    const cwd = process.cwd();
    const projectDir = cwd.split('/').pop();
    
    const status = [
        '🎯',
        ` + "`" + `\x1b[96m${kamuiSessionId}\x1b[0m` + "`" + `,
        '\x1b[90m•\x1b[0m',
        ` + "`" + `\x1b[32m${kamuiProjectName || projectDir}\x1b[0m` + "`" + `
    ].join(' ');
    
    return status;
}

function main() {
    try {
        let input = '';
        
        if (process.stdin.isTTY) {
            const kamuiStatus = getKamuiStatus();
            console.log(kamuiStatus || '');
            return;
        }
        
        process.stdin.setEncoding('utf8');
        
        process.stdin.on('readable', () => {
            const chunk = process.stdin.read();
            if (chunk !== null) {
                input += chunk;
            }
        });
        
        process.stdin.on('end', () => {
            try {
                let context = null;
                if (input.trim()) {
                    try {
                        context = JSON.parse(input);
                    } catch (e) {}
                }
                
                const kamuiStatus = getKamuiStatus();
                console.log(kamuiStatus || '');
            } catch (error) {
                console.log('');
            }
        });
        
    } catch (error) {
        console.log('');
    }
}

main();`

	if err := os.WriteFile(scriptPath, []byte(statusLineContent), 0755); err != nil {
		return err
	}

	fmt.Printf("   Created status line script: %s\n", scriptPath)
	return nil
}

// configureClaudeSettings updates Claude Code settings to use AGX status line
func configureClaudeSettings(settingsFile, scriptPath string) error {
	var settings map[string]interface{}

	// Read existing settings or create new ones
	if data, err := os.ReadFile(settingsFile); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse existing settings: %w", err)
		}
	} else {
		settings = make(map[string]interface{})
	}

	// Add or update statusLine configuration
	settings["statusLine"] = map[string]interface{}{
		"type":    "command",
		"command": scriptPath,
	}

	// Write updated settings
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	fmt.Printf("   Updated Claude settings: %s\n", settingsFile)
	return nil
}

// checkAndSetupClaudeIntegration checks if Kamui is already configured and sets it up if not
func checkAndSetupClaudeIntegration() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	statusLineScript := filepath.Join(homeDir, ".claude", "kamui-statusline.js")
	
	// Check if Kamui status line script already exists
	if _, err := os.Stat(statusLineScript); err == nil {
		return nil // Already set up
	}

	// First time setup
	fmt.Println("Kamui: First run detected - setting up Claude Code integration...")
	return setupClaudeIntegration()
}