package spider

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temoto/robotstxt"
)

func TestIsInternalURLPredicate(t *testing.T) {
	testURL, err := url.Parse("http://willdemaine.co.uk")
	require.NoError(t, err)

	noSubPred := createIsInternalPredicate(testURL, false)
	subPred := createIsInternalPredicate(testURL, true)

	cases := []struct {
		name     string
		pred     urlPredicate
		uri      string
		expected bool
	}{
		{"local", noSubPred, "/foo", true},
		{"local no /", noSubPred, "foo", true},
		{"same host", noSubPred, "http://willdemaine.co.uk", true},
		{"path", noSubPred, "http://willdemaine.co.uk/foo", true},
		{"subdomain", noSubPred, "http://foo.willdemaine.co.uk", false},
		{"external", noSubPred, "http://foo.bar.co.uk", false},

		{"local (sub)", subPred, "/foo", true},
		{"local no / (sub)", subPred, "foo", true},
		{"same host (sub)", subPred, "http://willdemaine.co.uk", true},
		{"path (sub)", subPred, "http://willdemaine.co.uk/foo", true},
		{"subdomain (sub)", subPred, "http://foo.willdemaine.co.uk", true},
		{"external (sub)", subPred, "http://foo.bar.co.uk", false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := url.Parse(test.uri)
			require.NoError(t, err)
			resolved := testURL.ResolveReference(parsed)
			assert.Equal(t, test.expected, test.pred(resolved))
		})
	}
}

func TestNotSeenPredicate(t *testing.T) {
	fooSeener := urlPredicate(func(input *url.URL) bool {
		return strings.HasSuffix(input.String(), "foo")
	})
	pred := createNotSeenPredicate(fooSeener)

	cases := []struct {
		name     string
		uri      string
		expected bool
	}{
		{"not seen", "notseen.com", true},
		{"seen", "seen.com/foo", false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := url.Parse(test.uri)
			require.NoError(t, err)
			assert.Equal(t, test.expected, pred(parsed))
		})
	}
}

func TestShoudRequestByRobotsPredicate(t *testing.T) {
	robots, err := robotstxt.FromStatusAndString(200, `
		User-agent: agent
		Allow: /foo/
		Disallow: /bar/
	`)
	require.NoError(t, err)

	cases := []struct {
		name     string
		agent    string
		path     string
		expected bool
	}{
		{"agent allow path", "agent", "/foo/123", true},
		{"agent disallow path", "agent", "/bar/123", false},
		{"wrong agent allow path", "foo", "/foo/123", true},
		{"wrong agent disallow path", "foo", "/bar/123", true},
	}

	fooCom, err := url.Parse("http://foo.com")
	require.NoError(t, err)

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			predicate := createShouldRequestByRobotsPredicate(test.agent, robots)

			parsed, err := url.Parse(test.path)
			require.NoError(t, err)

			finalURL := fooCom.ResolveReference(parsed)
			res := predicate(finalURL)
			assert.Equal(t, test.expected, res)
		})
	}
}

func TestShouldRequestByRobotsNil(t *testing.T) {
	predicate := createShouldRequestByRobotsPredicate("foo", nil)
	fooURL, err := url.Parse("/foo")
	require.NoError(t, err)

	assert.True(t, predicate(fooURL))
}
