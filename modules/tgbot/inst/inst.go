package inst

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	pw "github.com/playwright-community/playwright-go"
	"superserver/loggers"
)

type ReelModule interface {
	DownloadReelFastdl(reelLink string) (string, error)
	DownloadVideoSavefrom(link string) (string, error)
}

type reelsDownloader struct {
	driver         *pw.Playwright
	browser        pw.Browser
	logger         loggers.Logger
	newHandleMutex sync.Mutex
}

// Парсит страницу с рилсом и скачивает в директорию
// @param reelURL - URL рилса
// @return - возвращает имя рилса в директории
func (r *reelsDownloader) DownloadReelFastdl(reelURL string) (string, error) {
	// Создание новой вкладки
	page, err := r.browser.NewPage()
	if err != nil {
		r.logger.Error(fmt.Sprintf("не удалось открыть вкладку: %s", err))
	}
	defer func() {
		err = page.Close()
		if err != nil {
			r.logger.Error(fmt.Sprintf("не удалось закрыть вкладку: %s", err))
		}
	}()

	// Переход на страницу для скачивания
	if _, err = page.Goto("https://fastdl.app/instagram-reels-download"); err != nil {
		r.logger.Error(fmt.Sprintf("не удалось открыть страницу для скачивания: %s", err))
		return "", err
	}
	r.logger.Debug("открыта страница для скачивания")

	// Ожидание появления и нажатие кнопки "Consent"
	for i := 0; i < 6; i++ {
		consentButton, err := page.QuerySelector("button.fc-button.fc-cta-consent.fc-primary-button")
		if err == nil && consentButton != nil {
			r.logger.Info("обнаружено всплывающее окно")
			err = consentButton.Click()
			if err != nil {
				r.logger.Error(fmt.Sprintf("не удалось нажать кнопку всплывающего окна: %v\n", err))
				return "", err
			}
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Поиск поля ввода
	inputElement, err := page.QuerySelector("#search-form-input")
	if err != nil || inputElement == nil {
		r.logger.Error(fmt.Sprintf("не удалось обнаружить поле ввода для ссылки: %v\n", err))
		return "", err
	}

	// Ввод ссылки на Reel
	if err = inputElement.Fill(reelURL); err != nil {
		r.logger.Error(fmt.Sprintf("не удалось ввести ссылку в поле для ввода: %v\n", err))
		return "", err
	}

	// Нажатие кнопки "Search"
	searchButton, err := page.QuerySelector("button.search-form__button")
	if err != nil || searchButton == nil {
		r.logger.Error(fmt.Sprintf("не удалось найти кнопку отправки ссылки: %v\n", err))
		return "", err
	}
	if err = searchButton.Click(); err != nil {
		r.logger.Error(fmt.Sprintf("не удалось нажать кнопку отправки ссылки: %v\n", err))
		return "", err
	}

	// Ожидание появления кнопки "Download"
	var downloadButton pw.ElementHandle
	for i := 0; i < 25; i++ {
		downloadButton, err = page.QuerySelector("a.button__download")
		if err == nil && downloadButton != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		r.logger.Error(fmt.Sprintf("не удалось найти кнопку скачивания рилса: %v\n", err))
		return "", err
	}

	// Получаем ссылку для скачивания
	downloadLink, err := downloadButton.GetAttribute("href")
	if err != nil || downloadLink == "" {
		r.logger.Error(fmt.Sprintf("не удалось получить ссылку для скачивания: %v\n", err))
		return "", err
	}

	// Путь для сохранения файла
	reelName, err := extractReelID(reelURL)
	if err != nil {
		hash := sha256.New()
		hash.Write([]byte(reelURL))
		hashBytes := hash.Sum(nil)
		reelName = hex.EncodeToString(hashBytes)
	}

	reelName += ".mp4"
	filePath := filepath.Join("./tmpData/inst/", reelName)
	err = r.downloadFile(downloadLink, filePath)
	if err != nil {
		r.logger.Error(fmt.Sprintf("не удалось скачать рилс по ссылке: %v", err))
		return "", nil
	}

	return filePath, nil
}

// Скачивает файл по ссылке
// @url 		- ссылка на видео
// @filePath 	- путь сохранения файла
func (r *reelsDownloader) downloadFile(url, filePath string) error {
	// Отправляем HTTP GET запрос для получения файла
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Ошибка при запросе файла: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Ошибка при создании файла: %v", err)
	}
	defer out.Close()

	// Копируем содержимое ответа в файл
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("Ошибка при записи файла: %v", err)
	}

	return nil
}

