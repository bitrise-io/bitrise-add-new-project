package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise-add-new-project/phases"
	"github.com/bitrise-io/go-utils/log"
	"github.com/spf13/cobra"
)

const (
	cmdFlagKeyOrganisation = "org"
	cmdFlagKeyPublic       = "public"
	cmdFlagKeyAPIToken     = "api-token"
	cmdFlagKeyVerbose      = "verbose"
)

var (
	cmdFlagAPIToken     string
	cmdFlagOrganisation string
	cmdFlagVerbose      bool
	cmdFlagPublic       bool
	rootCmd             = &cobra.Command{
		Run:   run,
		Use:   "bitrise-add-new-project",
		Short: "Register a new Bitrise Project on bitrise.io",
		Long:  "A guided process for creating a pipeline on bitrise.io.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flag(cmdFlagKeyAPIToken).Value.String() == "" {
				return errors.New("--api-token not defined")
			}
			return nil
		},
	}
)

func init() {
	rootCmd.Flags().StringVar(&cmdFlagOrganisation, cmdFlagKeyOrganisation, "", "The slug of the organization to assign the project")
	rootCmd.Flags().BoolVar(&cmdFlagPublic, cmdFlagKeyPublic, false, "Create a public app")
	rootCmd.Flags().StringVar(&cmdFlagAPIToken, cmdFlagKeyAPIToken, "", "Bitrise personal access token")
	rootCmd.Flags().BoolVar(&cmdFlagVerbose, cmdFlagKeyVerbose, false, "Enable verbose logging")
}

func executePhases(cmd cobra.Command) (phases.Progress, error) {
	progress := phases.Progress{}

	if cmd.Flags().Changed(cmdFlagKeyOrganisation) {
		progress.OrganizationSlug = cmdFlagOrganisation
	} else {
		account, err := phases.Account(cmdFlagAPIToken)
		if err != nil {
			return phases.Progress{}, err
		}
		progress.OrganizationSlug = account
	}

	if cmd.Flags().Changed(cmdFlagKeyPublic) {
		progress.Public = cmdFlagPublic
	} else {
		public, err := phases.IsPublic()
		if err != nil {
			return phases.Progress{}, err
		}
		progress.Public = public
	}

	// Search dir
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return phases.Progress{}, fmt.Errorf("failed to get current directory, error: %s", err)
	}

	// repo
	log.Infof("SCANNING WORKDIR FOR GIT REPO")
	log.Infof("=============================")

	repoURL, err := phases.Repo(currentDir, progress.Public)
	if err != nil {
		return phases.Progress{}, err
	}
	progress.RepoURL = repoURL

	log.Donef("REPOSITORY SCANNED. DETAILS:")
	log.Donef("- scheme: %s", repoURL.Scheme)
	log.Donef("- url: %s", repoURL.URL)
	log.Donef("- provider: %s", repoURL.Provider)
	log.Donef("- owner: %s", repoURL.Owner)
	log.Donef("- slug: %s", repoURL.Slug)
	log.Donef("- username: %s", repoURL.SSHUsername)

	// ssh key
	if repoURL.Scheme == phases.SSH {
		publicKeyPth, privateKeyPth, register, err := phases.PrivateKey(progress.RepoURL)
		if err != nil {
			return phases.Progress{}, err
		}
		progress.SSHPrivateKeyPth = privateKeyPth
		progress.SSHPublicKeyPth = publicKeyPth
		progress.RegisterSSHKey = register
	}

	// bitrise.yml
	bitriseYML, primaryWorkflow, err := phases.BitriseYML(currentDir)
	if err != nil {
		return phases.Progress{}, err
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
		return phases.Progress{}, err
	}
	progress.Stack = stack

	// webhook
	wh, err := phases.AddWebhook()
	if err != nil {
		return phases.Progress{}, err
	}
	progress.AddWebhook = wh

	// codesign
	codesign, err := phases.AutoCodesign(bitriseYML, currentDir)
	if err != nil {
		return phases.Progress{}, err
	}
	progress.Codesign = codesign

	return progress, nil
}

func run(cmd *cobra.Command, args []string) {
	log.SetEnableDebugLog(cmdFlagVerbose)

	progress, err := executePhases(*cmd)
	if err != nil {
		fmt.Println("failed to execute phases, error:", err)
		os.Exit(1)
	}

	if err := phases.Register(cmdFlagAPIToken, progress, os.Stdin); err != nil {
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
