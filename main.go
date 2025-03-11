package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Profile represents a Chromium browser profile
type Profile struct {
	Name      string
	Proxy     string
	ProxyType string
	Flags     string
}

// ChromiumManager handles the application state
type ChromiumManager struct {
	profiles     map[string]Profile
	configFile   string
	chromePath   string
	profileDir   string
	currentView  string
	mainList     list.Model
	profileList  list.Model
	manageList   list.Model
	message      string
	selected     string
	profileName  string
	profileProxy string
	profileType  string
	profileFlags string
	err          error
}

// Helper styles for application UI
var (
	docStyle  = lipgloss.NewStyle().Margin(1, 2)
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	okStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)
)

// Create a new model
func initialModel() *ChromiumManager {
	cm := &ChromiumManager{
		profiles:    make(map[string]Profile),
		currentView: "main",
	}

	// Set paths
	homeDir, _ := os.UserHomeDir()
	cm.profileDir = filepath.Join(homeDir, ".chrome_profiles")
	cm.configFile = filepath.Join(cm.profileDir, "profiles.conf")

	// Find browser
	cm.chromePath = "/Applications/Chromium.app/Contents/MacOS/Chromium"
	if _, err := os.Stat(cm.chromePath); os.IsNotExist(err) {
		cm.chromePath = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	}

	// Create directories & load profiles
	os.MkdirAll(cm.profileDir, 0755)
	cm.loadProfiles()

	// Create main menu
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(3) // Increase item height for better visibility
	delegate.SetSpacing(1) // Add spacing between items
	
	items := []list.Item{
		item{title: "Launch Browser", desc: "Start with a profile"},
		item{title: "Manage Profiles", desc: "Add, edit or remove profiles"},
		item{title: "Clean Profile", desc: "Clear browsing data"},
		item{title: "Quit", desc: "Exit application"},
	}

	cm.mainList = list.New(items, delegate, 80, 24)
	cm.mainList.Title = "Chromium Profile Manager"
	cm.mainList.SetShowStatusBar(true)
	cm.mainList.SetFilteringEnabled(false)
	
	// Create management menu
	cm.updateManageList()

	return cm
}

// Load profiles from config file
func (cm *ChromiumManager) loadProfiles() {
	// Create default profile if needed
	if _, err := os.Stat(cm.configFile); os.IsNotExist(err) {
		defaults := []Profile{
			{Name: "default", Proxy: "none", ProxyType: "none", Flags: "--no-first-run --disable-features=RendererCodeIntegrity"},
			{Name: "clean", Proxy: "none", ProxyType: "none", Flags: "--no-first-run --disable-features=RendererCodeIntegrity,UseChromeOSDirectVideoDecoder --disable-gpu-driver-bug-workarounds --ignore-gpu-blacklist --disable-gpu-compositing --disable-infobars"},
		}
		
		var content string
		for _, p := range defaults {
			content += fmt.Sprintf("%s|%s|%s|%s\n", p.Name, p.Proxy, p.ProxyType, p.Flags)
		}
		
		ioutil.WriteFile(cm.configFile, []byte(content), 0644)
	}

	// Read profiles
	data, err := ioutil.ReadFile(cm.configFile)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			cm.profiles[parts[0]] = Profile{
				Name:      parts[0],
				Proxy:     parts[1],
				ProxyType: parts[2],
				Flags:     parts[3],
			}
		}
	}

	// Update profile list
	cm.updateProfileList()
}

// Update the profile list
func (cm *ChromiumManager) updateProfileList() {
	items := []list.Item{}
	for name := range cm.profiles {
		items = append(items, item{title: name, desc: ""})
	}

	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(2)
	delegate.SetSpacing(1)
	
	cm.profileList = list.New(items, delegate, 80, 24)
	cm.profileList.Title = "Select Profile"
	cm.profileList.SetShowStatusBar(true)
	cm.profileList.SetFilteringEnabled(false)
}

