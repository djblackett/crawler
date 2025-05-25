package main

import (
	"net/url"
	"reflect"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		expected string
	}{
		{
			name:     "remove https scheme",
			inputURL: "https://blog.boot.dev/path",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "remove https scheme and trailing /",
			inputURL: "https://blog.boot.dev/path/",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "remove http scheme",
			inputURL: "http://blog.boot.dev/path",
			expected: "blog.boot.dev/path",
		},
		{
			name:     "remove http scheme and trailing /",
			inputURL: "http://blog.boot.dev/path/",
			expected: "blog.boot.dev/path",
		},
		// add more test cases here
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := normalizeURL(tc.inputURL)
			if err != nil {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
				return
			}
			if actual != tc.expected {
				t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
			}
		})
	}
}

func TestGetURLsFromHTML(t *testing.T) {
	tests := []struct {
		name       string
		rawBaseURL string
		inputHTML  string
		expected   []string
	}{
		{
			name:       "absolute and relative URLs on Boot.dev",
			rawBaseURL: "https://blog.boot.dev",
			inputHTML: `
		<html>
			<body>
				<a href="/path/one">
					<span>Boot.dev</span>
				</a>
				<a href="https://other.com/path/one">
					<span>Boot.dev</span>
				</a>
			</body>
		</html>
		`,
			expected: []string{"https://blog.boot.dev/path/one", "https://other.com/path/one"},
		},
		{
			name:       "Absolute and relative links",
			rawBaseURL: "https://example.com",
			inputHTML: `
				<html>
					<head><title>Test Page</title></head>
					<body>
						<a href="https://google.com">Google</a>
						<a href="/about">About Us</a>
						<a href="contact.html">Contact</a>
					</body>
				</html>`,
			expected: []string{
				"https://google.com",
				"https://example.com/about",
				"https://example.com/contact.html",
			},
		},
		{
			name:       "Mixed link formats",
			rawBaseURL: "http://testsite.org/dir/",
			inputHTML: `
				<html>
					<body>
						<a href="../parent">Parent Dir</a>
						<a href="./sibling">Sibling</a>
						<a href="http://external.com/page">External Page</a>
					</body>
				</html>`,
			expected: []string{
				"http://testsite.org/parent",
				"http://testsite.org/dir/sibling",
				"http://external.com/page",
			},
		},
		{
			name:       "Base URL with trailing slash",
			rawBaseURL: "https://site.io/blog/",
			inputHTML: `
				<html>
					<body>
						<a href="/home">Home</a>
						<a href="2025/05/24/article">Article</a>
					</body>
				</html>`,
			expected: []string{
				"https://site.io/home",
				"https://site.io/blog/2025/05/24/article",
			},
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseURL, err := url.Parse(tc.rawBaseURL)
			if err != nil {
				return
			}
			actual, err := getURLsFromHTML(tc.inputHTML, baseURL)
			if err != nil {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
				return
			}
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
			}
		})
	}
}
