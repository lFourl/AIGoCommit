package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	openai "github.com/sashabaranov/go-openai"
)

// Model represents the state of our application.
type model struct {
	commitMsg string
	err       error
}

// Init initializes the model.
func (m model) Init() tea.Cmd {
	return nil
}

// Update updates the model based on incoming messages.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			m.commitMsg, m.err = generateCommitMessageFromStagedChanges()
			if m.err != nil {
				m.commitMsg = fmt.Sprintf("Error: %v", m.err)
			}
			return m, nil
		}
	}

	return m, nil
}

// View renders the user interface.
func (m model) View() string {
	return fmt.Sprintf(
		"Commit message:\n%s\n\n%s",
		m.commitMsg,
		"(press enter to generate, ctrl+c to quit)",
	)
}

// getStagedChanges retrieves the staged changes using git diff.
func getStagedChanges() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// generateCommitMessageFromStagedChanges uses OpenAI GPT to create a commit message based on staged changes.
func generateCommitMessageFromStagedChanges() (string, error) {
	changes, err := getStagedChanges()
	if err != nil {
		return "", err
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an assistant that generates concise and meaningful git commit messages.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Generate a git commit message for the following changes:\n%s", changes),
			},
		},
	})

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func main() {
	m := model{}

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}
