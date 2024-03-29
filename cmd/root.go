package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise-add-new-project/bitriseio"
	"github.com/bitrise-io/bitrise-add-new-project/phases"
	"github.com/bitrise-io/go-utils/log"
	"github.com/spf13/cobra"
)

const (
	cmdFlagKeyOrganisation    = "org"
	cmdFlagKeyPublic          = "public"
	cmdFlagKeyAPIToken        = "api-token"
	cmdFlagKeyVerbose         = "verbose"
	cmdFlagKeyPersonal        = "personal"
	cmdFlagKeyIsWebsiteSource = "website"
)

var (
	cmdFlagAPIToken        string
	cmdFlagOrganisation    string
	cmdFlagVerbose         bool
	cmdFlagPublic          bool
	cmdFlagPersonal        bool
	cmdFlagIsWebsiteSource bool
	rootCmd                = &cobra.Command{
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
	rootCmd.Flags().BoolVar(&cmdFlagPersonal, cmdFlagKeyPersonal, false, "Assign the project to the owner of the personal access token")
	rootCmd.Flags().BoolVar(&cmdFlagIsWebsiteSource, cmdFlagKeyIsWebsiteSource, false, "Set this flag if the registration started from the Bitrise.io website")
}

func executePhases(cmd cobra.Command) (phases.Progress, error) {
	progress := phases.Progress{}

	account, err := phases.Account(cmdFlagAPIToken, cmdFlagPersonal, cmdFlagOrganisation)
	if err != nil {
		return phases.Progress{}, err
	}
	progress.OrganizationSlug = account

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
	repoURL, err := phases.Repo(currentDir, progress.Public)
	if err != nil {
		return phases.Progress{}, err
	}
	progress.RepoDetails = repoURL

	log.Debugf("REPOSITORY SCANNED. DETAILS:")
	log.Debugf("- url: %s", repoURL.URL)
	log.Debugf("- provider: %s", repoURL.Provider)
	log.Debugf("- owner: %s", repoURL.Owner)
	log.Debugf("- slug: %s", repoURL.Slug)
	log.Debugf("- username: %s", repoURL.SSHUsername)

	// ssh key
	if repoURL.Scheme == phases.SSH {
		SSHKeys, register, err := phases.PrivateKey(progress.RepoDetails)
		if err != nil {
			return phases.Progress{}, err
		}
		progress.SSHKeys = SSHKeys
		progress.RegisterSSHKey = register
	}

	// bitrise.yml
	bitriseYML, primaryWorkflow, branch, err := phases.BitriseYML(currentDir, progress.RegisterSSHKey)
	if err != nil {
		return phases.Progress{}, err
	}
	projectType := bitriseYML.ProjectType
	if projectType == "" {
		projectType = "other"
		bitriseYML.ProjectType = "other"
	}
	progress.BitriseYML = bitriseYML
	progress.PrimaryWorkflow = primaryWorkflow
	progress.Branch = branch
	progress.ProjectType = projectType

	log.Debugf("project type\nprogress: %s, yml; %s", projectType, bitriseYML.ProjectType)

	// stack
	stack, err := phases.Stack(progress.OrganizationSlug, cmdFlagAPIToken, projectType)
	if err != nil {
		return phases.Progress{}, err
	}
	progress.Stack = stack
	progress.BitriseYML.Meta = map[string]interface{}{
		"bitrise.io": map[string]string{
			"stack": stack,
		},
	}

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

	source := bitriseio.SourceBanp
	if cmdFlagIsWebsiteSource {
		source = bitriseio.SourceBanpWebsite
	}

	if err := phases.Register(cmdFlagAPIToken, source, progress, os.Stdin); err != nil {
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
