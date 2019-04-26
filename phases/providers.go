package phases

type providerHandler interface {
	provider() string
	repoType() string
	parseURL(string) urlParts
	buildURL(urlParts, string) string
}