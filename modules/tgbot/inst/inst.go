package inst

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/tebeka/selenium"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"superserver/loggers"
	"time"
)

const (
	seleniumPath      = "chromedriver.exe"
	seleniumPathLinux = "chromedriver"
	port              = 9515
)

type ReelModule interface {
	DownloadReel(reelLink string) (string, error)
}

type reelsDownloader struct {
	driver selenium.WebDriver
	logger loggers.Logger
}

// Парсит страницу с рилсом и скачивает в директорию
// @param reelURL - URL рилса
// @return - возвращает имя рилса в директории
func (r *reelsDownloader) DownloadReel(reelURL string) (string, error) {
	if err := r.driver.Get("https://fastdl.app/instagram-reels-download"); err != nil {
		r.logger.Error(err)
		return "", err
	}

	time.Sleep(1500 * time.Millisecond)
	consentButton, err := r.driver.FindElement(selenium.ByCSSSelector, "button.fc-button.fc-cta-consent.fc-primary-button")
	if err == nil {
		r.logger.Info("обнаружено всплывающее окно")
		err = consentButton.Click()
		if err != nil {
			r.logger.Info(fmt.Sprintf("не удалось нажать кнопку всплывающего окна: %v", err))
			return "", err
		}
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
	r.logger.Info(fmt.Sprintf("получена ссылка для скачивания видео: %s", downloadLink))

	// Путь для сохранения файла
	var reelName string
	reelName, err = extractReelID(reelURL)
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

	caps := selenium.Capabilities{"browserName": "chrome"}

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
				"--headless",
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
