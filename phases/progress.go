package phases

// Progress ...
type Progress struct {
	Account      *string
	Public       *bool
	Repo         *string
	RepoURL      *string
	RepoProvider *string
	RepoOwner    *string
	RepoSlug     *string
	RepoType     *string
	PrivateKey   *string
	BitriseYML   *string
	Stack        *string
	AddWebhook   *bool
	AutoCodesign *bool
}
