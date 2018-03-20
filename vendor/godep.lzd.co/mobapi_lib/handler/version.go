package handler

import (
	"net/http"
	"regexp"
	"strings"
)

var appVersionRegExp = regexp.MustCompile(`^(\d{1,2})(?:\.(\d{1,2}))?(?:\.(\d{1,2}))?(?:\.(\d{1,2}))?(?:\.(\d{1,2}))?$`)

func filter(in []string) (out []string) {
	for _, i := range in {
		if i != "" {
			out = append(out, i)
		}
	}
	return
}

func GetVersionFromRequest(req *http.Request) (result string) {
	if v := req.Header.Get(HeaderAppVersion); v != "" {
		m := appVersionRegExp.FindStringSubmatch(v)
		if len(m) > 0 {
			m = filter(m[1:5])
			result += GetPlatformFromUserAgent(req.UserAgent())
			if result != "" {
				result += "_"
			}
			result += strings.Join(m, "_")
			return
		}
	}
	return GetVersionFromUserAgent(req.UserAgent())
}

var userAgentRegExp = regexp.MustCompile(`^(?:Lazada(?:-\w+)?|Dalvik)/(.*?) \(.+ ((?:iOS|Android)) .+\)$`)

func GetVersionFromUserAgent(ua string) (result string) {
	m := userAgentRegExp.FindStringSubmatch(ua)
	if len(m) == 3 {
		result += m[2]
		if m[2] == "iOS" {
			v := appVersionRegExp.FindStringSubmatch(m[1])
			if len(v) > 0 {
				v = filter(v[1:5])
				if result != "" {
					result += "_"
				}
				result += strings.Join(v, "_")
			}
		}
	}
	return
}

func GetPlatformFromUserAgent(ua string) string {
	if strings.HasPrefix(ua, "okhttp") || strings.Contains(ua, " Android ") {
		return "Android"
	}
	if strings.Contains(ua, " iOS ") {
		return "iOS"
	}
	return ""
}
