package handler_test

import (
	"motify_core_api/godep_libs/mobapi_lib/handler"
	"net/http"
	"testing"
)

func TestGetVersionFromRequest(t *testing.T) {
	tests := []struct {
		in  *http.Request
		out string
	}{
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/11.8.1 (iPhone; iOS 9.0.2; Scale/2.00)",
			}),
			out: "iOS_11_8_1",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/1.82 (iPhone; iOS 9.0; Scale/2.00)",
			}),
			out: "iOS_1_82",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/1.8bad (iPhone; iOS 9.1; Scale/2.00)",
			}),
			out: "iOS",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/1.8.1 (iPhone; iOS 9.0.2; Scale/2.00)",
			}),
			out: "iOS_1_8_1",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/1.8.1 (iPhone; iOS 9.0; Scale/2.00)",
			}),
			out: "iOS_1_8_1",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/1.8.1 (iPhone; iOS 9.1; Scale/2.00)",
			}),
			out: "iOS_1_8_1",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/1.8.1 (iPhone; iOS 9.2.1; Scale/2.00)",
			}),
			out: "iOS_1_8_1",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/1.8.1 (iPhone; iOS 9.2; Scale/2.00)",
			}),
			out: "iOS_1_8_1",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/5.0 (iPhone; iOS 9.2; Scale/3.00)",
			}),
			out: "iOS_5_0",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/5.0 (iPhone; iOS 9.3.1; Scale/3.00)",
			}),
			out: "iOS_5_0",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/5.0 (iPhone; iOS 9.3.2; Scale/3.00)",
			}),
			out: "iOS_5_0",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/5.0a (iPhone; iOS 9.3; Scale/3.00)",
			}),
			out: "iOS",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/5.0 (iPod touch; iOS 7.1.2; Scale/2.00)",
			}),
			out: "iOS_5_0",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/5.0 (iPod touch; iOS 8.1.2; Scale/2.00)",
			}),
			out: "iOS_5_0",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/5.0 (iPod touch; iOS 8.1; Scale/2.00)",
			}),
			out: "iOS_5_0",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/5.0.2 (iPod touch; iOS 8.1; Scale/2.00)",
			}),
			out: "iOS_5_0_2",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/5.0.0.4 (iPod touch; iOS 8.1; Scale/2.00)",
			}),
			out: "iOS_5_0_0_4",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada/5.0.0.4.1 (iPod touch; iOS 8.1; Scale/2.00)",
			}),
			out: "iOS_5_0_0_4",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada-/5.6 (iPod touch; iOS 8.1; Scale/2.00)",
			}),
			out: "",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada-d+d/5.6 (iPod touch; iOS 8.1; Scale/2.00)",
			}),
			out: "",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada-Test/5.6 (iPod touch; iOS 8.1; Scale/2.00)",
			}),
			out: "iOS_5_6",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Lazada-Dev/5.6 (iPod touch; iOS 8.1; Scale/2.00)",
			}),
			out: "iOS_5_6",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Dalvik/1.6.0 (Linux; U; Android 4.0.3; MediaPad 7 Lite Build/HuaweiMediaPad)",
			}),
			out: "Android",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Dalvik/1.6.0 (Linux; U; Android 4.0.3; Next7P12 Build/IML74K)",
			}),
			out: "Android",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Dalvik/1.6.0 (Linux; U; Android 4.0.3; SC-02C Build/IML74K)",
			}),
			out: "Android",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent":             "Dalvik/1.6a (Linux; U; Android 4.0.3; Next7P12 Build/IML74K)",
				handler.HeaderAppVersion: "5.0.0.5",
			}),
			out: "Android_5_0_0_5",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent":             "Dalvik/1.6a (Linux; U; Android 4.0.3; Next7P12 Build/IML74K)",
				handler.HeaderAppVersion: "5.0.5",
			}),
			out: "Android_5_0_5",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent":             "Dalvik/1.6a (Linux; U; Android 4.0.3; Next7P12 Build/IML74K)",
				handler.HeaderAppVersion: "5.0.0.5.1",
			}),
			out: "Android_5_0_0_5",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent":             "Dalvik/1.6a (Linux; U; Android 4.0.3; Next7P12 Build/IML74K)",
				handler.HeaderAppVersion: "5.0.0.5.1.1",
			}),
			out: "Android",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Android 4.0.3",
			}),
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "iOS",
			}),
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent":             "test",
				handler.HeaderAppVersion: "5.1",
			}),
			out: "5_1",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent":             "",
				handler.HeaderAppVersion: "",
			}),
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent":             "test",
				handler.HeaderAppVersion: "5",
			}),
			out: "5",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Mozilla/5.0 (Linux; U; Android 4.1.1; en-us; ALCATEL ONE TOUCH 6033X Build/JRO03C) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.1 Mobile Safari/534.30",
			}),
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Mozilla/5.0 (Linux; U; Android 4.1.2; en-us; ALCATEL ONE TOUCH 5020T Build/JZO54K) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.1 Mobile Safari/534.30",
			}),
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent": "Mozilla/5.0 (Linux; U; Android 4.2.1; en-us; ALCATEL ONE TOUCH 8008D Build/JOP40D) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.2 Mobile Safari/534.30",
			}),
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent":             "okhttp/2.5.0",
				handler.HeaderAppVersion: "5.0.0.4",
			}),
			out: "Android_5_0_0_4",
		},
		{
			in: genRequestWithHeaders(map[string]string{
				"User-Agent":             "test okhttp/2.5.0",
				handler.HeaderAppVersion: "5.0.0.4",
			}),
			out: "5_0_0_4",
		},
	}
	for _, tt := range tests {
		actual := handler.GetVersionFromRequest(tt.in)
		if actual != tt.out {
			t.Errorf("GetVersionFromRequest(%#v): expected %s, actual %s", tt.in, tt.out, actual)
		}
	}
}

func genRequestWithHeaders(hh map[string]string) (h *http.Request) {
	h = &http.Request{Header: make(http.Header)}
	for k, v := range hh {
		h.Header.Add(k, v)
	}
	return
}