// Update the manage menu
func (cm *ChromiumManager) updateManageList() {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(2)
	delegate.SetSpacing(1)
	
	items := []list.Item{
		item{title: "Add New Profile", desc: "Create a new browser profile"},
		item{title: "Edit Profile", desc: "Modify an existing profile"},
		item{title: "Delete Profile", desc: "Remove a profile"},
	}

	cm.manageList = list.New(items, delegate, 80, 24)
	cm.manageList.Title = "Profile Management"
	cm.manageList.SetShowStatusBar(true)
	cm.manageList.SetFilteringEnabled(false)
}

// Save profiles to config file
func (cm *ChromiumManager) saveProfiles() {
	var content string
	for _, profile := range cm.profiles {
		content += fmt.Sprintf("%s|%s|%s|%s\n", 
			profile.Name, profile.Proxy, profile.ProxyType, profile.Flags)
	}
	
	ioutil.WriteFile(cm.configFile, []byte(content), 0644)
}

// Launch browser with profile
func (cm *ChromiumManager) launchBrowser(profileName string) string {
	profile, exists := cm.profiles[profileName]
	if !exists {
		return fmt.Sprintf("Profile '%s' not found", profileName)
	}

	// Create profile directory
	profilePath := filepath.Join(cm.profileDir, profile.Name)
	os.MkdirAll(profilePath, 0755)
	
	// Create Local State file for API key warnings
	prefsFile := filepath.Join(profilePath, "Local State")
	if _, err := os.Stat(prefsFile); os.IsNotExist(err) {
		prefsData := `{"browser":{"enabled_labs_experiments":["ignore-gpu-blocklist@1"]},"distribution":{"suppress_first_run_bubble":true,"suppress_api_keys_warning":true}}`
		ioutil.WriteFile(prefsFile, []byte(prefsData), 0644)
	}

	// Build command line with all arguments - use a different approach
	cmdArgs := []string{}
	
	// Add profile directory
	cmdArgs = append(cmdArgs, "--user-data-dir="+profilePath)
	
	// Force new window
	cmdArgs = append(cmdArgs, "--new-window")
	cmdArgs = append(cmdArgs, "about:blank") // Open a blank page to ensure window opens
	
	// Add proxy if specified
	if profile.Proxy != "none" {
		proxyFlag := "--proxy-server="
		if profile.ProxyType == "http" {
			proxyFlag += "http://"
		}
		proxyFlag += profile.Proxy
		cmdArgs = append(cmdArgs, proxyFlag)
	}
	
	// Add profile flags by splitting on spaces (proper handling)
	if profile.Flags != "" {
		for _, flag := range strings.Split(profile.Flags, " ") {
			if flag != "" {
				cmdArgs = append(cmdArgs, flag)
			}
		}
	}
	
	// Add standard suppression flags
	standardFlags := []string{
		// Logging and notification suppression
		"--disable-logging",
		"--disable-breakpad",
		"--disable-infobars",
		"--disable-notifications",
		"--no-default-browser-check",
		"--silent-launch",
		
		// GPU artifact suppression
		"--disable-gpu",
		"--disable-gpu-compositing",
		"--disable-gpu-sandbox",
		"--disable-gpu-driver-bug-workarounds",
		"--disable-features=UseChromeOSDirectVideoDecoder",
		"--disable-accelerated-2d-canvas",
		"--disable-accelerated-video-decode",
		"--disable-accelerated-video-encode",
		"--disable-webgl",
		"--disable-threaded-animation",
		"--disable-webgl-image-chromium",
		"--force-dark-mode",
	}
	
	for _, flag := range standardFlags {
		cmdArgs = append(cmdArgs, flag)
	}
	
	// Try different approaches for launching
	var err error
	
	// First attempt: standard exec approach
	cmd := exec.Command(cm.chromePath, cmdArgs...)
	err = cmd.Start()
	
	// If that fails, try the open command on macOS
	if err != nil {
		// Create a shell script in temp directory
		scriptPath := filepath.Join(os.TempDir(), "launch_chrome.sh")
		scriptContent := "#!/bin/bash\n" + cm.chromePath + " " + strings.Join(cmdArgs, " ") + " &\n"
		if err := ioutil.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			return fmt.Sprintf("Error creating launcher script: %s", err)
		}
		
		// Execute the script
		cmd = exec.Command("/bin/bash", scriptPath)
		if err = cmd.Start(); err != nil {
			// Last resort - use 'open' command on macOS
			openArgs := []string{cm.chromePath, "--args"}
			openArgs = append(openArgs, cmdArgs...)
			cmd = exec.Command("open", openArgs...)
			err = cmd.Start()
		}
	}
	
	if err != nil {
		return fmt.Sprintf("Error launching browser: %s", err)
	}
	
	return fmt.Sprintf("Launched with profile: %s", profile.Name)
}

