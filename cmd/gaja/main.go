package main

import (
	"os"

	"github.com/nilsocket/gaja"
	"github.com/nilsocket/gaja/ghttp"
)

func main() {
	gaja.Download(ghttp.NewFiles(os.Args[1:]...)...)
	gaja.Close()
}

// urls := []string{
// "https://happiness-report.s3.amazonaws.com/2021/WHR+21.pdf",
// "https://cartographicperspectives.org/index.php/journal/article/view/cp43-complete-issue/pdf",
// "http://www3.weforum.org/docs/WEF_GGGR_2021.pdf",
// "https://apps.who.int/iris/bitstream/10665/186463/1/9789240694811_eng.pdf",
// "https://www.worldobesityday.org/assets/downloads/COVID-19-and-Obesity-The-2021-Atlas.pdf",
// "https://sustainabledevelopment.un.org/content/documents/5987our-common-future.pdf",
// "https://www.fda.gov/media/120060/download",
// "https://www.ilo.org/wcmsp5/groups/public/@dgreports/@dcomm/documents/briefingnote/wcms_767028.pdf",
// "https://vod-progressive.akamaized.net/exp=1618098051~acl=%2Fvimeo-prod-skyfire-std-us%2F01%2F4378%2F19%2F496894990%2F2368408923.mp4~hmac=bf694b92410673cb28ef399dcc8f0eb0da3d62f4d57deeaee314300c9dd884e3/vimeo-prod-skyfire-std-us/01/4378/19/496894990/2368408923.mp4",
// }
