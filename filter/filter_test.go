package filter

import (
	"net/http"
	"strings"
	"testing"
)

type matchTestCase struct {
	url  string
	want bool
}

type matchTest struct {
	name  string
	rules []string
	cases []matchTestCase
}

func (mt *matchTest) run(t *testing.T) {
	filter := NewFilter()
	rulesReader := strings.NewReader(strings.Join(mt.rules, "\n"))
	filter.AddRules(rulesReader, nil)
	for _, u := range mt.cases {
		req, _ := http.NewRequest("GET", u.url, nil)
		block, _ := filter.HandleRequest(req)
		if block != u.want {
			t.Errorf("filter.HandleRequest(%q) = %v, want %v", u.url, block, u.want)
		}
	}
}

func TestFilterByAddressParts(t *testing.T) {
	t.Parallel()
	tests := []matchTest{
		{
			name:  "by address parts",
			rules: []string{"/banner/img"},
			cases: []matchTestCase{
				{"http://example.com/banner/img", true},
				{"https://example.com/banner/img", true},
				{"http://example.com/example/banner/img", true},
				{"http://example.com/banner/img/example", true},
				{"http://example.com/banner-img", false},
				{"https://example.com/banner?img", false},
				{"http://example.com", false},
			},
		},
		{
			name:  "by segments",
			rules: []string{"-banner-ad-"},
			cases: []matchTestCase{
				{"http://example.com/-banner-ad-", true},
				{"https://example.com/-example-banner-ad-example", true},
				{"http://example.com/-banner-ad-example", true},
				{"http://example.com/banner-ad", false},
				{"https://example.com/banner-ad", false},
				{"https://example.com/this-is-a-banner-ad", false},
				{"http://example.com/ad-banner", false},
				{"http://example.com/banner-ad-", false},
				{"https://example.com/-banner-ad", false},
				{"http://example.com/banner-ad-example", false},
				{"http://example.com/banner?ad", false},
			},
		},
		{
			name: "by multiple segments",
			rules: []string{
				"-banner-ad-",
				"-ad-banner-",
				"-adfliction/",
				"-adframe.",
				".html?clicktag=",
				".html?ad=",
				".html?ad_",
				"/ad-top-",
			},
			cases: []matchTestCase{
				{"http://example.com/-banner-ad-", true},
				{"https://example.com/-ad-banner-", true},
				{"http://example.com/-adfliction/", true},
				{"http://example.com/-adframe.", true},
				{"http://example.com/-adframe.html", true},
				{"http://example.com/innocent.html?clicktag=", true},
				{"http://example.com/innocent.html?ad=", true},
				{"http://example.com/innocent.html?ad_", true},
				{"http://example.com/ad-top-", true},
				{"http://example.com/-banner-ad-example", true},
				{"http://example.com/banner-ad", false},
				{"https://example.com/banner-ad", false},
				{"http://test.org", false},
				{"https://example.com/this-is-a-banner-ad", false},
				{"http://example.com/ad-banner", false},
				{"http://example.com/test.html?click=", false},
				{"http://example.com/test.html?ad", false},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, test.run)
	}
}

func TestFilterHosts(t *testing.T) {
	t.Parallel()
	tests := []matchTest{
		{
			name:  "single host",
			rules: []string{"0.0.0.0 example.com"},
			cases: []matchTestCase{
				{"http://example.com", true},
				{"https://example.com", true},
				{"http://example.com/", true},
				{"http://example.com/?q=example", true},
				{"https://example.com/subdir/doc?foo1=bar1&foo2=bar2", true},
				{"http://example.com:8080", true},
				{"https://example.co", false},
				{"http://example.co", false},
				{"http://example.com.co", false},
				{"http://example.com.co/", false},
				{"http://example.com.co/?q=example", false},
			},
		},
		{
			name:  "multiple components",
			rules: []string{"0.0.0.0 sub.test.example.com"},
			cases: []matchTestCase{
				{"http://sub.test.example.com", true},
				{"https://sub.test.example.com", true},
				{"http://sub.test.example.com/", true},
				{"http://sub.test.example.com/?q=example", true},
				{"https://sub.test.example.com/subdir/doc?foo1=bar1&foo2=bar2", true},
				{"http://sub.test.example.com:8080", true},
				{"https://test.example.com", false},
				{"http://test.example.com", false},
				{"http://sub.test.example.co", false},
				{"http://sub.test.example.co/", false},
				{"http://sub.test.example.co/?q=example", false},
			},
		},
		{
			name: "multiple hosts",
			rules: []string{
				"0.0.0.0 example.com",
				"127.0.0.1 example.org",
				"0.0.0.0 test.sub.foo.xyz",
			},
			cases: []matchTestCase{
				{"http://example.com", true},
				{"https://example.com", true},
				{"http://example.org", true},
				{"https://example.org", true},
				{"http://test.sub.foo.xyz", true},
				{"https://test.sub.foo.xyz", true},
				{"http://example.com/", true},
				{"https://example.com/", true},
				{"http://example.com/?q=example", true},
				{"https://example.com/?q=example", true},
				{"https://example.com/subdir/doc?foo1=bar1&foo2=bar2", true},
				{"http://example.com:8080", true},
				{"https://example.co", false},
				{"http://test.sub.foo", false},
				{"http://example.edu", false},
				{"http://example.edu/doc", false},
			},
		},
		{
			name: "multiple overlapping hosts",
			rules: []string{
				"0.0.0.0 example.com",
				"0.0.0.0 example.com.co",
				"0.0.0.0 example.com.co.uk",
				"0.0.0.0 example.com.co.uk.co.uk",
			},
			cases: []matchTestCase{
				{"http://example.com", true},
				{"http://example.com.co", true},
				{"http://example.com.co.uk", true},
				{"http://example.com.co.uk.co.uk", true},
				{"http://example.com.co.uk.co", false},
				{"http://example.edu", false},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, test.run)
	}
}

func TestFilterByDomainName(t *testing.T) {
	t.Parallel()
	tests := []matchTest{
		{
			name: "single rule",
			rules: []string{
				"||example.org",
			},
			cases: []matchTestCase{
				{"http://example.org", true},
				{"https://example.org", true},
				{"http://example.org/", true},
				{"http://example.org/?q=example", true},
				{"https://example.org/subdir/doc?foo1=bar1&foo2=bar2", true},
				{"http://example.org:8080", true},
				{"https://example.org.co", true},
				{"https://example.com", false},
			},
		},
		{
			name: "single rule with separator at the end",
			rules: []string{
				"||example.org^",
			},
			cases: []matchTestCase{
				{"http://example.org", true},
				{"https://example.org", true},
				{"http://example.org/", true},
				{"http://example.org/?q=example", true},
				{"https://example.org/subdir/doc?foo1=bar1&foo2=bar2", true},
				{"http://example.org:8080", true},
				{"https://example.com", false},
				{"http://example.com", false},
				{"http://example.com.co", false},
				{"http://example.com.co/", false},
			},
		},
		{
			name: "multiple rules",
			rules: []string{
				"||example.org^",
				"||example.com^",
				"||test.example.com^",
			},
			cases: []matchTestCase{
				{"http://example.org", true},
				{"https://example.org", true},
				{"http://example.org/", true},
				{"http://example.org/?q=example", true},
				{"https://example.org/subdir/doc?foo1=bar1&foo2=bar2", true},
				{"http://example.org:8080", true},
				{"https://example.com", true},
				{"http://example.com", true},
				{"http://example.com.co", false},
				{"http://example.com.co/", false},
				{"http://example.com.co/?q=example", false},
				{"http://example.edu", false},
				{"http://example.edu/doc", false},
				{"http://test.example.com", true},
				{"http://test.example.com/", true},
				{"http://test.example.com/?q=example", true},
				{"http://test.example.com:8080", true},
				{"http://test.example.com.co", false},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, test.run)
	}
}

func TestFilterByExactAddress(t *testing.T) {
	t.Parallel()
	tests := []matchTest{
		{
			name: "single rule",
			rules: []string{
				"|http://example.org/",
			},
			cases: []matchTestCase{
				{"http://example.org", false},
				{"https://example.org", false},
				{"http://example.org/", true},
				{"https://example.org/", false},
				{"http://example.org/?q=example", false},
				{"https://example.org/banner/img", false},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, test.run)
	}
}

func TestFilterWildcard(t *testing.T) {
	t.Parallel()
	tests := []matchTest{
		{
			name: "single wildcard rule",
			rules: []string{
				"/beacon/track/*",
			},
			cases: []matchTestCase{
				{"http://example.org/beacon/track/foo", true},
				{"http://example.org/beacon/track/foo/bar", true},
				{"http://example.org/beacon/track", false},
				{"http://example.org/beacon/track/", false},
				{"http://example.org/beacon", false},
			},
		},
		{
			name: "multiple wildcard rules",
			rules: []string{
				"/beacon/track/*",
				"/beacon/*/events/*",
			},
			cases: []matchTestCase{
				{"http://example.org/beacon/track/foo", true},
				{"http://example.org/beacon/track/foo/bar", true},
				{"http://example.org/beacon/track", false},
				{"http://example.org/beacon/track/", false},
				{"http://example.org/beacon", false},
				{"http://example.org/beacon/foo/events/bar", true},
				{"http://example.org/beacon/foo/events/bar/baz", true},
				{"http://example.org/beacon/foo/events", false},
				{"http://example.org/beacon/foo/events/", false},
				{"http://example.org/beacon/foo", false},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, test.run)
	}
}