// Item for lists
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// Init implements tea.Model
func (cm *ChromiumManager) Init() tea.Cmd {
	// Set initial size to show items
	if cm.mainList.Items() != nil {
		cm.mainList.SetSize(80, 20)
	}
	if cm.profileList.Items() != nil {
		cm.profileList.SetSize(80, 20)
	}
	if cm.manageList.Items() != nil {
		cm.manageList.SetSize(80, 20)
	}
	return nil
}

// Update implements tea.Model
func (cm *ChromiumManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update window sizes for all lists
		if cm.manageList.Items() != nil {
			cm.manageList.SetSize(msg.Width, msg.Height-6)
		}
		if cm.mainList.Items() != nil {
			cm.mainList.SetSize(msg.Width, msg.Height-6)
		}
		if cm.profileList.Items() != nil {
			cm.profileList.SetSize(msg.Width, msg.Height-6)
		}

	case tea.KeyMsg:
		// Global keys
		switch msg.Type {
		case tea.KeyCtrlC:
			return cm, tea.Quit
		case tea.KeyEsc:
			if cm.currentView != "main" {
				cm.currentView = "main"
				cm.message = ""
				return cm, nil
			}
		}

		// View-specific handling
		switch cm.currentView {
		case "main":
			if msg.Type == tea.KeyEnter {
				i, ok := cm.mainList.SelectedItem().(item)
				if ok {
					switch i.title {
					case "Launch Browser":
						cm.updateProfileList()
						cm.currentView = "select_profile"
					case "Manage Profiles":
						cm.updateManageList()
						cm.currentView = "manage"
					case "Clean Profile":
						cm.updateProfileList()
						cm.currentView = "select_clean"
					case "Quit":
						return cm, tea.Quit
					}
				}
			}
			cm.mainList, cmd = cm.mainList.Update(msg)
			return cm, cmd

		case "select_profile":
			if msg.Type == tea.KeyEnter {
				i, ok := cm.profileList.SelectedItem().(item)
				if ok {
					cm.message = cm.launchBrowser(i.title)
					cm.currentView = "main"
				}
			}
			cm.profileList, cmd = cm.profileList.Update(msg)
			return cm, cmd
			
		case "manage":
			if msg.Type == tea.KeyEnter {
				i, ok := cm.manageList.SelectedItem().(item)
				if ok {
					switch i.title {
					case "Add New Profile":
						cm.currentView = "add_profile"
						cm.profileName = ""
						cm.profileProxy = "none"
						cm.profileType = "none"
						cm.profileFlags = "--no-first-run --disable-features=RendererCodeIntegrity"
					case "Edit Profile":
						cm.updateProfileList()
						cm.currentView = "select_edit"
					case "Delete Profile":
						cm.updateProfileList()
						cm.currentView = "select_delete"
					}
				}
			}
			cm.manageList, cmd = cm.manageList.Update(msg)
			return cm, cmd
			
		case "select_edit":
			if msg.Type == tea.KeyEnter {
				i, ok := cm.profileList.SelectedItem().(item)
				if ok {
					profile := cm.profiles[i.title]
					cm.profileName = profile.Name
					cm.profileProxy = profile.Proxy
					cm.profileType = profile.ProxyType
					cm.profileFlags = profile.Flags
					cm.selected = i.title
					cm.currentView = "edit_profile"
				}
			}
			cm.profileList, cmd = cm.profileList.Update(msg)
			return cm, cmd
			
		case "select_delete":
			if msg.Type == tea.KeyEnter {
				i, ok := cm.profileList.SelectedItem().(item)
				if ok {
					cm.selected = i.title
					cm.currentView = "confirm_delete"
				}
			}
			cm.profileList, cmd = cm.profileList.Update(msg)
			return cm, cmd
			
		case "confirm_delete":
			switch msg.String() {
			case "y", "Y":
				delete(cm.profiles, cm.selected)
				cm.saveProfiles()
				cm.message = fmt.Sprintf("Profile '%s' deleted", cm.selected)
				cm.currentView = "main"
				return cm, nil
			case "n", "N":
				cm.currentView = "main"
				return cm, nil
			}
			
		case "select_clean":
			if msg.Type == tea.KeyEnter {
				i, ok := cm.profileList.SelectedItem().(item)
				if ok {
					profilePath := filepath.Join(cm.profileDir, i.title)
					if _, err := os.Stat(profilePath); os.IsNotExist(err) {
						cm.message = "Profile directory does not exist"
					} else {
						// Clean the entire profile directory
						files, err := ioutil.ReadDir(profilePath)
						if err != nil {
							cm.message = fmt.Sprintf("Error reading directory: %s", err)
						} else {
							// Remove all files in the directory
							for _, file := range files {
								filePath := filepath.Join(profilePath, file.Name())
								if err := os.RemoveAll(filePath); err != nil {
									cm.message = fmt.Sprintf("Error cleaning profile: %s", err)
									cm.currentView = "main"
									return cm, nil
								}
							}
							cm.message = fmt.Sprintf("Profile '%s' completely cleared and reset", i.title)
						}
					}
					cm.currentView = "main"
				}
			}
			cm.profileList, cmd = cm.profileList.Update(msg)
			return cm, cmd
			
		case "edit_profile", "add_profile":
			// Handle field editing with number keys
			switch msg.String() {
			case "1":
				cm.currentView = "edit_name"
				return cm, nil
			case "2":
				cm.currentView = "edit_proxy"
				return cm, nil
			case "3":
				cm.currentView = "edit_type"
				return cm, nil
			case "4":
				cm.currentView = "edit_flags"
				return cm, nil
			}
			
			if msg.Type == tea.KeyEnter {
				// Save the edited profile
				oldName := cm.selected
				
				// Check if name is provided
				if cm.profileName == "" {
					cm.message = "Profile name is required"
					return cm, nil
				}
				
				// Check if name already exists (if changed)
				if oldName != cm.profileName {
					if _, exists := cm.profiles[cm.profileName]; exists {
						cm.message = fmt.Sprintf("Profile '%s' already exists", cm.profileName)
						return cm, nil
					}
				}
				
				// Remove the old profile if name changed
				if oldName != cm.profileName {
					delete(cm.profiles, oldName)
				}
				
				// Add/update the profile
				cm.profiles[cm.profileName] = Profile{
					Name:      cm.profileName,
					Proxy:     cm.profileProxy,
					ProxyType: cm.profileType,
					Flags:     cm.profileFlags,
				}
				
				// Save profiles
				cm.saveProfiles()
				cm.message = fmt.Sprintf("Profile '%s' updated", cm.profileName)
				cm.currentView = "main"
				return cm, nil
			}
			
		// Text input views
		case "edit_name", "edit_proxy", "edit_type", "edit_flags":
			if msg.Type == tea.KeyEnter {
				// Return to the edit/add view
				if strings.HasPrefix(cm.currentView, "edit_") {
					if cm.selected != "" {
						cm.currentView = "edit_profile"
					} else {
						cm.currentView = "add_profile"
					}
				}
				return cm, nil
			}
			
			// Handle text input
			switch cm.currentView {
			case "edit_name":
				if msg.Type == tea.KeyBackspace && len(cm.profileName) > 0 {
					cm.profileName = cm.profileName[:len(cm.profileName)-1]
				} else if msg.Type == tea.KeyRunes {
					cm.profileName += msg.String()
				}
			case "edit_proxy":
				if msg.Type == tea.KeyBackspace && len(cm.profileProxy) > 0 {
					cm.profileProxy = cm.profileProxy[:len(cm.profileProxy)-1]
				} else if msg.Type == tea.KeyRunes {
					cm.profileProxy += msg.String()
				}
			case "edit_type":
				if msg.Type == tea.KeyBackspace && len(cm.profileType) > 0 {
					cm.profileType = cm.profileType[:len(cm.profileType)-1]
				} else if msg.Type == tea.KeyRunes {
					cm.profileType += msg.String()
				}
			case "edit_flags":
				if msg.Type == tea.KeyBackspace && len(cm.profileFlags) > 0 {
					cm.profileFlags = cm.profileFlags[:len(cm.profileFlags)-1]
				} else if msg.Type == tea.KeyRunes {
					cm.profileFlags += msg.String()
				}
			}
		}
	}

	return cm, nil
}

