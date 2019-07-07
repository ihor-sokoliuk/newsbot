package bot

import "testing"

func TestFormatLink(t *testing.T) {
	testCases := [][]string{{"https://24tv.ua/yak_prohodyat_vibori_2019_2_tur_ukrayina_zaraz_novini_21_04_2019_n1141866?utm_source=rss", "https://24tv.ua/yak_prohodyat_vibori_2019_2_tur_ukrayina_zaraz_novini_21_04_2019_n1141866"},
		{"https://www.rbc.ua/rus/news/shri-lanke-proizoshel-novyy-vzryv-1555837616.html", "https://www.rbc.ua/rus/news/shri-lanke-proizoshel-novyy-vzryv-1555837616.html"}}
	for i := range testCases {
		formatedLink := FormatLink(testCases[i][0])
		if testCases[i][1] != formatedLink {
			t.Errorf("Expected: %s\nActual:%s", testCases[i][1], formatedLink)
		}
	}
}
