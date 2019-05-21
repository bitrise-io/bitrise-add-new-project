package bitrise

// CreateProjectParams ...
type CreateProjectParams struct {
	Repository      RegisterParams
	SSHKey          RegisterSSHKeyParams
	RegisterWebhook bool
	Project         RegisterFinishParams
}

// CreateProject ...
func CreateProject(token string, params CreateProjectParams) (string, error) {
	client, err := NewClient(token)
	if err != nil {
		return "", err
	}
	slug, err := client.Register(params.Repository)
	if err != nil {
		return "", err
	}
	if err := client.RegisterSSHKey(slug, params.SSHKey); err != nil {
		return "", err
	}
	if params.RegisterWebhook {
		if err := client.RegisterWebhook(slug); err != nil {
			return "", err
		}
	}
	return client.RegisterFinish(slug, params.Project)
}
