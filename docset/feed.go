package docset

// Dash documentation feeds are very simple:
//   <entry>
//     <version>0.10.26</version>
//     <url>http://newyork.kapeli.com/feeds/NodeJS.tgz</url>
//     <url>http://sanfrancisco.kapeli.com/feeds/NodeJS.tgz</url>
//     <url>http://london.kapeli.com/feeds/NodeJS.tgz</url>
//     <url>http://tokyo.kapeli.com/feeds/NodeJS.tgz</url>
//   </entry>
//

type Feed struct {
	Entry FeedEntry `xml:"entry"`
}

type FeedEntry struct {
	// You can use any versioning system you want. Dash/Zeal will use string comparison to
	// determine whether or not to download an update.
	Version string `xml:"version"`

	// One or several <url> elements. These point to the URL of the archived docset. They
	// should refer to the same file.
	Url []string `xml:"url"`
}