func (r *reelsDownloader) DownloadVideoSavefrom(link string) (string, error) {
	// Создание новой вкладки
	page, err := r.browser.NewPage()
	if err != nil {
		r.logger.Error(fmt.Sprintf("не удалось открыть вкладку: %s", err))
	}
	defer func() {
		err = page.Close()
		if err != nil {
			r.logger.Error(fmt.Sprintf("не удалось закрыть вкладку: %s", err))
		}
	}()

	// Переход на страницу для скачивания
	if _, err = page.Goto("https://savefrom.in.net/"); err != nil {
		r.logger.Error(fmt.Sprintf("не удалось открыть страницу для скачивания: %s", err))
		return "", err
	}
	r.logger.Debug("открыта страница для скачивания")

	// Ожидание появления и нажатие кнопки "Consent"
	/*for i := 0; i < 6; i++ {
		consentButton, err := page.QuerySelector("button.fc-button.fc-cta-consent.fc-primary-button")
		if err == nil && consentButton != nil {
			r.logger.Info("обнаружено всплывающее окно")
			err = consentButton.Click()
			if err != nil {
				r.logger.Error(fmt.Sprintf("не удалось нажать кнопку всплывающего окна: %v\n", err))
				return "", err
			}
			break
		}
		time.Sleep(500 * time.Millisecond)
	}*/

	// Поиск поля ввода
	inputElement, err := page.QuerySelector("input#id_url.form__form-input")
	if err != nil || inputElement == nil {
		r.logger.Error(fmt.Sprintf("не удалось обнаружить поле ввода для ссылки: %v\n", err))
		return "", err
	}

	// Ввод ссылки на видео
	if err = inputElement.Fill(link); err != nil {
		r.logger.Error(fmt.Sprintf("не удалось ввести ссылку в поле для ввода: %v\n", err))
		return "", err
	}

	// Нажатие кнопки "Start"
	searchButton, err := page.QuerySelector("button#search.form__form-button")
	if err != nil || searchButton == nil {
		r.logger.Error(fmt.Sprintf("не удалось найти кнопку отправки ссылки: %v\n", err))
		return "", err
	}
	if err = searchButton.Click(); err != nil {
		r.logger.Error(fmt.Sprintf("не удалось нажать кнопку отправки ссылки: %v\n", err))
		return "", err
	}

	// Ожидание появления кнопки "Download"
	var downloadButton pw.ElementHandle
	for i := 0; i < 25; i++ {
		table, err := page.QuerySelector("a.results__table")
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Находим первую строку таблицы
		rows, err := table.QuerySelectorAll("tr")
		if err != nil || len(rows) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		downloadButton, err = rows[0].QuerySelector("a.results__btn-download")
		if err == nil && downloadButton != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil || downloadButton == nil {
		r.logger.Error(fmt.Sprintf("не удалось найти кнопку скачивания видео: %v\n", err))
		return "", err
	}

	// Получаем ссылку для скачивания
	downloadLink, err := downloadButton.GetAttribute("href")
	if err != nil || downloadLink == "" {
		r.logger.Error(fmt.Sprintf("не удалось получить ссылку для скачивания: %v\n", err))
		return "", err
	}

	// Путь для сохранения файла
	reelName, err := extractReelID(link)
	if err != nil {
		hash := sha256.New()
		hash.Write([]byte(link))
		hashBytes := hash.Sum(nil)
		reelName = hex.EncodeToString(hashBytes)
	}

	reelName += ".mp4"
	filePath := filepath.Join("./tmpData/inst/", reelName)
	err = r.downloadFile(downloadLink, filePath)
	if err != nil {
		r.logger.Error(fmt.Sprintf("не удалось скачать рилс по ссылке: %v", err))
		return "", nil
	}

	return filePath, nil
}

func NewReelsDownloader() (ReelModule, error) {
	dl := reelsDownloader{}
	logger := loggers.CreateLogger(loggers.LoggerConfig{
		Name:           "MainLog",
		Path:           "./DefLogs.txt",
		Level:          loggers.InfoLevel,
		WriteToConsole: true,
		UseColor:       true,
	})

	dl.logger = logger

	var err error
	dl.driver, err = pw.Run()
	if err != nil {
		dl.logger.Error(fmt.Sprintf("не удалось запустить драйвер: %s", err))
		return nil, err
	}

	// Запуск браузера (в headless-режиме)
	dl.browser, err = dl.driver.Chromium.Launch(pw.BrowserTypeLaunchOptions{
		Headless: pw.Bool(true),
	})
	if err != nil {
		dl.logger.Error(fmt.Sprintf("не удалось запустить браузер: %s", err))
	}

	return &dl, nil
}

func extractReelID(url string) (string, error) {
	// Регулярное выражение для извлечения TgID
	re := regexp.MustCompile(`https:\/\/www\.instagram\.com\/(reel|reels)\/([A-Za-z0-9_-]+)\/?`)
	match := re.FindStringSubmatch(url)
	if len(match) < 3 {
		return "", fmt.Errorf("не удалось найти TgID в ссылке")
	}
	return match[2], nil
}