// View renders the current UI
func (cm *ChromiumManager) View() string {
	// Handle errors
	if cm.err != nil {
		return errStyle.Render(fmt.Sprintf("Error: %s", cm.err))
	}

	var s string

	// Render the appropriate view
	switch cm.currentView {
	case "main":
		s = cm.mainList.View()
		
	case "select_profile", "select_edit", "select_delete", "select_clean":
		s = cm.profileList.View()
		
	case "manage":
		s = cm.manageList.View()
		
	case "confirm_delete":
		s = fmt.Sprintf("Delete Profile\n\nAre you sure you want to delete profile '%s'? (y/n)", cm.selected)
		
	case "add_profile", "edit_profile":
		s = "Profile Editor\n\n"
		s += fmt.Sprintf("1. Name: %s\n", cm.profileName)
		s += fmt.Sprintf("2. Proxy: %s\n", cm.profileProxy)
		s += fmt.Sprintf("3. Proxy Type: %s\n", cm.profileType)
		s += fmt.Sprintf("4. Flags: %s\n\n", cm.profileFlags)
		s += "Press 1-4 to edit a field, Enter to save, Esc to cancel"
		
	case "edit_name":
		s = "Edit Profile Name\n\n"
		s += fmt.Sprintf("Name: %s\n\n", cm.profileName)
		s += "Press Enter when done, Esc to cancel"
		
	case "edit_proxy":
		s = "Edit Proxy Address\n\n"
		s += fmt.Sprintf("Proxy: %s\n\n", cm.profileProxy)
		s += "Enter 'none' for no proxy, or server address (e.g. 127.0.0.1:8080)"
		s += "\nPress Enter when done, Esc to cancel"
		
	case "edit_type":
		s = "Edit Proxy Type\n\n"
		s += fmt.Sprintf("Proxy Type: %s\n\n", cm.profileType)
		s += "Enter 'none', 'http', or 'socks5'"
		s += "\nPress Enter when done, Esc to cancel"
		
	case "edit_flags":
		s = "Edit Browser Flags\n\n"
		s += fmt.Sprintf("Flags: %s\n\n", cm.profileFlags)
		s += "Enter the browser command-line flags"
		s += "\nPress Enter when done, Esc to cancel"
		
	default:
		s = "Unknown view: " + cm.currentView
	}

	// Add any messages
	if cm.message != "" {
		if strings.HasPrefix(cm.message, "Error") {
			s += "\n\n" + errStyle.Render(cm.message)
		} else {
			s += "\n\n" + okStyle.Render(cm.message)
		}
	}

	// Add help at the bottom
	s += "\n\n" + helpStyle.Render(fmt.Sprintf("View: %s | Press Esc to go back, Ctrl+C to quit", cm.currentView))

	return docStyle.Render(s)
}

func main() {
	// Use physical screen for terminal environments
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	
	// Set explicit sizing
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
