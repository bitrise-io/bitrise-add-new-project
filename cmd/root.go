package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise-add-new-project/phases"
	"github.com/spf13/cobra"
)

const (
	cmdFlagKeyAccount  = "account"
	cmdFlagKeyPublic   = "public"
	cmdFlagKeyAPIToken = "api-token"
)

var (
	cmdFlagAPIToken string
	cmdFlagAccount  string
	cmdFlagPublic   bool
	rootCmd         = &cobra.Command{
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
	rootCmd.Flags().StringVar(&cmdFlagAPIToken, cmdFlagKeyAPIToken, "", "Your Bitrise personal access token")
}

func executePhases(cmd cobra.Command, progress *phases.Progress) error {
	if cmd.Flags().Changed(cmdFlagKeyAccount) {
		progress.Account = cmdFlagAccount
	} else {
		account, err := phases.Account(cmdFlagAPIToken)
		if err != nil {
			return err
		}
		progress.Account = account
	}

	if cmd.Flags().Changed(cmdFlagKeyPublic) {
		progress.Public = cmdFlagPublic
	} else {
		public, err := phases.IsPublic()
		if err != nil {
			return err
		}
		progress.Public = public
	}

	// repo
	repoDetails, err := phases.Repo(progress.Public)
	if err != nil {
		return err
	}

	progress.RepoURL = repoDetails.URL
	progress.RepoProvider = repoDetails.Provider
	progress.RepoOwner = repoDetails.Owner
	progress.RepoSlug = repoDetails.Slug
	progress.RepoType = repoDetails.RepoType

	// ssh key
	publicKeyPth, privateKeyPth, register, err := phases.PrivateKey()
	if err != nil {
		return err
	}
	progress.SSHPrivateKeyPth = privateKeyPth
	progress.SSHPublicKeyPth = publicKeyPth
	progress.RegisterSSHKey = register

	// bitrise.yml
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory, error: %s", err)
	}
	bitriseYML, primaryWorkflow, err := phases.BitriseYML(currentDir)
	if err != nil {
		return err
	}
	projectType := bitriseYML.ProjectType
	if projectType == "" {
		projectType = "other"
	}
	progress.BitriseYML = bitriseYML
	progress.PrimaryWorkflow = primaryWorkflow
	progress.ProjectType = projectType

	// stack
	stack, err := phases.Stack(projectType)
	if err != nil {
		return err
	}
	progress.Stack = stack

	// webhook
	wh, err := phases.AddWebhook()
	if err != nil {
		return err
	}
	progress.AddWebhook = wh

	// codesign
	codesign, err := phases.AutoCodesign(projectType)
	if err != nil {
		return err
	}
	progress.Codesign = codesign

	return nil
}

func run(cmd *cobra.Command, args []string) {
	progress := &phases.Progress{}
	if err := executePhases(*cmd, progress); err != nil {
		fmt.Println("failed to execute phases, error:", err)
		os.Exit(1)
	}

	if err := phases.Register(*progress, cmdFlagAPIToken); err != nil {
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
