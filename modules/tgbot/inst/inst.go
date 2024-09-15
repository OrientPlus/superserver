package inst

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
)

type ReelsDownloader struct {
}

// Парсит страницу с рилсом и скачивает в директорию
// @param url - URL рилса
// @return - возвращает имя рилса в директории
func (r *ReelsDownloader) DownloadInstagramReel(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("ошибка при загрузке страницы, статус код: %d", res.StatusCode)
	}

	// Парсим страницу
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	var videoURL string

	// Ищем тег <meta> с атрибутом property="og:video"
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if name, _ := s.Attr("property"); name == "og:video" {
			videoURL, _ = s.Attr("content")
		}
	})

	if videoURL == "" {
		return "", fmt.Errorf("видео не найдено")
	}

	// Возвращаем URL видео
	return videoURL, nil
}
