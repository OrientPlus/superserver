package inst

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	pw "github.com/playwright-community/playwright-go"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"superserver/loggers"
	"sync"
	"time"
)

const (
	chromiumPath      = "chrome.exe"
	chromiumPathLinux = "chromium"
)

type ReelModule interface {
	DownloadReel(reelLink string) (string, error)
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
func (r *reelsDownloader) DownloadReel(reelURL string) (string, error) {
	/*r.newHandleMutex.Lock()

	r.driver.NewSession()

	// Открываем новую вкладку
	_, err := r.driver.ExecuteScript("window.open()", nil)
	if err != nil {
		r.newHandleMutex.Unlock()
		r.logger.Error(fmt.Sprintf("не удалось открыть новую вкладку: %s", err))
		return "", err
	}
	time.Sleep(200 * time.Millisecond)
	r.logger.Debug("открыта новая вкладка")

	// Получаем список хендлов всех вкладок
	tabs, err := r.driver.WindowHandles()
	if err != nil {
		r.newHandleMutex.Unlock()
		r.logger.Error(fmt.Sprintf("не удалось получить список вкладок: %s", err))
		return "", err
	}

	// Последняя вкладка - только что созданная. Переключаемся на нее
	newTabHandle := tabs[len(tabs)-1]
	err = r.driver.SwitchWindow(newTabHandle)
	if err != nil {
		r.newHandleMutex.Unlock()
		r.logger.Error(fmt.Sprintf("не удалось переключиться на новую вкладку: %s", err))
		return "", err
	}
	r.logger.Debug("удалось переключиться на новую вкладку")
	defer func() {
		r.driver.Close()
		err = r.driver.SwitchWindow(tabs[0])
		if err != nil {
			r.logger.Error(fmt.Sprintf("не удалось переключиться на нулевую вкладку: %s", err))
		}
	}()
	r.newHandleMutex.Unlock()

	if err = r.driver.Get("https://fastdl.app/instagram-reels-download"); err != nil {
		r.logger.Error(err)
		return "", err
	}
	r.logger.Debug("открыта страница для скачивания")

	var consentButton selenium.WebElement
	for range 7 {
		consentButton, err = r.driver.FindElement(selenium.ByCSSSelector, "button.fc-button.fc-cta-consent.fc-primary-button")
		if err == nil {
			r.logger.Info("обнаружено всплывающее окно")
			err = consentButton.Click()
			if err != nil {
				r.logger.Info(fmt.Sprintf("не удалось нажать кнопку всплывающего окна: %v", err))
				return "", err
			}
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Ищем поле для ввода
	inputElement, err := r.driver.FindElement(selenium.ByCSSSelector, "#search-form-input")
	if err != nil {
		r.logger.Error(fmt.Sprintf("не удалось обнаружить поле ввода для ссылки: %v", err))
		return "", err
	}

	// Ввод ссылки на Reel
	if err = inputElement.SendKeys(reelURL); err != nil {
		r.logger.Error(fmt.Sprintf("не удалось ввести ссылку в поле для ввода: %v", err))
		return "", err
	}

	// Нажать кнопку "Search"
	searchButton, err := r.driver.FindElement(selenium.ByCSSSelector, "button.search-form__button") //button[type='submit']
	if err != nil {
		r.logger.Warn(fmt.Sprintf("не удалось найти кнопку отправки ссылки: %v", err))
	}
	if err = searchButton.Click(); err != nil {
		r.logger.Warn(fmt.Sprintf("не удалось нажать кнопку отправки ссылки: %v", err))
	}

	// Найдите кнопку "Download"
	time.Sleep(2 * time.Second) // Здесь лучше использовать ожидание видимости элемента
	var downloadButton selenium.WebElement
	for range 25 {
		downloadButton, err = r.driver.FindElement(selenium.ByCSSSelector, "a.button__download")
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second) // Здесь лучше использовать ожидание видимости элемента
	}
	if err != nil {
		r.logger.Error(fmt.Sprintf("не удалось найти кнопку скачивания рилса: %v", err))
		return "", err
	}

	// Получаем ссылку для скачивания
	downloadLink, err := downloadButton.GetAttribute("href")
	if err != nil {
		r.logger.Error(fmt.Sprintf("не удалось получить ссылку для скачивания: %v", err))
		return "", err
	}
	r.logger.Info(fmt.Sprintf("получена ссылка для скачивания видео: %s", downloadLink))*/

	// Создание новой вкладки
	page, err := r.browser.NewPage()
	if err != nil {
		r.logger.Error(fmt.Sprintf("не удалось открыть вкладку: %s", err))
	}

	// Переход на страницу для скачивания
	if _, err = page.Goto("https://fastdl.app/instagram-reels-download"); err != nil {
		r.logger.Error(fmt.Sprintf("не удалось открыть страницу для скачивания: %s", err))
		return "", err
	}
	r.logger.Debug("открыта страница для скачивания")

	// Ожидание появления и нажатие кнопки "Consent"
	for i := 0; i < 7; i++ {
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

func NewReelsDownloader() (ReelModule, error) {
	dl := reelsDownloader{}
	logger := loggers.CreateLogger(loggers.LoggerConfig{
		Name:           "MainLog",
		Path:           "./DefLogs.txt",
		Level:          loggers.DebugLevel,
		WriteToConsole: true,
		UseColor:       true,
	})

	dl.logger = logger

	/*caps := selenium.Capabilities{"browserName": "chrome"}

	var err error
	dl.driver, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", 46429))
	if err != nil {
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command(seleniumPath, fmt.Sprintf("--port=%d", port))
		}
		if runtime.GOOS == "linux" {
			cmd = exec.Command(seleniumPathLinux, fmt.Sprintf("--port=%d", port))
		}

		err = cmd.Start()
		if err != nil {
			return nil, err
		}
		time.Sleep(3 * time.Second)

		caps = selenium.Capabilities{"browserName": "chrome"}
		chromeCaps := map[string]interface{}{
			"args": []string{
				//"--headless",
				"--disable-gpu",
				"--no-sandbox",
				"--disable-dev-shm-usage",
			},
		}

		caps["goog:chromeOptions"] = chromeCaps

		dl.driver, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", port))
		if err != nil {
			return nil, err
		}
	}*/

	var err error
	dl.driver, err = pw.Run()
	if err != nil {
		dl.logger.Error(fmt.Sprintf("не удалось запустить драйвер: %s", err))
		return nil, err
	}

	// Запуск браузера (в headless-режиме)
	dl.browser, err = dl.driver.Chromium.Launch(pw.BrowserTypeLaunchOptions{
		Headless:       pw.Bool(true),
		ExecutablePath: pw.String(chromiumPath),
	})
	if err != nil {
		dl.logger.Error(fmt.Sprintf("не удалось запустить браузер: %s", err))
	}

	return &dl, nil
}

func extractReelID(url string) (string, error) {
	// Регулярное выражение для извлечения ID
	re := regexp.MustCompile(`https:\/\/www\.instagram\.com\/(reel|reels)\/([A-Za-z0-9_-]+)\/?`)
	match := re.FindStringSubmatch(url)
	if len(match) < 3 {
		return "", fmt.Errorf("не удалось найти ID в ссылке")
	}
	return match[2], nil
}
