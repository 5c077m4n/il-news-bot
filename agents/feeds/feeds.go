// Package feeds holds the data fetching functions
package feeds

var (
	GetIsrealHayom   = getRSSFeed("https://www.israelhayom.co.il/rss.xml")
	GetYNet          = getRSSFeed("https://www.ynet.co.il/Integration/StoryRss2.xml")
	GetJPost         = getRSSFeed("https://www.jpost.com/rss/rssfeedsfrontpage.aspx")
	GetMakorRishon   = getRSSFeed("https://www.makorrishon.co.il/feed/")
	GetAbuAliExpress = getChannelFeed("@abualiexpress")
)
