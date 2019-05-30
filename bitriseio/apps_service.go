package bitriseio

// AppsServiceURL ...
const AppsServiceURL = "apps/"

// AppsService ...
type AppsService struct {
	client *Client
}

// AppService ...
type AppService struct {
	client *Client
	Slug   string
}
