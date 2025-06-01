package fuzzer

import (
	"github.com/ffuf/ffuf/v2/pkg/output"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// Output wraps ffuf's Stdoutput to provide a cleaner interface for use.
type Output struct {
	inner *output.Stdoutput
}

// NewOutput creates a new Output instance using the provided ffuf config.
func NewOutput(conf *ffuf.Config) *Output {
	return &Output{
		inner: output.NewStdoutput(conf),
	}
}

// Banner displays the ffuf banner.
func (o *Output) Banner() {
	o.inner.Banner()
}

// Reset clears the current scan state.
func (o *Output) Reset() {
	o.inner.Reset()
}

// Cycle moves to the next scan cycle.
func (o *Output) Cycle() {
	o.inner.Cycle()
}

// GetCurrentResults returns the currently collected results.
func (o *Output) GetCurrentResults() []ffuf.Result {
	return o.inner.GetCurrentResults()
}

// SetCurrentResults sets the current result set.
func (o *Output) SetCurrentResults(results []ffuf.Result) {
	o.inner.SetCurrentResults(results)
}

// Progress updates the progress output with current status.
func (o *Output) Progress(status ffuf.Progress) {
	o.inner.Progress(status)
}

// Info logs an informational message.
func (o *Output) Info(infostring string) {
	o.inner.Info(infostring)
}

// Error logs an error message.
func (o *Output) Error(errstring string) {
	o.inner.Error(errstring)
}

// Warning logs a warning message.
func (o *Output) Warning(warnstring string) {
	o.inner.Warning(warnstring)
}

// Raw outputs a raw string to stdout.
func (o *Output) Raw(output string) {
	o.inner.Raw(output)
}

// SaveFile saves the current results to a file with the specified format.
func (o *Output) SaveFile(filename, format string) error {
	return o.inner.SaveFile(filename, format)
}

// Finalize is called after all ffuf jobs are complete to perform final output operations.
func (o *Output) Finalize() error {
	return o.inner.Finalize()
}

// Result processes a single ffuf.Response and appends it to the current results.
func (o *Output) Result(resp ffuf.Response) {
	// Convert request input to a map with string keys and byte slices
	inputs := make(map[string][]byte, len(resp.Request.Input))
	for k, v := range resp.Request.Input {
		inputs[k] = v
	}

	// Build a simplified Result struct from the Response
	sResult := ffuf.Result{
		Input:            inputs,
		Position:         resp.Request.Position,
		StatusCode:       resp.StatusCode,
		ContentLength:    resp.ContentLength,
		ContentWords:     resp.ContentWords,
		ContentLines:     resp.ContentLines,
		ContentType:      resp.ContentType,
		RedirectLocation: resp.GetRedirectLocation(false),
		ScraperData:      resp.ScraperData,
		Url:              resp.Request.Url,
		// Duration:         resp.Duration, // Uncomment if needed
		ResultFile:       resp.ResultFile,
		Host:             resp.Request.Host,
	}

	// Append the result to the output's current results
	o.inner.CurrentResults = append(o.inner.CurrentResults, sResult)
}

// PrintResult prints a single result using the inner output mechanism.
func (o *Output) PrintResult(res ffuf.Result) {
	o.inner.PrintResult(res)
}
