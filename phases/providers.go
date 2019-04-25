package phases

type providerHandler interface {
	provider() string
	parseURL(string) urlParts
	buildURL(urlParts, string) string
}