package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise-add-new-project/phases"
	"github.com/spf13/cobra"
)

const (
	cmdFlagKeyAccount      = "account"
	cmdFlagKeyPublic       = "public"
	cmdFlagKeyRepo         = "repo"
	cmdFlagKeyPrivateKey   = "private-key"
	cmdFlagKeyBitriseYML   = "bitrise-yml"
	cmdFlagKeyStack        = "stack"
	cmdFlagKeyAddWebhook   = "add-webhook"
	cmdFlagKeyAutoCodesign = "auto-codesign"
	cmdFlagKeyAPIToken     = "api-token"
)

var (
	cmdFlagAPIToken     string
	cmdFlagAccount      string
	cmdFlagPublic       bool
	cmdFlagRepo         string
	cmdFlagPrivateKey   string
	cmdFlagBitriseYML   string
	cmdFlagStack        string
	cmdFlagAddWebhook   bool
	cmdFlagAutoCodesign bool
	rootCmd             = &cobra.Command{
		Run:   run,
		Use:   "bitrise-add-new-project",
		Short: "Register a new Bitrise Project on bitrise.io",
		Long: `A guided process for creating a pipeline on bitrise.io

	You can quit the process at any phase and continue from where you left off later.`,
	}
)

func init() {
	rootCmd.Flags().StringVar(&cmdFlagAccount, cmdFlagKeyAccount, "", "Name of Bitrise account to use")
	rootCmd.Flags().BoolVar(&cmdFlagPublic, cmdFlagKeyPublic, false, "Visibility of the Bitrise app")
	rootCmd.Flags().StringVar(&cmdFlagRepo, cmdFlagKeyRepo, "", "Git URL for the repository to register")
	rootCmd.Flags().StringVar(&cmdFlagPrivateKey, cmdFlagKeyPrivateKey, "", "Path to the private key file")
	rootCmd.Flags().StringVar(&cmdFlagBitriseYML, cmdFlagKeyBitriseYML, "", "Path to the bitrise.yml file")
	rootCmd.Flags().StringVar(&cmdFlagStack, cmdFlagKeyStack, "", "The stack to run the builds on")
	rootCmd.Flags().BoolVar(&cmdFlagAddWebhook, cmdFlagKeyAddWebhook, false, "To register a webhook for the git provider")
	rootCmd.Flags().BoolVar(&cmdFlagAutoCodesign, cmdFlagKeyAutoCodesign, false, "Upload codesign files for iOS project")
	rootCmd.Flags().StringVar(&cmdFlagAPIToken, cmdFlagKeyAPIToken, "", "Your Bitrise personal access token")
}

func executePhases(cmd cobra.Command, progress *phases.Progress) error {
	if cmd.Flags().Changed(cmdFlagKeyAccount) {
		progress.Account = &cmdFlagAccount
	}
	if progress.Account == nil {
		account, err := phases.Account(cmdFlagAPIToken)
		if err != nil {
			return err
		}
		progress.Account = &account
	}

	if cmd.Flags().Changed(cmdFlagKeyPublic) {
		progress.Public = &cmdFlagPublic
	}
	if progress.Public == nil {
		public, err := phases.IsPublic()
		if err != nil {
			return err
		}
		progress.Public = &public
	}

	if cmd.Flags().Changed(cmdFlagKeyRepo) {
		progress.Repo = &cmdFlagRepo
	}
	if progress.Repo == nil {
		repoDetails, err := phases.Repo(*progress.Public)
		if err != nil {
			return err
		}

		progress.RepoURL = &repoDetails.URL
		progress.RepoProvider = &repoDetails.Provider
		progress.RepoOwner = &repoDetails.Owner
		progress.RepoSlug = &repoDetails.Slug
		progress.RepoType = &repoDetails.RepoType
	}

	if cmd.Flags().Changed(cmdFlagKeyPrivateKey) {
		progress.PrivateKey = &cmdFlagPrivateKey
	}
	if progress.PrivateKey == nil {
		_, privKey, _, err := phases.PrivateKey()
		if err != nil {
			return err
		}
		progress.PrivateKey = &privKey
	}

	if cmd.Flags().Changed(cmdFlagKeyBitriseYML) {
		DSLFile, err := os.Open(cmdFlagKeyBitriseYML)
		if err != nil {
			return fmt.Errorf("failed to open bitrise.yml, error: %s", err)
		}
		DSL, err := phases.ParseDSLFile(DSLFile)
		if err != nil {
			return fmt.Errorf("failed to parse bitrise.yml, error: %s", err)
		}
		progress.BitriseYML = &DSL
	}
	if progress.BitriseYML == nil {
		currentDir, err := filepath.Abs(".")
		if err != nil {
			return fmt.Errorf("failed to get current directory, error: %s", err)
		}
		yml, _, err := phases.BitriseYML(currentDir)
		if err != nil {
			return err
		}
		progress.BitriseYML = &yml
	}

	if cmd.Flags().Changed(cmdFlagKeyStack) {
		progress.Stack = &cmdFlagStack
	}
	if progress.Stack == nil {
		stack, err := phases.Stack()
		if err != nil {
			return err
		}
		progress.Stack = &stack
	}

	if cmd.Flags().Changed(cmdFlagKeyAddWebhook) {
		progress.AddWebhook = &cmdFlagAddWebhook
	}
	if progress.AddWebhook == nil {
		wh, err := phases.AddWebhook()
		if err != nil {
			return err
		}
		progress.AddWebhook = &wh
	}

	if cmd.Flags().Changed(cmdFlagKeyAutoCodesign) {
		progress.AutoCodesign = &cmdFlagAutoCodesign
	}
	if progress.AutoCodesign == nil {
		codesign, err := phases.AutoCodesign(*progress.BitriseYML)
		if err != nil {
			return err
		}
		_ = codesign
		progress.AutoCodesign = nil
	}

	return nil
}

func run(cmd *cobra.Command, args []string) {
	progress := &phases.Progress{}
	if err := executePhases(*cmd, progress); err != nil {
		fmt.Println("failed to execute phases, error:", err)
		os.Exit(1)
	}

	if err := phases.Register(*progress); err != nil {
		fmt.Println("failed to add Bitrise app, error:", err)
		os.Exit(1)
	}
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Failed to execute the command, error: %s\n", err)
		os.Exit(1)
	}
}
